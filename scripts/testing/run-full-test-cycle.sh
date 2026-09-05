#!/bin/bash
# Full Test Cycle - Linux/Mac (Isolated Testing Model)
# This script orchestrates the complete testing cycle with full stack isolation.
# Dev and personal stacks are stopped during testing to prevent resource conflicts.
# All temporary snapshots are saved to scripts/testing/temp/snapshots/.

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Shared phase tracking lives in lib/phase-tracking.sh so this script and the
# regression test exercise the same code path.
. "$SCRIPT_DIR/lib/phase-tracking.sh"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
SNAPSHOT_BASE="$SCRIPT_DIR/temp/snapshots"
SNAPSHOT_DIR="$SNAPSHOT_BASE/$TIMESTAMP"
mkdir -p "$SNAPSHOT_DIR"

restore_stacks() {
    echo ""
    echo "Restoring dev and personal stacks..."
    (
        cd "$PROJECT_ROOT"
        # Build dev/personal images only if they are missing (e.g., after full cleanup)
        if docker image inspect knowledge-graph-backend:latest >/dev/null 2>&1; then
            docker compose up -d --wait || true
        else
            docker compose up -d --build --wait || true
        fi
        if docker image inspect knowledge-graph-backend_personal:latest >/dev/null 2>&1; then
            docker compose -f docker-compose.personal.yml up -d --wait || true
        else
            docker compose -f docker-compose.personal.yml up -d --build --wait || true
        fi
    )
    echo "  ✓ Dev and personal stacks restored"
}

summary_printed=0
on_exit() {
    restore_stacks
    if [[ $summary_printed -eq 0 ]]; then
        write_final_summary false
    fi
}
trap on_exit EXIT

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
echo "[Step 0/24] Capturing dev stack state snapshot..."
docker ps --filter "name=kg-" > "$SNAPSHOT_DIR/pre-test-ps.txt"
echo "  ✓ Container snapshot saved to $SNAPSHOT_DIR/pre-test-ps.txt"

if curl -s -f http://127.0.0.1:18080/health > /dev/null; then
    curl -s http://127.0.0.1:18080/health > "$SNAPSHOT_DIR/pre-test-health.json"
    echo "  ✓ Health snapshot saved to $SNAPSHOT_DIR/pre-test-health.json"
else
    echo "  ⚠ Dev health endpoint not available (stack may be stopped)"
fi

if curl -s -f "http://127.0.0.1:18080/api/v1/notes?limit=1" > /dev/null; then
    curl -s "http://127.0.0.1:18080/api/v1/notes?limit=1" > "$SNAPSHOT_DIR/pre-test-notes.json"
    echo "  ✓ Notes snapshot saved to $SNAPSHOT_DIR/pre-test-notes.json"
else
    echo "  ⚠ Dev API not available (stack may be stopped)"
fi

# Step 1: Stop dev stack
echo ""
echo "[Step 1/24] Stopping dev stack..."
docker compose down
stop_dev_exit=$?
register_phase "Stop dev stack" "$stop_dev_exit"
if [[ $stop_dev_exit -eq 0 ]]; then
    echo "  ✓ Dev stack stopped"
fi

# Step 2: Stop personal stack
echo ""
echo "[Step 2/24] Stopping personal stack..."
docker compose -f docker-compose.personal.yml down
stop_personal_exit=$?
register_phase "Stop personal stack" "$stop_personal_exit"
if [[ $stop_personal_exit -eq 0 ]]; then
    echo "  ✓ Personal stack stopped"
fi

# Step 3: Check stacks identity
echo ""
echo "[Step 3/24] Checking stacks identity..."
"$SCRIPT_DIR/../ci/check-stacks-identity.sh"
identity_exit=$?
register_phase "Check stacks identity" "$identity_exit"
if [[ $identity_exit -ne 0 ]]; then
    echo "ERROR: Stacks have differences"
    echo "Please fix the differences before running tests"
    echo "⚠ Continuing with isolated testing despite identity differences"
else
    echo "✓ Stacks are identical"
fi

# Step 4: Start test stack
echo ""
echo "[Step 4/24] Starting test stack..."
"$SCRIPT_DIR/start-test.sh"
start_test_exit=$?
register_phase "Start test stack" "$start_test_exit"
if [[ $start_test_exit -ne 0 ]]; then
    echo "ERROR: Failed to start test stack"
    exit 1
fi
echo "✓ Test stack started"

# Step 5: Seed test data
echo ""
echo "[Step 5/24] Seeding test data..."
"$SCRIPT_DIR/seed-test-data.sh"
seed_exit=$?
register_phase "Seed test data" "$seed_exit"
if [[ $seed_exit -ne 0 ]]; then
    echo "WARNING: Failed to seed test data"
    echo "Continuing anyway (data might already exist)"
