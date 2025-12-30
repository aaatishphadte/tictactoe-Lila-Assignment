import { useState, useEffect } from 'react';
import './GameBoard.css';

export default function GameBoard({
    gameState,
    onMove,
    playerSymbol,
    playerName,
    opponentName
}) {

    const handleCellClick = (row, col) => {
        if (gameState.board[row][col] !== '' ||
            gameState.current_player !== playerSymbol ||
            gameState.status !== 'active') {
            return;
        }

        onMove(row, col);
    };

    const isMyTurn = gameState.current_player === playerSymbol;

    const renderCell = (row, col) => {
        const value = gameState.board[row][col];

        return (
            <button
                key={`${row}-${col}`}
                className={`cell ${value} ${value ? 'filled' : ''}`}
                onClick={() => handleCellClick(row, col)}
                disabled={!isMyTurn || value !== ''}
            >
                {value && (
                    <span className={`symbol ${value.toLowerCase()}`}>
                        {value}
                    </span>
                )}
            </button>
        );
    };

    return (
        <div className="game-board fade-in">
            <div className="game-card">
                {/* Player Info */}
                <div className="players-info">
                    <div className={`player ${playerSymbol === 'X' ? 'active' : ''}`}>
                        <span className="player-name">{playerName}</span>
                        <span className="player-label">(you)</span>
                    </div>

                    <div className={`player ${playerSymbol === 'O' ? 'active' : ''}`}>
                        <span className="player-name">{opponentName}</span>
                        <span className="player-label">(opp)</span>
                    </div>
                </div>

                {/* Turn Indicator */}
                <div className="turn-indicator">
                    <div className={`turn-symbol ${isMyTurn ? 'my-turn' : ''}`}>
                        {gameState.current_player}
                    </div>
                    <span className="turn-text">Turn</span>
                </div>

                {/* Game Board */}
                <div className="board">
                    {[0, 1, 2].map(row => (
                        <div key={row} className="board-row">
                            {[0, 1, 2].map(col => renderCell(row, col))}
                        </div>
                    ))}
                </div>

                {/* Status Message */}
                {gameState.status === 'active' && (
                    <div className="status-message">
                        {isMyTurn ? "Your turn!" : "Opponent's turn..."}
                    </div>
                )}

                {/* Leave Room Button */}
                <button className="leave-button">
                    Leave room (?)
                </button>
            </div>
        </div>
    );
}
