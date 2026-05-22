# Docker Cleanup Script for Knowledge Graph (Windows with WSL2)
# Removes dangling images, stopped containers, unused networks, and build cache
# Optionally compresses WSL2 disk for efficient storage
# Usage: .\cleanup-docker.ps1 [-Full] [-WslOptimize]

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
if ($WslOptimize) {
    Write-Host "`n8️⃣ Optimizing WSL2 disk..." -ForegroundColor Cyan
    
    $wsl_check = wsl --list 2>$null
    if ($wsl_check) {
        Write-Status "Shutting down WSL..." "INFO"
        wsl --shutdown 2>$null
        Start-Sleep -Seconds 3
        Write-Status "WSL shut down successfully" "SUCCESS"
        
        $searchPaths = @(
            Join-Path $env:LOCALAPPDATA 'Packages\CanonicalGroupLimited.Ubuntu*\LocalState\ext4.vhdx'
            Join-Path $env:LOCALAPPDATA 'Docker\wsl\data\ext4.vhdx'
            Join-Path $env:LOCALAPPDATA 'Docker\wsl\main\ext4.vhdx'
            Join-Path $env:LOCALAPPDATA 'Docker\wsl\disk\docker_data.vhdx'
        )

        $vhdx_file = $null
        foreach ($path in $searchPaths) {
            $candidate = Get-Item $path -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($candidate) {
                $vhdx_file = $candidate
                break
            }
        }

        if (-not $vhdx_file) {
            $vhdx_file = Get-ChildItem -Path (Join-Path $env:LOCALAPPDATA 'Docker\wsl') -Filter *.vhdx -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        }

        if ($vhdx_file) {
            $old_size = [math]::Round($vhdx_file.Length / 1GB, 2)
            Write-Status "Found WSL2 disk: $($vhdx_file.FullName) ($old_size GB)" "INFO"
            
            if ((Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -ErrorAction SilentlyContinue).State -eq "Enabled") {
                try {
                    Write-Status "Compressing VHD file..." "INFO"
                    Optimize-VHD -Path $vhdx_file.FullName -Mode Full -ErrorAction Stop | Out-Null
                    $new_size = [math]::Round((Get-Item $vhdx_file.FullName).Length / 1GB, 2)
                    $saved = [math]::Round($old_size - $new_size, 2)
                    Write-Status "VHD compressed: $old_size GB → $new_size GB (saved: $saved GB)" "SUCCESS"
                } catch {
                    Write-Status "VHD compression failed (Hyper-V required or path locked)" "WARNING"
                }
            } else {
                Write-Status "Hyper-V not enabled. Cannot compress VHD." "WARNING"
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
