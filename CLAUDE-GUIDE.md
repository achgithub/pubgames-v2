# PubGames V2 - Claude Assistant Guide

**📋 Purpose**: This document provides essential context for AI assistants (Claude, ChatGPT, etc.) working on this project. Read this file at the start of each session to avoid common mistakes and understand the architecture.

**✅ Workflow Tested**: 2026-01-18 - Mac → GitHub → Pi workflow verified and working

---

## 🖥️ Development Environment

### System Architecture

```
┌─────────────────────────────────────┐
│   Mac Client (Development)          │
│                                      │
│ - Claude Code UI runs here          │
│ - Git repository (can pull/push)    │
│ - NO runtime (no Go/services)       │
│ - Claude edits files here           │
└──────────────┬──────────────────────┘
               │
               │ git push
               ▼
        ┌──────────────┐
        │   GitHub     │
        │ (Central)    │
        └──────┬───────┘
               │
               │ git pull
               ▼
┌──────────────────────────────────────┐
│   Raspberry Pi (Runtime)             │
│                                      │
│ - Runs all services (Go/React)      │
│ - Port 3001, 30000, 30010, etc.     │
│ - Claude has MCP access to logs     │
│ - Network: 192.168.x.x              │
└──────────────────────────────────────┘
```

### Key Points

1. **Mac Client (Development)**:
   - Claude Code UI runs here
   - Full Git access (pull, commit, push)
   - Development environment ONLY
   - **NO runtime** - cannot run Go/React services
   - Claude edits files here
   - Location: `~/Documents/Projects/pubgames-v2`

2. **GitHub**:
   - Central repository
   - All changes flow through here
   - Mac pushes → GitHub → Pi pulls

3. **Raspberry Pi (Runtime)**:
   - Runtime environment for all services
   - Pulls from GitHub to get updates
   - Services run here (Go backends, React frontends)
   - Accessible at 192.168.x.x on local network
   - **Claude has MCP connection to read logs**

4. **Development Workflow (STANDARD - USE THIS)**:

   **Every change follows this path:**
   1. Claude edits files on Mac
   2. Claude commits locally on Mac
   3. **User decides when to push to GitHub** (allows offline work)
   4. User pulls on Pi from GitHub
   5. Services run on Pi
   6. If errors, Claude reads Pi logs via MCP

   **Commands:**
   ```bash
   # On Mac - Claude commits locally (automatic)
   git add .
   git commit -m "Description"

   # On Mac - User pushes when ready (manual, or Claude if explicitly asked)
   git push origin master

   # On Pi (user does this via SSH)
   ssh user@pi
   cd /home/user/pubgames-v2
   git pull origin master
   ```

5. **Claude Can Execute Commands**:
   - Edit files on Mac
   - Commit locally on Mac (automatic after changes)
   - Push to GitHub (ONLY if user explicitly asks)
   - Read logs from Pi via MCP
   - Cannot run services (Mac has no runtime)
   - **Don't just suggest commands - run them!**

---

## 🎨 **CRITICAL: Shared CSS Architecture**

### ⚠️ NEVER FORGET THIS

**All apps share a single CSS file served from Identity Service.**

### How It Works

1. **Shared CSS Location**:
   ```
   /identity-service/static/pubgames.css
   ```
   - Served at: `http://localhost:3001/static/pubgames.css`
   - Also at: `http://192.168.x.x:3001/static/pubgames.css`

2. **Apps Load CSS Dynamically**:
   Every app's `public/index.html` contains:
   ```html
   <script>
     (function() {
       var hostname = window.location.hostname;
       var cssUrl = 'http://' + hostname + ':3001/static/pubgames.css';
       var link = document.createElement('link');
       link.rel = 'stylesheet';
       link.href = cssUrl;
       document.head.appendChild(link);
     })();
   </script>
   ```

3. **Why Dynamic Loading?**:
   - Works on both `localhost` (desktop) AND `192.168.x.x` (mobile)
   - Automatically uses correct hostname
   - Single source of truth for styles

### 🚨 Common Mistakes to Avoid

❌ **WRONG**: Creating app-specific CSS files
❌ **WRONG**: Hardcoding `http://localhost:3001/static/pubgames.css`
❌ **WRONG**: Suggesting inline styles for layout/buttons/forms

✅ **CORRECT**: Update `/identity-service/static/pubgames.css`
✅ **CORRECT**: Use dynamic hostname detection
✅ **CORRECT**: All apps automatically get style updates

### When to Modify Shared CSS

**Modify shared CSS when**:
- Adding new UI components
- Updating button styles
- Changing layouts/forms
- Fixing visual bugs
- Adding responsive styles

**After modifying**, restart Identity Service:
```bash
./stop_services.sh
./start_services.sh
```

---

## 🏗️ Project Architecture

### Port Scheme

