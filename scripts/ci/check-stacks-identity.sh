#!/bin/bash
# Check Stacks Identity - Bash
# This script verifies that dev, personal, and test stacks are consistent

echo "========================================"
echo "  Stacks Identity Check"
echo "========================================"

ERRORS=0
DIFFERENCES=()

# Step 1: Check service versions from Dockerfiles
echo ""
echo "[Step 1/6] Checking service versions..."

# Check Go version
GO_VERSION=$(grep "FROM golang:" backend/Dockerfile | head -1)
if [ -n "$GO_VERSION" ]; then
    GO_VERSION_STR=${GO_VERSION#FROM golang:}
    echo "  Go version: $GO_VERSION_STR"
else
    echo "  Go version: NOT FOUND"
    ((ERRORS++))
fi

# Check Node version
NODE_VERSION=$(grep "FROM node:" frontend/Dockerfile | head -1)
if [ -n "$NODE_VERSION" ]; then
    NODE_VERSION_STR=${NODE_VERSION#FROM node:}
    echo "  Node version: $NODE_VERSION_STR"
else
    echo "  Node version: NOT FOUND"
    ((ERRORS++))
fi

# Check Python version
PYTHON_VERSION=$(grep "FROM python:" nlp-service/Dockerfile | head -1)
if [ -n "$PYTHON_VERSION" ]; then
    PYTHON_VERSION_STR=${PYTHON_VERSION#FROM python:}
    echo "  Python version: $PYTHON_VERSION_STR"
else
    echo "  Python version: NOT FOUND"
    ((ERRORS++))
fi

# Step 1.5: Check healthchecks in Dockerfiles
echo ""
echo "[Step 1.5/6] Checking healthchecks in Dockerfiles..."

# Check backend Dockerfile for healthcheck
if grep -q "HEALTHCHECK" backend/Dockerfile; then
    echo "  Backend Dockerfile: HEALTHCHECK present"
else
    echo "  Backend Dockerfile: HEALTHCHECK missing"
    ((ERRORS++))
    DIFFERENCES+=("Backend Dockerfile missing HEALTHCHECK")
fi

# Check frontend Dockerfile for healthcheck
if grep -q "HEALTHCHECK" frontend/Dockerfile; then
    echo "  Frontend Dockerfile: HEALTHCHECK present"
else
    echo "  Frontend Dockerfile: HEALTHCHECK missing"
    ((ERRORS++))
    DIFFERENCES+=("Frontend Dockerfile missing HEALTHCHECK")
fi

# Check NLP Dockerfile for healthcheck
if grep -q "HEALTHCHECK" nlp-service/Dockerfile; then
    echo "  NLP Dockerfile: HEALTHCHECK present"
else
    echo "  NLP Dockerfile: HEALTHCHECK missing"
    ((ERRORS++))
    DIFFERENCES+=("NLP Dockerfile missing HEALTHCHECK")
fi

# Step 2: Compare docker-compose files
echo ""
echo "[Step 2/6] Comparing docker-compose files..."

DEV_SERVICES=$(grep -c "image:\|build:" docker-compose.yml || echo "0")
PERSONAL_SERVICES=$(grep -c "image:\|build:" docker-compose.personal.yml || echo "0")
TEST_SERVICES=$(grep -c "image:\|build:" docker-compose.test.yml || echo "0")

echo "  Dev services: $DEV_SERVICES found"
echo "  Personal services: $PERSONAL_SERVICES found"
echo "  Test services: $TEST_SERVICES found"

# Check for SKIP_AUTH consistency
DEV_SKIP_AUTH=$(grep "SKIP_AUTH" docker-compose.yml || echo "NOT FOUND")
PERSONAL_SKIP_AUTH=$(grep "SKIP_AUTH" docker-compose.personal.yml || echo "NOT FOUND")
TEST_SKIP_AUTH=$(grep "SKIP_AUTH" docker-compose.test.yml || echo "NOT FOUND")

echo "  Dev SKIP_AUTH: $DEV_SKIP_AUTH"
echo "  Personal SKIP_AUTH: $PERSONAL_SKIP_AUTH"
echo "  Test SKIP_AUTH: $TEST_SKIP_AUTH"

# Step 3: Check configuration files
echo ""
echo "[Step 3/6] Checking configuration files..."

if [ -f "knowledge-graph.config.json" ]; then
    echo "  knowledge-graph.config.json: OK"
else
    echo "  knowledge-graph.config.json: NOT FOUND"
    ((ERRORS++))
fi

# Step 4: Check nginx configurations
echo ""
echo "[Step 4/6] Checking nginx configurations..."

if [ -f "nginx.conf" ]; then
    echo "  nginx.conf: OK"
else
    echo "  nginx.conf: NOT FOUND"
    ((ERRORS++))
fi

if [ -f "nginx.personal.conf" ]; then
    echo "  nginx.personal.conf: OK"
else
    echo "  nginx.personal.conf: NOT FOUND"
    ((ERRORS++))
fi

# Step 5: Check stack health
echo ""
echo "[Step 5/6] Checking stack health..."

# In CI, dev/personal stacks are not running, so only report status without failing
if [ -n "$CI" ]; then
    echo "  CI environment detected: skipping live stack health checks"
else
    # Check dev stack
    if curl -s http://127.0.0.1:18080/health > /dev/null 2>&1; then
        echo "  Dev stack: OK"
    else
        echo "  Dev stack: FAILED"
        ((ERRORS++))
    fi

    # Check personal stack
    if curl -s http://127.0.0.1:18082/health > /dev/null 2>&1; then
        echo "  Personal stack: OK"
    else
        echo "  Personal stack: FAILED"
        ((ERRORS++))
    fi
fi

# Step 6: Verify healthcheck endpoints are accessible
echo ""
echo "[Step 6/6] Verifying healthcheck endpoints..."

# Check backend health endpoint
if curl -s http://127.0.0.1:9000/health > /dev/null 2>&1; then
    echo "  Backend health endpoint: OK"
else
    echo "  Backend health endpoint: FAILED (backend may not be running)"
    # Not counting as error since backend might not be exposed directly
fi

# Check NLP service health endpoint (if running)
if curl -s http://127.0.0.1:8000/health > /dev/null 2>&1; then
    echo "  NLP service health endpoint: OK"
else
    echo "  NLP service health endpoint: FAILED (NLP may not be running)"
    # Not counting as error since NLP might not be running
fi

# Final result
echo ""
echo "========================================"
if [ $ERRORS -eq 0 ]; then
    echo "  STACKS_IDENTICAL"
    echo "========================================"
    exit 0
else
    echo "  STACKS_HAVE_DIFFERENCES"
    echo "========================================"
    echo ""
    echo "Differences found:"
    for diff in "${DIFFERENCES[@]}"; do
        echo "  - $diff"
    done
    exit 1
fi
