import { useState, useEffect, useRef, useCallback } from 'react';

export const useGameWebSocket = (gameId, user, onGameUpdate, onGameEnded) => {
  const [status, setStatus] = useState('disconnected'); // disconnected, connecting, handshaking, connected, error
  const [error, setError] = useState(null);
  const wsRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef(null);
  const handshakeTimeoutRef = useRef(null);

  const connect = useCallback(() => {
    if (!gameId || !user) return;

    console.log(`🔌 Connecting WebSocket for Game ${gameId}...`);
    setStatus('connecting');
    setError(null);

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = '30041';
    const token = localStorage.getItem('jwt_token');
    
    const ws = new WebSocket(`${protocol}//${host}:${port}/api/ws/game/${gameId}?token=${token}`);
    
    wsRef.current = ws;

    // Connection opened
    ws.onopen = () => {
      console.log('✅ WebSocket connection opened');
      setStatus('handshaking');
      performHandshake(ws);
    };

    // Message received
    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      handleMessage(msg);
    };

    // Connection closed
    ws.onclose = (event) => {
      console.log(`🔌 WebSocket closed (code: ${event.code})`);
      setStatus('disconnected');
      
      // Only attempt reconnect if it wasn't a clean close
      if (event.code !== 1000) {
        attemptReconnect();
      }
    };

    // Error
    ws.onerror = (err) => {
      console.error('❌ WebSocket error:', err);
      setError('Connection error');
      setStatus('error');
    };

  }, [gameId, user]);

  const performHandshake = (ws) => {
    console.log('🤝 Starting handshake...');
    
    // 1. Send PING
    ws.send(JSON.stringify({ type: 'ping' }));
    console.log('📤 Sent PING');
    
    // Set timeout for handshake
    handshakeTimeoutRef.current = setTimeout(() => {
      console.error('⏰ Handshake timeout');
      setError('Connection timeout');
      setStatus('error');
      ws.close();
    }, 10000); // 10 second timeout
  };

  const handleMessage = (msg) => {
    console.log('📨 WS Message:', msg.type);

    switch (msg.type) {
      case 'pong':
        console.log('📨 Received PONG');
        // 2. Received PONG, send ACK
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({ type: 'ack' }));
          console.log('📤 Sent ACK');
        }
        break;

      case 'ready':
        console.log('✅ Received READY - Handshake complete!');
        // Clear handshake timeout
        if (handshakeTimeoutRef.current) {
          clearTimeout(handshakeTimeoutRef.current);
          handshakeTimeoutRef.current = null;
        }
        
        // 3. Handshake complete!
        setStatus('connected');
        reconnectAttemptsRef.current = 0;
        setError(null);
        
        // Update game state from READY payload
        if (onGameUpdate && msg.payload) {
          onGameUpdate(msg.payload);
        }
        break;

      case 'move_update':
        console.log('📨 Received move_update');
        if (onGameUpdate && msg.payload) {
          onGameUpdate(msg.payload);
        }
        break;

      case 'game_ended':
        console.log('📨 Received game_ended');
        if (onGameEnded && msg.payload) {
          onGameEnded(msg.payload);
        }
        break;

      case 'opponent_disconnected':
        console.log('⚠️ Opponent disconnected');
        // Could show notification to user
        break;

      default:
        console.log('📨 Unknown message type:', msg.type);
    }
  };

  const attemptReconnect = () => {
    if (reconnectAttemptsRef.current >= 3) {
      console.error('❌ Max reconnection attempts reached');
      setStatus('error');
      setError('Connection lost. Please try again.');
      return;
    }

    reconnectAttemptsRef.current++;
    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 10000);
    
    console.log(`🔄 Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current}/3)`);
    setStatus('reconnecting');
    
    reconnectTimeoutRef.current = setTimeout(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'reconnecting' }));
      }
      connect();
    }, delay);
  };

  const disconnect = useCallback(() => {
    console.log('🔌 Disconnecting WebSocket...');
    
    // Clear any pending timeouts
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (handshakeTimeoutRef.current) {
      clearTimeout(handshakeTimeoutRef.current);
      handshakeTimeoutRef.current = null;
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

  // Connect on mount if gameId provided
  useEffect(() => {
    if (gameId && user) {
      connect();
    }
    
    return () => {
      disconnect();
    };
  }, [gameId, user, connect, disconnect]);

  return {
    status,      // 'disconnected' | 'connecting' | 'handshaking' | 'connected' | 'reconnecting' | 'error'
    error,       // Error message if status is 'error'
    disconnect   // Function to manually disconnect
  };
};
