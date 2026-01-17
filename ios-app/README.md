# PubGames Mini Apps - iOS

A lightweight native iOS app that serves as a dedicated browser container for PubGames web-based mini apps.

## Overview

This iOS app provides:

- **Native Authentication** - Login/Register with secure Keychain storage
- **App Discovery** - Fetches available mini apps from Identity Service
- **WebView Container** - Loads mini apps with automatic token injection
- **SSO Integration** - Seamless authentication with JWT tokens
- **Future-Ready** - Architecture prepared for Face ID, Apple Pay, and local caching

## Architecture (Phase 1)

```
┌─────────────────────────────────────────┐
│          Native iOS App                  │
├─────────────────────────────────────────┤
│  LoginView (Native SwiftUI)             │
│  LauncherView (Native Grid)             │
│  WebViewContainer (WKWebView)           │
└─────────────────────────────────────────┘
           ↓ HTTP/HTTPS
┌─────────────────────────────────────────┐
│    Identity Service (Go Backend)        │
│    - POST /api/login                    │
│    - POST /api/register                 │
│    - GET /api/validate-token            │
│    - GET /api/apps                      │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│    Mini Apps (React Frontends)          │
│    - Tic Tac Toe (30040)                │
│    - Last Man Standing (30020)          │
│    - Sweepstakes (30030)                │
└─────────────────────────────────────────┘
```

## Project Structure

```
PubGamesMiniApps/
├── App/
│   └── PubGamesMiniAppsApp.swift      # Main entry point
├── Models/
│   ├── User.swift                      # User and auth models
│   └── MiniApp.swift                   # Mini app model
├── Services/
│   ├── AuthService.swift               # Authentication & token management
│   └── AppService.swift                # App discovery & fetching
├── Views/
│   ├── Auth/
│   │   ├── LoginView.swift             # Native login screen
│   │   └── RegisterView.swift          # Native registration screen
│   ├── Launcher/
│   │   ├── LauncherView.swift          # App grid launcher
│   │   └── SettingsView.swift          # Settings screen
│   └── WebView/
│       └── WebViewContainer.swift      # WebView with token injection
├── Utilities/
│   ├── KeychainHelper.swift            # Secure token storage
│   └── Config.swift                    # App configuration
└── Info.plist
```

## Setup Instructions

### Prerequisites

- macOS with Xcode 14.0 or later
- iOS 15.0+ target device or simulator
- PubGames backend services running (Identity Service + Mini Apps)

### Step 1: Create Xcode Project

1. Open Xcode
2. Create new project: **File → New → Project**
3. Select **iOS → App**
4. Configure project:
   - Product Name: `PubGamesMiniApps`
   - Team: Your development team
   - Organization Identifier: `com.pubgames`
   - Interface: **SwiftUI**
   - Language: **Swift**
   - Minimum Deployments: **iOS 15.0**

### Step 2: Add Source Files

1. Delete the default `ContentView.swift` and `PubGamesMiniAppsApp.swift` that Xcode created
2. In Finder, navigate to `/home/user/pubgames-v2/ios-app/PubGamesMiniApps/PubGamesMiniApps/`
3. Drag all folders (`App/`, `Models/`, `Services/`, `Views/`, `Utilities/`) into your Xcode project
4. When prompted, select:
   - ✅ Copy items if needed
   - ✅ Create groups
   - ✅ Add to target: PubGamesMiniApps

### Step 3: Configure Info.plist

1. Replace the default `Info.plist` with the one from this repository
2. Or manually add these keys to your Info.plist:

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
<key>NSFaceIDUsageDescription</key>
<string>We use Face ID to securely authenticate you and protect your account.</string>
```

**Note:** `NSAllowsArbitraryLoads` is needed for development to allow HTTP connections to localhost. For production, configure specific domains with HTTPS.

### Step 4: Configure Server URL

By default, the app points to `http://localhost:3001` for development.

**For iOS Simulator:**
- Localhost works if backend is running on your Mac

**For Physical Device:**
- Find your Mac's local IP: `ifconfig | grep "inet " | grep -v 127.0.0.1`
- Update `Config.swift` line 16:
  ```swift
  return "http://192.168.1.XXX:3001"  // Replace with your IP
  ```

Or update dynamically in code:
```swift
Config.updateServerURL("http://192.168.1.100:3001")
```

### Step 5: Build and Run

1. Select your target device (iPhone 15 Pro simulator or physical device)
2. Press **Cmd+R** to build and run
3. The app should launch and show the login screen

