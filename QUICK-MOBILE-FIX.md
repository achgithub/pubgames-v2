# Quick Mobile Fix - Manual Commands

## Option 1: Quick Sed Fix (Recommended)

Run these commands to fix all apps at once:

```bash
cd /home/andrew/pubgames-v2

# Fix Last Man Standing
sed -i "s|const API_BASE = 'http://localhost:30021/api';|const API_BASE = \`http://\${window.location.hostname}:30021/api\`;|g" last-man-standing/src/App.js
sed -i "s|'http://localhost:3001/api/validate-token'|\`http://\${window.location.hostname}:3001/api/validate-token\`|g" last-man-standing/src/App.js
sed -i "s|window.location.href = 'http://localhost:30000';|window.location.href = \`http://\${window.location.hostname}:30000\`;|g" last-man-standing/src/App.js
sed -i "s|window.location.href = 'http://localhost:30000?logout=true';|window.location.href = \`http://\${window.location.hostname}:30000?logout=true\`;|g" last-man-standing/src/App.js
sed -i 's|href="http://localhost:30000"|href={\`http://${window.location.hostname}:30000\`}|g' last-man-standing/src/App.js

# Fix Smoke Test
sed -i "s|const API_BASE = 'http://localhost:30011/api';|const API_BASE = \`http://\${window.location.hostname}:30011/api\`;|g" smoke-test/src/App.js
sed -i "s|'http://localhost:3001/api/validate-token'|\`http://\${window.location.hostname}:3001/api/validate-token\`|g" smoke-test/src/App.js
sed -i "s|window.location.href = IDENTITY_URL;|window.location.href = \`http://\${window.location.hostname}:30000\`;|g" smoke-test/src/App.js
sed -i "s|window.location.href = \`\${IDENTITY_URL}?logout=true\`;|window.location.href = \`http://\${window.location.hostname}:30000?logout=true\`;|g" smoke-test/src/App.js
sed -i 's|href={IDENTITY_URL}|href={\`http://${window.location.hostname}:30000\`}|g' smoke-test/src/App.js

# Fix Sweepstakes
sed -i "s|const API_BASE = 'http://localhost:30031/api';|const API_BASE = \`http://\${window.location.hostname}:30031/api\`;|g" sweepstakes/src/App.js
sed -i "s|'http://localhost:3001/api/validate-token'|\`http://\${window.location.hostname}:3001/api/validate-token\`|g" sweepstakes/src/App.js  
sed -i "s|window.location.href = 'http://localhost:30000';|window.location.href = \`http://\${window.location.hostname}:30000\`;|g" sweepstakes/src/App.js
sed -i "s|window.location.href = 'http://localhost:30000?logout=true';|window.location.href = \`http://\${window.location.hostname}:30000?logout=true\`;|g" sweepstakes/src/App.js
sed -i 's|href="http://localhost:30000"|href={\`http://${window.location.hostname}:30000\`}|g' sweepstakes/src/App.js

echo "✅ All apps fixed! Now restart services:"
echo "./stop_services.sh && ./start_services.sh"
```

## Option 2: Let Claude Fix Them

Just let me know and I'll update each file properly with the correct dynamic URLs.

## What Gets Fixed

For each app, these URLs become dynamic:
1. API_BASE - calls to backend
2. Token validation - SSO check
3. Back to Apps button - navigation
4. Logout redirect - navigation  
5. Login required link - navigation

## After Running

```bash
# Restart to recompile React
./stop_services.sh
./start_services.sh

# Test from desktop (should still work)
# Then test from mobile (should now work!)
```

## Logout Error Fix

The syntax error on logout is because we're calling `window.location.href` which causes React to unmount before cleanup finishes. The fix above will also help, but if it persists, we need to add a small delay:

```javascript
const handleLogout = () => {
  // Clear state first
  setUser(null);
  localStorage.removeItem('user');
  localStorage.removeItem('jwt_token');
  
  // Small delay before redirect to let React cleanup
  setTimeout(() => {
    window.location.href = `http://${window.location.hostname}:30000?logout=true`;
  }, 100);
};
```

Which approach do you prefer?
