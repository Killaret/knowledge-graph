
#!/usr/bin/env pwsh
# Backfill recommendations for personal stack
# Runs CLI for mass recalculation of recommendations

param(
    [switch]$DryRun
)

Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Backfill Recommendations - Personal Stack" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

if ($DryRun) {
    Write-Host "`nDRY RUN MODE - tasks will not be created" -ForegroundColor Yellow
}

# Rebuild image with CLI
Write-Host "`n[1/3] Rebuilding backend image with CLI..." -ForegroundColor Green
docker compose -f docker-compose.personal.yml build backend_personal

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to build backend" -ForegroundColor Red
    exit 1
}

# Run CLI
Write-Host "`n[2/3] Running CLI..." -ForegroundColor Green
$cliArgs = @("-f", "docker-compose.personal.yml", "run", "--rm")
if ($DryRun) {
    $cliArgs += @("--env", "DRY_RUN=1")
}
$cliArgs += @("cli_personal", "./cli")
if ($DryRun) {
    $cliArgs += "--dry-run"
}

docker compose @cliArgs

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: CLI failed" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3/3] Done!" -ForegroundColor Green
Write-Host "`nRecommendations are being computed in background." -ForegroundColor Cyan
Write-Host "Monitor worker logs:" -ForegroundColor Cyan
Write-Host "  docker compose -f docker-compose.personal.yml logs -f kg-worker-personal" -ForegroundColor White

