package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

// InitModule initializes the Nakama runtime module
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("TicTacToe module loaded")

	// Register Authentication RPCs
	if err := initializer.RegisterRpc("authenticate_device", RpcAuthenticateDevice); err != nil {
		return err
	}
	logger.Info("Registered RPC: authenticate_device")

	// Register Game Logic RPCs
	if err := initializer.RegisterRpc("make_move", RpcMakeMove); err != nil {
		return err
	}
	logger.Info("Registered RPC: make_move")

	if err := initializer.RegisterRpc("get_game_state", RpcGetGameState); err != nil {
		return err
	}
	logger.Info("Registered RPC: get_game_state")

	if err := initializer.RegisterRpc("resign_game", RpcResignGame); err != nil {
		return err
	}
	logger.Info("Registered RPC: resign_game")

	// Register Matchmaking RPCs
	if err := initializer.RegisterRpc("join_queue", RpcJoinQueue); err != nil {
		return err
	}
	logger.Info("Registered RPC: join_queue")

	if err := initializer.RegisterRpc("cancel_queue", RpcCancelQueue); err != nil {
		return err
	}
	logger.Info("Registered RPC: cancel_queue")

	// Register Leaderboard RPCs
	if err := initializer.RegisterRpc("get_leaderboard", RpcGetLeaderboard); err != nil {
		return err
	}
	logger.Info("Registered RPC: get_leaderboard")

	if err := initializer.RegisterRpc("get_player_rank", RpcGetPlayerRank); err != nil {
		return err
	}
	logger.Info("Registered RPC: get_player_rank")

	// Register Match Handler for real-time gameplay
	if err := initializer.RegisterMatch("tictactoe", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &TicTacToeMatch{}, nil
	}); err != nil {
		return err
	}
	logger.Info("Registered Match Handler: tictactoe")

	// Register Matchmaker Matched Handler for built-in matchmaking
	if err := initializer.RegisterMatchmakerMatched(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
		logger.Info("Matchmaker matched %d players", len(entries))

		if len(entries) != 2 {
			logger.Error("Invalid number of players matched: %d, expected 2", len(entries))
			return "", nil
		}

		playerIDs := make([]string, 0, 2)
		for _, entry := range entries {
			playerIDs = append(playerIDs, entry.GetPresence().GetUserId())
			logger.Info("Player matched - UserID: %s, Username: %s",
				entry.GetPresence().GetUserId(),
				entry.GetPresence().GetUsername())
		}

		params := map[string]interface{}{
			"player1": playerIDs[0],
			"player2": playerIDs[1],
		}

		matchID, err := nk.MatchCreate(ctx, "tictactoe", params)
		if err != nil {
			logger.Error("Failed to create match: %v", err)
			return "", err
		}

		logger.Info("Match created - Match ID: %s, Players: %s vs %s", matchID, playerIDs[0], playerIDs[1])
		return matchID, nil
	}); err != nil {
		return err
	}
	logger.Info("Registered Matchmaker Matched Handler")

	// Initialize Leaderboard
	if err := InitializeLeaderboard(ctx, logger, nk); err != nil {
		logger.Error("Failed to initialize leaderboard: %v", err)
		return err
	}
	logger.Info("Leaderboard initialized")

	logger.Info("TicTacToe module initialization complete")
	return nil
}
