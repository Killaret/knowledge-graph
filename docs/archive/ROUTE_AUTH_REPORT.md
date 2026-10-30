# Route Protection and Authentication Logic Report

**Date:** 2026-07-04
**Frontend:** SvelteKit 5 with Svelte 5 Runes
**Auth Store:** `auth.svelte.ts`

---

## 1. Central Route Protection (+layout.svelte)

### Location
`frontend/src/routes/+layout.svelte`

### Public Routes (No Authentication Required)
```typescript
const publicRoutes = [
  '/',           // Main page - accessible for guests ✅
  '/graph',      // Graph page - accessible for guests ✅
  '/auth/login',
  '/auth/register',
  auth/forgot-password',
  '/auth/reset-password',
  '/auth/yandex/callback',
  '/health',
  '/test'        // Test routes for visual regression testing
];
```

### Protection Logic
```typescript
$effect(() => {
  const currentPath = $page.url.pathname;
  const isPublicRoute = publicRoutes.some(route => currentPath.startsWith(route));

  if (
    isInitialized() &&
    !isLoading() &&
    !isPublicRoute &&
    !isAuthenticated() &&
    !isSkipAuth
  ) {
    const returnUrl = encodeURIComponent(currentPath);
    goto(`/auth/login?redirect=${returnUrl}`);
  }
});
```

**Conditions for redirect to login:**
1. ✅ Auth is initialized (`isInitialized()`)
2. ✅ Not loading (`!isLoading()`)
3. ✅ Not a public route (`!isPublicRoute`)
4. ✅ Not authenticated (`!isAuthenticated()`)
5. ✅ Not in SKIP_AUTH mode (`!isSkipAuth`)

**Result:** Unauthenticated users are redirected to `/auth/login` with return URL.

---

## 2. Page-Level Protection

### Profile Page (`/profile`)
**Location:** `frontend/src/routes/profile/+page.svelte`

```typescript
$effect(() => {
  if (!isAuthenticated()) {
    goto('/auth/login?redirect=/profile');
  }
});
```

**Protection:** Redirects to login if not authenticated.

---

### Notes Detail Page (`/notes/[id]`)
**Location:** `frontend/src/routes/notes/[id]/+page.svelte`

**Protection:** None - accessible to guests (public notes).

**Behavior:**
- 404 error → redirects to `/` after 3 seconds
- No auth check

---

### Search Page (`/search`)
**Location:** `frontend/src/routes/search/+page.svelte`

**Protection:** None - accessible to guests.

**Behavior:**
- Empty query → redirects to `/`
- No auth check

---

### Graph Detail Page (`/graph/[id]`)
**Location:** `frontend/src/routes/graph/[id]/+page.svelte`

**Protection:** None - accessible to guests.

**Behavior:**
- No auth check
- Loads graph data via API

---

### Graph 3D Pages (`/graph/3d/*`)
**Location:** `frontend/src/routes/graph/3d/+page.svelte`, `/graph/3d/[id]/+page.svelte`

**Protection:** None - accessible to guests.

**Behavior:**
- No auth check
- Loads 3D graph data

---

## 3. Auth Pages (Self-Protected)

### Login Page (`/auth/login`)
**Location:** `frontend/src/routes/auth/login/+page.svelte`

```typescript
$effect(() => {
  if (isAuthenticated()) {
    const redirectTo = $page.url.searchParams.get('redirect') || '/';
    goto(redirectTo);
  }
});
```

**Protection:** Redirects authenticated users away (to `redirect` param or `/`).

---

### Register Page (`/auth/register`)
**Location:** `frontend/src/routes/auth/register/+page.svelte`

```typescript
$effect(() => {
  if (isAuthenticated()) {
    goto('/');
  }
});
```

**Protection:** Redirects authenticated users to `/`.

---

### Forgot Password (`/auth/forgot-password`)
**Location:** `frontend/src/routes/auth/forgot-password/+page.svelte`

```typescript
$effect(() => {
  if (isAuthenticated()) {
    goto('/');
  }
});
```

**Protection:** Redirects authenticated users to `/`.

---

### Reset Password (`/auth/reset-password`)
**Location:** `frontend/src/routes/auth/reset-password/+page.svelte`

```typescript
$effect(() => {
  if (isAuthenticated()) {
    goto('/');
  }
});
```

**Protection:** Redirects authenticated users to `/`.

---

## 4. Authentication Store Logic

### Location
`frontend/src/shared/stores/auth.svelte.ts`

### Auth State
```typescript
const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  refreshToken: null as string | null,
  isInitialized: false,
  isLoading: false,
  error: null as string | null,
  apiKey: null as string | null
});
```

### Initialization Flow
```typescript
export async function initAuth(): Promise<void> {
  // 1. Check for SKIP_AUTH mode (dev only)
  if (import.meta.env.DEV) {
    const url = new URL(window.location.href);
    if (url.searchParams.get('skip_auth') === 'true') {
      localStorage.setItem('__SKIP_AUTH__', 'true');
    }
  }

  // 2. Load tokens from localStorage
  const storedAccessToken = localStorage.getItem('access_token');
  const storedRefreshToken = localStorage.getItem('refresh_token');
  const storedApiKey = localStorage.getItem('api_key');

  // 3. If no refresh token, no user
  if (!storedRefreshToken) {
    authState.currentUser = null;
    return;
  }

  // 4. Try to refresh token and get user info
  const refreshed = await refreshAccessToken();
  if (refreshed) {
    const user = await usersApi.getMe();
    authState.currentUser = user;
    void preloadAuthenticatedGraph();
  } else {
    clearAuthState();
  }

  authState.isInitialized = true;
}
```

