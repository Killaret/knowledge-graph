# Start Test Stack - Windows PowerShell
# This script stops any existing test stack, then starts a fresh test stack

Write-Host "Starting test stack setup..." -ForegroundColor Cyan

# Stop and remove previous test stack
Write-Host "Stopping previous test stack..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml down -v

# Remove any orphaned kg-test-* containers that might have been left behind
# by a previous incomplete shutdown or a different compose project.
$orphans = docker ps -aq --filter "name=kg-test"
if ($orphans) {
    Write-Host "Removing orphaned test containers..." -ForegroundColor Yellow
    docker rm -f $orphans | Out-Null
}

# Start test stack
Write-Host "Starting test stack..." -ForegroundColor Yellow
docker compose -f docker-compose.test.yml up -d --build --wait
$upExit = $LASTEXITCODE
if ($upExit -ne 0) {
    Write-Host "ERROR: Test stack failed to start (exit $upExit)" -ForegroundColor Red
    exit $upExit
}

# Wait for all containers to be healthy
Write-Host "Waiting for containers to be healthy..." -ForegroundColor Yellow
$timeout = 120 # 2 minutes
$startTime = Get-Date

while (((Get-Date) - $startTime).TotalSeconds -lt $timeout) {
    $healthy = docker compose -f docker-compose.test.yml ps --format json | ConvertFrom-Json | Where-Object { $_.State -eq "running" -and $_.Health -eq "healthy" }
    $total = docker compose -f docker-compose.test.yml ps --format json | ConvertFrom-Json | Where-Object { $_.State -eq "running" } | Measure-Object | Select-Object -ExpandProperty Count
    
    if ($healthy.Count -eq $total) {
        Write-Host "All containers are healthy!" -ForegroundColor Green
        break
    }
    
    Write-Host "Healthy: $($healthy.Count)/$total containers" -ForegroundColor Gray
    Start-Sleep -Seconds 5
}

# Final check
$finalCheck = docker compose -f docker-compose.test.yml ps
Write-Host "`nTest stack status:" -ForegroundColor Cyan
Write-Host $finalCheck

Write-Host "`nTest stack ready: http://127.0.0.1:3002" -ForegroundColor Green
Write-Host "Backend API: http://127.0.0.1:18083" -ForegroundColor Green
Write-Host "Frontend: http://127.0.0.1:3002" -ForegroundColor Green
