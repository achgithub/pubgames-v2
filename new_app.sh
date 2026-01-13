#!/bin/bash

# PubGames V2 - New App Creator
# Creates a new app from the template
#
# ═══════════════════════════════════════════════════════════════════════════
# ⚠️  IMPORTANT FOR AI ASSISTANTS (Claude, ChatGPT, etc.):
# ═══════════════════════════════════════════════════════════════════════════
#
# When creating new apps or modifying the template, ALWAYS ensure CORS is 
# configured to allow BOTH localhost AND the network IP address!
#
# ✅ CORRECT CORS Configuration:
#   handlers.AllowedOrigins([]string{
#       "http://localhost:" + FRONTEND_PORT,
#       "http://192.168.1.45:" + FRONTEND_PORT,  // CRITICAL for mobile!
#   })
#
# ❌ WRONG - Will break mobile access:
#   handlers.AllowedOrigins([]string{
#       "http://localhost:" + FRONTEND_PORT,     // Only localhost = mobile fails!
#   })
#
# WHY THIS MATTERS:
# - Desktop browsers access via http://localhost:30020
# - Mobile devices access via http://192.168.1.45:30020 (network IP)
# - If CORS only allows localhost, mobile gets "Network Error" 
# - This is the #1 cause of "works on desktop, fails on mobile" issues
#
# The template at /home/andrew/pubgames-v2/template/main.go has been updated
# with the correct CORS configuration. Always verify this when creating apps!
# ═══════════════════════════════════════════════════════════════════════════

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$BASE_DIR/template"

echo -e "${BLUE}🎮 PubGames V2 - New App Creator${NC}"
echo "========================================"
echo ""
echo -e "${YELLOW}Input Rules:${NC}"
echo "  • App name: lowercase, numbers, hyphens only (e.g., 'poker-night')"
echo "  • Display name: letters, numbers, spaces, basic punctuation"
echo "  • Avoid: < > \" ' \\ (these break Go syntax)"
echo "  • Icon: Select from list or enter custom"
echo ""

# Check if template exists
if [ ! -d "$TEMPLATE_DIR" ]; then
    echo -e "${RED}✗ Template directory not found: $TEMPLATE_DIR${NC}"
    exit 1
fi

# Get app details
echo ""
read -p "App name (e.g., 'poker-night'): " APP_NAME
if [ -z "$APP_NAME" ]; then
    echo -e "${RED}✗ App name is required${NC}"
    exit 1
fi

# Validate app name - only lowercase letters, numbers, hyphens
if ! [[ "$APP_NAME" =~ ^[a-z0-9-]+$ ]]; then
    echo -e "${RED}✗ Invalid app name. Use only lowercase letters, numbers, and hyphens${NC}"
    echo "  Example: poker-night, todo-app, game-tracker"
    exit 1
fi

read -p "App display name (e.g., 'Poker Night'): " APP_DISPLAY_NAME
if [ -z "$APP_DISPLAY_NAME" ]; then
    APP_DISPLAY_NAME="$APP_NAME"
fi

