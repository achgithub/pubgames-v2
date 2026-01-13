# PubGames Last Man Standing

This is the standard template for creating new apps in the PubGames ecosystem.

## Architecture

- **Backend**: Go API (Port 30021)
- **Frontend**: React dev server (Port 30020)
- **Database**: SQLite (./data/app.db)
- **Authentication**: SSO via Identity Service

## File Structure

```
/template/
├── main.go           # Entry point, routing
├── handlers.go       # HTTP handlers
├── models.go         # Data structures
├── database.go       # DB initialization
├── auth.go           # Uses shared/auth library
├── /src/            # React source
│   ├── index.js
│   └── App.js       # SSO built-in
├── /public/         # Static files
│   └── index.html
├── /data/           # SQLite database
├── package.json     # NPM config
└── go.mod           # Go module
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Installation

1. **Install Go dependencies**:
   ```bash
   go mod download
   ```

2. **Install npm dependencies**:
   ```bash
   npm install
   ```

### Running

1. **Start backend** (in one terminal):
   ```bash
   go run *.go
   ```

2. **Start frontend** (in another terminal):
   ```bash
   npm start
   ```

The app will be available at:
- Frontend: http://localhost:30020
- Backend API: http://localhost:30021

## SSO Flow

1. User logs into Identity Service (http://localhost:3001)
2. Clicks app tile
3. Gets redirected to: `http://localhost:30020?token=JWT`
4. App auto-validates token and logs user in

## Customization

To create a new app from this template:

1. Copy the template directory
2. Replace placeholders in these files:
   - `main.go`: Update BACKEND_PORT, FRONTEND_PORT
   - `package.json`: Update PORT in start script, name, description
   - `go.mod`: Update module name
   - `src/App.js`: Update API_BASE
   - `README.md`: Update app name and description

3. Update the database schema in `database.go`
4. Add your business logic to `handlers.go`
5. Define your data models in `models.go`

## API Endpoints

### Public
- `GET /api/config` - App configuration

### Protected (requires authentication)
- `GET /api/data` - Sample data
- `GET /api/items` - List items
- `POST /api/items` - Create item

### Admin (requires admin role)
- `GET /api/admin/stats` - Admin statistics

## Database

SQLite database at `./data/app.db`

Default tables:
- `users` - Local user reference
- `items` - Sample items

Add your app-specific tables in `database.go`

## Development

### Hot Reload
- Backend: Restart with `go run *.go`
- Frontend: Automatic via React dev server

### Testing
```bash
# Backend tests
go test ./...

# Frontend tests
npm test
```

## Deployment

```bash
# Build frontend
npm run build

# Build backend
go build -o app

# Run
./app
```

## Notes

- All authentication is handled by Identity Service
- Use shared/auth library for token validation
- Follow the standard port scheme (XX0/XX1)
- Keep Go files modular (don't put everything in main.go)
