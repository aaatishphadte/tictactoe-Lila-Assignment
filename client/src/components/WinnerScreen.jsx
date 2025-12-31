import './WinnerScreen.css';

export default function WinnerScreen({
    winner,
    playerScore,
    leaderboard,
    onPlayAgain
}) {
    const isWinner = winner === 'you';
    const isDraw = winner === 'draw';

    return (
        <div className="winner-screen fade-in">
            <div className="winner-card">
                {/* Winner Announcement */}
                <div className={`winner-announcement ${isWinner ? 'win' : isDraw ? 'draw' : 'lose'}`}>
                    {isDraw ? (
                        <div className="draw-icon">⚔️</div>
                    ) : (
                        <div className="winner-symbol">
                            {isWinner ? '✓' : '✗'}
                        </div>
                    )}

                    <h1 className="winner-title">
                        {isDraw ? 'DRAW!' : isWinner ? 'WINNER!' : 'GAME OVER'}
                    </h1>

                    {!isDraw && (
                        <div className="points">
                            {isWinner ? '+' : ''}{playerScore} pts
                        </div>
                    )}
                </div>

                {/* Leaderboard */}
                <div className="leaderboard-section">
                    <div className="leaderboard-header">
                        <div className="trophy-icon">🏆</div>
                        <span>Leaderboard</span>
                    </div>

                    <table className="leaderboard-table">
                        <thead>
                            <tr>
                                <th>RANK</th>
                                <th>PLAYER</th>
                                <th className="score-col">SCORE</th>
                            </tr>
                        </thead>
                        <tbody>
                            {leaderboard.length === 0 ? (
                                <tr>
                                    <td colSpan="3" style={{ textAlign: 'center', padding: '20px', opacity: 0.5 }}>
                                        No leaderboard data yet
                                    </td>
                                </tr>
                            ) : (
                                leaderboard.slice(0, 10).map((entry, index) => (
                                    <tr key={entry.user_id || index}>
                                        <td className="rank">{entry.rank || (index + 1)}</td>
                                        <td className="username">{entry.username || 'Unknown'}</td>
                                        <td className="score">{entry.rating || entry.score || 1000}</td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Play Again Button */}
                <button onClick={onPlayAgain} className="play-again-button">
                    Play Again
                </button>
            </div>
        </div>
    );
}
