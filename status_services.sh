#!/bin/bash

# PubGames V2 - Status Check Script
# Shows the status of all services

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$BASE_DIR/.pids"

echo -e "${BLUE}📊 PubGames V2 - Service Status${NC}"
echo "========================================"
echo ""

# Function to check if port is in use
check_port_status() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${GREEN}✓ RUNNING${NC}"
        return 0
    else
        echo -e "${RED}✗ STOPPED${NC}"
        return 1
    fi
}

# Function to get PID for port
get_port_pid() {
    local port=$1
    lsof -ti:$port 2>/dev/null
}

# Function to check if PID or any of its children are running
is_pid_family_running() {
    local pid=$1
    # Check if the PID itself is running
    if ps -p $pid > /dev/null 2>&1; then
        return 0
    fi
    # Check if any children of this PID are running
    if pgrep -P $pid > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Function to check service status
check_service() {
    local name=$1
    local backend_port=$2
    local frontend_port=$3
    local pid_file_backend="$PID_DIR/${name// /-}-backend.pid"
    local pid_file_frontend="$PID_DIR/${name// /-}-frontend.pid"
    
    echo -e "${BLUE}$name${NC}"
    
    # Check backend
    echo -n "  Backend  (port $backend_port): "
    local backend_running=$(check_port_status $backend_port)
    local backend_pid=$(get_port_pid $backend_port)
    if [ -n "$backend_pid" ]; then
        echo "    PID: $backend_pid"
    fi
    
    # Verify tracked PID is still valid
    if [ -f "$pid_file_backend" ]; then
        local tracked_pid=$(cat "$pid_file_backend")
        if ! is_pid_family_running $tracked_pid; then
            echo -e "    ${YELLOW}⚠ Tracked PID $tracked_pid is no longer running${NC}"
            echo -e "    ${YELLOW}  (Cleaning stale PID file)${NC}"
            rm -f "$pid_file_backend"
        fi
    fi
    
    # Check frontend
    echo -n "  Frontend (port $frontend_port): "
    local frontend_running=$(check_port_status $frontend_port)
    local frontend_pid=$(get_port_pid $frontend_port)
    if [ -n "$frontend_pid" ]; then
        echo "    PID: $frontend_pid"
    fi
    
    # Verify tracked PID is still valid
    if [ -f "$pid_file_frontend" ]; then
        local tracked_pid=$(cat "$pid_file_frontend")
        if ! is_pid_family_running $tracked_pid; then
            echo -e "    ${YELLOW}⚠ Tracked PID $tracked_pid is no longer running${NC}"
            echo -e "    ${YELLOW}  (Cleaning stale PID file)${NC}"
            rm -f "$pid_file_frontend"
        fi
    fi
    
    echo ""
}

# Check Identity Service
check_service "Identity Service" 3001 30000

# Check Smoke Test (only if running)
if lsof -Pi :30010 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :30011 -sTCP:LISTEN -t >/dev/null 2>&1; then
    check_service "Smoke Test" 30011 30010
fi

# Check Last Man Standing (only if running)
if lsof -Pi :30020 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :30021 -sTCP:LISTEN -t >/dev/null 2>&1; then
    check_service "Last Man Standing" 30021 30020
fi

# Check Sweepstakes (only if running)
if lsof -Pi :30030 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :30031 -sTCP:LISTEN -t >/dev/null 2>&1; then
    check_service "Sweepstakes" 30031 30030
fi

# Check Tic-Tac-Toe (only if running)
if lsof -Pi :30040 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :30041 -sTCP:LISTEN -t >/dev/null 2>&1; then
    check_service "Tic-Tac-Toe" 30041 30040
fi

# Check Template (only if running)
if lsof -Pi :30050 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :30051 -sTCP:LISTEN -t >/dev/null 2>&1; then
    check_service "Template" 30051 30050
fi

echo "========================================"
echo ""

# Count running services by checking ports
running_count=0
[ $(lsof -Pi :30000 -sTCP:LISTEN -t 2>/dev/null | wc -l) -gt 0 ] && running_count=$((running_count + 1))
[ $(lsof -Pi :30010 -sTCP:LISTEN -t 2>/dev/null | wc -l) -gt 0 ] && running_count=$((running_count + 1))
[ $(lsof -Pi :30020 -sTCP:LISTEN -t 2>/dev/null | wc -l) -gt 0 ] && running_count=$((running_count + 1))
[ $(lsof -Pi :30030 -sTCP:LISTEN -t 2>/dev/null | wc -l) -gt 0 ] && running_count=$((running_count + 1))
[ $(lsof -Pi :30040 -sTCP:LISTEN -t 2>/dev/null | wc -l) -gt 0 ] && running_count=$((running_count + 1))

echo "Summary:"
echo "  Services running: $running_count"

# Check for truly orphaned processes (no PID file AND not on expected ports)
echo ""
echo "Checking for orphaned processes..."

# Get all pubgames processes, excluding MCP server and this status script
all_go_pids=$(pgrep -f "go run.*pubgames" 2>/dev/null | grep -v "$$")
all_node_pids=$(pgrep -af "node.*pubgames" 2>/dev/null | grep -v "status_services" | grep -v "mcp" | grep -v "claude" | awk '{print $1}')

# Get PIDs that are listening on expected ports
expected_pids=""
for port in 3001 30000 30010 30011 30020 30021 30030 30031 30040 30041 30050 30051; do
    port_pid=$(lsof -ti:$port 2>/dev/null)
    if [ -n "$port_pid" ]; then
        expected_pids="$expected_pids $port_pid"
        # Also get parent PIDs
        for pid in $port_pid; do
            parent_pid=$(ps -o ppid= -p $pid 2>/dev/null | tr -d ' ')
            if [ -n "$parent_pid" ] && [ "$parent_pid" != "1" ]; then
                expected_pids="$expected_pids $parent_pid"
                # Also get grandparent (for npm -> node -> react-scripts chain)
                grandparent_pid=$(ps -o ppid= -p $parent_pid 2>/dev/null | tr -d ' ')
                if [ -n "$grandparent_pid" ] && [ "$grandparent_pid" != "1" ]; then
                    expected_pids="$expected_pids $grandparent_pid"
                fi
            fi
        done
    fi
done

# Count truly orphaned (not in expected PIDs)
orphaned_count=0
orphaned_details=""
for pid in $all_go_pids $all_node_pids; do
    if ! echo " $expected_pids " | grep -q " $pid "; then
        orphaned_count=$((orphaned_count + 1))
        # Get process details
        process_info=$(ps -p $pid -o pid,ppid,cmd --no-headers 2>/dev/null)
        if [ -n "$process_info" ]; then
            orphaned_details="${orphaned_details}  ${process_info}\n"
        fi
    fi
done

if [ $orphaned_count -gt 0 ]; then
    echo -e "${YELLOW}⚠ Found $orphaned_count orphaned process(es):${NC}"
    echo -e "$orphaned_details"
    echo "These processes are not associated with any running service."
    echo "Run './stop_services.sh' to clean up"
else
    echo -e "${GREEN}✓ No orphaned processes${NC}"
fi

echo ""

# Show URLs if services are running
identity_running=$(lsof -Pi :30000 -sTCP:LISTEN -t >/dev/null 2>&1 && echo "yes" || echo "no")

if [ "$identity_running" = "yes" ]; then
    echo "Access the system:"
    echo -e "  ${BLUE}Identity Service:${NC} http://localhost:30000"
    
    if lsof -Pi :30010 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "  ${BLUE}Smoke Test:${NC}       http://localhost:30010"
    fi
    
    if lsof -Pi :30020 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "  ${BLUE}Last Man Standing:${NC} http://localhost:30020"
    fi
    
    if lsof -Pi :30030 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "  ${BLUE}Sweepstakes:${NC}      http://localhost:30030"
    fi
    
    if lsof -Pi :30040 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "  ${BLUE}Tic-Tac-Toe:${NC}      http://localhost:30040"
    fi
    
    echo ""
fi

# Show log locations
if [ -d "$BASE_DIR/logs" ]; then
    echo "Log files: $BASE_DIR/logs/"
    echo "  View logs: tail -f logs/<service>-backend.log"
    echo ""
fi
