# Testing Guide

This document describes the testing infrastructure and procedures for Knowledge Graph.

## Overview

Knowledge Graph uses three Docker stacks:
- **Dev stack** (docker-compose.yml) - Development environment (ports 3000/8080)
- **Personal stack** (docker-compose.personal.yml) - Personal environment (ports 3001/8082)
- **Test stack** (docker-compose.test.yml) - Isolated testing environment (ports 3002/8083)

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
| postgres-test | kg-test-postgres | 5434 | Test database |
| redis-test | kg-test-redis | 6381 | Test cache/queue |
| mongo-test | kg-test-mongo | 27018 | Test drafts |
| nlp-test | kg-test-nlp | 15002 | Test NLP service |
| backend-test | kg-test-backend | 8083 | Test backend API |
| frontend-test | kg-test-frontend | 3002 | Test frontend |

### Configuration

- **SKIP_AUTH: true** - Authentication bypassed for testing
- **REDIS_FLUSH_ON_STARTUP: true** - Redis cleared on startup
- **Database: knowledge_test** - Separate test database

### Test Stack URLs

- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:8083

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

### Benefits

- **Docker stability** - Prevents Docker API instability from running multiple stacks
- **Resource efficiency** - Only test stack uses resources during testing
- **Accurate results** - Tests run on clean, isolated environment
- **State verification** - Automatic comparison of dev stack state before/after testing
- **No conflicts** - Eliminates port and resource conflicts between stacks

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
.\scripts\check-stacks-health.ps1 -Stack <dev|personal|test|all>
```

**Linux/Mac:**
```bash
./scripts/check-stacks-health.sh --stack <dev|personal|test|all>
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
.\scripts\start-test.ps1
```

**Linux/Mac:**
```bash
./scripts/start-test.sh
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
.\scripts\stop-test.ps1
```

**Linux/Mac:**
```bash
./scripts/stop-test.sh
```

**Actions:**
- Stops test stack
- Removes volumes (complete cleanup)

### seed-test-data
Seeds the test database with test data.

**Windows:**
```powershell
.\scripts\seed-test-data.ps1
```

**Linux/Mac:**
```bash
./scripts/seed-test-data.sh
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
.\scripts\run-full-test-cycle.ps1
```

**Linux/Mac:**
```bash
./scripts/run-full-test-cycle.sh
```

**Isolated Testing Model Steps:**
1. **Capture dev stack state snapshot** - Save container state, health endpoint, and API response
2. **Stop dev stack** - `docker compose down`
3. **Stop personal stack** - `docker compose -f docker-compose.personal.yml down`
4. **Check stacks identity** - Verify dev/personal/test consistency
5. **Start test stack** - `start-test.ps1`
6. **Seed test data** - `seed-test-data.ps1`
7. **Docker build verification** - Check Docker images
8. **NLP service tests** - Verify NLP health and functionality
9. **Backend unit tests** - Run Go unit tests
10. **Backend API verification** - Test critical endpoints
11. **Asynchronous tasks verification** - Check worker and Redis
12. **PGVECTOR verification** - Verify pgvector extension
13. **Redis & MongoDB verification** - Check data layer
14. **Frontend unit tests** - Run Vitest tests
15. **Manual testing instructions** - Display URLs and credentials
16. **Public graph verification** - Manual verification
17. **CI/CD verification** - Manual verification
18. **Stop test stack** - `stop-test.ps1`
19. **Start dev stack** - `docker compose up -d --wait`
20. **Start personal stack** - `docker compose -f docker-compose.personal.yml up -d --wait`
21. **Compare dev stack state** - Compare with pre-test snapshot
22. **Check stacks health** - Verify dev and personal stacks are healthy

**Benefits of Isolated Testing:**
- **Resource efficiency** - Only test stack uses resources during testing
- **No conflicts** - Eliminates port and resource conflicts between stacks
- **Accurate results** - Tests run on clean, isolated environment
- **State verification** - Automatic comparison of dev stack state before/after testing
- **Docker stability** - Prevents Docker API instability from running multiple stacks

**Snapshots:**
- Pre-test snapshots saved to `test-snapshots_YYYYMMDD_HHMMSS/` directory
- Includes: container state, health endpoint, API response
- Post-test snapshots for comparison

## Manual Testing

### Test Environment

After starting the test stack, access the test environment at:
- **Frontend:** http://localhost:3002
- **Backend API:** http://localhost:8083

### Test User Credentials

- **Login:** testuser
- **Password:** TestPassword123!

### Manual Test Checklist

Follow the manual test checklist at `docs/MANUAL_TEST_CHECKLISTS.md` for detailed testing procedures.

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

**Latest Test Results (from FINAL_TEST_REPORT.md):**

| Layer | Category | Total | Passed | Failed | Skipped | Status |
|-------|----------|-------|--------|--------|---------|--------|
| Backend | Unit Tests | 118 | 118 | 0 | 0 | ✅ Excellent |
| Backend | Integration Tests | 2 | 0 | 2 | 0 | ❌ Failed (Windows limitation) |
| Frontend | Unit Tests | 563 | 521 | 5 | 37 | ✅ Good |
| Frontend | E2E Tests | 94 | 84 | 0 | 10 | ✅ Excellent |
| Frontend | BDD Tests | 5 | 5 | 0 | 0 | ✅ Excellent |
| NLP | API Tests | 17 | 17 | 0 | 0 | ✅ Excellent |
| NLP | Utils Tests | 16 | 11 | 0 | 5 | ✅ Good |
| **NLP Total** | - | **33** | **28** | **0** | **5** | ✅ **Excellent** |

**Notes:**
- Backend integration tests fail on Windows due to testcontainers rootless Docker limitation (not a code issue)
- Frontend unit test failures are due to Russian text in LoginForm tests (non-blocking, functionality correct)

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

## Test Data Isolation

The test stack ensures complete isolation:
- Separate database (knowledge_test vs knowledge_base/knowledge_personal)
- Separate volumes (test_postgres_data vs postgres_data)
- Separate container names (kg-test-* vs kg-*)
- Separate ports (13002/18083 vs 3000/8080 and 3001/8082)

## Cleanup

After testing, always run the cleanup script to ensure complete isolation:

```powershell
.\scripts\stop-test.ps1
```

or

```bash
./scripts/stop-test.sh
```

This ensures:
- Test containers are stopped
- Test volumes are removed
- No data leakage to dev/personal stacks

## Troubleshooting

### Dev stack 502 error (FIXED)

**Issue:** Dev nginx returns 502 Bad Gateway for API requests

**Root Cause:** Incorrect container names in nginx.conf

**Fix:** Updated nginx.conf to use correct container names (kg-backend, kg-frontend, kg-graph-service)

**Verification:**
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/notes?limit=1
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
- Port conflicts (3002, 8083, 5434, 6381, 27018, 15002)
- Docker Desktop not running
- Insufficient resources

### Test data seeding fails

**Check backend health:**
```bash
curl http://localhost:8083/health
```

**Check API:**
```bash
curl http://localhost:8083/api/v1/notes
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
  run: ./scripts/start-test.sh

