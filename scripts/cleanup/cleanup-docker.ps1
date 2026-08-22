# Docker Cleanup Script for Knowledge Graph (Windows with WSL2)
# Removes dangling images, stopped containers, unused networks, and build cache
# Optionally compresses WSL2 disk for efficient storage
# Usage: .\cleanup-docker.ps1 [-Full] [-WslOptimize]
# Note: -WslOptimize requires admin rights (diskpart). Hyper-V is NOT required.

param(
    [switch]$Full = $false,
    [switch]$WslOptimize = $false
)

$ErrorActionPreference = "Continue"

Write-Host "🧹 Knowledge Graph Docker Cleanup" -ForegroundColor Cyan
Write-Host "$(Get-Date -Format 'HH:mm:ss') Starting cleanup..." -ForegroundColor Gray

function Write-Status {
    param([string]$Message, [string]$Status = "INFO")
    $color = switch ($Status) {
        "SUCCESS" { "Green" }
        "WARNING" { "Yellow" }
        "ERROR" { "Red" }
        default { "Cyan" }
    }
    Write-Host "  [$Status] $Message" -ForegroundColor $color
}

# 1. Stop all containers
Write-Host "`n1️⃣ Stopping containers..." -ForegroundColor Cyan
try {
    $running = docker ps -q 2>$null
    if ($running) {
        docker stop $running 2>$null | Out-Null
        Write-Status "Stopped running containers" "SUCCESS"
    } else {
        Write-Status "No running containers" "INFO"
    }
} catch {
    Write-Status "Failed to stop containers" "WARNING"
}

# 2. Remove dangling images
Write-Host "`n2️⃣ Removing dangling images..." -ForegroundColor Cyan
try {
    docker image prune -f 2>$null | Out-Null
    Write-Status "Dangling images removed" "SUCCESS"
} catch {
    Write-Status "Failed to remove dangling images" "WARNING"
}

# 3. Remove stopped containers
Write-Host "`n3️⃣ Removing stopped containers..." -ForegroundColor Cyan
try {
    docker container prune -f 2>$null | Out-Null
    Write-Status "Stopped containers removed" "SUCCESS"
} catch {
    Write-Status "Failed to remove containers" "WARNING"
}

# 4. Remove unused networks
Write-Host "`n4️⃣ Removing unused networks..." -ForegroundColor Cyan
try {
    docker network prune -f 2>$null | Out-Null
    Write-Status "Unused networks removed" "SUCCESS"
} catch {
    Write-Status "Failed to remove networks" "WARNING"
}

# 5. Remove build cache
Write-Host "`n5️⃣ Clearing Docker build cache..." -ForegroundColor Cyan
try {
    docker builder prune -f 2>$null | Out-Null
    Write-Status "Build cache cleared" "SUCCESS"
} catch {
    Write-Status "Failed to clear build cache" "WARNING"
}

# 6. Remove unused volumes
Write-Host "`n6️⃣ Removing unused volumes..." -ForegroundColor Cyan
try {
    $protectedVolumes = docker volume ls --filter "label=com.knowledgegraph.protected=true" --format "{{.Name}}" 2>$null
    if ($protectedVolumes) {
        Write-Status "Protected volumes found (will be kept)" "INFO"
        docker volume prune --filter "label!=com.knowledgegraph.protected=true" -f 2>$null | Out-Null
    } else {
        docker volume prune -f 2>$null | Out-Null
    }
    Write-Status "Unused volumes removed" "SUCCESS"
} catch {
    Write-Status "Failed to remove volumes" "WARNING"
}

# 7. Full cleanup mode (optional)
if ($Full) {
    Write-Host "`n7️⃣ Full cleanup mode (removing ALL unused images)..." -ForegroundColor Cyan
    try {
        docker system prune -af --volumes 2>$null | Out-Null
        Write-Status "Full system cleanup completed" "SUCCESS"
    } catch {
        Write-Status "Full cleanup completed with warnings" "WARNING"
    }
}

