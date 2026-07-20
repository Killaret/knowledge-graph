# Cleanup temporary test artifacts
param(
    [string]$ProjectRoot = (Join-Path $PSScriptRoot ..)
)

$ProjectRoot = Resolve-Path $ProjectRoot
$removed = 0

$patterns = @(
    "backend/coverage.out",
    "backend/*.cov",
    "frontend/coverage",
    "logs/test-outputs/*.log",
    "scripts/testing/temp/snapshots/*"
)

foreach ($pattern in $patterns) {
    $fullPattern = Join-Path $ProjectRoot $pattern
    $items = Get-ChildItem -Path $fullPattern -Recurse -Force -ErrorAction SilentlyContinue
    foreach ($item in $items) {
        Remove-Item $item.FullName -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "  Removed $($item.FullName)" -ForegroundColor DarkGray
        $removed++
    }
}

Write-Host "  ✓ Temporary test artifacts cleaned ($removed items)" -ForegroundColor Green
