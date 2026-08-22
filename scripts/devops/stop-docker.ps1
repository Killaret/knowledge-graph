# Properly stop Docker Desktop and WSL
Write-Output "Stopping Docker Desktop and WSL processes..."

try {
    # Stop Docker Desktop process
    $dockerProcess = Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue
    if ($dockerProcess) {
        Write-Output "Stopping Docker Desktop..."
        Stop-Process -Name "Docker Desktop" -Force
        Start-Sleep -Seconds 3
    } else {
        Write-Output "Docker Desktop not running"
    }
    
    # Stop all docker-related processes
    $dockerProcesses = @("com.docker.backend", "com.docker.service", "Vmmem", "Wsl", "dockerd")
    foreach ($procName in $dockerProcesses) {
        $procs = Get-Process -Name $procName -ErrorAction SilentlyContinue
        if ($procs) {
            Write-Output "Stopping $procName..."
            Stop-Process -Name $procName -Force
        }
    }
    
    # Shutdown WSL completely
    Write-Output "Shutting down WSL..."
    wsl --shutdown
    Start-Sleep -Seconds 5
    
    Write-Output "Docker and WSL stopped successfully"
} catch {
    Write-Output "Error stopping processes: $_"
}

Write-Output "Checking if files are still locked..."
$testFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"
if (Test-Path $testFile) {
    try {
        $file = [System.IO.File]::Open($testFile, 'Open', 'Read', 'ReadWrite')
        $file.Close()
        Write-Output "File is accessible - ready for compression!"
    } catch {
        Write-Output "File is still locked: $_"
    }
} else {
    Write-Output "File not found"
}