# Clean and Compress "lunix" Image (Windows PowerShell)
# Usage: .\clean_and_compress_lunix.ps1 [-Compress] [-ImagePath <path>] [-Search] [-Force]

param(
    [switch]$Compress = $false,
    [string]$ImagePath = $null,
    [switch]$Search = $false,
    [switch]$Force = $false,
    [switch]$DryRun = $false
)

$ErrorActionPreference = "Continue"

Write-Host "🧰 Clean & Compress lunix" -ForegroundColor Cyan
Write-Host "$(Get-Date -Format 'HH:mm:ss') Starting..." -ForegroundColor Gray

function Write-Status { param([string]$Message, [string]$Status = "INFO")
    $color = switch ($Status) {
        "SUCCESS" { "Green" }
        "WARNING" { "Yellow" }
        "ERROR" { "Red" }
        default { "Cyan" }
    }
    Write-Host "  [$Status] $Message" -ForegroundColor $color
}

# 1. (Optional) Reuse docker cleanup steps if needed
Write-Host "\n1️⃣ Running lightweight Docker cleanup (dangling images, stopped containers)" -ForegroundColor Cyan
if ($DryRun) {
    Write-Status "DRY RUN: would run 'docker image prune -f'" "INFO"
    Write-Status "DRY RUN: would run 'docker container prune -f'" "INFO"
} else {
    try {
        docker image prune -f 2>$null | Out-Null
        docker container prune -f 2>$null | Out-Null
        Write-Status "Docker prune (images & containers) executed" "SUCCESS"
    } catch {
        Write-Status "Docker prune failed or Docker not available" "WARNING"
    }
}

# 2. Find lunix image if requested
$foundImage = $null
if ($ImagePath) {
    if (Test-Path $ImagePath) { $foundImage = Get-Item $ImagePath }
    else { Write-Status "Provided ImagePath not found: $ImagePath" "ERROR" }
} elseif ($Search) {
    Write-Host "\n2️⃣ Searching common locations for 'lunix' image..." -ForegroundColor Cyan
    $candidates = @(
        Join-Path $env:USERPROFILE 'lunix.vhdx',
        Join-Path $env:USERPROFILE 'lunix.img',
        Join-Path $env:LOCALAPPDATA 'Docker\wsl\lunix.vhdx',
        Join-Path $env:LOCALAPPDATA 'Docker\wsl\disk\lunix.vhdx',
        'D:\lunix.vhdx',
        'D:\images\lunix.vhdx'
    )
    foreach ($p in $candidates) {
        $it = Get-Item $p -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($it) { $foundImage = $it; break }
    }
    if (-not $foundImage) {
        # fallback: search for files with 'lunix' in name under user profile
        $search = Get-ChildItem -Path $env:USERPROFILE -Filter '*lunix*' -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($search) { $foundImage = $search }
    }
    if ($foundImage) { Write-Status "Found image: $($foundImage.FullName)" "INFO" }
    else { Write-Status "No lunix image found in common locations" "WARNING" }
}

# 3. Compress (if requested and found)
if ($Compress) {
    if (-not $foundImage) { Write-Status "No image available to compress" "ERROR" }
    else {
        $file = $foundImage.FullName
        $oldSize = [math]::Round((Get-Item $file).Length / 1GB, 2)
        Write-Host "\n3️⃣ Compressing image: $file" -ForegroundColor Cyan

        if ((Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -ErrorAction SilentlyContinue).State -eq "Enabled") {
            try {
                if ($DryRun) {
                    Write-Status "DRY RUN: would run Optimize-VHD -Path $file -Mode Full" "INFO"
                } else {
                    if ($Force -or (Read-Host "Proceed with Optimize-VHD on $file? (y/N)") -match '^[yY]') {
                        Write-Status "Running Optimize-VHD (Full)..." "INFO"
                        Optimize-VHD -Path $file -Mode Full -ErrorAction Stop | Out-Null
                        $newSize = [math]::Round((Get-Item $file).Length / 1GB, 2)
                        $saved = [math]::Round($oldSize - $newSize, 2)
                        Write-Status "Compression complete: $oldSize GB -> $newSize GB (saved: $saved GB)" "SUCCESS"
                    } else { Write-Status "Compression aborted by user" "WARNING" }
                }
            } catch {
                Write-Status "Compression failed (Optimize-VHD error)" "ERROR"
            }
        } else {
            Write-Status "Hyper-V feature not enabled. Cannot use Optimize-VHD." "ERROR"
        }
    }
}

Write-Host "\n✅ Done." -ForegroundColor Green
Write-Host "$(Get-Date -Format 'HH:mm:ss') Finished." -ForegroundColor Gray

Write-Host "\nUsage examples:" -ForegroundColor Cyan
Write-Host "  .\clean_and_compress_lunix.ps1 -Search -Compress    # Find and compress lunix image" -ForegroundColor Gray
Write-Host "  .\clean_and_compress_lunix.ps1 -ImagePath 'D:\images\lunix.vhdx' -Compress -Force  # Compress given image without prompt" -ForegroundColor Gray