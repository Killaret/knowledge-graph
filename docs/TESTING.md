# Testing Guide

This document describes the testing infrastructure and procedures for Knowledge Graph.

## Overview

Knowledge Graph uses three Docker stacks:
- **Dev stack** (docker-compose.yml) - Development environment (frontend dev server 5173, backend 9000, nginx API 18080/frontend 18081)
- **Personal stack** (docker-compose.personal.yml) - Personal environment (frontend 3001, backend direct 18085, nginx API 18082/frontend 18084)
- **Test stack** (docker-compose.test.yml) - Isolated testing environment (frontend 3002, backend 18083, postgres 15434, redis 16381, mongo 27019, nlp 15002, graph-service 9095)

## Test Stack

The test stack is fully isolated from dev and personal stacks:
- Separate PostgreSQL database (knowledge_test)
- Separate Redis instance
- Separate MongoDB instance
- Separate NLP service
- Separate backend and frontend containers
- Separate volumes (test_postgres_data, test_mongodb_data)

### Services

| Service | Container Name | Port | Purpose |
|---------|---------------|------|---------|
| postgres-test | kg-test-postgres | 15434 | Test database |
| redis-test | kg-test-redis | 16381 | Test cache/queue |
| mongo-test | kg-test-mongo | 27019 | Test drafts |
| nlp-test | kg-test-nlp | 15002 | Test NLP service |
| backend-test | kg-test-backend | 18083 | Test backend API |
| graph-service-test | kg-test-graph-service | 19090/19091 | Test graph analytics service |
| frontend-test | kg-test-frontend | 3002 | Test frontend |

### Configuration

- **SKIP_AUTH: true** - Authentication bypassed for testing
- **REDIS_FLUSH_ON_STARTUP: true** - Redis cleared on startup
- **Database: knowledge_test** - Separate test database

### Test Stack URLs

- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:18083
- **Graph Service (HTTP):** http://localhost:19091

### Health Checks

- **Backend:** `curl http://127.0.0.1:18083/health`
- **Graph Service:** `curl http://127.0.0.1:19091/health`

## Isolated Testing Model

**⚠️ IMPORTANT:** Knowledge Graph uses an isolated testing model to ensure accurate test results and prevent resource conflicts.

### Overview

The isolated testing model ensures that:
- **Only the test stack runs during testing** - dev and personal stacks are stopped
- **No resource conflicts** - Eliminates Docker API instability from running multiple stacks
- **Accurate test results** - Tests run on clean, isolated environment
- **State verification** - Automatic comparison of dev stack state before/after testing
- **Resource efficiency** - Optimizes resource usage during testing

### Testing Process

1. **Pre-test snapshot** - Capture dev stack state (containers, health, API)
2. **Stop dev/personal stacks** - Free up resources for testing
3. **Run tests on isolated test stack** - Clean environment, no conflicts
4. **Stop test stack** - Complete cleanup with volume removal
5. **Restore dev/personal stacks** - Bring back development environments
6. **Post-test comparison** - Verify dev stack state unchanged
7. **Dev/Personal identity check** - Verify stacks are identical
8. **Auto-commit** - Commit with test success marker if all checks pass

### Benefits

- **Docker stability** - Prevents Docker API instability from running multiple stacks
- **Resource efficiency** - Only test stack uses resources during testing
- **Accurate results** - Tests run on clean, isolated environment
- **State verification** - Automatic comparison of dev stack state before/after testing
- **No conflicts** - Eliminates port and resource conflicts between stacks
- **Auto-commit** - Automatic commit with test success marker when all checks pass
- **Identity verification** - Automatic comparison of dev and personal stacks

### When to Use Isolated Testing

- **Full regression testing** - Before production deployment
- **E2E and BDD testing** - Always use isolated test stack
- **Integration testing** - When testing with real databases
- **Performance testing** - When measuring system performance

### When to Use Concurrent Stacks

- **Manual testing** - When testing features across dev/personal stacks
- **Feature development** - When working on features in dev stack
- **Personal use** - When using personal stack for daily work

