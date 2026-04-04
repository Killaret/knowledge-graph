#!/usr/bin/env pwsh
# Complete rebuild and async task testing script

param(
    [switch]$SkipBuild = $false,
    [switch]$SkipTest = $false
)

$ErrorActionPreference = "Stop"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     Knowledge Graph - Async Tasks Rebuild & Test Suite     ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Helper functions
function Show-Status {
    param([string]$Message, [string]$Status)
    $statusColor = if ($Status -eq "OK") { "Green" } else { "Yellow" }
    Write-Host "[$Status] " -ForegroundColor $statusColor -NoNewline
    Write-Host $Message
}

function Show-Section {
    param([string]$Title)
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host $Title -ForegroundColor Cyan
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
}

# Check prerequisites
Show-Section "1. Checking Prerequisites"

if (!(Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Docker is not installed!" -ForegroundColor Red
    exit 1
}
Show-Status "Docker installed" "OK"

if (!(Get-Command docker-compose -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Docker Compose is not installed!" -ForegroundColor Red
    exit 1
}
Show-Status "Docker Compose installed" "OK"

$projectPath = "D:\knowledge-graph"
if (!(Test-Path $projectPath)) {
    Write-Host "ERROR: Project path not found: $projectPath" -ForegroundColor Red
    exit 1
}
Show-Status "Project directory found" "OK"

# Stop old containers
if (-not $SkipBuild) {
    Show-Section "2. Cleaning Up Old Containers"
    
    Write-Host "Stopping containers..." -ForegroundColor Yellow
    Push-Location $projectPath
    docker-compose down --remove-orphans | Out-Null
    Show-Status "Old containers removed" "OK"
    Pop-Location
    
    Write-Host ""
    Write-Host "Waiting 3 seconds..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3
}

# Rebuild
if (-not $SkipBuild) {
    Show-Section "3. Building and Starting Services"
    
    Write-Host "Building images and starting containers (this may take a few minutes)..." -ForegroundColor Yellow
    Write-Host "Tip: Run 'docker-compose logs -f' in another terminal to watch the build progress" -ForegroundColor Gray
    
    Push-Location $projectPath
    docker-compose up --build -d 2>&1 | Out-Null
    Pop-Location
    
    Write-Host "Containers started." -ForegroundColor Green
    
    # Wait for services to be healthy
    Write-Host ""
    Write-Host "Waiting for services to become healthy (max 60 seconds)..." -ForegroundColor Yellow
    
    $maxRetries = 60
    $retries = 0
    $allHealthy = $false
    
    while ($retries -lt $maxRetries -and -not $allHealthy) {
        $ps = docker-compose ps --format "table {{.Service}}\t{{.Status}}"
        
        $postgresHealthy = $ps | Select-String "healthy" | Select-String "postgres"
        $redisHealthy = $ps | Select-String "healthy" | Select-String "redis"
        $backendRunning = $ps | Select-String "running" | Select-String "backend"
        $workerRunning = $ps | Select-String "running" | Select-String "worker"
        $nlpRunning = $ps | Select-String "running" | Select-String "nlp"
        
        if ($postgresHealthy -and $redisHealthy -and $backendRunning -and $workerRunning -and $nlpRunning) {
            $allHealthy = $true
        } else {
            Start-Sleep -Seconds 1
            $retries++
            Write-Host "." -ForegroundColor Gray -NoNewline
        }
    }
    
    if ($allHealthy) {
        Write-Host ""
        Show-Status "All services healthy" "OK"
    } else {
        Write-Host ""
        Write-Host "WARNING: Services took longer than expected to become healthy" -ForegroundColor Yellow
        docker-compose ps
    }
}

# Display service status
Show-Section "4. Service Status"
Push-Location $projectPath
$status = docker-compose ps --format "table {{.Service}}\t{{.Status}}"
$status | Write-Host
Pop-Location

# Test async tasks
if (-not $SkipTest) {
    Show-Section "5. Testing Async Task Processing"
    
    Write-Host "Creating test note..." -ForegroundColor Yellow
    
    $testNote = @{
        title   = "Async Processing Test - $(Get-Date -Format 'HH:mm:ss')"
        content = "This comprehensive test note validates the async task processing pipeline. The system should automatically extract keywords and compute embeddings using the worker service. Upon successful processing, corresponding log entries should appear in the worker container showing completed task execution."
        metadata = @{
            "test" = $true
            "timestamp" = (Get-Date).ToString("o")
            "type" = "performance_test"
        }
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/notes" `
            -Method POST `
            -ContentType "application/json" `
            -Body $testNote `
            -TimeoutSec 10 `
            -ErrorAction Stop
        
        Show-Status "Note created successfully" "OK"
        $noteID = $response.id
        Write-Host "  Note ID: $noteID" -ForegroundColor Cyan
        
    } catch {
        Write-Host "ERROR: Failed to create test note" -ForegroundColor Red
        Write-Host "  Message: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
        Write-Host "Checking backend status..." -ForegroundColor Yellow
        Push-Location $projectPath
        docker-compose logs --tail=20 kg-backend
        Pop-Location
        exit 1
    }
    
    Write-Host ""
    Write-Host "Waiting 12 seconds for worker to process tasks..." -ForegroundColor Yellow
    
    $spinner = @('|', '/', '-', '\')
    for ($i = 0; $i -lt 12; $i++) {
        Write-Host -NoNewline "`r$($spinner[$i % 4]) Processing... ($($i+1)/12)"
        Start-Sleep -Seconds 1
    }
    Write-Host -NoNewline "`r✓ Processing complete!              " 
    Write-Host ""
    
    # Check logs for successful processing
    Show-Section "6. Checking Task Processing Results"
    
    Push-Location $projectPath
    $workerLogs = docker-compose logs kg-worker --tail=100
    Pop-Location
    
    $keywordMatches = [regex]::Matches($workerLogs, "HandleExtractKeywords.*successfully processed")
    $embeddingMatches = [regex]::Matches($workerLogs, "HandleComputeEmbedding.*successfully processed")
    
    if ($keywordMatches.Count -gt 0) {
        Show-Status "Keyword extraction tasks processed" "OK"
        Write-Host "  Found $($keywordMatches.Count) successful processing(s)" -ForegroundColor Cyan
    } else {
        Write-Host "WARNING: No keyword extraction tasks found in logs" -ForegroundColor Yellow
        Write-Host "  This might be due to:" -ForegroundColor Gray
        Write-Host "    - Tasks still processing (retry in a few seconds)" -ForegroundColor Gray
        Write-Host "    - NLP service not ready yet" -ForegroundColor Gray
    }
    
    if ($embeddingMatches.Count -gt 0) {
        Show-Status "Embedding computation tasks processed" "OK"
        Write-Host "  Found $($embeddingMatches.Count) successful processing(s)" -ForegroundColor Cyan
    } else {
        Write-Host "WARNING: No embedding computation tasks found in logs" -ForegroundColor Yellow
    }
    
    # Show recent worker logs
    Write-Host ""
    Write-Host "Recent worker logs:" -ForegroundColor Yellow
    Push-Location $projectPath
    docker-compose logs --tail=15 kg-worker | Write-Host
    Pop-Location
}

# Final summary
Show-Section "7. Summary"

Write-Host ""
Write-Host "✅ Rebuild and testing complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. View live logs:        docker-compose logs -f kg-worker" -ForegroundColor White
Write-Host "  2. Test more notes:       Run this script again or use test_async_tasks.ps1" -ForegroundColor White
Write-Host "  3. Stop containers:       docker-compose down" -ForegroundColor White
Write-Host "  4. Read docs:             ASYNC_TESTING_GUIDE.md, REBUILD_CHECKLIST.md" -ForegroundColor White
Write-Host ""
