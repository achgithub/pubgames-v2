# PubGames Clean Architecture Design
**Version:** 2.0  
**Date:** January 7, 2026  
**Purpose:** Define the ONE architecture pattern for all apps

---

## Core Principles

1. **Consistency:** Every app uses the same structure
2. **Simplicity:** Minimize moving parts
3. **Modularity:** Clean separation of concerns
4. **Reliability:** Scripts that always work
5. **Speed:** Hot reload everywhere in development
6. **Scalability:** Adding new apps is trivial

---

## System Overview

### Three-Tier Architecture

```
┌─────────────────────────────────────────────────┐
│  Identity Service (Port 3001 + 30000)           │
│  - Gateway and authentication                   │
│  - App launcher                                 │
│  - Token generation                             │
└─────────────────────────────────────────────────┘
                      │
                      │ JWT Token
                      ▼
┌─────────────────────────────────────────────────┐
│  Mini-App N (Ports 300N0 + 300N1)               │
│  - Self-contained application                   │
│  - Validates tokens via Identity Service        │
│  - Own database                                 │
└─────────────────────────────────────────────────┘
```

### Key Concepts

**Identity Service:**
- Central authentication hub
- Generates and validates JWT tokens
- Provides app launcher UI
- Tracks user sessions
- **Same architecture as mini-apps** (critical change!)

**Mini-Apps:**
- Self-contained with own database
- Validate tokens against Identity Service
- SSO via URL parameter
- Independent development and deployment

---

## Port Allocation Standard

### Fixed Scheme

```
┌──────────┬─────────────────────────────────────┐
│   Port   │  Purpose                            │
├──────────┼─────────────────────────────────────┤
│  3001    │  Identity Service Backend (Go API)  │
│  30000   │  Identity Service Frontend (React)  │
├──────────┼─────────────────────────────────────┤
│  30010   │  App 1 Frontend (React dev server)  │
│  30011   │  App 1 Backend (Go API)             │
├──────────┼─────────────────────────────────────┤
│  30020   │  App 2 Frontend (React dev server)  │
│  30021   │  App 2 Backend (Go API)             │
├──────────┼─────────────────────────────────────┤
│  30030   │  App 3 Frontend (React dev server)  │
│  30031   │  App 3 Backend (Go API)             │
├──────────┼─────────────────────────────────────┤
│  ...     │  Supports up to 99 apps (3009X)     │
└──────────┴─────────────────────────────────────┘
```

### Rules

1. **Backend always XX1** (odd port)
2. **Frontend always XX0** (even port, ends in 0)
3. **Identity Service** follows same pattern (3001/30000)
4. **First mini-app** starts at 30010/30011
5. **No exceptions** to this pattern

---

## Standard File Structure

### Every App (Including Identity Service)

```
/app-name/
│
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
│
├── main.go                         # Entry point, routing only
├── handlers.go                     # HTTP handlers
├── models.go                       # Data structures
├── database.go                     # DB init and schema
├── auth.go                         # Auth handlers (uses shared lib)
│
├── package.json                    # NPM config
├── package-lock.json               # NPM lock file
│
├── /src/                          # React source code
│   ├── index.js                   # React entry point
│   ├── App.js                     # Main app component
│   ├── /components/               # (optional) React components
│   └── index.css                  # (optional) App-specific styles
│
├── /public/                       # Static files for React
│   ├── index.html                 # HTML template
│   └── favicon.ico                # (optional)
│
└── /data/                         # Database storage
    └── app.db                     # SQLite database
```

### File Responsibilities

#### main.go
```go
package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    "pubgames/shared/auth"
)

var db *sql.DB

const (
    BACKEND_PORT       = "30X1"  // Replace X with app number
    FRONTEND_PORT      = "30X0"
    DB_PATH            = "./data/app.db"
    IDENTITY_SERVICE   = "http://localhost:3001"
)

func main() {
    log.Println("🚀 Starting App Name...")
    
    // Initialize database
    initDB()
    defer db.Close()
    
    // Setup router
    r := mux.NewRouter()
    api := r.PathPrefix("/api").Subrouter()
    
    // Public routes
    api.HandleFunc("/config", getConfigHandler).Methods("GET")
    
    // Protected routes
    authMw := auth.AuthMiddleware(auth.Config{
        IdentityServiceURL: IDENTITY_SERVICE,
    })
    api.HandleFunc("/data", authMw(getDataHandler)).Methods("GET")
    
    // Admin routes
    adminMw := auth.AdminMiddleware
    api.HandleFunc("/admin", authMw(adminMw(adminHandler))).Methods("POST")
    
    // CORS
    corsHandler := handlers.CORS(
        handlers.AllowedOrigins([]string{
            "http://localhost:" + FRONTEND_PORT,
        }),
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
        handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
    )
    
    log.Printf("✅ Backend running on :%s", BACKEND_PORT)
    log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler(r)))
}
```

