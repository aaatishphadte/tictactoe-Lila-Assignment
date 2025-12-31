# Tic-Tac-Toe Nakama Backend
## Gameplay Demo

<table>
  <tr>
    <th width="50%">Player 1 Perspective</th>
    <th width="50%">Player 2 Perspective</th>
  </tr>
  <tr>
    <td>
      <video src="https://github.com/aaatishphadte/tictactoe-Lila-Assignment/Recodings/Player-1.mp4" controls="controls" muted="muted" style="max-width: 100%;"></video>
    </td>
    <td>
      <video src="https://github.com/aaatishphadte/tictactoe-Lila-Assignment/Recodings/player-2mobile.mp4" controls="controls" muted="muted" style="max-width: 100%;"></video>
    </td>
  </tr>
</table>

A multiplayer Tic-Tac-Toe game backend built with Nakama and Go plugins.

## Features

- **Device-based Authentication**: JWT tokens with 2-hour expiry
- **Server-Authoritative Game Logic**: All moves validated on server
- **Real-time Communication**: WebSocket-based gameplay
- **Matchmaking System**: Casual and ranked modes with ELO-based pairing
- **Leaderboard**: Global rankings with ELO rating system
- **Player Stats**: Track wins, losses, draws, and ratings

## Architecture

```
tictactoe-Lila-Assignment/
├── modules/                    # Go plugin source code
│   ├── main.go                # Plugin entry point
│   ├── auth.go                # Authentication system
│   ├── game_state.go          # Game state and validation
│   ├── game_logic.go          # Game RPCs and logic
│   ├── matchmaking.go         # Matchmaking system
│   ├── leaderboard.go         # ELO ratings and leaderboard
│   └── match_handler.go       # Real-time match handler
├── nakama/                    # Docker configuration
│   ├── docker-compose.yml     # Service definition
│   └── data/                  # Nakama data and modules
└── go.mod                     # Go module definition
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)

## Quick Start

### 1. Build the Plugin

The plugin must be compiled as a Linux shared library (`.so` file) for Nakama to load it.

**Option A: Using Docker (Recommended)**
```bash
# Build using Docker container (ensures Linux compatibility)
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.21 \
  go build -buildmode=plugin -o ./nakama/data/modules/tictactoe.so ./modules/
```

**Option B: Using Make**
```bash
make docker-build
```

### 2. Start Nakama Services

```bash
cd nakama
docker-compose up -d
```

Verify services are running:
```bash
docker ps
```

You should see `nakama` and `postgres` containers running.

### 3. Check Logs

```bash
docker-compose logs -f nakama
```

Look for messages indicating the plugin was loaded successfully:
- "TicTacToe module loaded"
- "Registered RPC: authenticate_device"
- etc.

## API Endpoints

All endpoints use `http://localhost:7350` as the base URL.

### Authentication

**Authenticate Device**
```bash
POST /v2/rpc/authenticate_device
Content-Type: application/json

{
  "device_id": "your-device-id"
}

Response:
{
  "user_id": "...",
  "username": "...",
  "session_token": "...",
  "profile": {
    "wins": 0,
    "losses": 0,
    "draws": 0,
    "rating": 1000
  }
}
```

### Matchmaking

**Join Queue**
```bash
POST /v2/rpc/join_queue
Authorization: Bearer <token>
Content-Type: application/json

{
  "game_mode": "casual"  // or "ranked"
}

Response:
{
  "token": "...",
  "message": "match found" // or "waiting for opponent"
  "match_id": "...",      // if matched
  "matched": true/false
}
```

**Cancel Queue**
```bash
POST /v2/rpc/cancel_queue
Authorization: Bearer <token>
Content-Type: application/json

{
  "token": "your-queue-token"
}
```

### Game Play

**Make Move**
```bash
POST /v2/rpc/make_move
Authorization: Bearer <token>
Content-Type: application/json

{
  "match_id": "...",
  "row": 0,     // 0-2
  "col": 0      // 0-2
}

Response:
{
  "success": true,
  "game_state": {
    "board": [["X","",""], ["","",""], ["","",""]],
    "current_player": "O",
    "status": "active",
    ...
  },
  "message": "move successful"
}
```

**Get Game State**
```bash
POST /v2/rpc/get_game_state
Authorization: Bearer <token>
Content-Type: application/json

{
  "match_id": "..."
}
```

**Resign Game**
```bash
POST /v2/rpc/resign_game
Authorization: Bearer <token>
Content-Type: application/json

{
  "match_id": "..."
}
```

### Leaderboard

**Get Leaderboard**
```bash
GET /v2/rpc/get_leaderboard
Authorization: Bearer <token>

Response:
{
  "entries": [
    {
      "user_id": "...",
      "username": "...",
      "rank": 1,
      "score": 1250,
      "num_score": 10
    },
    ...
  ]
}
```

**Get Player Rank**
```bash
POST /v2/rpc/get_player_rank
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "..."  // optional, defaults to authenticated user
}

Response:
{
  "user_id": "...",
  "username": "...",
  "rank": 42,
  "rating": 1180,
  "wins": 15,
  "losses": 10,
  "draws": 3
}
```

## WebSocket Real-time Gameplay

For real-time gameplay, connect to `ws://localhost:7350/ws` using Nakama's WebSocket protocol.

**Operation Codes:**
- `1` - Move
- `2` - Game State Update
- `3` - Player Joined
- `4` - Player Left
- `5` - Game Over

## Development

### Project Structure

- **modules/main.go**: Plugin initialization and RPC registration
- **modules/auth.go**: Device authentication and user profiles
- **modules/game_state.go**: Game state structure and validation logic
- **modules/game_logic.go**: RPC handlers for game operations
- **modules/matchmaking.go**: Player queue and matching system
- **modules/leaderboard.go**: ELO rating calculation and leaderboard
- **modules/match_handler.go**: Real-time WebSocket match handler

### Making Changes

1. Edit the Go source files in `modules/`
2. Rebuild the plugin: `make docker-build`
3. Restart Nakama: `cd nakama && docker-compose restart nakama`
4. Check logs: `docker-compose logs -f nakama`

### Testing

**Nakama Console**: Access at http://localhost:7351
- Default credentials: admin / password
- View users, storage, leaderboards, and more

**API Testing**: Use curl, Postman, or any HTTP client

## Deployment

### Google Cloud Deployment (Guide)

1. **Build for Linux**: Use Docker build to ensure Linux compatibility
2. **Container Registry**: Push Nakama image with plugin to Google Container Registry
3. **Cloud Run / GKE**: Deploy Nakama container
4. **Cloud SQL**: Use PostgreSQL instance
5. **Environment Variables**: Configure database connection
6. **Networking**: Set up load balancer and SSL

See Google Cloud documentation for detailed deployment steps.

## Troubleshooting

**Plugin not loading:**
- Check that `.so` file exists in `nakama/data/modules/`
- Ensure plugin was built for Linux (use Docker build)
- Check Nakama logs for error messages

**Port conflicts:**
- Ensure ports 5432, 7349, 7350, 7351 are available
- No local PostgreSQL should be running on 5432

**Connection errors:**
- Verify Docker services are running: `docker ps`
- Check logs: `docker-compose logs -f`

## License

See LICENSE file for details.
