# CSS Segregation Guide - PubGames V2

**Last Updated**: January 20, 2026
**Purpose**: Define clear boundaries between shared and app-specific CSS

---

## Core Principle

**Shared CSS** = Used by 2+ apps OR fundamental platform patterns
**App-Specific CSS** = Unique to one app's domain/features

---

## SHARED CSS (`pubgames.css`)

### What MUST Be Shared

#### 1. **Platform Foundation**
- CSS reset/normalize
- Box-sizing, font-family
- Root CSS variables (--game-color, --game-accent)
- Body and html base styles

#### 2. **Layout Structure**
- `.App` - Main app container
- `.container` - Content wrapper
- `header` - Top navigation bar
- `nav.tabs` - Tab navigation
- `main` - Main content area
- Flexbox/grid base utilities

#### 3. **SSO & Navigation**
- `.user-info` - User info display
- `.back-to-apps-btn` - Return to directory
- `.admin-badge` - Admin indicator
- Login/logout buttons in header

#### 4. **Core Components**
- **Buttons**: `.button`, `.button-primary`, `.button-secondary`, `.button-danger`, `.cta-button`, `.action-button`
- **Badges**: `.badge` (base), status badges (active/eliminated/pending/completed)
- **Forms**: `input`, `select`, `textarea`, `.admin-form`, `.inline-form`
- **Tables**: `table`, `th`, `td` base styles
- **Alerts**: `.warning`, `.warning-box`, `.status-message`

#### 5. **Common Patterns**
- `.dashboard` - Dashboard layout
- `.card` - Generic card base
- `.info-text` - Informational text
- `.no-data` / `.no-entries` - Empty state messages
- Loading states, error states

#### 6. **Typography**
- Heading sizes (h1-h6)
- Paragraph spacing
- Link styles
- `.stage`, `.date-range` - Common text patterns

#### 7. **Responsive Base**
- Mobile breakpoints (@media queries)
- Touch-friendly sizing
- Mobile navigation adjustments

---

## APP-SPECIFIC CSS (e.g., `sweepstakes.css`, `lms.css`)

### What SHOULD Be App-Specific

#### 1. **Domain-Specific Layouts**
Examples:
- `.blind-boxes-grid` - Sweepstakes blind box layout
- `.spinner-modal` - Sweepstakes random selection
- `.match-grid` - LMS match display
- `.countdown-card` - LMS round countdown

**Rule**: If the component represents app-specific business logic/workflow

#### 2. **Feature-Specific Components**
Examples:
- `.competition-card` - Sweepstakes competitions
- `.entry-item` - Sweepstakes entries
- `.prediction-card` - LMS predictions
- `.standings-table` - LMS leaderboard

**Rule**: If terminology/structure is unique to that app's domain

#### 3. **App-Unique Interactions**
Examples:
- `.blind-box` selection interface
- `.spinner-wheel` animation
- `.team-selector` dropdowns
- `.round-timer` displays

**Rule**: If the interaction pattern doesn't apply to other apps

#### 4. **Domain-Specific Badges**
Examples:
- `.seed-badge` - Tournament seeding (Sweepstakes)
- `.number-badge` - Race numbers (Sweepstakes)
- `.round-badge` - Round indicators (LMS)

**Rule**: If the badge type is specific to one app's data model

#### 5. **App-Specific Responsive**
Examples:
- Spinner modal mobile sizing
- Competition grid mobile columns
- Match grid tablet breakpoints

**Rule**: If responsive behavior is unique to app layout

---

## GRAY AREAS - Decision Framework

When uncertain, ask these questions:

### Question 1: "Will another app need this exact pattern?"
- **Yes** → Shared CSS
- **Maybe** → Shared CSS (reusable)
- **No** → App-specific CSS

### Question 2: "Does this represent app business logic?"
- **Yes** → App-specific CSS
- **No** → Shared CSS

### Question 3: "Could this be generalized?"
- **Easily** → Generalize and put in Shared
- **With effort** → Shared if worth it, otherwise app-specific
- **No, very specific** → App-specific

### Question 4: "Is this styling or structure?"
- **Structure** (layout, grid, flex) → More likely shared
- **Styling** (colors, specific sizes) → Can be either
- **Business logic UI** (blind boxes, match cards) → App-specific

---

## EXAMPLES

### ✅ GOOD: Properly Segregated

**Shared CSS**:
```css
/* Generic card base - used by all apps */
.card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

/* Generic action button - used everywhere */
.action-button {
  padding: 10px 20px;
  background: var(--game-color);
  color: white;
  border: none;
  border-radius: 4px;
}
```