### isAuthenticated() Function
```typescript
export function isAuthenticated(): boolean {
  // Dev mode: check SKIP_AUTH flags
  if (browser && import.meta.env.DEV) {
    if ((window as any).__SKIP_AUTH__ === true) return true;
    if (localStorage.getItem('__SKIP_AUTH__') === 'true') return true;
    const url = new URL(window.location.href);
    if (url.searchParams.get('skip_auth') === 'true') {
      localStorage.setItem('__SKIP_AUTH__', 'true');
      return true;
    }
  }
  // Production: check tokens
  return !!authState.accessToken || !!authState.apiKey;
}
```

**Auth Methods:**
- `login(login, password)` → saves tokens, gets user, preloads graph
- `register(login, password, email)` → same as login
- `logout()` → clears tokens, clears preload cache, redirects to `/auth/login`
- `handleYandexCallback(code, state)` → OAuth flow
- `loginWithApiKey(key)` → API key auth

---

## 5. SKIP_AUTH Mode (Dev Only)

### Activation Methods
1. **Query parameter:** `?skip_auth=true` (persists to localStorage)
2. **LocalStorage:** `localStorage.setItem('__SKIP_AUTH__', 'true')`
3. **Window flag:** `(window as any).__SKIP_AUTH__ = true` (Playwright)

### Usage
```typescript
export function skipAuthMode(): boolean {
  if (!browser) return false;
  if (!import.meta.env.DEV) return false;
  return localStorage.getItem('__SKIP_AUTH__') === 'true';
}
```

**Effect:** Bypasses all auth checks in production (dev only).

---

## 6. Redirect Parameters

### Login Redirect
```typescript
// In +layout.svelte
const returnUrl = encodeURIComponent(currentPath);
goto(`/auth/login?redirect=${returnUrl}`);
```

### Usage in Login Page
```typescript
// In auth/login/+page.svelte
const redirectTo = $page.url.searchParams.get('redirect') || '/';
goto(redirectTo);
```

**Example:**
- User visits `/profile` (not authenticated)
- Redirected to `/auth/login?redirect=%2Fprofile`
- After login, redirected back to `/profile`

---

## 7. Preload Service

### Guest Preload
```typescript
// In +layout.svelte
if (isInitialized() && !isAuthenticated()) {
  startPreload();
}
```

**Purpose:** Preloads public graph data for guests to improve UX.

### Authenticated Preload
```typescript
// In auth.svelte.ts (after login/register)
void preloadAuthenticatedGraph();
```

**Purpose:** Preloads user's personal graph data after auth.

---

## 8. Summary Table

| Route | Auth Required | Protection Level | Redirect Target | Notes |
|-------|--------------|-----------------|----------------|-------|
| `/` | ❌ | Layout | `/auth/login` | ✅ Public (guests) |
| `/graph` | ❌ | Layout | `/auth/login` | ✅ Public (guests) |
| `/profile` | ✅ | Page | `/auth/login?redirect=/profile` | Protected |
| `/notes/[id]` | ❌ | None | - | Public notes |
| `/search` | ❌ | None | `/` (if empty query) | Public search |
| `/graph/[id]` | ❌ | None | - | Public graph |
| `/graph/3d/*` | ❌ | None | - | Public 3D graph |
| `/auth/login` | ❌ | Page | `/` or `redirect` | Redirects if authenticated |
| `/auth/register` | ❌ | Page | `/` | Redirects if authenticated |
| `/auth/forgot-password` | ❌ | Page | `/` | Redirects if authenticated |
| `/auth/reset-password` | ❌ | Page | `/` | Redirects if authenticated |
| `/auth/yandex/callback` | ❌ | Page | - | OAuth callback |
| `/health` | ❌ | Layout | `/auth/login` | Health check endpoint |
| `/test/*` | ❌ | Layout | `/auth/login` | Test routes |

---

## 9. Current State (After Fixes)

### Recent Changes
1. **Added `'/'` and `'/graph'` to `publicRoutes`** in `+layout.svelte`
   - **Before:** Guests were redirected to login
   - **After:** Guests can access main page and graph without auth

### Test Data Available
- 5 test notes created via API
- 5 test links created via API (different types/weights)
- `source_type` field added to links schema

### Services Status
- Backend: ✅ (port 8085)
- Graph Service: ✅ (port 8092)
- Frontend: ✅ (port 3001)
- PostgreSQL: ✅
- Redis: ✅

---

## 10. Recommendations

### Current Issues
1. **No protection on `/notes/[id]`** - guests can access any note by ID
2. **No protection on `/graph/[id]`** - guests can access any graph by ID
3. **No protection on `/search`** - guests can search all notes
4. **No protection on `/graph/3d/*`** - guests can access 3D graphs

### Suggested Improvements
1. Add protection to `/notes/[id]` if note is private
2. Add protection to `/graph/[id]` if graph is private
3. Add protection to `/search` if you want to restrict search
4. Consider adding `/notes` list page protection

### SKIP_AUTH Mode
- **Status:** Active in dev mode
- **Usage:** Testing without auth
- **Recommendation:** Keep for dev, disable in production (already disabled)