## Automated Testing Scripts

### check-stacks-health
Checks the health of specified stack(s).

**Windows:**
```powershell
.\scripts\ci\check-stacks-health.ps1 -Stack <dev|personal|test|all>
```

**Linux/Mac:**
```bash
./scripts/ci/check-stacks-health.sh --stack <dev|personal|test|all>
```

**Parameters:**
- `-Stack` / `--stack`: Specify which stack to check (default: all)
  - `dev` - Check only dev stack
  - `personal` - Check only personal stack
  - `test` - Check only test stack
  - `all` - Check all stacks (default)

**Checks (for each stack):**
- Containers running
- Health endpoint
- API endpoint

### start-test
Starts the isolated test stack.

**Windows:**
```powershell
.\scripts\testing\start-test.ps1
```

**Linux/Mac:**
```bash
./scripts/testing/start-test.sh
```

**Actions:**
- Stops and removes previous test stack (with volumes)
- Builds and starts test stack
- Waits for all containers to be healthy
- Displays test stack URLs

### stop-test
Stops and destroys the test stack.

**Windows:**
```powershell
.\scripts\testing\stop-test.ps1
```

**Linux/Mac:**
```bash
./scripts/testing/stop-test.sh
```

**Actions:**
- Stops test stack
- Removes volumes (complete cleanup)

### seed-test-data
Seeds the test database with test data.

**Windows:**
```powershell
.\scripts\testing\seed-test-data.ps1
```

**Linux/Mac:**
```bash
./scripts/testing/seed-test-data.sh
```

**Creates:**
- Test user (login: testuser, password: TestPassword123!)
- 5 test notes (star, planet, comet, galaxy, asteroid)
- 2 test links between notes

### run-full-test-cycle
Orchestrates the complete testing cycle with **full stack isolation**.

**⚠️ IMPORTANT:** This script uses an isolated testing model where dev and personal stacks are stopped during testing to prevent resource conflicts and ensure accurate test results.

**Windows:**
```powershell
.\scripts\testing\run-full-test-cycle.ps1
```

**Linux/Mac:**
```bash
./scripts/testing/run-full-test-cycle.sh
```

**Isolated Testing Model Steps (25 total):**
1. **Capture dev stack state snapshot** - Save container state, health endpoint, and API response
2. **Stop dev stack** - `docker compose down`
3. **Stop personal stack** - `docker compose -f docker-compose.personal.yml down`
4. **Check stacks identity** - Verify dev/personal/test consistency
5. **Start test stack** - `start-test.ps1`
6. **Seed test data** - `seed-test-data.ps1`
7. **Docker build verification** - Check Docker images
8. **NLP service tests** - Verify NLP health and functionality
9. **Backend unit tests** - `go test ./...`
10. **Backend integration tests** - `go test -tags=integration ./...` (requires Linux/WSL Docker)
11. **Backend API verification** - Test critical endpoints
12. **Asynchronous tasks verification** - Check worker and Redis
13. **PGVECTOR verification** - Verify pgvector extension
14. **Redis & MongoDB verification** - Check data layer
15. **Frontend unit tests** - `npm run test:unit`
16. **Manual testing instructions** - Display URLs and credentials
17. **Public graph verification** - Manual verification
18. **CI/CD verification** - Manual verification
19. **Documentation verification** - Verify `docs/AGENTS.md`, `.windsurfrules` and architecture docs are updated if boundaries changed
20. **Stop test stack** - `stop-test.ps1`
21. **Cleanup temporary files** - `cleanup-test-artifacts.ps1` (removes `coverage.out`, `*.cov`, `frontend/coverage`, `backend/.coverage_tmp`, `*.log`)
22. **Start dev stack** - `docker compose up -d --wait`
23. **Start personal stack** - `docker compose -f docker-compose.personal.yml up -d --wait`
24. **Compare dev stack state / identity / health** - Compare with pre-test snapshot, verify dev/personal identity, check health
25. **Auto-commit** - If all checks passed, commit with test success marker