else
    echo "✓ Test data seeded"
fi

# Step 6: Docker Build Verification
echo ""
echo "[Step 6/24] Docker Build Verification..."
echo "  Checking Docker images..."
docker images --format "{{.Repository}}:{{.Tag}}" | grep '^knowledge-graph' || true
echo "  ✓ Docker images checked"

# Step 7: NLP Service Tests
echo ""
echo "[Step 7/24] NLP Service Tests..."
if curl -s -f http://127.0.0.1:15002/health > /dev/null; then
    echo "  ✓ NLP health endpoint: OK"
else
    echo "  ⚠ NLP health endpoint: FAILED"
fi

# Step 8: Backend Unit Tests
echo ""
echo "[Step 8/24] Backend Unit Tests..."
echo "  Running backend unit tests..."
if cd "$PROJECT_ROOT/backend"; then
    go test ./... -count=1
    backend_unit_exit=$?
    cd "$PROJECT_ROOT"
    register_phase "Backend unit tests" "$backend_unit_exit"
    if [[ $backend_unit_exit -eq 0 ]]; then
        echo "  ✓ Backend unit tests completed"
    fi
else
    register_phase "Backend unit tests" 1
    echo "  ERROR: Failed to enter backend directory"
fi

# Step 9: Backend Integration Tests
echo ""
echo "[Step 9/24] Backend Integration Tests..."
echo "  Running backend integration tests (requires Linux/WSL Docker)..."
if cd "$PROJECT_ROOT/backend"; then
    go test -tags=integration -p=1 -count=1 ./...
    backend_integration_exit=$?
    cd "$PROJECT_ROOT"
    register_phase "Backend integration tests" "$backend_integration_exit"
    if [[ $backend_integration_exit -eq 0 ]]; then
        echo "  ✓ Backend integration tests completed"
    else
        echo "  WARNING: Backend integration tests failed"
        echo "  This is often testcontainers on Windows rootless Docker; use WSL2 or CI."
    fi
else
    register_phase "Backend integration tests" 1
    echo "  ERROR: Failed to enter backend directory"
fi

# Step 10: Backend API Verification
echo ""
echo "[Step 10/24] Backend API Verification..."
if curl -s -f http://127.0.0.1:18083/health > /dev/null; then
    echo "  ✓ Test backend health: OK"
else
    echo "  ⚠ Test backend health: FAILED"
fi

if curl -s -f "http://127.0.0.1:18083/api/v1/notes?limit=1" > /dev/null; then
    echo "  ✓ Test backend API: OK"
else
    echo "  ⚠ Test backend API: FAILED"
fi

# Step 11: Asynchronous Tasks Verification
echo ""
echo "[Step 11/24] Asynchronous Tasks Verification..."
docker logs kg-test-worker --tail 10 || echo "  ⚠ Worker logs not available"
echo "  ✓ Worker logs checked"

# Step 12: PGVECTOR Verification
echo ""
echo "[Step 12/24] PGVECTOR Verification..."
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
pgvector_exit=$?
register_phase "PGVECTOR verification" "$pgvector_exit"
if [[ $pgvector_exit -eq 0 ]]; then
    echo "  ✓ PGVECTOR extension checked"
fi

# Step 13: Redis & MongoDB Verification
echo ""
echo "[Step 13/24] Redis & MongoDB Verification..."
docker exec kg-test-redis redis-cli PING
redis_exit=$?
docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')"
mongo_exit=$?
redis_mongo_exit=$(( redis_exit != 0 ? redis_exit : mongo_exit ))
register_phase "Redis and MongoDB verification" "$redis_mongo_exit"
if [[ $redis_mongo_exit -eq 0 ]]; then
    echo "  ✓ Redis and MongoDB checked"
fi

# Step 14: Frontend Unit Tests
echo ""
echo "[Step 14/24] Frontend Unit Tests..."
echo "  Running frontend unit tests..."
if cd "$PROJECT_ROOT/frontend"; then
    npm run test:unit
    frontend_unit_exit=$?
    cd "$PROJECT_ROOT"
    register_phase "Frontend unit tests" "$frontend_unit_exit"
    if [[ $frontend_unit_exit -eq 0 ]]; then
        echo "  ✓ Frontend unit tests completed"
    fi
else
    register_phase "Frontend unit tests" 1
    echo "  ERROR: Failed to enter frontend directory"
fi

