# Start Knowledge Graph Test Stack
# This script launches a completely isolated test environment

Write-Host "Starting Knowledge Graph Test Stack..." -ForegroundColor Green

# Stop and remove any existing test stack
Write-Host "Cleaning up any existing test stack..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml down -v 2>$null

# Build and start the test stack
Write-Host "Building and starting test stack..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml up -d --build

# Wait for services to be healthy
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# Check service status
Write-Host "Checking service status..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml ps

Write-Host ""
Write-Host "Test stack ready!" -ForegroundColor Green
Write-Host "Frontend: http://localhost:3002" -ForegroundColor Cyan
Write-Host "Backend API: http://localhost:8083" -ForegroundColor Cyan
Write-Host "Graph Service: http://localhost:8083/graph-service" -ForegroundColor Cyan
Write-Host "NLP Service: http://localhost:5002" -ForegroundColor Cyan
Write-Host ""
Write-Host "To stop the test stack, run: .\scripts\stop-test.ps1" -ForegroundColor Yellow
