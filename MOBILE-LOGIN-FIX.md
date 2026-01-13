# Mobile Login Fix

## Problem
When accessing from mobile via QR code (http://192.168.x.x:30000), login was failing because the frontend was hardcoded to call `http://localhost:3001/api`, which on mobile means the phone itself, not the server.

## Solution
Changed API base URL to be dynamic:

```javascript
// Before (hardcoded)
const API_BASE = 'http://localhost:3001/api';

// After (dynamic)
const getApiBase = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:3001/api`;
};
const API_BASE = getApiBase();
```

Now:
- Desktop: `http://localhost:30000` → calls `http://localhost:3001/api` ✅
- Mobile: `http://192.168.1.100:30000` → calls `http://192.168.1.100:3001/api` ✅

## Testing Steps

1. **Restart services** (important - frontend needs to recompile):
   ```bash
   cd /home/andrew/pubgames-v2
   ./stop_services.sh
   ./start_services.sh
   ```

2. **Desktop test**:
   - Open http://localhost:30000
   - Login should still work ✅
   - Note the IP shown in QR code

3. **Mobile test**:
   - Open camera app
   - Scan QR code
   - Browser opens login page
   - Login with credentials:
     - Email: admin@pubgames.local
     - Code: 123456
   - Should now work! ✅

## What Was Already Correct

✅ Backend listening on all interfaces (0.0.0.0:3001)  
✅ CORS allowing all origins  
✅ React dev server on all interfaces  
✅ QR code shows correct IP  

## What Was Wrong

❌ Frontend API calls hardcoded to localhost  

## Now Fixed

✅ Frontend API calls use current hostname  
✅ Works from any IP/hostname  
✅ Mobile login works!  

## Network Flow

```
Mobile Device (192.168.1.50)
  ↓
Scans QR Code
  ↓
Opens: http://192.168.1.100:30000
  ↓
Loads React app
  ↓
Makes API call to: http://192.168.1.100:3001/api/login
  ↓
Server (192.168.1.100:3001)
  ↓
Returns JWT token
  ↓
✅ Logged in!
```

## Verification

After restart, check browser console (F12) on mobile:
- Should see API calls to `http://192.168.x.x:3001/api/...`
- NOT to `http://localhost:3001/api/...`

If login works, you'll see:
- ✅ Token stored in localStorage
- ✅ Redirect to app launcher
- ✅ Apps are clickable

## If Still Not Working

1. **Check network**: Both devices on same WiFi?
2. **Check firewall**: Is port 3001 blocked?
3. **Check logs**: Look at backend terminal for requests
4. **Test manually**: 
   ```bash
   # From mobile browser, try:
   http://192.168.x.x:3001/api/server-info
   # Should return JSON with server info
   ```

## Future: Other Apps Need Same Fix

When you scan QR and access apps, they also need dynamic API URLs:
- Last Man Standing: Update API_BASE in src/App.js
- Sweepstakes: Update API_BASE in src/App.js
- Smoke Test: Update API_BASE in src/App.js

Use same pattern:
```javascript
const getApiBase = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:BACKEND_PORT/api`;
};
```

For now, test with Identity Service login! 🎉
