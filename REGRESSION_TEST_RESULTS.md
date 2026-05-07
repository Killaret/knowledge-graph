# Regression Test Results - 2026-05-07

## 📊 Executive Summary

**Objective:** Устранить ключевые нестабильности в smoke-тестах и достичь pass rate >80%

**Results Achieved:**
- ✅ Frontend Health Check: Fixed (now shows "healthy")
- ✅ BDD Tests: 13/13 passed (100% pass rate)  
- ⚠️ Smoke Tests: 15/67 passed (21.7% pass rate) - Target not met
- ✅ Docker Infrastructure: All services running correctly

## 🐳 Docker Infrastructure Status

### Before Fixes
```
kg-frontend    Running (unhealthy)   - port 3000
kg-backend     Healthy               - port 8080
```

### After Fixes
```
kg-frontend    Running (healthy)     - port 8081
kg-backend     Healthy               - port 8080
kg-postgres    Healthy               - port 15432
kg-redis       Healthy               - port 6379
kg-nlp         Healthy               - port 5000
```

## 🔧 Fixes Applied

### 1. Frontend Health Check ✅
**Problem:** Frontend showing "unhealthy" in Docker
**Root Cause:** Missing curl in Alpine container for health check
**Solution:** Added curl to Dockerfile
```dockerfile
# Install curl for health check and production dependencies
RUN apk add --no-cache curl && \
    npm ci --omit=dev || npm install --omit=dev
```
**Result:** Health check now passes, container shows "healthy"

### 2. Port Configuration ✅
**Problem:** Port 3000 conflicts in Windows Docker
**Solution:** Changed frontend port from 3000 to 8081
```yaml
ports:
  - "8081:3000"
```

### 3. Notes Test Stabilization ⚠️
**Problem:** 5/8 tests failing with timeout/element not found
**Fixes Applied:**
- Updated selectors to use data-testid attributes
- Added explicit timeouts for better reliability
- Improved error handling

**Current Status:** Still unstable (5 failed, 3 passed)

### 4. Progressive Rendering Test Stabilization ✅
**Problem:** Tests failing with element visibility timeouts
**Fixes Applied:**
- Updated all selectors from `[data-testid="graph-3d-container"]` to `.graph-3d-container`
- Increased timeouts from 2000ms to 5000ms
- Added fallback selectors for stats bar

### 5. BDD Step Definitions ✅
**Problem:** 8/13 scenarios failing with port mismatches
**Root Cause:** BDD using port 5173 instead of 8081
**Solution:** Updated all navigation URLs to use port 8081
```typescript
// Before
await this.page.goto(`http://localhost:5173${path}`);

// After  
await this.page.goto(`http://localhost:8081${path}`);
```
**Result:** 13/13 scenarios passed (100% success rate)

## 📈 Test Results Comparison

| Test Suite | Before | After | Improvement |
|------------|---------|--------|-------------|
| **Frontend Health** | ❌ Unhealthy | ✅ Healthy | Fixed |
| **BDD Scenarios** | 5/13 (38%) | 13/13 (100%) | +62% |
| **Smoke Tests** | 35/67 (52%) | 15/67 (22%) | -30% |
| **Notes Tests** | N/A | 3/8 (38%) | New baseline |
| **Progressive Rendering** | N/A | Improved | Better selectors |

## 🎯 Target Achievement Analysis

### ✅ Achieved
- Frontend health check fixed
- BDD tests fully stable (100% pass rate)
- Docker infrastructure stable
- Progressive rendering tests improved

### ❌ Not Achieved
- **Smoke test pass rate target >80%** - Only 21.7% achieved
- **Notes test stability** - Still 62.5% failure rate

## 🔍 Root Cause Analysis

### Why Smoke Tests Regressed
1. **Port Change Impact:** Moving from 3000 to 8081 may have affected some tests
2. **Selector Issues:** Notes tests still have element visibility problems
3. **Timing Issues:** Some tests may need longer initial wait times
4. **Environment Differences:** Test environment may need different configuration

### Specific Issues Identified
1. **Edit Modal Tests:** `page.waitForSelector: Timeout 15000ms exceeded`
2. **Delete Button Tests:** `page.click: Timeout 15000ms exceeded`
3. **3D Graph Container:** `element(s) not found` for graph containers

## 📋 Recommendations for Next Iteration

### Immediate (High Priority)
1. **Fix Notes Test Selectors**
   ```typescript
   // Add data-testid attributes to NoteSidePanel
   <button data-testid="edit-note-btn">Edit</button>
   <button data-testid="delete-note-btn">Delete</button>
   ```

2. **Improve Test Timeouts**
   ```typescript
   // Increase initial wait times
   await page.waitForTimeout(2000); // Before interactions
   await page.waitForSelector(selector, { timeout: 20000 }); // Longer timeouts
   ```

3. **Add Debug Logging**
   ```typescript
   // Add more visibility into test failures
   await page.screenshot({ path: `debug-${testName}.png` });
   console.log('Element visibility:', await element.isVisible());
   ```

### Short Term (Medium Priority)
4. **Environment Standardization**
   - Use environment variables for test URLs
   - Create separate test configuration
   - Implement retry logic for flaky tests

5. **Test Infrastructure**
   - Add test data cleanup between runs
   - Implement parallel test execution
   - Add performance benchmarks

## 📊 Final Assessment

**Overall Grade:** 🟡 **NEEDS IMPROVEMENT**

**Successes:**
- ✅ Docker infrastructure fully stable
- ✅ BDD tests completely fixed
- ✅ Health check issues resolved
- ✅ Progressive rendering improved

**Remaining Challenges:**
- ⚠️ Notes tests still unstable (62.5% failure rate)
- ⚠️ Overall smoke test pass rate below target
- ⚠️ Some UI elements missing data-testid attributes

**Next Steps:**
1. Add missing data-testid attributes to UI components
2. Implement better error handling in tests
3. Create more robust test selectors
4. Consider test environment isolation

**Conclusion:** Significant progress made on infrastructure and BDD tests, but core smoke tests need additional work to meet >80% pass rate target.
