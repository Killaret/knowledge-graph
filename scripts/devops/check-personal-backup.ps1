# Checks that a fresh, non-empty Personal-stack backup exists before a
# destructive Docker operation. Mirrors the rule in
# scripts/devops/guard-personal-data.py — the backup globs and the
# freshness threshold come from the same policy, so the two cannot drift.
#
# Exit code 0 — a usable backup exists (prints its path and age).
# Exit code 1 — no usable backup (prints the reason and where to put one).

$ErrorActionPreference = "Stop"

$repoRoot = (git rev-parse --show-toplevel 2>$null)
if (-not $repoRoot) { $repoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path }
$backupDir = Join-Path $repoRoot "backups"
$policyFile = Join-Path $PSScriptRoot "backup-policy.env"

# Threshold: env var first, then the shared policy file. Refuse rather than
# guess — a cleanup that cannot tell how fresh the backup must be is not safe.
$maxAgeHours = $env:KG_BACKUP_MAX_AGE_HOURS
if (-not $maxAgeHours) {
    if (-not (Test-Path $policyFile)) {
        Write-Host "  [ERROR] Backup policy file missing: $policyFile" -ForegroundColor Red
        exit 1
    }
    $line = Get-Content $policyFile | Where-Object { $_ -match '^\s*KG_BACKUP_MAX_AGE_HOURS\s*=' } | Select-Object -First 1
    $maxAgeHours = ($line -split '=', 2)[1].Trim()
}
if (-not $maxAgeHours) {
    Write-Host "  [ERROR] KG_BACKUP_MAX_AGE_HOURS not set in $policyFile" -ForegroundColor Red
    exit 1
}
$maxAgeHours = [double]$maxAgeHours

if (-not (Test-Path $backupDir)) {
    Write-Host "  [ERROR] No backup directory: $backupDir" -ForegroundColor Red
    Write-Host "          Run scripts/devops/backup-personal.ps1 first." -ForegroundColor Yellow
    exit 1
}

$patterns = @("backup-personal-*", "personal-volumes-raw-*")
$latest = $null
foreach ($pattern in $patterns) {
    $latest = @(Get-ChildItem -Path $backupDir -Filter $pattern -File -ErrorAction SilentlyContinue |
        Where-Object { $_.Length -gt 0 } |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1) + @($latest) |
        Sort-Object LastWriteTime -Descending | Select-Object -First 1
}

if (-not $latest) {
    Write-Host "  [ERROR] No non-empty backup matching backup-personal-* or personal-volumes-raw-* in $backupDir" -ForegroundColor Red
    Write-Host "          Run scripts/devops/backup-personal.ps1 first." -ForegroundColor Yellow
    exit 1
}

$ageHours = ((Get-Date) - $latest.LastWriteTime).TotalHours
if ($ageHours -gt $maxAgeHours) {
    Write-Host ("  [ERROR] Newest backup {0} is {1:N1} h old; allowed: {2:N0} h" -f $latest.Name, $ageHours, $maxAgeHours) -ForegroundColor Red
    Write-Host "          Make a fresh backup via scripts/devops/backup-personal.ps1." -ForegroundColor Yellow
    exit 1
}

Write-Host ("  [PASS] Backup {0} ({1:N0} KB, {2:N1} h old) is fresh enough" -f $latest.Name, ($latest.Length / 1KB), $ageHours) -ForegroundColor Green
exit 0
