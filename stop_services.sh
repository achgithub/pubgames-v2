#!/bin/bash

# PubGames V2 - Stop Services Script
# Stops all running services using PID files and port cleanup

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$BASE_DIR/.pids"

echo -e "${BLUE}🛑 PubGames V2 - Stopping Services${NC}"
echo "========================================"

# Function to stop a process by PID file
stop_by_pidfile() {
    local pid_file=$1
    local service_name=$2
    
    if [ ! -f "$pid_file" ]; then
        return 0
    fi
    
    local pid=$(cat "$pid_file")
    
    if ps -p $pid > /dev/null 2>&1; then
        echo -e "${YELLOW}Stopping $service_name (PID: $pid)...${NC}"
        kill $pid 2>/dev/null
        
        # Wait up to 5 seconds for graceful shutdown
        local count=0
        while ps -p $pid > /dev/null 2>&1 && [ $count -lt 5 ]; do
            sleep 1
            count=$((count + 1))
        done
        
        # Force kill if still running
        if ps -p $pid > /dev/null 2>&1; then
            echo "  Force killing..."
            kill -9 $pid 2>/dev/null
        fi
        
        echo -e "${GREEN}✓ Stopped $service_name${NC}"
    fi
    
    rm -f "$pid_file"
}

# Function to kill processes on a port (fallback)
kill_port() {
    local port=$1
    local name=$2
    
    local pids=$(lsof -ti:$port 2>/dev/null)
    
    if [ -n "$pids" ]; then
        echo -e "${YELLOW}Cleaning up port $port ($name)...${NC}"
        echo "  Killing PIDs: $pids"
        kill -9 $pids 2>/dev/null
        echo -e "${GREEN}✓ Port $port cleared${NC}"
    fi
}

echo ""

# Stop services using PID files first
if [ -d "$PID_DIR" ]; then
    echo "Stopping services by PID files..."
    echo ""
    
    for pid_file in "$PID_DIR"/*.pid; do
        if [ -f "$pid_file" ]; then
            service_name=$(basename "$pid_file" .pid | sed 's/-/ /g')
            stop_by_pidfile "$pid_file" "$service_name"
        fi
    done
    
    echo ""
fi

# Fallback: Clean up known ports
echo "Cleaning up known ports..."
echo ""

# Identity Service
kill_port 3001 "Identity Backend"
kill_port 30000 "Identity Frontend"

# Smoke Test
kill_port 30011 "Smoke Test Backend"
kill_port 30010 "Smoke Test Frontend"

# Last Man Standing
kill_port 30021 "Last Man Standing Backend"
kill_port 30020 "Last Man Standing Frontend"

# Sweepstakes
kill_port 30031 "Sweepstakes Backend"
kill_port 30030 "Sweepstakes Frontend"

# Tic-Tac-Toe
kill_port 30041 "Tic-Tac-Toe Backend"
kill_port 30040 "Tic-Tac-Toe Frontend"

# Template (if running)
kill_port 30051 "Template Backend"
kill_port 30050 "Template Frontend"

echo ""
echo "Cleaning up any remaining processes..."

# Kill any remaining go processes from pubgames
pkill -f "go run.*pubgames" 2>/dev/null && echo "  ✓ Killed remaining Go processes"

# Kill any remaining npm/react processes from pubgames directory
for dir in identity-service smoke-test last-man-standing sweepstakes tic-tac-toe template; do
    if [ -d "$BASE_DIR/$dir" ]; then
        pkill -f "node.*$dir" 2>/dev/null
    fi
done
echo "  ✓ Killed remaining Node processes"

# Clean up PID directory
if [ -d "$PID_DIR" ]; then
    rm -f "$PID_DIR"/*.pid
    echo "  ✓ Cleaned up PID files"
fi

echo ""
echo "========================================"
echo -e "${GREEN}✓ All services stopped${NC}"
echo ""
