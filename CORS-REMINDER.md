# 🚨 CRITICAL: CORS Configuration for Mobile Access

## For AI Assistants (Claude, ChatGPT, etc.)

**READ THIS BEFORE CREATING OR MODIFYING ANY APP!**

---

## The Problem

Apps work perfectly on desktop but fail on mobile with "Network Error" when CORS is misconfigured.

## Root Cause

- **Desktop**: Accesses via `http://localhost:30020` 
- **Mobile**: Accesses via `http://192.168.1.45:30020` (network IP)
- **Result**: If CORS only allows `localhost`, mobile requests are **rejected**

---

## ✅ CORRECT CORS Configuration

```go
// CORS configuration - IMPORTANT: Allow both localhost AND network IP for mobile access!
// Users on mobile devices will access via 192.168.1.45 (or similar), not localhost
// Without the network IP, mobile devices will get "Network Error" when making API calls
corsHandler := handlers.CORS(
    handlers.AllowedOrigins([]string{
        "http://localhost:" + FRONTEND_PORT,
        "http://192.168.1.45:" + FRONTEND_PORT, // CRITICAL: Allows mobile access on local network
    }),
    handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
    handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
    handlers.AllowCredentials(),
)
```

---

## ❌ WRONG - Breaks Mobile Access

```go
// This ONLY works on desktop!
corsHandler := handlers.CORS(
    handlers.AllowedOrigins([]string{
        "http://localhost:" + FRONTEND_PORT,  // Mobile will fail!
    }),
    // ...
)
```

---

## Checklist for New Apps

When creating a new app or modifying `main.go`:

- [ ] Verify CORS includes `http://localhost:` + FRONTEND_PORT
- [ ] Verify CORS includes `http://192.168.1.45:` + FRONTEND_PORT
- [ ] Test on mobile device after creation
- [ ] If mobile fails with "Network Error", check CORS first!

---

## Files Already Fixed

✅ `/template/main.go` - Template has correct CORS  
✅ `/identity-service/main.go` - Allows all origins  
✅ `/smoke-test/main.go` - Fixed  
✅ `/last-man-standing/main.go` - Fixed  
✅ `/sweepstakes/main.go` - Fixed  

---

## Testing Mobile Access

1. Start services: `./start_services.sh`
2. On mobile, navigate to `http://192.168.1.45:30020`
3. Login and test functionality
4. If "Network Error" appears when clicking buttons → **CHECK CORS!**

---

**Remember**: This is the #1 cause of "works on desktop, fails on mobile" issues in PubGames V2!
