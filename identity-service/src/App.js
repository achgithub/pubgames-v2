import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';

// Dynamic API base - uses same host as frontend
const getApiBase = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:3001/api`;
};

const API_BASE = getApiBase();

function App() {
  const [view, setView] = useState('login');
  const [user, setUser] = useState(null);
  const [apps, setApps] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [serverInfo, setServerInfo] = useState(null);
  
  // Form states
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  
  // QR code div ref (QRCode library needs a div, not canvas)
  const qrDivRef = useRef(null);
  const qrCodeInstance = useRef(null);

  // Check for existing session on mount
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('logout') === 'true') {
      localStorage.removeItem('user');
      localStorage.removeItem('token');
      setUser(null);
      setView('login');
      window.history.replaceState({}, document.title, window.location.pathname);
      return;
    }

    const savedUser = localStorage.getItem('user');
    const savedToken = localStorage.getItem('token');
    if (savedUser && savedToken) {
      try {
        const userData = JSON.parse(savedUser);
        setUser(userData);
        setView('landing');
      } catch (e) {
        localStorage.removeItem('user');
        localStorage.removeItem('token');
      }
    }
  }, []);

  // Load server info for QR code
  useEffect(() => {
    let isMounted = true;
    
    const loadServerInfo = async () => {
      try {
        const response = await axios.get(`${API_BASE}/server-info`);
        if (isMounted) {
          setServerInfo(response.data);
        }
      } catch (err) {
        console.error('Failed to load server info:', err);
      }
    };
    
    loadServerInfo();
    
    return () => {
      isMounted = false;
    };
  }, []);

  // Generate QR code when server info is available and on login view
  useEffect(() => {
    if (serverInfo && qrDivRef.current && view === 'login' && window.QRCode) {
      // Clear previous QR code if exists
      if (qrCodeInstance.current) {
        qrDivRef.current.innerHTML = '';
        qrCodeInstance.current = null;
      }
      
      // Generate new QR code
      try {
        qrCodeInstance.current = new window.QRCode(qrDivRef.current, {
          text: serverInfo.qr_url,
          width: 200,
          height: 200,
          colorDark: '#000000',
          colorLight: '#ffffff',
          correctLevel: window.QRCode.CorrectLevel.M
        });
      } catch (err) {
        console.error('Failed to generate QR code:', err);
      }
    }
    
    // Cleanup function
    return () => {
      if (qrCodeInstance.current && qrDivRef.current) {
        qrDivRef.current.innerHTML = '';
        qrCodeInstance.current = null;
      }
    };
  }, [serverInfo, view]);

  // Load apps when user is logged in
  useEffect(() => {
    let isMounted = true;
    
    if (user && view === 'landing') {
      const loadApps = async () => {
        try {
          const response = await axios.get(`${API_BASE}/apps`);
          if (isMounted) {
            setApps(response.data || []);
          }
        } catch (err) {
          console.error('Failed to load apps:', err);
        }
      };
      loadApps();
    }
    
    return () => {
      isMounted = false;
    };
  }, [user, view]);

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await axios.post(`${API_BASE}/login`, {
        email,
        code
      });

      const { token, user: userData } = response.data;
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      localStorage.setItem('token', token);
      setCode(''); // Clear code for security
      setView('landing');
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await axios.post(`${API_BASE}/register`, {
        email,
        name,
        code,
        is_admin: false
      });

      // Auto-login after registration
      const response = await axios.post(`${API_BASE}/login`, {
        email,
        code
      });

      const { token, user: userData } = response.data;
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      localStorage.setItem('token', token);
      setCode(''); // Clear code for security
      setView('landing');
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    // Clean up all state
    setApps([]);
    setUser(null);
    setEmail('');
    setName('');
    setCode('');
    setError('');
    
    // Clear storage
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    
    // Switch to login view
    setView('login');
  };

  const launchApp = (app) => {
    const token = localStorage.getItem('token');
    if (token) {
      // Replace localhost with current hostname for mobile support
      const hostname = window.location.hostname;
      const dynamicUrl = app.url.replace('localhost', hostname);
      window.location.href = `${dynamicUrl}?token=${token}`;
    }
  };

  // Login View with QR Code
  if (view === 'login') {
    return (
      <div style={styles.container}>
        <div style={styles.loginContainer}>
          {/* QR Code Section - Desktop Only */}
          <div style={styles.qrSection} className="desktop-only">
            <div style={styles.qrCard}>
              <h3 style={styles.qrTitle}>📱 Quick Login</h3>
              <p style={styles.qrText}>Scan with your phone</p>
              <div style={styles.qrCodeWrapper}>
                <div ref={qrDivRef} style={styles.qrDiv}></div>
              </div>
              {serverInfo && (
                <div style={styles.qrInfo}>
                  <p style={styles.qrUrl}>{serverInfo.local_ip}:{serverInfo.frontend_port}</p>
                  <p style={styles.qrHint}>Same WiFi required</p>
                </div>
              )}
              <div style={styles.futureFeature}>
                <p style={styles.futureText}>🔐 Coming Soon</p>
                <p style={styles.futureHint}>Face ID / Touch ID login</p>
              </div>
            </div>
          </div>

          {/* Login Form */}
          <div style={styles.formCard}>
            <h1 style={styles.title}>🎮 PubGames</h1>
            <h2 style={styles.subtitle}>Login</h2>
            
            {error && <div style={styles.error}>{error}</div>}
            
            <form onSubmit={handleLogin}>
              <div style={styles.formGroup}>
                <label style={styles.label}>Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  style={styles.input}
                  placeholder="your@email.com"
                  autoComplete="email"
                />
              </div>
              
              <div style={styles.formGroup}>
                <label style={styles.label}>6-Character Code</label>
                <input
                  type="password"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  required
                  maxLength={6}
                  style={styles.input}
                  placeholder="******"
                  autoComplete="current-password"
                />
              </div>
              
              <button type="submit" style={styles.button} disabled={loading}>
                {loading ? 'Logging in...' : 'Login'}
              </button>
            </form>
            
            <div style={styles.switchView}>
              Don't have an account?{' '}
              <button onClick={() => { setView('register'); setError(''); }} style={styles.linkButton}>
                Register
              </button>
            </div>
          </div>
        </div>

        {/* CSS for responsive behavior */}
        <style>{`
          @media (max-width: 900px) {
            .desktop-only {
              display: none !important;
            }
          }
        `}</style>
      </div>
    );
  }

  // Register View
  if (view === 'register') {
    return (
      <div style={styles.container}>
        <div style={styles.formCard}>
          <h1 style={styles.title}>🎮 PubGames</h1>
          <h2 style={styles.subtitle}>Register</h2>
          
          {error && <div style={styles.error}>{error}</div>}
          
          <form onSubmit={handleRegister}>
            <div style={styles.formGroup}>
              <label style={styles.label}>Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                style={styles.input}
                placeholder="Your Name"
                autoComplete="name"
              />
            </div>
            
            <div style={styles.formGroup}>
              <label style={styles.label}>Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                style={styles.input}
                placeholder="your@email.com"
                autoComplete="email"
              />
            </div>
            
            <div style={styles.formGroup}>
              <label style={styles.label}>6-Character Code</label>
              <input
                type="password"
                value={code}
                onChange={(e) => setCode(e.target.value.slice(0, 6))}
                required
                maxLength={6}
                style={styles.input}
                placeholder="******"
                autoComplete="new-password"
              />
              <small style={styles.hint}>Choose a 6-character code for login</small>
            </div>
            
            <button type="submit" style={styles.button} disabled={loading}>
              {loading ? 'Creating account...' : 'Register'}
            </button>
          </form>
          
          <div style={styles.switchView}>
            Already have an account?{' '}
            <button onClick={() => { setView('login'); setError(''); }} style={styles.linkButton}>
              Login
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Landing / App Launcher View
  if (view === 'landing' && user) {
    return (
      <div style={styles.container}>
        <div style={styles.landingCard}>
          <header style={styles.header}>
            <div>
              <h1 style={styles.title}>🎮 PubGames</h1>
              <p style={styles.welcome}>Welcome, <strong>{user.name}</strong>!</p>
            </div>
            <button onClick={handleLogout} style={styles.logoutButton}>
              Logout
            </button>
          </header>

          <div style={styles.appsSection}>
            <h2 style={styles.sectionTitle}>Available Apps</h2>
            
            {apps.length === 0 ? (
              <div style={styles.noApps}>
                <p>No apps available yet.</p>
                {user.is_admin && <p>Add apps in the admin panel.</p>}
              </div>
            ) : (
              <div style={styles.appsGrid}>
                {apps.map(app => (
                  <div 
                    key={app.id} 
                    style={styles.appCard}
                    onClick={() => launchApp(app)}
                  >
                    <div style={styles.appIcon}>{app.icon}</div>
                    <h3 style={styles.appName}>{app.name}</h3>
                    <p style={styles.appDescription}>{app.description}</p>
                    <div style={styles.launchButton}>Launch →</div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {user.is_admin && (
            <div style={styles.adminBadge}>
              <span>👑 Admin User</span>
            </div>
          )}
        </div>
      </div>
    );
  }

  // Fallback
  return null;
}

// Inline styles
const styles = {
  container: {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#f0f2f5',
    padding: '20px',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
  },
  loginContainer: {
    display: 'flex',
    gap: '40px',
    alignItems: 'flex-start',
    flexWrap: 'wrap',
    justifyContent: 'center',
    maxWidth: '900px',
    width: '100%'
  },
  qrSection: {
    flex: '0 0 auto',
    order: 1
  },
  qrCard: {
    backgroundColor: 'white',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    padding: '30px',
    width: '280px',
    textAlign: 'center'
  },
  qrTitle: {
    fontSize: '20px',
    margin: '0 0 10px 0',
    color: '#333'
  },
  qrText: {
    fontSize: '14px',
    margin: '0 0 20px 0',
    color: '#666'
  },
  qrCodeWrapper: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: '20px',
    minHeight: '220px'
  },
  qrDiv: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center'
  },
  qrInfo: {
    marginBottom: '20px',
    paddingTop: '20px',
    borderTop: '2px solid #eee'
  },
  qrUrl: {
    fontSize: '14px',
    margin: '0 0 8px 0',
    color: '#007bff',
    fontFamily: 'monospace',
    fontWeight: '600'
  },
  qrHint: {
    fontSize: '12px',
    margin: '0',
    color: '#666'
  },
  futureFeature: {
    padding: '15px',
    backgroundColor: '#f8f9fa',
    borderRadius: '8px',
    border: '2px dashed #ddd'
  },
  futureText: {
    fontSize: '13px',
    margin: '0 0 5px 0',
    color: '#666',
    fontWeight: '600'
  },
  futureHint: {
    fontSize: '11px',
    margin: '0',
    color: '#999'
  },
  formCard: {
    backgroundColor: 'white',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    padding: '40px',
    width: '100%',
    maxWidth: '400px',
    flex: '1 1 400px',
    order: 2
  },
  landingCard: {
    backgroundColor: 'white',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    padding: '40px',
    width: '100%',
    maxWidth: '1000px',
    minHeight: '600px'
  },
  title: {
    fontSize: '32px',
    margin: '0 0 10px 0',
    textAlign: 'center',
    color: '#1a1a1a'
  },
  subtitle: {
    fontSize: '24px',
    margin: '0 0 30px 0',
    textAlign: 'center',
    color: '#666'
  },
  formGroup: {
    marginBottom: '20px'
  },
  label: {
    display: 'block',
    marginBottom: '8px',
    fontWeight: '600',
    color: '#333'
  },
  input: {
    width: '100%',
    padding: '12px',
    fontSize: '16px',
    border: '2px solid #ddd',
    borderRadius: '8px',
    boxSizing: 'border-box',
    transition: 'border-color 0.3s'
  },
  hint: {
    display: 'block',
    marginTop: '5px',
    fontSize: '12px',
    color: '#666'
  },
  button: {
    width: '100%',
    padding: '14px',
    fontSize: '16px',
    fontWeight: '600',
    backgroundColor: '#007bff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'background-color 0.3s'
  },
  switchView: {
    marginTop: '20px',
    textAlign: 'center',
    color: '#666',
    fontSize: '14px'
  },
  linkButton: {
    background: 'none',
    border: 'none',
    color: '#007bff',
    cursor: 'pointer',
    textDecoration: 'underline',
    fontSize: '14px',
    padding: 0
  },
  error: {
    padding: '12px',
    marginBottom: '20px',
    backgroundColor: '#fee',
    border: '1px solid #fcc',
    borderRadius: '8px',
    color: '#c33',
    fontSize: '14px'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingBottom: '30px',
    borderBottom: '2px solid #eee',
    marginBottom: '40px',
    flexWrap: 'wrap',
    gap: '20px'
  },
  welcome: {
    margin: '10px 0 0 0',
    fontSize: '16px',
    color: '#666',
    textAlign: 'center'
  },
  logoutButton: {
    padding: '10px 20px',
    backgroundColor: '#dc3545',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: '600'
  },
  appsSection: {
    marginBottom: '40px'
  },
  sectionTitle: {
    fontSize: '24px',
    marginBottom: '30px',
    color: '#333'
  },
  appsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))',
    gap: '20px'
  },
  appCard: {
    padding: '30px',
    backgroundColor: '#f8f9fa',
    borderRadius: '12px',
    border: '2px solid #ddd',
    cursor: 'pointer',
    transition: 'all 0.3s',
    textAlign: 'center'
  },
  appIcon: {
    fontSize: '48px',
    marginBottom: '15px'
  },
  appName: {
    fontSize: '20px',
    margin: '0 0 10px 0',
    color: '#333'
  },
  appDescription: {
    fontSize: '14px',
    color: '#666',
    margin: '0 0 20px 0'
  },
  launchButton: {
    padding: '10px',
    backgroundColor: '#007bff',
    color: 'white',
    borderRadius: '8px',
    fontWeight: '600',
    fontSize: '14px'
  },
  noApps: {
    textAlign: 'center',
    padding: '60px 20px',
    color: '#666'
  },
  adminBadge: {
    marginTop: '30px',
    padding: '15px',
    backgroundColor: '#fff3cd',
    borderRadius: '8px',
    textAlign: 'center',
    border: '2px solid #ffc107',
    fontWeight: '600'
  }
};

export default App;
