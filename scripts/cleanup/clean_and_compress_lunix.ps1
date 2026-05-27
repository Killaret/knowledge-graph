# Clean and Compress Docker Images (Windows PowerShell)
# Usage: .\clean_and_compress_lunix.ps1 [-Compress] [-ImagePath <path>] [-Search] [-Force] [-DryRun] [-UseCompact]

param(
    [switch]$Compress = $false,
    [string]$ImagePath = $null,
    [switch]$Search = $false,
    [switch]$Force = $false,
    [switch]$DryRun = $false,
    [switch]$UseCompact = $false
)

$ErrorActionPreference = "Continue"

Write-Output "🧰 Clean & Compress Docker Images"
Write-Output "$(Get-Date -Format 'HH:mm:ss') Starting..."

function Write-Status { param([string]$Message, [string]$Status = "INFO")
    Write-Output "  [$Status] $Message"
}

# 1. Lightweight Docker cleanup
Write-Output ""
Write-Output "1️⃣ Running lightweight Docker cleanup"
if ($DryRun) {
    Write-Status "DRY RUN: would run 'docker image prune -f'"
    Write-Status "DRY RUN: would run 'docker container prune -f'"
} else {
    try {
        docker image prune -f 2>$null | Out-Null
        docker container prune -f 2>$null | Out-Null
        Write-Status "Docker prune executed" "SUCCESS"
    } catch {
        Write-Status "Docker prune failed or Docker not available" "WARNING"
    }
}

# 2. Find image(s)
$foundImages = @()

if ($ImagePath) {
    if (Test-Path $ImagePath) {
        $item = Get-Item $ImagePath
        if ($item.PSIsContainer) {
            Write-Status "Searching in directory: $($item.FullName)"
            $foundImages = Get-ChildItem -Path $ImagePath -File -Recurse -Force -ErrorAction SilentlyContinue
        } else {
            $foundImages = @($item)
        }
    }
    else { Write-Status "Provided ImagePath not found: $ImagePath" "ERROR" }
} elseif ($Search) {
    Write-Output ""
    Write-Output "2️⃣ Searching common locations for Docker images..."
    
    # Check Docker WSL directory
    $dockerWSL = "C:\Users\$env:USERNAME\AppData\Local\Docker\wsl"
    if (Test-Path $dockerWSL) {
        Write-Status "Searching in Docker WSL directory..."
        $foundImages = Get-ChildItem -Path $dockerWSL -File -Recurse -Force -ErrorAction SilentlyContinue
    }
    
    # Check specific files
    $specificFiles = @(
        "D:\docker_data.vhdx",
        "D:\ext4.vhdx",
        "D:\lunix.vhdx"
    )
    
    foreach ($filePath in $specificFiles) {
        if (Test-Path $filePath) {
            $foundImages += Get-Item $filePath
        }
    }
    
    if ($foundImages.Count -gt 0) {
        Write-Status "Total found: $($foundImages.Count) image(s)" "SUCCESS"
    } else { Write-Status "No Docker images found" "WARNING" }
}

# Remove duplicates
$foundImages = $foundImages | Sort-Object FullName -Unique

# 3. Compress if requested
if ($Compress -and $foundImages.Count -gt 0) {
    Write-Output ""
    Write-Output "3️⃣ Compressing found images..."
    
    foreach ($image in $foundImages) {
        $file = $image.FullName
        
        # Validate file
        if (-not $file -or -not (Test-Path $file)) {
            Write-Status "Skipping invalid path" "WARNING"
            continue
        }
        
        $oldSize = [math]::Round((Get-Item $file).Length / 1GB, 2)
        Write-Output ""
        Write-Output "Processing: $file"
        Write-Status "Original size: $oldSize GB"
        
        if ($DryRun) {
            Write-Status "DRY RUN: would compress $file with Compact.exe"
            Write-Status "DRY RUN: would enable sparse file optimization"
            continue
        }
        
        # Try Compact.exe
        try {
            if ($Force -or (Read-Host "Proceed with Compact.exe on $file? (y/N)") -match '^[yY]') {
                Write-Status "Running Compact.exe..."
                compact /C /F "$file" | Out-Null
                if ($LASTEXITCODE -eq 0) {
                    $newSize = [math]::Round((Get-Item $file).Length / 1GB, 2)
                    $saved = [math]::Round($oldSize - $newSize, 2)
                    Write-Status "Compact.exe complete: $oldSize GB -> $newSize GB (saved: $saved GB)" "SUCCESS"
                } else {
                    Write-Status "Compact.exe failed with exit code: $LASTEXITCODE" "WARNING"
                }
            } else {
                Write-Status "Skipped by user" "WARNING"
            }
        } catch {
            Write-Status "Compact.exe operation failed" "ERROR"
        }
        
        # Try sparse file
        try {
            Write-Status "Attempting sparse file optimization..."
            fsutil sparse setflag "$file" 1 2>$null | Out-Null
            Write-Status "Sparse flag enabled" "SUCCESS"
        } catch {
            Write-Status "Sparse operation failed" "WARNING"
        }
    }
} elseif ($Compress -and $foundImages.Count -eq 0) {
    Write-Status "No images available to compress" "ERROR"
}

Write-Output ""
Write-Output "✅ Done."
Write-Output "$(Get-Date -Format 'HH:mm:ss') Finished."

Write-Output ""
Write-Output "Usage examples:"
Write-Output "  .\clean_and_compress_lunix.ps1 -ImagePath 'C:\Users\USERNAME\AppData\Local\Docker\wsl' -Compress -UseCompact"
Write-Output "  .\clean_and_compress_lunix.ps1 -Search -Compress -Force"
Write-Output "  .\clean_and_compress_lunix.ps1 -ImagePath 'C:\Users\USERNAME\AppData\Local\Docker\wsl' -Compress -UseCompact -DryRun"