# 8. WSL2 optimization (optional)
# Uses diskpart to compact the VHD - requires admin rights, but NOT Hyper-V.
if ($WslOptimize) {
    Write-Host "`n8️⃣ Optimizing WSL2 disk..." -ForegroundColor Cyan

    $wsl_check = wsl --list 2>$null
    if ($wsl_check) {
        Write-Status "Shutting down WSL..." "INFO"
        wsl --shutdown 2>$null
        Start-Sleep -Seconds 5
        # Force kill any leftover WSL/Docker processes that may lock the VHD
        $wslProcs = Get-Process | Where-Object { $_.ProcessName -like "*wsl*" -or $_.ProcessName -like "*vmmem*" -or $_.ProcessName -like "*docker*" }
        if ($wslProcs) {
            Write-Status "Stopping leftover WSL/Docker processes..." "INFO"
            $wslProcs | Stop-Process -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 3
        }
        Write-Status "WSL shut down successfully" "SUCCESS"

        # Find ALL VHDs and pick the largest one (docker_data.vhdx is the real target,
        # not the small main\ext4.vhdx utility disk)
        $allVhds = @()
        $allVhds += Get-ChildItem -Path (Join-Path $env:LOCALAPPDATA 'Docker\wsl') -Filter *.vhdx -Recurse -ErrorAction SilentlyContinue
        $allVhds += Get-ChildItem -Path (Join-Path $env:LOCALAPPDATA 'Packages') -Filter 'ext4.vhdx' -Recurse -ErrorAction SilentlyContinue
        $vhdx_file = $allVhds | Sort-Object Length -Descending | Select-Object -First 1

        if ($vhdx_file) {
            $old_size = [math]::Round($vhdx_file.Length / 1GB, 2)
            Write-Status "Found WSL2 disk: $($vhdx_file.FullName) ($old_size GB)" "INFO"

            # Verify file is not locked before diskpart runs
            $fileLocked = $true
            try {
                $stream = [System.IO.File]::Open($vhdx_file.FullName, 'Open', 'ReadWrite', 'None')
                $stream.Close()
                $fileLocked = $false
            } catch {
                Write-Status "VHD is still locked: $($_.Exception.Message)" "ERROR"
            }

            if (-not $fileLocked) {
                $diskpartScript = @"
select vdisk file="$($vhdx_file.FullName)"
attach vdisk readonly
compact vdisk
detach vdisk
exit
"@
                $scriptPath = Join-Path $env:TEMP 'kg_diskpart_compress.txt'
                $diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII
                Write-Status "Compacting VHD via diskpart..." "INFO"
                try {
                    & diskpart /s $scriptPath 2>&1 | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
                    $new_size = [math]::Round((Get-Item $vhdx_file.FullName).Length / 1GB, 2)
                    $saved = [math]::Round($old_size - $new_size, 2)
                    if ($saved -gt 0) {
                        Write-Status "VHD compressed: $old_size GB -> $new_size GB (saved: $saved GB)" "SUCCESS"
                    } else {
                        Write-Status "VHD already optimized: $old_size GB -> $new_size GB" "INFO"
                    }
                } catch {
                    Write-Status "diskpart failed: $($_.Exception.Message)" "ERROR"
                }
                Remove-Item $scriptPath -ErrorAction SilentlyContinue
            }
        } else {
            Write-Status "WSL2 VHD file not found" "WARNING"
        }

        Write-Status "Restarting WSL..." "INFO"
        wsl -e ls /home 2>$null | Out-Null
        Write-Status "WSL restarted" "SUCCESS"
    } else {
        Write-Status "WSL not found" "WARNING"
    }
}

# Show status
Write-Host "`n📊 Docker system status:" -ForegroundColor Cyan
docker system df 2>$null | Out-String | ForEach-Object { Write-Host $_ -ForegroundColor Gray }

Write-Host "`n✅ Cleanup completed!" -ForegroundColor Green
Write-Host "$(Get-Date -Format 'HH:mm:ss') Done." -ForegroundColor Gray

Write-Host "`nℹ️  Usage:" -ForegroundColor Cyan
Write-Host "  .\cleanup-docker.ps1              # Basic cleanup" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -Full        # Full system cleanup (aggressive)" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -WslOptimize # Include WSL2 disk optimization" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -Full -WslOptimize  # Full cleanup + WSL optimization" -ForegroundColor Gray
