Write-Output "=== Checking Docker Disk Usage ==="
Write-Output ""

# Check if Docker is running
try {
    $dockerRunning = docker info 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Output "Docker is running"
        Write-Output ""

        Write-Output "=== Docker System Disk Usage ==="
        docker system df -v

        Write-Output ""
        Write-Output "=== Docker Images ==="
        docker images

        Write-Output ""
        Write-Output "=== Docker Containers ==="
        docker ps -a

        Write-Output ""
        Write-Output "=== Docker Volumes ==="
        docker volume ls

    } else {
        Write-Output "Docker is not running"
        Write-Output "Start Docker with: docker start"
    }
} catch {
    Write-Output "Docker command failed: $_"
    Write-Output "Docker may not be installed or not in PATH"
}

Write-Output ""
Write-Output "=== VHDX File Size ==="
$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'
if (Test-Path $vhdxFile) {
    $size = (Get-Item $vhdxFile).Length / 1GB
    Write-Output "docker_data.vhdx: $($size.ToString('F2')) GB"
}