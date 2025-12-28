import { useState, useEffect } from 'react';
import './MatchmakingScreen.css';

export default function MatchmakingScreen({ onCancel, onMatchFound }) {
    const [dots, setDots] = useState('');
    const [countdown, setCountdown] = useState(30);

    useEffect(() => {
        const interval = setInterval(() => {
            setDots(prev => prev.length >= 3 ? '' : prev + '.');
        }, 500);

        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        const timer = setInterval(() => {
            setCountdown(prev => prev > 0 ? prev - 1 : 0);
        }, 1000);

        return () => clearInterval(timer);
    }, []);

    return (
        <div className="matchmaking-screen fade-in">
            <div className="matchmaking-card">
                <div className="spinner-container">
                    <div className="spinner"></div>
                </div>

                <h2 className="matchmaking-title">
                    Finding a random player{dots}
                </h2>

                <p className="matchmaking-subtitle">
                    {countdown > 0 ? `Waiting... ${countdown}s` : 'Still searching...'}
                </p>

                <button onClick={onCancel} className="cancel-button">
                    Cancel
                </button>
            </div>
        </div>
    );
}
