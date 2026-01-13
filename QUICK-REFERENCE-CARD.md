# PubGames V2 - Quick Start Card

**Copy this to the new directory for instant reference**

---

## The ONE Architecture Pattern

```
Every App (including Identity Service):

/app-name/
├── main.go           # Routing only
├── handlers.go       # HTTP handlers
├── models.go         # Data structures
├── database.go       # DB init + schema
├── auth.go           # Auth (uses shared lib)
├── /src/            # React source
│   └── App.js       # SSO detection built-in
├── /data/           # SQLite database
├── package.json     # NPM config
└── go.mod           # Go module
```

---

## Port Rules

```
Backend:  XX1 (odd, e.g., 3001, 30011, 30021)
Frontend: XX0 (even ending in 0, e.g., 30000, 30010, 30020)

Identity:  3001 (backend) + 30000 (frontend)
App 1:     30011 (backend) + 30010 (frontend)
App 2:     30021 (backend) + 30020 (frontend)
```

---

## SSO Flow

```
1. User logs into Identity (3001)
2. Gets JWT token
3. Clicks app tile
4. Redirects to: http://localhost:30X0?token=JWT
5. App detects ?token= in URL
6. Validates with Identity: GET /api/validate-token
7. Logs user in
8. Removes ?token= from URL
```

---

## File Responsibilities

| File | Purpose | Size |
|------|---------|------|
| main.go | Entry point, routing setup | ~100 lines |
| handlers.go | HTTP request handlers | ~300 lines |
| models.go | Data structures (structs) | ~100 lines |
| database.go | DB init, schema, migrations | ~150 lines |
| auth.go | Uses shared/auth library | ~50 lines |

---

## Standard Commands

```bash
# Start backend
go run *.go

# Start frontend (separate terminal)
npm start

# Build frontend (not needed in dev)
npm run build

# Test backend
go test ./...

# Test frontend
npm test
```

---

## Template Checklist

Before declaring template done:

- [ ] Compiles without errors
- [ ] Backend starts on XX1
- [ ] Frontend starts on XX0
- [ ] SSO token detection works
- [ ] Token validation works
- [ ] Protected routes work
- [ ] Admin routes work
- [ ] Hot reload works
- [ ] Database creates
- [ ] Can copy to create new app

---

## Critical Constants

Every app needs these in main.go:

```go
const (
    BACKEND_PORT       = "30X1"  // Change X
    FRONTEND_PORT      = "30X0"  // Change X
    DB_PATH            = "./data/app.db"
    IDENTITY_SERVICE   = "http://localhost:3001"
)
```

---

## React SSO Template

```javascript
useEffect(() => {
  // Check for ?token= in URL
  const params = new URLSearchParams(window.location.search);
  const token = params.get('token');
  
  if (token) {
    validateToken(token);
    window.history.replaceState({}, '', window.location.pathname);
    return;
  }
  
  // Check localStorage
  const saved = localStorage.getItem('user');
  if (saved) {
    setUser(JSON.parse(saved));
  }
}, []);

const validateToken = async (token) => {
  const res = await fetch('http://localhost:3001/api/validate-token', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (res.ok) {
    const user = await res.json();
    setUser(user);
    localStorage.setItem('user', JSON.stringify(user));
  }
};
```

---

## Shared Library

```
/shared/
└── /auth/
    ├── go.mod
    ├── middleware.go
    └── identity.go

Usage in app:
import "pubgames/shared/auth"

authMw := auth.AuthMiddleware(auth.Config{
    IdentityServiceURL: IDENTITY_SERVICE,
})

api.HandleFunc("/protected", authMw(handler))
```

---

## Database Standard

Every app's database.go:

```go
func initDB() {
    var err error
    db, err = sql.Open("sqlite3", DB_PATH)
    if err != nil {
        log.Fatal(err)
    }
    
    schema := `
    CREATE TABLE IF NOT EXISTS users (...);
    -- Add app-specific tables
    `
    
    _, err = db.Exec(schema)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## Don't Forget

1. ✅ Every app must validate tokens with Identity
2. ✅ Frontend must detect ?token= in URL
3. ✅ Backend must use shared/auth library
4. ✅ Ports must follow XX0/XX1 pattern
5. ✅ Go files must be modular (not one huge file)
6. ✅ React must have hot reload (dev server)
7. ✅ CORS must allow frontend port
8. ✅ All apps must look/feel same

---

## If Something Breaks

1. Stop all services
2. Check ports are free: `lsof -i :PORT`
3. Check Go compiles: `go run *.go`
4. Check React starts: `npm start`
5. Check logs in terminal
6. Verify constants (ports, URLs)
7. Test SSO step by step

---

## Build Order

1. **Template first** - Get one app working perfectly
2. **Identity Service** - Rebuild to match template
3. **Test SSO** - Prove token flow works
4. **Migrate App 1** - Use template pattern
5. **Migrate App 2** - Use template pattern
6. **Only then** - Consider new apps

---

## Success = Simplicity

```
1 template
  × N apps
  = consistent system

NOT:

N apps
  × M patterns
  = chaos
```

---

**Keep this card visible while building!**
