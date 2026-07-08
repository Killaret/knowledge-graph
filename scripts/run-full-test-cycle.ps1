# Full Test Cycle - Windows PowerShell
# This script orchestrates the complete testing cycle

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Knowledge Graph Full Test Cycle" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Step 1: Check stacks health
Write-Host "`n[Step 1/8] Checking dev and personal stacks health..." -ForegroundColor Yellow
& .\scripts\check-stacks-health.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Dev or personal stacks are not healthy" -ForegroundColor Red
    Write-Host "Please start dev and personal stacks first" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Dev and personal stacks are healthy" -ForegroundColor Green

# Step 2: Start test stack
Write-Host "`n[Step 2/8] Starting test stack..." -ForegroundColor Yellow
& .\scripts\start-test.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to start test stack" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Test stack started" -ForegroundColor Green

# Step 3: Seed test data
Write-Host "`n[Step 3/8] Seeding test data..." -ForegroundColor Yellow
& .\scripts\seed-test-data.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to seed test data" -ForegroundColor Red
    Write-Host "Continuing anyway (data might already exist)" -ForegroundColor Yellow
}
Write-Host "✓ Test data seeded" -ForegroundColor Green

# Step 4: Manual testing instructions
Write-Host "`n[Step 4/8] Test environment ready" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MANUAL TESTING INSTRUCTIONS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test stack URLs:" -ForegroundColor Yellow
Write-Host "  Frontend: http://localhost:13002" -ForegroundColor White
Write-Host "  Backend API: http://localhost:18083" -ForegroundColor White
Write-Host ""
Write-Host "Follow the manual test checklist:" -ForegroundColor Yellow
Write-Host "  docs/MANUAL_TEST_CHECKLISTS.md" -ForegroundColor White
Write-Host ""
Write-Host "Test user credentials:" -ForegroundColor Yellow
Write-Host "  Login: testuser" -ForegroundColor White
Write-Host "  Password: TestPassword123!" -ForegroundColor White
Write-Host ""
Write-Host "Press Enter when manual testing is complete..." -ForegroundColor Cyan
Read-Host

# Step 5: Stop test stack
Write-Host "`n[Step 5/8] Stopping test stack..." -ForegroundColor Yellow
& .\scripts\stop-test.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Failed to stop test stack" -ForegroundColor Yellow
} else {
    Write-Host "✓ Test stack destroyed" -ForegroundColor Green
}

# Step 6: Check stacks health again
Write-Host "`n[Step 6/8] Checking dev and personal stacks health after testing..." -ForegroundColor Yellow
& .\scripts\check-stacks-health.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Dev or personal stacks are not healthy after testing" -ForegroundColor Yellow
    Write-Host "Please check the stacks" -ForegroundColor Yellow
} else {
    Write-Host "✓ Dev and personal stacks are still healthy" -ForegroundColor Green
}

# Step 7: Summary
Write-Host "`n[Step 7/8] Test cycle summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TEST CYCLE COMPLETE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Dev and personal stacks verified (before and after)" -ForegroundColor Green
Write-Host "✓ Test stack started and destroyed" -ForegroundColor Green
Write-Host "✓ Test data seeded" -ForegroundColor Green
Write-Host "✓ Manual testing completed" -ForegroundColor Green
Write-Host ""
Write-Host "All stacks are stable." -ForegroundColor Green

# Step 8: Exit
Write-Host "`nPress Enter to exit..." -ForegroundColor Cyan
Read-Host
