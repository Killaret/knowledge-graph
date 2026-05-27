# VHDX Compression via DiskPart (Admin required)
# With automatic WSL shutdown and file unlock

$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'

Write-Output "=== VHDX Compression via DiskPart ==="
Write-Output "File: $vhdxFile"
Write-Output ""

# Check file exists
if (-not (Test-Path $vhdxFile)) {
    Write-Output "ERROR: File not found: $vhdxFile"
    exit 1
}

# Get original size
$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original size: $($originalSize.ToString('F2')) GB"
Write-Output ""

# Step 1: Stop WSL completely
Write-Output "Step 1: Stopping WSL completely..."
wsl --shutdown
Start-Sleep -Seconds 10

# Force kill any remaining WSL processes
Write-Output "Step 2: Checking for remaining WSL processes..."
$wslProcesses = Get-Process | Where-Object { $_.ProcessName -like "*wsl*" -or $_.ProcessName -like "*vmmem*" }
if ($wslProcesses) {
    Write-Output "Found WSL processes, stopping them..."
    $wslProcesses | Stop-Process -Force
    Start-Sleep -Seconds 5
} else {
    Write-Output "No WSL processes found"
}

# Step 3: Check if file is locked
Write-Output "Step 3: Checking if VHDX file is locked..."
try {
    $stream = [System.IO.File]::Open($vhdxFile, 'Open', 'ReadWrite', 'None')
    $stream.Close()
    Write-Output "File is unlocked - ready for compression"
} catch {
    Write-Output "ERROR: File is still locked: $($_.Exception.Message)"
    Write-Output "Please close any applications using Docker and try again"
    exit 1
}

Write-Output ""

# Create diskpart script (same commands as terminal)
$diskpartScript = @"
select vdisk file="$vhdxFile"
attach vdisk readonly
compact vdisk
detach vdisk
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_compress.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Step 4: Running diskpart with compression commands..."
Write-Output "Commands:"
Write-Output "  select vdisk file='$vhdxFile'"
Write-Output "  attach vdisk readonly"
Write-Output "  compact vdisk"
Write-Output "  detach vdisk"
Write-Output ""

# Run diskpart
& diskpart /s $scriptPath

Write-Output ""
Start-Sleep -Seconds 3

# Check new size
$newSize = (Get-Item $vhdxFile).Length / 1GB
$saved = $originalSize - $newSize

Write-Output "=== Compression Results ==="
Write-Output "Before: $($originalSize.ToString('F2')) GB"
Write-Output "After:  $($newSize.ToString('F2')) GB"
Write-Output "Saved:  $($saved.ToString('F2')) GB"

if ($saved -gt 0) {
    Write-Output "SUCCESS: VHDX compressed successfully!"
    $percentSaved = ($saved / $originalSize) * 100
    Write-Output "Efficiency: $($percentSaved.ToString('F1'))% reduction"
} else {
    Write-Output "No size change - file may already be optimized"
}

Write-Output ""
Write-Output "Step 5: Restart WSL"
Write-Output "Run: wsl"