import { useState, useEffect } from 'react';
import './GameBoard.css';

export default function GameBoard({
    gameState,
    onMove,
    playerSymbol,
    playerName,
    opponentName
}) {
    const [selectedCell, setSelectedCell] = useState(null);

    const handleCellClick = (row, col) => {
        const index = row * 3 + col;

        if (gameState.board[index] !== '' ||
            gameState.currentPlayer !== playerSymbol ||
            gameState.status !== 'active') {
            return;
        }

        setSelectedCell({ row, col });
        onMove(row, col);
    };

    const isMyTurn = gameState.currentPlayer === playerSymbol;

    const renderCell = (index) => {
        const row = Math.floor(index / 3);
        const col = index % 3;
        const value = gameState.board[index];

        return (
            <button
                key={index}
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
                        {gameState.currentPlayer}
                    </div>
                    <span className="turn-text">Turn</span>
                </div>

                {/* Game Board */}
                <div className="board">
                    {[0, 1, 2].map(row => (
                        <div key={row} className="board-row">
                            {[0, 1, 2].map(col => {
                                const index = row * 3 + col;
                                return renderCell(index);
                            })}
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
