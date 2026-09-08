# CLEAN-1 regression test: the Docker cleanup scripts must report failures
# honestly and exit non-zero when any step fails.
#
# Cases:
#   1. Docker unreachable (stub exits 1) -> [FAIL] in output, exit code != 0.
#      This is the mutation anchor for the spec's criterion 8: if the script
#      goes back to try/catch without checking the exit code, the stub's
#      failure is swallowed, output shows [SUCCESS]/[PASS]-only and the exit
#      code is 0 — this test goes red.
#   2. -DryRun previews steps and removes nothing.
#   3. -RemoveVolumes without a fresh backup refuses (phase FAIL, exit != 0);
#      with a fresh non-empty backup the gate passes.
#
# Runs against a stub `docker` on PATH so no real Docker is touched.

$ErrorActionPreference = "Continue"
$repoRoot = (git rev-parse --show-toplevel 2>$null).Trim()
if (-not $repoRoot) { $repoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path }

$script:Failures = 0
function Check {
    param([string]$Name, [bool]$Ok, [string]$Detail = "")
    if ($Ok) { Write-Host "  [PASS] $Name" -ForegroundColor Green }
    else { $script:Failures++; Write-Host "  [FAIL] $Name — $Detail" -ForegroundColor Red }
}

$stubDir = Join-Path $env:TEMP "kg-cleanup-stub-$(Get-Random)"
New-Item -ItemType Directory -Path $stubDir -Force | Out-Null
$cleanupPs1 = Join-Path $repoRoot "scripts\cleanup\cleanup-docker.ps1"

# Stub docker that always fails, like a stopped daemon.
@'
@echo off
echo error during connect: docker daemon is not running 1>&2
exit /b 1
'@ | Set-Content (Join-Path $stubDir "docker.cmd") -Encoding ASCII

$env:PATH = "$stubDir;$env:PATH"

Write-Host "Case 1: docker unreachable -> failures reported" -ForegroundColor Cyan
$out = pwsh -NoProfile -File $cleanupPs1 2>&1 | Out-String
$code = $LASTEXITCODE
Check "ps1 exits non-zero when docker is down" ($code -ne 0) "exit=$code"
Check "ps1 prints [FAIL] not [SUCCESS]" (($out -match '\[FAIL\]') -and ($out -notmatch '\[SUCCESS\]')) "exit=$code"

# Stub docker that works: everything returns empty success, `system df` prints a table.
@'
@echo off
if "%1"=="system" (
    echo TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
    echo Images          0         0         0B        0B
    echo Build Cache     0                   0B        0B
) 
exit /b 0
'@ | Set-Content (Join-Path $stubDir "docker.cmd") -Encoding ASCII

Write-Host "Case 2: -DryRun previews and changes nothing" -ForegroundColor Cyan
$out = pwsh -NoProfile -File $cleanupPs1 -DryRun 2>&1 | Out-String
$code = $LASTEXITCODE
Check "ps1 dry-run exits zero" ($code -eq 0) "exit=$code`n$out"
Check "ps1 dry-run marks steps skipped" ($out -match 'dry-run') "no dry-run markers`n$out"
Check "ps1 dry-run does not call prune/rm/stop" ($out -notmatch 'Removed') "unexpected removal output`n$out"

Write-Host "Case 3: -RemoveVolumes gated by backup freshness" -ForegroundColor Cyan
$backupDir = Join-Path $repoRoot "backups"
$fakeBackup = Join-Path $backupDir "backup-personal-clean1-test.tar"
$hadDir = Test-Path $backupDir
try {
    $out = pwsh -NoProfile -File $cleanupPs1 -RemoveVolumes 2>&1 | Out-String
    $code = $LASTEXITCODE
    Check "ps1 -RemoveVolumes without backup refuses" ($code -ne 0 -and $out -match 'backup') "exit=$code`n$out"

    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    "fake" | Set-Content $fakeBackup -Encoding ASCII
    $out = pwsh -NoProfile -File $cleanupPs1 -RemoveVolumes 2>&1 | Out-String
    $code = $LASTEXITCODE
    Check "ps1 -RemoveVolumes with fresh backup proceeds" ($code -eq 0 -and $out -match 'volume-cleanup') "exit=$code`n$out"
} finally {
    Remove-Item $fakeBackup -ErrorAction SilentlyContinue
    if (-not $hadDir) { Remove-Item $backupDir -Recurse -Force -ErrorAction SilentlyContinue }
}

# Bash version, when available — same contract. Prefer Git-bash: it
# understands Windows paths; a bare `bash` here may be WSL.
$bash = $null
$gitBash = "C:\Program Files\Git\bin\bash.exe"
if (Test-Path $gitBash) {
    $bash = $gitBash
} elseif (Get-Command bash -ErrorAction SilentlyContinue) {
    $bash = "bash"
}
if ($bash) {
    @'
#!/bin/sh
echo "error during connect: docker daemon is not running" >&2
exit 1
'@ | Set-Content (Join-Path $stubDir "docker") -Encoding ASCII -NoNewline
    Write-Host "Case 4 (sh): docker unreachable -> failures reported" -ForegroundColor Cyan
    $shScript = (Join-Path $repoRoot "scripts/cleanup/cleanup-docker.sh") -replace '\\', '/'
    $out = & $bash $shScript 2>&1 | Out-String
    $code = $LASTEXITCODE
    Check "sh exits non-zero when docker is down" ($code -ne 0) "exit=$code"
    Check "sh prints [FAIL]" ($out -match '\[FAIL\]') "no [FAIL]`n$out"
} else {
    Write-Host "  [SKIP] bash not available — sh cases skipped" -ForegroundColor Yellow
}

Remove-Item $stubDir -Recurse -Force -ErrorAction SilentlyContinue

Write-Host ""
if ($script:Failures -eq 0) {
    Write-Host "CLEAN-1 honesty test: ALL PASSED" -ForegroundColor Green
    exit 0
} else {
    Write-Host "CLEAN-1 honesty test: $($script:Failures) FAILED" -ForegroundColor Red
    exit 1
}
