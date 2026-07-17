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
.\scripts\run-full-test-cycle.ps1      # Windows
./scripts/run-full-test-cycle.sh       # Linux/Mac

# Start test stack
.\scripts\start-test.ps1              # Windows
./scripts\start-test.sh               # Linux/Mac

# Seed test data
.\scripts\seed-test-data.ps1          # Windows
./scripts\seed-test-data.sh           # Linux/Mac

# Stop and destroy test stack
.\scripts\stop-test.ps1               # Windows
./scripts\stop-test.sh                # Linux/Mac

# Manual test stack management
docker compose -f docker-compose.test.yml up -d --build
docker compose -f docker-compose.test.yml down -v
```

### Stack Health Checks
```bash
# Check all stacks (default)
.\scripts\check-stacks-health.ps1              # Windows
./scripts/check-stacks-health.sh               # Linux/Mac

# Check specific stack
.\scripts\check-stacks-health.ps1 -Stack dev   # Windows
.\scripts\check-stacks-health.ps1 -Stack personal
.\scripts\check-stacks-health.ps1 -Stack test
./scripts/check-stacks-health.sh --stack dev    # Linux/Mac
./scripts/check-stacks-health.sh --stack personal
./scripts/check-stacks-health.sh --stack test
```

### Regression Testing
```bash
# Full regression cycle (isolated model - 24 steps)
.\scripts\run-full-test-cycle.ps1      # Windows
./scripts/run-full-test-cycle.sh       # Linux/Mac

# Stacks identity check
.\scripts\check-stacks-identity.ps1    # Windows
./scripts\check-stacks-identity.sh     # Linux/Mac

# Individual regression steps
# Step 0: Capture dev stack state snapshot
# Step 1-2: Stop dev and personal stacks
# Step 3: Check stacks identity
# Step 4: Start test stack
# Step 5: Seed test data
# Step 6-16: Run tests and verifications
# Step 17: Stop test stack
# Step 18-19: Restore dev and personal stacks
# Step 20-22: Compare states and verify identity
# Step 23: Auto-commit if all checks pass
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
- **ALWAYS** use English text patterns in frontend tests (language policy)
- **ALWAYS** run unit tests before integration tests
- **ALWAYS** verify health endpoints before API testing
- **ALWAYS** use the full test cycle script for regression testing (isolated model)
- **ALWAYS** check dev stack state before/after testing for data leakage
- **ALWAYS** verify dev/personal identity after testing

---
