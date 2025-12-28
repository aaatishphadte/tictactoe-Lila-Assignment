package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

// AuthenticateDeviceRequest represents the device authentication request
type AuthenticateDeviceRequest struct {
	DeviceID string `json:"device_id"`
}

// AuthenticateDeviceResponse represents the authentication response
type AuthenticateDeviceResponse struct {
	UserID       string      `json:"user_id"`
	Username     string      `json:"username"`
	SessionToken string      `json:"session_token"`
	Profile      UserProfile `json:"profile"`
}

// UserProfile represents user game statistics
type UserProfile struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
	Rating int `json:"rating"`
}

// RpcAuthenticateDevice handles device-based authentication
func RpcAuthenticateDevice(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request AuthenticateDeviceRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		return "", runtime.NewError("invalid request payload", 3) // INVALID_ARGUMENT
	}

	if request.DeviceID == "" {
		return "", runtime.NewError("device_id is required", 3)
	}

	// Authenticate or create user with device ID
	userID, username, created, err := nk.AuthenticateDevice(ctx, request.DeviceID, "", true)
	if err != nil {
		logger.Error("Failed to authenticate device: %v", err)
		return "", runtime.NewError("authentication failed", 13) // INTERNAL
	}

	logger.Info("Device authenticated - UserID: %s, Username: %s, Created: %v", userID, username, created)

	// Initialize user profile if newly created
	if created {
		profile := UserProfile{
			Wins:   0,
			Losses: 0,
			Draws:  0,
			Rating: 1000, // Default ELO rating
		}

		// Store user profile in storage
		profileData, err := json.Marshal(profile)
		if err != nil {
			logger.Error("Failed to marshal profile: %v", err)
			return "", runtime.NewError("failed to create profile", 13)
		}

		writes := []*runtime.StorageWrite{
			{
				Collection:      "profiles",
				Key:             userID,
				UserID:          userID,
				Value:           string(profileData),
				PermissionRead:  2, // Owner read
				PermissionWrite: 0, // No client write
			},
		}

		if _, err := nk.StorageWrite(ctx, writes); err != nil {
			logger.Error("Failed to write profile: %v", err)
			return "", runtime.NewError("failed to save profile", 13)
		}

		logger.Info("Created new profile for user: %s", userID)
	}

	// Retrieve user profile
	profile, err := GetUserProfile(ctx, logger, nk, userID)
	if err != nil {
		logger.Error("Failed to get user profile: %v", err)
		return "", runtime.NewError("failed to retrieve profile", 13)
	}

	// Generate session token
	token, _, err := nk.AuthenticateTokenGenerate(userID, username, 0, nil)
	if err != nil {
		logger.Error("Failed to generate token: %v", err)
		return "", runtime.NewError("failed to generate session token", 13)
	}

	response := AuthenticateDeviceResponse{
		UserID:       userID,
		Username:     username,
		SessionToken: token,
		Profile:      profile,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response: %v", err)
		return "", runtime.NewError("failed to create response", 13)
	}

	return string(responseJSON), nil
}

// GetUserProfile retrieves a user's profile from storage
func GetUserProfile(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string) (UserProfile, error) {
	var profile UserProfile

	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: "profiles",
			Key:        userID,
			UserID:     userID,
		},
	})

	if err != nil {
		return profile, err
	}

	if len(objects) == 0 {
		// Return default profile if not found
		return UserProfile{
			Wins:   0,
			Losses: 0,
			Draws:  0,
			Rating: 1000,
		}, nil
	}

	if err := json.Unmarshal([]byte(objects[0].Value), &profile); err != nil {
		return profile, err
	}

	return profile, nil
}

// UpdateUserProfile updates a user's profile in storage
func UpdateUserProfile(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string, profile UserProfile) error {
	profileData, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	writes := []*runtime.StorageWrite{
		{
			Collection:      "profiles",
			Key:             userID,
			UserID:          userID,
			Value:           string(profileData),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	}

	if _, err := nk.StorageWrite(ctx, writes); err != nil {
		return err
	}

	return nil
}
