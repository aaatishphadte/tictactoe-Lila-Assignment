# Tic-Tac-Toe React Client

A modern React frontend for the multiplayer Tic-Tac-Toe game powered by Nakama.

## Features

- 🎮 Real-time multiplayer gameplay
- 🎨 Modern dark theme UI with teal accents
- 🔐 Device-based authentication
- 🎯 Automatic matchmaking
- 🏆 Live leaderboard
- 📱 Responsive design

## Prerequisites

- Node.js 18+ 
- Nakama server running on `localhost:7350`

## Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at http://localhost:5173

## Building for Production

```bash
npm run build
```

The optimized build will be in the `dist` folder.

## Project Structure

```
client/
├── src/
│   ├── components/         # React components
│   │   ├── LoginScreen.jsx
│   │   ├── MatchmakingScreen.jsx
│   │   ├── GameBoard.jsx
│   │   └── WinnerScreen.jsx
│   ├── services/
│   │   └── nakama.js       # Nakama API integration
│   ├── App.jsx             # Main app with state management
│   ├── main.jsx
│   └── index.css           # Global styles
└── package.json
```

## Configuration

The Nakama server configuration is in `src/services/nakama.js`:

```javascript
const SERVER_KEY = 'defaultkey';
const HOST = 'localhost';
const PORT = '7350';
const USE_SSL = false;
```

Update these values for production deployment.

## Game Flow

1. **Login**: Enter nickname → Device authentication
2. **Matchmaking**: Wait for opponent (usually ~26 seconds)
3. **Game**: Play Tic-Tac-Toe in real-time
4. **Results**: View winner and leaderboard
5. **Repeat**: Play again!

## Technologies

- **React 18** - UI framework
- **Vite** - Build tool and dev server
- **@heroiclabs/nakama-js** - Nakama client SDK
- **CSS3** - Modern styling with animations

## Development

The app uses polling for game updates. In production, this should be replaced with WebSocket connections for better performance.

To enable WebSocket support, update the `App.jsx` to use `nakamaService.connectSocket()` instead of polling.
