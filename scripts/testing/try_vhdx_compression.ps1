# Try proper VHDX compression methods

Write-Output "=== Trying VHDX compression methods ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original size: $([math]::Round($originalSize,2)) GB"

# Method 1: WSL optimize (for WSL2 VHDX)
Write-Output "Trying WSL2 optimization..."
try {
    wsl --optimize
    Write-Output "WSL optimize completed"
    Start-Sleep -Seconds 3
} catch {
    Write-Output "WSL optimize failed: $_"
}

$newSize = (Get-Item $vhdxFile).Length / 1GB
$saved = $originalSize - $newSize
Write-Output "After WSL optimize: $([math]::Round($originalSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"

# Method 2: Try qemu-img if available
Write-Output "Checking for qemu-img..."
if (Get-Command qemu-img -ErrorAction SilentlyContinue) {
    Write-Output "qemu-img found, trying compression..."
    $compressedFile = "$vhdxFile.compressed"
    
    try {
        # Try to convert with compression
        qemu-img convert -O vhdx -o suballocation=on "$vhdxFile" "$compressedFile"
        
        if (Test-Path $compressedFile) {
            $compressedSize = (Get-Item $compressedFile).Length / 1GB
            Write-Output "Compressed file: $([math]::Round($compressedSize,2)) GB"
            
            if ($compressedSize -lt $originalSize) {
                Write-Output "SUCCESS: Would save $([math]::Round($originalSize - $compressedSize,2)) GB"
                Write-Output "To use: replace original with compressed file"
            } else {
                Write-Output "No improvement with qemu-img"
                Remove-Item $compressedFile
            }
        }
    } catch {
        Write-Output "qemu-img compression failed: $_"
    }
} else {
    Write-Output "qemu-img not available"
}

# Method 3: Try VHD resize (shrinking)
Write-Output "Trying VHD shrink..."
try {
    # This requires Hyper-V, try anyway
    Resize-VHD -Path $vhdxFile -ToMinimumSize
    Write-Output "Resize-VHD completed"
    
    $finalSize = (Get-Item $vhdxFile).Length / 1GB
    $finalSaved = $originalSize - $finalSize
    Write-Output "After resize: $([math]::Round($originalSize,2)) GB -> $([math]::Round($finalSize,2)) GB (saved: $([math]::Round($finalSaved,2)) GB)"
} catch {
    Write-Output "Resize-VHD failed: $_"
}

Write-Output "=== Compression attempts completed ==="