```
3001   - Identity Service Backend (serves shared CSS at /static/pubgames.css)
30000  - Identity Service Frontend
30010  - App 1 Frontend
30011  - App 1 Backend
30020  - App 2 Frontend
30021  - App 2 Backend
...
3009X  - App 9X (up to 99 apps)
```

**Rule**:
- Frontend: `300X0` (even, ends in 0)
- Backend: `300X1` (odd, ends in 1)

### Standard App Structure

```
/app-name/
├── main.go              # Entry point, CORS config, routing
├── handlers.go          # HTTP request handlers
├── models.go            # Data structures
├── database.go          # SQLite schema and init
├── auth.go             # Uses shared/auth middleware
├── /src/               # React source code
│   ├── index.js
│   └── App.js          # SSO integration
├── /public/
│   └── index.html      # Loads shared CSS dynamically
├── /data/              # SQLite databases
├── package.json        # NPM dependencies
└── go.mod              # Go dependencies
```

### 🔴 CORS Configuration - CRITICAL FOR MOBILE

**Every app's `main.go` MUST have BOTH localhost AND network IP**:

```go
handlers.AllowedOrigins([]string{
    "http://localhost:" + FRONTEND_PORT,
    "http://192.168.1.45:" + FRONTEND_PORT,  // CRITICAL - adjust IP!
})
```

**Why**:
- Desktop browsers use `localhost`
- Mobile devices use network IP (e.g., `192.168.1.45`)
- Missing network IP = "Network Error" on mobile

**Always verify CORS when**:
- Creating new apps
- Modifying template
- Debugging mobile issues

---

## 🚀 Scripts and Automation

### Available Scripts

| Script | Purpose | Claude Can Run? | Interactive? |
|--------|---------|----------------|--------------|
| `./start_services.sh` | Start all services | ✅ Yes | ❌ No |
| `./stop_services.sh` | Stop all services | ✅ Yes | ❌ No |
| `./status_services.sh` | Check service status | ✅ Yes | ❌ No |
| `./new_app.sh` | Create new app | ✅ Yes | ✅ Both modes |

### Using new_app.sh

**Non-interactive mode (for Claude)**:
```bash
./new_app.sh --name poker-game --display "Poker Night" --number 5 --icon "🃏" --yes
```

**Interactive mode (for humans)**:
```bash
./new_app.sh
```

**Flags**:
- `-n, --name` - App name (required, lowercase-with-hyphens)
- `-d, --display` - Display name (optional)
- `-num, --number` - App number 1-99 (required)
- `-desc, --description` - Description (optional)
- `-i, --icon` - Icon emoji (optional)
- `-y, --yes` - Skip confirmation (important for automation)
- `-h, --help` - Show help

### Example: Claude Creating an App

```bash
# Claude should run this directly
./new_app.sh \
  --name trivia-night \
  --display "Trivia Night" \
  --number 8 \
  --description "Pub quiz game" \
  --icon "🧠" \
  --yes
```

---

## 🔐 SSO (Single Sign-On) Flow

1. User logs into Identity Service (`http://localhost:30000`)
2. Identity Service issues JWT token
3. User clicks app tile
4. Redirects to: `http://localhost:30X0?token=JWT_TOKEN_HERE`
5. App's `src/App.js` detects `?token=` parameter
6. Validates token with: `GET http://localhost:3001/api/validate-token`
7. Auto-logs user in
8. Removes `?token=` from URL (clean history)

### Token Validation

**All protected routes use shared middleware**:
```go
import "pubgames/shared/auth"

authMw := auth.AuthMiddleware(auth.Config{
    IdentityServiceURL: "http://localhost:3001",
})

api.HandleFunc("/protected", authMw(protectedHandler))
```

---

## 📁 Directory Structure

```
/pubgames-v2/
├── identity-service/           # Central auth + app launcher
│   ├── static/
│   │   └── pubgames.css       # ⚠️ SHARED CSS - All apps use this!
│   ├── main.go
│   ├── handlers.go
│   └── ...
├── template/                   # Standard template for new apps
│   ├── public/
│   │   └── index.html         # Loads shared CSS dynamically
│   ├── main.go
│   └── ...
├── shared/                     # Shared Go libraries
│   └── auth/                  # JWT validation middleware
├── [app-name]/                # Individual apps
├── start_services.sh          # ✅ Claude can run
├── stop_services.sh           # ✅ Claude can run
├── status_services.sh         # ✅ Claude can run
├── new_app.sh                 # ✅ Claude can run (use --yes flag)
└── CLAUDE-GUIDE.md            # 📖 This file!
```

---

## ⚠️ Common Mistakes & How to Avoid Them

### 1. Forgetting Shared CSS

**❌ Mistake**: Creating app-specific CSS or suggesting inline styles

