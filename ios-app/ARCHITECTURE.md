# iOS App Architecture - PubGames Mini Apps

## Phase 1: Foundation (Current Implementation)

### Objective
Create a lightweight native iOS app that acts as a dedicated browser for web-based mini apps, with architecture ready for future enhancements (caching, biometrics, digital wallets).

### Architecture Patterns

#### 1. **MVVM with Services Layer**

```
┌─────────────────────────────────────────────────┐
│              Views (SwiftUI)                     │
│  - LoginView                                     │
│  - RegisterView                                  │
│  - LauncherView                                  │
│  - WebViewContainer                              │
└─────────────────┬───────────────────────────────┘
                  │ observes
┌─────────────────▼───────────────────────────────┐
│         Services (@Published)                    │
│  - AuthService    (authentication)               │
│  - AppService     (app discovery)                │
└─────────────────┬───────────────────────────────┘
                  │ uses
┌─────────────────▼───────────────────────────────┐
│              Utilities                           │
│  - KeychainHelper (secure storage)               │
│  - Config         (configuration)                │
└──────────────────────────────────────────────────┘
```

#### 2. **Service Layer Design**

**AuthService** - Singleton with `@Published` properties
- Manages authentication state globally
- Handles login, register, logout, token validation
- Uses Keychain for secure token storage
- Publishes `isAuthenticated` for reactive UI updates

**AppService** - Singleton for app discovery
- Fetches mini apps from Identity Service
- Manages app list state
- Handles refresh and error states

**Why Singleton Pattern?**
- Authentication state is global across app
- Prevents multiple auth states
- Simplifies dependency injection
- Easy to access from any view

#### 3. **Security Architecture**

```
┌─────────────────────────────────────────────────┐
│               iOS Keychain                       │
│  - authToken     (JWT)                           │
│  - refreshToken  (future)                        │
│  - userID                                        │
│  - username                                      │
└─────────────────────────────────────────────────┘
         ▲
         │ KeychainHelper
         │
┌────────┴──────────────────────────────────────┐
│            AuthService                         │
│  - Encrypts sensitive data                     │
│  - Per-device storage                          │
│  - Survives app deletion (device level)        │
└────────────────────────────────────────────────┘
```

**Security Features:**
- Keychain uses hardware encryption (Secure Enclave on modern devices)
- Tokens never stored in UserDefaults or files
- Automatic cleanup on logout
- `kSecAttrAccessibleWhenUnlockedThisDeviceOnly` - tokens only accessible when device unlocked

#### 4. **WebView Integration**

```swift
WebViewContainer
  ├─ WKWebView (UIKit wrapped in SwiftUI)
  ├─ NavigationDelegate (loading states, errors)
  ├─ Token Injection (URL query parameter)
  └─ JavaScript Bridge (future: native features)
```

**SSO Flow:**
1. User authenticated in native app → JWT stored in Keychain
2. User taps mini app → `LauncherView` passes app to `WebViewContainer`
3. Container builds URL with `?token=JWT`
4. WebView loads mini app with token
5. Mini app validates token, removes from URL, stores in localStorage

**Future JavaScript Bridge:**
```javascript
// Injected by native app
window.NATIVE_APP = true;
window.NATIVE_PLATFORM = 'ios';

// Future handlers
window.nativeApplePay(amount, description) → calls native code
window.nativeFaceID() → calls native biometric auth
```

#### 5. **Navigation Flow**

```
App Launch
    │
    ├─ Has Token? ──No──> LoginView
    │                        │
    │                        └──> RegisterView (sheet)
    │
    ├─ Yes ──> Validate Token
                    │
                    ├─ Valid ──> LauncherView
                    │              │
                    │              └──> WebViewContainer (sheet)
                    │                     │
                    │                     └──> SettingsView (sheet)
                    │
                    └─ Invalid ──> Logout → LoginView
```

**State Management:**
- `@StateObject` for service instances (owned by view)
- `@EnvironmentObject` for shared auth state
- `@Published` for reactive state changes
- SwiftUI automatically re-renders on state changes

### Technical Decisions

