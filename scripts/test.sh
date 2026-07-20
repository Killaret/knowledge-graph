#!/bin/bash
# Unified test entry point for Knowledge Graph
# Usage: ./scripts/test.sh [unit|integration|e2e|bdd|coverage|clean|all]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TARGET="${1:-all}"

run_unit() {
    echo "Running backend unit tests..."
    cd "$PROJECT_ROOT/backend" && go test ./... -count=1

    echo "Running frontend unit tests..."
    cd "$PROJECT_ROOT/frontend" && npm run test:unit
}

run_integration() {
    echo "Running backend integration tests..."
    cd "$PROJECT_ROOT/backend" && go test -tags=integration ./... -count=1 -p=1
}

run_e2e() {
    echo "Running frontend E2E tests..."
    cd "$PROJECT_ROOT/frontend" && npm run test
}

run_bdd() {
    echo "Running frontend BDD tests..."
    cd "$PROJECT_ROOT/frontend" && npm run test:bdd
}

run_coverage() {
    echo "Generating backend coverage..."
    cd "$PROJECT_ROOT/backend" && go test ./... -coverprofile=coverage.out

    echo "Generating frontend coverage..."
    cd "$PROJECT_ROOT/frontend" && npm run test:coverage
}

run_clean() {
    "$SCRIPT_DIR/cleanup-test-artifacts.sh"
}

case "$TARGET" in
    unit) run_unit ;;
    integration) run_integration ;;
    e2e) run_e2e ;;
    bdd) run_bdd ;;
    coverage) run_coverage ;;
    clean) run_clean ;;
    all)
        "$SCRIPT_DIR/test.sh" unit
        "$SCRIPT_DIR/test.sh" integration
        "$SCRIPT_DIR/test.sh" e2e
        "$SCRIPT_DIR/test.sh" bdd
        ;;
    *)
        echo "Usage: $0 {unit|integration|e2e|bdd|coverage|clean|all}"
        exit 1
        ;;
esac