## Usage

### First Time Setup

1. **Start Backend Services**
   ```bash
   cd /home/user/pubgames-v2
   ./start_services.sh
   ```

2. **Create an Account**
   - Open the iOS app
   - Tap "Don't have an account? Register"
   - Enter username and password
   - Optionally enter email

3. **Browse Mini Apps**
   - After login, you'll see a grid of available apps
   - Tap any app tile to launch it in a WebView
   - The app automatically injects your auth token

### Features

#### Authentication
- ✅ Native login/register screens
- ✅ Secure token storage in iOS Keychain
- ✅ Automatic token validation on app launch
- ✅ Logout functionality

#### App Launcher
- ✅ Fetches apps from `/api/apps` endpoint
- ✅ Grid display with generated icons and colors
- ✅ Pull to refresh
- ✅ Settings menu

#### WebView Container
- ✅ Loads mini apps from web URLs
- ✅ Automatic SSO token injection (`?token=JWT`)
- ✅ Navigation controls (back, forward, reload)
- ✅ Full screen presentation
- ✅ Native bridge script injection (prepared for future features)

### Settings

Access settings via the menu icon (•••) in the top right:
- View user information
- Check server configuration
- Logout

## Development

### Debugging

**Enable verbose logging:**
```swift
// In AuthService.swift or AppService.swift
print("DEBUG: URL: \(url)")
print("DEBUG: Response: \(String(data: data, encoding: .utf8) ?? "nil")")
```

**Check Network Requests:**
- Use **Charles Proxy** or **Proxyman** to inspect HTTP traffic
- View console logs in Xcode: **View → Debug Area → Show Debug Area**

### Common Issues

**"Invalid URL" error:**
- Check `Config.swift` server URL
- Ensure backend is running
- Verify you can access `http://your-ip:3001` from Safari on your device

**"Network error" or timeout:**
- Check that your device is on the same WiFi network
- Ensure firewall allows connections to ports 3001, 30000-30041
- Try accessing the backend URL in Safari first

**"No apps available":**
- Verify `/api/apps` endpoint returns apps
- Check that apps have `is_active = true` in the database
- Ensure you're authenticated (check Keychain for token)

**WebView shows blank page:**
- Check mini app frontend URL is accessible from device
- Inspect Console in Xcode for JavaScript errors
- Verify CORS settings allow requests from the WebView

## Future Enhancements (Phases 2-3)

### Phase 2: Backend Bundle Support
- [ ] Add `/api/apps/manifest` endpoint
- [ ] Add `/api/apps/{id}/bundle` endpoint
- [ ] Create build scripts to package mini apps
- [ ] Add version tracking

### Phase 3: Caching & Offline
- [ ] Download app bundles to local storage
- [ ] Version checking and delta updates
- [ ] Load WebViews from local cache
- [ ] Offline mode support

### Additional Features
- [ ] Face ID / Touch ID authentication
- [ ] Apple Pay integration via JavaScript bridge
- [ ] Push notifications
- [ ] App-specific permissions
- [ ] Dark mode support
- [ ] iPad optimization

## API Integration

### Identity Service Endpoints Used

```
POST /api/login
Body: {"username": "user", "password": "pass"}
Response: {"message": "...", "token": "JWT...", "user": {...}}

POST /api/register
Body: {"username": "user", "password": "pass", "email": "..."}
Response: {"message": "...", "token": "JWT...", "user": {...}}

GET /api/validate-token
Headers: Authorization: Bearer {token}
Response: {"valid": true, "user": {...}}

GET /api/apps
Headers: Authorization: Bearer {token}
Response: {"apps": [{id, name, url, description, ...}]}
```

### SSO Flow

1. User logs in → iOS app receives JWT token
2. User taps mini app → iOS app opens WebView
3. iOS app appends `?token=JWT` to mini app URL
4. Mini app receives token, validates with Identity Service
5. Mini app stores token in localStorage
6. Mini app removes token from URL

## Security

- ✅ Tokens stored in iOS Keychain (encrypted, per-device)
- ✅ HTTPS recommended for production
- ✅ Token validation on every API call
- ✅ Automatic logout on invalid token
- ⚠️ `NSAllowsArbitraryLoads` should be removed for production (use ATS exceptions)

## License

Part of the PubGames V2 project.

## Support

For issues or questions, refer to the main PubGames documentation or create an issue in the repository.
