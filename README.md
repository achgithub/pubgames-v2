# PubGames V2

A clean, consistent, and scalable architecture for the PubGames ecosystem.

## 🎯 Overview

PubGames V2 is a complete rebuild with:
- **ONE architecture pattern** for all apps (including Identity Service)
- **Modular Go backend** with clean separation of concerns
- **React frontend** with hot reload everywhere
- **SSO authentication** via JWT tokens
- **Template-based app creation** for rapid development

## 📁 Structure

```
/pubgames-v2/
├── identity-service/     # Central authentication hub (Port 3001/30000)
├── template/             # Standard template for new apps
├── shared/              # Shared libraries
│   └── auth/           # Token validation middleware
├── start_services.sh    # Start all services
├── stop_services.sh     # Stop all services
└── new_app.sh          # Create new app from template
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **Node.js 18+**
- **npm**

### Installation

1. **Clone or navigate to this directory**:
   ```bash
   cd /home/andrew/pubgames-v2
   ```

2. **Start all services**:
   ```bash
   ./start_services.sh
   ```

3. **Access the system**:
   - Open http://localhost:30000
   - Login with default admin:
     - Email: `admin@pubgames.local`
     - Code: `123456`

4. **Stop services when done**:
   ```bash
   ./stop_services.sh
   ```

## 🏗️ Architecture

### Port Scheme

```
3001   - Identity Service Backend (Go API)
30000  - Identity Service Frontend (React)
30010  - App 1 Frontend
30011  - App 1 Backend
30020  - App 2 Frontend
30021  - App 2 Backend
...    - Up to 99 apps (3009X)
```

**Rule**: Backend = XX1 (odd), Frontend = XX0 (even ending in 0)

### Standard App Structure

Every app (including Identity Service) follows this structure:

```
/app-name/
├── main.go           # Entry point, routing
├── handlers.go       # HTTP handlers
├── models.go         # Data structures
├── database.go       # DB init and schema
├── auth.go          # Uses shared/auth library
├── /src/            # React source
│   ├── index.js
│   └── App.js       # SSO built-in
├── /public/         # Static files
├── /data/           # SQLite database
├── package.json     # NPM config
└── go.mod           # Go module
```

### SSO Flow

1. User logs into Identity Service (port 30000)
2. Gets JWT token
3. Clicks app tile
4. Redirects to: `http://localhost:30X0?token=JWT`
5. App detects `?token=` in URL
6. Validates with Identity Service: `GET /api/validate-token`
7. Auto-logs user in
8. Removes `?token=` from URL

## 🎨 Creating a New App

### Using the Script (Recommended)

```bash
./new_app.sh
```

Follow the prompts to create a new app with:
- Custom name and description
- App number (determines ports)
- Icon emoji
- Auto-configured from template

### Manual Creation

1. Copy the template:
   ```bash
   cp -r template/ my-new-app/
   cd my-new-app/
   ```

2. Replace placeholders in these files:
   - `main.go`: Update `BACKEND_PORT`, `FRONTEND_PORT`
   - `package.json`: Update `PORT` in start script, name, description
   - `go.mod`: Update module name
   - `src/App.js`: Update `API_BASE`
   - `database.go`: Update database filename

3. Install dependencies:
   ```bash
   go mod download
   npm install
   ```

4. Start your app:
   ```bash
   # Terminal 1
   go run *.go
   
   # Terminal 2
   npm start
   ```

## 🔧 Development

### Running Individual Apps

**Backend** (in app directory):
```bash
go run *.go
```

**Frontend** (in app directory, separate terminal):
```bash
npm start
```

### Hot Reload

- **Frontend**: Automatic via React dev server
- **Backend**: Restart with `go run *.go`

### Testing

**Backend**:
```bash
go test ./...
```

**Frontend**:
```bash
npm test
```

## 📝 Components

### Identity Service

Central authentication and app launcher.

**Features**:
- User registration and login
- JWT token generation
- Token validation endpoint
- App directory management
- Admin panel

