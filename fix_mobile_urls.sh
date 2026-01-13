#!/bin/bash

# Fix Mobile URLs in All Apps
# This script updates hardcoded localhost URLs to use dynamic hostnames

set -e

BASE_DIR="/home/andrew/pubgames-v2"
cd "$BASE_DIR"

echo "🔧 Fixing mobile URLs in all apps..."
echo ""

# Function to update an app
fix_app() {
    local app_name=$1
    local backend_port=$2
    local frontend_port=$3
    local app_dir="$BASE_DIR/$app_name"
    local app_file="$app_dir/src/App.js"
    
    if [ ! -f "$app_file" ]; then
        echo "⚠️  Skipping $app_name - file not found"
        return
    fi
    
    echo "📝 Fixing $app_name..."
    
    # Create backup
    cp "$app_file" "$app_file.backup"
    
    # Create temp file with fixes
    cat > /tmp/fix_${app_name}.js << 'HEREDOC'
// Dynamic URL helpers
const getHostname = () => window.location.hostname;
const getApiBase = () => `http://${getHostname()}:BACKEND_PORT/api`;
const getIdentityUrl = () => `http://${getHostname()}:30000`;
const getIdentityApiUrl = () => `http://${getHostname()}:3001/api`;
HEREDOC
    
    # Replace BACKEND_PORT placeholder
    sed -i "s/BACKEND_PORT/$backend_port/g" /tmp/fix_${app_name}.js
    
    # Now update the actual file
    # 1. Replace hardcoded API_BASE
    sed -i "s|const API_BASE = 'http://localhost:$backend_port/api';|const getApiBase = () => \`http://\${window.location.hostname}:$backend_port/api\`;\nconst API_BASE = getApiBase();|g" "$app_file"
    
    # 2. Replace token validation URL
    sed -i "s|'http://localhost:3001/api/validate-token'|\`http://\${window.location.hostname}:3001/api/validate-token\`|g" "$app_file"
    
    # 3. Replace logout redirect
    sed -i "s|'http://localhost:30000?logout=true'|\`http://\${window.location.hostname}:30000?logout=true\`|g" "$app_file"
    
    # 4. Replace back to apps
    sed -i "s|'http://localhost:30000'|\`http://\${window.location.hostname}:30000\`|g" "$app_file"
    
    # 5. Replace login required link (JSX)
    sed -i 's|href="http://localhost:30000"|href={`http://${window.location.hostname}:30000`}|g' "$app_file"
    
    echo "   ✅ Fixed $app_name"
    echo "   📦 Backup saved: $app_file.backup"
    echo ""
}

# Fix each app
fix_app "smoke-test" "30011" "30010"
fix_app "last-man-standing" "30021" "30020"
fix_app "sweepstakes" "30031" "30030"
fix_app "template" "30041" "30040"

echo ""
echo "================================"
echo "✅ All apps fixed!"
echo ""
echo "Next steps:"
echo "1. Review the changes:"
echo "   diff smoke-test/src/App.js smoke-test/src/App.js.backup"
echo ""
echo "2. Restart services:"
echo "   ./stop_services.sh && ./start_services.sh"
echo ""
echo "3. Test from mobile:"
echo "   - Scan QR code"
echo "   - Login"
echo "   - Try each app"
echo "   - Test logout"
echo ""
echo "If something breaks, restore from backup:"
echo "   mv smoke-test/src/App.js.backup smoke-test/src/App.js"
echo "================================"
