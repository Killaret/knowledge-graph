# Try diskpart for VHDX compression

Write-Output "=== Trying diskpart for VHDX compression ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original VHDX size: $([math]::Round($originalSize,2)) GB"

# Create diskpart script
$diskpartScript = @"
select vdisk file="$vhdxFile"
attach vdisk
compact vdisk
detach vdisk
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_script.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Running diskpart script..."
try {
    diskpart /s $scriptPath
    Write-Output "Diskpart completed"
    
    $newSize = (Get-Item $vhdxFile).Length / 1GB
    $saved = $originalSize - $newSize
    Write-Output "After diskpart compact: $([math]::Round($originalSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"
    
    if ($saved -gt 0.1) {
        Write-Output "SUCCESS: Diskpart compressed the VHDX file!"
    } else {
        Write-Output "No compression with diskpart compact"
    }
} catch {
    Write-Output "Diskpart failed: $_"
}

# Clean up
Remove-Item $scriptPath -ErrorAction SilentlyContinue

Write-Output "=== Diskpart attempt completed ==="