# Validate display name - no special characters that break Go syntax
if [[ "$APP_DISPLAY_NAME" =~ [\<\>\"\'\\] ]]; then
    echo -e "${RED}✗ Invalid display name. Cannot contain: < > \" ' \\${NC}"
    echo "  These characters break Go code syntax"
    exit 1
fi

read -p "App number (1-99, e.g., 3): " APP_NUMBER
if [ -z "$APP_NUMBER" ] || ! [[ "$APP_NUMBER" =~ ^[0-9]+$ ]] || [ "$APP_NUMBER" -lt 1 ] || [ "$APP_NUMBER" -gt 99 ]; then
    echo -e "${RED}✗ Invalid app number. Must be a number between 1 and 99${NC}"
    exit 1
fi

read -p "App description: " APP_DESCRIPTION
if [ -z "$APP_DESCRIPTION" ]; then
    APP_DESCRIPTION="A PubGames application"
fi

# Validate description - no special characters that break Go/JSON syntax
if [[ "$APP_DESCRIPTION" =~ [\<\>\"\'\\] ]]; then
    echo -e "${RED}✗ Invalid description. Cannot contain: < > \" ' \\${NC}"
    echo "  These characters break Go code syntax"
    exit 1
fi

# Icon selection with suggestions
echo ""
echo -e "${YELLOW}App Icon Selection${NC}"

# Define available emojis
ALL_EMOJIS=(
    "🎮" "🃏" "🎲" "🏆" "🎯" "⚽" "🏀" "⚾" "🎾" "🏐"
    "🎱" "🏓" "🏸" "🏒" "🏑" "🏏" "⛳" "🎣" "🎿" "⛷️"
    "🎰" "🎪" "🎨" "🎭" "🎬" "🎤" "🎧" "🎼" "🎹" "🎺"
    "🎻" "🎸" "🥁" "📱" "💻" "⌨️" "🖥️" "🖨️" "🖱️" "🕹️"
    "📊" "📈" "📉" "📋" "📌" "📍" "📎" "📝" "✏️" "📚"
    "📖" "📰" "🗞️" "📧" "📨" "📩" "📤" "📥" "📦" "📫"
    "🔔" "🔕" "🔊" "🔇" "📢" "📣" "📯" "🔈" "🔉" "🔊"
    "💡" "🔦" "🕯️" "🔥" "💧" "🌊" "⚡" "❄️" "☀️" "🌙"
    "⭐" "✨" "⚡" "💫" "🌟" "☄️" "🔮" "🎁" "🎈" "🎉"
    "🎊" "🎀" "🎗️" "🏅" "🥇" "🥈" "🥉" "🏵️" "🎖️" "🔰"
)

# Get used icons from Identity Service database
USED_ICONS=()
IDENTITY_DB="$BASE_DIR/identity-service/data/identity.db"
if [ -f "$IDENTITY_DB" ]; then
    # Read used icons from database
    while IFS= read -r icon; do
        if [ -n "$icon" ]; then
            USED_ICONS+=("$icon")
        fi
    done < <(sqlite3 "$IDENTITY_DB" "SELECT icon FROM apps WHERE is_active = 1;" 2>/dev/null)
fi

# Filter out used icons
AVAILABLE_EMOJIS=()
for emoji in "${ALL_EMOJIS[@]}"; do
    used=0
    for used_emoji in "${USED_ICONS[@]}"; do
        if [ "$emoji" = "$used_emoji" ]; then
            used=1
            break
        fi
    done
    if [ $used -eq 0 ]; then
        AVAILABLE_EMOJIS+=("$emoji")
    fi
done

# Show icon selection
show_icon_choices() {
    local start=$1
    local count=5
    local end=$((start + count))
    
    if [ $start -ge ${#AVAILABLE_EMOJIS[@]} ]; then
        start=0
        end=$count
    fi
    
    echo ""
    echo "Available icons:"
    for i in $(seq $start $((end - 1))); do
        if [ $i -lt ${#AVAILABLE_EMOJIS[@]} ]; then
            local num=$((i - start + 1))
            echo "  $num) ${AVAILABLE_EMOJIS[$i]}"
        fi
    done
    echo "  6) Show 5 more options"
    echo "  7) Enter custom emoji/text"
}

ICON_START=0
while true; do
    show_icon_choices $ICON_START
    echo ""
    read -p "Select icon (1-7): " ICON_CHOICE
    
    if [[ "$ICON_CHOICE" =~ ^[1-5]$ ]]; then
        # Valid selection from list
        ICON_INDEX=$((ICON_START + ICON_CHOICE - 1))
        if [ $ICON_INDEX -lt ${#AVAILABLE_EMOJIS[@]} ]; then
            APP_ICON="${AVAILABLE_EMOJIS[$ICON_INDEX]}"
            echo -e "${GREEN}✓ Selected: $APP_ICON${NC}"
            break
        else
            echo -e "${RED}✗ Invalid selection${NC}"
        fi
    elif [ "$ICON_CHOICE" = "6" ]; then
        # Show more options
        ICON_START=$((ICON_START + 5))
        if [ $ICON_START -ge ${#AVAILABLE_EMOJIS[@]} ]; then
            ICON_START=0
            echo -e "${YELLOW}Wrapping back to start...${NC}"
        fi
    elif [ "$ICON_CHOICE" = "7" ]; then
        # Custom entry
        echo ""
        read -p "Enter custom icon: " APP_ICON
        if [ -z "$APP_ICON" ]; then
            APP_ICON="📝"
        fi
        # Validate custom icon
        if [[ "$APP_ICON" =~ [\<\>\"\'\\] ]]; then
            echo -e "${RED}✗ Invalid icon. Cannot contain: < > \" ' \\${NC}"
            echo "  Use a simple emoji or symbol"
            exit 1
        fi
        ICON_LENGTH=$(echo -n "$APP_ICON" | wc -m)
        if [ "$ICON_LENGTH" -gt 10 ]; then
            echo -e "${RED}✗ Icon too long. Use a single emoji or 1-2 character symbol${NC}"
            exit 1
        fi
        echo -e "${GREEN}✓ Using: $APP_ICON${NC}"
        break
    else
        echo -e "${RED}✗ Invalid choice. Enter 1-7${NC}"
    fi
done

# Calculate ports
FRONTEND_PORT="300${APP_NUMBER}0"
BACKEND_PORT="300${APP_NUMBER}1"

# Show summary
echo ""
echo -e "${YELLOW}Creating new app with:${NC}"
echo "  Name: $APP_NAME"
echo "  Display Name: $APP_DISPLAY_NAME"
echo "  Description: $APP_DESCRIPTION"
echo "  Icon: $APP_ICON"
echo "  App Number: $APP_NUMBER"
echo "  Frontend Port: $FRONTEND_PORT"
echo "  Backend Port: $BACKEND_PORT"
echo ""
read -p "Continue? (y/n): " CONFIRM

if [ "$CONFIRM" != "y" ]; then
    echo "Cancelled"
    exit 0
fi

APP_DIR="$BASE_DIR/$APP_NAME"

# Check if app already exists
if [ -d "$APP_DIR" ]; then
    echo -e "${RED}✗ App directory already exists: $APP_DIR${NC}"
    exit 1
fi

# Copy template
echo ""
echo -e "${YELLOW}Copying template...${NC}"
cp -r "$TEMPLATE_DIR" "$APP_DIR"
echo -e "${GREEN}✓ Template copied${NC}"

# Replace placeholders in go.mod
echo -e "${YELLOW}Updating go.mod...${NC}"
sed -i "s|module pubgames/template|module pubgames/$APP_NAME|g" "$APP_DIR/go.mod"
echo -e "${GREEN}✓ Updated go.mod${NC}"

# Replace placeholders in main.go
echo -e "${YELLOW}Updating main.go...${NC}"
sed -i "s|30X1|$BACKEND_PORT|g" "$APP_DIR/main.go"
sed -i "s|30X0|$FRONTEND_PORT|g" "$APP_DIR/main.go"
sed -i "s|PLACEHOLDER_APP_NAME|$APP_DISPLAY_NAME|g" "$APP_DIR/main.go"
sed -i "s|PLACEHOLDER_ICON|$APP_ICON|g" "$APP_DIR/main.go"
sed -i "s|./data/app.db|./data/${APP_NAME}.db|g" "$APP_DIR/main.go"
echo -e "${GREEN}✓ Updated main.go${NC}"

# Replace placeholders in database.go
echo -e "${YELLOW}Updating database.go...${NC}"
sed -i "s|app.db|${APP_NAME}.db|g" "$APP_DIR/database.go"
echo -e "${GREEN}✓ Updated database.go${NC}"

# Replace placeholders in package.json
echo -e "${YELLOW}Updating package.json...${NC}"
sed -i "s|pubgames-template|pubgames-$APP_NAME|g" "$APP_DIR/package.json"
sed -i "s|Template app for PubGames ecosystem|$APP_DESCRIPTION|g" "$APP_DIR/package.json"
sed -i "s|PORT=30X0|PORT=$FRONTEND_PORT|g" "$APP_DIR/package.json"
echo -e "${GREEN}✓ Updated package.json${NC}"

# Replace placeholders in src/App.js
echo -e "${YELLOW}Updating src/App.js...${NC}"
sed -i "s|30X1|$BACKEND_PORT|g" "$APP_DIR/src/App.js"
sed -i "s|30X0|$FRONTEND_PORT|g" "$APP_DIR/src/App.js"
sed -i "s|PLACEHOLDER_ICON|$APP_ICON|g" "$APP_DIR/src/App.js"
sed -i "s|PLACEHOLDER_APP_NAME|$APP_DISPLAY_NAME|g" "$APP_DIR/src/App.js"
echo -e "${GREEN}✓ Updated src/App.js${NC}"

# Update README.md
echo -e "${YELLOW}Updating README.md...${NC}"
sed -i "s|Template App|$APP_DISPLAY_NAME|g" "$APP_DIR/README.md"
sed -i "s|30X0|$FRONTEND_PORT|g" "$APP_DIR/README.md"
sed -i "s|30X1|$BACKEND_PORT|g" "$APP_DIR/README.md"
echo -e "${GREEN}✓ Updated README.md${NC}"

# Update public/index.html
echo -e "${YELLOW}Updating public/index.html...${NC}"
sed -i "s|Template App|$APP_DISPLAY_NAME|g" "$APP_DIR/public/index.html"
echo -e "${GREEN}✓ Updated public/index.html${NC}"

echo ""
echo "========================================"
echo -e "${GREEN}✓ New app created successfully!${NC}"
echo ""
echo "App location: $APP_DIR"
echo "Frontend: http://localhost:$FRONTEND_PORT"
echo "Backend:  http://localhost:$BACKEND_PORT"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. cd $APP_DIR"
echo "2. go mod download"
echo "3. npm install"
echo "4. Start backend: go run *.go"
echo "5. Start frontend: npm start"
echo ""
echo "Or add this app to start_services.sh and run ./start_services.sh"
echo ""
echo -e "${YELLOW}Don't forget to:${NC}"
echo "- Add app to Identity Service database (apps table)"
echo "- Customize database schema in database.go"
echo "- Implement your business logic in handlers.go"
echo ""