#### handlers.go
```go
package main

import (
    "encoding/json"
    "net/http"
)

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
    config := Config{
        AppName: "My App",
    }
    json.NewEncoder(w).Encode(config)
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {
    // Protected route - user is authenticated
    // ... implementation
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    // Admin only route
    // ... implementation
}
```

#### models.go
```go
package main

import "time"

type User struct {
    ID        int       `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    IsAdmin   bool      `json:"is_admin"`
    CreatedAt time.Time `json:"created_at"`
}

type Config struct {
    AppName string `json:"app_name"`
}

// ... other models
```

#### database.go
```go
package main

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", DB_PATH)
    if err != nil {
        log.Fatal(err)
    }
    
    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        name TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    -- Add app-specific tables here
    `
    
    _, err = db.Exec(schema)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ Database initialized")
}
```

#### auth.go
```go
package main

// This file uses the shared auth library
// No need to reimplement authentication
// Identity Service handles token generation
// Shared library handles token validation
```

#### React App.js
```javascript
import React, { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE = 'http://localhost:30X1/api';  // Replace X

function App() {
  const [user, setUser] = useState(null);
  const [view, setView] = useState('loading');

  useEffect(() => {
    // SSO: Check for token in URL
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');
    
    if (token) {
      validateToken(token);
      window.history.replaceState({}, '', window.location.pathname);
      return;
    }
    
    // Check localStorage
    const savedUser = localStorage.getItem('user');
    if (savedUser) {
      setUser(JSON.parse(savedUser));
      setView('dashboard');
    } else {
      setView('login-required');
    }
  }, []);

  const validateToken = async (token) => {
    try {
      const response = await fetch('http://localhost:3001/api/validate-token', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        setView('dashboard');
      }
    } catch (error) {
      setView('login-required');
    }
  };

  const handleLogout = () => {
    setUser(null);
    localStorage.removeItem('user');
    window.location.href = 'http://localhost:3001';
  };

  if (view === 'loading') return <div>Loading...</div>;
  
  if (view === 'login-required') {
    return (
      <div style={{textAlign: 'center', marginTop: '100px'}}>
        <h2>Authentication Required</h2>
        <p>Please log in through the main portal</p>
        <button onClick={() => window.location.href = 'http://localhost:3001'}>
          Go to Login
        </button>
      </div>
    );
  }

  return (
    <div className="app">
      <header>
        <h1>App Name</h1>
        <div>
          <span>Welcome, {user.name}</span>
          <button onClick={handleLogout}>Logout</button>
        </div>
      </header>
      <main>
        {/* Your app content */}
      </main>
    </div>
  );
}

export default App;
```

---

## Shared Components

### Shared Authentication Library

**Location:** `/shared/auth/`

**Structure:**
```
/shared/auth/
├── go.mod
├── middleware.go        # Auth & admin middleware
└── identity.go          # Token validation functions
```

**Purpose:**
- Validate JWT tokens with Identity Service
- Provide middleware for protected routes
- Handle admin authorization
- **Used by ALL apps** (including Identity Service for validation)

**Key Functions:**
```go
// Middleware for protected routes
func AuthMiddleware(config Config) func(http.Handler) http.Handler

// Middleware for admin-only routes
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc

// Validate token with Identity Service
func ValidateToken(config Config, token string) (*User, error)
```

### Shared Styles

**Location:** `/shared/styles/pubgames.css`

**Served By:** Identity Service at `/static/pubgames.css`

**Usage:** All apps import:
```html
<link rel="stylesheet" href="http://localhost:3001/static/pubgames.css">
```

**Contains:**
- Common form styles
- Button styles
- Card layouts
- Badge styles
- Responsive grid
- Color scheme variables

---

## SSO Flow (Standard)

### Step-by-Step

1. **User visits Identity Service**
   ```
   http://localhost:3001
   ```

2. **User logs in**
   - Email + 6-character code
   - Identity Service validates credentials
   - Generates JWT token
   - Stores in localStorage

