#!/bin/bash
# Check Stacks Health - Linux/Mac
# This script checks the health of specified stack(s)
# Usage: ./check-stacks-health.sh [--stack <dev|personal|test|all>]

set -e

STACK="all"

# Parse arguments
while [ "$#" -gt 0 ]; do
    case "$1" in
        --stack|-s)
            if [ -n "$2" ]; then
                STACK="$2"
                shift 2
            else
                echo "Error: --stack requires a value (dev|personal|test|all)" >&2
                exit 1
            fi
            ;;
        *)
            echo "Unknown argument: $1" >&2
            echo "Usage: ./check-stacks-health.sh [--stack <dev|personal|test|all>]" >&2
            exit 1
            ;;
    esac
done

# Validate stack value
case "$STACK" in
    dev|personal|test|all) ;;
    *)
        echo "Invalid stack: $STACK" >&2
        echo "Usage: ./check-stacks-health.sh [--stack <dev|personal|test|all>]" >&2
        exit 1
        ;;
esac

echo "Checking stacks health..."
echo "Stack: $STACK"

errors=0

check_api() {
    local port=$1
    if curl -s -f "http://127.0.0.1:$port/api/v1/notes?limit=1" > /dev/null || curl -s -f "http://127.0.0.1:$port/api/v1/graph/all?limit=1" > /dev/null; then
        echo "  API: OK"
        return 0
    else
        echo "  API: FAILED"
        return 1
    fi
}

check_graph_service() {
    local port=$1
    if curl -s -f "http://127.0.0.1:$port/health" > /dev/null; then
        echo "  Graph-service: OK"
        return 0
    else
        echo "  Graph-service: FAILED"
        return 1
    fi
}

check_dev_stack() {
    echo ""
    echo "Checking dev stack..."

    dev_containers=$(docker ps --filter "name=kg-" --format '{{.Names}}' | grep -v -E '(test|personal)' | wc -l | tr -d '[:space:]')
    if [ "$dev_containers" -eq 0 ]; then
        echo "  No dev containers running"
        return 1
    else
        echo "  Dev containers: $dev_containers running"
    fi

    if curl -s -f http://127.0.0.1:18080/health > /dev/null; then
        echo "  Dev health endpoint: OK"
    else
        echo "  Dev health endpoint: FAILED"
        return 1
    fi

    check_api 18080 || return 1
    check_graph_service 9091 || return 1

    return 0
}

check_personal_stack() {
    echo ""
    echo "Checking personal stack..."

    personal_containers=$(docker ps --filter "name=kg-" --format '{{.Names}}' | grep -E 'personal' | wc -l | tr -d '[:space:]')
    if [ "$personal_containers" -eq 0 ]; then
        echo "  No personal containers running"
        return 1
    else
        echo "  Personal containers: $personal_containers running"
    fi

    if curl -s -f http://127.0.0.1:18082/health > /dev/null; then
        echo "  Personal health endpoint: OK"
    else
        echo "  Personal health endpoint: FAILED"
        return 1
    fi

    check_api 18082 || return 1
    check_graph_service 9092 || return 1

    return 0
}

check_test_stack() {
    echo ""
    echo "Checking test stack..."

    test_containers=$(docker ps --filter "name=kg-" --format '{{.Names}}' | grep -E 'test' | wc -l | tr -d '[:space:]')
    if [ "$test_containers" -eq 0 ]; then
        echo "  No test containers running"
        return 1
    else
        echo "  Test containers: $test_containers running"
    fi

    if curl -s -f http://127.0.0.1:18083/health > /dev/null; then
        echo "  Test health endpoint: OK"
    else
        echo "  Test health endpoint: FAILED"
        return 1
    fi

    check_api 18083 || return 1
    check_graph_service 19091 || return 1

    return 0
}

# Check requested stacks
if [ "$STACK" = "all" ] || [ "$STACK" = "dev" ]; then
    check_dev_stack || ((errors++)) || true
fi

if [ "$STACK" = "all" ] || [ "$STACK" = "personal" ]; then
    check_personal_stack || ((errors++)) || true
fi

if [ "$STACK" = "all" ] || [ "$STACK" = "test" ]; then
    check_test_stack || ((errors++)) || true
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
