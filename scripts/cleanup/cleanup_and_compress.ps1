Write-Output "=== Docker Cleanup and VHDX Compression ==="
Write-Output ""

Write-Output "Step 1: Stop WSL and Docker..."
wsl --shutdown
Start-Sleep -Seconds 5

Write-Output "Step 2: Clean Docker (if Docker is running)"
try {
    docker system prune -a --volumes --force
} catch {
    Write-Output "Docker not running or docker command failed - continuing"
}

Write-Output ""
Write-Output "Step 3: Check VHDX size before compression..."
$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'
$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original size: $($originalSize.ToString('F2')) GB"

Write-Output ""
Write-Output "Step 4: Run diskpart compression"
Write-Output "Please run manually in elevated diskpart:"
Write-Output "  select vdisk file='$vhdxFile'"
Write-Output "  attach vdisk readonly"
Write-Output "  compact vdisk"
Write-Output "  detach vdisk"
Write-Output ""
Write-Output "Or run: diskpart /s scripts\diskpart_vdisk.txt (as admin)"

Write-Output ""
Write-Output "Step 5: Restart WSL"
Write-Output "Run: wsl"