# Check Stacks Health - Windows PowerShell
# This script checks the health of specified stack(s)
# Usage: .\check-stacks-health.ps1 [-Stack <dev|personal|test|all>]

param(
    [ValidateSet("dev", "personal", "test", "all")]
    [string]$Stack = "all"
)

Write-Host "Checking stacks health..." -ForegroundColor Cyan
Write-Host "Stack: $Stack" -ForegroundColor Yellow

$errors = 0

function Check-DevStack {
    Write-Host "`nChecking dev stack..." -ForegroundColor Yellow
    
    # Check dev containers
    $devContainers = docker ps --filter "name=kg-" --format json | ConvertFrom-Json | Where-Object { $_.Names -notlike "*test*" -and $_.Names -notlike "*personal*" }
    if ($devContainers.Count -eq 0) {
        Write-Host "  No dev containers running" -ForegroundColor Red
        return 1
    } else {
        Write-Host "  Dev containers: $($devContainers.Count) running" -ForegroundColor Green
    }
    
    # Check dev health endpoint
    try {
        $devHealth = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -TimeoutSec 5
        Write-Host "  Dev health endpoint: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Dev health endpoint: FAILED" -ForegroundColor Red
        return 1
    }
    
    # Check dev API
    try {
        $devNotes = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
        Write-Host "  Dev API: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Dev API: FAILED" -ForegroundColor Red
        return 1
    }
    
    return 0
}

function Check-PersonalStack {
    Write-Host "`nChecking personal stack..." -ForegroundColor Yellow
    
    # Check personal containers
    $personalContainers = docker ps --filter "name=kg-" --format json | ConvertFrom-Json | Where-Object { $_.Names -like "*personal*" }
    if ($personalContainers.Count -eq 0) {
        Write-Host "  No personal containers running" -ForegroundColor Red
        return 1
    } else {
        Write-Host "  Personal containers: $($personalContainers.Count) running" -ForegroundColor Green
    }
    
    # Check personal health endpoint
    try {
        $personalHealth = Invoke-RestMethod -Uri "http://localhost:8082/health" -Method Get -TimeoutSec 5
        Write-Host "  Personal health endpoint: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Personal health endpoint: FAILED" -ForegroundColor Red
        return 1
    }
    
    # Check personal API
    try {
        $personalNotes = Invoke-RestMethod -Uri "http://localhost:8082/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
        Write-Host "  Personal API: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Personal API: FAILED" -ForegroundColor Red
        return 1
    }
    
    return 0
}

function Check-TestStack {
    Write-Host "`nChecking test stack..." -ForegroundColor Yellow
    
    # Check test containers
    $testContainers = docker ps --filter "name=kg-" --format json | ConvertFrom-Json | Where-Object { $_.Names -like "*test*" }
    if ($testContainers.Count -eq 0) {
        Write-Host "  No test containers running" -ForegroundColor Red
        return 1
    } else {
        Write-Host "  Test containers: $($testContainers.Count) running" -ForegroundColor Green
    }
    
    # Check test health endpoint
    try {
        $testHealth = Invoke-RestMethod -Uri "http://localhost:8083/health" -Method Get -TimeoutSec 5
        Write-Host "  Test health endpoint: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Test health endpoint: FAILED" -ForegroundColor Red
        return 1
    }
    
    # Check test API
    try {
        $testNotes = Invoke-RestMethod -Uri "http://localhost:8083/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
        Write-Host "  Test API: OK" -ForegroundColor Green
    } catch {
        Write-Host "  Test API: FAILED" -ForegroundColor Red
        return 1
    }
    
    return 0
}

# Check requested stacks
if ($Stack -eq "all" -or $Stack -eq "dev") {
    $errors += Check-DevStack
}

if ($Stack -eq "all" -or $Stack -eq "personal") {
    $errors += Check-PersonalStack
}

if ($Stack -eq "all" -or $Stack -eq "test") {
    $errors += Check-TestStack
}

# Final result
Write-Host "`n" -NoNewline
if ($errors -eq 0) {
    if ($Stack -eq "all") {
        Write-Host "All stacks are healthy" -ForegroundColor Green
    } else {
        Write-Host "$Stack stack is healthy" -ForegroundColor Green
    }
    exit 0
} else {
    if ($Stack -eq "all") {
        Write-Host "Stacks have $errors error(s)" -ForegroundColor Red
    } else {
        Write-Host "$Stack stack has $errors error(s)" -ForegroundColor Red
    }
    exit 1
}
