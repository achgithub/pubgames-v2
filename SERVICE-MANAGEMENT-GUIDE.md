# Service Management Improvements

## What Changed

The startup system has been completely redesigned to be reliable and robust:

### Key Improvements

1. **Background Process Management**
   - Services run as proper background processes (no terminal emulator dependency)
   - PID files track all processes in `.pids/` directory
   - Proper cleanup on shutdown

2. **Health Checks**
   - Waits for ports to actually be in use before marking as started
   - 30-second timeout for backends, 60-second timeout for frontends
   - Clear error messages if services fail to start

3. **Better Error Handling**
   - Continues starting other services even if one fails
   - Non-critical services (like smoke-test) don't block startup
   - Clear success/failure reporting

4. **Logging**
   - All output captured in `logs/` directory
   - Separate logs for each service's backend and frontend
   - Easy to debug with `tail -f logs/<service>-backend.log`

5. **Status Monitoring**
   - New `status_services.sh` script shows what's running
   - Displays PIDs and port status
   - Detects orphaned processes

## Setup

Make the scripts executable (run once):

```bash
cd /home/andrew/pubgames-v2
chmod +x start_services.sh stop_services.sh status_services.sh
```

## Usage

### Start All Services

```bash
./start_services.sh
```

**What it does:**
- Creates `.pids/` and `logs/` directories
- Checks port availability before starting
- Starts Identity Service first (critical)
- Starts other apps (smoke-test, last-man-standing, sweepstakes) if they exist
- Waits for each service to actually start (health check)
- Reports success/failure for each service
- Creates PID files for all processes

**Output:**
- Real-time progress as each service starts
- Clear ✓ or ✗ for each component
- URLs to access services
- Log file locations

### Stop All Services

```bash
./stop_services.sh
```

**What it does:**
- Gracefully stops processes using PID files
- Waits 5 seconds for graceful shutdown
- Force kills if needed
- Cleans up known ports as fallback
- Removes PID files

### Check Status

```bash
./status_services.sh
```

**What it shows:**
- Which services are running (✓) or stopped (✗)
- Port numbers and PIDs for each service
- PID file mismatches (if any)
- Orphaned processes
- URLs to access running services
- Log file locations

## How It Works

### Process Management

1. **Starting a Service:**
   ```
   Check port available → Install dependencies → Start in background
   → Save PID → Wait for port to be in use → Verify started
   ```

2. **Health Checks:**
   - Polls the port every second
   - Succeeds when port is in LISTEN state
   - Times out with clear error if service doesn't start

3. **PID Tracking:**
   ```
   .pids/
   ├── Identity-Service-backend.pid
   ├── Identity-Service-frontend.pid
   ├── Smoke-Test-backend.pid
   └── Smoke-Test-frontend.pid
   ```

### Logging

All output goes to log files:
```
logs/
├── Identity-Service-backend.log
├── Identity-Service-frontend.log
├── Smoke-Test-backend.log
└── Smoke-Test-frontend.log
```

**View logs in real-time:**
```bash
tail -f logs/Identity-Service-backend.log
```

**View all backend logs:**
```bash
tail -f logs/*-backend.log
```

## Troubleshooting

### Service Won't Start

1. **Check the logs:**
   ```bash
   cat logs/<service>-backend.log
   cat logs/<service>-frontend.log
   ```

2. **Common issues:**
   - Dependencies not installed: Check for "go mod download" or "npm install" errors
   - Port already in use: Run `./status_services.sh` to see what's using ports
   - Database locked: Check if another instance is running

### Port Already in Use

```bash
# See what's using the port
lsof -i :3001

# Stop all services
./stop_services.sh

# If that doesn't work, manually kill
kill -9 $(lsof -ti:3001)
```

### Services Keep Stopping

Check logs for errors:
```bash
tail -100 logs/<service>-backend.log
```

Common causes:
- Database errors
- Missing environment variables
- Go compilation errors
- npm dependency issues

### Orphaned Processes

If `status_services.sh` shows orphaned processes:

```bash
./stop_services.sh  # Should clean them up

# Or manually
pkill -f "go run.*pubgames"
pkill -f "node.*pubgames"
```

## Advantages Over Manual Startup

### Manual (old way)
```
Terminal 1: cd identity-service && go run *.go
Terminal 2: cd identity-service && npm start
Terminal 3: cd smoke-test && go run *.go
Terminal 4: cd smoke-test && npm start
... (8+ terminals for 4 apps)
```

**Issues:**
- Hard to manage multiple terminals
- No way to track PIDs
- Difficult to stop cleanly
- No logs preserved
- Easy to forget which terminals are which

### Automated (new way)
```
./start_services.sh
```

**Benefits:**
- Single command
- All logs captured
- Easy to stop: `./stop_services.sh`
- Check status: `./status_services.sh`
- Health checks ensure services actually started
- Background processes don't clutter terminal

## Advanced Usage

### Start Individual Service

You can still manually start individual services if needed:

```bash
cd identity-service

# Terminal 1 (or background)
go run *.go > ../logs/identity-backend-manual.log 2>&1 &
echo $! > ../.pids/identity-backend-manual.pid

# Terminal 2 (or background)
BROWSER=none npm start > ../logs/identity-frontend-manual.log 2>&1 &
echo $! > ../.pids/identity-frontend-manual.pid
```

### Debug Mode

To see more detail, edit `start_services.sh` and change:
```bash
nohup go run *.go > "$log_file" 2>&1 &
```
to:
```bash
go run *.go 2>&1 | tee "$log_file" &
```

This will show output in terminal AND log file.

### Faster Startup (Skip Health Checks)

If you're sure services will start, you can reduce timeouts in `start_services.sh`:
```bash
# Change these lines:
if wait_for_port $port 30 "$name Backend"; then  # Reduce 30 to 10
if wait_for_port $port 60 "$name Frontend"; then # Reduce 60 to 20
```

## Files Created

```
/home/andrew/pubgames-v2/
├── start_services.sh          # New: reliable startup
├── stop_services.sh           # Updated: PID file support
├── status_services.sh         # New: status monitoring
├── .pids/                     # New: PID tracking
│   ├── Identity-Service-backend.pid
│   └── ...
└── logs/                      # New: all service logs
    ├── Identity-Service-backend.log
    └── ...
```

## Migration from Old Scripts

The old scripts relied on terminal emulators (gnome-terminal/xterm) which:
- Aren't always available
- Fail silently
- Make process management difficult

The new scripts:
- Work everywhere (no terminal emulator needed)
- Track processes properly
- Provide clear feedback
- Are easy to debug

**You can delete the old scripts or keep them as backup.**

## Testing

After setup, verify everything works:

```bash
# Start all services
./start_services.sh

# Check status (should show all ✓)
./status_services.sh

# Open browser to http://localhost:30000
# Login should work

# Stop all services
./stop_services.sh

# Check status (should show all ✗)
./status_services.sh
```

## Next Steps

1. Run `chmod +x *.sh` to make scripts executable
2. Test with `./start_services.sh`
3. Check status with `./status_services.sh`
4. Review logs if any issues: `ls -l logs/`
5. Stop when done: `./stop_services.sh`

The startup process should now be completely reliable! 🚀
