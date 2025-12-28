package main

import (
	"encoding/json"
	"fmt"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

// GameResult represents the outcome of a finished game
type GameResult string

const (
	GameResultXWins GameResult = "x_wins"
	GameResultOWins GameResult = "o_wins"
	GameResultDraw  GameResult = "draw"
	GameResultNone  GameResult = "none"
)

// PlayerSymbol represents X or O
type PlayerSymbol string

const (
	SymbolX     PlayerSymbol = "X"
	SymbolO     PlayerSymbol = "O"
	SymbolEmpty PlayerSymbol = ""
)

// GameState represents the complete state of a Tic-Tac-Toe game
type GameState struct {
	MatchID       string             `json:"match_id"`
	Board         [3][3]PlayerSymbol `json:"board"`
	CurrentPlayer PlayerSymbol       `json:"current_player"`
	PlayerX       string             `json:"player_x"`
	PlayerO       string             `json:"player_o"`
	Status        GameStatus         `json:"status"`
	Result        GameResult         `json:"result"`
	Winner        string             `json:"winner"`
	MoveCount     int                `json:"move_count"`
	GameMode      string             `json:"game_mode"` // "casual" or "ranked"
}

// Move represents a player's move
type Move struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// NewGameState creates a new game state
func NewGameState(matchID, playerX, playerO, gameMode string) *GameState {
	return &GameState{
		MatchID:       matchID,
		Board:         [3][3]PlayerSymbol{},
		CurrentPlayer: SymbolX, // X always starts
		PlayerX:       playerX,
		PlayerO:       playerO,
		Status:        GameStatusActive,
		Result:        GameResultNone,
		Winner:        "",
		MoveCount:     0,
		GameMode:      gameMode,
	}
}

// ValidateMove checks if a move is legal
func (gs *GameState) ValidateMove(row, col int, playerID string) error {
	// Check if game is active
	if gs.Status != GameStatusActive {
		return fmt.Errorf("game is not active")
	}

	// Check if it's the player's turn
	if gs.CurrentPlayer == SymbolX && playerID != gs.PlayerX {
		return fmt.Errorf("not your turn")
	}
	if gs.CurrentPlayer == SymbolO && playerID != gs.PlayerO {
		return fmt.Errorf("not your turn")
	}

	// Check if move is within bounds
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return fmt.Errorf("move out of bounds")
	}

	// Check if cell is empty
	if gs.Board[row][col] != SymbolEmpty {
		return fmt.Errorf("cell already occupied")
	}

	return nil
}

// ApplyMove applies a move to the game state
func (gs *GameState) ApplyMove(row, col int, playerID string) error {
	if err := gs.ValidateMove(row, col, playerID); err != nil {
		return err
	}

	// Place the symbol
	gs.Board[row][col] = gs.CurrentPlayer
	gs.MoveCount++

	// Check for win
	if gs.CheckWin() {
		gs.Status = GameStatusFinished
		gs.Winner = playerID
		if gs.CurrentPlayer == SymbolX {
			gs.Result = GameResultXWins
		} else {
			gs.Result = GameResultOWins
		}
		return nil
	}

	// Check for draw
	if gs.CheckDraw() {
		gs.Status = GameStatusFinished
		gs.Result = GameResultDraw
		return nil
	}

	// Switch player
	if gs.CurrentPlayer == SymbolX {
		gs.CurrentPlayer = SymbolO
	} else {
		gs.CurrentPlayer = SymbolX
	}

	return nil
}

// CheckWin checks if the current player has won
func (gs *GameState) CheckWin() bool {
	symbol := gs.CurrentPlayer

	// Check rows
	for i := 0; i < 3; i++ {
		if gs.Board[i][0] == symbol && gs.Board[i][1] == symbol && gs.Board[i][2] == symbol {
			return true
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if gs.Board[0][i] == symbol && gs.Board[1][i] == symbol && gs.Board[2][i] == symbol {
			return true
		}
	}

	// Check diagonals
	if gs.Board[0][0] == symbol && gs.Board[1][1] == symbol && gs.Board[2][2] == symbol {
		return true
	}
	if gs.Board[0][2] == symbol && gs.Board[1][1] == symbol && gs.Board[2][0] == symbol {
		return true
	}

	return false
}

// CheckDraw checks if the game is a draw
func (gs *GameState) CheckDraw() bool {
	// All cells must be filled
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if gs.Board[i][j] == SymbolEmpty {
				return false
			}
		}
	}
	return true
}

// ToJSON converts game state to JSON string
func (gs *GameState) ToJSON() (string, error) {
	data, err := json.Marshal(gs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GameStateFromJSON creates a GameState from JSON string
func GameStateFromJSON(data string) (*GameState, error) {
	var gs GameState
	if err := json.Unmarshal([]byte(data), &gs); err != nil {
		return nil, err
	}
	return &gs, nil
}
