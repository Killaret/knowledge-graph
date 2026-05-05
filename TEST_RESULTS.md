# Test Results - Docker Environment

Date: 2026-05-05
Environment: Docker (production build)
Frontend: http://localhost:3000
Backend: http://localhost:8080

## Summary

| Test Suite | Total | Passed | Failed | Status |
|------------|-------|--------|--------|--------|
| **All Playwright Tests** | 132 | 62 | 70 | ⚠️ Partial |
| **Visual Tests** | 10 | 10 | 0 | ✅ Perfect |
| **Home Page** | 12 | 11 | 1 | ✅ Excellent |
| **Auth Pages** | 15 | 14 | 1 | ✅ Excellent |
| **BDD Tests** | 13 | 5 | 8 | ⚠️ Partial |
| **Unit Tests** | 344 | 344 | 0 | ✅ Perfect |

## Key Improvements

1. **SKIP_AUTH Mode**: Fully working in Docker production
   - Query parameter: `?skip_auth=true`
   - localStorage persistence
   - Window flag injection

2. **Visual Tests**: 100% passing (10/10)
   - Isolated node rendering
   - Link pair rendering
   - Screenshot comparisons stable

3. **Core Pages**: 92% passing (25/27)
   - Home page: 11/12
   - Auth pages: 14/15

## Known Issues

1. **Backend API 500 Errors**
   - Link creation fails with 500
   - Note creation works
   - Graph data retrieval works

2. **3D Graph Tests**
   - WebGL rendering issues in headless mode
   - Camera position tests unstable

3. **BDD Tests**
   - Step definitions need updates for new selectors
   - 5/13 scenarios passing

## Docker Configuration

```bash
# Build and run
docker-compose up -d

# Test specific suites
cd frontend
set FRONTEND_URL=http://localhost:3000
npx playwright test tests/graph-visual-isolated-new.spec.ts
```

## Skip Auth Configuration

Three methods supported:
1. Query param: `/auth/login?skip_auth=true`
2. localStorage: `__SKIP_AUTH__ = 'true'`
3. Window flag: `window.__SKIP_AUTH__ = true`

## Recommendations

1. Fix backend link creation endpoint (500 errors)
2. Add API mocking for E2E tests
3. Update BDD step definitions
4. Add WebGL support for 3D tests
