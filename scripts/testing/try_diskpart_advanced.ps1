# Try diskpart shrink + compact combination

Write-Output "=== Trying diskpart shrink + compact ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original VHDX size: $([math]::Round($originalSize,2)) GB"

# Multi-step diskpart script
$diskpartScript = @"
select vdisk file="$vhdxFile"
attach vdisk
select partition 1
shrink querymax
shrink querymax
detach vdisk
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_query.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Checking diskpart shrink querymax..."
try {
    diskpart /s $scriptPath
    Write-Output "Shrink querymax completed"
} catch {
    Write-Output "Shrink querymax failed: $_"
}

# Try resize with diskpart
$resizeScript = @"
select vdisk file="$vhdxFile"
attach vdisk
create partition primary
format quick fs=ntfs label="DockerData"
detach vdisk
exit
"@

Write-Output "WARNING: This would reformat the VHDX - NOT RECOMMENDED"
Write-Output "Skipping resize approach"

# Try simple compact again
$compactScript = @"
select vdisk file="$vhdxFile"
attach vdisk
compact vdisk
detach vdisk
exit
"@

$compactPath = "C:\Users\89209\AppData\Local\Temp\diskpart_compact.txt"
$compactScript | Out-File -FilePath $compactPath -Encoding ASCII

Write-Output "Trying diskpart compact again..."
try {
    diskpart /s $compactPath
    Write-Output "Second diskpart compact completed"
    
    $newSize = (Get-Item $vhdxFile).Length / 1GB
    $saved = $originalSize - $newSize
    Write-Output "After second compact: $([math]::Round($originalSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"
} catch {
    Write-Output "Second compact failed: $_"
}

Remove-Item $scriptPath -ErrorAction SilentlyContinue
Remove-Item $compactPath -ErrorAction SilentlyContinue

Write-Output "=== Diskpart attempts completed ==="