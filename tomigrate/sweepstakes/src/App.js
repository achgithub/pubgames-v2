import React, { useState, useEffect } from 'react';
import axios from 'axios';
// import './App.css'; // Using shared CSS

const API_BASE = 'http://localhost:30021/api';

function App() {
  const [user, setUser] = useState(null);
  const [config, setConfig] = useState({});
  const [view, setView] = useState('login');
  const [competitions, setCompetitions] = useState([]);
  const [selectedCompetition, setSelectedCompetition] = useState(null);
  const [entries, setEntries] = useState([]);
  const [myDraws, setMyDraws] = useState([]);
  const [blindBoxes, setBlindBoxes] = useState([]);
  const [availableCount, setAvailableCount] = useState(0);
  const [selectionLock, setSelectionLock] = useState(null);
  const [lockCheckInterval, setLockCheckInterval] = useState(null);
  const [showSpinner, setShowSpinner] = useState(false);
  const [spinnerResult, setSpinnerResult] = useState(null);
  const [leaderboardData, setLeaderboardData] = useState([]);
  const [leaderboardComp, setLeaderboardComp] = useState(null);

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
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      const u = JSON.parse(storedUser);
      setUser(u);
      setView(u.is_admin ? 'admin-dashboard' : 'dashboard');
    }
  }, []);

  const validateAndLoginWithToken = async (token) => {
    try {
      // Validate token with Identity Service
      const response = await fetch('http://localhost:3001/api/validate-token', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (!response.ok) {
        console.error('Token validation failed');
        setView('login');
        return;
      }
      
      const userData = await response.json();
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      setView(userData.is_admin ? 'admin-dashboard' : 'dashboard');
    } catch (error) {
      console.error('SSO error:', error);
      setView('login');
    }
  };

  useEffect(() => {
    if (user) {
      loadCompetitions();
      if (view === 'dashboard' || view === 'my-entries' || view === 'pick-box' || view === 'leaderboard') {
        loadMyDraws();
      }
      if (selectedCompetition && view === 'pick-box') {
        loadBlindBoxes();
        loadAvailableCount();
        checkSelectionLock();
      }
      if (leaderboardComp) {
        loadLeaderboard(leaderboardComp.id);
      }
    }
    
    return () => {
      if (lockCheckInterval) {
        clearInterval(lockCheckInterval);
      }
    };
  }, [user, view, selectedCompetition, leaderboardComp]);

  const loadConfig = async () => {
    try {
      const res = await axios.get(`${API_BASE}/config`);
      setConfig(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  const loadCompetitions = async () => {
    try {
      const res = await axios.get(`${API_BASE}/competitions`);
      setCompetitions(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const loadEntries = async (compId) => {
    try {
      const res = await axios.get(`${API_BASE}/competitions/${compId}/entries`);
      setEntries(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const loadMyDraws = async () => {
    if (!user) return;
    try {
      const res = await axios.get(`${API_BASE}/draws?user_id=${user.id}`);
      setMyDraws(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const loadBlindBoxes = async () => {
    if (!selectedCompetition || !user) return;
    
    const compDraws = myDraws.filter(d => d.competition_id === selectedCompetition.id);
    if (compDraws.length > 0) {
      setBlindBoxes([]);
      return;
    }
    
    try {
      const res = await axios.get(`${API_BASE}/competitions/${selectedCompetition.id}/blind-boxes?user_id=${user.id}`);
      setBlindBoxes(res.data || []);
    } catch (err) {
      console.error('Error loading blind boxes:', err);
      setBlindBoxes([]);
    }
  };

  const loadAvailableCount = async () => {
    if (!selectedCompetition) return;
    try {
      const res = await axios.get(`${API_BASE}/competitions/${selectedCompetition.id}/available-count`);
      setAvailableCount(res.data?.count || 0);
    } catch (err) {
      console.error(err);
    }
  };

  const loadLeaderboard = async (compId) => {
    try {
      const res = await axios.get(`${API_BASE}/competitions/${compId}/all-draws`);
      setLeaderboardData(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpdatePosition = async (compId, entryId, position) => {
    try {
      await axios.post(`${API_BASE}/competitions/${compId}/update-position`, {
        entry_id: entryId,
        position: position
      });
      loadEntries(compId);
    } catch (err) {
      alert('Failed to update position: ' + (err.response?.data || err.message));
    }
  };

  const checkSelectionLock = async () => {
    if (!selectedCompetition || !user) return;
    
    try {
      const res = await axios.get(`${API_BASE}/competitions/${selectedCompetition.id}/lock-status?user_id=${user.id}`);
      setSelectionLock(res.data);
      
      if (res.data.locked && !res.data.is_me) {
        if (!lockCheckInterval) {
          const interval = setInterval(checkSelectionLock, 5000);
          setLockCheckInterval(interval);
        }
      } else {
        if (lockCheckInterval) {
          clearInterval(lockCheckInterval);
          setLockCheckInterval(null);
        }
      }
    } catch (err) {
      console.error('Error checking lock:', err);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    try {
      const res = await axios.post(`${API_BASE}/register`, {
        email: formData.get('email'),
        name: formData.get('name'),
        code: formData.get('code')
      });
      setUser(res.data);
      localStorage.setItem('user', JSON.stringify(res.data));
      setView('dashboard');
    } catch (err) {
      alert(err.response?.data || 'Registration failed');
    }
  };

  const handleAdminRegister = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    try {
      const res = await axios.post(`${API_BASE}/register/admin`, {
        email: formData.get('email'),
        name: formData.get('name'),
        code: formData.get('code'),
        admin_password: formData.get('admin_password')
      });
      setUser(res.data);
      localStorage.setItem('user', JSON.stringify(res.data));
      setView('admin-dashboard');
    } catch (err) {
      alert(err.response?.data || 'Admin registration failed');
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    try {
      const res = await axios.post(`${API_BASE}/login`, {
        email: formData.get('email'),
        code: formData.get('code')
      });
      setUser(res.data);
      localStorage.setItem('user', JSON.stringify(res.data));
      setView(res.data.is_admin ? 'admin-dashboard' : 'dashboard');
    } catch (err) {
      alert('Invalid credentials');
    }
  };

  const handleLogout = () => {
    setUser(null);
    localStorage.removeItem('user');
    setView('login');
  };

  const handleCreateCompetition = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    const payload = {
      name: formData.get('name'),
      type: formData.get('type'),
      status: 'draft',
      description: formData.get('description') || '',
    };
    
    const startDate = formData.get('start_date');
    const endDate = formData.get('end_date');
    
    if (payload.type === 'race' && endDate) {
      payload.end_date = new Date(endDate).toISOString();
    } else if (payload.type === 'knockout') {
      if (startDate) {
        payload.start_date = new Date(startDate).toISOString();
      }
      if (endDate) {
        payload.end_date = new Date(endDate).toISOString();
      }
    }
    
    try {
      await axios.post(`${API_BASE}/competitions`, payload);
      e.target.reset();
      loadCompetitions();
      alert('Competition created successfully!');
    } catch (err) {
      alert('Failed to create competition: ' + (err.response?.data || err.message));
    }
  };

  const handleUpdateCompetition = async (compId, updates) => {
    try {
      await axios.put(`${API_BASE}/competitions/${compId}`, updates);
      loadCompetitions();
      if (selectedCompetition?.id === compId) {
        setSelectedCompetition({...selectedCompetition, ...updates});
      }
    } catch (err) {
      const errorMsg = err.response?.data || 'Failed to update competition';
      alert(errorMsg);
      throw err; // Re-throw so caller knows it failed
    }
  };

  const handleUploadEntries = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    const compId = formData.get('competition_id');
    const type = formData.get('type');
    const file = formData.get('file');
    
    if (!compId || !type || !file) {
      alert('Please fill in all fields');
      return;
    }
    
    try {
      const res = await axios.post(`${API_BASE}/entries/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      alert(res.data);
      loadEntries(compId);
      e.target.reset();
    } catch (err) {
      alert('Failed to upload entries: ' + (err.response?.data || err.message));
    }
  };

  const handleUpdateEntry = async (entryId, updates) => {
    try {
      await axios.put(`${API_BASE}/entries/${entryId}`, updates);
      loadEntries(updates.competition_id);
    } catch (err) {
      alert('Failed to update entry: ' + (err.response?.data || err.message));
    }
  };

  const handleDeleteEntry = async (entryId, compId) => {
    if (!window.confirm('Delete this entry?')) return;
    try {
      await axios.delete(`${API_BASE}/entries/${entryId}`);
      loadEntries(compId);
    } catch (err) {
      alert('Failed to delete entry');
    }
  };

  const handleChooseBlindBox = async (boxNumber) => {
    if (!window.confirm(`Select Box #${boxNumber}?`)) return;
    
    try {
      const res = await axios.post(`${API_BASE}/competitions/${selectedCompetition.id}/choose-blind-box`, {
        user_id: user.id,
        box_number: boxNumber
      });
      
      alert(`You got: ${res.data.entry_name}!`);
      loadMyDraws();
      loadBlindBoxes();
      loadAvailableCount();
      setView('my-entries');
    } catch (err) {
      alert(err.response?.data || 'Failed to select box.');
      loadBlindBoxes();
      loadAvailableCount();
    }
  };

  const handleRandomSpin = async () => {
    if (!window.confirm('Let the computer pick a random box for you?')) return;
    
    setShowSpinner(true);
    setSpinnerResult(null);
    
    // Show spinner animation for 3 seconds
    await new Promise(resolve => setTimeout(resolve, 3000));
    
    try {
      const res = await axios.post(`${API_BASE}/competitions/${selectedCompetition.id}/random-pick`, {
        user_id: user.id
      });
      
      setSpinnerResult(`You got: ${res.data.entry_name}!`);
      
      setTimeout(() => {
        setShowSpinner(false);
        setSpinnerResult(null);
        loadMyDraws();
        loadBlindBoxes();
        loadAvailableCount();
        setView('my-entries');
      }, 2000);
    } catch (err) {
      setShowSpinner(false);
      alert(err.response?.data || 'Failed to pick.');
    }
  };

  const getCompetitionDraws = (compId) => {
    return myDraws.filter(d => d.competition_id === compId);
  };

  const hasDrawsForCompetition = (compId) => {
    return getCompetitionDraws(compId).length > 0;
  };

  // Auth views - USES SHARED CSS
  if (!user) {
    return (
      <div className="container">
        <div className="auth-box">
          <h1>🎯 {config.venue_name || 'Sweepstake'}</h1>
          
          {view === 'login' && (
            <>
              <h2>Login</h2>
              <form onSubmit={handleLogin}>
                <input type="email" name="email" placeholder="Email" required />
                <input type="password" name="code" placeholder="Code" required />
                <button type="submit">Login</button>
              </form>
              <p className="switch-view">
                Don't have an account? <a onClick={() => setView('register')}>Register as Player</a> or <a onClick={() => setView('register-admin')}>Register as Admin</a>
              </p>
            </>
          )}

          {view === 'register' && (
            <>
              <h2>Register Player</h2>
              <form onSubmit={handleRegister}>
                <input type="email" name="email" placeholder="Email" required />
                <input type="text" name="name" placeholder="Full Name" required />
                <input type="password" name="code" placeholder="Create Code" required />
                <button type="submit">Register</button>
              </form>
              <p className="switch-view">
                Already have an account? <a onClick={() => setView('login')}>Login</a>
              </p>
            </>
          )}

          {view === 'register-admin' && (
            <>
              <h2>Register Admin</h2>
              <form onSubmit={handleAdminRegister}>
                <input type="email" name="email" placeholder="Email" required />
                <input type="text" name="name" placeholder="Full Name" required />
                <input type="password" name="code" placeholder="Create Code" required />
                <input type="password" name="admin_password" placeholder="Admin Password" required />
                <button type="submit">Register Admin</button>
              </form>
              <p className="switch-view">
                Already have an account? <a onClick={() => setView('login')}>Login</a>
              </p>
            </>
          )}
        </div>
      </div>
    );
  }


  // User Dashboard
  if (!user.is_admin) {
    const activeCompetitions = competitions.filter(c => 
    c.status === 'open' || c.status === 'locked' || c.status === 'completed'
  );
    
    return (
      <div className="app">
        <header className="header">
          <div className="header-content">
            <h1>{config.venue_name || 'Sweepstake'}</h1>
            <div className="user-info">
              <span>Welcome, {user.name}</span>
              <button onClick={handleLogout} className="logout-button">Logout</button>
            </div>
          </div>
        </header>

        <nav className="nav">
          <button onClick={() => setView('dashboard')} className={view === 'dashboard' ? 'active' : ''}>
            Competitions
          </button>
          <button onClick={() => setView('my-entries')} className={view === 'my-entries' ? 'active' : ''}>
            My Entries ({myDraws.length})
          </button>
          <button onClick={() => setView('leaderboard')} className={view === 'leaderboard' ? 'active' : ''}>
            Participants
          </button>
        </nav>

        <main className="main">
          {view === 'dashboard' && (
            <UserDashboard 
              competitions={activeCompetitions}
              myDraws={myDraws}
              onSelectCompetition={(comp) => {
                setSelectedCompetition(comp);
                setView('pick-box');
              }}
              onViewCompetition={(comp) => {
                setLeaderboardComp(comp);
                setView('leaderboard');
              }}
            />
          )}

          {view === 'pick-box' && selectedCompetition && (
            <PickBoxView 
              competition={selectedCompetition}
              blindBoxes={blindBoxes}
              availableCount={availableCount}
              selectionLock={selectionLock}
              onChooseBox={handleChooseBlindBox}
              onRandomSpin={handleRandomSpin}
              onBack={() => setView('dashboard')}
            />
          )}

          {view === 'my-entries' && (
            <MyEntriesView 
              myDraws={myDraws}
              competitions={competitions}
            />
          )}

          {view === 'leaderboard' && (
            <LeaderboardView 
              competitions={competitions}
              leaderboardData={leaderboardData}
              selectedComp={leaderboardComp}
              onSelectCompetition={(comp) => {
                setLeaderboardComp(comp);
                loadLeaderboard(comp.id);
              }}
            />
          )}
        </main>

        {showSpinner && (
          <SpinnerModal entries={entries} result={spinnerResult} />
        )}
      </div>
    );
  }

  // Admin Dashboard
  return (
    <AdminDashboard 
      user={user}
      config={config}
      competitions={competitions}
      entries={entries}
      selectedCompetition={selectedCompetition}
      leaderboardData={leaderboardData}
      leaderboardComp={leaderboardComp}
      onLogout={handleLogout}
      onCreateCompetition={handleCreateCompetition}
      onUpdateCompetition={handleUpdateCompetition}
      onSelectCompetition={(comp) => {
        setSelectedCompetition(comp);
        loadEntries(comp.id);
      }}
      onUploadEntries={handleUploadEntries}
      onUpdateEntry={handleUpdateEntry}
      onUpdatePosition={handleUpdatePosition}
      onDeleteEntry={handleDeleteEntry}
      onLoadEntries={loadEntries}
      onLoadParticipants={(compId) => {
        loadLeaderboard(compId);
        setLeaderboardComp(competitions.find(c => c.id === compId));
      }}
    />
  );
}

// User Components
function UserDashboard({ competitions, myDraws, onSelectCompetition, onViewCompetition }) {
  return (
    <div className="dashboard">
      <h2>Open Competitions</h2>
      
      {competitions.length > 0 ? (
        <div className="competitions-grid">
          {competitions.map(comp => {
            const draws = myDraws.filter(d => d.competition_id === comp.id);
            const hasEntry = draws.length > 0;
            
            return (
              <div key={comp.id} className="competition-card">
                <h3>{comp.name}</h3>
                <div className="competition-details">
                  <span className={`badge badge-${comp.status}`}>{comp.status}</span>
                  <span className="type-badge">{comp.type}</span>
                </div>
                
                {comp.description && <p>{comp.description}</p>}
                
                {comp.type === 'race' && comp.end_date && (
                  <p className="info-text">
                    Race Date: {new Date(comp.end_date).toLocaleDateString()}
                  </p>
                )}
                
                {comp.type === 'knockout' && comp.start_date && comp.end_date && (
                  <div className="date-range">
                    <p className="info-text">
                      {new Date(comp.start_date).toLocaleDateString()} - {new Date(comp.end_date).toLocaleDateString()}
                    </p>
                  </div>
                )}
                
                {hasEntry && (
                  <div className="my-entries">
                    <h4>Your Entry</h4>
                    <div className="entry-list">
                      {draws.map(draw => (
                        <div key={draw.id} className={`entry-item entry-${draw.entry_status}`}>
                          <div className="entry-main">
                            <span className="entry-name">{draw.entry_name}</span>
                            <span className={`badge badge-${draw.entry_status}`}>
                              {draw.entry_status}
                            </span>
                          </div>
                          {draw.seed && <span className="stage">Seed #{draw.seed}</span>}
                          {draw.number && <span className="stage">#{draw.number}</span>}
                          {draw.entry_stage && <span className="stage">Stage: {draw.entry_stage}</span>}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
                
                <div className="card-actions">
                  {comp.status === 'open' && !hasEntry && (
                    <button 
                      onClick={() => onSelectCompetition(comp)}
                      className="cta-button"
                    >
                      📦 Pick Your Box
                    </button>
                  )}
                  
                  {(comp.status === 'locked' || comp.status === 'completed') && (
                    <button 
                      onClick={() => onViewCompetition(comp)}
                      className="action-button"
                    >
                      View Participants
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <p className="info-text">No active competitions at the moment.</p>
      )}
    </div>
  );
}

function PickBoxView({ competition, blindBoxes, availableCount, selectionLock, onChooseBox, onRandomSpin, onBack }) {
  return (
    <div className="blind-draw">
      <button onClick={onBack} className="action-button" style={{marginBottom: '20px'}}>
        ← Back to Competitions
      </button>
      
      <h2>📦 Pick Your Mystery Box - {competition.name}</h2>
      <p className="info-text">{availableCount} boxes remaining</p>
      
      {selectionLock?.locked && !selectionLock?.is_me && (
        <div className="lock-warning">
          <h3>⏳ Please Wait</h3>
          <p><strong>{selectionLock.locked_by}</strong> is currently picking their box...</p>
          <p className="lock-timer">Locked for {selectionLock.locked_for} seconds</p>
          <div className="lock-spinner"></div>
        </div>
      )}
      
      {(!selectionLock?.locked || selectionLock?.is_me) && (
        <>
          <div style={{marginBottom: '30px', textAlign: 'center'}}>
            <button 
              onClick={onRandomSpin}
              className="cta-button"
              style={{fontSize: '18px', padding: '15px 40px'}}
            >
              🎲 Random Spin - Let Computer Pick!
            </button>
          </div>
          
          <div style={{textAlign: 'center', margin: '20px 0', color: '#6b7280'}}>
            <p>— OR —</p>
          </div>
          
          <p className="info-text" style={{textAlign: 'center'}}>Choose a box number manually:</p>
          
          {blindBoxes.length > 0 ? (
            <div className="blind-boxes-grid">
              {blindBoxes.map(box => (
                <div key={box.box_number} className="blind-box">
                  <div className="box-icon">📦</div>
                  <h3>Box #{box.box_number}</h3>
                  <button 
                    onClick={() => onChooseBox(box.box_number)}
                    className="select-button"
                  >
                    Pick This Box
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <div className="no-entries">
              <p>All boxes have been selected or you've already picked one!</p>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function MyEntriesView({ myDraws, competitions }) {
  return (
    <div className="my-draws">
      <h2>My Entries ({myDraws.length})</h2>
      {myDraws.length > 0 ? (
        <div className="draws-grid">
          {myDraws.map(draw => {
            const comp = competitions.find(c => c.id === draw.competition_id);
            return (
              <div key={draw.id} className={`draw-card draw-${draw.entry_status}`}>
                <p className="info-text" style={{marginBottom: '10px', fontWeight: 'bold'}}>
                  {comp?.name || 'Unknown Competition'}
                </p>
                <h3>{draw.entry_name}</h3>
                {draw.seed && <span className="seed-badge">Seed #{draw.seed}</span>}
                {draw.number && <span className="number-badge">#{draw.number}</span>}
                <div className="draw-details">
                  <span className={`badge badge-${draw.entry_status}`}>
                    {draw.entry_status}
                  </span>
                  {draw.entry_stage && <span className="stage-badge">{draw.entry_stage}</span>}
                </div>
                <p className="draw-date">
                  Selected: {new Date(draw.drawn_at).toLocaleDateString()}
                </p>
              </div>
            );
          })}
        </div>
      ) : (
        <p className="info-text">You haven't entered any competitions yet.</p>
      )}
    </div>
  );
}

function LeaderboardView({ competitions, leaderboardData, selectedComp, onSelectCompetition }) {
  const lockedOrCompleted = competitions.filter(c => 
    c.status === 'locked' || c.status === 'completed'
  );
  
  return (
    <div className="admin-results">
      <h2>Participants & Results</h2>
      
      <select 
        onChange={(e) => {
          const comp = competitions.find(c => c.id === parseInt(e.target.value));
          if (comp) onSelectCompetition(comp);
        }}
        value={selectedComp?.id || ''}
        className="competition-select"
      >
        <option value="">Select Competition</option>
        {lockedOrCompleted.map(c => (
          <option key={c.id} value={c.id}>{c.name} ({c.status})</option>
        ))}
      </select>

      {selectedComp && (
        <>
          {selectedComp.status === 'open' && (
            <div className="status-message info" style={{marginTop: '20px'}}>
              <h4>🔒 Competition Not Yet Locked</h4>
              <p>Participants list will be visible once the admin locks the competition.</p>
            </div>
          )}
          
          {(selectedComp.status === 'locked' || selectedComp.status === 'completed') && leaderboardData.length > 0 && (
            <div className="results-grid" style={{marginTop: '30px'}}>
              {leaderboardData.map((draw, idx) => {
                const isEliminated = draw.entry_status === 'eliminated';
                const hasPosition = draw.position && draw.position !== null;
                
                return (
                  <div 
                    key={idx} 
                    className={`result-card result-${draw.entry_status}`}
                    style={isEliminated ? {opacity: 0.6, filter: 'grayscale(30%)'} : {}}
                  >
                    <h3>{draw.user_name}</h3>
                    <p className="entry-name" style={hasPosition ? {fontWeight: 'bold', fontSize: '22px'} : {}}>
                      {draw.entry_name}
                    </p>
                    {hasPosition && (
                      <div style={{fontSize: '24px', fontWeight: 'bold', margin: '10px 0'}}>
                        {draw.position === 999 ? (
                          '🏴 Last Place'
                        ) : (
                          `🏆 ${draw.position}${draw.position === 1 ? 'st' : draw.position === 2 ? 'nd' : draw.position === 3 ? 'rd' : 'th'} Place`
                        )}
                      </div>
                    )}
                    <div className="result-details">
                      <span className={`badge badge-${draw.entry_status}`}>
                        {draw.entry_status}
                      </span>
                      {draw.entry_stage && <span className="stage-badge">{draw.entry_stage}</span>}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}
    </div>
  );
}

// Admin Component
function AdminDashboard({ user, config, competitions, entries, selectedCompetition, leaderboardData, leaderboardComp, onLogout, onCreateCompetition, onUpdateCompetition, onSelectCompetition, onUploadEntries, onUpdateEntry, onUpdatePosition, onDeleteEntry, onLoadEntries, onLoadParticipants }) {
  const [view, setView] = useState('manage');
  
  return (
    <div className="app">
      <header className="header">
        <div className="header-content">
          <h1>{config.venue_name || 'Sweepstake'} - Admin</h1>
          <div className="user-info">
            <span>{user.name}</span>
            <button onClick={onLogout} className="logout-button">Logout</button>
          </div>
        </div>
      </header>

      <nav className="nav">
        <button onClick={() => setView('manage')} className={view === 'manage' ? 'active' : ''}>
          Manage Competitions
        </button>
        <button onClick={() => setView('entries')} className={view === 'entries' ? 'active' : ''}>
          Entries
        </button>
        <button onClick={() => setView('participants')} className={view === 'participants' ? 'active' : ''}>
          View Participants
        </button>
      </nav>

      <main className="main">
        {view === 'manage' && (
          <ManageCompetitions 
            competitions={competitions}
            onCreateCompetition={onCreateCompetition}
            onUpdateCompetition={onUpdateCompetition}
            onSelectCompetition={(comp) => {
              onSelectCompetition(comp);
              setView('entries');
            }}
          />
        )}

        {view === 'entries' && (
          <ManageEntries 
            competitions={competitions}
            entries={entries}
            selectedCompetition={selectedCompetition}
            onSelectCompetition={(comp) => {
              onSelectCompetition(comp);
              onLoadEntries(comp.id);
            }}
            onUploadEntries={onUploadEntries}
            onUpdateEntry={onUpdateEntry}
            onUpdatePosition={onUpdatePosition}
            onDeleteEntry={onDeleteEntry}
          />
        )}

        {view === 'participants' && (
          <AdminParticipantsView 
            competitions={competitions}
            leaderboardData={leaderboardData}
            selectedComp={leaderboardComp}
            onSelectCompetition={onLoadParticipants}
          />
        )}
      </main>
    </div>
  );
}

function ManageCompetitions({ competitions, onCreateCompetition, onUpdateCompetition, onSelectCompetition }) {
  return (
    <div className="admin-dashboard">
      <h2>Manage Competitions</h2>
      
      <div className="admin-section">
        <h3>Create New Competition</h3>
        <form onSubmit={onCreateCompetition} className="admin-form">
          <input type="text" name="name" placeholder="Competition Name" required />
          
          <select name="type" required>
            <option value="">Select Type</option>
            <option value="knockout">Knockout (e.g., World Cup)</option>
            <option value="race">Race (e.g., Grand National)</option>
          </select>
          
          <textarea name="description" placeholder="Description"></textarea>
          
          <button type="submit" className="cta-button">Create Competition</button>
        </form>
      </div>

      <h3>All Competitions</h3>
      <div className="competitions-grid">
        {competitions.map(comp => (
          <div key={comp.id} className="competition-card">
            <h3>{comp.name}</h3>
            <div className="competition-details">
              <span className={`badge badge-${comp.status}`}>{comp.status}</span>
              <span className="type-badge">{comp.type}</span>
            </div>
            <p>{comp.description}</p>
            
            <div className="competition-actions">
              {comp.status === 'draft' && (
                <button 
                  onClick={() => onUpdateCompetition(comp.id, {...comp, status: 'open'})}
                  className="action-button"
                >
                  📢 Open for Users
                </button>
              )}
              
              {comp.status === 'open' && (
                <button 
                  onClick={() => {
                    if (window.confirm('Lock the competition? Users can no longer pick, and everyone will see all selections.')) {
                      onUpdateCompetition(comp.id, {...comp, status: 'locked'});
                    }
                  }}
                  className="action-button"
                >
                  🔒 Lock Competition
                </button>
              )}
              
              
              {comp.status === 'locked' && (
                <button 
                  onClick={async () => {
                    if (window.confirm('Mark as completed? This requires at least one 1st place winner to be set.')) {
                      try {
                        await onUpdateCompetition(comp.id, {...comp, status: 'completed'});
                      } catch (err) {
                        // Error will be shown by the update handler
                      }
                    }
                  }}
                  className="action-button"
                >
                  🏁 Complete Competition
                </button>
              )}
              
              {comp.status === 'completed' && (
                <button 
                  onClick={() => {
                    if (window.confirm('Archive this competition? It will be hidden from users but can be unarchived later.')) {
                      onUpdateCompetition(comp.id, {...comp, status: 'archived'});
                    }
                  }}
                  className="action-button"
                  style={{background: '#6b7280'}}
                >
                  📦 Archive
                </button>
              )}
              
              {comp.status === 'archived' && (
                <button 
                  onClick={() => {
                    if (window.confirm('Unarchive this competition? It will become visible to users again as completed.')) {
                      onUpdateCompetition(comp.id, {...comp, status: 'completed'});
                    }
                  }}
                  className="action-button"
                  style={{background: '#10b981'}}
                >
                  ↩️ Unarchive
                </button>
              )}
              
              <button 
                onClick={() => onSelectCompetition(comp)}
                className="action-button small"
              >
                Manage Entries
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function ManageEntries({ competitions, entries, selectedCompetition, onSelectCompetition, onUploadEntries, onUpdateEntry, onUpdatePosition, onDeleteEntry }) {
  return (
    <div className="admin-entries">
      <h2>Manage Entries</h2>
      
      <select 
        onChange={(e) => {
          const comp = competitions.find(c => c.id === parseInt(e.target.value));
          if (comp) onSelectCompetition(comp);
        }}
        value={selectedCompetition?.id || ''}
        className="competition-select"
      >
        <option value="">Select Competition</option>
        {competitions.map(c => (
          <option key={c.id} value={c.id}>{c.name} ({c.status})</option>
        ))}
      </select>

      {selectedCompetition && (
        <>
          <div className="admin-section">
            <h3>Bulk Upload (CSV)</h3>
            <form onSubmit={onUploadEntries} className="admin-form">
              <input type="hidden" name="competition_id" value={selectedCompetition.id} />
              <input type="hidden" name="type" value={selectedCompetition.type} />
              <input type="file" name="file" accept=".csv" required />
              <button type="submit" className="cta-button">Upload CSV</button>
            </form>
            <p className="info-text">
              CSV Format - Knockout: Name, Seed | Race: Name, Number
            </p>
          </div>

          {entries.length > 0 && (
            <div className="entries-table">
              <h3>Current Entries ({entries.length})</h3>
              <p className="info-text" style={{marginBottom: '15px'}}>
                <strong>Available/Taken</strong> are information only (auto-managed). 
                For {selectedCompetition.type === 'race' ? 'races' : 'knockouts'}: 
                {selectedCompetition.type === 'race' ? ' Set positions (1st-5th, Last) to mark placements.' : ' Mark as eliminated as teams are knocked out.'}
              </p>
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Seed/Number</th>
                    <th>Status</th>
                    <th>Stage</th>
                    <th>Position</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map(entry => (
                    <tr key={entry.id}>
                      <td>{entry.name}</td>
                      <td>{entry.seed || entry.number || '-'}</td>
                      <td>
                        <span className={`badge badge-${entry.status}`} style={{fontSize: '12px'}}>
                          {entry.status}
                        </span>
                      </td>
                      <td>
                        <input 
                          type="text"
                          value={entry.stage || ''}
                          onChange={(e) => onUpdateEntry(entry.id, {
                            ...entry,
                            stage: e.target.value,
                            competition_id: entry.competition_id
                          })}
                          placeholder="Stage"
                          style={{width: '120px'}}
                        />
                      </td>
                      <td>
                        <select
                          value={entry.position || ''}
                          onChange={(e) => {
                            const val = e.target.value;
                            onUpdatePosition(
                              selectedCompetition.id,
                              entry.id,
                              val ? parseInt(val) : null
                            );
                          }}
                          style={{width: '100px'}}
                        >
                          <option value="">None</option>
                          <option value="1">1st</option>
                          <option value="2">2nd</option>
                          <option value="3">3rd</option>
                          <option value="4">4th</option>
                          <option value="5">5th</option>
                          <option value="999">Last</option>
                        </select>
                      </td>
                      <td>
                        {selectedCompetition.type === 'knockout' && entry.status !== 'available' && entry.status !== 'taken' && (
                          <button 
                            onClick={() => onUpdateEntry(entry.id, {
                              ...entry,
                              status: 'eliminated',
                              competition_id: entry.competition_id
                            })}
                            className="action-button small"
                            style={{marginRight: '5px'}}
                          >
                            Eliminate
                          </button>
                        )}
                        <button 
                          onClick={() => onDeleteEntry(entry.id, entry.competition_id)}
                          className="delete-button"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function AdminParticipantsView({ competitions, leaderboardData, selectedComp, onSelectCompetition }) {
  return (
    <div className="admin-results">
      <h2>View All Participants</h2>
      <p className="info-text">Select a competition to view all participants and their selections.</p>
      
      <select 
        onChange={(e) => e.target.value && onSelectCompetition(parseInt(e.target.value))}
        value={selectedComp?.id || ''}
        className="competition-select"
      >
        <option value="">Select Competition</option>
        {competitions.map(c => (
          <option key={c.id} value={c.id}>{c.name} ({c.status})</option>
        ))}
      </select>

      {selectedComp && leaderboardData.length > 0 && (
        <div className="results-grid" style={{marginTop: '30px'}}>
          {leaderboardData.map((draw, idx) => {
            const isEliminated = draw.entry_status === 'eliminated';
            const hasPosition = draw.position && draw.position !== null;
            
            return (
              <div 
                key={idx} 
                className={`result-card result-${draw.entry_status}`}
                style={isEliminated ? {opacity: 0.6, filter: 'grayscale(30%)'} : {}}
              >
                <h3>{draw.user_name}</h3>
                <p className="entry-name" style={hasPosition ? {fontWeight: 'bold', fontSize: '22px'} : {}}>
                  {draw.entry_name}
                </p>
                {hasPosition && (
                  <div style={{fontSize: '24px', fontWeight: 'bold', margin: '10px 0'}}>
                    {draw.position === 999 ? (
                      '🏴 Last Place'
                    ) : (
                      `🏆 ${draw.position}${draw.position === 1 ? 'st' : draw.position === 2 ? 'nd' : draw.position === 3 ? 'rd' : 'th'} Place`
                    )}
                  </div>
                )}
                <div className="result-details">
                  <span className={`badge badge-${draw.entry_status}`}>
                    {draw.entry_status}
                  </span>
                  {draw.entry_stage && <span className="stage-badge">{draw.entry_stage}</span>}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function SpinnerModal({ entries, result }) {
  return (
    <div className="spinner-modal">
      <div className="spinner-content">
        <div className="spinner-wheel">
          <div className="spinner-arrow"></div>
          {entries.slice(0, 16).map((entry, idx) => (
            <div 
              key={entry.id} 
              className="spinner-segment"
              style={{ transform: `rotate(${idx * 22.5}deg)` }}
            >
              <span>{entry.name}</span>
            </div>
          ))}
        </div>
        {result && (
          <div className="spinner-result">
            <h3>✨ {result}</h3>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;