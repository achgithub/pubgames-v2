import { useState, useEffect, useRef, useCallback } from 'react';

/**
 * Smart Lobby WebSocket - MOBILE-SAFE
 * 
 * KEY DIFFERENCES FROM OLD VERSION:
 * 1. Only connects when there's an ACTIVE challenge (sent or received)
 * 2. Auto-disconnects after 30s (backend enforces this too)
 * 3. Disconnects immediately when challenge resolved
 * 4. NO persistent connections = NO mobile battery drain
 * 
 * WHEN IT CONNECTS:
 * - You just sent a challenge → Connect for 30s max
 * - You have pending challenges → Connect for 30s max
 * - Challenge accepted/declined → Disconnect immediately
 * - Just browsing lobby → NO connection (zero overhead)
 */

export const useSmartLobbyWebSocket = (shouldConnect, user, callbacks) => {
  const [status, setStatus] = useState('disconnected');
  const wsRef = useRef(null);
  const timeoutRef = useRef(null);

  const {
    onChallengeReceived,
    onChallengeAccepted,
    onChallengeDeclined,
    onUserOffline,
  } = callbacks || {};

  const disconnect = useCallback(() => {
    console.log('🔌 Disconnecting Smart Lobby WS...');
    
    // Clear timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    
    // Close WebSocket
    if (wsRef.current) {
      wsRef.current.close(1000, 'Challenge resolved');
      wsRef.current = null;
    }
    
    setStatus('disconnected');
  }, []); // EMPTY dependencies - this function never changes

  const connect = useCallback(() => {
    if (!shouldConnect || !user) {
      disconnect();
      return;
    }

    // Don't reconnect if already connected
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      return;
    }

    console.log('🏛️ Connecting Smart Lobby WS (30s max)...');
    setStatus('connecting');

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = '30041';
    const token = localStorage.getItem('jwt_token');
    
    const ws = new WebSocket(`${protocol}//${host}:${port}/api/ws/lobby?token=${token}`);
    wsRef.current = ws;

    // Auto-disconnect failsafe (client-side, in case backend fails)
    timeoutRef.current = setTimeout(() => {
      console.log('⏰ Client-side 30s timeout - disconnecting');
      disconnect();
    }, 30000);

    ws.onopen = () => {
      console.log('✅ Smart Lobby WS connected (will auto-disconnect in 30s)');
      setStatus('connected');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log('📨 Smart Lobby WS:', msg.type);

        switch (msg.type) {
          case 'lobby_connected':
            console.log('✅ Lobby connection confirmed');
            break;

          case 'challenge_received':
            console.log('📨 Challenge received instantly!', msg.payload);
            if (onChallengeReceived) {
              onChallengeReceived(msg.payload);
            }
            break;

          case 'challenge_accepted':
            console.log('✅ Challenge accepted - switching to game!', msg.payload);
            if (onChallengeAccepted) {
              onChallengeAccepted(msg.payload);
            }
            // Disconnect immediately - challenge resolved
            disconnect();
            break;

          case 'challenge_declined':
            console.log('❌ Challenge declined', msg.payload);
            if (onChallengeDeclined) {
              onChallengeDeclined(msg.payload);
            }
            // Disconnect immediately - challenge resolved
            disconnect();
            break;

          case 'user_offline':
            console.log('👋 User went offline', msg.payload);
            if (onUserOffline) {
              onUserOffline(msg.payload);
            }
            break;

          default:
            console.log('📨 Unknown message type:', msg.type);
        }
      } catch (err) {
        console.error('Failed to parse lobby message:', err);
      }
    };

    ws.onclose = (event) => {
      console.log(`🔌 Smart Lobby WS closed (code: ${event.code})`);
      setStatus('disconnected');
      wsRef.current = null;
    };

    ws.onerror = (err) => {
      console.error('❌ Smart Lobby WS error:', err);
      setStatus('error');
    };

  }, [shouldConnect, user, onChallengeReceived, onChallengeAccepted, onChallengeDeclined, onUserOffline, disconnect]);

  // Connect/disconnect based on shouldConnect flag
  // CRITICAL: Only shouldConnect and user in dependencies to avoid infinite loops
  useEffect(() => {
    if (shouldConnect && user) {
      connect();
    } else if (wsRef.current) {
      // Only disconnect if we actually have a connection
      disconnect();
    }
    
    // Cleanup on unmount
    return () => {
      if (wsRef.current) {
        disconnect();
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [shouldConnect, user]); // ONLY these two - connect/disconnect would cause loops

  return {
    status,
    disconnect,
  };
};
