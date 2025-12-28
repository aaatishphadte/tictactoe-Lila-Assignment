# Nakama Backend API Documentation

## Base URL
`http://localhost:7350`

Production: Replace with your deployed Nakama server URL

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <session_token>
```

---

## Endpoints

### 1. Device Authentication

**Endpoint:** `POST /v2/rpc/authenticate_device`

**Description:** Authenticate a device and create/retrieve user account.

**Request Body:**
```json
{
  "device_id": "string (required)"
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "username": "string",
  "session_token": "jwt_token",
  "profile": {
    "wins": 0,
    "losses": 0,
    "draws": 0,
    "rating": 1000
  }
}
```

**Errors:**
- `3 (INVALID_ARGUMENT)`: device_id is required
- `13 (INTERNAL)`: Authentication or profile creation failed

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/authenticate_device \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device-123"}'
```

---

### 2. Join Matchmaking Queue

**Endpoint:** `POST /v2/rpc/join_queue`

**Description:** Join matchmaking queue for a game.

**Authentication:** Required

**Request Body:**
```json
{
  "game_mode": "casual|ranked"
}
```

**Response (Waiting):**
```json
{
  "token": "uuid",
  "message": "waiting for opponent",
  "matched": false
}
```

**Response (Matched):**
```json
{
  "token": "uuid",
  "message": "match found",
  "match_id": "uuid",
  "matched": true
}
```

**Errors:**
- `16 (UNAUTHENTICATED)`: User not authenticated
- `3 (INVALID_ARGUMENT)`: Invalid game_mode
- `13 (INTERNAL)`: Matchmaking failed

**Game Modes:**
- `casual`: No rating restrictions, instant matching
- `ranked`: Rating-based matching (±200 rating difference)

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/join_queue \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"game_mode": "ranked"}'
```

---

### 3. Cancel Matchmaking Queue

**Endpoint:** `POST /v2/rpc/cancel_queue`

**Description:** Leave matchmaking queue.

**Authentication:** Required

**Request Body:**
```json
{
  "token": "string (matchmaking token)"
}
```

**Response:**
```json
{
  "success": true,
  "message": "removed from queue"
}
```

**Errors:**
- `16 (UNAUTHENTICATED)`: User not authenticated
- `3 (INVALID_ARGUMENT)`: Token is required
- `13 (INTERNAL)`: Failed to remove from queue

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/cancel_queue \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"token": "matchmaking-token-uuid"}'
```

---

### 4. Make Move

**Endpoint:** `POST /v2/rpc/make_move`

**Description:** Make a move in an active game.

**Authentication:** Required

**Request Body:**
```json
{
  "match_id": "string (uuid)",
  "row": 0,  // 0-2
  "col": 0   // 0-2
}
```

**Response (Success):**
```json
{
  "success": true,
  "game_state": {
    "match_id": "uuid",
    "board": [
      ["X", "", ""],
      ["", "O", ""],
      ["", "", ""]
    ],
    "current_player": "X|O",
    "player_x": "user_id",
    "player_o": "user_id",
    "status": "waiting|active|finished",
    "result": "x_wins|o_wins|draw|none",
    "winner": "user_id",
    "move_count": 3,
    "game_mode": "casual|ranked"
  },
  "message": "move successful"
}
```

**Response (Invalid Move):**
```json
{
  "success": false,
  "game_state": { ... },
  "message": "error description"
}
```

**Errors:**
- `16 (UNAUTHENTICATED)`: User not authenticated
- `3 (INVALID_ARGUMENT)`: Invalid request payload or missing match_id
- `5 (NOT_FOUND)`: Game not found
- `13 (INTERNAL)`: Failed to save game state

**Move Validation:**
- Game must be active
- Must be player's turn
- Position must be within bounds (0-2)
- Cell must be empty

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/make_move \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"match_id": "game-uuid", "row": 1, "col": 1}'
```

---

### 5. Get Game State

**Endpoint:** `POST /v2/rpc/get_game_state`

**Description:** Retrieve current game state (for reconnection).

**Authentication:** Optional (public read)

**Request Body:**
```json
{
  "match_id": "string (uuid)"
}
```

**Response:**
```json
{
  "match_id": "uuid",
  "board": [["X","O",""], ["","X",""], ["","","O"]],
  "current_player": "X|O",
  "player_x": "user_id",
  "player_o": "user_id",
  "status": "active|finished",
  "result": "none|x_wins|o_wins|draw",
  "winner": "user_id",
  "move_count": 5,
  "game_mode": "casual|ranked"
}
```

**Errors:**
- `3 (INVALID_ARGUMENT)`: match_id is required
- `5 (NOT_FOUND)`: Game not found
- `13 (INTERNAL)`: Failed to load game state

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/get_game_state \
  -H "Content-Type: application/json" \
  -d '{"match_id": "game-uuid"}'
```

---

### 6. Resign Game

**Endpoint:** `POST /v2/rpc/resign_game`

**Description:** Forfeit the current game.

