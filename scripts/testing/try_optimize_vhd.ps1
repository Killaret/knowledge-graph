Write-Output "Trying Optimize-VHD on Docker files..."

$files = @(
    "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx",
    "C:\Users\89209\AppData\Local\Docker\wsl\main\ext4.vhdx"
)

foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Output "File: $file"
        $oldSize = (Get-Item $file).Length / 1GB
        Write-Output "Current size: $([math]::Round($oldSize,2)) GB"
        
        try {
            Write-Output "Running Optimize-VHD..."
            Optimize-VHD -Path $file -Mode Full
            $newSize = (Get-Item $file).Length / 1GB
            $saved = $oldSize - $newSize
            Write-Output "SUCCESS: $([math]::Round($oldSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"
        } catch {
            Write-Output "FAILED: $_"
        }
    } else {
        Write-Output "File not found: $file"
    }
}

Write-Output "Done."
