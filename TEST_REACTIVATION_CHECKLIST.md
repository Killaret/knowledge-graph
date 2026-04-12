# Checklist: Re-enabling Graph Tests After SSR Fix

## Overview
After the SSR fix for three-forcegraph is implemented, follow this checklist to ensure all tests are properly re-enabled and passing.

---

## Files to Check

### 1. `frontend/tests/ui-components.spec.ts`

**Tests to verify:**
- [x] ✅ `'should show note popup on node click'` - Line 122 (RE-ENABLED)
- [x] ✅ `'should show empty state for isolated node'` - Line 147 (RE-ENABLED)

**Verification steps:**
```bash
cd frontend
npx playwright test ui-components.spec.ts --grep "Graph Page Features" --reporter=list
```

### 2. `frontend/tests/graph-3d.spec.ts`

**Check if exists and re-enable all skipped tests:**
- [ ] Check if file exists: `ls tests/graph-3d.spec.ts`
- [ ] If exists, remove `.skip()` from all graph-related tests
- [ ] If file doesn't exist, consider creating basic graph tests:
  - `'should render 3D graph canvas'`
  - `'should display nodes as stars'`
  - `'should show node labels'`
  - `'should handle zoom and pan'`

---

## Pre-Flight Checks

Before running tests, ensure:

- [ ] Backend is running (`docker compose up` or `go run cmd/server/main.go`)
- [ ] Frontend dev server is running (`npm run dev`)
- [ ] Build passes without errors (`npm run build`)
- [ ] TypeScript check passes (`npm run check`)

---

## Test Execution

### Step 1: Run specific graph tests
```bash
npx playwright test ui-components.spec.ts --grep "Graph" --reporter=list
```

### Step 2: Run all UI component tests
```bash
npx playwright test ui-components.spec.ts --reporter=list
```

### Step 3: Run full test suite
```bash
npx playwright test
```

---

## Expected Results After Fix

All tests should pass:
```
✅ should show note popup on node click (5s)
✅ should show empty state for isolated node (3s)
```

---

## Troubleshooting

### If tests still fail after SSR fix:

1. **Check canvas rendering:**
   ```typescript
   await page.waitForSelector('canvas', { timeout: 10000 });
   ```

2. **Add extra wait time for graph initialization:**
   ```typescript
   await page.waitForTimeout(3000); // Increase if needed
   ```

3. **Check for console errors:**
   ```typescript
   page.on('console', msg => console.log(msg.text()));
   ```

4. **Verify Three.js is loading:**
   Check if `THREE` is available in window object

---

## CI/CD Considerations

When creating PR with SSR fix:

1. **Update CI workflow** (if needed):
   ```yaml
   - name: Run E2E tests
     run: npx playwright test
     env:
       CI: true
   ```

2. **Update test configuration** (playwright.config.ts):
   - Ensure timeout is sufficient for 3D rendering (30s+ for graph tests)
   - Add retry logic for flaky canvas tests

3. **Document breaking changes** in PR description:
   - SSR fix may change initial render behavior
   - Document any new environment variables or build steps

---

## Status Tracking

| Test Name | Status | Notes |
|-----------|--------|-------|
| `should show note popup on node click` | ⏳ ACTIVE | Waiting for SSR fix |
| `should show empty state for isolated node` | ⏳ ACTIVE | Waiting for SSR fix |
| `3D graph canvas rendering` | ⏳ PENDING | Need to create if missing |
| `node interaction tests` | ⏳ PENDING | Need to create |

---

## Related Files

- `frontend/src/lib/components/Graph3D.svelte` - Main component needing SSR fix
- `frontend/src/routes/+page.svelte` - New home with global graph
- `frontend/src/routes/graph/[id]/+page.svelte` - Note-specific graph
- `frontend/tests/ui-components.spec.ts` - Re-enabled tests

---

## Commands Summary

```bash
# Run specific test
npx playwright test ui-components.spec.ts -g "should show note popup"

# Run with UI mode for debugging
npx playwright test ui-components.spec.ts --ui

# Run with trace viewer
npx playwright test ui-components.spec.ts --trace on
```

---

**Last Updated:** $(date)
**Branch:** feature/ui-ux-improvements