**Authentication:** Required

**Request Body:**
```json
{
  "match_id": "string (uuid)"
}
```

**Response:**
```json
{
  "match_id": "uuid",
  "status": "finished",
  "result": "x_wins|o_wins",
  "winner": "opponent_user_id",
  ...
}
```

**Errors:**
- `16 (UNAUTHENTICATED)`: User not authenticated
- `3 (INVALID_ARGUMENT)`: match_id is required
- `5 (NOT_FOUND)`: Game not found
- `13 (INTERNAL)`: Failed to save game state

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/resign_game \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"match_id": "game-uuid"}'
```

---

### 7. Get Leaderboard

**Endpoint:** `GET /v2/rpc/get_leaderboard`

**Description:** Retrieve top 100 players.

**Authentication:** Optional

**Response:**
```json
{
  "entries": [
    {
      "user_id": "uuid",
      "username": "player1",
      "rank": 1,
      "score": 1250,
      "num_score": 25
    },
    {
      "user_id": "uuid",
      "username": "player2",
      "rank": 2,
      "score": 1180,
      "num_score": 18
    }
  ]
}
```

**Errors:**
- `13 (INTERNAL)`: Failed to retrieve leaderboard

**Example:**
```bash
curl -X GET http://localhost:7350/v2/rpc/get_leaderboard
```

---

### 8. Get Player Rank

**Endpoint:** `POST /v2/rpc/get_player_rank`

**Description:** Get specific player's rank and statistics.

**Authentication:** Required

**Request Body:**
```json
{
  "user_id": "string (optional, defaults to authenticated user)"
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "username": "player1",
  "rank": 42,
  "rating": 1180,
  "wins": 15,
  "losses": 10,
  "draws": 3
}
```

**Errors:**
- `16 (UNAUTHENTICATED)`: User not authenticated
- `13 (INTERNAL)`: Failed to get profile or account

**Example:**
```bash
curl -X POST http://localhost:7350/v2/rpc/get_player_rank \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## WebSocket Real-time Gameplay

**WebSocket URL:** `ws://localhost:7350/ws`

**Connection:** Use Nakama client SDK or WebSocket library

**Message Format:**
```json
{
  "op_code": 1,  // Operation code
  "data": { ... }  // Message payload
}
```

**Operation Codes:**

| OpCode | Name | Direction | Description |
|--------|------|-----------|-------------|
| 1 | Move | Client → Server | Send player move |
| 2 | GameState | Server → Client | Broadcast game state update |
| 3 | PlayerJoined | Server → Client | Player joined the match |
| 4 | PlayerLeft | Server → Client | Player left the match |
| 5 | GameOver | Server → Client | Game finished |

**Move Message (Client → Server):**
```json
{
  "op_code": 1,
  "data": {
    "row": 1,
    "col": 2
  }
}
```

**Game State Update (Server → Client):**
```json
{
  "op_code": 2,
  "data": {
    "match_id": "uuid",
    "board": [["X","O",""], ["","X",""], ["","",""]],
    "current_player": "O",
    "status": "active",
    ...
  }
}
```

**Player Joined (Server → Client):**
```json
{
  "op_code": 3,
  "data": {
    "user_id": "uuid",
    "username": "player1"
  }
}
```

**Player Left (Server → Client):**
```json
{
  "op_code": 4,
  "data": {
    "user_id": "uuid"
  }
}
```

---

## Error Codes

Standard gRPC error codes:

| Code | Name | Description |
|------|------|-------------|
| 3 | INVALID_ARGUMENT | Invalid request parameters |
| 5 | NOT_FOUND | Resource not found |
| 13 | INTERNAL | Internal server error |
| 16 | UNAUTHENTICATED | User not authenticated |

---

## ELO Rating System

**Formula:** Standard ELO with K-factor of 32

**Initial Rating:** 1000

**Rating Update (Win):**
```
Expected = 1 / (1 + 10^((OpponentRating - PlayerRating) / 400))
NewRating = PlayerRating + 32 * (1 - Expected)
```

**Rating Update (Loss):**
```
Expected = 1 / (1 + 10^((OpponentRating - PlayerRating) / 400))
NewRating = PlayerRating + 32 * (0 - Expected)
```

**Rating Update (Draw):**
```
Expected = 1 / (1 + 10^((OpponentRating - PlayerRating) / 400))
NewRating = PlayerRating + 32 * (0.5 - Expected)
```

**Minimum Rating:** 100

---

## Rate Limits

No rate limits enforced by default. Configure in `nakama/docker-compose.yml` if needed.

---

## Storage Collections

**profiles:** User game statistics and ratings
**games:** Active and finished game states
**matchmaking_queue:** Players waiting for matches

---

## Notes

- Session tokens expire after 2 hours (7200 seconds)
- Queue entries expire after 60 seconds
- Matchmaking for ranked mode considers ±200 rating difference
- Game states persist for replay/analysis
- WebSocket connections auto-reconnect on network issues