- name: Seed test data
  run: ./scripts/seed-test-data.sh

- name: Run tests
  run: npm run test

- name: Stop test stack
  run: ./scripts/stop-test.sh
```

## References

- [Manual Test Checklist](MANUAL_TEST_CHECKLISTS.md)
- [Regression Test Plan](REGRESSION_TEST_PLAN.md)
- [Final Test Report](FINAL_TEST_REPORT.md)
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
- **Ports:** Frontend 3002, Backend 8083, PostgreSQL 5434, Redis 6381
- **Status:** ✅ Fully automated and verified

### Smoke Tests (July 2026)
- **Coverage:** Public access, authentication, profile, graph, note creation, logout
- **Status:** ✅ All smoke tests passing
- **Documentation:** Added to MANUAL_TEST_CHECKLISTS.md

### Public Graph Verification (July 2026)
- **Feature:** Public notes and links accessible without authentication
- **Testing:** API and frontend verification for public graph access
- **Status:** ✅ Documented in MANUAL_TEST_CHECKLISTS.md

### Regression Test Plan (July 2026)
- **New Document:** REGRESSION_TEST_PLAN.md
- **Coverage:** 20-part comprehensive regression testing plan
- **Includes:** Stacks identity, Docker builds, dependencies, security, infrastructure, all test layers
- **Status:** ✅ Documented and ready for execution
