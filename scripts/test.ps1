# Unified test entry point for Knowledge Graph
# Usage: .\scripts\test.ps1 [-Target unit|integration|e2e|bdd|coverage|clean|all]
param(
    [ValidateSet('unit', 'integration', 'e2e', 'bdd', 'coverage', 'clean', 'all')]
    [string]$Target = 'all'
)

$scriptDir = $PSScriptRoot
$root = Join-Path $scriptDir .. | Resolve-Path

switch ($Target) {
    'unit' {
        Write-Host "Running backend unit tests..." -ForegroundColor Cyan
        Set-Location "$root\backend"
        go test ./... -count=1

        Write-Host "Running frontend unit tests..." -ForegroundColor Cyan
        Set-Location "$root\frontend"
        npm run test:unit
    }
    'integration' {
        Write-Host "Running backend integration tests (requires Linux/WSL Docker)..." -ForegroundColor Cyan
        Set-Location "$root\backend"
        go test -tags=integration ./... -count=1 -p=1
    }
    'e2e' {
        Write-Host "Running frontend E2E tests..." -ForegroundColor Cyan
        Set-Location "$root\frontend"
        npm run test
    }
    'bdd' {
        Write-Host "Running frontend BDD tests..." -ForegroundColor Cyan
        Set-Location "$root\frontend"
        npm run test:bdd
    }
    'coverage' {
        Write-Host "Generating backend coverage..." -ForegroundColor Cyan
        Set-Location "$root\backend"
        go test ./... -count=1 -coverprofile=.\coverage.out
        go tool cover -func .\coverage.out | Select-Object -Last 1

        Write-Host "Generating frontend coverage..." -ForegroundColor Cyan
        Set-Location "$root\frontend"
        npm run test:coverage
    }
    'clean' {
        & "$scriptDir\cleanup-test-artifacts.ps1" -ProjectRoot $root
    }
    'all' {
        & "$scriptDir\test.ps1" -Target unit
        & "$scriptDir\test.ps1" -Target integration
        & "$scriptDir\test.ps1" -Target e2e
        & "$scriptDir\test.ps1" -Target bdd
    }
}

Set-Location $root
