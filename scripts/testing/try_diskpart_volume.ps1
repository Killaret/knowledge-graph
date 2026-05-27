# Check volume and use diskpart correctly

Write-Output "=== Checking volume for VHDX ==="

$vhdxFile = "C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx"

if (-not (Test-Path $vhdxFile)) {
    Write-Output "File not found: $vhdxFile"
    exit
}

# Get volume letter
$volLetter = (Get-Item $vhdxFile).PSDrive.Name
Write-Output "VHDX is on volume: $volLetter"

$originalSize = (Get-Item $vhdxFile).Length / 1GB
Write-Output "Original VHDX size: $([math]::Round($originalSize,2)) GB"

# Correct diskpart script as user described:
# 1. Select volume
# 2. Block/Select partition  
# 3. Shrink
# 4. Compact

$diskpartScript = @"
select volume $volLetter
select partition 1
shrink
compact
exit
"@

$scriptPath = "C:\Users\89209\AppData\Local\Temp\diskpart_volume.txt"
$diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII

Write-Output "Running diskpart on volume $volLetter (as you described)..."
try {
    diskpart /s $scriptPath
    Write-Output "Diskpart completed"
    
    $newSize = (Get-Item $vhdxFile).Length / 1GB
    $saved = $originalSize - $newSize
    Write-Output "After diskpart: $([math]::Round($originalSize,2)) GB -> $([math]::Round($newSize,2)) GB (saved: $([math]::Round($saved,2)) GB)"
    
    if ($saved -gt 0.1) {
        Write-Output "SUCCESS! Diskpart compressed the volume and VHDX!"
    } else {
        Write-Output "No compression with this method"
    }
} catch {
    Write-Output "Diskpart failed: $_"
}

Remove-Item $scriptPath -ErrorAction SilentlyContinue

Write-Output "=== Volume diskpart attempt completed ==="