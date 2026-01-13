# PubGames Current State Analysis
**Date:** January 7, 2026  
**Purpose:** Complete audit before architecture redesign

---

## Executive Summary

**What Works:**
- ✅ Identity Service authentication (after fixes)
- ✅ User registration and login
- ✅ JWT token generation
- ✅ Last Man Standing SSO integration
- ✅ Database schemas are sound
- ✅ Shared authentication library concept

**What's Broken:**
- ❌ Inconsistent architecture across 3 apps
- ❌ Sweepstakes won't start (compile errors)
- ❌ Identity Service requires manual rebuild (no hot reload)
- ❌ Start/stop scripts unreliable
- ❌ Mixed patterns: single-file vs multi-file Go code
- ❌ No standardized testing approach

**Root Cause:** Each app evolved separately with different patterns. No enforced standard.

---

## Current Architecture Inventory

### Identity Service (Port 3001)

**Pattern:** Monolithic single-process serving built React

**Structure:**
```
/pubgames-identity-service/
├── main.go                    (1,000+ lines, everything in one file)
├── /src/                      (React source)
│   ├── App.js
│   ├── Login.js
│   ├── Register.js
│   └── Landing.js
├── /build/                    (Built React - served by Go)
├── /static/
│   └── pubgames.css          (SHARED CSS - used by all apps)
├── /data/
│   └── identity.db
├── package.json
└── go.mod
```

**Go Backend (main.go contains):**
- HTTP server setup
- Database initialization
- All API handlers inline
- User models inline
- Serves static files from /build
- JWT token generation
- CORS configuration

**React Frontend:**
- Login/Register forms
- App launcher (tiles for LMS, Sweepstakes)
- Redirects to apps with ?token=JWT

**Issues:**
- ❌ No hot reload - must run `npm run build` after every React change
- ❌ Everything in one 1000+ line file
- ❌ Different pattern than mini-apps
- ❌ Hard to debug
- ✅ BUT: Authentication works correctly after fixes

**Database Schema:**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    code TEXT NOT NULL,              -- bcrypt hashed 6-char code
    is_admin INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    icon TEXT,
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

### Last Man Standing (Ports 30010 + 30011)

**Pattern:** Dual-process with React dev server

**Structure:**
```
/last-man-standing/
├── main.go                    (3,000+ lines)
├── seed_data.go              (Test data seeding)
├── /src/                     (React source)
│   └── App.js
├── /data/
│   └── lastmanstanding.db
├── package.json
└── go.mod
```

**Go Backend (Port 30011):**
- Self-contained main.go with ALL logic inline
- Database init + schema
- All handlers in main.go
- Models defined inline
- Uses shared auth library for SSO ✅
- CORS configured for port 30010

**React Frontend (Port 30010):**
- Dev server with hot reload ✅
- SSO via ?token= parameter ✅
- Validates token with Identity Service
- Full app UI (games, rounds, predictions, admin)

**Issues:**
- ❌ Massive single file (3000+ lines)
- ❌ API_BASE was wrong port (9080 vs 30011) - FIXED
- ✅ SSO works after fixes
- ✅ Hot reload works
- ✅ Clear separation of concerns (frontend/backend)

**Database Schema:**
```sql
-- Key tables: users, games, rounds, matches, predictions, game_players
-- Tracks tournament games with rounds and player predictions
```

---

### Sweepstakes (Ports 30020 + 30021)

**Pattern:** Dual-process with modular Go files

**Structure:**
```
/sweepstakes/
├── main.go                   (Entry point, routing)
├── handlers.go               (API handlers)
├── models.go                 (Data structures)
├── database.go               (DB initialization)
├── auth.go                   (Auth handlers using shared lib)
├── seed.go                   (Seed data)
├── /src/                     (React source)
│   └── App.js
├── /data/
│   └── sweepstake.db
├── package.json
└── go.mod
```

**Go Backend (Port 30021):**
- **Best practice pattern:** Modular files ✅
- Clean separation: routing, handlers, models, database
- Uses shared auth library ✅
- CORS configured

**React Frontend (Port 30020):**
- Dev server should work
- SSO implementation added ✅
- Admin and user views
- Blind box selection, competitions