**✅ Fix**: Always modify `/identity-service/static/pubgames.css`

### 2. Hardcoding localhost in CORS

**❌ Mistake**:
```go
handlers.AllowedOrigins([]string{
    "http://localhost:30020",
})
```

**✅ Fix**:
```go
handlers.AllowedOrigins([]string{
    "http://localhost:30020",
    "http://192.168.1.45:30020",  // Network IP for mobile
})
```

### 3. Hardcoding localhost in Shared CSS URL

**❌ Mistake**:
```html
<link rel="stylesheet" href="http://localhost:3001/static/pubgames.css">
```

**✅ Fix**:
```html
<script>
  var hostname = window.location.hostname;
  var cssUrl = 'http://' + hostname + ':3001/static/pubgames.css';
  // ... dynamic loading
</script>
```

### 4. Just Suggesting Commands Instead of Running Them

**❌ Mistake**: "You should run `./start_services.sh`"

**✅ Fix**: Actually run `./start_services.sh` using Bash tool

### 5. Creating Files Instead of Editing

**❌ Mistake**: Creating new documentation/config files unnecessarily

**✅ Fix**: Edit existing files when possible

### 6. Inconsistent App Structure

**❌ Mistake**: Deviating from template structure

**✅ Fix**: Always follow template pattern (main.go, handlers.go, models.go, etc.)

---

## 🧪 Testing Checklist

When making changes, verify:

- [ ] **Desktop**: Works on `http://localhost:30X0`
- [ ] **Mobile**: Works on `http://192.168.x.x:30X0`
- [ ] **CORS**: Both localhost AND network IP in AllowedOrigins
- [ ] **Shared CSS**: Loaded from Identity Service
- [ ] **SSO**: Token validation works
- [ ] **Services**: All start with `./start_services.sh`

---

## 🔧 Quick Commands

### Service Management
```bash
# Start everything
./start_services.sh

# Stop everything
./stop_services.sh

# Check status
./status_services.sh
```

### Create New App
```bash
# Non-interactive (Claude)
./new_app.sh --name my-app --number 7 --icon "🎮" --yes

# Interactive (human)
./new_app.sh
```

### Check Ports
```bash
# See what's running
lsof -i :3001
lsof -i :30000

# Kill specific port
kill -9 $(lsof -ti:3001)
```

### Git Workflow (STANDARD)

**On Mac (where Claude works)**:
```bash
# Claude does this automatically after making changes
git add .
git commit -m "Description of changes"

# User pushes when ready (or tells Claude to push)
git push origin master
```

**On Raspberry Pi (you do this via SSH)**:
```bash
# SSH to Pi
ssh user@pi
cd /home/user/pubgames-v2

# Pull latest changes
git pull origin master

# Services auto-update
```

---

## 📚 Related Documentation

- **README.md** - Project overview and quick start
- **CLEAN-ARCHITECTURE-DESIGN.md** - Full architecture specification
- **SERVICE-MANAGEMENT-GUIDE.md** - Service management details
- **CORS-REMINDER.md** - CORS configuration details
- **TESTING-CHECKLIST.md** - Comprehensive testing guide

---

## 🎯 Key Reminders for Claude

1. **You can execute commands directly** - Use Bash tool, don't just suggest
2. **Shared CSS is SACRED** - All styling goes in `/identity-service/static/pubgames.css`
3. **CORS needs BOTH IPs** - localhost AND network IP in every app
4. **Dynamic CSS loading** - Never hardcode localhost in `index.html`
5. **Use `--yes` flag** - When running `./new_app.sh` non-interactively
6. **Follow template structure** - Every app matches the pattern
7. **Test on mobile** - CORS/CSS issues often only appear on mobile
8. **Read this file first** - At the start of each session to avoid repeated mistakes
9. **Commit locally, user pushes** - Commit after changes, but DON'T push unless explicitly asked
10. **Standard workflow** - Mac (edit + commit) → GitHub (when user pushes) → Pi (user pulls)

---

## 🚨 Emergency Procedures

### Services Won't Start
```bash
./stop_services.sh
# Wait 5 seconds
./start_services.sh
```

### Port Conflicts
```bash
# Find what's using the port
lsof -i :3001

# Kill it
kill -9 <PID>
```

### CSS Not Loading
1. Check Identity Service is running: `http://localhost:3001/static/pubgames.css`
2. Verify dynamic loading script in `public/index.html`
3. Check browser console for 404 errors

### Mobile Not Working
1. **Verify CORS** includes network IP in `main.go`
2. **Verify CSS URL** uses dynamic hostname detection
3. **Check network**: Mobile and Pi on same WiFi
4. **Test manually**: Visit `http://192.168.x.x:30000` on mobile

---

**Remember: This project prioritizes consistency. If it works in the template, it should work in all apps. When in doubt, check the template!**
