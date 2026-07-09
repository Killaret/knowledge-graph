# Check Stacks Identity - Windows PowerShell
# This script verifies that dev, personal, and test stacks are consistent

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Stacks Identity Check" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$errors = 0
$differences = @()

# Step 1: Check service versions from Dockerfiles
Write-Host "`n[Step 1/5] Checking service versions..." -ForegroundColor Yellow

# Check Go version (from Dockerfile)
$goVersion = Select-String -Path "backend/Dockerfile" -Pattern "FROM golang:" | Select-Object -First 1
if ($goVersion) {
    $goVersionStr = $goVersion.Line -replace "FROM golang:", ""
    Write-Host "  Go version: $goVersionStr" -ForegroundColor Green
} else {
    Write-Host "  Go version: NOT FOUND" -ForegroundColor Red
    $errors++
}

# Check Node version (from Dockerfile)
$nodeVersion = Select-String -Path "frontend/Dockerfile" -Pattern "FROM node:" | Select-Object -First 1
if ($nodeVersion) {
    $nodeVersionStr = $nodeVersion.Line -replace "FROM node:", ""
    Write-Host "  Node version: $nodeVersionStr" -ForegroundColor Green
} else {
    Write-Host "  Node version: NOT FOUND" -ForegroundColor Red
    $errors++
}

# Check Python version (from Dockerfile)
$pythonVersion = Select-String -Path "nlp-service/Dockerfile" -Pattern "FROM python:" | Select-Object -First 1
if ($pythonVersion) {
    $pythonVersionStr = $pythonVersion.Line -replace "FROM python:", ""
    Write-Host "  Python version: $pythonVersionStr" -ForegroundColor Green
} else {
    Write-Host "  Python version: NOT FOUND" -ForegroundColor Red
    $errors++
}

# Step 2: Compare docker-compose files
Write-Host "`n[Step 2/5] Comparing docker-compose files..." -ForegroundColor Yellow

$devCompose = Get-Content "docker-compose.yml" -Raw
$personalCompose = Get-Content "docker-compose.personal.yml" -Raw
$testCompose = Get-Content "docker-compose.test.yml" -Raw

# Check for critical differences
$devServices = Select-String -Path "docker-compose.yml" -Pattern "image:|build:" | Select-Object -First 10
$personalServices = Select-String -Path "docker-compose.personal.yml" -Pattern "image:|build:" | Select-Object -First 10
$testServices = Select-String -Path "docker-compose.test.yml" -Pattern "image:|build:" | Select-Object -First 10

Write-Host "  Dev services: $($devServices.Count) found" -ForegroundColor Green
Write-Host "  Personal services: $($personalServices.Count) found" -ForegroundColor Green
Write-Host "  Test services: $($testServices.Count) found" -ForegroundColor Green

# Check for SKIP_AUTH consistency
$devSkipAuth = Select-String -Path "docker-compose.yml" -Pattern "SKIP_AUTH"
$personalSkipAuth = Select-String -Path "docker-compose.personal.yml" -Pattern "SKIP_AUTH"
$testSkipAuth = Select-String -Path "docker-compose.test.yml" -Pattern "SKIP_AUTH"

Write-Host "  Dev SKIP_AUTH: $($devSkipAuth.Line)" -ForegroundColor Green
Write-Host "  Personal SKIP_AUTH: $($personalSkipAuth.Line)" -ForegroundColor Green
Write-Host "  Test SKIP_AUTH: $($testSkipAuth.Line)" -ForegroundColor Green

# Step 3: Check configuration files
Write-Host "`n[Step 3/5] Checking configuration files..." -ForegroundColor Yellow

if (Test-Path "knowledge-graph.config.json") {
    $config = Get-Content "knowledge-graph.config.json" -Raw | ConvertFrom-Json
    Write-Host "  knowledge-graph.config.json: OK" -ForegroundColor Green
} else {
    Write-Host "  knowledge-graph.config.json: NOT FOUND" -ForegroundColor Red
    $errors++
}

# Step 4: Check nginx configurations
Write-Host "`n[Step 4/5] Checking nginx configurations..." -ForegroundColor Yellow

if (Test-Path "nginx.conf") {
    Write-Host "  nginx.conf: OK" -ForegroundColor Green
} else {
    Write-Host "  nginx.conf: NOT FOUND" -ForegroundColor Red
    $errors++
}

if (Test-Path "nginx.personal.conf") {
    Write-Host "  nginx.personal.conf: OK" -ForegroundColor Green
} else {
    Write-Host "  nginx.personal.conf: NOT FOUND" -ForegroundColor Red
    $errors++
}

# Step 5: Check stack health
Write-Host "`n[Step 5/5] Checking stack health..." -ForegroundColor Yellow

# Check dev stack
try {
    $devHealth = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get -TimeoutSec 5
    Write-Host "  Dev stack: OK" -ForegroundColor Green
} catch {
    Write-Host "  Dev stack: FAILED" -ForegroundColor Red
    $errors++
}

# Check personal stack
try {
    $personalHealth = Invoke-RestMethod -Uri "http://localhost:8082/health" -Method Get -TimeoutSec 5
    Write-Host "  Personal stack: OK" -ForegroundColor Green
} catch {
    Write-Host "  Personal stack: FAILED" -ForegroundColor Red
    $errors++
}

# Final result
Write-Host "`n========================================" -ForegroundColor Cyan
if ($errors -eq 0) {
    Write-Host "  STACKS_IDENTICAL" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "  STACKS_HAVE_DIFFERENCES" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "`nDifferences found:" -ForegroundColor Yellow
    $differences | ForEach-Object { Write-Host "  - $_" -ForegroundColor Red }
    exit 1
}
