package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

// TicTacToeMatch represents the match handler for real-time gameplay
type TicTacToeMatch struct{}

// MatchState holds the state for a match
type MatchState struct {
	MatchID      string                      `json:"match_id"`
	GameState    *GameState                  `json:"game_state"`
	PresenceList map[string]runtime.Presence `json:"-"`
}

// OpCode represents message operation codes
const (
	OpCodeMove         int64 = 1
	OpCodeGameState    int64 = 2
	OpCodePlayerJoined int64 = 3
	OpCodePlayerLeft   int64 = 4
	OpCodeGameOver     int64 = 5
)

// MatchMessage represents a message sent in the match
type MatchMessage struct {
	OpCode int64           `json:"op_code"`
	Data   json.RawMessage `json:"data"`
}

// MatchInit initializes the match
func (m *TicTacToeMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.Info("Match initialized")

	matchID := ""
	if id, ok := params["match_id"].(string); ok {
		matchID = id
	}

	// Extract player IDs from matchmaker (if provided)
	var player1, player2 string
	if p1, ok := params["player1"].(string); ok {
		player1 = p1
		logger.Info("Player 1 (X) assigned: %s", player1)
	}
	if p2, ok := params["player2"].(string); ok {
		player2 = p2
		logger.Info("Player 2 (O) assigned: %s", player2)
	}

	// Create match state with player assignments
	state := &MatchState{
		MatchID:      matchID,
		GameState:    nil,
		PresenceList: make(map[string]runtime.Presence),
	}

	// If both players are assigned, we can pre-initialize the game state
	// It will be finalized when players actually join
	if player1 != "" && player2 != "" {
		state.GameState = NewGameState(matchID, player1, player2, "casual")
		logger.Info("Game state pre-initialized for matchmaker match")
	}

	// Tick rate: 10 times per second
	tickRate := 10
	label := ""

	return state, tickRate, label
}

// MatchJoinAttempt validates whether a player can join the match
func (m *TicTacToeMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	matchState, ok := state.(*MatchState)
	if !ok {
		return state, false, "invalid match state"
	}

	// Allow up to 2 players
	if len(matchState.PresenceList) >= 2 {
		return state, false, "match is full"
	}

	logger.Info("Player attempting to join - UserID: %s", presence.GetUserId())
	return state, true, ""
}

// MatchJoin is called when a player successfully joins the match
func (m *TicTacToeMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		return state
	}

	for _, presence := range presences {
		matchState.PresenceList[presence.GetUserId()] = presence
		logger.Info("Player joined match - UserID: %s, Total players: %d", presence.GetUserId(), len(matchState.PresenceList))

		// Broadcast player joined event
		message := map[string]interface{}{
			"user_id":  presence.GetUserId(),
			"username": presence.GetUsername(),
		}
		messageJSON, _ := json.Marshal(message)

		envelope := &MatchMessage{
			OpCode: OpCodePlayerJoined,
			Data:   messageJSON,
		}
		envelopeJSON, _ := json.Marshal(envelope)

		dispatcher.BroadcastMessage(OpCodePlayerJoined, envelopeJSON, nil, nil, true)
	}

	// If both players are present, ensure game state is ready
	if len(matchState.PresenceList) == 2 {
		// If game state was pre-initialized by matchmaker, just broadcast it
		if matchState.GameState != nil {
			logger.Info("Both players joined matchmaker match - Match ID: %s", matchState.MatchID)
			m.broadcastGameState(dispatcher, matchState.GameState)
		} else {
			// Manual match creation (fallback for non-matchmaker matches)
			players := make([]string, 0, 2)
			for userID := range matchState.PresenceList {
				players = append(players, userID)
			}

			// Generate match ID if not set
			if matchState.MatchID == "" {
				matchState.MatchID = "match_" + players[0] + "_" + players[1]
			}

			matchState.GameState = NewGameState(
				matchState.MatchID,
				players[0],
				players[1],
				"casual", // Default to casual
			)

			logger.Info("Game started (manual match) - Match ID: %s", matchState.MatchID)
			m.broadcastGameState(dispatcher, matchState.GameState)
		}
	}

	return matchState
}

