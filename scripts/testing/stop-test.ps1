# Stop Test Stack - Windows PowerShell
# This script stops and destroys the test stack including volumes

Write-Host "Stopping test stack..." -ForegroundColor Yellow

# Stop and remove test stack with volumes
docker compose -f docker-compose.test.yml down -v

Write-Host "Test stack destroyed" -ForegroundColor Green
