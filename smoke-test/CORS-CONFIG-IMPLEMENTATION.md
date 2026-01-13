# Smoke Test - Shared CORS Config Implementation

**Date:** January 13, 2026  
**Status:** ✅ Updated to use shared CORS config

---

## What Changed

### Files Modified

1. **`smoke-test/main.go`**
   - Added import: `pubgames/shared/config`
   - Replaced hardcoded CORS origins with dynamic config loading
   - Added logging for CORS mode and allowed origins
   - Added CORS blocking notifications

2. **`smoke-test/go.mod`**
   - Added dependency: `pubgames/shared/config v0.0.0`
   - Added replace directive: `replace pubgames/shared/config => ../shared/config`

### Before (Hardcoded)
```go
handlers.AllowedOrigins([]string{
    "http://localhost:" + FRONTEND_PORT,
    "http://192.168.1.45:" + FRONTEND_PORT,  // ❌ Hardcoded IP
}),
```

### After (Dynamic)
```go
// Load CORS configuration from shared config
corsConfig, err := config.LoadCORSConfig()
log.Printf("📋 CORS Mode: %s", corsConfig.CORS.Mode)
log.Printf("📋 Allowed Origins: %v", corsConfig.GetAllowedOrigins())

// Use pattern matching from config file
handlers.AllowedOriginValidator(func(origin string) bool {
    allowed := corsConfig.IsOriginAllowed(origin)
    if !allowed {
        log.Printf("❌ CORS blocked: %s", origin)
    }
    return allowed
}),
```

---

## Benefits

✅ **No IP Hardcoding** - Works on any network IP  
✅ **Pattern Matching** - `http://192.168.1.*:*` matches entire subnet  
✅ **Centralized Config** - One file for all apps  
✅ **Hot Swappable** - Edit config for prod/dev modes  
✅ **Debug Logging** - See exactly what's allowed/blocked  

---

## CORS Config File

**Location:** `~/pubgames-v2/shared/config/cors-config.json`

**Current Settings:**
```json
{
  "environment": "development",
  "pub_id": "dev-local",
  "pub_name": "Development Environment",
  "cors": {
    "mode": "pattern",
    "patterns": [
      "http://localhost:*",      // Any localhost port
      "http://192.168.1.*:*"     // Any IP in 192.168.1.x subnet
    ],
    "explicit_origins": []
  },
  "updated_at": "2026-01-13T12:00:00Z",
  "updated_by": "system"
}
```

### To Change Network Subnet

If your Pi moves to a different network (e.g., 192.168.0.x):

```bash
nano ~/pubgames-v2/shared/config/cors-config.json
```

Change:
```json
"patterns": [
  "http://localhost:*",
  "http://192.168.0.*:*"    // Updated subnet
]
```

Save and restart services - no code changes needed!

---

## Testing

### 1. Build Test
```bash
cd ~/pubgames-v2/smoke-test
./test-cors-config.sh
```

This will:
- Run `go mod tidy`
- Test compilation
- Show current CORS config
- Provide startup instructions

### 2. Run Backend
```bash
cd ~/pubgames-v2/smoke-test
./start-backend.sh
```

Or manually:
```bash
cd ~/pubgames-v2/smoke-test
go run .
```

### 3. Watch for Log Output

**Expected startup logs:**
```
🚀 Starting Smoke test...
✅ Loaded CORS config: mode=pattern, environment=development
📋 CORS Mode: pattern
📋 Allowed Origins: [http://localhost:* http://192.168.1.*:*]
✅ Backend running on :30011
   Frontend should be at :30010
```

### 4. Test CORS

**From another terminal:**
```bash
# Test from localhost - should work
curl -H "Origin: http://localhost:30010" -I http://localhost:30011/api/config

# Test from network IP - should work
curl -H "Origin: http://192.168.1.45:30010" -I http://localhost:30011/api/config

# Test from blocked origin - should fail
curl -H "Origin: http://evil.com" -I http://localhost:30011/api/config
```

**Backend logs should show:**
```
✅ Request from: http://localhost:30010
✅ Request from: http://192.168.1.45:30010
❌ CORS blocked: http://evil.com
```

### 5. Start Frontend
```bash
cd ~/pubgames-v2/smoke-test
npm start
```

Frontend will run on port 30010 and connect to backend on 30011.

---

## Production Mode

When deploying to a pub, edit the config:

```json
{
  "environment": "production",
  "pub_id": "the-crown-london-01",
  "pub_name": "The Crown",
  "cors": {
    "mode": "explicit",  // Changed from "pattern"
    "patterns": [],
    "explicit_origins": [
      "http://localhost:3001",
      "http://localhost:30010",
      "http://192.168.1.100:3001",
      "http://192.168.1.100:30010"
    ]
  }
}
```

Restart services - same code works with exact origin matching.

---

## Migration Pattern for Other Apps

**This is now the template for:**
1. `last-man-standing/main.go`
2. `sweepstakes/main.go`
3. `identity-service/main.go` (after refactoring)
4. Any future mini-apps

**Steps to migrate each app:**
1. Update `import` to include `pubgames/shared/config`
2. Add to `go.mod`: `pubgames/shared/config v0.0.0`
3. Add replace: `replace pubgames/shared/config => ../shared/config`
4. Replace hardcoded CORS with `config.LoadCORSConfig()`
5. Use `handlers.AllowedOriginValidator()` with pattern matching
6. Test with `go mod tidy && go build`

---

## Troubleshooting

### Error: "module not found"
```bash
cd ~/pubgames-v2/smoke-test
go mod tidy
```

### Error: "config file not found"
Check file exists:
```bash
ls -l ~/pubgames-v2/shared/config/cors-config.json
```

If missing, it will use safe defaults (localhost only).

### CORS still blocking network access
Check your actual IP:
```bash
hostname -I
```

Update the pattern in `cors-config.json` to match your network.

### Want to see what's being blocked?
Backend logs show every CORS decision:
```
✅ Request allowed from: http://192.168.1.45:30010
❌ CORS blocked: http://192.168.1.99:8080
```

---

## Next Steps

1. ✅ **Test smoke-test** - Run the test script
2. **Migrate LMS** - Apply same pattern to last-man-standing
3. **Migrate Sweepstakes** - Apply same pattern to sweepstakes
4. **Refactor Identity Service** - Split to dual-port + use shared config
5. **Update start_services.sh** - Ensure all apps use shared config

---

## Questions for Future Discussion

1. **Admin UI for CORS config?** - Should Identity Service have a UI to edit cors-config.json?
2. **Multiple network support?** - Support patterns for multiple networks simultaneously?
3. **Pub-specific configs?** - Should each pub have unique pub_id in config?
4. **Config validation?** - Add startup validation for CORS patterns?

---

**This smoke-test is now the working reference implementation for the entire PubGames platform.**
