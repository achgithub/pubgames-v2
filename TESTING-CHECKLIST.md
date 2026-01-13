# PubGames V2 - Testing Checklist

This checklist ensures the template and architecture are working correctly.

## ✅ Phase 1: Template Verification

### Backend Compilation

```bash
cd /home/andrew/pubgames-v2/template
go mod download
go run *.go
```

**Expected Output**:
```
🚀 Starting Template App...
✅ Database initialized at ./data/app.db
✅ Backend running on :30X1
   Frontend should be at :30X0
```

**Checklist**:
- [ ] No compilation errors
- [ ] Database created at `./data/app.db`
- [ ] Server starts and listens on port 30X1
- [ ] All endpoints registered

### Frontend Compilation

```bash
cd /home/andrew/pubgames-v2/template
npm install
npm start
```

**Expected**:
- [ ] Dependencies install without errors
- [ ] React dev server starts on port 30X0
- [ ] Browser opens automatically
- [ ] No console errors

### SSO Token Detection

1. Visit: `http://localhost:30X0?token=fake-token`
2. Check browser console

**Expected**:
- [ ] App attempts to validate token
- [ ] Token validation fails (expected with fake token)
- [ ] Shows "Authentication Required" screen
- [ ] URL cleaned (no ?token= visible)

---

## ✅ Phase 2: Identity Service Verification

### Backend

```bash
cd /home/andrew/pubgames-v2/identity-service
go mod download
go run *.go
```

**Expected**:
```
🚀 Starting PubGames Identity Service...
✅ Database initialized at ./data/identity.db
   Creating default admin user...
   ✅ Default admin created: admin@pubgames.local / 123456
   Creating sample apps...
   ✅ Sample apps created
✅ Identity Service backend running on :3001
   Frontend should be at :30000
```

**Checklist**:
- [ ] No compilation errors
- [ ] Database created with admin user
- [ ] Sample apps seeded
- [ ] Server running on port 3001
- [ ] Static files accessible at /static/pubgames.css

### Frontend

```bash
cd /home/andrew/pubgames-v2/identity-service
npm install
npm start
```

**Expected**:
- [ ] React dev server starts on port 30000
- [ ] Login page displays
- [ ] No console errors

---

## ✅ Phase 3: Full Integration Test

### 1. User Registration

1. Open http://localhost:30000
2. Click "Register"
3. Fill in:
   - Name: Test User
   - Email: test@example.com
   - Code: 123456
4. Submit

**Expected**:
- [ ] Registration succeeds
- [ ] Auto-logged in
- [ ] Redirected to app launcher
- [ ] User stored in localStorage

### 2. User Login

1. Logout
2. Login with:
   - Email: test@example.com
   - Code: 123456

**Expected**:
- [ ] Login succeeds
- [ ] JWT token generated
- [ ] User data in localStorage
- [ ] App launcher visible

### 3. Token Validation

Test the validation endpoint:

```bash
# Get token from localStorage in browser console:
localStorage.getItem('token')

# Test validation:
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     http://localhost:3001/api/validate-token
```

**Expected**:
- [ ] Returns user data
- [ ] Status 200
- [ ] User fields correct

### 4. Admin Check

Login with admin account:
- Email: admin@pubgames.local
- Code: 123456

**Expected**:
- [ ] Shows "ADMIN" badge
- [ ] All apps visible

---

## ✅ Phase 4: SSO Integration Test

### Setup

1. Start Identity Service (both backend and frontend)
2. Start Template App (both backend and frontend)
3. Login to Identity Service as admin

### Add Template App to Directory

Manually add to Identity Service database:

```bash
sqlite3 /home/andrew/pubgames-v2/identity-service/data/identity.db

INSERT INTO apps (name, url, description, icon, is_active)
VALUES ('Template App', 'http://localhost:30010', 'Test template', '📝', 1);

.quit
```

Refresh Identity Service frontend - should see new app tile.

### Test SSO Flow

1. Click "Template App" tile in Identity Service
2. Should redirect to: `http://localhost:30010?token=JWT_TOKEN`

**Expected**:
- [ ] Template app receives token in URL
- [ ] Token validated with Identity Service
- [ ] User auto-logged in
- [ ] URL cleaned (no ?token=)
- [ ] User name displayed
- [ ] Logout button works

### Test Protected Routes

With Template App open and logged in:

1. Try loading items (protected route)
2. Try creating item (protected route)

