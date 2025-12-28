package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	LeaderboardID = "global_rankings"
	KFactor       = 32.0 // ELO K-factor
)

// LeaderboardEntry represents a player's leaderboard entry
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Rank     int64  `json:"rank"`
	Score    int64  `json:"score"`
	NumScore int    `json:"num_score"`
}

// GetLeaderboardResponse represents the leaderboard response
type GetLeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}

// GetPlayerRankRequest represents a request to get player rank
type GetPlayerRankRequest struct {
	UserID string `json:"user_id,omitempty"`
}

// GetPlayerRankResponse represents player rank response
type GetPlayerRankResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Rank     int64  `json:"rank"`
	Rating   int64  `json:"rating"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Draws    int    `json:"draws"`
}

// InitializeLeaderboard creates the global leaderboard on startup
func InitializeLeaderboard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	// Create leaderboard with descending order (higher score = better)
	err := nk.LeaderboardCreate(ctx, LeaderboardID, false, "desc", "best", "", nil)
	if err != nil {
		// Leaderboard might already exist, which is fine
		logger.Info("Leaderboard already exists or created")
	} else {
		logger.Info("Created leaderboard: %s", LeaderboardID)
	}

	return nil
}

// RpcGetLeaderboard retrieves the top players from the leaderboard
func RpcGetLeaderboard(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Get top 100 players
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, LeaderboardID, nil, 100, "", 0)
	if err != nil {
		logger.Error("Failed to get leaderboard: %v", err)
		return "", runtime.NewError("failed to retrieve leaderboard", 13)
	}

	var entries []LeaderboardEntry
	for _, record := range records {
		entries = append(entries, LeaderboardEntry{
			UserID:   record.OwnerId,
			Username: record.Username.Value,
			Rank:     record.Rank,
			Score:    record.Score,
			NumScore: int(record.NumScore),
		})
	}

	response := GetLeaderboardResponse{
		Entries: entries,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// RpcGetPlayerRank retrieves a specific player's rank and stats
func RpcGetPlayerRank(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Get user ID from request or context
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		// Try to get from payload
		var request GetPlayerRankRequest
		if err := json.Unmarshal([]byte(payload), &request); err == nil && request.UserID != "" {
			userID = request.UserID
		} else {
			return "", runtime.NewError("user not authenticated", 16)
		}
	}

	// Get user profile
	profile, err := GetUserProfile(ctx, logger, nk, userID)
	if err != nil {
		logger.Error("Failed to get user profile: %v", err)
		return "", runtime.NewError("failed to get profile", 13)
	}

	// Get user's account info
	account, err := nk.AccountGetId(ctx, userID)
	if err != nil {
		logger.Error("Failed to get account: %v", err)
		return "", runtime.NewError("failed to get account", 13)
	}

	// Get leaderboard record
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, LeaderboardID, []string{userID}, 1, "", 0)
	if err != nil {
		logger.Error("Failed to get player rank: %v", err)
		return "", runtime.NewError("failed to get rank", 13)
	}

	var rank int64 = 0
	if len(records) > 0 {
		rank = records[0].Rank
	}

	response := GetPlayerRankResponse{
		UserID:   userID,
		Username: account.User.Username,
		Rank:     rank,
		Rating:   int64(profile.Rating),
		Wins:     profile.Wins,
		Losses:   profile.Losses,
		Draws:    profile.Draws,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// SubmitLeaderboardScore submits a player's rating to the leaderboard
func SubmitLeaderboardScore(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string, rating int64) error {
	// Get username
	account, err := nk.AccountGetId(ctx, userID)
	if err != nil {
		return err
	}

	// Submit score to leaderboard
	_, err = nk.LeaderboardRecordWrite(ctx, LeaderboardID, userID, account.User.Username, rating, 0, nil, nil)
	if err != nil {
		return err
	}

	logger.Info("Updated leaderboard - UserID: %s, Rating: %d", userID, rating)
	return nil
}

// UpdateRatings updates player ratings using ELO algorithm
func UpdateRatings(playerA, playerB *UserProfile, scoreA float64) {
	// scoreA: 1.0 if A wins, 0.0 if B wins, 0.5 for draw

	ratingA := float64(playerA.Rating)
	ratingB := float64(playerB.Rating)

	// Calculate expected scores
	expectedA := 1.0 / (1.0 + math.Pow(10, (ratingB-ratingA)/400.0))
	expectedB := 1.0 / (1.0 + math.Pow(10, (ratingA-ratingB)/400.0))

	// Update ratings
	newRatingA := ratingA + KFactor*(scoreA-expectedA)
	newRatingB := ratingB + KFactor*((1.0-scoreA)-expectedB)

	// Ensure ratings don't go below 100
	if newRatingA < 100 {
		newRatingA = 100
	}
	if newRatingB < 100 {
		newRatingB = 100
	}

	playerA.Rating = int(math.Round(newRatingA))
	playerB.Rating = int(math.Round(newRatingB))
}
