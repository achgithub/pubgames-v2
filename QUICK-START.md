# 🚀 QUICK START GUIDE

## ⚠️ CRITICAL: Mobile Access Requires Proper CORS!

**Before creating any apps, read `/home/andrew/pubgames-v2/CORS-REMINDER.md`**

New apps MUST include BOTH localhost AND network IP in CORS configuration, or mobile devices will get "Network Error".

---

## What Just Happened?

I've built the complete PubGames V2 template system from scratch based on your clean architecture design. Everything follows the ONE architecture pattern.

## What's Been Created

```
/home/claude/pubgames-v2/
├── identity-service/        ✅ Complete Identity Service
│   ├── Go backend (modular: main, handlers, models, database, auth)
│   ├── React frontend (login, register, app launcher)
│   └── Ports: 3001 (backend) + 30000 (frontend)
│
├── template/                ✅ Working template for new apps
│   ├── Go backend (modular: main, handlers, models, database, auth)
│   ├── React frontend (SSO built-in)
│   └── Ports: 30X0/30X1 (configurable)
│
├── shared/auth/             ✅ Shared authentication library
│   ├── Token validation middleware
│   └── Admin middleware
│
├── start_services.sh        ✅ Start all services
├── stop_services.sh         ✅ Stop all services
├── new_app.sh              ✅ Create new apps from template
├── README.md               ✅ Complete documentation
├── TESTING-CHECKLIST.md    ✅ Testing guide
└── [Design docs]           ✅ All architecture documentation
```

## Immediate Next Steps

### 1. Copy to Raspberry Pi

```bash
# On your Mac (where you're reading this):
# First, this directory is on Claude's computer at /home/claude/pubgames-v2

# You'll need to access this through your MCP setup, or manually transfer:
# - Copy the entire pubgames-v2 directory to your Raspberry Pi
# - Suggested location: /home/andrew/pubgames-v2
```

### 2. Install Dependencies (on Raspberry Pi)

```bash
cd /home/andrew/pubgames-v2

# For Identity Service
cd identity-service
go mod download
npm install
cd ..

# For Template (optional, for testing)
cd template
go mod download
npm install
cd ..

# For Shared Auth
cd shared/auth
go mod download
cd ..
```

### 3. Test the Template

```bash
cd /home/andrew/pubgames-v2/template

# Terminal 1 - Backend
go run *.go

# Terminal 2 - Frontend  
npm start
```

Visit http://localhost:30X0 (replace X with your app number)

### 4. Test Identity Service

```bash
cd /home/andrew/pubgames-v2/identity-service

# Terminal 1 - Backend
go run *.go

# Terminal 2 - Frontend
npm start
```

Visit http://localhost:30000

Login with:
- Email: `admin@pubgames.local`
- Code: `123456`

### 5. Test SSO Flow

1. Start Identity Service (both terminals)
2. Start Template App (both terminals)
3. Login to Identity Service
4. Manually add template to apps database:

```bash
sqlite3 /home/andrew/pubgames-v2/identity-service/data/identity.db

INSERT INTO apps (name, url, description, icon, is_active)
VALUES ('Template Test', 'http://localhost:30010', 'Testing SSO', '📝', 1);

.quit
```

5. Refresh Identity Service, click "Template Test" tile
6. Should SSO into template app automatically!

### 6. Use Scripts

Once everything works manually:

```bash
# Start all services
./start_services.sh

# Stop all services
./stop_services.sh

# Create new app
./new_app.sh
```

## Key Features to Verify

### ✅ Template App
- [x] Compiles without errors
- [ ] Backend starts on correct port
- [ ] Frontend starts on correct port
- [ ] SSO token detection works
- [ ] Protected routes require auth
- [ ] Admin routes require admin
- [ ] Hot reload works (React)

