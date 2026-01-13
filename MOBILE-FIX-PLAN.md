# Fix All Apps for Mobile Access

## Problem
All apps have hardcoded `localhost` URLs which don't work from mobile devices.

## Apps to Fix
1. ✅ identity-service (already fixed)
2. ⏳ last-man-standing
3. ⏳ smoke-test  
4. ⏳ sweepstakes
5. ⏳ template

## Required Changes Per App

### 1. Dynamic API Base
```javascript
// Before
const API_BASE = 'http://localhost:30021/api';

// After
const getApiBase = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:30021/api`;  // Use correct backend port
};
const API_BASE = getApiBase();
```

### 2. Dynamic Identity Service URL
```javascript
// Before
window.location.href = 'http://localhost:30000';
window.location.href = 'http://localhost:30000?logout=true';

// After
const getIdentityUrl = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:30000`;
};
window.location.href = getIdentityUrl();
window.location.href = `${getIdentityUrl()}?logout=true`;
```

### 3. Dynamic Token Validation
```javascript
// Before
const response = await fetch('http://localhost:3001/api/validate-token', {

// After
const getIdentityApiUrl = () => {
  const hostname = window.location.hostname;
  return `http://${hostname}:3001/api`;
};
const response = await fetch(`${getIdentityApiUrl()}/validate-token`, {
```

### 4. Dynamic Login Required Links
```javascript
// Before
<a href="http://localhost:30000">Go to Login</a>

// After
const identityUrl = `http://${window.location.hostname}:30000`;
<a href={identityUrl}>Go to Login</a>
```

## Port Mapping
- Identity Service: Backend 3001, Frontend 30000
- Smoke Test: Backend 30011, Frontend 30010
- Last Man Standing: Backend 30021, Frontend 30020
- Sweepstakes: Backend 30031, Frontend 30030

## Testing After Fix
1. Stop and restart services
2. Desktop: Verify login still works
3. Mobile: Scan QR code
4. Mobile: Login to Identity Service
5. Mobile: Click each app - should work!
6. Mobile: Logout - should redirect correctly

## Automated Fix
Run this script to apply fixes to all apps:

```bash
./fix_mobile_urls.sh
```

Then restart services:
```bash
./stop_services.sh
./start_services.sh
```