**Expected**:
- [ ] Protected routes work
- [ ] Data loads/saves correctly
- [ ] No authorization errors

### Test Admin Routes

Login as admin, access Template App via SSO:

**Expected**:
- [ ] Admin badge visible
- [ ] Admin section visible
- [ ] Admin routes accessible

---

## ✅ Phase 5: Scripts Test

### Start Services Script

```bash
cd /home/andrew/pubgames-v2
./start_services.sh
```

**Expected**:
- [ ] Port availability checked
- [ ] Identity Service starts (backend + frontend)
- [ ] New terminal windows open (if available)
- [ ] All services accessible
- [ ] No errors in any terminal

### Stop Services Script

```bash
./stop_services.sh
```

**Expected**:
- [ ] All services stopped
- [ ] All ports freed
- [ ] No processes left running
- [ ] Can restart without port conflicts

### New App Script

```bash
./new_app.sh
```

Fill in:
- Name: test-app
- Display Name: Test Application
- App Number: 5
- Description: Testing app creation
- Icon: 🧪

**Expected**:
- [ ] New directory created at `/home/andrew/pubgames-v2/test-app`
- [ ] All placeholders replaced correctly
- [ ] Ports set to 30050/30051
- [ ] App compiles: `cd test-app && go run *.go`
- [ ] Frontend starts: `npm install && npm start`

---

## ✅ Phase 6: Edge Cases

### Port Conflicts

1. Start Identity Service
2. Try to start it again

**Expected**:
- [ ] Script detects port in use
- [ ] Error message displayed
- [ ] Script exits cleanly

### Missing Dependencies

1. Delete `node_modules` from template
2. Try `npm start`

**Expected**:
- [ ] Script installs dependencies
- [ ] Starts successfully

### Invalid Token

1. In Template App, try:
```javascript
localStorage.setItem('token', 'invalid-token')
```
2. Reload page

**Expected**:
- [ ] Token validation fails
- [ ] Shows login required screen
- [ ] Redirects to Identity Service

### Database Corruption

1. Delete `identity.db`
2. Restart Identity Service

**Expected**:
- [ ] New database created
- [ ] Schema initialized
- [ ] Admin user recreated
- [ ] Sample apps reseeded

---

## ✅ Phase 7: Performance Check

### Response Times

Test API endpoints:

```bash
# Token validation
time curl -H "Authorization: Bearer TOKEN" \
          http://localhost:3001/api/validate-token

# App list
time curl http://localhost:3001/api/apps
```

**Expected**:
- [ ] All responses < 100ms
- [ ] No database locks
- [ ] No memory leaks

### Concurrent Users

Simulate multiple users:
1. Open Identity Service in 3 different browsers
2. Login with different accounts
3. Access Template App from each

**Expected**:
- [ ] All sessions independent
- [ ] No token conflicts
- [ ] No database contention

---

## 🎯 Success Criteria

**The system is ready when:**

- ✅ Template compiles and runs first try
- ✅ Identity Service compiles and runs first try
- ✅ SSO flow works end-to-end
- ✅ Protected routes require authentication
- ✅ Admin routes require admin privilege
- ✅ Scripts work reliably
- ✅ New app creation works
- ✅ Hot reload works everywhere
- ✅ No manual build steps needed
- ✅ Zero configuration required

---

## 🐛 Common Issues

### Issue: "Port already in use"
**Solution**: Run `./stop_services.sh` first

### Issue: "Module not found"
**Solution**: Run `go mod download` in app directory

### Issue: "npm ERR!"
**Solution**: Delete `node_modules` and `package-lock.json`, run `npm install`

### Issue: "Token validation failed"
**Solution**: 
1. Check Identity Service is running on 3001
2. Check CORS configuration
3. Check JWT_SECRET matches

### Issue: "Database locked"
**Solution**: 
1. Stop all services
2. Delete `*.db-shm` and `*.db-wal` files
3. Restart services

---

## 📝 Testing Notes

Record any issues or observations here:

```
Date: ___________
Tester: ___________

Phase 1 - Template:
- 

Phase 2 - Identity Service:
- 

Phase 3 - Integration:
- 

Phase 4 - SSO:
- 

Phase 5 - Scripts:
- 

Phase 6 - Edge Cases:
- 

Phase 7 - Performance:
- 

Overall Notes:
- 
```

---

**Remember**: If ANY checklist item fails, stop and fix it before proceeding. The template must be perfect before migrating any existing apps.
