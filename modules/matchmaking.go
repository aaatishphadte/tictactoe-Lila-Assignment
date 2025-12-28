package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
)

// MatchmakingQueue represents players waiting for a match
type MatchmakingQueue struct {
	UserID    string    `json:"user_id"`
	GameMode  string    `json:"game_mode"`
	Rating    int       `json:"rating"`
	Token     string    `json:"token"`
	Timestamp time.Time `json:"timestamp"`
}

// JoinQueueRequest represents a request to join matchmaking
type JoinQueueRequest struct {
	GameMode string `json:"game_mode"` // "casual" or "ranked"
}

// JoinQueueResponse represents the response after joining queue
type JoinQueueResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	MatchID string `json:"match_id,omitempty"`
	Matched bool   `json:"matched"`
}

// CancelQueueRequest represents a request to cancel matchmaking
type CancelQueueRequest struct {
	Token string `json:"token"`
}

// RpcJoinQueue handles a player joining the matchmaking queue
func RpcJoinQueue(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("user not authenticated", 16)
	}

	var request JoinQueueRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		return "", runtime.NewError("invalid request payload", 3)
	}

	// Validate game mode
	if request.GameMode != "casual" && request.GameMode != "ranked" {
		return "", runtime.NewError("invalid game_mode, must be 'casual' or 'ranked'", 3)
	}

	// Get user profile for rating
	profile, err := GetUserProfile(ctx, logger, nk, userID)
	if err != nil {
		logger.Error("Failed to get user profile: %v", err)
		return "", runtime.NewError("failed to get user profile", 13)
	}

	// Generate matchmaking token
	token := uuid.New().String()

	// Create queue entry
	queueEntry := MatchmakingQueue{
		UserID:    userID,
		GameMode:  request.GameMode,
		Rating:    profile.Rating,
		Token:     token,
		Timestamp: time.Now(),
	}

	// Try to find a match
	matchID, opponent, err := FindMatch(ctx, logger, nk, &queueEntry)
	if err != nil {
		logger.Error("Error finding match: %v", err)
		return "", runtime.NewError("matchmaking failed", 13)
	}

	if matchID != "" {
		// Match found!
		logger.Info("Match found - Match ID: %s, Players: %s vs %s", matchID, userID, opponent.UserID)

		response := JoinQueueResponse{
			Token:   token,
			Message: "match found",
			MatchID: matchID,
			Matched: true,
		}

		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	// No match found, add to queue
	if err := AddToQueue(ctx, nk, &queueEntry); err != nil {
		logger.Error("Failed to add to queue: %v", err)
		return "", runtime.NewError("failed to join queue", 13)
	}

	logger.Info("Player added to queue - UserID: %s, Mode: %s, Token: %s", userID, request.GameMode, token)

	response := JoinQueueResponse{
		Token:   token,
		Message: "waiting for opponent",
		Matched: false,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// RpcCancelQueue handles a player canceling matchmaking
func RpcCancelQueue(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("user not authenticated", 16)
	}

	var request CancelQueueRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		return "", runtime.NewError("invalid request payload", 3)
	}

	if request.Token == "" {
		return "", runtime.NewError("token is required", 3)
	}

	// Remove from queue
	if err := RemoveFromQueue(ctx, nk, userID); err != nil {
		logger.Error("Failed to remove from queue: %v", err)
		return "", runtime.NewError("failed to cancel queue", 13)
	}

	logger.Info("Player removed from queue - UserID: %s, Token: %s", userID, request.Token)

	return `{"success": true, "message": "removed from queue"}`, nil
}

// FindMatch attempts to find a suitable opponent in the queue
func FindMatch(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, player *MatchmakingQueue) (string, *MatchmakingQueue, error) {
	// List all queue entries for the same game mode
	objects, _, err := nk.StorageList(ctx, "", "", "matchmaking_queue", 100, "")
	if err != nil {
		return "", nil, err
	}

	var bestMatch *MatchmakingQueue
	var bestMatchKey string

	for _, obj := range objects {
		var opponent MatchmakingQueue
		if err := json.Unmarshal([]byte(obj.Value), &opponent); err != nil {
			continue
		}

		// Skip if same user or different game mode
		if opponent.UserID == player.UserID || opponent.GameMode != player.GameMode {
			continue
		}

		// For ranked, check rating difference (within 200 rating)
		if player.GameMode == "ranked" {
			ratingDiff := abs(player.Rating - opponent.Rating)
			if ratingDiff > 200 {
				continue
			}
		}

		// Found a suitable match
		bestMatch = &opponent
		bestMatchKey = obj.Key
		break
	}

	if bestMatch == nil {
		// No match found
		return "", nil, nil
	}

	// Create a new match
	matchID := uuid.New().String()

	// Randomly assign X and O
	playerX := player.UserID
	playerO := bestMatch.UserID
	if time.Now().UnixNano()%2 == 0 {
		playerX, playerO = playerO, playerX
	}

	// Create game state
	gameState := NewGameState(matchID, playerX, playerO, player.GameMode)
	if err := SaveGameState(ctx, nk, gameState); err != nil {
		return "", nil, err
	}

	// Remove opponent from queue
	if err := nk.StorageDelete(ctx, []*runtime.StorageDelete{
		{
			Collection: "matchmaking_queue",
			Key:        bestMatchKey,
			UserID:     "",
		},
	}); err != nil {
		logger.Error("Failed to remove opponent from queue: %v", err)
	}

	logger.Info("Created match - ID: %s, X: %s, O: %s", matchID, playerX, playerO)

	return matchID, bestMatch, nil
}

// AddToQueue adds a player to the matchmaking queue
func AddToQueue(ctx context.Context, nk runtime.NakamaModule, entry *MatchmakingQueue) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	writes := []*runtime.StorageWrite{
		{
			Collection:      "matchmaking_queue",
			Key:             entry.UserID,
			UserID:          "",
			Value:           string(data),
			PermissionRead:  0,
			PermissionWrite: 0,
		},
	}

	if _, err := nk.StorageWrite(ctx, writes); err != nil {
		return err
	}

	return nil
}

// RemoveFromQueue removes a player from the matchmaking queue
func RemoveFromQueue(ctx context.Context, nk runtime.NakamaModule, userID string) error {
	deletes := []*runtime.StorageDelete{
		{
			Collection: "matchmaking_queue",
			Key:        userID,
			UserID:     "",
		},
	}

	if err := nk.StorageDelete(ctx, deletes); err != nil {
		return err
	}

	return nil
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// CleanupExpiredQueueEntries removes queue entries older than 60 seconds
func CleanupExpiredQueueEntries(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	objects, _, err := nk.StorageList(ctx, "", "", "matchmaking_queue", 100, "")
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-60 * time.Second)
	var deletes []*runtime.StorageDelete

	for _, obj := range objects {
		var entry MatchmakingQueue
		if err := json.Unmarshal([]byte(obj.Value), &entry); err != nil {
			continue
		}

		if entry.Timestamp.Before(cutoff) {
			deletes = append(deletes, &runtime.StorageDelete{
				Collection: "matchmaking_queue",
				Key:        obj.Key,
				UserID:     "",
			})
		}
	}

	if len(deletes) > 0 {
		if err := nk.StorageDelete(ctx, deletes); err != nil {
			return fmt.Errorf("failed to cleanup expired entries: %w", err)
		}
		logger.Info("Cleaned up %d expired queue entries", len(deletes))
	}

	return nil
}
