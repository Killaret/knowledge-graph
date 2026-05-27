param(
    [string]$ImagePath = $null,
    [switch]$DryRun = $false
)

Write-Output "=== Test Cleanup Script ==="
Write-Output "ImagePath: $ImagePath"
Write-Output "DryRun: $DryRun"

if ($ImagePath) {
    if (Test-Path $ImagePath) {
        Write-Output "Path exists: $ImagePath"
        $item = Get-Item $ImagePath
        Write-Output "Is directory: $($item.PSIsContainer)"
        
        if ($item.PSIsContainer) {
            Write-Output "Searching for files..."
            $files = Get-ChildItem -Path $ImagePath -File -Recurse -Force -ErrorAction SilentlyContinue
            Write-Output "Found $($files.Count) files"
            foreach ($f in $files) {
                $size = [math]::Round($f.Length/1GB,2)
                Write-Output "  - $($f.Name) ($size GB)"
            }
        } else {
            $size = [math]::Round($item.Length/1GB,2)
            Write-Output "File: $($item.Name) ($size GB)"
        }
    } else {
        Write-Output "Path does not exist: $ImagePath"
    }
} else {
    Write-Output "No path provided"
}

Write-Output "=== Test Complete ==="