# QR Code Login & Passkey Authentication

## Overview

This document describes the enhanced authentication system for PubGames, including QR code login (Priority 1 - IMPLEMENTED) and passkey authentication (Priority 2 - FUTURE).

## Priority 1: QR Code Login ✅ IMPLEMENTED

### What It Does

Users can scan a QR code on the desktop login page to instantly access the site on their mobile device (same WiFi required).

### Architecture

```
Desktop Browser                     Mobile Device
┌─────────────────────┐            ┌──────────────────┐
│  Login Page         │            │  Camera App      │
│  ┌───────────┐      │            │                  │
│  │ QR Code   │ ◄────┼────────────┤  [Scan QR]       │
│  │  192.168  │      │            │                  │
│  │  .1.100   │      │            │  Opens:          │
│  │  :30000   │      │            │  http://192...   │
│  └───────────┘      │            │                  │
│                     │            │  ↓               │
│  [Email/Code Form]  │            │  [Login Form]    │
└─────────────────────┘            └──────────────────┘
```

### Technical Implementation

#### Backend Changes

1. **New Endpoint: `/api/server-info`**
   - Returns server's local IP address
   - Provides frontend/backend ports
   - Generates QR URL

2. **CORS Updates**
   - Now allows requests from any local network IP
   - Enables mobile devices on same WiFi to access

3. **Database Schema**
   - Added passkey columns (for future use):
     - `passkey_id` - WebAuthn credential ID
     - `passkey_public_key` - Public key for verification
     - `passkey_counter` - Signature counter (anti-replay)
     - `passkey_transports` - How credential can be used
     - `passkey_created_at` - When passkey was registered

#### Frontend Changes

1. **QR Code Display**
   - Fetches server info on load
   - Generates QR code with IP address
   - Hidden on mobile (responsive)
   - Shows "Passkeys coming soon" hint

2. **Layout**
   - Side-by-side: QR card + Login form
   - Desktop only (hidden on mobile)
   - Professional, clean design

### Files Modified

```
identity-service/
├── main.go                    # Added /api/server-info endpoint
├── server_info.go             # NEW - IP detection & server info
├── database.go                # Added passkey columns to schema
├── src/App.js                 # Added QR code display
└── public/index.html          # Added QR library CDN
```

### How to Use

1. **Desktop User:**
   - Navigate to http://localhost:30000
   - See QR code on left side
   - Login normally with email/code

2. **Mobile User:**
   - Open camera app
   - Scan QR code on desktop screen
   - Automatically opens http://192.168.x.x:30000
   - Login with same credentials
   - Stay logged in on mobile!

### Network Requirements

- Desktop and mobile must be on **same WiFi network**
- No internet required (fully local)
- Works with private home/office networks

---

## Priority 2: Passkey Authentication 🔐 FUTURE

### Vision

Ultimate goal: User scans QR code → Face ID/Touch ID activates → Instant login (no email/code needed)

### How It Will Work

```
User Experience Flow:
┌─────────────────────────────────────────────────────┐
│                                                      │
│  Desktop: Shows QR code                             │
│  ↓                                                   │
│  Mobile: Scan QR                                    │
│  ↓                                                   │
│  Mobile: Detects passkey exists                     │
│  ↓                                                   │
│  Mobile: "Use Face ID to login?"                    │
│  ↓                                                   │
│  User: Face ID scan                                 │
│  ↓                                                   │
│  Mobile: Instantly logged in! 🎉                    │
│                                                      │
│  Fallback: If no passkey, show email/code form     │
└─────────────────────────────────────────────────────┘
```

### Technical Architecture

#### WebAuthn Flow

Passkeys use WebAuthn standard (supported by iOS 16+, Android 9+):

**Registration (Setup):**
```
User Profile → "Add Passkey" button
↓
Backend generates challenge (random bytes)
↓
Frontend calls navigator.credentials.create()
↓
iOS shows Face ID prompt
↓
Device creates public/private key pair
↓
Public key sent to server
↓
Server stores: user_id, credential_id, public_key
```

