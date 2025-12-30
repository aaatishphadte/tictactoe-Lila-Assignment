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

            // Set opponent name
            const oppId = state.player_x === currentUser.user_id ? state.player_o : state.player_x;
            const opponent = match.presences.find(p => p.user_id === oppId);
            if (opponent) setOpponentName(opponent.username);

            // Match end logic
            if (state.status === 'finished' || matchData.op_code === 5) {
              let winner = 'draw';
              let score = 0;

              if (state.winner === currentUser.user_id) {
                winner = 'you';
                score = 200;
              } else if (state.winner !== "") {
                winner = 'opponent';
                score = -100;
              }

              setGameResult({ winner, score });

              setLeaderboard([
                {
                  user_id: currentUser.user_id,
                  username: currentUser.username,
                  wins: winner === 'you' ? 1 : 0,
                  rating: 1000 + score
                },
                {
                  user_id: oppId,
                  username: opponent ? opponent.username : 'Opponent',
                  wins: winner === 'opponent' ? 1 : 0,
                  rating: 1000 - score / 2
                }
              ]);

              setTimeout(() => setScreen('winner'), 800);
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
