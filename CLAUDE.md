# PubGames V2 - Claude Context Document

**Last Updated**: January 20, 2026  
**Latest Commit**: `f8e490c` - "Complete Sweepstakes V2 migration with blind box selection"  
**GitHub**: https://github.com/achgithub/pubgames-v2

---

## 🎯 Current Status

### Platform Overview
PubGames V2 is a multi-application gaming platform running on Raspberry Pi with:
- **Centralized SSO**: Identity Service for authentication
- **V2 Architecture**: Separate Go backends + React frontends
- **Mobile Support**: Configured for IP access (not localhost)
- **Applications**: Identity Service, Last Man Standing, Sweepstakes, Tic-Tac-Toe, Smoke Test

### Recent Major Work (Jan 20, 2026)
**✅ COMPLETED: Sweepstakes V2 Migration**
- Migrated from template structure to full V2 architecture
- Backend: Full Go implementation (database.go, models.go, handlers.go, main.go)
- Frontend: Complete React app (881 lines, src/App.js)
- Features: Blind box selection, admin management, competition lifecycle
- Database: SQLite with competitions, entries, draws tables
- **Committed to GitHub**: Commit `f8e490c`

---

## 🚨 KNOWN ISSUE - URGENT

**Problem**: Sweepstakes app is coded to `localhost` but rest of platform uses IP addresses for mobile access.

**Symptoms**:
- All other apps (Identity Service, Last Man Standing, etc.) work fine
- Sweepstakes login fails when accessing from browser
- Sweepstakes works when tested locally on Pi

**Root Cause**: Sweepstakes still has hardcoded `localhost` URLs instead of IP addresses

**Files to Check**:
- `/home/andrew/pubgames-v2/sweepstakes/main.go` - CORS and backend config
- `/home/andrew/pubgames-v2/sweepstakes/src/App.js` - API URLs (lines 1-10)

**Expected URLs** (based on other working apps):
```javascript
const API_BASE = 'http://192.168.1.148:30031/api';  // NOT localhost:30031
const IDENTITY_URL = 'http://192.168.1.148:30000';  // NOT localhost:30000
const IDENTITY_API = 'http://192.168.1.148:3001';   // NOT localhost:3001
```

**Next Steps**:
1. Update sweepstakes URLs to use 192.168.1.148
2. Update CORS in sweepstakes/main.go
3. Test SSO flow
4. Commit fix to GitHub

---

## 📁 Application Structure

### Identity Service (Ports: 3001/30000)
**Status**: ✅ Working  
**Location**: `/home/andrew/pubgames-v2/identity-service/`  
**URLs**: Configured for 192.168.1.148  
**Role**: Central authentication, JWT tokens, app directory

### Last Man Standing (Ports: 30011/30010)
**Status**: ✅ Working (migrated to V2)  
**Location**: `/home/andrew/pubgames-v2/last-man-standing/`  
**URLs**: Configured for 192.168.1.148  
**Features**: Knockout tournament, WebSocket (reverted to polling)

### Sweepstakes (Ports: 30031/30030)
**Status**: ⚠️ NEEDS FIX - localhost URLs  
**Location**: `/home/andrew/pubgames-v2/sweepstakes/`  
**URLs**: ❌ Currently using localhost (NEEDS UPDATE)  
**Features**: Blind box selection, competitions, draws, admin management

**Backend Files**:
- `main.go` - Entry point, routing, CORS
- `handlers.go` - API endpoints (blind boxes, competitions, draws)
- `models.go` - Competition, Entry, Draw, SelectionLock structs
- `database.go` - SQLite schema (competitions, entries, draws)

**Frontend Files**:
- `src/App.js` - 881 lines, complete React app
  - Main App component with SSO
  - UserDashboard - competition grid
  - PickBoxView - blind box selection with random spin
  - MyEntriesView - user's draws
  - LeaderboardView - participants & results
  - AdminDashboard - admin navigation
  - ManageCompetitions - create, lifecycle management
  - ManageEntries - CSV upload, position management
  - AdminParticipantsView - view all selections
  - SpinnerModal - animated random selection

**Database Schema**:
```sql
competitions (id, name, type, status, description, start_date, end_date)
entries (id, competition_id, name, seed, number, status, position)
draws (id, competition_id, user_email, entry_id, box_number, drawn_at)
selection_locks (competition_id, user_email, locked_at)
```

### Tic-Tac-Toe (Ports: 30021/30020)
**Status**: ✅ Working  
**Location**: `/home/andrew/pubgames-v2/tic-tac-toe/`

### Smoke Test (Ports: 30041/30040)
**Status**: ✅ Working  
**Location**: `/home/andrew/pubgames-v2/smoke-test/`

---

## 🏗️ Architecture Pattern

### Port Scheme
```
3001   - Identity Service Backend (Go)
30000  - Identity Service Frontend (React)
30011  - Last Man Standing Backend
30010  - Last Man Standing Frontend
30031  - Sweepstakes Backend
30030  - Sweepstakes Frontend
30021  - Tic-Tac-Toe Backend
30020  - Tic-Tac-Toe Frontend
30041  - Smoke Test Backend
30040  - Smoke Test Frontend
```
**Rule**: Backend = XX1 (odd), Frontend = XX0 (even)

### Standard App Structure
```
/app-name/
├── main.go           # Entry point, routing, CORS
├── handlers.go       # HTTP handlers
├── models.go         # Data structures
├── database.go       # DB init and schema
├── auth.go          # Uses shared/auth
├── /src/            # React source
│   ├── index.js
│   └── App.js       # SSO integration
├── /public/         # Static files
├── /data/           # SQLite database
├── package.json
└── go.mod
```

