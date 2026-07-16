# Clean up generated log/replay/output files older than retention thresholds.
# Run from repo root.

param(
    [int]$LogRetentionDays = 30,
    [int]$SnapshotRetentionDays = 7,
    [switch]$DryRun
)

$repoRoot = Split-Path -Parent $PSScriptRoot | Split-Path -Parent
Set-Location $repoRoot

function Remove-IfOld {
    param(
        [Parameter(ValueFromPipeline)]$Item,
        [int]$Days,
        [string]$Label
    )
    process {
        if ($Item -is [System.IO.FileInfo]) {
            $age = (Get-Date) - $Item.LastWriteTime
            if ($age.Days -ge $Days -or $Days -eq 0) {
                if ($DryRun) {
                    Write-Host "[DRY-RUN] Would delete $Label`: $($Item.FullName) ($([math]::Round($age.TotalDays,1)) days old)"
                } else {
                    Remove-Item $Item.FullName -Force -ErrorAction SilentlyContinue
                    Write-Host "Deleted $Label`: $($Item.FullName)"
                }
            }
        } elseif ($Item -is [System.IO.DirectoryInfo]) {
            $age = (Get-Date) - $Item.LastWriteTime
            if ($age.Days -ge $Days -or $Days -eq 0) {
                if ($DryRun) {
                    Write-Host "[DRY-RUN] Would delete $Label`: $($Item.FullName) ($([math]::Round($age.TotalDays,1)) days old)"
                } else {
                    Remove-Item $Item.FullName -Recurse -Force -ErrorAction SilentlyContinue
                    Write-Host "Deleted $Label`: $($Item.FullName)"
                }
            }
        }
    }
}

Write-Host "Cleaning ad-hoc log files in project root..." -ForegroundColor Cyan
$rootPatterns = @('*.log', 'hs_err_pid*.log', 'replay_pid*.log', '*.results.txt')
foreach ($pattern in $rootPatterns) {
    Get-ChildItem -Path $repoRoot -Filter $pattern -File -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -notlike '*\logs\*' } |
        Remove-IfOld -Days 0 -Label 'root log'
}

Write-Host "Cleaning frontend generated logs..." -ForegroundColor Cyan
Get-ChildItem -Path (Join-Path $repoRoot 'frontend') -Filter '*.log' -File -ErrorAction SilentlyContinue |
    Remove-IfOld -Days 0 -Label 'frontend log'

Write-Host "Cleaning backend generated logs and test result files..." -ForegroundColor Cyan
Get-ChildItem -Path (Join-Path $repoRoot 'backend') -Include '*.log','*-test-results.txt' -File -ErrorAction SilentlyContinue |
    Remove-IfOld -Days 0 -Label 'backend artifact'

Write-Host "Cleaning old logs in logs/ (keeping .gitkeep and README)..." -ForegroundColor Cyan
Get-ChildItem -Path (Join-Path $repoRoot 'logs') -Recurse -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -ne '.gitkeep' -and $_.Name -notlike 'README*' } |
    Remove-IfOld -Days $LogRetentionDays -Label 'log file'

Write-Host "Cleaning old test snapshots in scripts/testing/temp/snapshots/..." -ForegroundColor Cyan
$snapshotsDir = [System.IO.Path]::Combine($repoRoot, 'scripts', 'testing', 'temp', 'snapshots')
if (Test-Path $snapshotsDir) {
    Get-ChildItem -Path $snapshotsDir -Directory -ErrorAction SilentlyContinue |
        Remove-IfOld -Days $SnapshotRetentionDays -Label 'snapshot dir'
}

Write-Host "Done." -ForegroundColor Green
