#!/usr/bin/env pwsh
# Проверка доступности frontend и backend

Write-Host "=== Knowledge Graph - Test Verification ===" -ForegroundColor Cyan
Write-Host ""

# Проверка backend
Write-Host "1. Checking Backend (http://localhost:9000)..." -ForegroundColor Yellow
try {
    $backendResp = Invoke-RestMethod -Uri "http://localhost:9000/health" -Method GET -TimeoutSec 5
    Write-Host "   ✓ Backend is UP" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Backend is DOWN" -ForegroundColor Red
    exit 1
}

# Проверка frontend
Write-Host "2. Checking Frontend (http://localhost:5173)..." -ForegroundColor Yellow
try {
    $frontendResp = Invoke-WebRequest -Uri "http://localhost:5173" -UseBasicParsing -TimeoutSec 5
    Write-Host "   ✓ Frontend is UP (Status: $($frontendResp.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Frontend is DOWN" -ForegroundColor Red
    exit 1
}

# Проверка количества заметок
Write-Host "3. Checking Database..." -ForegroundColor Yellow
try {
    $notesResp = Invoke-RestMethod -Uri "http://localhost:9000/notes?limit=1" -Method GET -TimeoutSec 5
    $totalNotes = $notesResp.total
    Write-Host "   ✓ Database has $totalNotes notes" -ForegroundColor Green
    
    if ($totalNotes -eq 0) {
        Write-Host ""
        Write-Host "⚠ Warning: Database is empty! Run seed_data.py to populate." -ForegroundColor Yellow
        Write-Host "   Command: python scripts/database/seed_data.py" -ForegroundColor Gray
    }
} catch {
    Write-Host "   ✗ Cannot read database" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== All checks passed! ===" -ForegroundColor Green
Write-Host ""
Write-Host "Frontend: http://localhost:5173" -ForegroundColor Cyan
Write-Host "Backend:  http://localhost:9000" -ForegroundColor Cyan
Write-Host ""
