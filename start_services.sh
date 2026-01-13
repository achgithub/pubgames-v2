#!/bin/bash

# PubGames V2 - Start Services Script
# Reliable background process management with health checks

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

# PID directory
PID_DIR="$BASE_DIR/.pids"
LOG_DIR="$BASE_DIR/logs"

# Create necessary directories
mkdir -p "$PID_DIR" "$LOG_DIR"

echo -e "${BLUE}🚀 PubGames V2 - Starting Services${NC}"
echo "========================================"
echo ""

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1
    else
        return 0
    fi
}

# Function to wait for port to be in use (service started)
wait_for_port() {
    local port=$1
    local timeout=$2
    local service_name=$3
    local count=0
    
    while [ $count -lt $timeout ]; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        count=$((count + 1))
        if [ $((count % 5)) -eq 0 ]; then
            echo -ne "\r  Waiting for $service_name... ${count}s"
        fi
    done
    echo ""
    return 1
}

# Function to wait for React compilation
wait_for_react_ready() {
    local log_file=$1
    local timeout=$2
    local service_name=$3
    local count=0
    
    echo "  Waiting for React to compile..."
    
    while [ $count -lt $timeout ]; do
        # Check for compilation success messages
        if grep -q "Compiled successfully\|webpack compiled\|Compiled with warnings" "$log_file" 2>/dev/null; then
            # Check if there are errors after compilation message
            if ! tail -20 "$log_file" | grep -q "Failed to compile\|Compilation failed"; then
                echo -e "\r  ✓ React compilation complete"
                return 0
            fi
        fi
        
        # Check for compilation errors
        if grep -q "Failed to compile\|Compilation failed" "$log_file" 2>/dev/null; then
            echo ""
            echo -e "${RED}  ✗ React compilation failed${NC}"
            return 1
        fi
        
        sleep 1
        count=$((count + 1))
        
        if [ $((count % 5)) -eq 0 ]; then
            echo -ne "\r  Waiting for React compilation... ${count}s"
        fi
    done
    
    echo ""
    echo -e "${YELLOW}  ⚠ Timeout waiting for compilation (proceeding anyway)${NC}"
    return 0  # Don't fail, just warn
}

# Function to verify frontend is serving content
verify_frontend_ready() {
    local port=$1
    local timeout=10
    local count=0
    
    echo "  Verifying frontend is serving content..."
    
    while [ $count -lt $timeout ]; do
        # Try to fetch the index page
        if curl -s -f -m 2 "http://localhost:$port" > /dev/null 2>&1; then
            echo "  ✓ Frontend responding to requests"
            return 0
        fi
        
        sleep 1
        count=$((count + 1))
    done
    
    echo -e "${YELLOW}  ⚠ Frontend not responding yet (may need a moment)${NC}"
    return 0  # Don't fail, just warn
}

# Function to start backend
start_backend() {
    local name=$1
    local dir=$2
    local port=$3
    local pid_file="$PID_DIR/${name// /-}-backend.pid"
    local log_file="$LOG_DIR/${name// /-}-backend.log"
    
    echo -e "${YELLOW}Starting $name Backend (port $port)...${NC}"
    
    # Check if already running
    if [ -f "$pid_file" ]; then
        local old_pid=$(cat "$pid_file")
        if ps -p $old_pid > /dev/null 2>&1; then
            echo -e "${YELLOW}  Backend already running (PID: $old_pid)${NC}"
            return 0
        else
            rm -f "$pid_file"
        fi
    fi
    
    # Check port availability
    if ! check_port $port; then
        echo -e "${RED}✗ Port $port already in use${NC}"
        return 1
    fi
    
    # Check if Go dependencies are installed
    if [ ! -f "$dir/go.sum" ]; then
        echo "  Installing Go dependencies..."
        (cd "$dir" && go mod download)
    fi
    
    # Start backend in background
    cd "$dir"
    nohup go run *.go > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    cd "$BASE_DIR"
    
    # Wait for service to start
    if wait_for_port $port 30 "$name Backend"; then
        echo -e "${GREEN}✓ Backend started successfully (PID: $pid)${NC}"
        echo "  Log: $log_file"
        return 0
    else
        echo -e "${RED}✗ Backend failed to start within 30 seconds${NC}"
        echo "  Check log: $log_file"
        if [ -f "$pid_file" ]; then
            local failed_pid=$(cat "$pid_file")
            kill -9 $failed_pid 2>/dev/null
            rm -f "$pid_file"
        fi
        return 1
    fi
}

# Function to start frontend
start_frontend() {
    local name=$1
    local dir=$2
    local port=$3
    local pid_file="$PID_DIR/${name// /-}-frontend.pid"
    local log_file="$LOG_DIR/${name// /-}-frontend.log"
    
    echo -e "${YELLOW}Starting $name Frontend (port $port)...${NC}"
    
    # Check if already running
    if [ -f "$pid_file" ]; then
        local old_pid=$(cat "$pid_file")
        if ps -p $old_pid > /dev/null 2>&1; then
            echo -e "${YELLOW}  Frontend already running (PID: $old_pid)${NC}"
            return 0
        else
            rm -f "$pid_file"
        fi
    fi
    
    # Check port availability
    if ! check_port $port; then
        echo -e "${RED}✗ Port $port already in use${NC}"
        return 1
    fi
    
    # Check if npm dependencies are installed
    if [ ! -d "$dir/node_modules" ]; then
        echo "  Installing npm dependencies..."
        (cd "$dir" && npm install --silent)
    fi
    
    # Clear old log file to avoid confusion
    > "$log_file"
    
    # Start frontend in background
    cd "$dir"
    # Set BROWSER=none to prevent auto-opening browser
    BROWSER=none nohup npm start > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    cd "$BASE_DIR"
    
    # Wait for port to start listening
    if ! wait_for_port $port 60 "$name Frontend"; then
        echo -e "${RED}✗ Frontend port failed to open within 60 seconds${NC}"
        echo "  Check log: $log_file"
        if [ -f "$pid_file" ]; then
            local failed_pid=$(cat "$pid_file")
            kill -9 $failed_pid 2>/dev/null
            rm -f "$pid_file"
        fi
        return 1
    fi
    
    # Wait for React to compile
    if ! wait_for_react_ready "$log_file" 45 "$name Frontend"; then
        echo -e "${RED}✗ Frontend compilation failed${NC}"
        echo "  Check log: $log_file"
        tail -20 "$log_file"
        if [ -f "$pid_file" ]; then
            local failed_pid=$(cat "$pid_file")
            kill -9 $failed_pid 2>/dev/null
            rm -f "$pid_file"
        fi
        return 1
    fi
    
    # Verify frontend is actually serving content
    verify_frontend_ready $port
    
    echo -e "${GREEN}✓ Frontend ready (PID: $pid)${NC}"
    echo "  Log: $log_file"
    return 0
}