3. **User clicks app tile**
   - Identity Service redirects to:
   ```
   http://localhost:30X0?token=JWT_TOKEN
   ```

4. **App detects token**
   - React useEffect checks URL params
   - Finds `?token=` parameter

5. **App validates token**
   - Calls Identity Service:
   ```
   GET http://localhost:3001/api/validate-token
   Headers: Authorization: Bearer JWT_TOKEN
   ```

6. **Identity Service responds**
   ```json
   {
     "id": 1,
     "email": "user@example.com",
     "name": "User Name",
     "is_admin": false
   }
   ```

7. **App logs user in**
   - Stores user data in localStorage
   - Removes `?token=` from URL
   - Shows dashboard

### Token Structure

```javascript
{
  "user_id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "is_admin": false,
  "exp": 1234567890  // Expiry timestamp
}
```

### Token Validation Endpoint

```go
// In Identity Service handlers.go
func validateTokenHandler(w http.ResponseWriter, r *http.Request) {
    tokenString := r.Header.Get("Authorization")
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")
    
    // Validate JWT
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(JWT_SECRET), nil
    })
    
    if err != nil || !token.Valid {
        http.Error(w, "Invalid token", 401)
        return
    }
    
    claims := token.Claims.(jwt.MapClaims)
    
    // Get user from database
    var user User
    db.QueryRow("SELECT id, email, name, is_admin FROM users WHERE id = ?", 
        claims["user_id"]).Scan(&user.ID, &user.Email, &user.Name, &user.IsAdmin)
    
    json.NewEncoder(w).Encode(user)
}
```

---

## Database Standards

### Common Pattern

**One SQLite database per app:**
- Location: `/data/app.db`
- Schema defined in `database.go`
- Migrations handled manually or via tool

**Standard Users Table:**
```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    is_admin INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**App-Specific Tables:**
- Defined by each app
- Follow SQLite best practices
- Use foreign keys appropriately

### Identity Service Database

**Special Purpose:**
- Master user directory
- App registry
- Activity tracking

**Tables:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    code TEXT NOT NULL,              -- bcrypt hashed
    is_admin INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,                -- e.g., http://localhost:30010
    description TEXT,
    icon TEXT,                        -- emoji or icon class
    port_frontend INTEGER,            -- 30010
    port_backend INTEGER,             -- 30011
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    app_id INTEGER NOT NULL,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (app_id) REFERENCES apps(id)
);
```

---

## Management Scripts

### start_services.sh

**Purpose:** Start all services reliably

**Logic:**
```bash
#!/bin/bash

# Configuration
APPS=(
  "identity-service:3001:30000"
  "last-man-standing:30011:30010"
  "sweepstakes:30021:30020"
)

# For each app
for app_config in "${APPS[@]}"; do
  IFS=':' read -r app backend frontend <<< "$app_config"
  
  # Check ports free
  check_port $backend
  check_port $frontend
  
  # Start backend
  cd "/path/to/$app"
  gnome-terminal -- bash -c "go run *.go"
  
  # Start frontend
  gnome-terminal -- bash -c "PORT=$frontend npm start"
  
  # Wait for services
  wait_for_port $backend 30
  wait_for_port $frontend 60
done
```

**Features:**
- ✅ Port availability check
- ✅ Graceful error handling
- ✅ Progress indication
- ✅ Timeout handling
- ✅ Summary report

### stop_services.sh

**Purpose:** Stop all services cleanly

**Logic:**
```bash
#!/bin/bash

PORTS=(3001 30000 30011 30010 30021 30020)

for port in "${PORTS[@]}"; do
  pid=$(lsof -ti:$port)
  if [ ! -z "$pid" ]; then
    kill -15 $pid  # SIGTERM
    sleep 2
    kill -9 $pid 2>/dev/null  # SIGKILL if still alive
  fi
done
```

### new_app.sh

**Purpose:** Create new app from template

**Usage:**
```bash
./new_app.sh "My New App" 4
# Creates app at /my-new-app/
# Configured for ports 30040 (frontend) and 30041 (backend)
```

