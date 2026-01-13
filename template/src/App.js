import React, { useState, useEffect } from 'react';
import axios from 'axios';

// Dynamic URLs for mobile access - these will be replaced during app creation
// PLACEHOLDER_BACKEND_PORT will be replaced with actual port (e.g., 30011, 30021, etc.)
const getHostname = () => window.location.hostname;
const API_BASE = `http://${getHostname()}:PLACEHOLDER_BACKEND_PORT/api`;
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
  const [view, setView] = useState('loading');
  const [items, setItems] = useState([]);
  const [config, setConfig] = useState({
    app_name: 'PLACEHOLDER_APP_NAME',
    app_icon: 'PLACEHOLDER_ICON'
  });
  const [newItemName, setNewItemName] = useState('');
  const [newItemDescription, setNewItemDescription] = useState('');

  // SSO: Check for token in URL or localStorage on mount
  useEffect(() => {
    let isMounted = true;
    
    const initAuth = async () => {
      // Check for ?token= in URL (SSO from Identity Service)
      const params = new URLSearchParams(window.location.search);
      const token = params.get('token');
      
      if (token) {
        // Validate token with Identity Service
        await validateToken(token, isMounted);
        // Clean up URL
        window.history.replaceState({}, '', window.location.pathname);
        return;
      }
      
      // Check localStorage for existing session
      const savedUser = localStorage.getItem('user');
      const savedToken = localStorage.getItem('jwt_token');
      if (savedUser && savedToken) {
        try {
          const userData = JSON.parse(savedUser);
          if (isMounted) {
            setUser(userData);
            setView('dashboard');
          }
        } catch (e) {
          localStorage.removeItem('user');
          localStorage.removeItem('jwt_token');
          if (isMounted) {
            setView('login-required');
          }
        }
      } else {
        if (isMounted) {
          setView('login-required');
        }
      }
    };

    initAuth();
    
    return () => {
      isMounted = false;
    };
  }, []);

  // Load app config
  useEffect(() => {
    let isMounted = true;
    
    axios.get(`${API_BASE}/config`)
      .then(res => {
        if (isMounted) {
          setConfig(res.data);
        }
      })
      .catch(err => console.log('Using default config'));
    
    return () => {
      isMounted = false;
    };
  }, []);

  // Load items when user is authenticated
  useEffect(() => {
    let isMounted = true;
    
    if (user) {
      loadItems(isMounted);
    }
    
    return () => {
      isMounted = false;
    };
  }, [user]);

  const validateToken = async (token, isMounted = true) => {
    try {
      const response = await fetch(`${IDENTITY_API}/validate-token`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        const userData = await response.json();
        if (isMounted) {
          setUser(userData);
          localStorage.setItem('user', JSON.stringify(userData));
          localStorage.setItem('jwt_token', token);
          setView('dashboard');
        }
      } else {
        if (isMounted) {
          setView('login-required');
        }
      }
    } catch (error) {
      console.error('Token validation failed:', error);
      if (isMounted) {
        setView('login-required');
      }
    }
  };

  const loadItems = async (isMounted = true) => {
    if (!user) return;
    
    try {
      const response = await axios.get(`${API_BASE}/items`);
      if (isMounted) {
        setItems(response.data || []);
      }
    } catch (error) {
      console.error('Failed to load items:', error);
      if (isMounted) {
        setItems([]);
      }
    }
  };

  const handleBackToApps = () => {
    window.location.href = IDENTITY_URL;
  };

  const handleLogout = () => {
    // Clear state first
    setUser(null);
    setItems([]);
    
    // Clear storage
    localStorage.removeItem('user');
    localStorage.removeItem('jwt_token');
    
    // Small delay to let React cleanup finish
    setTimeout(() => {
      window.location.href = `${IDENTITY_URL}?logout=true`;
    }, 100);
  };

  const handleCreateItem = async (e) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    try {
      await axios.post(`${API_BASE}/items`, {
        name: newItemName,
        description: newItemDescription
      });
      
      setNewItemName('');
      setNewItemDescription('');
      loadItems();
    } catch (error) {
      console.error('Failed to create item:', error);
      alert('Failed to create item');
    }
  };

  // Loading state
  if (view === 'loading') {
    return (
      <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
        <h2>Loading...</h2>
      </div>
    );
  }
  
  // Login required state
  if (view === 'login-required') {
    return (
      <>
        <style>
          {`:root {
            --game-color: PLACEHOLDER_COLOR;
            --game-accent: PLACEHOLDER_ACCENT;
          }`}
        </style>
        <div className="container" style={{textAlign: 'center', padding: '100px 20px'}}>
          <h1 style={{marginBottom: '20px'}}>PLACEHOLDER_ICON PLACEHOLDER_APP_NAME</h1>
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
              backgroundColor: 'PLACEHOLDER_COLOR',
              color: 'white'
            }}
          >
            Go to Login
          </a>
        </div>
      </>
    );
  }

  // Main dashboard
  if (user) {
    return (
      <div className="App">
        <style>
          {`:root {
            --game-color: PLACEHOLDER_COLOR;
            --game-accent: PLACEHOLDER_ACCENT;
          }`}
        </style>
        
        <header>
          <div>
            <h1>{config.app_icon} {config.app_name}</h1>
          </div>
          <div className="user-info">
            <span>Welcome, {user.name}! {user.is_admin && <span className="admin-badge">ADMIN</span>}</span>
            <button onClick={handleBackToApps} className="back-to-apps-btn">← Back to Apps</button>
            <button onClick={handleLogout}>Logout</button>
          </div>
        </header>

        <nav className="tabs">
          <button className={view === 'dashboard' ? 'active' : ''} onClick={() => setView('dashboard')}>
            Dashboard
          </button>
          <button className={view === 'items' ? 'active' : ''} onClick={() => setView('items')}>
            Items
          </button>
          {user.is_admin && (
            <button className={view === 'admin' ? 'active' : ''} onClick={() => setView('admin')}>
              Admin Panel
            </button>
          )}
        </nav>

        <main>
          {view === 'dashboard' && (
            <div className="dashboard">
              <h2>Welcome to {config.app_name}!</h2>
              
              <div className="rules">
                <h3>Getting Started</h3>
                <p>This is a template app that demonstrates the PubGames architecture.</p>
                <ul>
                  <li>✅ SSO authentication via Identity Service</li>
                  <li>✅ Protected API routes with JWT</li>
                  <li>✅ Shared CSS styling</li>
                  <li>✅ Admin functionality</li>
                  <li>✅ Modular Go backend</li>
                  <li>✅ React frontend with hot reload</li>
                  <li>✅ Mobile-friendly (dynamic URLs)</li>
                </ul>
              </div>

              {user.is_admin && (
                <div className="admin-dashboard">
                  <h3>Admin Quick Actions</h3>
                  <button onClick={() => setView('items')}>Manage Items</button>
                  <button onClick={() => setView('admin')}>Admin Panel</button>
                </div>
              )}
            </div>
          )}

          {view === 'items' && (
            <div>
              <h2>Sample Items</h2>
              
              <div className="admin-section">
                <h3>Create New Item</h3>
                <form onSubmit={handleCreateItem} className="inline-form">
                  <input
                    type="text"
                    placeholder="Item name"
                    value={newItemName}
                    onChange={(e) => setNewItemName(e.target.value)}
                    required
                  />
                  <input
                    type="text"
                    placeholder="Description (optional)"
                    value={newItemDescription}
                    onChange={(e) => setNewItemDescription(e.target.value)}
                  />
                  <button type="submit">Create Item</button>
                </form>
              </div>

              <div className="admin-section">
                <h3>All Items</h3>
                {items.length === 0 ? (
                  <p className="info-text">No items yet. Create one above!</p>
                ) : (
                  <table>
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Description</th>
                        <th>Created</th>
                      </tr>
                    </thead>
                    <tbody>
                      {items.map(item => (
                        <tr key={item.id}>
                          <td><strong>{item.name}</strong></td>
                          <td>{item.description || '-'}</td>
                          <td>{new Date(item.created_at).toLocaleString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            </div>
          )}

          {view === 'admin' && user.is_admin && (
            <div className="admin">
              <h2>Admin Panel</h2>
              
              <div className="admin-section">
                <h3>Admin Tools</h3>
                <p className="info-text">This section is only visible to administrators.</p>
                <p>You can add admin-only functionality here.</p>
              </div>

              <div className="admin-section">
                <h3>System Information</h3>
                <table className="compact-table">
                  <tbody>
                    <tr>
                      <td><strong>App Name:</strong></td>
                      <td>{config.app_name}</td>
                    </tr>
                    <tr>
                      <td><strong>Total Items:</strong></td>
                      <td>{items.length}</td>
                    </tr>
                    <tr>
                      <td><strong>Admin User:</strong></td>
                      <td>{user.name}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
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
