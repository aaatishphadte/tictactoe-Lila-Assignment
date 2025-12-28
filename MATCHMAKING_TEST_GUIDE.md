# How to Test Multiplayer Matchmaking

## Quick Method (Easiest)

### Option 1: Two Browser Windows

1. **Open First Player Window**
   - Open Chrome/Edge: http://localhost:5173
   - Enter nickname: `Player1`
   - Click Continue
   - You'll see "Finding a random player..."

2. **Open Second Player Window** 
   - Open a NEW BROWSER WINDOW (or Incognito/Private window)
   - Go to: http://localhost:5173
   - Enter nickname: `Player2`
   - Click Continue

3. **Watch Them Match**
   - Within a few seconds, BOTH windows should show the game board
   - One player will be **X** (blue), the other will be **O** (red)
   - The game board will appear with "Your turn" or "Opponent's turn"

4. **Play the Game**
   - Click on empty squares in the grid
   - Alternate turns between the two windows
   - Win by getting 3 in a row!

---

## Option 2: Two Different Browsers

Use two completely different browsers for cleaner testing:

1. **Browser 1** (e.g., Chrome)
   - Open http://localhost:5173
   - Enter nickname: `Alice`
   - Start matchmaking

2. **Browser 2** (e.g., Edge or Firefox)
   - Open http://localhost:5173
   - Enter nickname: `Bob`
   - Start matchmaking

3. **They should match immediately!**

---

## Option 3: Two Devices (Best for Real Testing)

If you have another computer or phone on the same network:

1. **Find your local IP**
   ```bash
   ipconfig
   # Look for IPv4 Address (e.g., 192.168.1.100)
   ```

2. **Update Nakama config** (if needed)
   - Backend should accept connections from network

3. **On Device 1** (Your Computer)
   - http://localhost:5173

4. **On Device 2** (Another Computer/Phone)
   - http://YOUR_IP:5173 (e.g., http://192.168.1.100:5173)

---

## What to Expect

### Successful Matchmaking:
✅ Both players see "Match found!"  
✅ Game board appears in both windows  
✅ One player is assigned **X** (goes first)  
✅ Other player is assigned **O** (goes second)  
✅ Turn indicator shows "Your turn" or "Opponent's turn"  

### Matchmaking Timeline:
- **0-5 seconds**: First player enters queue
- **5-10 seconds**: Second player joins
- **10-15 seconds**: Nakama matches them together
- **15-30 seconds**: If no match, player waits (can cancel)

---

## Troubleshooting

### Players Not Matching?

**Check Backend:**
```bash
cd nakama
docker logs nakama --tail 50
# Look for matchmaker messages
```

**Check Browser Console:**
- Press F12 in browser
- Go to "Console" tab
- Look for errors in red

**Common Issues:**
- ❌ Backend not running → Start with `docker-compose up -d`
- ❌ Frontend not running → Start with `npm run dev` in client folder
- ❌ Wrong port configured → Check port is 8350 in `nakama.js`

---

## Testing Checklist

- [ ] Both players can enter nicknames
- [ ] Matchmaking spinner appears
- [ ] Players match within 30 seconds
- [ ] Game board loads for both
- [ ] Turns alternate correctly
- [ ] Moves appear in both windows
- [ ] Win detection works
- [ ] Winner screen shows
- [ ] Leaderboard updates

---

## Quick Test Command

If you want to see matchmaking status in backend:
```bash
docker logs nakama -f | findstr "matchmaker"
```

Keep this running while testing to see matchmaking events in real-time!

---

## Currently Running

Your servers are already running:
- ✅ Frontend: http://localhost:5173
- ✅ Backend: localhost:8350

**Just open 2 browser windows and you're ready to test!**