**Automatic State Verification:**
- **Pre-test snapshot:** Captures dev stack state before testing
- **Post-test comparison:** Compares dev stack state after testing
- **Dev/Personal identity:** Verifies dev and personal stacks are identical
- **Auto-commit:** Only if dev state unchanged and dev/personal identical
- **Failure handling:** Stops with exit code 1 if differences found

**Benefits of Isolated Testing:**
- **Resource efficiency** - Only test stack uses resources during testing
- **No conflicts** - Eliminates port and resource conflicts between stacks
- **Accurate results** - Tests run on clean, isolated environment
- **State verification** - Automatic comparison of dev stack state before/after testing
- **Docker stability** - Prevents Docker API instability from running multiple stacks

**Temporary Files and Snapshots:**
- All temporary snapshots are saved to `scripts/testing/temp/snapshots/YYYYMMDD_HHMMSS/`.
- Includes: container state, health endpoint, API response.
- Post-test snapshots are saved to the same directory for comparison.
- Argos visual screenshots are saved to `frontend/argos-screenshots/` (see docs/ARGOS.md).
- These directories are ignored by Git — only source changes are committed.

## Auto-Commit on Successful Testing

**When all checks pass:**
- Dev stack state unchanged (pre-test vs post-test)
- Dev and personal stacks identical
- Dev and personal stacks healthy

**Auto-commit action:**
```bash
git add -A
git commit -m "test: successful regression cycle — dev and personal identical"
git push
```

**Commit message includes:**
- Test success marker
- Dev/Personal identity confirmation
- Co-authored-by tag for Devin

**When checks fail:**
- Dev stack state changed → Skip auto-commit, show warning
- Dev/Personal not identical → Exit with code 1, skip auto-commit
- Stacks not healthy → Exit with code 1, skip auto-commit

**Manual investigation required:**
- Check snapshot differences in `scripts/testing/temp/snapshots/YYYYMMDD_HHMMSS/`
- Review diff output for dev/personal differences
- Fix issues before re-running test cycle

## Manual Testing

### Test Environment

After starting the test stack, access the test environment at:
- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:18083

### Test User Credentials

- **Login:** testuser
- **Password:** TestPassword123!

### Manual Test Checklist

Follow the manual test checklist at `docs/MANUAL_TEST_CHECKLISTS_RU.md` for detailed testing procedures.

### Test Coverage

The manual test checklist covers:
- **Pre-testing setup** - Stack health checks, test stack startup, data seeding
- **Smoke tests** - Public access, authentication, profile, graph, note creation, logout
- **Public graph verification** - Public note/link creation, API access without auth, frontend public access
- **Canvas features** - Ghost node, black hole, drag-and-drop links, hotkeys, tooltips, new indicators
- **Note cards** - Visual style, batch operations, undo, sorting, dust style, card tooltips
- **General UX** - Galactic lexicon, browser console, language switch
- **Post-testing cleanup** - Test stack destruction, stacks health verification, defect reporting

## Automated Tests

### Current Test Statistics

**Latest Test Results (run 2026-07-20):**

| Layer | Category | Total | Passed | Failed | Skipped | Status |
|-------|----------|-------|--------|--------|---------|--------|
| Backend | Unit Tests | 1089 | 1085 | 0 | 4 | ✅ Excellent |
| Backend | Integration Tests | - | - | - | - | ⚠️ Not run — requires Linux/WSL Docker (`-tags=integration`) |
| Frontend | Unit Tests | 617 | 580 | 0 | 37 | ✅ Good |
| Frontend | E2E Tests | - | - | - | - | ⚠️ Run separately with `npm run test` |
| Frontend | BDD Tests | - | - | - | - | ⚠️ Run separately with `npm run test:bdd` |
| NLP | API + Utils Tests | 46 | 46 | 0 | 0 | ✅ Excellent |
| **NLP Total** | - | **46** | **46** | **0** | **0** | ✅ **Excellent** |

