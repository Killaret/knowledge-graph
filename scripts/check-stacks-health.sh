#!/bin/bash
# Check Stacks Health - Linux/Mac
# This script checks the health of specified stack(s)
# Usage: ./check-stacks-health.sh [--stack <dev|personal|test|all>]

set -e

STACK="${1:-all}"

echo "Checking stacks health..."
echo "Stack: $STACK"

errors=0

check_dev_stack() {
    echo ""
    echo "Checking dev stack..."
    
    # Check dev containers
    dev_containers=$(docker ps --filter "name=kg-" --format json | jq '[.[] | select(.Names | contains("test") | not) | select(.Names | contains("personal") | not)] | length')
    if [ "$dev_containers" -eq 0 ]; then
        echo "  No dev containers running"
        return 1
    else
        echo "  Dev containers: $dev_containers running"
    fi
    
    # Check dev health endpoint
    if curl -s -f http://localhost:8080/health > /dev/null; then
        echo "  Dev health endpoint: OK"
    else
        echo "  Dev health endpoint: FAILED"
        return 1
    fi
    
    # Check dev API
    if curl -s -f "http://localhost:8080/api/v1/notes?limit=1" > /dev/null; then
        echo "  Dev API: OK"
    else
        echo "  Dev API: FAILED"
        return 1
    fi
    
    return 0
}

check_personal_stack() {
    echo ""
    echo "Checking personal stack..."
    
    # Check personal containers
    personal_containers=$(docker ps --filter "name=kg-" --format json | jq '[.[] | select(.Names | contains("personal"))] | length')
    if [ "$personal_containers" -eq 0 ]; then
        echo "  No personal containers running"
        return 1
    else
        echo "  Personal containers: $personal_containers running"
    fi
    
    # Check personal health endpoint
    if curl -s -f http://localhost:8082/health > /dev/null; then
        echo "  Personal health endpoint: OK"
    else
        echo "  Personal health endpoint: FAILED"
        return 1
    fi
    
    # Check personal API
    if curl -s -f "http://localhost:8082/api/v1/notes?limit=1" > /dev/null; then
        echo "  Personal API: OK"
    else
        echo "  Personal API: FAILED"
        return 1
    fi
    
    return 0
}

check_test_stack() {
    echo ""
    echo "Checking test stack..."
    
    # Check test containers
    test_containers=$(docker ps --filter "name=kg-" --format json | jq '[.[] | select(.Names | contains("test"))] | length')
    if [ "$test_containers" -eq 0 ]; then
        echo "  No test containers running"
        return 1
    else
        echo "  Test containers: $test_containers running"
    fi
    
    # Check test health endpoint
    if curl -s -f http://localhost:8083/health > /dev/null; then
        echo "  Test health endpoint: OK"
    else
        echo "  Test health endpoint: FAILED"
        return 1
    fi
    
    # Check test API
    if curl -s -f "http://localhost:8083/api/v1/notes?limit=1" > /dev/null; then
        echo "  Test API: OK"
    else
        echo "  Test API: FAILED"
        return 1
    fi
    
    return 0
}

# Check requested stacks
if [ "$STACK" = "all" ] || [ "$STACK" = "dev" ]; then
    check_dev_stack || ((errors++))
fi

if [ "$STACK" = "all" ] || [ "$STACK" = "personal" ]; then
    check_personal_stack || ((errors++))
fi

if [ "$STACK" = "all" ] || [ "$STACK" = "test" ]; then
    check_test_stack || ((errors++))
fi

# Final result
echo ""
if [ $errors -eq 0 ]; then
    if [ "$STACK" = "all" ]; then
        echo "All stacks are healthy"
    else
        echo "$STACK stack is healthy"
    fi
    exit 0
else
    if [ "$STACK" = "all" ]; then
        echo "Stacks have $errors error(s)"
    else
        echo "$STACK stack has $errors error(s)"
    fi
    exit 1
fi
