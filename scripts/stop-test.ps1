# Stop Knowledge Graph Test Stack
# This script destroys the test environment and removes all data

Write-Host "Destroying Knowledge Graph Test Stack..." -ForegroundColor Red

# Stop and remove all containers, networks, and volumes
Write-Host "Stopping containers and removing volumes..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml down -v

# Remove any orphaned containers
Write-Host "Removing orphaned containers..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml down --remove-orphans

Write-Host ""
Write-Host "Test stack destroyed. All test data has been removed." -ForegroundColor Green
Write-Host "Volumes removed: pgdata_test, mongodbdata_test" -ForegroundColor Cyan