# Step 15: Manual testing instructions
echo ""
echo "[Step 15/24] Test environment ready for manual testing"
echo "========================================"
echo "  MANUAL TESTING INSTRUCTIONS"
echo "========================================"
echo ""
echo "Test stack URLs:"
echo "  Frontend: http://127.0.0.1:3002"
echo "  Backend API: http://127.0.0.1:18083"
echo ""
echo "Follow the manual test checklist:"
echo "  docs/MANUAL_TEST_CHECKLISTS_RU.md"
echo ""
echo "Test user credentials:"
echo "  Login: testuser"
echo "  Password: TestPassword123!"
echo ""
MANUAL_TEST_FLAG="$SNAPSHOT_DIR/continue-manual-test.flag"
echo ""
echo "  Manual testing in progress..."
echo "  Press Enter in the interactive terminal, or create the flag file:"
echo "  $MANUAL_TEST_FLAG"
echo "  The script will continue automatically when the flag is created."
while [ ! -f "$MANUAL_TEST_FLAG" ]; do
    # Try to read a single Enter keypress non-blocking (interactive terminals)
    if IFS= read -rs -t 1 -n 1 key && [ "$key" = "" ]; then
        break
    fi
    sleep 1
done
rm -f "$MANUAL_TEST_FLAG"

# Step 16: Public Graph Verification
echo ""
echo "[Step 16/24] Public Graph Verification..."
echo "  ⏳ Manual verification required"
echo "  Please verify public graph functionality manually"

# Step 17: CI/CD Verification
echo ""
echo "[Step 17/24] CI/CD Verification..."
echo "  ⏳ Manual verification required"
echo "  Please verify CI/CD workflows manually"

# Step 18: Documentation Verification
echo ""
echo "[Step 18/24] Documentation Verification..."
for f in docs/AGENTS.md .windsurfrules; do
    if [ -f "$PROJECT_ROOT/$f" ]; then
        echo "  ✓ $f exists"
    else
        echo "  ⚠ $f missing"
    fi
done
dirty=$(git diff --name-only 2>/dev/null | grep -E "^(docs/AGENTS\.md|\.windsurfrules|internal/(domain|infrastructure|application|interfaces))" || true)
if [ -n "$dirty" ]; then
    echo "  ⚠ Architecture files changed; verify docs/AGENTS.md and .windsurfrules are updated:"
    echo "$dirty" | sed 's/^/    - /'
else
    echo "  ✓ No architecture boundary changes detected"
fi

# Step 19: Stop test stack
echo ""
echo "[Step 19/24] Stopping test stack..."
"$SCRIPT_DIR/stop-test.sh"
stop_test_exit=$?
register_phase "Stop test stack" "$stop_test_exit"
if [[ $stop_test_exit -eq 0 ]]; then
    echo "✓ Test stack destroyed"
else
    echo "WARNING: Failed to stop test stack"
fi

# Step 20: Cleanup Temporary Files
echo ""
echo "[Step 20/24] Cleanup Temporary Files..."
python "$SCRIPT_DIR/../cleanup/cleanup-test-artifacts.py"
cleanup_exit=$?
register_phase "Cleanup temporary files" "$cleanup_exit"
if [[ $cleanup_exit -eq 0 ]]; then
    echo "  ✓ Temporary files cleaned"
fi

# Step 21: Start dev stack
echo ""
echo "[Step 21/24] Starting dev stack..."
if docker image inspect knowledge-graph-backend:latest >/dev/null 2>&1; then
    docker compose up -d --wait
else
    docker compose up -d --build --wait
fi
dev_restore_exit=$?
register_phase "Restore dev stack" "$dev_restore_exit"
if [[ $dev_restore_exit -eq 0 ]]; then
    echo "  ✓ Dev stack started"
fi

# Step 22: Start personal stack
echo ""
echo "[Step 22/24] Starting personal stack..."
if docker image inspect knowledge-graph-backend_personal:latest >/dev/null 2>&1; then
    docker compose -f docker-compose.personal.yml up -d --wait
else
    docker compose -f docker-compose.personal.yml up -d --build --wait
fi
personal_restore_exit=$?
register_phase "Restore personal stack" "$personal_restore_exit"
if [[ $personal_restore_exit -eq 0 ]]; then
    echo "  ✓ Personal stack started"
fi

# Step 23: State, identity and health checks
echo ""
echo "[Step 23/24] State, identity and health checks"
docker ps --filter "name=kg-" > "$SNAPSHOT_DIR/post-test-ps.txt"
echo "  ✓ Post-test container snapshot saved"

curl -s -f http://127.0.0.1:18080/health > /dev/null
dev_health_exit=$?
if [[ $dev_health_exit -eq 0 ]]; then
    curl -s http://127.0.0.1:18080/health > "$SNAPSHOT_DIR/post-test-health.json"
    echo "  ✓ Post-test health snapshot saved"
else
    echo "  ⚠ Dev health endpoint not available after restoration"
    echo "  ERROR: Dev stack restoration failed"
    register_phase "Dev stack restoration (health endpoint)" "$dev_health_exit"
    exit 1
fi