**App-Specific CSS** (sweepstakes.css):
```css
/* Sweepstakes blind box selection grid */
.blind-boxes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 15px;
}

/* Sweepstakes competition card extends base .card */
.competition-card {
  /* Specific layout for competition display */
  display: flex;
  flex-direction: column;
  gap: 15px;
}
```

### ❌ BAD: Improperly Segregated

**Shared CSS** (too specific):
```css
/* DON'T: This is too specific to Sweepstakes */
.blind-box-number-display-for-racing-competitions {
  /* ... */
}

/* DON'T: This should be in lms.css */
.last-man-standing-round-countdown-timer {
  /* ... */
}
```

**App CSS** (should be shared):
```css
/* DON'T: Generic card should be shared */
.my-custom-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
}

/* DON'T: Standard button should use shared */
.my-submit-button {
  padding: 10px 20px;
  background: blue;
  color: white;
}
```

---

## MIGRATION STRATEGY

### Phase 1: Conservative Cleanup (Current)
1. Consolidate duplicates in shared CSS
2. Create base classes with variants
3. Document what stays shared
4. **No file splitting yet**

### Phase 2: Identify App-Specific Classes
1. Review each app's unique classes
2. Mark classes for extraction
3. Create segregation plan
4. Document dependencies

### Phase 3: Split CSS Files
1. Create app-specific CSS files
2. Move identified classes
3. Update `index.html` to load both files
4. Test all apps thoroughly
5. Remove extracted classes from shared CSS

### Phase 4: Template & Documentation
1. Create app template with proper CSS structure
2. Update ARCHITECTURE.md
3. Create CSS checklist for new apps
4. Document the segregation principles

---

## LOADING STRATEGY (Future)

Each app's `public/index.html` will load:

```html
<!-- Shared platform CSS (always loaded) -->
<script>
  var hostname = window.location.hostname;
  var sharedCSS = 'http://' + hostname + ':3001/static/pubgames.css';
  var link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = sharedCSS;
  document.head.appendChild(link);
</script>

<!-- App-specific CSS (conditionally loaded) -->
<script>
  var appCSS = 'http://' + hostname + ':3001/static/apps/sweepstakes.css';
  var appLink = document.createElement('link');
  appLink.rel = 'stylesheet';
  appLink.href = appCSS;
  document.head.appendChild(appLink);
</script>
```

**CSS File Structure**:
```
identity-service/static/
├── pubgames.css          # Shared platform CSS
└── apps/
    ├── sweepstakes.css   # Sweepstakes-specific
    ├── lms.css           # Last Man Standing-specific
    ├── tictactoe.css     # Tic-Tac-Toe-specific
    └── template.css      # Template for new apps
```

---

## MAINTENANCE RULES

### When Creating New Apps:
1. Start with shared CSS only
2. Only create app CSS when you need app-specific patterns
3. Always check if pattern can be generalized
4. Document why each app-specific class exists

### When Modifying Shared CSS:
1. Consider impact on ALL apps
2. Test changes across platform
3. Don't add app-specific classes to shared CSS
4. Maintain backwards compatibility

### When Adding to App CSS:
1. First check if shared CSS has suitable class
2. Document the business reason for app-specific class
3. Consider if pattern will be reused later
4. Keep app CSS focused on domain logic

---

## DECISION CHECKLIST

Before adding a new CSS class, ask:

- [ ] Does this represent platform-wide structure? → Shared
- [ ] Is this used/needed by 2+ apps? → Shared
- [ ] Does this represent app-specific business logic? → App-specific
- [ ] Could this be generalized to a reusable pattern? → Shared (generalized)
- [ ] Is this a one-off styling need? → App-specific
- [ ] Does this extend/customize a shared pattern? → App-specific
- [ ] Will this change frequently with app features? → App-specific

---

## BENEFITS OF THIS APPROACH

### For Shared CSS:
- ✅ Smaller file size (only platform essentials)
- ✅ Faster loading for all apps
- ✅ Easier to maintain consistency
- ✅ Clear "contract" of what's available
- ✅ Less cognitive load when debugging

### For App-Specific CSS:
- ✅ Freedom to iterate on app features
- ✅ No fear of breaking other apps
- ✅ Clear ownership and responsibility
- ✅ Easier to understand app-specific styling
- ✅ Can be deleted when app is removed

### For Developers:
- ✅ Clear decision framework
- ✅ Faster development (know where to look)
- ✅ Reduced merge conflicts
- ✅ Better code organization
- ✅ Easier onboarding for new apps

---

**End of Guide**

*This guide will evolve as we learn from implementing the strategy across the platform.*
