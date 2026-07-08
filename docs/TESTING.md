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
| postgres-test | kg-test-postgres | 15434 | Test database |
| redis-test | kg-test-redis | 16381 | Test cache/queue |
| mongo-test | kg-test-mongo | 27018 | Test drafts |
| nlp-test | kg-test-nlp | 15002 | Test NLP service |
| backend-test | kg-test-backend | 18083 | Test backend API |
| frontend-test | kg-test-frontend | 13002 | Test frontend |

### Configuration

- **SKIP_AUTH: true** - Authentication bypassed for testing
- **REDIS_FLUSH_ON_STARTUP: true** - Redis cleared on startup
- **Database: knowledge_test** - Separate test database

## Automated Testing Scripts

### check-stacks-health
Checks the health of dev and personal stacks.

**Windows:**
```powershell
.\scripts\check-stacks-health.ps1
```

**Linux/Mac:**
```bash
./scripts/check-stacks-health.sh
```

**Checks:**
- Dev containers running
- Dev health endpoint (http://localhost:8080/health)
- Dev API (http://localhost:8080/api/v1/notes?limit=1)
- Personal containers running
- Personal health endpoint (http://localhost:8082/health)
- Personal API (http://localhost:8082/api/v1/notes?limit=1)

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
Orchestrates the complete testing cycle.

**Windows:**
```powershell
.\scripts\run-full-test-cycle.ps1
```

**Linux/Mac:**
```bash
./scripts/run-full-test-cycle.sh
```

**Steps:**
1. Check dev and personal stacks health
2. Start test stack
3. Seed test data
4. Display manual testing instructions
5. Wait for manual testing (press Enter)
6. Stop test stack
7. Check dev and personal stacks health again
8. Display summary

## Manual Testing

### Test Environment

After starting the test stack, access the test environment at:
- **Frontend:** http://localhost:13002
- **Backend API:** http://localhost:18083

### Test User Credentials

- **Login:** testuser
- **Password:** TestPassword123!

### Manual Test Checklist

Follow the manual test checklist at `docs/MANUAL_TEST_CHECKLISTS.md` for detailed testing procedures.

### Test Coverage

The manual test checklist covers:
- Canvas features (ghost node, black hole, drag-and-drop links, hotkeys)
- Note cards (visual style, batch operations, undo, sorting)
- General UX (galactic lexicon, browser console, language switch)

## Automated Tests

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
- Port conflicts (13002, 18083, 15434, 16381, 27018, 15002)
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
- [Test Execution Report](TEST_EXECUTION_REPORT.md)
- [Backend Testing](../backend/README.md#testing)
- [Frontend Testing](../frontend/tests/README.md)