curl -s -f "http://127.0.0.1:18080/api/v1/notes?limit=1" > /dev/null
dev_api_exit=$?
if [[ $dev_api_exit -eq 0 ]]; then
    curl -s "http://127.0.0.1:18080/api/v1/notes?limit=1" > "$SNAPSHOT_DIR/post-test-notes.json"
    echo "  ✓ Post-test notes snapshot saved"
else
    echo "  ⚠ Dev API not available after restoration"
    echo "  ERROR: Dev stack restoration failed"
    register_phase "Dev stack restoration (API endpoint)" "$dev_api_exit"
    exit 1
fi

dev_state_changed=false

if ! diff -q "$SNAPSHOT_DIR/pre-test-ps.txt" "$SNAPSHOT_DIR/post-test-ps.txt" > /dev/null; then
    echo "  ⚠ Dev container state changed during testing"
    dev_state_changed=true
else
    echo "  ✓ Dev container state unchanged"
fi

if [ -f "$SNAPSHOT_DIR/pre-test-health.json" ] && [ -f "$SNAPSHOT_DIR/post-test-health.json" ]; then
    if ! diff -q "$SNAPSHOT_DIR/pre-test-health.json" "$SNAPSHOT_DIR/post-test-health.json" > /dev/null; then
        echo "  ⚠ Dev health endpoint changed during testing"
        dev_state_changed=true
    else
        echo "  ✓ Dev health endpoint unchanged"
    fi
fi

if ! diff -q "$SNAPSHOT_DIR/pre-test-notes.json" "$SNAPSHOT_DIR/post-test-notes.json" > /dev/null; then
    echo "  ⚠ Dev API response changed during testing"
    dev_state_changed=true
else
    echo "  ✓ Dev API response unchanged"
fi

if [ "$dev_state_changed" = true ]; then
    register_phase "Dev stack state verification" 1
    echo "  WARNING: Dev stack state changed during testing"
    echo "  This may indicate data leakage or side effects"
else
    register_phase "Dev stack state verification" 0
    echo "  ✓ Dev stack state verified - no changes detected"
fi

curl -s -f "http://127.0.0.1:18080/api/v1/notes?limit=1" > /dev/null
dev_notes_exit=$?
if [[ $dev_notes_exit -eq 0 ]]; then
    curl -s "http://127.0.0.1:18080/api/v1/notes?limit=1" > "$SNAPSHOT_DIR/dev-notes.json"
    echo "  ✓ Dev notes snapshot saved"
else
    echo "  ERROR: Failed to get dev notes"
    register_phase "Dev notes snapshot" "$dev_notes_exit"
    exit 1
fi

curl -s -f "http://127.0.0.1:18082/api/v1/notes?limit=1" > /dev/null
personal_notes_exit=$?
if [[ $personal_notes_exit -eq 0 ]]; then
    curl -s "http://127.0.0.1:18082/api/v1/notes?limit=1" > "$SNAPSHOT_DIR/personal-notes.json"
    echo "  ✓ Personal notes snapshot saved"
else
    echo "  ERROR: Failed to get personal notes"
    register_phase "Personal notes snapshot" "$personal_notes_exit"
    exit 1
fi

diff -q "$SNAPSHOT_DIR/dev-notes.json" "$SNAPSHOT_DIR/personal-notes.json" > /dev/null
stacks_identity_exit=$?
register_phase "Stacks identity (dev vs personal)" "$stacks_identity_exit"
if [[ $stacks_identity_exit -ne 0 ]]; then
    echo "  ⚠ Dev and Personal stacks are NOT identical"
    echo "  ERROR: Stacks have differences - manual investigation required"
    echo "  Difference details:"
    diff "$SNAPSHOT_DIR/dev-notes.json" "$SNAPSHOT_DIR/personal-notes.json"
    echo ""
    echo "  Please investigate and fix the differences manually"
    exit 1
else
    echo "  ✓ Dev and Personal stacks are identical"
fi

"$SCRIPT_DIR/../ci/check-stacks-health.sh" --stack dev
dev_health_check_exit=$?
register_phase "Dev stack health check" "$dev_health_check_exit"
if [[ $dev_health_check_exit -ne 0 ]]; then
    echo "  ERROR: Dev stack is not healthy"
    exit 1
fi

"$SCRIPT_DIR/../ci/check-stacks-health.sh" --stack personal
personal_health_check_exit=$?
register_phase "Personal stack health check" "$personal_health_check_exit"
if [[ $personal_health_check_exit -ne 0 ]]; then
    echo "  ERROR: Personal stack is not healthy"
    exit 1
fi
echo "  ✓ Dev and personal stacks health verified"

# Final Summary:
summary_printed=1
if test_any_failed; then
    write_final_summary false
    exit 1
else
    write_final_summary true
fi
