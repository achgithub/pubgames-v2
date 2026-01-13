import React, { useState, useEffect } from 'react';
import axios from 'axios';
// import './App.css'; // Using shared CSS from Identity Service

// Dynamic URLs for mobile access
const getHostname = () => window.location.hostname;
const API_BASE = `http://${getHostname()}:30021/api`;
const IDENTITY_URL = `http://${getHostname()}:30000`;
const IDENTITY_API = `http://${getHostname()}:3001/api`;

// Setup axios interceptor to add JWT token to all requests
axios.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('jwt_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

function App() {
  const [user, setUser] = useState(null);
  const [currentGame, setCurrentGame] = useState(null);
  const [userGameStatus, setUserGameStatus] = useState(null);
  const [view, setView] = useState('loading');
  const [games, setGames] = useState([]);
  const [matches, setMatches] = useState([]);
  const [rounds, setRounds] = useState([]);
  const [openRounds, setOpenRounds] = useState([]);
  const [selectedRound, setSelectedRound] = useState(null);
  const [roundMatches, setRoundMatches] = useState([]);
  const [usedTeams, setUsedTeams] = useState([]);
  const [standings, setStandings] = useState([]);
  const [predictions, setPredictions] = useState([]);
  const [summary, setSummary] = useState(null);
  const [countdowns, setCountdowns] = useState({});
  const [config, setConfig] = useState({
    venue_name: 'Football Prediction - Last Man Standing',
    logo_url: ''
  });

  useEffect(() => {
    loadConfig();
    
    // SSO: Check for token in URL from Identity Service
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    
    if (token) {
      // Validate token with Identity Service
      validateAndLoginWithToken(token);
      // Clean URL
      window.history.replaceState({}, document.title, window.location.pathname);
      return;
    }
    
    // Otherwise check localStorage
    const savedUser = localStorage.getItem('user');
    const savedToken = localStorage.getItem('jwt_token');
    if (savedUser && savedToken) {
      const u = JSON.parse(savedUser);
      setUser(u);
      loadCurrentGame(u);
      setView('dashboard');
    } else {
      setView('login-required');
    }
  }, []);

  const validateAndLoginWithToken = async (token) => {
    try {
      // Validate token with Identity Service
      const response = await fetch(`${IDENTITY_API}/validate-token`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (!response.ok) {
        console.error('Token validation failed');
        setView('login-required');
        return;
      }
      
      const userData = await response.json();
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      localStorage.setItem('jwt_token', token);
      loadCurrentGame(userData);
      setView('dashboard');
    } catch (error) {
      console.error('SSO error:', error);
      setView('login-required');
    }
  };

  const loadConfig = async () => {
    try {
      const response = await axios.get(`${API_BASE}/config`);
      setConfig(response.data);
    } catch (error) {
      console.log('Using default config');
    }
  };

  // Countdown timer effect
  useEffect(() => {
    if (!rounds || rounds.length === 0) return;

    const updateCountdowns = () => {
      const newCountdowns = {};
      rounds.filter(r => r.status === 'open').forEach(round => {
        const deadline = new Date(round.submission_deadline);
        const now = new Date();
        const diff = deadline - now;

        if (diff <= 0) {
          newCountdowns[round.round_number] = 'CLOSED';
        } else {
          const days = Math.floor(diff / (1000 * 60 * 60 * 24));
          const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
          const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
          const seconds = Math.floor((diff % (1000 * 60)) / 1000);

          if (days > 0) {
            newCountdowns[round.round_number] = `${days}d ${hours}h ${minutes}m`;
          } else if (hours > 0) {
            newCountdowns[round.round_number] = `${hours}h ${minutes}m ${seconds}s`;
          } else {
            newCountdowns[round.round_number] = `${minutes}m ${seconds}s`;
          }
        }
      });
      setCountdowns(newCountdowns);
    };

    updateCountdowns();
    const interval = setInterval(updateCountdowns, 1000);
    return () => clearInterval(interval);
  }, [rounds]);

  // Load rounds and open rounds for countdown on dashboard
  useEffect(() => {
    if (user && !user.is_admin && userGameStatus?.joined && view === 'dashboard') {
      loadRounds();
      loadOpenRounds();
    }
  }, [userGameStatus, view]);

  const loadCurrentGame = async (u = user) => {
    try {
      const response = await axios.get(`${API_BASE}/games/current`);
      setCurrentGame(response.data);
      if (u && !u.is_admin) {
        loadUserGameStatus(u.id, response.data.id);
      }
    } catch (error) {
      console.log('No current game');
      setCurrentGame(null);
    }
  };

  const loadUserGameStatus = async (userId, gameId) => {
    try {
      const response = await axios.get(`${API_BASE}/games/status?game_id=${gameId || ''}`);
      setUserGameStatus(response.data);
    } catch (error) {
      console.log('Error loading game status');
    }
  };

  const handleBackToApps = () => {
    // Just redirect to Identity Service dashboard (stay logged in)
    window.location.href = IDENTITY_URL;
  };

  const handleLogout = () => {
    // Clear all state
    setUser(null);
    setCurrentGame(null);
    setUserGameStatus(null);
    setGames([]);
    setMatches([]);
    setRounds([]);
    setOpenRounds([]);
    setSelectedRound(null);
    setRoundMatches([]);
    setUsedTeams([]);
    setStandings([]);
    setPredictions([]);
    setSummary(null);
    setCountdowns({});
    
    // Clear local storage
    localStorage.removeItem('user');
    localStorage.removeItem('jwt_token');
    
    // Small delay to let React cleanup finish
    setTimeout(() => {
      window.location.href = `${IDENTITY_URL}?logout=true`;
    }, 100);
  };

  const joinCurrentGame = async () => {
    try {
      await axios.post(`${API_BASE}/games/join`, {
        game_id: currentGame.id
      });
      alert('Successfully joined the game!');
      loadUserGameStatus(user.id, currentGame.id);
    } catch (error) {
      alert('Failed to join game');
    }
  };

  const loadGames = async () => {
    const response = await axios.get(`${API_BASE}/games`);
    setGames(response.data || []);
  };

  const createGame = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    try {
      await axios.post(`${API_BASE}/games`, {
        name: formData.get('name'),
        postponement_rule: formData.get('postponement_rule') || 'loss'
      });
      alert('Game created successfully!');
      loadGames();
      e.target.reset();
    } catch (error) {
      alert('Failed to create game');
    }
  };

  const setAsCurrentGame = async (gameId) => {
    try {
      await axios.put(`${API_BASE}/games/${gameId}/set-current`);
      alert('Current game updated!');
      loadCurrentGame();
      loadGames();
    } catch (error) {
      alert('Failed to set current game');
    }
  };

  const completeGame = async (gameId) => {
    if (!window.confirm('Complete this game? Active players will be declared winners.')) return;
    try {
      const response = await axios.put(`${API_BASE}/games/${gameId}/complete`);
      alert(`Game completed! ${response.data.winners} winner(s)`);
      loadGames();
      loadCurrentGame();
    } catch (error) {
      alert(error.response?.data || 'Failed to complete game');
    }
  };

  const loadMatches = async () => {
    if (!currentGame) return;
    const response = await axios.get(`${API_BASE}/matches?game_id=${currentGame.id}`);
    setMatches(response.data || []);
  };

  const loadRounds = async () => {
    if (!currentGame) return;
    const response = await axios.get(`${API_BASE}/rounds?game_id=${currentGame.id}`);
    setRounds(response.data || []);
  };

  const loadOpenRounds = async () => {
    if (!currentGame) return;
    const response = await axios.get(`${API_BASE}/rounds/open?game_id=${currentGame.id}`);
    setOpenRounds(response.data || []);
    // Also load full rounds data for deadline info
    await loadRounds();
  };

  const loadRoundMatches = async (round) => {
    if (!currentGame) return;
    const response = await axios.get(`${API_BASE}/matches/${currentGame.id}/round/${round}`);
    setRoundMatches(response.data || []);
    setSelectedRound(round);
    loadUsedTeams();
  };

  const loadUsedTeams = async () => {
    if (!currentGame || !user) return;
    try {
      const response = await axios.get(`${API_BASE}/predictions/used-teams?game_id=${currentGame.id}`);
      setUsedTeams(response.data || []);
    } catch (error) {
      console.log('Error loading used teams');
      setUsedTeams([]);
    }
  };

  const makePrediction = async (matchId, team) => {
    if (usedTeams.includes(team)) {
      alert('You have already picked this team in this game!');
      return;
    }
    try {
      await axios.post(`${API_BASE}/predictions`, {
        game_id: currentGame.id,
        match_id: matchId,
        predicted_team: team
      });
      alert('Prediction submitted successfully!');
      loadOpenRounds();
      setSelectedRound(null);
    } catch (error) {
      alert(error.response?.data || 'Prediction failed');
    }
  };

  const loadStandings = async () => {
    if (!currentGame) return;
    const response = await axios.get(`${API_BASE}/standings?game_id=${currentGame.id}`);
    setStandings(response.data || []);
  };

  const loadPredictions = async (viewAll = false) => {
    if (!currentGame) return;
    try {
      const viewAllParam = viewAll ? '?view_all=true' : '';
      const response = await axios.get(`${API_BASE}/predictions${viewAllParam}&game_id=${currentGame.id}`);
      setPredictions(response.data || []);
    } catch (error) {
      console.error('Failed to load predictions:', error);
      alert(viewAll ? 'Only admins can view all predictions' : 'Failed to load predictions');
      setPredictions([]);
    }
  };

  const loadRoundSummary = async (roundNum) => {
    if (!currentGame) return;
    try {
      const response = await axios.get(`${API_BASE}/rounds/${currentGame.id}/${roundNum}/summary`);
      setSummary(response.data);
    } catch (error) {
      console.error('Failed to load round summary:', error);
      alert('Failed to load round summary');
    }
  };

  const uploadMatches = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('game_id', currentGame.id);

    try {
      const response = await axios.post(`${API_BASE}/matches/upload`, formData);
      alert(`${response.data.uploaded} matches uploaded successfully!`);
      loadMatches();
      loadRounds();
    } catch (error) {
      alert('Upload failed');
    }
  };

  const createRound = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    try {
      await axios.post(`${API_BASE}/rounds`, {
        game_id: currentGame.id,
        round_number: parseInt(formData.get('round_number')),
        submission_deadline: formData.get('deadline')
      });
      alert('Round created successfully!');
      loadRounds();
      e.target.reset();
    } catch (error) {
      alert('Failed to create round');
    }
  };

  const updateRoundStatus = async (roundNum, status) => {
    try {
      await axios.put(`${API_BASE}/rounds/${currentGame.id}/${roundNum}/status`, { status });
      alert(`Round ${roundNum} ${status}!`);
      loadRounds();
    } catch (error) {
      alert('Failed to update round status');
    }
  };

  const updateResult = async (matchId, result) => {
    try {
      await axios.put(`${API_BASE}/matches/${matchId}/result`, { result });
      alert('Result updated successfully!');
      loadMatches();
      loadStandings();
      if (user && !user.is_admin) {
        loadUserGameStatus(user.id, currentGame.id);
      }
    } catch (error) {
      alert('Update failed');
    }
  };

  const closeRoundSubmission = async (roundNum) => {
    if (!window.confirm(`Close submissions for Round ${roundNum}?`)) return;
    await updateRoundStatus(roundNum, 'closed');
    loadRoundSummary(roundNum);
    setView('summary');
  };

  // ============================================================================
  // LOADING VIEW - While checking SSO
  // ============================================================================
  
  if (view === 'loading') {
    return (
      <>
        <style>
          {`:root {
            --game-color: #2ecc71;
            --game-accent: #27ae60;
          }`}
        </style>
        <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
          <h2>Loading...</h2>
        </div>
      </>
    );
  }
  
  // ============================================================================
  // LOGIN REQUIRED VIEW - SSO via Identity Service
  // ============================================================================
  
  if (view === 'login-required') {
    return (
      <>
        <style>
          {`:root {
            --game-color: #2ecc71;
            --game-accent: #27ae60;
          }`}
        </style>
        <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
          <h1 style={{marginBottom: '20px'}}>⚽ Last Man Standing</h1>
          <p style={{fontSize: '18px', margin: '30px 0', color: '#666'}}>
            Please login via the Identity Service to access this app.
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
              backgroundColor: '#2ecc71',
              color: 'white'
            }}
          >
            Go to Login
          </a>
        </div>
      </>
    );
  }

  // ============================================================================
  // MAIN APP - Dashboard and all views
  // ============================================================================
  
  if (user) {
    return (
      <div className="App">
      <header>
        <div>
          {config.logo_url && (
            <img src={config.logo_url} alt="Logo" className="header-logo" />
          )}
          <h1>⚽ {config.venue_name}</h1>
          {currentGame && (
            <div className="current-game-badge">
              {currentGame.name} {currentGame.status === 'completed' && '(Completed)'}
            </div>
          )}
        </div>
        <div className="user-info">
          <span>Welcome, {user.name}! {user.is_admin && <span className="admin-badge">ADMIN</span>}</span>
          {!user.is_admin && userGameStatus && (
            <span className={userGameStatus.is_active ? 'status-active' : 'status-eliminated'}>
              {userGameStatus.is_active ? '✓ Active' : '✗ Eliminated'}
            </span>
          )}
          <button onClick={handleBackToApps} className="back-to-apps-btn">← Back to Apps</button>
          <button onClick={handleLogout}>Logout</button>
        </div>
      </header>

      <nav className="tabs">
        <button className={view === 'dashboard' ? 'active' : ''} onClick={() => {
          setView('dashboard');
          if (!user.is_admin && userGameStatus?.joined) {
            loadRounds();
            loadOpenRounds();
          }
        }}>
          Dashboard
        </button>
        {!user.is_admin && userGameStatus?.joined && (
          <>
            <button className={view === 'predict' ? 'active' : ''} onClick={() => {
              setView('predict');
              loadOpenRounds();
            }}>
              Make Prediction
            </button>
            <button className={view === 'my-predictions' ? 'active' : ''} onClick={() => {
              setView('my-predictions');
              loadPredictions(false);
            }}>
              My Predictions
            </button>
            <button className={view === 'round-summary' ? 'active' : ''} onClick={() => {
              setView('round-summary');
              loadRounds();
              setSummary(null);
            }}>
              Round Summaries
            </button>
          </>
        )}
        {user.is_admin && (
          <>
            <button className={view === 'standings' ? 'active' : ''} onClick={() => {
              setView('standings');
              loadStandings();
            }}>
              Standings
            </button>
            <button className={view === 'games' ? 'active' : ''} onClick={() => {
              setView('games');
              loadGames();
            }}>
              Manage Games
            </button>
            <button className={view === 'rounds' ? 'active' : ''} onClick={() => {
              setView('rounds');
              loadRounds();
              loadMatches();
            }}>
              Manage Rounds
            </button>
            <button className={view === 'predictions' ? 'active' : ''} onClick={() => {
              setView('predictions');
              loadPredictions(true);
            }}>
              All Predictions
            </button>
            <button className={view === 'admin' ? 'active' : ''} onClick={() => {
              setView('admin');
              loadMatches();
            }}>
              Admin Tools
            </button>
          </>
        )}
      </nav>

      <main>
        {view === 'dashboard' && (
          <div className="dashboard">
            <h2>Welcome to Last Man Standing!</h2>
            
            {!user.is_admin && !userGameStatus?.joined && currentGame && currentGame.status === 'active' && (
              <div className="join-game-prompt">
                <h3>🎮 Join {currentGame.name}</h3>
                <p>You're not in this game yet. Click below to join!</p>
                <button onClick={joinCurrentGame} className="cta-button">
                  Join Game
                </button>
              </div>
            )}

            {!user.is_admin && userGameStatus?.joined && currentGame?.status === 'completed' && (
              <div className="game-completed">
                <h3>🏆 Game Completed!</h3>
                <p>This game has ended. Wait for the admin to start a new game.</p>
              </div>
            )}

            <div className="rules">
              <h3>How to Play:</h3>
              <ul>
                <li>Each competition is a separate <strong>Game</strong></li>
                <li>Admin creates rounds with submission deadlines</li>
                <li>Each round, pick ONE team you think will win</li>
                <li>If your team wins, you advance to the next round</li>
                <li>If your team draws or loses, you're eliminated from that game</li>
                <li>When a game ends, eliminated players can join new games!</li>
                <li>Last player standing wins the game!</li>
              </ul>
            </div>

            {!user.is_admin && userGameStatus?.joined && rounds.filter(r => r.status === 'open').length > 0 && (
              <div className="open-rounds-countdown">
                <h3>⏰ Open Rounds</h3>
                {rounds.filter(r => r.status === 'open').map(round => {
                  const hasSubmitted = openRounds && !openRounds.includes(round.round_number);
                  return (
                    <div key={round.round_number} className={'countdown-card' + (hasSubmitted ? ' submitted' : '')}>
                      <div className="countdown-info">
                        <span className="round-label">Round {round.round_number}</span>
                        <span className="countdown-time">
                          {countdowns[round.round_number] === 'CLOSED' ? (
                            <span className="expired">⚠️ EXPIRED</span>
                          ) : (
                            countdowns[round.round_number] || 'Loading...'
                          )}
                        </span>
                      </div>
                      {hasSubmitted ? (
                        <span className="status-badge submitted-badge">✓ Submitted</span>
                      ) : (
                        <span className="status-badge pending-badge">⏳ Pending</span>
                      )}
                    </div>
                  );
                })}
              </div>
            )}

            {!user.is_admin && userGameStatus?.is_active && (
              <div className="cta">
                <button onClick={() => {
                  setView('predict');
                  loadOpenRounds();
                }} className="cta-button">
                  Make Your Prediction
                </button>
              </div>
            )}

            {!user.is_admin && userGameStatus?.joined && !userGameStatus?.is_active && currentGame?.status === 'active' && (
              <div className="eliminated-message">
                <h3>You've been eliminated from {currentGame.name}</h3>
                <p>Wait for the next game to start fresh!</p>
              </div>
            )}

            {user.is_admin && (
              <div className="admin-dashboard">
                <h3>Admin Quick Actions</h3>
                {!currentGame ? (
                  <div className="warning-box">
                    <p>⚠️ No current game set. Create a game first!</p>
                    <button onClick={() => setView('games')}>Create Game</button>
                  </div>
                ) : (
                  <>
                    <button onClick={() => setView('games')}>Manage Games</button>
                    <button onClick={() => setView('rounds')}>Manage Rounds</button>
                    <button onClick={() => setView('admin')}>Upload Matches</button>
                    <button onClick={() => {
                      setView('standings');
                      loadStandings();
                    }}>View Standings</button>
                  </>
                )}
              </div>
            )}
          </div>
        )}

        {view === 'predict' && (
          <div className="predict">
            <h2>Make Your Prediction - {currentGame?.name}</h2>
            {!userGameStatus?.is_active ? (
              <p className="warning">You have been eliminated from this game.</p>
            ) : openRounds.length === 0 ? (
              <p>No rounds currently open for prediction. Check back later!</p>
            ) : !selectedRound ? (
              <div className="rounds-list">
                <h3>Select a Round:</h3>
                {openRounds.map(round => {
                  const roundInfo = rounds.find(r => r.round_number === round);
                  const deadline = roundInfo ? new Date(roundInfo.submission_deadline) : null;
                  const isExpired = deadline && new Date() > deadline;
                  
                  return (
                    <button 
                      key={round} 
                      onClick={() => !isExpired && loadRoundMatches(round)} 
                      className={'round-btn' + (isExpired ? ' round-expired' : '')}
                      disabled={isExpired}
                    >
                      Round {round}
                      {isExpired && <span style={{fontSize: '12px', display: 'block', marginTop: '5px'}}>⚠️ EXPIRED</span>}
                      {!isExpired && countdowns[round] && (
                        <span style={{fontSize: '12px', display: 'block', marginTop: '5px'}}>
                          ⏰ {countdowns[round]}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            ) : (
              <div className="games-list">
                {(() => {
                  const roundInfo = rounds.find(r => r.round_number === selectedRound);
                  const deadline = roundInfo ? new Date(roundInfo.submission_deadline) : null;
                  const isExpired = deadline && new Date() > deadline;
                  
                  if (isExpired) {
                    return (
                      <div>
                        <button onClick={() => setSelectedRound(null)} className="back-btn">← Back to Rounds</button>
                        <div className="warning-box" style={{marginTop: '20px'}}>
                          <p>⚠️ This round has expired. You can no longer submit predictions.</p>
                          <button onClick={() => setSelectedRound(null)}>Back to Rounds</button>
                        </div>
                      </div>
                    );
                  }
                  
                  return (
                    <>
                      <h3>Round {selectedRound} - Select Your Team:</h3>
                      <button onClick={() => setSelectedRound(null)} className="back-btn">← Back to Rounds</button>
                      {deadline && (
                        <div className="deadline-warning">
                          <strong>⏰ Time Remaining:</strong> {countdowns[selectedRound] || 'Loading...'}
                        </div>
                      )}
                      {usedTeams.length > 0 && (
                        <div className="used-teams-notice">
                          <strong>Already used:</strong> {usedTeams.join(', ')}
                        </div>
                      )}
                      {roundMatches.map(match => {
                        const homeUsed = usedTeams.includes(match.home_team);
                        const awayUsed = usedTeams.includes(match.away_team);
                        return (
                          <div key={match.id} className="game-card">
                            <div className="game-info">
                              <div className="game-date">{match.date}</div>
                              <div className="game-location">{match.location}</div>
                            </div>
                            <div className="teams">
                              <button 
                                onClick={() => {
                                  const currentDeadline = roundInfo ? new Date(roundInfo.submission_deadline) : null;
                                  if (currentDeadline && new Date() > currentDeadline) {
                                    alert('⚠️ Deadline has passed! You can no longer submit predictions for this round.');
                                    setSelectedRound(null);
                                    return;
                                  }
                                  if (!homeUsed) makePrediction(match.id, match.home_team);
                                }}
                                className={'team-btn home' + (homeUsed ? ' team-used' : '')}
                                disabled={homeUsed}
                              >
                                {match.home_team}
                                {homeUsed && <span className="used-badge">✗ Used</span>}
                              </button>
                              <span className="vs">VS</span>
                              <button 
                                onClick={() => {
                                  const currentDeadline = roundInfo ? new Date(roundInfo.submission_deadline) : null;
                                  if (currentDeadline && new Date() > currentDeadline) {
                                    alert('⚠️ Deadline has passed! You can no longer submit predictions for this round.');
                                    setSelectedRound(null);
                                    return;
                                  }
                                  if (!awayUsed) makePrediction(match.id, match.away_team);
                                }}
                                className={'team-btn away' + (awayUsed ? ' team-used' : '')}
                                disabled={awayUsed}
                              >
                                {match.away_team}
                                {awayUsed && <span className="used-badge">✗ Used</span>}
                              </button>
                            </div>
                          </div>
                        );
                      })}
                    </>
                  );
                })()}
              </div>
            )}
          </div>
        )}

        {view === 'my-predictions' && !user.is_admin && (
          <div className="predictions">
            <h2>My Predictions - {currentGame?.name}</h2>
            {predictions.length === 0 ? (
              <p className="info-text">You haven't made any predictions yet. Go to "Make Prediction" to get started!</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Round</th>
                    <th>Date</th>
                    <th>Match</th>
                    <th>My Pick</th>
                    <th>Result</th>
                    <th>Outcome</th>
                  </tr>
                </thead>
                <tbody>
                  {predictions.map(pred => (
                    <tr key={pred.id} className={pred.voided ? 'voided-row' : ''}>
                      <td>Round {pred.round_number}</td>
                      <td>{pred.match_date || 'N/A'}</td>
                      <td>{pred.home_team} vs {pred.away_team}</td>
                      <td><strong>{pred.predicted_team}</strong></td>
                      <td>{pred.result || 'Pending'}</td>
                      <td>
                        {pred.voided ? (
                          <span className="badge-voided">VOIDED</span>
                        ) : pred.is_correct === null ? (
                          <span className="badge-pending">Pending</span>
                        ) : pred.is_correct ? (
                          <span className="badge-correct">✓ Correct</span>
                        ) : (
                          <span className="badge-wrong">✗ Wrong</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {view === 'round-summary' && !user.is_admin && (
          <div className="round-summary">
            <h2>Round Summaries - {currentGame?.name}</h2>
            {!summary ? (
              <div>
                <p className="info-text">Select a round to view its summary</p>
                <div className="rounds-list">
                  {rounds.filter(r => {
                    if (r.status === 'closed') return true;
                    if (r.status === 'open') {
                      const deadline = new Date(r.submission_deadline);
                      return new Date() > deadline;
                    }
                    return false;
                  }).map(round => (
                    <button 
                      key={round.id} 
                      onClick={() => loadRoundSummary(round.round_number)} 
                      className="round-btn"
                    >
                      Round {round.round_number}
                      {round.status === 'open' && <span style={{fontSize: '12px', display: 'block'}}>⏳ Awaiting Results</span>}
                    </button>
                  ))}
                  {rounds.filter(r => r.status === 'closed' || (r.status === 'open' && new Date() > new Date(r.submission_deadline))).length === 0 && (
                    <p>No rounds available yet.</p>
                  )}
                </div>
              </div>
            ) : (
              <div>
                <button onClick={() => setSummary(null)} className="back-btn">← Back to Round Selection</button>
                
                <h3>Round {summary.round_number} Summary</h3>
                
                {summary.total_players === 0 ? (
                  <p className="info-text">⏳ Awaiting match results to calculate statistics</p>
                ) : (
                  <>
                    <div className="summary-stats">
                      <div className="stat-card">
                        <h3>{summary.total_players}</h3>
                        <p>Total Players</p>
                      </div>
                    </div>

                    <div className="team-breakdown">
                      <h3>Team Selections</h3>
                      {summary.team_stats && summary.team_stats.length > 0 ? (
                        <table>
                          <thead>
                            <tr>
                              <th>Team</th>
                              <th>Players Selected</th>
                            </tr>
                          </thead>
                          <tbody>
                            {summary.team_stats.map(stat => (
                              <tr key={stat.team_name}>
                                <td><strong>{stat.team_name}</strong></td>
                                <td>{stat.player_count}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      ) : (
                        <p className="info-text">⏳ Awaiting match results</p>
                      )}
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        )}

        {view === 'standings' && (
          <div className="standings">
            <h2>Standings - {currentGame?.name}</h2>
            <table>
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Player</th>
                  <th>Status</th>
                  <th>Last Round</th>
                </tr>
              </thead>
              <tbody>
                {standings.map((entry, index) => (
                  <tr key={entry.user_id} className={entry.is_active ? 'active' : 'eliminated'}>
                    <td>{index + 1}</td>
                    <td>{entry.user_name}</td>
                    <td>
                      <span className={entry.is_active ? 'badge-active' : 'badge-eliminated'}>
                        {entry.is_active ? 'Active' : 'Eliminated'}
                      </span>
                    </td>
                    <td>{entry.last_round || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {view === 'games' && user.is_admin && (
          <div className="games-management">
            <h2>Game Management</h2>
            
            <div className="admin-section">
              <h3>Create New Game</h3>
              <form onSubmit={createGame} className="inline-form">
                <input type="text" name="name" placeholder="Game Name (e.g., Game 2, Spring 2026)" required />
                <select name="postponement_rule">
                  <option value="loss">Postponed = Loss (Default)</option>
                  <option value="win">Postponed = Win</option>
                </select>
                <button type="submit">Create Game</button>
              </form>
              <p className="hint">
                New games start with no players. Eliminated players from previous games can join new ones.<br/>
                <strong>Postponement Rule:</strong> If a match is marked as "P - P", all players either advance (Win) or are eliminated (Loss).
              </p>
            </div>

            <div className="admin-section">
              <h3>All Games</h3>
              <table>
                <thead>
                  <tr>
                    <th>Game Name</th>
                    <th>Status</th>
                    <th>Postponement Rule</th>
                    <th>Winners</th>
                    <th>Start Date</th>
                    <th>End Date</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {games.map(game => (
                    <tr key={game.id} className={currentGame?.id === game.id ? 'current-game-row' : ''}>
                      <td>
                        <strong>{game.name}</strong>
                        {currentGame?.id === game.id && <span className="badge-current">CURRENT</span>}
                      </td>
                      <td>
                        <span className={'badge-' + game.status}>
                          {game.status.toUpperCase()}
                        </span>
                      </td>
                      <td>
                        <span style={{fontSize: '12px'}}>
                          P-P = {game.postponement_rule === 'win' ? '✓ Win' : '✗ Loss'}
                        </span>
                      </td>
                      <td>{game.winner_count || '-'}</td>
                      <td>{new Date(game.start_date).toLocaleDateString()}</td>
                      <td>{game.end_date ? new Date(game.end_date).toLocaleDateString() : '-'}</td>
                      <td>
                        {game.status === 'active' && currentGame?.id !== game.id && (
                          <button onClick={() => setAsCurrentGame(game.id)} className="btn-info">
                            Set as Current
                          </button>
                        )}
                        {game.status === 'active' && currentGame?.id === game.id && (
                          <button onClick={() => completeGame(game.id)} className="btn-warning">
                            Complete Game
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {view === 'rounds' && user.is_admin && (
          <div className="rounds-management">
            <h2>Round Management - {currentGame?.name}</h2>
            
            {!currentGame && (
              <p className="warning">No current game selected. Create or select a game first.</p>
            )}

            {currentGame && (
              <>
                <div className="admin-section">
                  <h3>Create New Round</h3>
                  <form onSubmit={createRound} className="inline-form">
                    <input type="number" name="round_number" placeholder="Round Number" required min="1" />
                    <input type="datetime-local" name="deadline" required />
                    <button type="submit">Create Round</button>
                  </form>
                </div>

                <div className="admin-section">
                  <h3>Existing Rounds</h3>
                  <table>
                    <thead>
                      <tr>
                        <th>Round</th>
                        <th>Submission Deadline</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rounds.map(round => (
                        <tr key={round.id}>
                          <td>Round {round.round_number}</td>
                          <td>{new Date(round.submission_deadline).toLocaleString()}</td>
                          <td>
                            <span className={'badge-' + round.status}>
                              {round.status.toUpperCase()}
                            </span>
                          </td>
                          <td>
                            {round.status === 'draft' && (
                              <button onClick={() => updateRoundStatus(round.round_number, 'open')} className="btn-success">
                                Open for Predictions
                              </button>
                            )}
                            {round.status === 'open' && (
                              <button onClick={() => closeRoundSubmission(round.round_number)} className="btn-warning">
                                Close & View Summary
                              </button>
                            )}
                            {round.status === 'closed' && (
                              <button onClick={() => {
                                loadRoundSummary(round.round_number);
                                setView('summary');
                              }} className="btn-info">
                                View Summary
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <div className="admin-section">
                  <h3>Matches by Round</h3>
                  {rounds.map(round => {
                    const roundMatches = matches.filter(m => m.round_number === round.round_number);
                    if (roundMatches.length === 0) return null;
                    return (
                      <div key={round.id} className="round-games">
                        <h4>Round {round.round_number}</h4>
                        <table className="compact-table">
                          <thead>
                            <tr>
                              <th>Match</th>
                              <th>Teams</th>
                              <th>Date</th>
                              <th>Result</th>
                              <th>Status</th>
                            </tr>
                          </thead>
                          <tbody>
                            {roundMatches.map(match => (
                              <tr key={match.id}>
                                <td>{match.match_number}</td>
                                <td>{match.home_team} vs {match.away_team}</td>
                                <td>{match.date}</td>
                                <td>{match.result || '-'}</td>
                                <td><span className={'badge-' + match.status}>{match.status}</span></td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    );
                  })}
                </div>
              </>
            )}
          </div>
        )}

        {view === 'summary' && user.is_admin && summary && (
          <div className="round-summary">
            <h2>Round {summary.round_number} Summary - {currentGame?.name}</h2>
            
            <div className="summary-stats">
              <div className="stat-card">
                <h3>{summary.total_players}</h3>
                <p>Total Players</p>
              </div>
              <div className="stat-card eliminated">
                <h3>{summary.players_eliminated}</h3>
                <p>Players Eliminated</p>
              </div>
              <div className="stat-card active">
                <h3>{summary.total_players - summary.players_eliminated}</h3>
                <p>Players Remaining</p>
              </div>
            </div>

            <div className="team-breakdown">
              <h3>Team Selections</h3>
              <table>
                <thead>
                  <tr>
                    <th>Team</th>
                    <th>Players Selected</th>
                    <th>Players Eliminated</th>
                    <th>Players Advanced</th>
                  </tr>
                </thead>
                <tbody>
                  {summary.team_stats.map(stat => (
                    <tr key={stat.team_name}>
                      <td><strong>{stat.team_name}</strong></td>
                      <td>{stat.player_count}</td>
                      <td className="eliminated-text">{stat.players_eliminated}</td>
                      <td className="active-text">{stat.player_count - stat.players_eliminated}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <button onClick={() => setView('rounds')} className="back-btn">
              ← Back to Round Management
            </button>
          </div>
        )}

        {view === 'predictions' && user.is_admin && (
          <div className="predictions">
            <h2>All Predictions (Admin View) - {currentGame?.name}</h2>
            <p className="info-text">Only admins can see all player predictions</p>
            <table>
              <thead>
                <tr>
                  <th>Round</th>
                  <th>Player</th>
                  <th>Date</th>
                  <th>Match</th>
                  <th>Predicted</th>
                  <th>Result</th>
                  <th>Outcome</th>
                </tr>
              </thead>
              <tbody>
                {predictions.map(pred => (
                  <tr key={pred.id} className={pred.voided ? 'voided-row' : ''}>
                    <td>{pred.round_number}</td>
                    <td>{pred.user_name}</td>
                    <td>{pred.match_date}</td>
                    <td>{pred.home_team} vs {pred.away_team}</td>
                    <td><strong>{pred.predicted_team}</strong></td>
                    <td>{pred.result || 'Pending'}</td>
                    <td>
                      {pred.voided ? (
                        <span className="badge-voided">VOIDED</span>
                      ) : pred.is_correct === null ? (
                        <span className="badge-pending">Pending</span>
                      ) : pred.is_correct ? (
                        <span className="badge-correct">✓ Correct</span>
                      ) : (
                        <span className="badge-wrong">✗ Wrong</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {view === 'admin' && user.is_admin && (
          <div className="admin">
            <h2>Admin Tools - {currentGame?.name}</h2>
            
            {!currentGame && (
              <p className="warning">No current game selected. Create or select a game first.</p>
            )}

            {currentGame && (
              <>
                <div className="admin-section">
                  <h3>Upload Matches</h3>
                  <input type="file" accept=".csv" onChange={uploadMatches} />
                  <p className="hint">
                    CSV format: Match Number, Round Number, Date, Location, Home Team, Away Team, Result<br/>
                    Leave Result empty for upcoming matches. If result is present, it will be auto-evaluated.<br/>
                    <strong>For postponed matches, enter: P - P</strong>
                  </p>
                </div>

                <div className="admin-section">
                  <h3>Set Results Manually</h3>
                  <p className="hint" style={{marginBottom: '15px'}}>
                    Enter results as "2 - 1" or "P - P" for postponed matches
                  </p>
                  <table>
                    <thead>
                      <tr>
                        <th>Round</th>
                        <th>Match</th>
                        <th>Date</th>
                        <th>Teams</th>
                        <th>Result</th>
                        <th>Status</th>
                        <th>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {matches.filter(m => m.status === 'upcoming').map(match => (
                        <tr key={match.id}>
                          <td>{match.round_number}</td>
                          <td>{match.match_number}</td>
                          <td>{match.date}</td>
                          <td>{match.home_team} vs {match.away_team}</td>
                          <td>{match.result || '-'}</td>
                          <td>
                            <span className={'badge-' + match.status}>{match.status}</span>
                          </td>
                          <td>
                            <button onClick={() => {
                              const result = prompt('Enter result (e.g., "2 - 1" or "P - P" for postponed):');
                              if (result) updateResult(match.id, result);
                            }}>
                              Set Result
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <div className="admin-section">
                  <h3>Edit Completed Results</h3>
                  <p className="hint" style={{marginBottom: '15px'}}>
                    Correct any mistakes before closing the round
                  </p>
                  {matches.filter(m => m.status === 'completed').length === 0 ? (
                    <p className="info-text">No completed matches yet</p>
                  ) : (
                    <table>
                      <thead>
                        <tr>
                          <th>Round</th>
                          <th>Match</th>
                          <th>Date</th>
                          <th>Teams</th>
                          <th>Result</th>
                          <th>Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {matches.filter(m => m.status === 'completed').map(match => (
                          <tr key={match.id}>
                            <td>{match.round_number}</td>
                            <td>{match.match_number}</td>
                            <td>{match.date}</td>
                            <td>{match.home_team} vs {match.away_team}</td>
                            <td><strong>{match.result}</strong></td>
                            <td>
                              <button onClick={() => {
                                const result = prompt(`Current result: ${match.result}\n\nEnter new result:`, match.result);
                                if (result && result !== match.result) updateResult(match.id, result);
                              }} className="btn-warning">
                                Edit Result
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              </>
            )}
          </div>
        )}
      </main>
    </div>
  );
  }
  
  // Fallback
  return null;
}

export default App;
