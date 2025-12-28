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
      // Generate a device ID (in production, store this persistently)
      const deviceId = localStorage.getItem('deviceId') ||
        `device_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

      localStorage.setItem('deviceId', deviceId);
      localStorage.setItem('nickname', nickname);

      const response = await nakamaService.authenticateDevice(deviceId, nickname);

      const user = {
        user_id: response.user_id,
        username: nickname,
        profile: response.profile
      };

      setUser(user);
      setUserData(user);

      // Start matchmaking
      setScreen('matchmaking');
      startMatchmaking();
    } catch (error) {
      console.error('Login failed:', error);
      throw new Error('Failed to authenticate. Please try again.');
    }
  };

  // Matchmaking
  const startMatchmaking = async () => {
    try {
      setScreen('matchmaking');

      // Start matchmaking with built-in matchmaker
      await nakamaService.startMatchmaking(
        // onMatched callback
        async (matched) => {
          console.log('Matchmaker matched! Full response:', matched);
          console.log('FULL JSON:', JSON.stringify(matched, null, 2));
          console.log('Match ID:', matched.match_id);
          console.log('Token:', matched.token);

          // IMPORTANT: Both players receive the SAME match_id from matchmaker
          if (matched.match_id) {
            console.log('Joining match from matchmaker:', matched.match_id);
            await joinMatch(matched.match_id);
          } else if (matched.token && matched.users) {
            // Coordinate match creation - only first player creates
            const userIds = matched.users.map(u => u.presence.user_id).sort();
            const isFirst = userIds[0] === matched.self.user_id;
            const storageKey = `match_${matched.token}`;

            console.log('Coordinating match creation. Am I first?', isFirst);

            if (isFirst) {
              // First player creates and stores match_id
              const newMatch = await nakamaService.socket.createMatch();
              await nakamaService.client.writeStorageObjects(nakamaService.session, [{
                collection: 'matchmaking',
                key: storageKey,
                value: { match_id: newMatch.match_id },
                permission_read: 2
              }]);
              console.log('Created and stored match:', newMatch.match_id);
              await joinMatch(newMatch.match_id);
            } else {
              // Second player reads match_id from storage
              let matchId = null;
              for (let i = 0; i < 10 && !matchId; i++) {
                try {
                  const result = await nakamaService.client.readStorageObjects(nakamaService.session, {
                    object_ids: [{ collection: 'matchmaking', key: storageKey }]
                  });
                  if (result.objects?.[0]) {
                    matchId = result.objects[0].value.match_id;
                  }
                } catch (e) { }
                if (!matchId) await new Promise(r => setTimeout(r, 500));
              }

              if (matchId) {
                console.log('Found match from storage:', matchId);
                await joinMatch(matchId);
              } else {
                setError('Failed to join match');
                setScreen('login');
              }
            }
          } else {
            console.error('No match_id or token in response!', matched);
            setError('Matchmaking failed');
            setScreen('login');
          }
        },
        // onError callback
        (error) => {
          console.error('Matchmaking error:', error);
          setError('Matchmaking failed. Please try again.');
          setScreen('login');
        }
      );
    } catch (error) {
      console.error('Matchmaking failed:', error);
      setError('Failed to start matchmaking. Please try again.');
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
  const joinMatch = async (matchId) => {
    try {
      console.log('Joining match:', matchId);
      const match = await nakamaService.joinMatch(matchId);

      console.log('✅ Match joined!');
      console.log('Presences in match:', match.presences);
      console.log('My user:', user);

      setMatchId(matchId);

      // Determine player symbol based on when they joined
      // First player to join is X
      const iAmX = match.presences.length === 1 ||
        (match.presences[0] && user && match.presences[0].user_id === user.user_id);
      const mySymbol = iAmX ? 'X' : 'O';
      setPlayerSymbol(mySymbol);

      // Find opponent
      const opponent = user ? match.presences?.find(p => p.user_id !== user.user_id) : null;

      if (opponent) {
        setOpponentName(opponent.username || `Player`);
      } else {
        // Opponent not here yet, listen for when they join
        nakamaService.socket.onmatchpresence = (event) => {
          console.log('Presence change:', event);
          if (event.joins && event.joins.length > 0) {
            setOpponentName(event.joins[0].username || 'Player');
          }
        };
      }

      // Initialize game state
      const initialState = {
        board: Array(9).fill(''),
        currentPlayer: 'X',
        status: 'active',
        winner: null
      };
      setGameState(initialState);

      setScreen('game');

      // Set up match data listener for opponent moves
      nakamaService.socket.onmatchdata = (matchData) => {
        try {
          console.log('Received match data:', matchData);
          const decoder = new TextDecoder();
          const dataStr = decoder.decode(matchData.data);
          const data = JSON.parse(dataStr);

          console.log('Parsed match data:', data);

          if (data.board) {
            // Update game state with opponent's move
            setGameState(data);
          }
        } catch (error) {
          console.error('Failed to parse match data:', error);
        }
      };
    } catch (error) {
      console.error('Failed to join match:', error);
      setError('Failed to join match');
      setScreen('login');
    }
  };

  const handleMove = async (row, col) => {
    if (!gameState || gameState.status !== 'active') {
      console.log('Game not active');
      return;
    }
    if (gameState.currentPlayer !== playerSymbol) {
      console.log('Not your turn');
      return;
    }

    const index = row * 3 + col;
    if (gameState.board[index]) {
      console.log('Cell already taken');
      return; // Cell already taken
    }

    console.log(`Making move at row=${row}, col=${col}, index=${index}`);

    // Make move locally
    const newBoard = [...gameState.board];
    newBoard[index] = playerSymbol;

    // Check for win or draw
    const winner = checkWinner(newBoard);
    const isDraw = !winner && newBoard.every(cell => cell !== '');

    const newState = {
      board: newBoard,
      currentPlayer: playerSymbol === 'X' ? 'O' : 'X',
      status: winner || isDraw ? 'finished' : 'active',
      winner: winner
    };

    console.log('New game state:', newState);
    setGameState(newState);

    // Send move to opponent via WebSocket
    try {
      if (!nakamaService.socket) {
        console.error('Socket not connected!');
        return;
      }

      if (!matchId) {
        console.error('No match ID!');
        return;
      }

      console.log('Sending match state to opponent...');
      await nakamaService.socket.sendMatchState(
        matchId,
        1, // op code for move
        JSON.stringify(newState)
      );
      console.log('Move sent successfully');
    } catch (error) {
      console.error('Failed to send move:', error);
    }

    // Check for game over
    if (winner || isDraw) {
      await handleGameOver(newState);
    }
  };

  const checkWinner = (board) => {
    const lines = [
      [0, 1, 2], [3, 4, 5], [6, 7, 8], // rows
      [0, 3, 6], [1, 4, 7], [2, 5, 8], // columns
      [0, 4, 8], [2, 4, 6] // diagonals
    ];

    for (const [a, b, c] of lines) {
      if (board[a] && board[a] === board[b] && board[a] === board[c]) {
        return board[a];
      }
    }
    return null;
  };

  const handleGameOver = async (finalState) => {
    let winner = 'draw';
    let score = 0;

    if (finalState.winner) {
      if (finalState.winner === playerSymbol) {
        winner = 'you';
        score = 200;
      } else {
        winner = 'opponent';
        score = -100;
      }
    }

    setGameResult({ winner, score });

    // Mock leaderboard for now (would use Nakama's leaderboard API in production)
    setLeaderboard([
      { username: userData.username || 'You', wins: 10, losses: 2, draws: 1, score: 2100 },
      { username: opponentName, wins: 2, losses: 10, draws: 1, score: 500 }
    ]);

    setScreen('winner');
  };

  const handlePlayAgain = () => {
    // Reset game state
    setGameState(null);
    setPlayerSymbol(null);
    setGameResult(null);
    setLeaderboard([]);

    // Start new matchmaking
    setScreen('matchmaking');
    startMatchmaking();
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
    </div>
  );
}

export default App;
