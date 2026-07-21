## Testing Commands & Procedures

**Comprehensive testing commands for AI agents and developers:**

### Frontend Testing
```bash
# Unit tests (Vitest)
cd frontend && npm run test:unit

# E2E tests (Playwright)
cd frontend && npx playwright test

# Visual regression tests
cd frontend && npx playwright test --grep="@visual"

# BDD tests (Cucumber)
cd frontend && npm run test:bdd

# Build verification
cd frontend && npm run build
```

### Backend Testing
```bash
# Unit tests
cd backend && go test ./...

# Integration tests
cd backend && go test -tags=integration ./...

# Race detection (requires CGO_ENABLED=1)
cd backend && CGO_ENABLED=1 go test -race ./...

# Build verification
cd backend && go build ./cmd/server
```

### Unified Test Entry Point
```bash
# All test layers (unit → integration → e2e → bdd)
.\scripts\testing\test.ps1 all          # Windows
./scripts/testing/test.sh all            # Linux/Mac

# Individual targets
.\scripts\testing\test.ps1 unit         # backend + frontend unit
.\scripts\testing\test.ps1 integration  # backend integration tests
.\scripts\testing\test.ps1 e2e          # Playwright E2E
.\scripts\testing\test.ps1 bdd          # Cucumber BDD
.\scripts\testing\test.ps1 coverage     # backend + frontend coverage
.\scripts\testing\test.ps1 clean        # cleanup temporary artifacts
```

### NLP Service Testing
```bash
# Unit tests
cd nlp-service && pytest tests/ -v

# Health check
curl http://localhost:5000/health

# API tests
curl -X POST http://localhost:5000/extract_keywords -H "Content-Type: application/json" -d '{"text":"test","top_n":3}'
curl -X POST http://localhost:5000/embed -H "Content-Type: application/json" -d '{"text":"test"}'
```

### Test Stack Management
```bash
# Full test cycle (isolated model - stops dev/personal stacks)
.\scripts\testing\run-full-test-cycle.ps1      # Windows
./scripts/testing/run-full-test-cycle.sh       # Linux/Mac

# Start test stack
.\scripts\testing\start-test.ps1              # Windows
./scripts/testing/start-test.sh               # Linux/Mac

# Seed test data
.\scripts\testing\seed-test-data.ps1          # Windows
./scripts\testing\seed-test-data.sh           # Linux/Mac

# Stop and destroy test stack
.\scripts\testing\stop-test.ps1               # Windows
./scripts\testing\stop-test.sh                # Linux/Mac

# Manual test stack management
docker compose -f docker-compose.test.yml up -d --build
docker compose -f docker-compose.test.yml down -v
```

### Stack Health Checks
```bash
# Check all stacks (default)
.\scripts\ci\check-stacks-health.ps1              # Windows
./scripts/ci/check-stacks-health.sh               # Linux/Mac

# Check specific stack
.\scripts\ci\check-stacks-health.ps1 -Stack dev   # Windows
.\scripts\ci\check-stacks-health.ps1 -Stack personal
.\scripts\ci\check-stacks-health.ps1 -Stack test
./scripts/ci/check-stacks-health.sh --stack dev    # Linux/Mac
./scripts/ci/check-stacks-health.sh --stack personal
./scripts/ci/check-stacks-health.sh --stack test
```

### Regression Testing
```bash
# Full regression cycle (isolated model - 25 steps)
.\scripts\testing\run-full-test-cycle.ps1      # Windows
./scripts/testing/run-full-test-cycle.sh       # Linux/Mac

# Stacks identity check
.\scripts\ci\check-stacks-identity.ps1    # Windows
./scripts/ci/check-stacks-identity.sh     # Linux/Mac

# Cleanup test artifacts
python .\scripts\cleanup\cleanup-test-artifacts.py    # Windows
python ./scripts/cleanup/cleanup-test-artifacts.py     # Linux/Mac

# Individual regression steps
# Step 0: Capture dev stack state snapshot
# Step 1-2: Stop dev and personal stacks
# Step 3: Check stacks identity
# Step 4: Start test stack
# Step 5: Seed test data
# Step 6-17: Run tests and verifications (unit, integration, API, manual)
# Step 18: Documentation verification
# Step 19: Stop test stack
# Step 20: Cleanup temporary files
# Step 21-22: Restore dev and personal stacks
# Step 23: Compare states and verify identity + health
# Step 24: Auto-commit if all checks pass
```

### Database Verification
```bash
# PostgreSQL (test stack)
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT COUNT(*) FROM note_embeddings;"

# Redis (test stack)
docker exec kg-test-redis redis-cli PING
docker exec kg-test-redis redis-cli KEYS "*"

# MongoDB (test stack)
docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')"
docker exec kg-test-mongo mongosh --eval "db.getCollectionNames()"
```

### Health Checks
```bash
# Dev stack
curl http://localhost:8080/health           # Nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:8080/api/v1/notes     # Notes API

# Personal stack
curl http://localhost:8085/health           # Personal backend
curl http://localhost:8082/health           # Personal API gateway
curl http://localhost:8092/health           # Personal graph service

# Test stack
curl http://localhost:8083/health           # Test backend
curl http://localhost:3002                  # Test frontend
curl http://localhost:15002/health          # Test NLP service
```

### Testing Best Practices
- **ALWAYS** use isolated test stack for E2E and BDD testing
- **NEVER** run E2E/BDD tests against dev or personal stacks
- **ALWAYS** verify stacks identity before regression testing
- **ALWAYS** destroy test stack with `down -v` after testing
- **ALWAYS** use i18n keys or `data-testid` selectors in frontend tests to avoid brittle locale-specific text
- **ALWAYS** run unit tests before integration tests
- **ALWAYS** verify health endpoints before API testing
- **ALWAYS** use the full test cycle script for regression testing (isolated model)
- **ALWAYS** check dev stack state before/after testing for data leakage
- **ALWAYS** verify dev/personal identity after testing

---
