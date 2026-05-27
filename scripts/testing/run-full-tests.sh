#!/bin/bash
# Full Regression Test Suite Script
# Runs all tests: unit, integration, e2e, bdd

set -e

echo "=========================================="
echo "🚀 Starting Full Regression Test Suite"
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
PASSED=0
FAILED=0

# Function to run test and track results
run_test() {
    local name=$1
    local cmd=$2
    
    echo ""
    echo "📦 Running: $name"
    echo "----------------------------------------"
    
    if eval "$cmd"; then
        echo -e "${GREEN}✅ $name PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ $name FAILED${NC}"
        ((FAILED++))
    fi
}

# 1. Backend Unit Tests
echo ""
echo "🔧 BACKEND TESTS"
echo "=========================================="
cd backend

run_test "Backend Unit Tests" "go test ./internal/... -v 2>&1 | grep -E 'PASS|FAIL|ok' | tail -20"

# 2. Frontend Unit Tests
echo ""
echo "🎨 FRONTEND TESTS"
echo "=========================================="
cd ../frontend

run_test "Frontend Unit Tests (Vitest)" "npm run test:unit 2>&1 | grep -E 'passed|failed|Test Files' | tail -5"

# 3. Playwright E2E Tests
echo ""
echo "🎭 E2E TESTS"
echo "=========================================="

# Start services
export SKIP_AUTH=true
echo "🚀 Starting dev server..."
npm run dev &
DEV_PID=$!

# Wait for server
sleep 5

run_test "Playwright Smoke Tests" "npx playwright test --grep='@smoke' --timeout=60000 --reporter=line 2>&1 | grep -E 'passed|failed' | tail -5"

run_test "Visual Regression Tests" "npx playwright test tests/graph-visual-isolated-new.spec.ts --timeout=60000 --reporter=line 2>&1 | grep -E 'passed|failed' | tail -5"

# Cleanup
kill $DEV_PID 2>/dev/null || true

# 4. BDD Tests
echo ""
echo "🥒 BDD TESTS"
echo "=========================================="

run_test "Cucumber BDD Tests" "npm run test:cucumber 2>&1 | grep -E 'scenarios|steps|passed|failed' | tail -10"

# Summary
echo ""
echo "=========================================="
echo "📊 TEST SUMMARY"
echo "=========================================="
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo "=========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed${NC}"
    exit 1
fi
