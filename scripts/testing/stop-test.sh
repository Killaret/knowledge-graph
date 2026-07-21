#!/bin/bash
# Stop Test Stack - Linux/Mac
# This script stops and destroys the test stack including volumes

set -e

echo "Stopping test stack..."

# Stop and remove test stack with volumes
docker compose -f docker-compose.test.yml down -v

echo "Test stack destroyed"
