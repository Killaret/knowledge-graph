# Force stop all Docker processes
Write-Output "=== Force stopping all Docker processes ==="

$processesToStop = @("Docker Desktop", "com.docker.build", "docker-sandbox")

foreach ($procName in $processesToStop) {
    $procs = Get-Process -Name $procName -ErrorAction SilentlyContinue
    if ($procs) {
        Write-Output "Stopping $procName (found $($procs.Count) instances)..."
        Stop-Process -Name $procName -Force
        Start-Sleep -Seconds 2
    } else {
        Write-Output "$procName not running"
    }
}

# Also try by killing service
Write-Output "Stopping Docker service..."
try {
    Stop-Service -Name "com.docker.service" -Force -ErrorAction SilentlyContinue
} catch {
    Write-Output "Docker service not found or error stopping"
}

# WSL shutdown
Write-Output "WSL shutdown..."
wsl --shutdown
Start-Sleep -Seconds 5

Write-Output "=== Checking if Docker processes are still running ==="
$remainingProcesses = Get-Process | Where-Object {
    $_.ProcessName -like "*docker*" -and $_.ProcessName -notlike "*Code*"
}

if ($remainingProcesses) {
    Write-Output "Still running Docker processes:"
    $remainingProcesses | Format-Table ProcessName, Id
} else {
    Write-Output "SUCCESS: All Docker processes stopped"
}

Write-Output "=== Checking file access ==="
$testFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"
if (Test-Path $testFile) {
    try {
        $file = [System.IO.File]::Open($testFile, 'Open', 'Read', 'ReadWrite')
        $file.Close()
        Write-Output "SUCCESS: File is accessible for compression"
    } catch {
        Write-Output "FAILED: File still locked: $_"
    }
} else {
    Write-Output "File not found"
}