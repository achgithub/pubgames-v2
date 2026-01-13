# Quick Service Management Reference

## Setup (One Time)
```bash
cd /home/andrew/pubgames-v2
chmod +x *.sh
```

## Daily Commands

### Start Everything
```bash
./start_services.sh
```
Then open: http://localhost:30000

### Check What's Running
```bash
./status_services.sh
```

### Stop Everything
```bash
./stop_services.sh
```

## Debugging

### View Logs
```bash
# Real-time
tail -f logs/Identity-Service-backend.log

# All recent errors
grep -i error logs/*.log

# Last 50 lines of all logs
tail -50 logs/*.log
```

### Check Specific Port
```bash
lsof -i :3001
```

### Kill Stuck Process
```bash
kill -9 $(lsof -ti:3001)
```

## File Locations

| What | Where |
|------|-------|
| Scripts | `/home/andrew/pubgames-v2/*.sh` |
| PID files | `/home/andrew/pubgames-v2/.pids/` |
| Log files | `/home/andrew/pubgames-v2/logs/` |

## Port Map

| Service | Backend | Frontend |
|---------|---------|----------|
| Identity Service | 3001 | 30000 |
| Smoke Test | 30011 | 30010 |
| Last Man Standing | 30021 | 30020 |
| Sweepstakes | 30031 | 30030 |

## Startup Flow

1. **start_services.sh** runs
2. Checks port availability
3. Installs dependencies if needed
4. Starts backend → waits for port to open
5. Starts frontend → waits for port to open
6. Reports success ✓ or failure ✗
7. Creates PID files in `.pids/`
8. Logs output to `logs/`

## Key Improvements

✓ No terminal emulator needed  
✓ Proper health checks  
✓ PID file tracking  
✓ Automatic logging  
✓ Status monitoring  
✓ Graceful shutdown  
✓ Clear error messages  

## What Changed

### Old (Unreliable)
- Tried to open gnome-terminal/xterm
- Failed silently if not available
- No process tracking
- No health checks
- No logs

### New (Reliable)
- Background processes with nohup
- PID files for tracking
- Health checks (waits for ports)
- All output logged
- Status script shows what's running
- Graceful shutdown with cleanup

## Common Workflows

### Development
```bash
# Morning
./start_services.sh
./status_services.sh  # verify all running

# Work on code...

# Evening
./stop_services.sh
```

### Debugging Service
```bash
# Stop everything
./stop_services.sh

# Start manually to see output
cd identity-service
go run *.go  # Terminal 1
npm start    # Terminal 2
```

### Clean Start
```bash
./stop_services.sh
rm -rf .pids/* logs/*
./start_services.sh
```

## Exit Codes

- `0` = Success
- `1` = Critical service (Identity) failed
- Script continues even if non-critical services fail

## Help

Full guide: `SERVICE-MANAGEMENT-GUIDE.md`