**Notes:**
- Backend unit tests (`go test ./...`) pass with 1085 passing, 4 skipped, 0 failures.
- Backend integration tests are excluded by default; run `go test -tags=integration ./...` on Linux/WSL or in CI.
- Frontend E2E and BDD tests are not part of `npm run test:unit`; they require the isolated test stack.

### Current Code Coverage

**Latest Coverage Results (run 2026-07-20):**

| Layer | Metric | Value | Target | Status |
|-------|--------|-------|--------|--------|
| Backend | Statements | **60.5%** | 70% (min 60%) | ⚠️ At minimum threshold |
| Frontend | Statements | **63.63%** | 70% (min 60%) | ⚠️ Below target |
| Frontend | Branches | **78.74%** | - | ✅ Good |
| Frontend | Functions | **56.91%** | 55% (min) | ✅ Above minimum |
| Frontend | Lines | **63.63%** | 60% (min) | ✅ Above minimum |

**Backend coverage gaps (packages below 60%):**
- `cmd/worker` (14.6%), `internal/infrastructure/mongo` (15.3%), `internal/infrastructure/db` (20.0%)
- `internal/infrastructure/cloud` (34.3%), `internal/infrastructure/db/postgres` (37.2%)
- `internal/interfaces/api/handlers/auth` (43.5%), `internal/infrastructure/queue` (46.0%)
- `cmd/checkconfig` (47.2%), `internal/domain/user` (50.7%)
- `internal/interfaces/api/handlers/share` (56.7%), `internal/application/cache` (57.1%)
- `internal/interfaces/api/handlers/draft` (58.7%), `internal/interfaces/api/notehandler` (59.8%)

**Frontend coverage gaps (files/directories below 60%):**
- `features/graph-interaction` (~28%), `features/graph-forms` (~22%), `features/graph-canvas` (~41%)
- `shared/stores` (~43%), `shared/api` (~54%), `shared/services` (~59%)
- Several form components (`ForgotPasswordForm`, `RegisterForm`, `ResetPasswordForm`) at 0%
- `GraphCanvas.svelte` interaction/zoom-pan/pan handlers and `delta.ts` largely uncovered

### Backend Tests

**Unit tests:**
```bash
cd backend
go test ./...
```

**Integration tests:**
```bash
cd backend
go test -tags=integration ./...
```

### Graph Service Tests

**Unit tests:**
```bash
cd services/graph-service
go test ./...
```

### Frontend Tests

**Unit tests:**
```bash
cd frontend
npm run test:unit
```

**E2E tests:**
```bash
cd frontend
npm run test
```

**BDD tests:**
```bash
cd frontend
npm run test:bdd
```

### NLP Tests

```bash
cd nlp-service
pytest tests/ -v
```

### Visual Regression / Argos

Visual regression tests are located in `frontend/tests/visual/visual-regression.spec.ts` and use `@argos-ci/playwright`. The reporter uploads screenshots to Argos automatically when `CI` or `ARGOS_UPLOAD_LOCAL` is set.

**Run locally (test stack):**
```powershell
# Windows
./scripts/testing/start-test.ps1
./scripts/testing/seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42

cd frontend
$env:FRONTEND_URL = "http://localhost:3002"
npm run test:visual:upload   # requires ARGOS_TOKEN env variable
```

```bash
# Linux/Mac
SKIP_AUTH=true ./scripts/testing/start-test.sh
NOTE_COUNT=20 LINK_COUNT=10 SEED=42 ./scripts/testing/seed-test-data.sh

cd frontend
FRONTEND_URL=http://localhost:3002 ARGOS_UPLOAD_LOCAL=true npm run test:visual
```

**Configuration:**
- `ARGOS_TOKEN` — required for upload.
- `FRONTEND_URL` — defaults to `http://localhost:3002` (test stack).
- `ARGOS_REFERENCE_BRANCH` — baseline branch (`ai-agents` for this work).
- `ARGOS_UPLOAD_LOCAL` — set to `true` to upload from a local run.

