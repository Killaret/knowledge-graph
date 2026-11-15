#!/bin/bash
# Start Test Stack - Linux/Mac
# This script stops any existing test stack, then starts a fresh test stack

set -e

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
cd "$repo_dir"
node scripts/testing/check-test-stack-ownership.mjs "$repo_dir"

# Use .env.test if the user has created one, otherwise fall back to a default test secret.
if [ -f .env.test ]; then
    while IFS='=' read -r name value; do
        name=$(printf '%s' "$name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        value=$(printf '%s' "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        case "$name" in \#*|"") continue ;; esac
        if [ -z "${!name:-}" ]; then
            export "$name=$value"
        fi
    done < .env.test
fi

if [ -z "$JWT_SECRET" ]; then
    export JWT_SECRET='test-jwt-secret-32-characters-long'
    echo "JWT_SECRET not set; using default test secret."
fi

echo "Starting test stack setup..."

# Stop and remove previous test stack
echo "Stopping previous test stack..."
docker compose -f docker-compose.test.yml down -v

# Start test stack
echo "Starting test stack..."
docker compose -f docker-compose.test.yml up -d --build --wait

# Wait for all containers to be healthy
echo "Waiting for containers to be healthy..."
timeout=120 # 2 minutes
start_time=$(date +%s)

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))
    
    if [ $elapsed -ge $timeout ]; then
        echo "Timeout waiting for containers to be healthy"
        exit 1
    fi
    
    healthy=$(docker compose -f docker-compose.test.yml ps --format json | jq '[.[] | select(.State == "running" and .Health == "healthy")] | length')
    total=$(docker compose -f docker-compose.test.yml ps --format json | jq '[.[] | select(.State == "running")] | length')
    
    if [ "$healthy" -eq "$total" ]; then
        echo "All containers are healthy!"
        break
    fi
    
    echo "Healthy: $healthy/$total containers"
    sleep 5
done

# Final check
echo ""
echo "Test stack status:"
docker compose -f docker-compose.test.yml ps

echo ""
echo "Test stack ready: http://localhost:3002"
echo "Backend API: http://localhost:18083"
echo "Frontend: http://localhost:3002"
