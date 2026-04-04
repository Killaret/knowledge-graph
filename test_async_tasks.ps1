# PowerShell script to test async task processing
# Run: .\test_async_tasks.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing Async Task Processing" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if services are running
Write-Host "Checking service status..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "Creating test note to trigger async tasks..." -ForegroundColor Yellow

# Create a test note
$testNote = @{
    title   = "Async Test Note - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    content = "This is a comprehensive test note to validate the async task processing pipeline. The system should extract keywords and compute embeddings automatically using the worker service. If everything works correctly, you should see corresponding log entries in the worker container showing successful task processing and completion."
    metadata = @{
        "test" = $true
        "timestamp" = (Get-Date).ToString("o")
    }
} | ConvertTo-Json

Write-Host "Request body:" -ForegroundColor Cyan
Write-Host $testNote
Write-Host ""

try {
    Write-Host "Sending POST /notes request..." -ForegroundColor Yellow
    $response = Invoke-RestMethod -Uri "http://localhost:8080/notes" `
        -Method POST `
        -ContentType "application/json" `
        -Body $testNote `
        -ErrorAction Stop
    
    Write-Host "Note created successfully!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    $response | ConvertTo-Json | Write-Host
    
    $noteID = $response.id
    Write-Host ""
    Write-Host "Note ID: $noteID" -ForegroundColor Cyan
    Write-Host ""
    
} catch {
    Write-Host "ERROR: Failed to create note" -ForegroundColor Red
    Write-Host $_.Exception.Message
    Write-Host ""
    Write-Host "Make sure backend is running: docker-compose ps" -ForegroundColor Yellow
    exit 1
}

Write-Host "Waiting 10 seconds for worker to process tasks..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
Write-Host ""

# Check worker logs
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Worker Logs (last 40 lines)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
docker-compose logs --tail=40 kg-worker

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Looking for successful task processing..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Check for successful processing
$workerLogs = docker-compose logs kg-worker
$hasKeywordProcessing = $workerLogs -match "HandleExtractKeywords.*successfully processed"
$hasEmbeddingProcessing = $workerLogs -match "HandleComputeEmbedding.*successfully processed"

if ($hasKeywordProcessing) {
    Write-Host "✓ Keywords extraction task processed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Keywords extraction task NOT found in logs" -ForegroundColor Yellow
}

if ($hasEmbeddingProcessing) {
    Write-Host "✓ Embedding computation task processed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Embedding computation task NOT found in logs" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Testing complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "For more details, run:" -ForegroundColor Yellow
Write-Host "  docker-compose logs -f kg-worker" -ForegroundColor Cyan