**Determinism helpers:**
- Tests append `?stableRender=true` to disable animated variations.
- `page.addInitScript` injects a seeded `Math.random` LCG and `__SKIP_AUTH__`.
- `data-visual-test="transparent"` masks dynamic timestamps and indicators.
- `data-testid="graph-canvas"` exposes `data-test-stable="true"` after the force simulation settles.

**CLI upload manually (legacy — not required with the Playwright reporter):**
```bash
cd frontend
npx argos upload ./argos-screenshots --token $ARGOS_TOKEN
```

See `docs/ARGOS.md` for the full visual regression workflow.

## Test Data Isolation

The test stack ensures complete isolation:
- Separate database (knowledge_test vs knowledge_base/knowledge_personal)
- Separate volumes (test_postgres_data vs postgres_data)
- Separate container names (kg-test-* vs kg-*)
- Separate ports (3002/18083 vs 5173/9000/18080/18081 and 3001/18085/18082/18084)

## Cleanup

After testing, always run the cleanup scripts to ensure complete isolation:

```powershell
.\scripts\testing\stop-test.ps1
python .\scripts\cleanup\cleanup-test-artifacts.py
```

or

```bash
./scripts/testing/stop-test.sh
python ./scripts/cleanup/cleanup-test-artifacts.py
```

This ensures:
- Test containers are stopped
- Test volumes are removed
- Temporary artifacts (`coverage.out`, `*.cov`, `frontend/coverage`, `backend/.coverage_tmp`, `*.log`) are removed
- No data leakage to dev/personal stacks

## Troubleshooting

### Dev stack 502 error (FIXED)

**Issue:** Dev nginx returns 502 Bad Gateway for API requests

**Root Cause:** Incorrect container names in nginx.conf

**Fix:** Updated nginx.conf to use correct container names (kg-backend, kg-frontend, kg-graph-service)

**Verification:**
```bash
curl http://localhost:9000/health
curl http://localhost:9000/api/v1/notes?limit=1
```

**Status:** ✅ Resolved - Dev stack API is now accessible

### Test stack won't start

**Check Docker:**
```bash
docker ps
docker compose -f docker-compose.test.yml ps
```

**Check logs:**
```bash
docker compose -f docker-compose.test.yml logs
```

**Common issues:**
- Port conflicts (3002, 18083, 15434, 16381, 27019, 15002)
- Docker Desktop not running
- Insufficient resources

### Test data seeding fails

**Check backend health:**
```bash
curl http://localhost:18083/health
```

**Check API:**
```bash
curl http://localhost:18083/api/v1/notes
```

**Common issues:**
- Backend not ready
- Network issues
- Invalid test data

### Dev/personal stacks affected

**Verify isolation:**
```bash
docker ps --filter "name=kg-"
```

**Check for test containers:**
```bash
docker ps --filter "name=kg-test"
```

**If test containers exist:**
```bash
docker compose -f docker-compose.test.yml down -v
```

### Multiple stacks cause Docker/Playwright failures

**Issue:** Docker becomes unstable, test stack containers fail health checks, or Playwright reports `ECONNREFUSED ::1:18083` / `net::ERR_CONNECTION_REFUSED`.

**Root Cause:** Running dev, personal, and test stacks simultaneously exhausts Docker resources and creates port/network conflicts. On Windows, Node/Playwright resolves `localhost` to `::1` first, but Docker Desktop binds published ports to `127.0.0.1` by default.

**Fix:**
1. Stop dev and personal stacks before E2E/BDD/regression:
   ```bash
   docker compose down
   docker compose -f docker-compose.personal.yml down
   ```
2. Start only the isolated test stack.
3. Use `127.0.0.1` URLs for Playwright/BDD:
   ```powershell
   $env:FRONTEND_URL = "http://127.0.0.1:3002"
   $env:BACKEND_URL = "http://127.0.0.1:18083"
   ```
4. Or rebuild the test frontend image with `VITE_API_URL=http://127.0.0.1:18083`:
   ```bash
   docker compose -f docker-compose.test.yml build --build-arg VITE_API_URL=http://127.0.0.1:18083 frontend-test
   ```

### Dev/personal PostgreSQL password mismatch

