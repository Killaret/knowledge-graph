# Check Stacks Health - Windows PowerShell
# This script checks the health of dev and personal stacks

Write-Host "Checking stacks health..." -ForegroundColor Cyan

$errors = 0

# Check dev stack
Write-Host "`nChecking dev stack..." -ForegroundColor Yellow

# Check dev containers
$devContainers = docker ps --filter "name=kg-" --format json | ConvertFrom-Json | Where-Object { $_.Names -notlike "*test*" }
if ($devContainers.Count -eq 0) {
    Write-Host "  No dev containers running" -ForegroundColor Red
    $errors++
} else {
    Write-Host "  Dev containers: $($devContainers.Count) running" -ForegroundColor Green
}

# Check dev health endpoint
try {
    $devHealth = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -TimeoutSec 5
    Write-Host "  Dev health endpoint: OK" -ForegroundColor Green
} catch {
    Write-Host "  Dev health endpoint: FAILED" -ForegroundColor Red
    $errors++
}

# Check dev API
try {
    $devNotes = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
    Write-Host "  Dev API: OK" -ForegroundColor Green
} catch {
    Write-Host "  Dev API: FAILED" -ForegroundColor Red
    $errors++
}

# Check personal stack
Write-Host "`nChecking personal stack..." -ForegroundColor Yellow

# Check personal containers
$personalContainers = docker ps --filter "name=kg-" --format json | ConvertFrom-Json | Where-Object { $_.Names -like "*personal*" }
if ($personalContainers.Count -eq 0) {
    Write-Host "  No personal containers running" -ForegroundColor Red
    $errors++
} else {
    Write-Host "  Personal containers: $($personalContainers.Count) running" -ForegroundColor Green
}

# Check personal health endpoint
try {
    $personalHealth = Invoke-RestMethod -Uri "http://localhost:8082/health" -Method Get -TimeoutSec 5
    Write-Host "  Personal health endpoint: OK" -ForegroundColor Green
} catch {
    Write-Host "  Personal health endpoint: FAILED" -ForegroundColor Red
    $errors++
}

# Check personal API
try {
    $personalNotes = Invoke-RestMethod -Uri "http://localhost:8082/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
    Write-Host "  Personal API: OK" -ForegroundColor Green
} catch {
    Write-Host "  Personal API: FAILED" -ForegroundColor Red
    $errors++
}

# Final result
Write-Host "`n" -NoNewline
if ($errors -eq 0) {
    Write-Host "Dev and Personal stacks are healthy" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Dev and Personal stacks have $errors error(s)" -ForegroundColor Red
    exit 1
}