# Function to start a complete service (backend + frontend)
start_service() {
    local name=$1
    local dir=$2
    local backend_port=$3
    local frontend_port=$4
    
    echo ""
    echo -e "${BLUE}════════════════════════════════════${NC}"
    echo -e "${BLUE}$name${NC}"
    echo -e "${BLUE}════════════════════════════════════${NC}"
    
    # Check if directory exists
    if [ ! -d "$dir" ]; then
        echo -e "${RED}✗ Directory not found: $dir${NC}"
        return 1
    fi
    
    # Start backend first
    if ! start_backend "$name" "$dir" $backend_port; then
        echo -e "${RED}✗ Failed to start $name backend${NC}"
        return 1
    fi
    
    echo ""
    
    # Then start frontend
    if ! start_frontend "$name" "$dir" $frontend_port; then
        echo -e "${RED}✗ Failed to start $name frontend${NC}"
        return 1
    fi
    
    echo ""
    echo -e "${GREEN}✓ $name fully ready${NC}"
    echo "  Backend:  http://localhost:$backend_port"
    echo "  Frontend: http://localhost:$frontend_port (CSS loaded)"
    
    return 0
}

# Track overall success
OVERALL_SUCCESS=true

# Start Identity Service first (critical)
if ! start_service "Identity Service" "$BASE_DIR/identity-service" 3001 30000; then
    echo -e "${RED}✗ Critical: Identity Service failed to start${NC}"
    OVERALL_SUCCESS=false
fi

# Start Smoke Test (optional)
if [ -d "$BASE_DIR/smoke-test" ]; then
    if ! start_service "Smoke Test" "$BASE_DIR/smoke-test" 30011 30010; then
        echo -e "${YELLOW}⚠ Warning: Smoke Test failed to start${NC}"
    fi
fi

# Start Last Man Standing (optional)
if [ -d "$BASE_DIR/last-man-standing" ]; then
    if ! start_service "Last Man Standing" "$BASE_DIR/last-man-standing" 30021 30020; then
        echo -e "${YELLOW}⚠ Warning: Last Man Standing failed to start${NC}"
    fi
fi

# Start Sweepstakes (optional)
if [ -d "$BASE_DIR/sweepstakes" ]; then
    if ! start_service "Sweepstakes" "$BASE_DIR/sweepstakes" 30031 30030; then
        echo -e "${YELLOW}⚠ Warning: Sweepstakes failed to start${NC}"
    fi
fi

# Start Tic-Tac-Toe (optional)
if [ -d "$BASE_DIR/tic-tac-toe" ]; then
    if ! start_service "Tic-Tac-Toe" "$BASE_DIR/tic-tac-toe" 30041 30040; then
        echo -e "${YELLOW}⚠ Warning: Tic-Tac-Toe failed to start${NC}"
    fi
fi

# Final status
echo ""
echo "========================================"
if [ "$OVERALL_SUCCESS" = true ]; then
    echo -e "${GREEN}✓ All services started successfully!${NC}"
    echo -e "${GREEN}  All frontends compiled and ready to use${NC}"
else
    echo -e "${RED}✗ Some services failed to start${NC}"
    echo -e "${YELLOW}Check the logs in: $LOG_DIR${NC}"
fi
echo ""
echo "Access the system:"
echo -e "  ${BLUE}Identity Service:${NC} http://localhost:30000"
if [ -d "$BASE_DIR/smoke-test" ] && lsof -Pi :30010 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "  ${BLUE}Smoke Test:${NC}       http://localhost:30010"
fi
if [ -d "$BASE_DIR/last-man-standing" ] && lsof -Pi :30020 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "  ${BLUE}Last Man Standing:${NC} http://localhost:30020"
fi
if [ -d "$BASE_DIR/sweepstakes" ] && lsof -Pi :30030 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "  ${BLUE}Sweepstakes:${NC}      http://localhost:30030"
fi
if [ -d "$BASE_DIR/tic-tac-toe" ] && lsof -Pi :30040 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "  ${BLUE}Tic-Tac-Toe:${NC}      http://localhost:30040"
fi
echo ""
echo "Default admin credentials:"
echo "  Email: admin@pubgames.local"
echo "  Code:  123456"
echo ""
echo -e "${YELLOW}Management commands:${NC}"
echo "  Stop all:    ./stop_services.sh"
echo "  View status: ./status_services.sh"
echo "  View logs:   tail -f $LOG_DIR/<service>-backend.log"
echo ""
echo -e "${GREEN}💡 Tip: All CSS and assets are now fully loaded - no refresh needed!${NC}"
echo ""

exit 0