// MatchLeave is called when a player leaves the match
func (m *TicTacToeMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		return state
	}

	for _, presence := range presences {
		delete(matchState.PresenceList, presence.GetUserId())
		logger.Info("Player left match - UserID: %s", presence.GetUserId())

		// Broadcast player left event
		message := map[string]interface{}{
			"user_id": presence.GetUserId(),
		}
		messageJSON, _ := json.Marshal(message)

		envelope := &MatchMessage{
			OpCode: OpCodePlayerLeft,
			Data:   messageJSON,
		}
		envelopeJSON, _ := json.Marshal(envelope)

		dispatcher.BroadcastMessage(OpCodePlayerLeft, envelopeJSON, nil, nil, true)
	}

	// If a player leaves during an active game, declare opponent as winner
	if matchState.GameState != nil && matchState.GameState.Status == GameStatusActive && len(matchState.PresenceList) < 2 {
		// Find remaining player
		for userID := range matchState.PresenceList {
			matchState.GameState.Status = GameStatusFinished
			matchState.GameState.Winner = userID

			if userID == matchState.GameState.PlayerX {
				matchState.GameState.Result = GameResultXWins
			} else {
				matchState.GameState.Result = GameResultOWins
			}

			logger.Info("Game ended due to player disconnect - Winner: %s", userID)
			m.broadcastGameState(dispatcher, matchState.GameState)

			// Broadcast Game Over OpCode (5)
			stateJSON, _ := json.Marshal(matchState.GameState)
			envelope := &MatchMessage{
				OpCode: OpCodeGameOver,
				Data:   stateJSON,
			}
			envelopeJSON, _ := json.Marshal(envelope)
			dispatcher.BroadcastMessage(OpCodeGameOver, envelopeJSON, nil, nil, true)
			break
		}
	}

	return matchState
}

// MatchLoop is called on each tick
func (m *TicTacToeMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState, ok := state.(*MatchState)
	if !ok {
		return state
	}

	// Process incoming messages
	for _, message := range messages {
		switch message.GetOpCode() {
		case OpCodeMove:
			m.handleMove(ctx, logger, nk, dispatcher, matchState, message)
		}
	}

	// End match if game is finished and no players remain
	if matchState.GameState != nil && matchState.GameState.Status == GameStatusFinished && len(matchState.PresenceList) == 0 {
		return nil
	}

	return matchState
}

// MatchTerminate is called when the match ends
func (m *TicTacToeMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	matchState, ok := state.(*MatchState)
	if ok {
		logger.Info("Match terminated - Match ID: %s", matchState.MatchID)
	}
	return state
}

// MatchSignal handles custom signals sent to the match
func (m *TicTacToeMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

// handleMove processes a move message from a player
func (m *TicTacToeMatch) handleMove(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, matchState *MatchState, message runtime.MatchData) {
	if matchState.GameState == nil {
		logger.Warn("Received move but game state is nil")
		return
	}

	var move Move
	if err := json.Unmarshal(message.GetData(), &move); err != nil {
		logger.Error("Failed to unmarshal move: %v", err)
		return
	}

	userID := message.GetUserId()

	// Apply the move
	if err := matchState.GameState.ApplyMove(move.Row, move.Col, userID); err != nil {
		logger.Warn("Invalid move from %s: %v", userID, err)
		return
	}

	logger.Info("Move applied - UserID: %s, Position: (%d,%d)", userID, move.Row, move.Col)

	// Broadcast updated game state
	m.broadcastGameState(dispatcher, matchState.GameState)

	// If game is finished, update stats and broadcast Game Over
	if matchState.GameState.Status == GameStatusFinished {
		logger.Info("Game finished - Result: %s, Winner: %s", matchState.GameState.Result, matchState.GameState.Winner)

		// Broadcast Game Over OpCode (5)
		stateJSON, _ := json.Marshal(matchState.GameState)
		envelope := &MatchMessage{
			OpCode: OpCodeGameOver,
			Data:   stateJSON,
		}
		envelopeJSON, _ := json.Marshal(envelope)
		dispatcher.BroadcastMessage(OpCodeGameOver, envelopeJSON, nil, nil, true)

		// Update player stats (async)
		go func() {
			if err := UpdatePlayerStats(ctx, logger, nk, matchState.GameState); err != nil {
				logger.Error("Failed to update player stats: %v", err)
			}
		}()
	}
}

// broadcastGameState sends the current game state to all players
func (m *TicTacToeMatch) broadcastGameState(dispatcher runtime.MatchDispatcher, gameState *GameState) {
	stateJSON, err := json.Marshal(gameState)
	if err != nil {
		return
	}

	envelope := &MatchMessage{
		OpCode: OpCodeGameState,
		Data:   stateJSON,
	}
	envelopeJSON, _ := json.Marshal(envelope)

	dispatcher.BroadcastMessage(OpCodeGameState, envelopeJSON, nil, nil, true)
}