#### Why SwiftUI?
- Declarative UI (less code)
- Native animations and transitions
- Great for rapid prototyping
- Easy to integrate with UIKit (WKWebView)
- Future-proof (Apple's direction)

#### Why iOS 15+?
- Access to modern async/await
- Better SwiftUI features (task modifier, async URLSession)
- Still covers 95%+ of active devices (as of 2024)

#### Why Phase 1 Loads from URLs?
- Zero backend changes needed
- Immediate value - app works today
- Validates architecture before complex caching
- Easy to test and debug

### File Organization

```
PubGamesMiniApps/
├── App/
│   └── PubGamesMiniAppsApp.swift          # Entry point, routing logic
├── Models/
│   ├── User.swift                          # Data models matching backend
│   └── MiniApp.swift
├── Services/
│   ├── AuthService.swift                   # Business logic
│   └── AppService.swift
├── Views/
│   ├── Auth/                               # Feature-based grouping
│   │   ├── LoginView.swift
│   │   └── RegisterView.swift
│   ├── Launcher/
│   │   ├── LauncherView.swift
│   │   └── SettingsView.swift
│   └── WebView/
│       └── WebViewContainer.swift
└── Utilities/
    ├── KeychainHelper.swift                # Reusable helpers
    └── Config.swift
```

### Extension Points for Future Phases

#### Phase 2: Bundle Download & Caching

**New Services:**
```swift
class BundleService {
    func downloadBundle(for app: MiniApp) async throws -> URL
    func getCachedBundle(for app: MiniApp) -> URL?
    func checkForUpdates() async throws -> [MiniApp]
}

class CacheManager {
    func saveBundle(data: Data, for app: MiniApp)
    func getBundle(for app: MiniApp) -> Data?
    func clearCache()
    func getCacheSize() -> Int
}
```

**Changes Needed:**
1. Add `BundleService` to fetch bundles from new backend endpoints
2. Modify `WebViewContainer` to load from local file:// URLs
3. Add version tracking in UserDefaults or local JSON
4. Add background download tasks for updates

#### Phase 3: Biometrics

**Integration Points:**
```swift
import LocalAuthentication

extension AuthService {
    func enableBiometrics() async throws {
        let context = LAContext()
        var error: NSError?

        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            throw AuthError.biometricsNotAvailable
        }

        // Store flag in Keychain
        Config.biometricsEnabled = true
    }

    func authenticateWithBiometrics() async throws {
        let context = LAContext()
        let reason = "Authenticate to access PubGames"

        let success = try await context.evaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics,
            localizedReason: reason
        )

        if success {
            // Load token from Keychain and validate
            await validateStoredToken()
        }
    }
}
```

**Changes Needed:**
1. Add Face ID authentication option to LoginView
2. Store biometric preference in Keychain
3. Add biometric auth on app foreground (if enabled)
4. Fallback to password login

#### Phase 4: Apple Pay

**JavaScript Bridge:**
```swift
// In WebViewContainer.Coordinator
class Coordinator: NSObject, WKScriptMessageHandler {
    func userContentController(_ controller: WKUserContentController,
                               didReceive message: WKScriptMessage) {
        if message.name == "applePay" {
            handleApplePay(message.body)
        }
    }

    func handleApplePay(_ body: Any) {
        // Use PassKit to process payment
        // Return result to JavaScript via evaluateJavaScript
    }
}
```

**JavaScript Side (in mini app):**
```javascript
if (window.NATIVE_APP && window.nativeApplePay) {
    const result = await window.nativeApplePay({
        amount: 10.00,
        currency: 'USD',
        description: 'Sweepstakes Entry'
    });
    // Handle result
}
```

### Performance Considerations

**Current (Phase 1):**
- Network latency: Full page load on each app launch
- No offline support
- Multiple HTTP requests for assets

**Future (Phase 2+):**
- First load: Download bundle (~2MB typical React app)
- Subsequent loads: Instant (local files)
- Updates: Delta downloads only
- Offline: Full functionality except API calls

### Testing Strategy

**Unit Tests:**
```swift
@testable import PubGamesMiniApps

class AuthServiceTests: XCTestCase {
    func testLoginSuccess() async throws {
        let authService = AuthService()
        let user = try await authService.login(username: "test", password: "test123")
        XCTAssertEqual(user.username, "test")
        XCTAssertTrue(authService.isAuthenticated)
    }
}
```

**UI Tests:**
```swift
class LoginFlowTests: XCTestCase {
    func testLoginFlow() throws {
        let app = XCUIApplication()
        app.launch()

        let usernameField = app.textFields["Username"]
        usernameField.tap()
        usernameField.typeText("testuser")

        let passwordField = app.secureTextFields["Password"]
        passwordField.tap()
        passwordField.typeText("password123")

        app.buttons["Login"].tap()

        XCTAssertTrue(app.navigationBars["PubGames"].exists)
    }
}
```

### Deployment

**Development:**
- Xcode builds for simulator or connected device
- Hot reload via Xcode
- Debug via Xcode console and breakpoints

**TestFlight (Beta):**
- Upload via Xcode or CI/CD
- Internal testing (up to 100 users)
- External testing (up to 10,000 users)
- No App Store review required

**App Store:**
- Submit via App Store Connect
- Review process (1-3 days typically)
- Requires: Screenshots, description, privacy policy
- Versioning: Update CFBundleShortVersionString in Info.plist

### Conclusion

Phase 1 establishes a solid foundation with:
- ✅ Modern SwiftUI architecture
- ✅ Secure authentication with Keychain
- ✅ Service layer ready for expansion
- ✅ WebView integration with token injection
- ✅ Clean separation of concerns
- ✅ Extension points for future features

The architecture is designed to evolve:
- Phase 2 adds caching without breaking existing code
- Phase 3 adds biometrics via extension to AuthService
- Phase 4 adds Apple Pay via WebView bridge extension

All without major refactoring of the core architecture.