**Authentication (Login):**
```
User scans QR code
↓
Frontend detects passkey available
↓
Backend generates auth challenge
↓
Frontend calls navigator.credentials.get()
↓
iOS shows Face ID prompt
↓
Device signs challenge with private key
↓
Server verifies signature with stored public key
↓
Server issues JWT token
↓
User logged in!
```

### Database Schema (Already Implemented)

```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  is_admin INTEGER DEFAULT 0,
  
  -- Passkey fields
  passkey_id TEXT,                    -- WebAuthn credential ID
  passkey_public_key TEXT,            -- For signature verification
  passkey_counter INTEGER DEFAULT 0,  -- Anti-replay protection
  passkey_transports TEXT,            -- USB, NFC, BLE, internal
  passkey_created_at TIMESTAMP,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Required Backend Endpoints

```go
// Passkey Registration Flow
POST /api/passkey/register-challenge
  → Returns challenge + user info for WebAuthn

POST /api/passkey/register-complete
  ← Client sends public key + credential
  → Server stores in database

// Passkey Authentication Flow  
POST /api/passkey/auth-challenge
  → Returns challenge for existing credential

POST /api/passkey/auth-complete
  ← Client sends signed challenge
  → Server verifies signature
  → Returns JWT token

// Passkey Management
GET /api/passkey/status
  → Check if user has passkey registered

DELETE /api/passkey/remove
  → Remove passkey from account
```

### Required Frontend Changes

1. **User Profile Page**
   ```jsx
   <div className="passkey-setup">
     {user.hasPasskey ? (
       <>
         <div>✅ Passkey Active</div>
         <button onClick={removePasskey}>Remove Passkey</button>
       </>
     ) : (
       <button onClick={setupPasskey}>
         Add Passkey (Face ID / Touch ID)
       </button>
     )}
   </div>
   ```

2. **Login Page Enhancement**
   ```jsx
   useEffect(() => {
     // Check if passkey available
     if (await isPasskeyAvailable()) {
       setShowPasskeyPrompt(true);
     }
   }, []);

   const loginWithPasskey = async () => {
     // Call WebAuthn
     const credential = await navigator.credentials.get({...});
     // Send to server
     const {token} = await verifyPasskey(credential);
     // Done!
   };
   ```

3. **QR Code Behavior**
   ```javascript
   // When QR scanned on mobile:
   if (device.hasPasskey && device.hasCamera) {
     showPasskeyPrompt(); // "Use Face ID?"
   } else {
     showLoginForm(); // Traditional email/code
   }
   ```

### iOS Implementation Notes

**WebAuthn Support:**
- iOS 16+ (September 2022)
- Safari, Chrome, all browsers
- Synced via iCloud Keychain
- Works across devices!

**User Experience:**
```
Setup Passkey:
"PubGames would like to save a passkey for admin@pubgames.local"
[Use Face ID] [Cancel]

