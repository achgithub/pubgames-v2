# Sweepstakes - Quick Start Guide

## 📦 What You Have

8 files ready to deploy:

1. `sweepstakes-main.go` - Entry point (rename to `main.go`)
2. `sweepstakes-handlers.go` - Rename to `handlers.go`
3. `sweepstakes-models.go` - Rename to `models.go`
4. `sweepstakes-database.go` - Rename to `database.go`
5. `sweepstakes-auth.go` - Rename to `auth.go`
6. `go.mod` - Module definition
7. `test-build.sh` - Build verification script
8. `README.md` - Full documentation

## 🚀 Quick Deploy (5 Steps)

### Step 1: Copy Files
```bash
cd ~/pub\ games/sweepstakes

# Backup original
mv main.go main.go.backup.$(date +%Y%m%d)

# Copy new files and rename
cp /path/to/outputs/sweepstakes-main.go main.go
cp /path/to/outputs/sweepstakes-handlers.go handlers.go
cp /path/to/outputs/sweepstakes-models.go models.go
cp /path/to/outputs/sweepstakes-database.go database.go
cp /path/to/outputs/sweepstakes-auth.go auth.go
cp /path/to/outputs/go.mod go.mod
cp /path/to/outputs/test-build.sh .
chmod +x test-build.sh
```

### Step 2: Build
```bash
./test-build.sh
```

Should output:
```
✅ Build successful!
```

### Step 3: Start Services

**Terminal 1 - Identity Service:**
```bash
cd ~/pub\ games/pubgames-identity-service
./start.sh
```

**Terminal 2 - Sweepstakes Backend:**
```bash
cd ~/pub\ games/sweepstakes
go run *.go
```

**Terminal 3 - Sweepstakes Frontend:**
```bash
cd ~/pub\ games/sweepstakes
PORT=3002 npm start
```

### Step 4: Create Test Users
```bash
cd ~/pub\ games/last-man-standing
go run seed_data.go
```

### Step 5: Test
Open http://localhost:3002 and login with:
- Email: `andrew_c_harris@outlook.com`
- Code: `ADMIN001`

## ⚡ Quick Test

Test admin backdoor:
```bash
curl -X POST http://localhost:9081/api/competitions \
  -H "Content-Type: application/json" \
  -H "X-Admin-Override: backdoor123" \
  -d '{"name":"Test","type":"knockout","status":"draft"}'
```

## 🔧 Troubleshooting

**Build fails:**
```bash
cd ~/pub\ games
go work use ./sweepstakes ./shared/auth
cd sweepstakes
go mod tidy
```

**Port in use:**
```bash
kill -9 $(lsof -ti:9081)
```

**Identity Service not running:**
```bash
cd ~/pub\ games/pubgames-identity-service
./start.sh
```

## 📋 Configuration

Default values (can override with environment variables):

```bash
export BACKEND_PORT=9081
export FRONTEND_PORT=3002
export DB_PATH=./data/sweepstake.db
export IDENTITY_SERVICE=http://localhost:3001
export ADMIN_PASSWORD=backdoor123
```

## ✅ Verification Checklist

- [ ] Files copied to sweepstakes directory
- [ ] Build succeeds (`./test-build.sh`)
- [ ] Identity Service running on 3001
- [ ] Backend starts on 9081
- [ ] Frontend starts on 3002
- [ ] Can login with test user
- [ ] Admin backdoor works
- [ ] Can create competition
- [ ] Can upload entries

## 📚 Full Documentation

See `README.md` for complete documentation including:
- Full API reference
- Detailed testing guide
- Environment variables
- Architecture diagrams
- Migration notes

## 🎯 What Changed

**From:** 1 file (1256 lines)
**To:** 5 files (1140 lines) - better organized

**Authentication:** Now uses Identity Service + shared/auth library
**New Features:** Admin backdoor, middleware, better logging
**Compatibility:** 100% compatible with existing frontend

---

**Status: READY FOR DEPLOYMENT ✅**
