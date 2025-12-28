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
                                <th></th>
                                <th>W/L/D</th>
                                <th className="score-col">Score</th>
                            </tr>
                        </thead>
                        <tbody>
                            {leaderboard.map((entry, index) => (
                                <tr key={entry.user_id}>
                                    <td className="rank-name">
                                        <span className="rank">{index + 1}.</span>
                                        <span className="username">{entry.username}</span>
                                    </td>
                                    <td className="stats">
                                        {entry.wins || 0}/{entry.losses || 0}/{entry.draws || 0}
                                    </td>
                                    <td className="score">{entry.score || entry.rating}</td>
                                </tr>
                            ))}
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
