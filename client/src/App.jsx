import { useState, useEffect } from 'react';
import LoginScreen from './components/LoginScreen';
import MatchmakingScreen from './components/MatchmakingScreen';
import GameBoard from './components/GameBoard';
import WinnerScreen from './components/WinnerScreen';
import nakamaService from './services/nakama';
import './App.css';

function App() {
  const [screen, setScreen] = useState('login'); // login, matchmaking, game, winner
  const [user, setUser] = useState(null);
  const [userData, setUserData] = useState(null);
  const [gameState, setGameState] = useState(null);
  const [matchId, setMatchId] = useState(null);
  const [playerSymbol, setPlayerSymbol] = useState(null);
  const [opponentName, setOpponentName] = useState('Opponent');
  const [gameResult, setGameResult] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [error, setError] = useState('');

  // Login Handler
  const handleLogin = async (nickname) => {
    try {
      const deviceId = nickname;
      localStorage.setItem('deviceId', deviceId);
      localStorage.setItem('nickname', nickname);

      const response = await nakamaService.authenticateDevice(deviceId, nickname);

      const authenticatedUser = {
        user_id: response.user_id,
        username: nickname,
        profile: response.profile
      };

      setUser(authenticatedUser);
      setUserData(authenticatedUser);

      // Start matchmaking
      startMatchmaking(authenticatedUser);
    } catch (error) {
      console.error('Login failed:', error);
      throw new Error('Failed to authenticate. Please try again.');
    }
  };

  // Matchmaking
  const startMatchmaking = async (currentUser) => {
    try {
      setScreen('matchmaking');

      await nakamaService.startMatchmaking(
        async (matched) => {
          console.log('Matchmaker matched!', matched);
          if (matched.match_id) {
            await joinMatch(matched.match_id, currentUser);
          } else {
            setError('Matchmaking failed: No match ID');
            setScreen('login');
          }
        },
        (error) => {
          console.error('Matchmaking error:', error);
          setError('Matchmaking failed. Please try again.');
          setScreen('login');
        }
      );
    } catch (error) {
      console.error('Matchmaking failed:', error);
      setScreen('login');
    }
  };

  const handleCancelMatchmaking = async () => {
    try {
      await nakamaService.cancelMatchmaking();
      setScreen('login');
    } catch (error) {
      console.error('Cancel failed:', error);
    }
  };

  // Game
  const joinMatch = async (id, currentUser) => {
    try {
      console.log('Joining match:', id);
      const match = await nakamaService.joinMatch(id);
      setMatchId(id);

      // Store match in a variable that the callback can access
      const matchRef = match;

      // Initial state
      setGameState({
        board: [['', '', ''], ['', '', ''], ['', '', '']],
        current_player: 'X',
        status: 'active'
      });

      setScreen('game');

      nakamaService.socket.onmatchdata = (matchData) => {
        try {
          const decoder = new TextDecoder();
          const envelope = JSON.parse(decoder.decode(matchData.data));

          console.log('--- Match Data --- OpCode:', matchData.op_code);

          // OpCode 2 = Game State, OpCode 5 = Game Over
          if (matchData.op_code === 2 || matchData.op_code === 5) {
            const state = typeof envelope.data === 'string'
              ? JSON.parse(envelope.data)
              : envelope.data;

            if (!state || !state.board) return;

            setGameState(state);

            // Authoritative symbol
            if (state.player_x === currentUser.user_id) {
              setPlayerSymbol('X');
            } else if (state.player_o === currentUser.user_id) {
              setPlayerSymbol('O');
            }

            // Set opponent name - FIXED: Check if presences exists
            const oppId = state.player_x === currentUser.user_id ? state.player_o : state.player_x;

            if (matchRef && matchRef.presences && matchRef.presences.length > 0) {
              const opponent = matchRef.presences.find(p => p.user_id === oppId);

              if (opponent && opponent.username) {
                console.log('Setting opponent name to:', opponent.username);
                setOpponentName(opponent.username);
              } else {
                // Fallback: try to find ANY other player
                const others = matchRef.presences.filter(p => p.user_id !== currentUser.user_id);
                if (others.length > 0 && others[0].username) {
                  console.log('Setting opponent name (fallback) to:', others[0].username);
                  setOpponentName(others[0].username);
                }
              }
            } else {
              console.warn('Match presences not available yet');
            }

            // Match end logic
            if (state.status === 'finished' || matchData.op_code === 5) {
              console.log('🏁 GAME OVER detected! OpCode:', matchData.op_code, 'Status:', state.status);
              console.log('Winner:', state.winner, 'Current User:', currentUser.user_id);
              console.log('Rating changes:', state.rating_change_x, state.rating_change_o);

              let winner = 'draw';
              let score = 0;

              // Determine which player we are (X or O) to get the correct rating change
              const isPlayerX = state.player_x === currentUser.user_id;
              const actualRatingChange = isPlayerX ? state.rating_change_x : state.rating_change_o;

              if (state.winner === currentUser.user_id) {
                winner = 'you';
                score = actualRatingChange || 0;  // Use real ELO change
              } else if (state.winner !== "") {
                winner = 'opponent';
                score = actualRatingChange || 0;  // Use real ELO change (will be negative)
              } else {
                // Draw
                score = actualRatingChange || 0;  // Use real ELO change (typically small)
              }

              console.log('Setting game result:', { winner, score });
              setGameResult({ winner, score });

              // Fetch REAL leaderboard from backend
              console.log('Fetching real leaderboard from Nakama...');
              nakamaService.getLeaderboard()
                .then(leaderboardData => {
                  console.log('Received leaderboard data:', leaderboardData);
                  if (leaderboardData && leaderboardData.length > 0) {
                    setLeaderboard(leaderboardData);
                  } else {
                    // Fallback to minimal data if backend returns empty
                    console.warn('Leaderboard is empty, showing empty state');
                    setLeaderboard([]);
                  }
                })
                .catch(error => {
                  console.error('Failed to fetch leaderboard:', error);
                  setLeaderboard([]);
                });

              console.log('Transitioning to winner screen in 800ms...');
              setTimeout(() => {
                console.log('NOW transitioning to winner screen');
                setScreen('winner');
              }, 800);
            }
          }
        } catch (error) {
          console.error('Failed to parse match data:', error);
        }
      };
    } catch (error) {
      console.error('Failed to join match:', error);
      setScreen('login');
    }
  };

  const handleMove = async (row, col) => {
    if (!gameState || gameState.status !== 'active') return;
    if (gameState.current_player !== playerSymbol) return;
    if (gameState.board[row][col] !== '') return;

    try {
      await nakamaService.socket.sendMatchState(
        matchId,
        1, // Move OpCode
        JSON.stringify({ row, col })
      );
    } catch (error) {
      console.error('Failed to send move:', error);
    }
  };

  const handlePlayAgain = () => {
    setGameState(null);
    setPlayerSymbol(null);
    setGameResult(null);
    setLeaderboard([]);
    startMatchmaking(user);
  };

  return (
    <div className="app">
      {screen === 'login' && (
        <LoginScreen onLogin={handleLogin} />
      )}

      {screen === 'matchmaking' && (
        <MatchmakingScreen
          onCancel={handleCancelMatchmaking}
        />
      )}

      {screen === 'game' && gameState && (
        <GameBoard
          gameState={gameState}
          onMove={handleMove}
          playerSymbol={playerSymbol}
          playerName={userData?.username || 'You'}
          opponentName={opponentName}
        />
      )}

      {screen === 'winner' && gameResult && (
        <WinnerScreen
          winner={gameResult.winner}
          playerScore={gameResult.score}
          leaderboard={leaderboard}
          onPlayAgain={handlePlayAgain}
        />
      )}

      {error && (
        <div className="error-toast">
          {error}
          <button onClick={() => setError('')}>×</button>
        </div>
      )}
    </div>
  );
}

export default App;
