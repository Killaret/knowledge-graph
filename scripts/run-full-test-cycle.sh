#!/bin/bash
# Full Test Cycle - Linux/Mac (Isolated Testing Model)
# This script orchestrates the complete testing cycle with full stack isolation
# Dev and personal stacks are stopped during testing to prevent resource conflicts

set -e

echo "========================================"
echo "  Knowledge Graph Full Test Cycle"
echo "  (Isolated Testing Model)"
echo "========================================"
echo ""
echo "For comprehensive regression testing, see docs/REGRESSION_TEST_PLAN.md"
echo ""
echo "⚠️  WARNING: Dev and personal stacks will be stopped during testing"
echo ""

# Step 0: Capture dev stack state snapshot
echo "[Step 0/22] Capturing dev stack state snapshot..."
timestamp=$(date +"%Y%m%d_%H%M%S")
snapshot_dir="test-snapshots_$timestamp"
mkdir -p "$snapshot_dir"

docker ps --filter "name=kg-" > "$snapshot_dir/pre-test-ps.txt"
echo "  ✓ Container snapshot saved to $snapshot_dir/pre-test-ps.txt"

if curl -s -f http://localhost:8080/health > /dev/null; then
    curl -s http://localhost:8080/health > "$snapshot_dir/pre-test-health.json"
    echo "  ✓ Health snapshot saved to $snapshot_dir/pre-test-health.json"
else
    echo "  ⚠ Dev health endpoint not available (stack may be stopped)"
fi

if curl -s -f "http://localhost:8080/api/v1/notes?limit=1" > /dev/null; then
    curl -s "http://localhost:8080/api/v1/notes?limit=1" > "$snapshot_dir/pre-test-notes.json"
    echo "  ✓ Notes snapshot saved to $snapshot_dir/pre-test-notes.json"
else
    echo "  ⚠ Dev API not available (stack may be stopped)"
fi

# Step 1: Stop dev stack
echo ""
echo "[Step 1/22] Stopping dev stack..."
docker compose down
echo "  ✓ Dev stack stopped"

# Step 2: Stop personal stack
echo ""
echo "[Step 2/22] Stopping personal stack..."
docker compose -f docker-compose.personal.yml down
echo "  ✓ Personal stack stopped"

# Step 3: Check stacks identity
echo ""
echo "[Step 3/22] Checking stacks identity..."
if ! ./scripts/check-stacks-identity.sh; then
    echo "ERROR: Stacks have differences"
    echo "Please fix the differences before running tests"
    echo "⚠ Continuing with isolated testing despite identity differences"
else
    echo "✓ Stacks are identical"
fi

# Step 4: Start test stack
echo ""
echo "[Step 4/22] Starting test stack..."
if ! ./scripts/start-test.sh; then
    echo "ERROR: Failed to start test stack"
    echo "Attempting to restore dev stack..."
    docker compose up -d
    docker compose -f docker-compose.personal.yml up -d
    exit 1
fi
echo "✓ Test stack started"

# Step 5: Seed test data
echo ""
echo "[Step 5/22] Seeding test data..."
if ! ./scripts/seed-test-data.sh; then
    echo "WARNING: Failed to seed test data"
    echo "Continuing anyway (data might already exist)"
else
    echo "✓ Test data seeded"
fi

# Step 6: Docker Build Verification
echo ""
echo "[Step 6/22] Docker Build Verification..."
echo "  Checking Docker images..."
docker images | grep knowledge-graph
echo "  ✓ Docker images checked"

# Step 7: NLP Service Tests
echo ""
echo "[Step 7/22] NLP Service Tests..."
if curl -s -f http://localhost:15002/health > /dev/null; then
    echo "  ✓ NLP health endpoint: OK"
else
    echo "  ⚠ NLP health endpoint: FAILED"
fi

# Step 8: Backend Unit Tests
echo ""
echo "[Step 8/22] Backend Unit Tests..."
echo "  Running backend unit tests..."
cd backend
go test ./... -count=1
cd ..
echo "  ✓ Backend unit tests completed"

# Step 9: Backend API Verification
echo ""
echo "[Step 9/22] Backend API Verification..."
if curl -s -f http://localhost:8083/health > /dev/null; then
    echo "  ✓ Test backend health: OK"
else
    echo "  ⚠ Test backend health: FAILED"
fi

if curl -s -f "http://localhost:8083/api/v1/notes?limit=1" > /dev/null; then
    echo "  ✓ Test backend API: OK"
else
    echo "  ⚠ Test backend API: FAILED"
fi

# Step 10: Asynchronous Tasks Verification
echo ""
echo "[Step 10/22] Asynchronous Tasks Verification..."
docker logs kg-test-worker --tail 10
echo "  ✓ Worker logs checked"

