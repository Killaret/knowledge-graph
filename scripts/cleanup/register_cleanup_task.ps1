<#
Register a Windows Scheduled Task to run the clean_and_compress_lunix PowerShell script.
Usage:
  .\register_cleanup_task.ps1 -Daily -Time "03:00" -TaskName "KG-CleanLunix" -Force
  .\register_cleanup_task.ps1 -Weekly -Day "Sunday" -Time "03:00" -Force
Notes:
- This script uses schtasks to register a task running as SYSTEM (no password required).
- Requires Administrator privileges to register a SYSTEM task.
#>
param(
    [switch]$Daily = $false,
    [switch]$Weekly = $false,
    [string]$Day = "Sunday",
    [string]$Time = "03:00",
    [string]$TaskName = "KG-CleanLunix",
    [switch]$Force = $false,
    [switch]$DryRun = $false
)

$scriptPath = Join-Path (Resolve-Path -Path $PSScriptRoot).Path '..\scripts\clean_and_compress_lunix.ps1'
$scriptPath = (Resolve-Path $scriptPath).Path
$action = "powershell -NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`" -Search -Compress -DryRun"

if ($DryRun) {
    Write-Host "DRY RUN: would register scheduled task with action:`n$action`nTrigger: $(if ($Daily) { 'Daily' } elseif ($Weekly) { "Weekly on $Day" } else { 'Daily' })" -ForegroundColor Cyan
    return
}

if (-not (Test-Path $scriptPath)) {
    Write-Error "Script not found: $scriptPath"
    exit 1
}

if ($Force) {
    Write-Host "Removing existing task (if any) $TaskName" -ForegroundColor Yellow
    schtasks /Delete /TN $TaskName /F | Out-Null
}

if ($Daily) {
    $schtaskCmd = "schtasks /Create /SC DAILY /TN $TaskName /TR \"$action\" /ST $Time /RU SYSTEM /RL HIGHEST"
} elseif ($Weekly) {
    $dayUpper = $Day.ToUpper()
    $schtaskCmd = "schtasks /Create /SC WEEKLY /D $dayUpper /TN $TaskName /TR \"$action\" /ST $Time /RU SYSTEM /RL HIGHEST"
} else {
    # default to daily
    $schtaskCmd = "schtasks /Create /SC DAILY /TN $TaskName /TR \"$action\" /ST $Time /RU SYSTEM /RL HIGHEST"
}

Write-Host "Registering scheduled task: $TaskName" -ForegroundColor Cyan
try {
    $out = cmd /c $schtaskCmd
    Write-Host $out -ForegroundColor Gray
    Write-Host "Task registered." -ForegroundColor Green
} catch {
    Write-Error "Failed to register task. You may need to run this script as Administrator. Error: $_"
    exit 1
}
