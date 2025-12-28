package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// MakeMoveRequest represents a move request
type MakeMoveRequest struct {
	MatchID string `json:"match_id"`
	Row     int    `json:"row"`
	Col     int    `json:"col"`
}

// MakeMoveResponse represents the response after a move
type MakeMoveResponse struct {
	Success   bool      `json:"success"`
	GameState GameState `json:"game_state"`
	Message   string    `json:"message"`
}

// GetGameStateRequest represents a request to get game state
type GetGameStateRequest struct {
	MatchID string `json:"match_id"`
}

// ResignGameRequest represents a resignation request
type ResignGameRequest struct {
	MatchID string `json:"match_id"`
}

// RpcMakeMove handles a player making a move
func RpcMakeMove(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Get user ID from context
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("user not authenticated", 16) // UNAUTHENTICATED
	}

	var request MakeMoveRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal move request: %v", err)
		return "", runtime.NewError("invalid request payload", 3)
	}

	if request.MatchID == "" {
		return "", runtime.NewError("match_id is required", 3)
	}

	// Load game state from storage
	gameState, err := LoadGameState(ctx, nk, request.MatchID)
	if err != nil {
		logger.Error("Failed to load game state: %v", err)
		return "", runtime.NewError("game not found", 5) // NOT_FOUND
	}

	// Apply the move
	if err := gameState.ApplyMove(request.Row, request.Col, userID); err != nil {
		logger.Warn("Invalid move: %v", err)
		response := MakeMoveResponse{
			Success:   false,
			GameState: *gameState,
			Message:   err.Error(),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	// Save updated game state
	if err := SaveGameState(ctx, nk, gameState); err != nil {
		logger.Error("Failed to save game state: %v", err)
		return "", runtime.NewError("failed to save game state", 13)
	}

	logger.Info("Move applied - Match: %s, Player: %s, Position: (%d,%d)", request.MatchID, userID, request.Row, request.Col)

	// If game is finished, update player stats
	if gameState.Status == GameStatusFinished {
		if err := UpdatePlayerStats(ctx, logger, nk, gameState); err != nil {
			logger.Error("Failed to update player stats: %v", err)
		}
	}

	response := MakeMoveResponse{
		Success:   true,
		GameState: *gameState,
		Message:   "move successful",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// RpcGetGameState retrieves the current game state
func RpcGetGameState(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request GetGameStateRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		return "", runtime.NewError("invalid request payload", 3)
	}

	if request.MatchID == "" {
		return "", runtime.NewError("match_id is required", 3)
	}

	gameState, err := LoadGameState(ctx, nk, request.MatchID)
	if err != nil {
		logger.Error("Failed to load game state: %v", err)
		return "", runtime.NewError("game not found", 5)
	}

	responseJSON, err := json.Marshal(gameState)
	if err != nil {
		logger.Error("Failed to marshal game state: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// RpcResignGame allows a player to resign from a game
func RpcResignGame(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("user not authenticated", 16)
	}

	var request ResignGameRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		return "", runtime.NewError("invalid request payload", 3)
	}

	if request.MatchID == "" {
		return "", runtime.NewError("match_id is required", 3)
	}

	gameState, err := LoadGameState(ctx, nk, request.MatchID)
	if err != nil {
		logger.Error("Failed to load game state: %v", err)
		return "", runtime.NewError("game not found", 5)
	}

	// Mark game as finished with opponent as winner
	gameState.Status = GameStatusFinished
	if userID == gameState.PlayerX {
		gameState.Result = GameResultOWins
		gameState.Winner = gameState.PlayerO
	} else {
		gameState.Result = GameResultXWins
		gameState.Winner = gameState.PlayerX
	}

	// Save updated state
	if err := SaveGameState(ctx, nk, gameState); err != nil {
		logger.Error("Failed to save game state: %v", err)
		return "", runtime.NewError("failed to save game state", 13)
	}

	// Update player stats
	if err := UpdatePlayerStats(ctx, logger, nk, gameState); err != nil {
		logger.Error("Failed to update player stats: %v", err)
	}

	logger.Info("Player resigned - Match: %s, Player: %s", request.MatchID, userID)

	responseJSON, err := json.Marshal(gameState)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// LoadGameState loads a game state from storage
func LoadGameState(ctx context.Context, nk runtime.NakamaModule, matchID string) (*GameState, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: "games",
			Key:        matchID,
			UserID:     "",
		},
	})

	if err != nil {
		return nil, err
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("game not found")
	}

	return GameStateFromJSON(objects[0].Value)
}

// SaveGameState saves a game state to storage
func SaveGameState(ctx context.Context, nk runtime.NakamaModule, gameState *GameState) error {
	data, err := gameState.ToJSON()
	if err != nil {
		return err
	}

	writes := []*runtime.StorageWrite{
		{
			Collection:      "games",
			Key:             gameState.MatchID,
			UserID:          "",
			Value:           data,
			PermissionRead:  1, // Public read
			PermissionWrite: 0, // No client write
		},
	}

	if _, err := nk.StorageWrite(ctx, writes); err != nil {
		return err
	}

	return nil
}

// UpdatePlayerStats updates player statistics after a game
func UpdatePlayerStats(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, gameState *GameState) error {
	// Update Player X stats
	profileX, err := GetUserProfile(ctx, logger, nk, gameState.PlayerX)
	if err != nil {
		return err
	}

	// Update Player O stats
	profileO, err := GetUserProfile(ctx, logger, nk, gameState.PlayerO)
	if err != nil {
		return err
	}

	// Determine outcome and update stats
	switch gameState.Result {
	case GameResultXWins:
		profileX.Wins++
		profileO.Losses++
		// Update ratings
		UpdateRatings(&profileX, &profileO, 1.0) // X wins
	case GameResultOWins:
		profileO.Wins++
		profileX.Losses++
		// Update ratings
		UpdateRatings(&profileX, &profileO, 0.0) // O wins
	case GameResultDraw:
		profileX.Draws++
		profileO.Draws++
		// Update ratings
		UpdateRatings(&profileX, &profileO, 0.5) // Draw
	}

	// Save updated profiles
	if err := UpdateUserProfile(ctx, logger, nk, gameState.PlayerX, profileX); err != nil {
		return err
	}
	if err := UpdateUserProfile(ctx, logger, nk, gameState.PlayerO, profileO); err != nil {
		return err
	}

	// Update leaderboard if ranked game
	if gameState.GameMode == "ranked" {
		if err := SubmitLeaderboardScore(ctx, logger, nk, gameState.PlayerX, int64(profileX.Rating)); err != nil {
			logger.Error("Failed to update leaderboard for player X: %v", err)
		}
		if err := SubmitLeaderboardScore(ctx, logger, nk, gameState.PlayerO, int64(profileO.Rating)); err != nil {
			logger.Error("Failed to update leaderboard for player O: %v", err)
		}
	}

	return nil
}
