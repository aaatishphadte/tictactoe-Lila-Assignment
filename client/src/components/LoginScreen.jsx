import { useState } from 'react';
import './LoginScreen.css';

export default function LoginScreen({ onLogin }) {
    const [nickname, setNickname] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!nickname.trim()) {
            setError('Please enter a nickname');
            return;
        }

        setLoading(true);
        setError('');

        try {
            await onLogin(nickname.trim());
        } catch (err) {
            setError(err.message || 'Authentication failed');
            setLoading(false);
        }
    };

    return (
        <div className="login-screen fade-in">
            <div className="login-card">
                <div className="login-header">
                    <h1>Tic-Tac-Toe</h1>
                    <p>Who are you?</p>
                </div>

                <form onSubmit={handleSubmit} className="login-form">
                    <input
                        type="text"
                        placeholder="Nickname"
                        value={nickname}
                        onChange={(e) => setNickname(e.target.value)}
                        maxLength={20}
                        disabled={loading}
                        className="nickname-input"
                        autoFocus
                    />

                    {error && <div className="error-message">{error}</div>}

                    <button
                        type="submit"
                        disabled={loading || !nickname.trim()}
                        className="continue-button"
                    >
                        {loading ? 'Connecting...' : 'Continue'}
                    </button>
                </form>
            </div>
        </div>
    );
}
