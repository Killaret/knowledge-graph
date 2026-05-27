# Open diskpart interactively for manual execution

Write-Output "=== Opening diskpart for manual execution ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original VHDX size: $([math]::Round($originalSize,2)) GB"

# Create script that opens diskpart with file pre-selected
$diskpartScript = @"
select vdisk file="$vhdxFile"
REM User will manually run: compact vdisk
REM Then: detach vdisk
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_manual.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Opening diskpart interactively..."
Write-Output "File will be pre-selected. You'll need to manually run:"
Write-Output "  1. compact vdisk"
Write-Output "  2. detach vdisk"
Write-Output ""
Write-Output "After completion, run: wsl to restart WSL"

try {
    Start-Process "diskpart.exe" -ArgumentList "/s $scriptPath"
    Write-Output "Diskpart opened! Please run the commands manually."
    
    # Wait for user to complete (simulate monitoring)
    Start-Sleep -Seconds 5
    
    # Check if file size changed
    $newSize = (Get-Item $vhdxFile).Length / 1GB
    $saved = $originalSize - $newSize
    Write-Output "Current VHDX size: $([math]::Round($newSize,2)) GB"
    
    if ($saved -gt 0.1) {
        Write-Output "File was compressed! Saved: $([math]::Round($saved,2)) GB"
    } else {
        Write-Output "File size unchanged. Try running compact vdisk manually."
    }
} catch {
    Write-Output "Failed to open diskpart: $_"
}

Write-Output ""
Write-Output "Or run manually in diskpart:"
Write-Output "  1. diskpart"
Write-Output "  2. select vdisk file=$vhdxFile"  
Write-Output "  3. compact vdisk"
Write-Output "  4. detach vdisk"
Write-Output "  5. exit"
Write-Output "  6. wsl"