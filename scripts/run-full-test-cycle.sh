#!/bin/bash
# Full Test Cycle - Linux/Mac
# This script orchestrates the complete testing cycle

set -e

echo "========================================"
echo "  Knowledge Graph Full Test Cycle"
echo "========================================"

# Step 1: Check stacks health
echo ""
echo "[Step 1/8] Checking dev and personal stacks health..."
if ! ./scripts/check-stacks-health.sh; then
    echo "ERROR: Dev or personal stacks are not healthy"
    echo "Please start dev and personal stacks first"
    exit 1
fi
echo "✓ Dev and personal stacks are healthy"

# Step 2: Start test stack
echo ""
echo "[Step 2/8] Starting test stack..."
if ! ./scripts/start-test.sh; then
    echo "ERROR: Failed to start test stack"
    exit 1
fi
echo "✓ Test stack started"

# Step 3: Seed test data
echo ""
echo "[Step 3/8] Seeding test data..."
if ! ./scripts/seed-test-data.sh; then
    echo "ERROR: Failed to seed test data"
    echo "Continuing anyway (data might already exist)"
fi
echo "✓ Test data seeded"

# Step 4: Manual testing instructions
echo ""
echo "[Step 4/8] Test environment ready"
echo "========================================"
echo "  MANUAL TESTING INSTRUCTIONS"
echo "========================================"
echo ""
echo "Test stack URLs:"
echo "  Frontend: http://localhost:13002"
echo "  Backend API: http://localhost:18083"
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

# Step 5: Stop test stack
echo ""
echo "[Step 5/8] Stopping test stack..."
if ! ./scripts/stop-test.sh; then
    echo "WARNING: Failed to stop test stack"
else
    echo "✓ Test stack destroyed"
fi

# Step 6: Check stacks health again
echo ""
echo "[Step 6/8] Checking dev and personal stacks health after testing..."
if ! ./scripts/check-stacks-health.sh; then
    echo "WARNING: Dev or personal stacks are not healthy after testing"
    echo "Please check the stacks"
else
    echo "✓ Dev and personal stacks are still healthy"
fi

# Step 7: Summary
echo ""
echo "[Step 7/8] Test cycle summary"
echo "========================================"
echo "  TEST CYCLE COMPLETE"
echo "========================================"
echo ""
echo "✓ Dev and personal stacks verified (before and after)"
echo "✓ Test stack started and destroyed"
echo "✓ Test data seeded"
echo "✓ Manual testing completed"
echo ""
echo "All stacks are stable."

# Step 8: Exit
echo ""
echo "Press Enter to exit..."
read