### SSO Flow
1. User logs into Identity Service (192.168.1.148:30000)
2. Gets JWT token
3. Clicks app tile → redirects to app with `?token=JWT`
4. App validates token with Identity Service
5. Auto-login, remove `?token` from URL

---

## 🔧 Configuration Details

### Network Configuration
**Raspberry Pi IP**: 192.168.1.148  
**Access**: Local network + mobile devices  
**CORS**: All apps configured for 192.168.1.148 (except Sweepstakes - needs fix)

### URL Pattern (for working apps)
```javascript
// Frontend (React)
const API_BASE = 'http://192.168.1.148:XXXXX/api';
const IDENTITY_URL = 'http://192.168.1.148:30000';
const IDENTITY_API = 'http://192.168.1.148:3001';

// Backend (Go)
handlers.AllowedOrigins([]string{
    "http://192.168.1.148:30000",    // Identity Service
    "http://192.168.1.148:XXXXX",    // This app's frontend
})
```

### Database Files
- Identity Service: `/home/andrew/pubgames-v2/identity-service/data/identity.db`
- Last Man Standing: `/home/andrew/pubgames-v2/last-man-standing/data/lms.db`
- Sweepstakes: `/home/andrew/pubgames-v2/sweepstakes/data/sweepstakes.db`

---

## 🚀 Running the Platform

### Start All Services
```bash
cd /home/andrew/pubgames-v2
./start_services.sh
```

### Start Individual App
```bash
# Terminal 1 - Backend
cd /home/andrew/pubgames-v2/APP-NAME
go run *.go

# Terminal 2 - Frontend
cd /home/andrew/pubgames-v2/APP-NAME
npm start
```

### Check Status
```bash
./status_services.sh
```

### Stop All Services
```bash
./stop_services.sh
```

---

## 📝 Development Workflow

### Git Status
- **Current branch**: master
- **Remote**: github.com:achgithub/pubgames-v2.git
- **Latest commit**: f8e490c
- **Status**: Clean working directory (as of Jan 20)

### Making Changes
```bash
# Make changes
git add .
git commit -m "Description"
git push origin master
```

### Testing Locally
```bash
# Access on Pi
http://localhost:30000

# Access from Mac/Mobile
http://192.168.1.148:30000
```

---

## 🔒 Security & Authentication

### Default Admin Account
- **Email**: admin@pubgames.local
- **Code**: 123456

### JWT Tokens
- **Expiry**: 24 hours
- **Storage**: localStorage
- **Validation**: Every protected route request

### Passwords
- **Hashing**: bcrypt (cost 12)
- **Format**: 6-character codes

---

## 📚 Key Documentation Files

- `README.md` - Main project documentation
- `ARCHITECTURE.md` - Technical architecture
- `QUICK-REFERENCE-CARD.md` - Quick reference
- `CLAUDE.md` - This file (for AI context)

---

## 🐛 Previous Issues (Resolved)

### Execution Environment Confusion (Jan 20)
**Issue**: Commands executing in local container vs Raspberry Pi  
**Solution**: 
- `bash_tool` executes in container (hostname: runsc)
- `raspberry-pi:*` MCP tools execute on actual Pi
- Always verify execution location

### File Transfer Method (Jan 20)
**Issue**: raspberry-pi:write_file truncated large files  
**Solution**: Use SCP for files >50KB
```bash
scp FILE andrew@192.168.1.148:/path/to/destination
```

### Duplicate Component Definitions (Jan 20)
**Issue**: App.js had duplicate function definitions  
**Solution**: Used `sed` to remove duplicate lines
```bash
sed -i '467,503d' App.js
```

---

## 💡 Tips for AI Assistants

### File Access
- User uploads → `/mnt/user-data/uploads`
- Output files → `/mnt/user-data/outputs`
- Raspberry Pi → Use `raspberry-pi:*` tools
- Local container → Use `bash_tool` (limited use)

### Large File Handling
- Files >50KB: Create locally, have user SCP
- Files <50KB: Use `raspberry-pi:write_file`

### Common Commands
```bash
# On Pi (via SSH or raspberry-pi tools)
cd /home/andrew/pubgames-v2
git status
git pull
./start_services.sh

# Check if app is running
ps aux | grep "go run"
lsof -i :30030

# View logs
journalctl -u APP-NAME -f
```

---

## 📋 Immediate Next Steps

1. **Fix Sweepstakes URLs** ⚠️ HIGH PRIORITY
   - Update `src/App.js` API_BASE, IDENTITY_URL, IDENTITY_API
   - Update `main.go` CORS origins
   - Test SSO flow
   - Commit to GitHub

2. **Test Full Platform**
   - Verify all apps accessible from mobile
   - Test SSO across all apps
   - Verify sweepstakes blind box feature

3. **Documentation**
   - Update README with sweepstakes features
   - Document blind box workflow

---

## 🎯 Project Goals & Vision

### Completed
- ✅ V2 architecture standardization
- ✅ Centralized SSO authentication
- ✅ Mobile access support
- ✅ Last Man Standing migration
- ✅ Sweepstakes V2 migration

### In Progress
- ⚠️ Fix sweepstakes localhost issue

### Future
- Template app improvements
- Additional game applications
- Enhanced mobile UI
- Real-time features (WebSocket where appropriate)

---

**End of Context Document**

*This file provides complete context for AI assistants working with this project. Reference the latest commit on GitHub and this document for full understanding of the current state.*