**API Endpoints**:
- `POST /api/register` - Create new user
- `POST /api/login` - Authenticate user
- `GET /api/validate-token` - Validate JWT token
- `GET /api/apps` - List available apps
- `GET /api/admin/apps` - Admin: Manage apps
- `GET /api/admin/users` - Admin: View users

### Template App

Standard template with:
- Modular Go backend (main, handlers, models, database, auth)
- React frontend with SSO
- Sample CRUD operations
- Protected and admin routes
- SQLite database

### Shared Auth Library

Located in `/shared/auth/`

**Provides**:
- `AuthMiddleware`: Validates JWT tokens
- `AdminMiddleware`: Requires admin privileges
- `GetUser`: Retrieves user from context

**Usage**:
```go
import "pubgames/shared/auth"

authMw := auth.AuthMiddleware(auth.Config{
    IdentityServiceURL: "http://localhost:3001",
})

api.HandleFunc("/protected", authMw(handler))
```

## 🔒 Security

### JWT Tokens

- **Expiry**: 24 hours
- **Secret**: Change `JWT_SECRET` in production
- **Validation**: Every request to protected routes
- **Storage**: Client-side in localStorage

### Passwords

- **Hashing**: bcrypt with cost 12
- **6-character codes**: For easy entry in pub setting

### CORS

Configured in each app's `main.go`:
```go
handlers.AllowedOrigins([]string{
    "http://localhost:30000",  // Adjust for your app
})
```

## 📊 Database

### SQLite Per App

Each app has its own SQLite database in `/data/`:
- Identity Service: `identity.db`
- Template: `app.db`
- Custom apps: `{app-name}.db`

### Schema Management

Define schema in each app's `database.go`:
```go
schema := `
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
```

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Check what's using a port
lsof -i :3001

# Kill process on port
kill -9 $(lsof -ti:3001)

# Or use stop script
./stop_services.sh
```

### Backend Won't Start

```bash
# Check Go installation
go version

# Clean and reinstall dependencies
rm go.sum
go mod download
```

### Frontend Won't Start

```bash
# Clean and reinstall
rm -rf node_modules package-lock.json
npm install
```

### SSO Not Working

1. Verify Identity Service is running on port 3001
2. Check browser console for errors
3. Verify token in localStorage
4. Test validation endpoint directly:
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" \
        http://localhost:3001/api/validate-token
   ```

## 📚 Documentation Files

- `QUICK-REFERENCE-CARD.md` - Quick reference for architecture
- `CLEAN-ARCHITECTURE-DESIGN.md` - Full design specification
- `CURRENT-STATE-ANALYSIS.md` - Analysis of old system
- `REDESIGN-SUMMARY.md` - Migration guide

## 🎯 Design Principles

1. **Consistency**: Same structure for every app
2. **Simplicity**: Minimum moving parts
3. **Modularity**: Clean separation of concerns
4. **Reliability**: Scripts that always work
5. **Speed**: Hot reload everywhere
6. **Scalability**: Easy to add new apps

## ✅ Success Criteria

You'll know it's working when:

- ✅ New app created in < 5 minutes
- ✅ All apps start with single command
- ✅ SSO works for all apps
- ✅ Hot reload works everywhere
- ✅ No manual build steps needed
- ✅ Scripts never fail
- ✅ File structure identical across apps

## 🚧 Next Steps

### Migrating Existing Apps

1. Create new app from template
2. Copy database from old app
3. Port business logic to new structure
4. Update React components
5. Test SSO integration
6. Switch to new version

### Adding Features

1. Update template first
2. Test thoroughly
3. Apply to existing apps
4. Document in README

## 📞 Support

For issues or questions:
1. Check troubleshooting section
2. Review design documentation
3. Examine template implementation
4. Test with fresh template app

## 🎉 Credits

Built with:
- Go (Gorilla Mux)
- React
- SQLite
- JWT
- bcrypt

---

**Remember**: One architecture, consistent everywhere. If it works in the template, it works in all apps.