**Logic:**
```bash
#!/bin/bash

APP_NAME=$1
APP_NUMBER=$2

FRONTEND_PORT="300${APP_NUMBER}0"
BACKEND_PORT="300${APP_NUMBER}1"

# Copy template
cp -r template/ "${APP_NAME}/"

# Replace placeholders
find "${APP_NAME}/" -type f -exec sed -i \
  -e "s/APP_NAME/${APP_NAME}/g" \
  -e "s/BACKEND_PORT/${BACKEND_PORT}/g" \
  -e "s/FRONTEND_PORT/${FRONTEND_PORT}/g" {} \;

# Initialize
cd "${APP_NAME}"
go mod init "${APP_NAME}"
npm install

echo "✅ New app created: ${APP_NAME}"
echo "   Backend:  http://localhost:${BACKEND_PORT}"
echo "   Frontend: http://localhost:${FRONTEND_PORT}"
```

---

## Development Workflow

### Starting Development

```bash
# Start all services
cd /path/to/pubgames
./start_services.sh

# Visit Identity Service
open http://localhost:3001

# Login and click app tiles
# Apps will SSO automatically
```

### Working on Specific App

```bash
# Frontend changes (auto hot-reload)
cd /path/to/app/src
nano App.js
# Save - browser refreshes automatically

# Backend changes (requires restart)
cd /path/to/app
nano handlers.go
# Save
# Restart: Ctrl+C in terminal, then go run *.go
```

### Adding New App

```bash
# Create from template
./new_app.sh "Poker Night" 3

# Register in Identity Service
# (Add to apps table in identity.db)

# Restart services
./stop_services.sh
./start_services.sh
```

---

## Build and Deployment

### Development Mode

**Current:** All services run in development mode
- Go: `go run *.go` (interpreted)
- React: `npm start` (webpack dev server)

**Characteristics:**
- ✅ Hot reload
- ✅ Source maps
- ✅ Fast iteration
- ❌ Not optimized
- ❌ Not production-ready

### Production Mode (Future)

**Backend:**
```bash
# Build binary
go build -o app-backend main.go handlers.go models.go database.go auth.go

# Run
./app-backend
```

**Frontend:**
```bash
# Build optimized bundle
npm run build

# Serve via nginx or Go backend
```

---

## Testing Strategy

### Unit Tests

**Backend (Go):**
```bash
go test ./...
```

**Files:**
```
handlers_test.go
models_test.go
database_test.go
```

**Frontend (React):**
```bash
npm test
```

**Files:**
```
App.test.js
```

### Integration Tests

**Test SSO Flow:**
1. Login to Identity Service
2. Get token
3. Call mini-app with token
4. Verify authentication

**Test API Endpoints:**
- Public routes work
- Protected routes require auth
- Admin routes require admin

### Manual Testing Checklist

- [ ] Can register new user
- [ ] Can login with correct credentials
- [ ] Login fails with wrong credentials
- [ ] Can see app tiles
- [ ] Clicking tile redirects with token
- [ ] App auto-logs in via SSO
- [ ] Protected routes work
- [ ] Admin routes work for admins only
- [ ] Logout returns to Identity Service

---

## Error Handling

### Backend Errors

**Standard Response:**
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Details string `json:"details,omitempty"`
}

func sendError(w http.ResponseWriter, message string, code int) {
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: message,
        Code:  code,
    })
}
```

**Usage:**
```go
if err != nil {
    sendError(w, "Database error", 500)
    return
}
```

### Frontend Errors

**Display to User:**
```javascript
try {
  const response = await axios.post(url, data);
} catch (error) {
  if (error.response) {
    alert(error.response.data.error);
  } else {
    alert('Network error');
  }
}
```

### Script Errors

**Fail Fast:**
```bash
set -e  # Exit on any error

check_port() {
  if ! lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
    return 0
  else
    echo "Port $1 in use"
    return 1
  fi
}
```

---

## Security Considerations

### JWT Tokens

**DO:**
- ✅ Set expiry time (24 hours)
- ✅ Use strong secret key
- ✅ Validate on every request
- ✅ Include minimal claims

**DON'T:**
- ❌ Store secrets in code (use env vars)
- ❌ Send tokens in URLs (use headers)
- ❌ Log tokens
- ❌ Store tokens in cookies without secure flag

### Password Handling

**DO:**
- ✅ Use bcrypt with cost 10+
- ✅ Hash before storing
- ✅ Compare hashes, never plaintext

**Code:**
```go
// Storing
hash, _ := bcrypt.GenerateFromPassword([]byte(code), 12)