Login with Passkey:
"Use passkey for admin@pubgames.local?"
[Continue] [Use Password Instead]
→ Face ID prompt
→ Logged in!
```

### Security Considerations

1. **Private Key Never Leaves Device**
   - Server only stores public key
   - Impossible to phish
   - No passwords to leak

2. **Biometric Required**
   - Face ID or Touch ID mandatory
   - Device PIN as fallback
   - Local authentication only

3. **Anti-Replay Protection**
   - Challenge-response prevents replay attacks
   - Counter prevents reuse of old signatures
   - Time-based validation

4. **Fallback Authentication**
   - Email/code still works
   - Users not forced to use passkeys
   - Can remove passkey anytime

### Migration Strategy

**Phase 1: QR Code** ✅ COMPLETE
- Desktop shows QR
- Mobile scans to access
- Traditional login only

**Phase 2a: Passkey UI** 🔜 NEXT
- Add "Setup Passkey" in user profile
- Store passkey data
- Don't change login flow yet

**Phase 2b: Passkey Login**
- Auto-detect passkey on mobile
- Show "Use Face ID?" prompt
- Fall back to email/code

**Phase 3: Enhancement**
- Passkey sync across devices (via iCloud)
- Passkey on desktop (USB keys, built-in sensors)
- Analytics on usage

### Libraries & Resources

**Backend (Go):**
- `github.com/go-webauthn/webauthn` - WebAuthn server library
- Handles challenge generation, verification
- Production-ready

**Frontend (React):**
- Native `navigator.credentials` API
- No library needed!
- Well-supported by browsers

**Testing:**
- Use real iOS device (simulator doesn't support Face ID for WebAuthn)
- Can test with USB security key on desktop
- Chrome DevTools has WebAuthn tab for debugging

### Future Enhancements

1. **Multiple Passkeys Per User**
   - Register passkey on phone + laptop
   - User chooses which to use

2. **Passkey Sync**
   - iCloud Keychain automatic sync
   - Works across user's Apple devices

3. **Desktop Passkeys**
   - Face ID on MacBook
   - Touch ID on Mac keyboard
   - USB security keys (YubiKey, etc)

4. **QR + Passkey Combo**
   - Scan QR → Instant Face ID → Logged in
   - Zero typing required!
   - Ultimate pub-friendly UX

---

## Current Status

### ✅ Completed (Priority 1)
- [x] Server IP detection endpoint
- [x] QR code generation on login page
- [x] Responsive layout (desktop only)
- [x] CORS configuration for local network
- [x] Database schema with passkey columns
- [x] Documentation

### 🔜 Next Steps (Priority 2)
- [ ] WebAuthn backend endpoints
- [ ] Passkey registration UI
- [ ] Passkey authentication flow
- [ ] Auto-detection on mobile
- [ ] Face ID/Touch ID prompts
- [ ] iOS testing

### 📝 Nice to Have (Future)
- [ ] Passkey management page
- [ ] Multiple passkeys support
- [ ] Desktop passkey support
- [ ] Usage analytics
- [ ] Passkey migration tools

---

## Testing Instructions

### Testing QR Code (Now)

1. **Start services:**
   ```bash
   cd /home/andrew/pubgames-v2
   ./start_services.sh
   ```

2. **Desktop:**
   - Open http://localhost:30000
   - Verify QR code appears on left
   - Note the IP address shown

3. **Mobile (same WiFi):**
   - Open camera app
   - Point at QR code
   - Tap notification to open link
   - Should see login page
   - Login works normally

4. **Verify:**
   - Mobile stays logged in
   - Can access all apps
   - SSO tokens work

### Testing Passkeys (Future)

Will require:
- Real iOS 16+ device
- HTTPS (WebAuthn requirement) OR localhost
- Face ID/Touch ID enabled
- Safari or iOS Chrome

---

## Architecture Benefits

### Current (QR Code)
✅ **Instant mobile access** - No typing URLs  
✅ **Same WiFi only** - Secure, local  
✅ **Professional UX** - Clean, modern  
✅ **Foundation ready** - Database schema in place  

### Future (Passkeys)
🔐 **Ultimate security** - No passwords to steal  
⚡ **Instant login** - Face ID → Done  
📱 **Mobile-first** - Perfect for pub setting  
🔄 **Sync across devices** - iCloud Keychain  
🎯 **Zero typing** - QR scan + Face ID  

---

## Questions & Answers

**Q: Does QR code work over internet?**  
A: No, same WiFi required. Server IP is local (192.168.x.x).

**Q: What if WiFi changes?**  
A: QR code auto-updates with new IP.

**Q: Can multiple people scan same QR?**  
A: Yes! Each gets their own login session.

**Q: Is passkey sync automatic?**  
A: Yes, via iCloud Keychain (iOS 16+).

**Q: What if user loses phone?**  
A: Traditional email/code still works. Or setup new passkey on new device.

**Q: Do passkeys work on Android?**  
A: Yes! Android 9+ with Google Password Manager.

**Q: Can I force all users to use passkeys?**  
A: Not recommended. Always keep email/code as fallback.

---

## Conclusion

**Priority 1 (QR Code)** provides immediate value:
- Scan → Login on phone
- Professional, modern UX
- Zero configuration needed

**Priority 2 (Passkeys)** is the future:
- Scan → Face ID → Instant login
- Maximum security
- Perfect for pub environment

The architecture is designed with both in mind, making the transition smooth and natural. Database schema is ready, CORS is configured, and the UI has placeholders for passkey features.

Ready to implement passkeys whenever you are! 🚀
