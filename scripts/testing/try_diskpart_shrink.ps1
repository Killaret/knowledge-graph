# Try diskpart shrink for VHDX

Write-Output "=== Trying diskpart shrink for VHDX ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original VHDX size: $([math]::Round($originalSize,2)) GB"

# Try diskpart shrink
$diskpartScript = @"
select vdisk file="$vhdxFile"
attach vdisk
select partition 1
shrink desired=10240 minimum=10240
detach vdisk
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_shrink.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Running diskpart shrink..."
try {
    diskpart /s $scriptPath
    Write-Output "Diskpart shrink completed"
    
    $newSize = (Get-Item $vhdxFile).Length / 1GB
    $saved = $originalSize - $newSize
    Write-Output "After diskpart shrink: $([math]::Round($originalSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"
} catch {
    Write-Output "Diskpart shrink failed: $_"
}

Remove-Item $scriptPath -ErrorAction SilentlyContinue

Write-Output "=== Diskpart shrink attempt completed ==="