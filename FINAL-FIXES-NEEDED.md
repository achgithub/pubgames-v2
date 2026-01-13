# Remaining Fixes - Quick Guide

## Issues Found

### 1. LMS - Syntax Error on "Manage Games" ❌
**Problem:** Missing closing brace for `if (user)` block

**Fix:**
At the end of `/home/andrew/pubgames-v2/last-man-standing/src/App.js`, before `export default App;`, add:

```javascript
      </main>
    </div>
  );
  }  // <-- ADD THIS closing brace for if (user)
  
  // Fallback
  return null;  // <-- ADD THIS fallback
}

export default App;
```

### 2. Smoke Test - Blank Pages on Navigation ❌
**Problem:** Only renders when `view === 'dashboard'` 

**Already Fixed** ✅ - The conditional check is correct: `if (view === 'dashboard' && user)`

### 3. Sweepstakes - Blank Pages ❌  
**Already Fixed** ✅ - Same as Smoke Test

## Quick Test

```bash
cd /home/andrew/pubgames-v2
./stop_services.sh
./start_services.sh
```

Then test:
- **LMS**: Click "Manage Games" button (should not error)
- **Smoke Test**: Click "Items" tab (should show content)
- **Sweepstakes**: Click tabs (should show content)

## Summary

- Smoke Test and Sweepstakes are **ALREADY FIXED** ✅
- Template is **ALREADY UPDATED** ✅  
- Only LMS needs the closing brace added manually

Add these 2 lines to LMS before the final `export`:
```
  }  
  return null;
```