// Validating
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(code))
```

### CORS Configuration

**Development:**
```go
handlers.AllowedOrigins([]string{
    "http://localhost:30000",
    "http://localhost:30010",
    "http://localhost:30020",
})
```

**Production:**
```go
handlers.AllowedOrigins([]string{
    "https://pubgames.example.com",
})
```

---

## Documentation Standards

### README.md (Per App)

**Required Sections:**
1. App name and description
2. Port numbers (frontend + backend)
3. Quick start commands
4. Database schema overview
5. API endpoints
6. Environment variables

**Example:**
```markdown
# Last Man Standing

Tournament prediction game

## Ports
- Frontend: 30010
- Backend: 30011

## Quick Start
```bash
go run *.go  # Backend
npm start    # Frontend (separate terminal)
```

## Database
- SQLite at /data/lastmanstanding.db
- Tables: users, games, rounds, matches, predictions

## API Endpoints
- GET /api/games - List all games
- POST /api/predictions - Submit prediction
```

### Code Comments

**Required:**
- Public functions
- Complex logic
- Non-obvious decisions

**Example:**
```go
// validateTokenHandler validates JWT tokens from Identity Service
// Returns user data if valid, 401 if invalid
func validateTokenHandler(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

---

## Migration Guide

### From Old Architecture to New

**For Each App:**

1. **Create new structure**
   ```bash
   mkdir app-name-new
   cd app-name-new
   # Copy template
   ```

2. **Move database**
   ```bash
   cp ../app-name-old/data/app.db ./data/
   ```

3. **Split main.go**
   - Extract handlers → `handlers.go`
   - Extract models → `models.go`
   - Extract DB code → `database.go`
   - Keep routing in `main.go`

4. **Update React**
   - Add SSO token detection
   - Update API_BASE URL
   - Import shared CSS

5. **Test**
   - Start backend: `go run *.go`
   - Start frontend: `npm start`
   - Verify SSO works
   - Test all features

6. **Replace old**
   ```bash
   mv ../app-name-old ../app-name-old.backup
   mv ../app-name-new ../app-name
   ```

---

## Success Metrics

**We'll know the redesign works when:**

1. ✅ New app created in < 5 minutes
2. ✅ All apps start with single command
3. ✅ SSO works for all apps
4. ✅ Hot reload works everywhere
5. ✅ No manual build steps needed
6. ✅ Scripts never fail
7. ✅ File structure identical across apps
8. ✅ Documentation stays current
9. ✅ Zero configuration needed
10. ✅ Developer can understand any app in < 30 minutes

---

## Implementation Checklist

### Phase 1: Template Creation
- [ ] Create template directory structure
- [ ] Write template main.go
- [ ] Write template handlers.go
- [ ] Write template models.go
- [ ] Write template database.go
- [ ] Write template auth.go
- [ ] Create template React app
- [ ] Add SSO implementation
- [ ] Test template thoroughly

### Phase 2: Script Development
- [ ] Write new_app.sh
- [ ] Write start_services.sh
- [ ] Write stop_services.sh
- [ ] Test scripts with template
- [ ] Handle all error cases

### Phase 3: Identity Service Migration
- [ ] Refactor to new structure
- [ ] Split monolithic file
- [ ] Add frontend dev server
- [ ] Update port to 3001/30000
- [ ] Test authentication still works
- [ ] Test token generation
- [ ] Test app launcher

### Phase 4: Mini-App Migration
- [ ] Migrate LMS to new structure
- [ ] Migrate Sweepstakes to new structure
- [ ] Test SSO with both apps
- [ ] Verify all features work
- [ ] Update documentation

### Phase 5: Polish
- [ ] Write comprehensive README
- [ ] Create architecture diagram
- [ ] Write developer guide
- [ ] Add troubleshooting section
- [ ] Update all documentation

---

## Appendix: Template Placeholders

**Placeholders to replace when creating new app:**

```
APP_NAME          → Actual app name (e.g., "Poker Night")
APP_NUMBER        → App number (e.g., 3)
FRONTEND_PORT     → Frontend port (e.g., 30030)
BACKEND_PORT      → Backend port (e.g., 30031)
APP_DESCRIPTION   → Short description
APP_ICON          → Emoji or icon (e.g., 🃏)
```

**Files containing placeholders:**
- main.go (ports, constants)
- package.json (name, description)
- go.mod (module name)
- src/App.js (API_BASE, app name)
- README.md (all references)

---

**End of Clean Architecture Design**

This design provides the foundation for rebuilding PubGames with a consistent, reliable, and scalable architecture.
