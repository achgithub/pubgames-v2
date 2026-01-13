import { useState, useEffect, useCallback, useRef } from 'react';
import {
  sendHeartbeat,
  getOnlineUsers,
  getActiveGame,
  getPendingChallenges
} from '../services/gameApi';

export const useGameState = (user) => {
  const [activeGame, setActiveGame] = useState(null);
  const [onlineUsers, setOnlineUsers] = useState([]);
  const [pendingChallenges, setPendingChallenges] = useState([]);
  const gamePollIntervalRef = useRef(null);
  const isInitialMount = useRef(true);

  // Manual refresh - called after user actions
  const refreshData = useCallback(async () => {
    if (!user) return;

    try {
      const game = await getActiveGame();
      setActiveGame(game);

      // Only fetch these when not in active game
      if (!game || game.status !== 'active') {
        const [users, challenges] = await Promise.all([
          getOnlineUsers(),
          getPendingChallenges()
        ]);
        setOnlineUsers(users);
        setPendingChallenges(challenges);
      }
    } catch (err) {
      console.error('Failed to refresh game state:', err);
      // Don't throw - just log and continue
    }
  }, [user]);

  // Refresh lobby data only (manual)
  const refreshLobby = useCallback(async () => {
    if (!user) return;
    
    try {
      // IMPORTANT: Also check for active game so challenger detects when challenge accepted
      const [game, users, challenges] = await Promise.all([
        getActiveGame(),
        getOnlineUsers(),
        getPendingChallenges()
      ]);
      setActiveGame(game);
      setOnlineUsers(users);
      setPendingChallenges(challenges);
    } catch (err) {
      console.error('Failed to refresh lobby:', err);
      // Set to empty arrays on error to prevent hanging
      setOnlineUsers([]);
      setPendingChallenges([]);
    }
  }, [user]);

  // Heartbeat - ONLY constant polling (keep user online)
  useEffect(() => {
    if (!user) return;

    const heartbeat = async () => {
      try {
        await sendHeartbeat();
      } catch (err) {
        console.error('Heartbeat failed:', err);
      }
    };

    // Initial heartbeat and data fetch
    const initialize = async () => {
      try {
        await heartbeat();
        await refreshData();
      } catch (err) {
        console.error('Failed to initialize:', err);
      } finally {
        isInitialMount.current = false;
      }
    };

    initialize();

    // Heartbeat every 30s (ONLY constant interval)
    const heartbeatInterval = setInterval(heartbeat, 30000);

    return () => {
      clearInterval(heartbeatInterval);
    };
  }, [user]); // Remove refreshData from dependencies to prevent loops

  // Poll for opponent's move ONLY when waiting for their turn
  useEffect(() => {
    // Clear any existing polling
    if (gamePollIntervalRef.current) {
      clearInterval(gamePollIntervalRef.current);
      gamePollIntervalRef.current = null;
    }

    if (!user || !activeGame || activeGame.status !== 'active') {
      return; // No polling when no active game
    }

    // Determine if it's opponent's turn
    const isPlayer1 = user.id === activeGame.player1_id;
    const playerNumber = isPlayer1 ? 1 : 2;
    const isMyTurn = activeGame.current_turn === playerNumber;

    if (!isMyTurn) {
      // It's opponent's turn - poll for their move every 10 seconds
      const pollForMove = async () => {
        try {
          const game = await getActiveGame();
          if (game && JSON.stringify(game) !== JSON.stringify(activeGame)) {
            setActiveGame(game);
          }
        } catch (err) {
          console.error('Failed to poll for move:', err);
        }
      };

      gamePollIntervalRef.current = setInterval(pollForMove, 10000);
    }

    return () => {
      if (gamePollIntervalRef.current) {
        clearInterval(gamePollIntervalRef.current);
        gamePollIntervalRef.current = null;
      }
    };
  }, [activeGame?.id, activeGame?.status, activeGame?.current_turn, user]); // Only watch specific props

  return {
    activeGame,
    setActiveGame,
    onlineUsers,
    pendingChallenges,
    refreshData,
    refreshLobby
  };
};
