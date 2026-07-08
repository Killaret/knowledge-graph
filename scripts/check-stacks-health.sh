#!/bin/bash
# Check Stacks Health - Linux/Mac
# This script checks the health of dev and personal stacks

set -e

echo "Checking stacks health..."

errors=0

# Check dev stack
echo ""
echo "Checking dev stack..."

# Check dev containers
dev_containers=$(docker ps --filter "name=kg-" --format json | jq '[.[] | select(.Names | contains("test") | not)] | length')
if [ "$dev_containers" -eq 0 ]; then
    echo "  No dev containers running"
    ((errors++))
else
    echo "  Dev containers: $dev_containers running"
fi

# Check dev health endpoint
if curl -s -f http://localhost:8080/health > /dev/null; then
    echo "  Dev health endpoint: OK"
else
    echo "  Dev health endpoint: FAILED"
    ((errors++))
fi

# Check dev API
if curl -s -f "http://localhost:8080/api/v1/notes?limit=1" > /dev/null; then
    echo "  Dev API: OK"
else
    echo "  Dev API: FAILED"
    ((errors++))
fi

# Check personal stack
echo ""
echo "Checking personal stack..."

# Check personal containers
personal_containers=$(docker ps --filter "name=kg-" --format json | jq '[.[] | select(.Names | contains("personal"))] | length')
if [ "$personal_containers" -eq 0 ]; then
    echo "  No personal containers running"
    ((errors++))
else
    echo "  Personal containers: $personal_containers running"
fi

# Check personal health endpoint
if curl -s -f http://localhost:8082/health > /dev/null; then
    echo "  Personal health endpoint: OK"
else
    echo "  Personal health endpoint: FAILED"
    ((errors++))
fi

# Check personal API
if curl -s -f "http://localhost:8082/api/v1/notes?limit=1" > /dev/null; then
    echo "  Personal API: OK"
else
    echo "  Personal API: FAILED"
    ((errors++))
fi

# Final result
echo ""
if [ $errors -eq 0 ]; then
    echo "Dev and Personal stacks are healthy"
    exit 0
else
    echo "Dev and Personal stacks have $errors error(s)"
    exit 1
fi
