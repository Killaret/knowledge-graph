# Manual Test Report

**Date:** July 12, 2026  
**Time:** 15:36 UTC (updated 13:55 UTC)  
**Environment:** Windows, Docker Desktop  
**Branch:** ai-agents  
**Test Stack:** docker-compose.test.yml (ports: 3002, 8083)  
**Tester:** Manual testing with AI assistant

---

## Executive Summary

Manual testing of Knowledge Graph system in progress. Multiple infrastructure issues were discovered and fixed during test stack setup. Test stack is now healthy with seeded data. Frontend testing can now proceed.

**Overall Status:** ⏳ **IN PROGRESS**

---

## Pre-Testing Setup

### ✅ Isolated Testing Model
- **Status:** PASSED
- **Details:**
  - Dev stack stopped to free resources
  - Personal stack stopped to free resources
  - Only test stack is running
- **Verdict:** Isolated model successfully applied

### ✅ Check Stacks Health
- **Status:** PASSED
- **Details:**
  - Test stack: 7 containers, health OK, API OK
- **Verdict:** Test stack healthy

### ✅ Test Stack Startup
- **Status:** PASSED (after multiple fixes)
- **Details:**
  - Initial issue: port 8083 conflict with personal stack
  - Fixed by stopping personal stack
  - Initial issue: `VITE_API_URL` not passed as build arg
  - Fixed by adding `build.args` to docker-compose.test.yml
  - Initial issue: backend connected to dev postgres (`postgres`) instead of test postgres (`postgres-test`)
  - Fixed by removing `env_file: .env` from backend-test and postgres-test
  - Initial issue: test postgres password mismatch
  - Fixed by using explicit credentials in docker-compose.test.yml
- **Verdict:** Test stack operational on ports 3002/8083

### ✅ Seed Test Data
- **Status:** PASSED
- **Details:**
  - Test user registered successfully
  - 5 test notes created (star, planet, comet, galaxy, asteroid)
  - 2 test links created
- **Verdict:** Test data ready

---

## Smoke Tests

### 🔄 Public Access
- **Status:** IN PROGRESS
- **URL:** http://localhost:3002
- **Expected:** Canvas loads successfully with public graph visible
- **Previous Issue:** Frontend requested wrong API URL (port 8080 dev stack)
- **Fix Applied:** Rebuilt frontend with `VITE_API_URL=http://localhost:8083`
- **Next Step:** Verify page loads correctly in browser

---

## Issues Found & Fixed

### Issue #1: Frontend API URL Misconfiguration ✅ FIXED
- **Severity:** 🔴 Critical
- **Component:** Frontend (test stack)
- **Description:** Frontend making API requests to `http://localhost:8080` (dev stack) instead of `http://localhost:8083` (test stack)
- **Root Cause:** `VITE_API_URL` passed as runtime `environment` instead of build-time `args`
- **Fix:** Added `build.args` section to `frontend-test` in `docker-compose.test.yml`
- **Status:** Resolved
- **Evidence:** HAR file showed API requests to port 8080 instead of expected test stack port

### Issue #2: Backend Connected to Dev Database ✅ FIXED
- **Severity:** 🔴 Critical
- **Component:** Backend (test stack)
- **Description:** Backend failed to connect to database
- **Root Cause:** `env_file: .env` loaded dev configuration with `DATABASE_URL` pointing to `postgres` host
- **Fix:** Removed `env_file: .env` from backend-test and postgres-test; set explicit test values
- **Status:** Resolved

### Issue #3: Test Postgres Password Mismatch ✅ FIXED
- **Severity:** 🔴 Critical
- **Component:** PostgreSQL (test stack)
- **Description:** Password authentication failed for `kb_user`
- **Root Cause:** `env_file: .env` set `POSTGRES_PASSWORD=change_me_in_production` but backend used `kb_password`
- **Fix:** Removed `env_file: .env` from postgres-test and set explicit password `kb_password`
- **Status:** Resolved

### Issue #4: Test Scripts Used Wrong Ports ✅ FIXED
- **Severity:** 🟡 Medium
- **Component:** Scripts `start-test.ps1`, `start-test.sh`, `seed-test-data.ps1`, `seed-test-data.sh`
- **Description:** Scripts referenced ports 13002/18083 instead of 3002/8083
- **Fix:** Updated all scripts to use correct ports
- **Status:** Resolved

### Issue #5: Type Filter UI Layout Problem ⏳ OPEN
- **Severity:** 🟡 Medium
- **Component:** Frontend UI (FloatingControls type filters)
- **Description:** Type filter chips have inconsistent spacing and are cut off on different sides
- **Details:** Filter chip container layout issue causing elements to be clipped
- **Status:** Open - needs verification after API fix
- **Priority:** P2 - Visual bug, doesn't block functionality
- **Screenshot:** Image 1

---

## Test Environment

### Test Stack Configuration (Isolated)
- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:8083
- **PostgreSQL:** localhost:15434
- **Redis:** localhost:16381
- **MongoDB:** localhost:27019
- **NLP Service:** localhost:15002
- **Test User:** testuser / TestPassword123!

### Dev Stack Configuration
- **Status:** STOPPED for isolated testing
- **Frontend:** http://localhost:5173
- **Backend:** http://localhost:9000
- **Nginx Gateway:** http://localhost:8080

### Personal Stack Configuration
- **Status:** STOPPED for isolated testing
- **Frontend:** http://localhost:3001
- **Backend:** http://localhost:8085
- **Nginx Gateway:** http://localhost:8082

---

## Next Steps

1. ✅ **Investigate Issue #1:** Frontend API URL - FIXED
2. ✅ **Verify test stack health:** All services healthy
3. 🔄 **Continue smoke tests:** Verify http://localhost:3002 loads correctly
4. ⏳ **Canvas features testing:** Ghost node, black hole, drag-and-drop links
5. ⏳ **Note cards testing:** Visual style, tooltips, indicators
6. ⏳ **Hotkeys testing:** N, F, ?, Esc, Delete
7. ⏳ **Public graph verification:** Public notes and links visible without auth
8. ⏳ **Final verdict:** Determine system readiness

---

## Configuration Changes Made

### `docker-compose.test.yml`
- Added `build.args` for `frontend-test`:
  - `VITE_API_URL: http://localhost:8083`
  - `VITE_BACKEND_URL: http://localhost:8083`
  - `VITE_SKIP_AUTH: true`
- Removed `env_file: .env` from `backend-test`
- Removed `env_file: .env` from `postgres-test`
- Set explicit environment variables for test services
- Updated ports: frontend 3002, backend 8083
- Updated CORS_ALLOWED_ORIGINS to include localhost:3002

### `scripts/start-test.ps1`
- Fixed PowerShell syntax error in while loop
- Updated ports in output messages

### `scripts/start-test.sh`
- Updated ports in output messages

### `scripts/seed-test-data.ps1`
- Updated API URL to port 8083

### `scripts/seed-test-data.sh`
- Updated API URL to port 8083

---

## Notes

- Isolated testing model successfully applied: dev and personal stacks stopped
- Test stack fully isolated on ports 3002/8083
- Multiple infrastructure bugs discovered and fixed during setup
- Frontend should now correctly connect to test backend
- Issue #5 (UI layout) needs verification in browser

---

**Report Generated:** July 12, 2026  
**Last Updated:** 13:55 UTC