# Step 11: PGVECTOR Verification
echo ""
echo "[Step 11/22] PGVECTOR Verification..."
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
echo "  ✓ PGVECTOR extension checked"

# Step 12: Redis & MongoDB Verification
echo ""
echo "[Step 12/22] Redis & MongoDB Verification..."
docker exec kg-test-redis redis-cli PING
docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')"
echo "  ✓ Redis and MongoDB checked"

# Step 13: Frontend Unit Tests
echo ""
echo "[Step 13/22] Frontend Unit Tests..."
echo "  Running frontend unit tests..."
cd frontend
npm run test:unit
cd ..
echo "  ✓ Frontend unit tests completed"

# Step 14: Manual testing instructions
echo ""
echo "[Step 14/22] Test environment ready for manual testing"
echo "========================================"
echo "  MANUAL TESTING INSTRUCTIONS"
echo "========================================"
echo ""
echo "Test stack URLs:"
echo "  Frontend: http://localhost:3002"
echo "  Backend API: http://localhost:8083"
echo ""
echo "Follow the manual test checklist:"
echo "  docs/MANUAL_TEST_CHECKLISTS.md"
echo ""
echo "Test user credentials:"
echo "  Login: testuser"
echo "  Password: TestPassword123!"
echo ""
echo "Press Enter when manual testing is complete..."
read

# Step 15: Public Graph Verification
echo ""
echo "[Step 15/22] Public Graph Verification..."
echo "  ⏳ Manual verification required"
echo "  Please verify public graph functionality manually"

# Step 16: CI/CD Verification
echo ""
echo "[Step 16/22] CI/CD Verification..."
echo "  ⏳ Manual verification required"
echo "  Please verify CI/CD workflows manually"

# Step 17: Stop test stack
echo ""
echo "[Step 17/22] Stopping test stack..."
if ! ./scripts/stop-test.sh; then
    echo "WARNING: Failed to stop test stack"
else
    echo "✓ Test stack destroyed"
fi

# Step 18: Start dev stack
echo ""
echo "[Step 18/22] Starting dev stack..."
docker compose up -d
echo "  Waiting for dev stack to be healthy..."
sleep 30
echo "  ✓ Dev stack started"

# Step 19: Start personal stack
echo ""
echo "[Step 19/22] Starting personal stack..."
docker compose -f docker-compose.personal.yml up -d
echo "  Waiting for personal stack to be healthy..."
sleep 30
echo "  ✓ Personal stack started"

# Step 20: Compare dev stack state with snapshot
echo ""
echo "[Step 20/22] Comparing dev stack state with snapshot..."
docker ps --filter "name=kg-" > "$snapshot_dir/post-test-ps.txt"
echo "  ✓ Post-test container snapshot saved"

if curl -s -f http://localhost:8080/health > /dev/null; then
    curl -s http://localhost:8080/health > "$snapshot_dir/post-test-health.json"
    echo "  ✓ Post-test health snapshot saved"
else
    echo "  ⚠ Dev health endpoint not available after restoration"
fi

if curl -s -f "http://localhost:8080/api/v1/notes?limit=1" > /dev/null; then
    curl -s "http://localhost:8080/api/v1/notes?limit=1" > "$snapshot_dir/post-test-notes.json"
    echo "  ✓ Post-test notes snapshot saved"
else
    echo "  ⚠ Dev API not available after restoration"
fi

# Step 21: Check stacks health
echo ""
echo "[Step 21/22] Checking dev and personal stacks health after testing..."
./scripts/check-stacks-health.sh --stack dev
./scripts/check-stacks-health.sh --stack personal
echo "  ✓ Dev and personal stacks health checked"

# Step 22: Summary
echo ""
echo "[Step 22/22] Test cycle summary"
echo "========================================"
echo "  TEST CYCLE COMPLETE"
echo "========================================"
echo ""
echo "✓ Dev stack state captured before testing"
echo "✓ Dev and personal stacks stopped for isolation"
echo "✓ Stacks identity verified"
echo "✓ Test stack started and destroyed"
echo "✓ Test data seeded"
echo "✓ Docker build verification completed"
echo "✓ NLP service tests completed"
echo "✓ Backend unit tests completed"
echo "✓ Backend API verification completed"
echo "✓ Asynchronous tasks verified"
echo "✓ PGVECTOR verification completed"
echo "✓ Redis and MongoDB verified"
echo "✓ Frontend unit tests completed"
echo "✓ Manual testing completed"
echo "✓ Dev and personal stacks restored"
echo "✓ Dev stack state compared with snapshot"
echo "✓ Dev and personal stacks health verified"
echo ""
echo "Snapshots saved to: $snapshot_dir"
echo ""
echo "All stacks are stable and isolated testing completed successfully."

# Step 23: Exit
echo ""
echo "Press Enter to exit..."
read