**Issue:** Backend fails with `password authentication failed for user "kb_user"` or `"personal"` after restoring dev/personal stacks.

**Root Cause:** Dev/personal `postgres` services load `env_file: .env`. If `.env` is missing or its `POSTGRES_PASSWORD` / `PERSONAL_POSTGRES_PASSWORD` does not match the password used when the `postgres_data` / `pgdata_personal` volume was initialized, authentication fails.

**Fix:** Ensure `.env` contains:
```env
JWT_SECRET=your-dev-jwt-secret
POSTGRES_PASSWORD=<password matching postgres_data volume>
PERSONAL_POSTGRES_PASSWORD=<password matching pgdata_personal volume>
```

## Best Practices

1. **Always check stacks health before testing** - Ensures dev/personal stacks are stable
2. **Use the full test cycle script** - Automates the entire process
3. **Clean up after testing** - Prevents data leakage
4. **Verify isolation** - Check that test stack doesn't affect dev/personal
5. **Document issues** - Use the reporting section in the manual checklist

## CI/CD Integration

The test stack can be integrated into CI/CD pipelines:

```yaml
- name: Start test stack
  run: ./scripts/testing/start-test.sh

- name: Seed test data
  run: ./scripts/testing/seed-test-data.sh

- name: Run tests
  run: npm run test

- name: Stop test stack
  run: ./scripts/testing/stop-test.sh
```

## References

- [Manual Test Checklist](MANUAL_TEST_CHECKLISTS_RU.md)
- [Regression Test Plan](REGRESSION_TEST_PLAN.md)
- [Final Test Report](archive/FINAL_TEST_REPORT.md)
- [Backend Testing](../backend/README.md#testing)
- [Frontend Testing](../frontend/tests/README.md)

## Recent Improvements

### Dev Stack Fix (July 2026)
- **Issue:** Nginx 502 Bad Gateway error on dev stack
- **Resolution:** Updated nginx.conf with correct container names
- **Status:** ✅ Resolved - Dev stack fully operational

### Test Stack Automation (July 2026)
- **New Scripts:** start-test, stop-test, seed-test-data, check-stacks-health, run-full-test-cycle
- **Isolation:** Complete separation from dev/personal stacks
- **Ports:** Frontend 3002, Backend 18083, PostgreSQL 15434, Redis 16381, MongoDB 27019, NLP 15002, Graph service 9095
- **Status:** ✅ Fully automated and verified

### Smoke Tests (July 2026)
- **Coverage:** Public access, authentication, profile, graph, note creation, logout
- **Status:** ✅ All smoke tests passing
- **Documentation:** Added to MANUAL_TEST_CHECKLISTS_RU.md

### Public Graph Verification (July 2026)
- **Feature:** Public notes and links accessible without authentication
- **Testing:** API and frontend verification for public graph access
- **Status:** ✅ Documented in MANUAL_TEST_CHECKLISTS_RU.md

### Regression Test Plan (July 2026)
- **New Document:** REGRESSION_TEST_PLAN.md
- **Coverage:** 20-part comprehensive regression testing plan
- **Includes:** Stacks identity, Docker builds, dependencies, security, infrastructure, all test layers
- **Status:** ✅ Documented and ready for execution

## Current Test Counts

| Category | Files | Tests/Scenarios | Notes |
|----------|-------|-----------------|-------|
| **Go Unit** | 99 | 596 test functions | `backend/**/*_test.go` |
| **Frontend Unit** | 84 | 976+ | `frontend/src/**/*.spec.ts` / `*.test.ts` |
| **Playwright E2E** | 16 | 122 | `frontend/tests/**/*.spec.ts` |
| **BDD (Cucumber)** | 14 | 127 scenarios | `tests/features/*.feature` |
| **NLP Python** | 2 | 46 | `nlp-service/tests/*.py` |

> Run `cd backend && go test ./...`, `cd frontend && npm run test:unit`, `cd frontend && npx playwright test`, `npm run test:bdd`, `cd nlp-service && pytest` to verify.
