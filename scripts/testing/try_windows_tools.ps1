# Try Windows system optimization tools

Write-Output "=== Trying Windows disk optimization tools ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original size: $([math]::Round($originalSize,2)) GB"

# Method 1: Compact.exe with different parameters
Write-Output "Trying Compact.exe with different parameters..."
try {
    # Try compact with C (compress) and I (index)
    compact /C /I "$vhdxFile"
    Write-Output "Compact /C /I completed"
    
    $size1 = (Get-Item $vhdxFile).Length / 1GB
    Write-Output "After compact /C /I: $([math]::Round($size1,2)) GB"
} catch {
    Write-Output "Compact /C /I failed: $_"
}

# Method 2: Defragment the volume
Write-Output "Trying volume defragmentation..."
try {
    $volume = (Get-Item $vhdxFile).PSDrive.Name
    Write-Output "Defragmenting volume $volume"
    defrag $volume /O
    Write-Output "Defrag completed"
    
    $size2 = (Get-Item $vhdxFile).Length / 1GB
    Write-Output "After defrag: $([math]::Round($size2,2)) GB"
} catch {
    Write-Output "Defrag failed: $_"
}

# Method 3: sdelete to zero free space (helps VHDX compression)
Write-Output "Checking for sdelete..."
$sdeletePath = "C:\Users\89209\AppData\Local\Programs\sysinternals-suite"
if (Test-Path "$sdeletePath\sdelete64.exe") {
    Write-Output "sdelete found, trying to zero free space on Docker volume"
    # Get Docker volume path
    $wslDistros = wsl.exe -l -v 2>$null
    Write-Output "WSL distros: $wslDistros"
} else {
    Write-Output "sdelete not found at $sdeletePath"
}

# Method 4: TRIM on SSD
Write-Output "Trying TRIM on the volume..."
try {
    $volume = (Get-Item $vhdxFile).PSDrive.Name
    Optimize-Volume -DriveLetter $volume -Trim
    Write-Output "TRIM completed"
    
    $size3 = (Get-Item $vhdxFile).Length / 1GB
    Write-Output "After TRIM: $([math]::Round($size3,2)) GB"
} catch {
    Write-Output "TRIM failed: $_"
}

$finalSize = (Get-Item $vhdxFile).Length / 1GB
$totalSaved = $originalSize - $finalSize

Write-Output "=== Final result: $([math]::Round($originalSize,2)) GB -> $([math]::Round($finalSize,2)) GB (saved: $([math]::Round($totalSaved,2)) GB) ==="