### ✅ Identity Service
- [x] Compiles without errors
- [ ] Creates admin user on first run
- [ ] Seeds sample apps
- [ ] Login/Register work
- [ ] JWT tokens generated
- [ ] Token validation endpoint works
- [ ] App launcher displays apps

### ✅ Shared Auth Library
- [x] Compiles without errors
- [ ] Token validation works
- [ ] Auth middleware works
- [ ] Admin middleware works

## Troubleshooting

### "Port already in use"
```bash
./stop_services.sh
# Or manually:
lsof -ti:3001 | xargs kill -9
```

### "Module not found"
```bash
cd [app-directory]
go mod download
```

### "npm install fails"
```bash
rm -rf node_modules package-lock.json
npm install
```

## What's Different from Old System?

### OLD (Broken):
- ❌ Identity Service: Monolithic, serves built React, no hot reload
- ❌ Mini-apps: Different patterns (LMS 3000-line file, Sweepstakes modular but won't start)
- ❌ Scripts: Unreliable, complex
- ❌ No standard template

### NEW (Clean):
- ✅ Identity Service: Same pattern as mini-apps, dual-process, hot reload
- ✅ All apps: Identical structure, modular Go files
- ✅ Scripts: Simple, reliable
- ✅ Template: Copy and customize

## Migration Strategy

Once template is proven working:

1. **Don't touch old system yet** - keep it as backup
2. **Create new apps** in pubgames-v2 using template
3. **Port business logic** from old apps one at a time
4. **Test thoroughly** before switching
5. **Keep databases** - just copy them over

## Files You Need to Customize

When creating a real app from template:

1. `main.go` - Update ports and app name
2. `handlers.go` - Your business logic
3. `models.go` - Your data structures
4. `database.go` - Your database schema
5. `src/App.js` - Your UI
6. `package.json` - Name and description
7. `go.mod` - Module name

## Default Admin Account

- Email: `admin@pubgames.local`
- Code: `123456`

**CHANGE THIS IN PRODUCTION!**

## Port Reference

```
3001    Identity Backend
30000   Identity Frontend
30010   App 1 Frontend
30011   App 1 Backend  
30020   App 2 Frontend
30021   App 2 Backend
...
```

## Success Checklist

Before declaring "done":

- [ ] Template compiles and runs
- [ ] Identity Service compiles and runs
- [ ] Can login to Identity Service
- [ ] Can register new user
- [ ] SSO flow works (token in URL → auto-login)
- [ ] Protected routes require authentication
- [ ] Admin routes require admin flag
- [ ] ./start_services.sh works
- [ ] ./stop_services.sh works
- [ ] ./new_app.sh creates working app
- [ ] Hot reload works everywhere

## Questions to Ask Yourself

1. Does the template compile without errors? → Test it
2. Does Identity Service start correctly? → Run it
3. Does SSO work end-to-end? → Try the full flow
4. Can I create a new app easily? → Run new_app.sh
5. Are the scripts reliable? → Test them

## If Something Doesn't Work

1. Check TESTING-CHECKLIST.md for detailed tests
2. Review README.md for architecture details
3. Check CLEAN-ARCHITECTURE-DESIGN.md for specifications
4. Look at the code comments

## Next Big Step: Migration

Once everything above works:

1. Read REDESIGN-SUMMARY.md
2. Follow migration guide in CLEAN-ARCHITECTURE-DESIGN.md
3. Migrate one app at a time
4. Test thoroughly before moving to next

## Remember

**The template is the source of truth.**

If it works in the template, it works everywhere. Get the template perfect before migrating anything.

---

## Need Help?

The system is fully documented in these files:
- `README.md` - Main documentation
- `TESTING-CHECKLIST.md` - How to test everything
- `CLEAN-ARCHITECTURE-DESIGN.md` - Full architecture spec
- `QUICK-REFERENCE-CARD.md` - Quick reference
- Each app has its own README.md

---

**You're ready to start testing! Begin with the template, then Identity Service, then SSO integration.**

Good luck! 🎮
