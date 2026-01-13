import React, { useState, useEffect } from 'react';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { useGameState } from './hooks/useGameState';
import { useGameWebSocket } from './hooks/useGameWebSocket';
import { useSmartLobbyWebSocket } from './hooks/useSmartLobbyWebSocket';
import { 
  createChallenge, 
  respondToChallenge, 
  makeMove as makeMoveApi,
  getPlayerStats,
  getLeaderboard,
  getGameHistory,
  getConfig,
  sendLogout
} from './services/gameApi';
import { IDENTITY_URL } from './services/api';
import { Lobby } from './components/Lobby';
import { GameView } from './components/GameView';
import { StatsView } from './components/StatsView';
import { ChallengeModal } from './components/ChallengeModal';

function GameApp() {
  const { user, loading, logout } = useAuth();
  const [view, setView] = useState('lobby');
  const [config, setConfig] = useState({
    app_name: 'Tic Tac Toe',
    app_icon: '📤'
  });

  // Game state from custom hook
  const {
    activeGame,
    setActiveGame,
    onlineUsers,
    pendingChallenges,
    refreshData,
    refreshLobby
  } = useGameState(user);

  // UI state
  const [pendingMove, setPendingMove] = useState(null);
  const [showChallengeModal, setShowChallengeModal] = useState(false);
  const [selectedOpponent, setSelectedOpponent] = useState(null);
  const [gameResult, setGameResult] = useState(null); // Show game end result
  const [justSentChallenge, setJustSentChallenge] = useState(false); // Track if we just sent challenge

  // Stats state
  const [playerStats, setPlayerStats] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [gameHistory, setGameHistory] = useState([]);

  // WebSocket for real-time game updates
  const { status: websocketStatus, error: websocketError, disconnect: disconnectWebSocket } = 
    useGameWebSocket(
      activeGame?.status === 'active' ? activeGame?.id : null, // Only connect for active games
      user,
      (game) => {
        // Game update received via WebSocket
        console.log('🎮 Game update received via WebSocket');
        setActiveGame(game);
      },
      (game) => {
        // Game ended
        console.log('🏁 Game ended via WebSocket');
        setActiveGame(game);
      }
    );

  // Smart Lobby WebSocket - MOBILE-SAFE!
  // Connect when: in lobby view OR just sent challenge OR have pending challenges
  // This ensures opponent can receive FIRST challenge notification
  const shouldConnectLobby = view === 'lobby' || justSentChallenge || pendingChallenges.length > 0;
  
  useSmartLobbyWebSocket(
    shouldConnectLobby,
    user,
    {
      onChallengeReceived: (challenge) => {
        console.log('📨 Challenge received instantly via WebSocket!');
        refreshLobby();
      },
      onChallengeAccepted: (game) => {
        console.log('✅ Challenge accepted - switching to game!');
        setActiveGame(game);
        setJustSentChallenge(false); // Clear flag
        setView('optimizing');
      },
      onChallengeDeclined: (data) => {
        console.log('❌ Challenge declined');
        alert('Your challenge was declined.');
        setJustSentChallenge(false); // Clear flag
        refreshLobby();
      },
      onUserOffline: (data) => {
        console.log('👋 User went offline:', data.user_name);
        // Refresh lobby to remove the offline user
        refreshLobby();
      }
    }
  );

  // Clear justSentChallenge flag after 30s (backup cleanup)
  useEffect(() => {
    if (justSentChallenge) {
      const timer = setTimeout(() => {
        console.log('⏰ Clearing justSentChallenge flag (30s timeout)');
        setJustSentChallenge(false);
      }, 30000);
      
      return () => clearTimeout(timer);
    }
  }, [justSentChallenge]);

  // Fallback polling for challenger - in case WebSocket message missed
  // Poll every 2s for 30s max when justSentChallenge is true
  useEffect(() => {
    if (!justSentChallenge || !user) return;
    
    console.log('🔄 Challenger fallback polling started (2s interval)');
    
    const pollInterval = setInterval(async () => {
      console.log('🔄 Checking for game start...');
      try {
        await refreshData();
      } catch (err) {
        console.error('Fallback poll failed:', err);
      }
    }, 2000);
    
    return () => {
      console.log('🛑 Challenger fallback polling stopped');
      clearInterval(pollInterval);
    };
  }, [justSentChallenge, user, refreshData]);

  // Auto-switch to game when challenge is accepted (game becomes active)
  useEffect(() => {
    if (activeGame && activeGame.status === 'active' && view === 'lobby') {
      console.log('✅ Challenge accepted - switching to game!');
      setView('optimizing');
    }
  }, [activeGame, view]);

  // Load config
  useEffect(() => {
    getConfig().then(setConfig).catch(console.error);
  }, []);

  // Handle WebSocket connection status
  useEffect(() => {
    if (websocketStatus === 'connected' && activeGame?.status === 'active' && view === 'optimizing') {
      // WebSocket connected successfully, enter game
      console.log('✅ WebSocket connected - entering game');
      setView('game');
    } else if (websocketStatus === 'error' && view === 'optimizing') {
      // WebSocket failed, show error and return to lobby
      alert('Technical issues. Please try again shortly.');
      setView('lobby');
      setActiveGame(null);
    }
  }, [websocketStatus, activeGame, view]);

  // Safety timeout: if stuck in optimizing for >10 seconds, return to lobby
  useEffect(() => {
    if (view === 'optimizing') {
      const timeout = setTimeout(() => {
        console.error('⏱️ Optimizing timeout - returning to lobby');
        alert('Connection timeout. Please try again.');
        setView('lobby');
        setActiveGame(null);
      }, 10000); // 10 seconds
      
      return () => clearTimeout(timeout);
    }
  }, [view]);

  // Clear pending move only when that position is actually filled
  useEffect(() => {
    if (!pendingMove || !activeGame) return;
    
    try {
      const board = JSON.parse(activeGame.board);
      if (board[pendingMove.position] === pendingMove.symbol) {
        // The move succeeded - clear pending
        setPendingMove(null);
      }
    } catch (e) {
      // Invalid board, clear pending
      setPendingMove(null);
    }
  }, [activeGame, pendingMove]);

  // Auto-return to lobby when game completes
  useEffect(() => {
    if (!activeGame) return;
    
    // When game just completed, show result then return to lobby
    if (activeGame.status === 'completed') {
      // Determine result message
      let message = '';
      if (activeGame.winner_id === user.id) {
        message = `🏆 You Won! (${activeGame.player1_score}-${activeGame.player2_score})`;
      } else if (activeGame.winner_id) {
        message = `Game Over. (${activeGame.player1_score}-${activeGame.player2_score})`;
      } else {
        message = `Game Draw! (${activeGame.player1_score}-${activeGame.player2_score})`;
      }
      
      setGameResult(message);
      
      const timer = setTimeout(() => {
        console.log('🏁 Game completed - returning to lobby');
        setGameResult(null);
        setActiveGame(null);
        setView('lobby');
        // Refresh lobby data
        refreshLobby();
      }, 3000);
      
      return () => clearTimeout(timer);
    }
  }, [activeGame?.status, activeGame?.id, activeGame?.winner_id, activeGame?.player1_score, activeGame?.player2_score, user?.id, refreshLobby]);

  // Load stats when viewing stats page
  const loadStatsAndHistory = async () => {
    try {
      const [stats, lb, history] = await Promise.all([
        getPlayerStats(),
        getLeaderboard(),
        getGameHistory()
      ]);
      setPlayerStats(stats);
      setLeaderboard(lb);
      setGameHistory(history);
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  };

  useEffect(() => {
    if (user && view === 'stats') {
      loadStatsAndHistory();
    }
  }, [user, view]);

  // Handlers
  const handleBackToApps = () => {
    window.location.href = IDENTITY_URL;
  };

  const handleLogout = async () => {
    // Clear user from online_users before logging out
    await sendLogout();
    
    logout();
    setTimeout(() => {
      window.location.href = `${IDENTITY_URL}?logout=true`;
    }, 100);
  };

  const handleChallenge = (opponent) => {
    setSelectedOpponent(opponent);
    setShowChallengeModal(true);
    // Pre-connect WebSocket for instant notification when opponent responds
    setJustSentChallenge(true);
  };

  const handleSendChallenge = async (opponentId, mode, moveTimeLimit, firstTo) => {
    try {
      await createChallenge(opponentId, mode, moveTimeLimit, firstTo);
      setShowChallengeModal(false);
      setSelectedOpponent(null);
      
      // WebSocket already connected (from handleChallenge)
      // Will receive instant notification when opponent responds
      // NO ALERT - WebSocket event will handle the flow
    } catch (err) {
      alert(err.response?.data?.error || 'Failed to send challenge');
      setJustSentChallenge(false); // Clear on error
    }
  };

  const handleRespondToChallenge = async (gameId, accept) => {
    try {
      await respondToChallenge(gameId, accept);
      
      if (accept) {
        // Show optimizing screen
        setView('optimizing');
        
        // Fetch initial game data
        await refreshData();
        
        // WebSocket will connect and switch to game view automatically
      } else {
        // Challenge declined, refresh lobby
        await refreshLobby();
      }
    } catch (err) {
      alert(err.response?.data?.error || 'Failed to respond to challenge');
    }
  };

  const handleMove = async (position) => {
    if (!activeGame || !user || pendingMove) return;

    const isPlayer1 = user.id === activeGame.player1_id;
    const playerNumber = isPlayer1 ? 1 : 2;

    if (activeGame.current_turn !== playerNumber) {
      alert("It's not your turn!");
      return;
    }

    const board = JSON.parse(activeGame.board);
    if (board[position] !== '') {
      alert('That position is already taken!');
      return;
    }

    // Optimistic update
    const symbol = isPlayer1 ? 'X' : 'O';
    setPendingMove({ position, symbol });

    try {
      const res = await makeMoveApi(activeGame.id, position);
      
      // WebSocket will broadcast the update to both players!
      // No need to call refreshData()
      
      setPendingMove(null);

      if (res.series_over) {
        // Series complete - modal will show via useEffect
      } else if (res.round_over) {
        if (res.is_draw) {
          alert(`Round ${activeGame.current_round} is a draw!`);
        } else if (!res.series_over) {
          alert(`Round ${activeGame.current_round} complete! Next round starting...`);
        }
      }
    } catch (err) {
      setPendingMove(null);
      alert(err.response?.data?.error || 'Failed to make move');
      // On error, refresh to get correct state
      refreshData();
    }
  };

  const handleGoToStandings = () => {
    // Disconnect WebSocket
    disconnectWebSocket();
    
    // Trigger one poll as safety net
    setTimeout(() => refreshData(), 500);
    
    setShowRematchModal(false);
    setRematchRequest(null);
    setView('stats');
    loadStatsAndHistory();
  };

  // Render states
  if (loading) {
    return (
      <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
        <h2>Loading...</h2>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
        <h1 style={{marginBottom: '20px'}}>📤 Tic Tac Toe</h1>
        <p style={{fontSize: '18px', margin: '30px 0', color: '#666'}}>
          Please login via the Identity Service to access this game.
        </p>
        <a 
          href={IDENTITY_URL}
          className="button button-primary" 
          style={{
            display: 'inline-block',
            padding: '12px 30px',
            fontSize: '16px',
            textDecoration: 'none',
            borderRadius: '4px',
            backgroundColor: '#3498db',
            color: 'white'
          }}
        >
          Go to Login
        </a>
      </div>
    );
  }

  return (
    <div className="App">
      <style>
        {`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }

          .game-board {
            display: grid;
            grid-template-columns: repeat(3, 100px);
            grid-template-rows: repeat(3, 100px);
            gap: 10px;
            margin: 20px auto;
            max-width: 330px;
          }
          
          .game-cell {
            width: 100px;
            height: 100px;
            font-size: 48px;
            font-weight: bold;
            border: 3px solid #3498db;
            border-radius: 8px;
            background: white;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s;
            touch-action: manipulation;
            user-select: none;
            -webkit-tap-highlight-color: transparent;
          }
          
          .game-cell:hover:not(.filled) {
            background: #ecf0f1;
            transform: scale(1.05);
          }
          
          .game-cell:active {
            transform: scale(0.95);
          }
          
          .game-cell.filled {
            cursor: not-allowed;
          }
          
          .game-cell.x { color: #e74c3c; }
          .game-cell.o { color: #3498db; }
          .game-cell.pending { color: #95a5a6; }
          
          .series-score {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 12px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
          }
          
          .series-score h3 {
            margin: 0 0 15px 0;
            font-size: 20px;
          }
          
          .series-score .scores {
            display: flex;
            justify-content: space-around;
            align-items: center;
            font-size: 24px;
            font-weight: bold;
          }
          
          .series-score .round-info {
            margin-top: 10px;
            font-size: 16px;
            opacity: 0.9;
          }
          
          .rematch-modal {
            text-align: center;
          }
          
          .rematch-modal h2 {
            margin-bottom: 20px;
            color: #2c3e50;
          }
          
          .rematch-buttons {
            display: flex;
            gap: 15px;
            margin-top: 25px;
          }
          
          .rematch-buttons button {
            flex: 1;
            padding: 15px;
            font-size: 18px;
            border-radius: 8px;
            border: none;
            cursor: pointer;
            transition: all 0.2s;
          }
          
          .rematch-yes {
            background: #27ae60;
            color: white;
          }
          
          .rematch-yes:hover {
            background: #229954;
          }
          
          .rematch-no {
            background: #e74c3c;
            color: white;
          }
          
          .rematch-no:hover {
            background: #c0392b;
          }
          
          .rematch-waiting {
            background: #95a5a6;
            color: white;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
          }
          
          .countdown {
            font-size: 48px;
            font-weight: bold;
            color: #e74c3c;
            margin: 20px 0;
          }
          
          @media (max-width: 768px) {
            .game-board {
              grid-template-columns: repeat(3, 80px);
              grid-template-rows: repeat(3, 80px);
              gap: 8px;
            }
            
            .game-cell {
              width: 80px;
              height: 80px;
              font-size: 40px;
            }
          }
        `}
      </style>
      
      <header>
        <div>
          <h1>{config.app_icon} {config.app_name}</h1>
        </div>
        <div className="user-info">
          <span>Welcome, {user.name}!</span>
          <button onClick={handleBackToApps} className="back-to-apps-btn">← Back to Apps</button>
          <button onClick={handleLogout}>Logout</button>
        </div>
      </header>

      <nav className="tabs">
        <button className={view === 'lobby' ? 'active' : ''} onClick={() => setView('lobby')}>
          Lobby {pendingChallenges.length > 0 && <span className="badge badge-active">{pendingChallenges.length}</span>}
        </button>
        {activeGame && activeGame.status === 'active' && (
          <button className={view === 'game' ? 'active' : ''} onClick={() => setView('game')}>
            Active Game
          </button>
        )}
        <button className={view === 'stats' ? 'active' : ''} onClick={() => setView('stats')}>
          Stats & Leaderboard
        </button>
      </nav>

      <main>
        {view === 'optimizing' && (
          <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
            <div style={{
              display: 'inline-block',
              width: '50px',
              height: '50px',
              border: '5px solid #f3f3f3',
              borderTop: '5px solid #3498db',
              borderRadius: '50%',
              animation: 'spin 1s linear infinite'
            }}></div>
            <h2 style={{marginTop: '30px'}}>Optimizing for gameplay...</h2>
            <p style={{color: '#666', fontSize: '16px', marginTop: '10px'}}>
              {websocketStatus === 'connecting' && 'Connecting...'}
              {websocketStatus === 'handshaking' && 'Establishing secure connection...'}
              {websocketStatus === 'reconnecting' && 'Reconnecting...'}
              {websocketStatus === 'error' && 'Connection failed'}
            </p>
          </div>
        )}

        {view === 'lobby' && (
          <>
            <Lobby
              onlineUsers={onlineUsers}
              pendingChallenges={pendingChallenges}
              onChallenge={handleChallenge}
              onRespondToChallenge={handleRespondToChallenge}
              onRefresh={refreshLobby}
            />
            
            {justSentChallenge && (
              <div style={{
                position: 'fixed',
                bottom: '20px',
                right: '20px',
                background: '#3498db',
                color: 'white',
                padding: '15px 25px',
                borderRadius: '8px',
                boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
                display: 'flex',
                alignItems: 'center',
                gap: '12px',
                zIndex: 1000
              }}>
                <div style={{
                  width: '20px',
                  height: '20px',
                  border: '3px solid rgba(255,255,255,0.3)',
                  borderTop: '3px solid white',
                  borderRadius: '50%',
                  animation: 'spin 1s linear infinite'
                }}></div>
                <span style={{ fontSize: '16px', fontWeight: '500' }}>
                  Waiting for opponent's response...
                </span>
              </div>
            )}
          </>
        )}

        {view === 'game' && activeGame && activeGame.status === 'active' && (
          <GameView
            game={activeGame}
            user={user}
            pendingMove={pendingMove}
            onMove={handleMove}
          />
        )}

        {view === 'game' && gameResult && (
          <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.8)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 9999
          }}>
            <div style={{
              background: 'white',
              padding: '40px',
              borderRadius: '12px',
              textAlign: 'center',
              boxShadow: '0 4px 20px rgba(0,0,0,0.3)'
            }}>
              <h2 style={{ fontSize: '32px', marginBottom: '20px' }}>{gameResult}</h2>
              <p style={{ fontSize: '16px', color: '#666' }}>Returning to lobby...</p>
            </div>
          </div>
        )}

        {view === 'stats' && (
          <StatsView
            user={user}
            playerStats={playerStats}
            leaderboard={leaderboard}
            gameHistory={gameHistory}
          />
        )}
      </main>

      {showChallengeModal && selectedOpponent && (
        <ChallengeModal
          opponent={selectedOpponent}
          onSend={handleSendChallenge}
          onCancel={() => {
            setShowChallengeModal(false);
            setJustSentChallenge(false); // Clear WS connection flag
          }}
        />
      )}
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <GameApp />
    </AuthProvider>
  );
}

export default App;
