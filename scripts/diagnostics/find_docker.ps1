# Find all Docker processes
Write-Output "=== Finding Docker processes ==="
$dockerProcesses = Get-Process | Where-Object {
    $_.ProcessName -like "*docker*" -or 
    $_.MainWindowTitle -like "*docker*"
}

if ($dockerProcesses) {
    $dockerProcesses | Format-Table ProcessName, Id, MainWindowTitle
    Write-Output "Total Docker processes: $($dockerProcesses.Count)"
} else {
    Write-Output "No Docker processes found"
}