**Issues:**
- ❌ **Won't compile/start** - script issues
- ❌ API_BASE was wrong port (9081 vs 30021) - FIXED
- ❌ SSO not tested (can't start)
- ✅ Best code organization of all three apps
- ✅ Clear file separation

**Database Schema:**
```sql
-- Key tables: users, competitions, entries, draws
-- Tracks sweepstake competitions with blind box selections
```

---

## Shared Components

### Shared Authentication Library

**Location:** `/shared/auth/`

**Files:**
```
/shared/auth/
├── go.mod
├── middleware.go        (Auth & Admin middleware)
└── dentity.go          (Token validation)
```

**What It Does:**
- ✅ Token validation against Identity Service
- ✅ Auth middleware for protected routes
- ✅ Admin middleware for admin routes
- ✅ Used by LMS and Sweepstakes

**Issues:**
- ❌ NOT used by Identity Service itself (it's the source)
- ❌ Typo: "dentity.go" should be "identity.go"

### Shared CSS

**Location:** `/pubgames-identity-service/static/pubgames.css`

**Usage:**
- Identity Service serves it
- Mini-apps reference it or copy styles
- Provides consistent look across all apps

**Issues:**
- ❌ Not truly "shared" - each app has own copy or CDN reference
- ❌ Should be centralized

---

## Port Allocation Scheme

**Current (NEW as of Jan 7, 2026):**
```
3001      - Identity Service (Backend + Built Frontend)
30010     - LMS Frontend (React dev server)
30011     - LMS Backend (Go API)
30020     - Sweepstakes Frontend (React dev server)
30021     - Sweepstakes Backend (Go API)
```

**Pattern:**
- Admin tier: 300X
- App N: 300NX where:
  - 300N0 = Frontend dev server
  - 300N1 = Backend API

**Issues:**
- ❌ Identity Service breaks pattern (no separate frontend port)
- ✅ Mini-apps follow pattern consistently

---

## Management Scripts

### start_services.sh

**What it does:**
- Checks port availability
- Builds Identity Service frontend
- Starts all services in terminal windows
- Logs to /logs/ directory

**Issues:**
- ❌ Goes through multiple iterations of `go run` command
- ❌ Latest: `go run *.go` but still fails for Sweepstakes
- ❌ Interactive prompts in Go code block script
- ❌ Doesn't handle build failures gracefully
- ✅ Port conflict detection works

### stop_services.sh

**What it does:**
- Kills processes on configured ports
- Cleans up PID files
- Summary report

**Issues:**
- ✅ Works reliably
- ✅ Clean implementation

---

## SSO Flow (As Designed)

**How It Should Work:**

1. User visits Identity Service (http://localhost:3001)
2. Logs in with email + 6-char code
3. Identity Service generates JWT token
4. User clicks app tile (e.g., "Last Man Standing")
5. Identity Service redirects to: `http://localhost:30010?token=JWT_TOKEN`
6. Mini-app frontend detects ?token= in URL
7. Validates token with Identity Service (http://localhost:3001/api/validate-token)
8. Auto-logs user in
9. Stores user data in localStorage
10. Removes ?token= from URL

**Status:**
- ✅ Identity Service token generation works
- ✅ LMS SSO works (after fixes)
- ❌ Sweepstakes SSO untested (app won't start)

---

## Technology Stack

**Backend:**
- Go 1.21+
- gorilla/mux (routing)
- SQLite3 (database)
- bcrypt (password hashing)
- JWT tokens (authentication)

**Frontend:**
- React 18
- react-router-dom 6
- axios (HTTP client)
- react-scripts 5 (dev server)

**Development:**
- gnome-terminal or xterm for process management
- npm for frontend dependencies
- go mod for backend dependencies

---

## What We Learned

### What Works Well:
1. **Shared auth library concept** - Good idea, reduces duplication
2. **Port scheme** - Logical and scalable
3. **SQLite databases** - Simple, effective, one per app
4. **JWT SSO** - Works when implemented correctly
5. **Dual-process pattern** (separate frontend/backend) - Much better than monolithic

### What Doesn't Work:
1. **Inconsistent file organization** - 3 different patterns
2. **No standard template** - Each app evolved differently
3. **Build complexity** - Identity Service requires manual builds
4. **Script brittleness** - Too many edge cases
5. **No testing** - Zero automated tests
6. **Massive single files** - Hard to maintain

### Critical Issues to Address:
1. **Architecture inconsistency** - Main blocker
2. **Go compilation** - Need one reliable method
3. **Identity Service pattern** - Should match mini-apps
4. **Script reliability** - Must work every time
5. **Developer experience** - Too many manual steps

---

## File Manifest

**Core Services:**
```
/pubgames-identity-service/
  - main.go (1000+ lines)
  - package.json
  - go.mod
  - /src/ (React)
  - /build/ (Built React)
  - /data/identity.db

/last-man-standing/
  - main.go (3000+ lines)
  - seed_data.go
  - package.json
  - go.mod
  - /src/ (React)
  - /data/lastmanstanding.db

/sweepstakes/
  - main.go
  - handlers.go
  - models.go
  - database.go
  - auth.go
  - seed.go
  - package.json
  - go.mod
  - /src/ (React)
  - /data/sweepstake.db
```

**Shared:**
```
/shared/auth/
  - go.mod
  - middleware.go
  - dentity.go
```

**Scripts:**
```
/start_services.sh
/stop_services.sh
/diagnose.sh (if exists)
```

**Documentation:**
```
/PUBGAMES-HANDOVER.md (original)
/DATABASE-FIXED.md
/FIXES-APPLIED-20260107.md
/QUICK-START.md
/PORT-SCHEME-FINAL.md
(and others)
```

---

## Dependencies

**Go Modules (typical):**
```go
require (
    github.com/gorilla/handlers v1.5.2
    github.com/gorilla/mux v1.8.1
    github.com/mattn/go-sqlite3 v1.14.33
    golang.org/x/crypto v0.46.0
    pubgames/shared/auth v0.0.0
)
```

**NPM Packages (typical):**
```json
{
  "react": "^18.x",
  "react-dom": "^18.x",
  "react-router-dom": "^6.x",
  "axios": "^1.x",
  "react-scripts": "5.x"
}
```

---

## Current vs Desired State

### Identity Service

| Aspect | Current | Desired |
|--------|---------|---------|
| Pattern | Monolithic, serves built React | Dual-process like mini-apps |
| Ports | 3001 only | 3001 backend, 30000 frontend |
| Dev Experience | Must rebuild after changes | Hot reload |
| Code Organization | One 1000-line file | Modular files |
| Matches Mini-Apps | ❌ No | ✅ Yes |

### Mini-Apps (LMS, Sweepstakes)

| Aspect | Current | Desired |
|--------|---------|---------|
| Pattern | Dual-process ✅ | Keep this |
| Code Organization | Inconsistent | Standardized |
| File Structure | Different per app | Identical template |
| Startup | Unreliable | Bulletproof |
| SSO | Works (LMS), Broken (Sweeps) | Works everywhere |

---

## Success Criteria for Redesign

**A successful redesign will have:**

1. ✅ **ONE architecture pattern** used by all apps (including Identity)
2. ✅ **ONE file structure** that's identical across apps
3. ✅ **ONE startup method** that works for any app
4. ✅ **Hot reload** everywhere during development
5. ✅ **Modular Go files** (handlers, models, database separate)
6. ✅ **Working SSO** out of the box
7. ✅ **Reliable scripts** that handle errors gracefully
8. ✅ **Template** that can be copied for new apps
9. ✅ **Documentation** that's current and accurate
10. ✅ **Testing** at least for critical paths

**A new app should be:**
```bash
cp -r template/ new-app/
# Edit 5 lines (name, port, description)
./start_services.sh
# IT JUST WORKS
```

---

## Transition Plan (Overview)

1. **Document** ← (We are here)
2. **Design** clean architecture
3. **Create** working template
4. **Test** template thoroughly
5. **Migrate** Identity Service to new pattern
6. **Migrate** LMS to new pattern
7. **Migrate** Sweepstakes to new pattern
8. **Verify** all three work identically
9. **Update** documentation

---

## Notes for Next Steps

**Preserve:**
- Database schemas (they work)
- Authentication logic (works after fixes)
- SSO token validation (correct)
- Port scheme (good)
- React UIs (functional)
- Shared auth library concept

**Throw Away:**
- Current file organization
- Start/stop script complexity
- Identity Service build-serve pattern
- Inconsistent patterns

**Build Fresh:**
- Standard template
- Clean modular structure
- Reliable scripts
- One architecture to rule them all

---

**End of Current State Analysis**
