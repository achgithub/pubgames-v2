import axios, { API_BASE } from './api';

// Config
export const getConfig = async () => {
  const res = await axios.get(`${API_BASE}/config`);
  return res.data;
};

// Heartbeat
export const sendHeartbeat = async () => {
  const res = await axios.post(`${API_BASE}/heartbeat`);
  return res.data;
};

// Logout - clears user from online_users
export const sendLogout = async () => {
  try {
    const res = await axios.post(`${API_BASE}/logout`);
    return res.data;
  } catch (err) {
    // Ignore errors - user might already be logged out
    console.log('Logout API call failed (might already be logged out)');
    return { success: true };
  }
};

// Online users
export const getOnlineUsers = async () => {
  const res = await axios.get(`${API_BASE}/online-users`);
  return res.data || [];
};

// Challenges
export const createChallenge = async (opponentId, mode, moveTimeLimit, firstTo) => {
  const res = await axios.post(`${API_BASE}/game/create-challenge`, {
    opponent_id: opponentId,
    mode,
    move_time_limit: mode === 'timed' ? moveTimeLimit : 0,
    first_to: firstTo
  });
  return res.data;
};

export const getPendingChallenges = async () => {
  const res = await axios.get(`${API_BASE}/game/pending-challenges`);
  return res.data || [];
};

export const respondToChallenge = async (gameId, accept) => {
  const res = await axios.post(`${API_BASE}/game/${gameId}/respond`, { accept });
  return res.data;
};

// Game state
export const getActiveGame = async () => {
  const res = await axios.get(`${API_BASE}/game/active`);
  return res.data;
};

export const makeMove = async (gameId, position) => {
  const res = await axios.post(`${API_BASE}/game/move`, {
    game_id: gameId,
    position
  });
  return res.data;
};

// Rematch
export const createRematchRequest = async (gameId) => {
  const res = await axios.post(`${API_BASE}/game/rematch`, { game_id: gameId });
  return res.data;
};

export const getRematchRequest = async (gameId) => {
  const res = await axios.get(`${API_BASE}/game/rematch/${gameId}`);
  return res.data;
};

export const respondToRematch = async (rematchId, accept) => {
  const res = await axios.post(`${API_BASE}/game/rematch/${rematchId}/respond`, { accept });
  return res.data;
};

// Stats
export const getPlayerStats = async () => {
  const res = await axios.get(`${API_BASE}/stats/player`);
  return res.data;
};

export const getLeaderboard = async () => {
  const res = await axios.get(`${API_BASE}/stats/leaderboard`);
  return res.data || [];
};

export const getGameHistory = async () => {
  const res = await axios.get(`${API_BASE}/history`);
  return res.data || [];
};
