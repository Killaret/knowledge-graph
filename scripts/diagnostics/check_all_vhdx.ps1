Write-Output "Checking all Docker VHDX files sizes..."
Write-Output ""

# Main WSL disk
$dockerData = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'
if (Test-Path $dockerData) {
    $size = (Get-Item $dockerData).Length / 1GB
    Write-Output "docker_data.vhdx: $($size.ToString('F2')) GB"
} else {
    Write-Output "docker_data.vhdx: NOT FOUND"
}

# Docker Desktop VM disk
$dockerDesktop = 'C:\ProgramData\DockerDesktop\vm-data\DockerDesktop.vhdx'
if (Test-Path $dockerDesktop) {
    $size = (Get-Item $dockerDesktop).Length / 1GB
    Write-Output "DockerDesktop.vhdx: $($size.ToString('F2')) GB"
} else {
    Write-Output "DockerDesktop.vhdx: NOT FOUND"
}

Write-Output ""
Write-Output "Check if Docker needs cleanup before compression for best results:"
Write-Output "1. Run: docker system prune -a --volumes"
Write-Output "2. Then run diskpart compression again"