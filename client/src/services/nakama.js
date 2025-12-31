import { Client, Session } from '@heroiclabs/nakama-js';

const SERVER_KEY = import.meta.env.VITE_NAKAMA_SERVER_KEY || 'defaultkey';
const HOST = import.meta.env.VITE_NAKAMA_HOST || window.location.hostname;
const PORT = import.meta.env.VITE_NAKAMA_PORT || '8350';
const USE_SSL = import.meta.env.VITE_NAKAMA_USE_SSL === 'true';

class NakamaService {
    constructor() {
        this.client = new Client(SERVER_KEY, HOST, PORT, USE_SSL);
        this.session = null;
        this.socket = null;
        this.matchId = null;
        this.queueToken = null;
    }

    // Authentication
    async authenticateDevice(deviceId, username) {
        try {
            // Use Nakama's built-in authentication
            const session = await this.client.authenticateCustom(deviceId, true, username);
            this.session = session;

            // Immediately fetch profile/rank information
            const profile = await this.getPlayerRank(session.user_id);

            return {
                user_id: session.user_id,
                username: session.username,
                session_token: session.token,
                profile: {
                    wins: profile.wins || 0,
                    losses: profile.losses || 0,
                    draws: profile.draws || 0,
                    rating: profile.rating || 1000
                }
            };
        } catch (error) {
            console.error('Authentication failed:', error);
            throw error;
        }
    }

    // Built-in Matchmaking
    async startMatchmaking(onMatched, onError) {
        if (!this.session) {
            throw new Error('Not authenticated');
        }

        try {
            // Connect socket if not already connected
            if (!this.socket) {
                await this.connectSocket();
            }

            // Set up matchmaker matched handler
            this.socket.onmatchmakermatched = (matched) => {
                console.log('Match found:', matched);
                if (this.matchmakerTicket) {
                    this.matchmakerTicket = null;
                }
                if (onMatched) {
                    onMatched(matched);
                }
            };

            // Set up error handler
            this.socket.onerror = (error) => {
                console.error('Socket error:', error);
                if (onError) onError(error);
            };

            // Add to matchmaker with simple query
            const query = '*'; // Match with anyone
            const minCount = 2;
            const maxCount = 2;

            const ticket = await this.socket.addMatchmaker(query, minCount, maxCount);
            this.matchmakerTicket = ticket.ticket;
            console.log('Joined matchmaker:', this.matchmakerTicket);

            return ticket;
        } catch (error) {
            console.error('Matchmaking failed:', error);
            throw error;
        }
    }

    async cancelMatchmaking() {
        if (!this.socket || !this.matchmakerTicket) {
            return;
        }

        try {
            await this.socket.removeMatchmaker(this.matchmakerTicket);
            this.matchmakerTicket = null;
            console.log('Left matchmaker');
        } catch (error) {
            console.error('Cancel matchmaking failed:', error);
            throw error;
        }
    }

    // Game Operations
    async makeMove(row, col) {
        if (!this.session || !this.matchId) {
            throw new Error('Not in a match');
        }

        try {
            const response = await this.client.rpc(
                this.session,
                'make_move',
                {
                    match_id: this.matchId,
                    row,
                    col
                }
            );

            return response;
        } catch (error) {
            console.error('Make move failed:', error);
            throw error;
        }
    }

    async getGameState() {
        if (!this.session || !this.matchId) {
            throw new Error('Not in a match');
        }

        try {
            const response = await this.client.rpc(
                this.session,
                'get_game_state',
                { match_id: this.matchId }
            );

            return response;
        } catch (error) {
            console.error('Get game state failed:', error);
            throw error;
        }
    }

    async resignGame() {
        if (!this.session || !this.matchId) {
            throw new Error('Not in a match');
        }

        try {
            await this.client.rpc(
                this.session,
                'resign_game',
                { match_id: this.matchId }
            );
        } catch (error) {
            console.error('Resign game failed:', error);
            throw error;
        }
    }

    // Leaderboard
    async getLeaderboard() {
        if (!this.session) {
            throw new Error('Not authenticated');
        }

        try {
            console.log('[LEADERBOARD] 📊 Fetching leaderboard from Nakama...');
            const response = await this.client.rpc(
                this.session,
                'get_leaderboard',
                {}
            );

            console.log('[LEADERBOARD] Raw response:', response);
            console.log('[LEADERBOARD] Response type:', typeof response);
            console.log('[LEADERBOARD] Response keys:', Object.keys(response || {}));
            console.log('[LEADERBOARD] response.payload:', response.payload);

            // CRITICAL FIX: Nakama RPC wraps response in 'payload' field
            const entries = response.payload?.entries || [];
            console.log('[LEADERBOARD] Entries array length:', entries.length);
            console.log('[LEADERBOARD] Entries:', entries);

            // Transform backend format to frontend format
            // Backend returns: { user_id, username, rank, score, num_score }
            // Frontend expects: { user_id, username, wins, losses, draws, rating }
            const transformedEntries = entries.map((entry, index) => {
                console.log(`[LEADERBOARD] Entry ${index}:`, entry);
                return {
                    user_id: entry.user_id,
                    username: entry.username,
                    rank: entry.rank,
                    rating: entry.score,  // Backend's 'score' is the ELO rating
                    wins: 0,              // TODO: Fetch from player profile
                    losses: 0,            // TODO: Fetch from player profile  
                    draws: 0              // TODO: Fetch from player profile
                };
            });

            console.log('[LEADERBOARD] ✅ Transformed entries:', transformedEntries);
            return transformedEntries;
        } catch (error) {
            console.error('[LEADERBOARD] ❌ Get leaderboard failed:', error);
            console.error('[LEADERBOARD] Error details:', error.message, error.stack);
            throw error;
        }
    }

    async getPlayerRank(userId = null) {
        if (!this.session) {
            throw new Error('Not authenticated');
        }

        try {
            const response = await this.client.rpc(
                this.session,
                'get_player_rank',
                userId ? { user_id: userId } : {}
            );

            return response;
        } catch (error) {
            console.error('Get player rank failed:', error);
            throw error;
        }
    }

    // WebSocket Operations  
    async connectSocket() {
        if (this.socket && this.socket.isConnected) {
            return this.socket;
        }

        const useSSL = false;
        const verboseLogging = true;

        this.socket = this.client.createSocket(useSSL, verboseLogging);
        await this.socket.connect(this.session, true);

        console.log('Socket connected');
        return this.socket;
    }

    async joinMatch(matchId) {
        if (!this.socket) {
            throw new Error('Socket not connected');
        }

        const match = await this.socket.joinMatch(matchId);
        this.matchId = matchId;
        this.currentMatch = match;

        // Set up match data handler
        this.socket.onmatchdata = (matchData) => {
            console.log('Match data received:', matchData);
            if (this.onMatchDataCallback) {
                this.onMatchDataCallback(matchData);
            }
        };

        return match;
    }

    async leaveMatch() {
        if (!this.socket || !this.matchId) {
            return;
        }

        await this.socket.leaveMatch(this.matchId);
        this.matchId = null;
        this.currentMatch = null;
    }

    async disconnectSocket() {
        if (this.socket) {
            this.socket.disconnect();
            this.socket = null;
        }
    }


    // Utility
    reset() {
        this.session = null;
        this.matchId = null;
        this.queueToken = null;
        this.disconnectSocket();
    }
}

// Export a singleton instance
const nakamaService = new NakamaService();
export default nakamaService;
