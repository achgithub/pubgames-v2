import { useState, useEffect, useRef, useCallback } from 'react';

export const useLobbyWebSocket = (enabled, user, callbacks) => {
  const [status, setStatus] = useState('disconnected'); // disconnected, connecting, connected, error
  const [error, setError] = useState(null);
  const wsRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef(null);

  const {
    onChallengeReceived,
    onChallengeAccepted,
    onChallengeDeclined,
  } = callbacks || {};

  const connect = useCallback(() => {
    if (!enabled || !user) return;

    console.log('🏛️ Connecting to Lobby WebSocket...');
    setStatus('connecting');
    setError(null);

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = '30041';
    const token = localStorage.getItem('jwt_token');
    
    const ws = new WebSocket(`${protocol}//${host}:${port}/api/ws/lobby?token=${token}`);
    
    wsRef.current = ws;

    // Connection opened
    ws.onopen = () => {
      console.log('✅ Lobby WebSocket connected');
      setStatus('connected');
      reconnectAttemptsRef.current = 0;
    };

    // Message received
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log('📨 Lobby WS Message:', msg.type);

        switch (msg.type) {
          case 'lobby_connected':
            console.log('✅ Lobby connection confirmed');
            break;

          case 'challenge_received':
            console.log('📨 Challenge received!', msg.payload);
            if (onChallengeReceived) {
              onChallengeReceived(msg.payload);
            }
            break;

          case 'challenge_accepted':
            console.log('✅ Challenge accepted - switching to game!', msg.payload);
            if (onChallengeAccepted) {
              onChallengeAccepted(msg.payload);
            }
            break;

          case 'challenge_declined':
            console.log('❌ Challenge declined', msg.payload);
            if (onChallengeDeclined) {
              onChallengeDeclined(msg.payload);
            }
            break;

          default:
            console.log('📨 Unknown lobby message type:', msg.type);
        }
      } catch (err) {
        console.error('Failed to parse lobby message:', err);
      }
    };

    // Connection closed
    ws.onclose = (event) => {
      console.log(`🔌 Lobby WebSocket closed (code: ${event.code})`);
      setStatus('disconnected');
      
      // Only attempt reconnect if still enabled and it wasn't a clean close
      if (enabled && event.code !== 1000) {
        attemptReconnect();
      }
    };

    // Error
    ws.onerror = (err) => {
      console.error('❌ Lobby WebSocket error:', err);
      setError('Connection error');
      setStatus('error');
    };

  }, [enabled, user, onChallengeReceived, onChallengeAccepted, onChallengeDeclined]);

  const attemptReconnect = useCallback(() => {
    if (reconnectAttemptsRef.current >= 3) {
      console.error('❌ Max lobby reconnection attempts reached');
      setStatus('error');
      setError('Connection lost. Please refresh.');
      return;
    }

    reconnectAttemptsRef.current++;
    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 10000);
    
    console.log(`🔄 Reconnecting lobby in ${delay}ms (attempt ${reconnectAttemptsRef.current}/3)`);
    
    reconnectTimeoutRef.current = setTimeout(() => {
      connect();
    }, delay);
  }, [connect]);

  const disconnect = useCallback(() => {
    console.log('🔌 Disconnecting Lobby WebSocket...');
    
    // Clear any pending timeouts
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    // Close WebSocket connection
    if (wsRef.current) {
      wsRef.current.close(1000, 'Client disconnecting'); // Clean close
      wsRef.current = null;
    }
    
    setStatus('disconnected');
    setError(null);
    reconnectAttemptsRef.current = 0;
  }, []);

  // Connect/disconnect based on enabled flag
  useEffect(() => {
    if (enabled && user) {
      connect();
    } else {
      disconnect();
    }
    
    return () => {
      disconnect();
    };
  }, [enabled, user, connect, disconnect]);

  return {
    status,      // 'disconnected' | 'connecting' | 'connected' | 'error'
    error,       // Error message if status is 'error'
    disconnect   // Function to manually disconnect
  };
};
