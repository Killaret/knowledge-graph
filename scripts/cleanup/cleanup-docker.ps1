# Docker Cleanup Script for Knowledge Graph (Windows with WSL2)
# Removes dangling images, stopped containers, unused networks, and build cache
# Optionally compresses the Docker WSL2 disk
# Usage: .\cleanup-docker.ps1 [-Full] [-RemoveVolumes] [-WslOptimize] [-DryRun]
#
# Statuses and the exit code are honest: every step registers its real
# $LASTEXITCODE, the script exits non-zero when any step failed, and a failed
# step prints the command's stderr. -DryRun previews each step and changes
# nothing. Steps that can touch volumes (6, 7) refuse without a fresh
# non-empty Personal-stack backup — the rule lives in
# scripts/devops/check-personal-backup.ps1 and backup-policy.env, the same
# policy the guard-personal-data.py hook enforces.

param(
    [switch]$Full = $false,
    [switch]$RemoveVolumes = $false,
    [switch]$WslOptimize = $false,
    [switch]$DryRun = $false
)

$ErrorActionPreference = "Continue"

. (Join-Path $PSScriptRoot "..\testing\lib\phase-tracking.ps1")
$script:SnapshotDir = $null
$BackupCheck = Join-Path $PSScriptRoot "..\devops\check-personal-backup.ps1"

Write-Host "Knowledge Graph Docker Cleanup" -ForegroundColor Cyan
Write-Host "$(Get-Date -Format 'HH:mm:ss') Starting cleanup$(if ($DryRun) { ' (dry-run)' })..." -ForegroundColor Gray

function Invoke-DockerStep {
    # Runs one docker command, registers its real exit code, and prints the
    # captured output on failure so there is something to fix.
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][scriptblock]$Command
    )
    $output = & $Command 2>&1
    $code = $LASTEXITCODE
    if ($code -ne 0 -and $output) {
        $output | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
    }
    Register-Phase $Name -ExitCode $code
    return $code
}

# 1. Stop all containers
Write-Host "`n1. Stopping containers..." -ForegroundColor Cyan
$running = @(docker ps -q 2>&1)
if ($LASTEXITCODE -ne 0) {
    $running | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
    Register-Phase "stop-containers" -ExitCode $LASTEXITCODE
} elseif (-not $running) {
    Register-Phase "stop-containers" -Skipped -Reason "no running containers"
} elseif ($DryRun) {
    Write-Host "  Would stop $($running.Count) container(s):" -ForegroundColor Yellow
    $running | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
    Register-Phase "stop-containers" -Skipped -Reason "dry-run"
} else {
    $code = Invoke-DockerStep "stop-containers" { docker stop $running }
}

# 2. Remove dangling images
Write-Host "`n2. Removing dangling images..." -ForegroundColor Cyan
if ($DryRun) {
    $dangling = @(docker images -f "dangling=true" -q 2>&1)
    if ($LASTEXITCODE -ne 0) {
        $dangling | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
        Register-Phase "prune-dangling-images" -ExitCode $LASTEXITCODE
    } else {
        Write-Host "  Would remove $($dangling.Count) dangling image(s)" -ForegroundColor Yellow
        $dangling | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        Register-Phase "prune-dangling-images" -Skipped -Reason "dry-run"
    }
} else {
    $code = Invoke-DockerStep "prune-dangling-images" { docker image prune -f }
}

# 3. Remove stopped containers
Write-Host "`n3. Removing stopped containers..." -ForegroundColor Cyan
if ($DryRun) {
    $stopped = @(docker ps -aq --filter "status=exited" --filter "status=created" --filter "status=dead" 2>&1)
    if ($LASTEXITCODE -ne 0) {
        $stopped | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
        Register-Phase "prune-stopped-containers" -ExitCode $LASTEXITCODE
    } else {
        Write-Host "  Would remove $($stopped.Count) stopped container(s)" -ForegroundColor Yellow
        $stopped | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        Register-Phase "prune-stopped-containers" -Skipped -Reason "dry-run"
    }
} else {
    $code = Invoke-DockerStep "prune-stopped-containers" { docker container prune -f }
}

# 4. Remove unused networks
Write-Host "`n4. Removing unused networks..." -ForegroundColor Cyan
if ($DryRun) {
    $networks = docker network ls --format "{{.Name}}" 2>&1
    if ($LASTEXITCODE -ne 0) {
        $networks | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
        Register-Phase "prune-networks" -ExitCode $LASTEXITCODE
    } else {
        Write-Host "  Would remove unused networks. Current networks:" -ForegroundColor Yellow
        $networks | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        Register-Phase "prune-networks" -Skipped -Reason "dry-run"
    }
} else {
    $code = Invoke-DockerStep "prune-networks" { docker network prune -f }
}

# 5. Remove build cache
Write-Host "`n5. Clearing Docker build cache..." -ForegroundColor Cyan
if ($DryRun) {
    $df = docker system df 2>&1
    if ($LASTEXITCODE -ne 0) {
        $df | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
        Register-Phase "prune-build-cache" -ExitCode $LASTEXITCODE
    } else {
        $df | Where-Object { $_ -match 'Build Cache' } | ForEach-Object { Write-Host "  Would clear: $_" -ForegroundColor Yellow }
        Register-Phase "prune-build-cache" -Skipped -Reason "dry-run"
    }
} else {
    $code = Invoke-DockerStep "prune-build-cache" { docker builder prune -f }
}

# 6. Volumes are preserved by default. With -RemoveVolumes only anonymous
# dangling volumes are eligible; personal-named and protected-labeled
# volumes are always skipped. A fresh non-empty backup is required first.
Write-Host "`n6. Volume cleanup..." -ForegroundColor Cyan
if (-not $RemoveVolumes) {
    Register-Phase "volume-cleanup" -Skipped -Reason "default safe mode"
} else {
    if ($DryRun) {
        Write-Host "  (dry-run: nothing will be removed)" -ForegroundColor Yellow
        & $BackupCheck
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  A real -RemoveVolumes run would stop here" -ForegroundColor Yellow
        }
    } else {
        & $BackupCheck
        if ($LASTEXITCODE -ne 0) {
            Register-Phase "volume-cleanup" -ExitCode 1
        }
    }
    if ($DryRun -or $LASTEXITCODE -eq 0) {
        $protectedVolumes = @(docker volume ls --filter "label=com.knowledgegraph.protected=true" --format "{{.Name}}" 2>&1)
        $lsCode = $LASTEXITCODE
        $danglingVolumes = @(docker volume ls --filter "dangling=true" --format "{{.Name}}" 2>&1)
        if ($LASTEXITCODE -ne 0) { $lsCode = $LASTEXITCODE }
        if ($lsCode -ne 0) {
            $protectedVolumes + $danglingVolumes | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
            Register-Phase "volume-cleanup" -ExitCode $lsCode
        } else {
        $eligible = @()
        $kept = @()
        foreach ($volume in $danglingVolumes) {
            $isAnonymous = $volume -match '^[0-9a-f]{64}$'
            $isPersonal = $volume -match 'personal'
            $isProtected = $protectedVolumes -contains $volume
            if (-not $isAnonymous -or $isPersonal -or $isProtected) {
                $kept += $volume
            } else {
                $eligible += $volume
            }
        }
        if ($DryRun) {
            Write-Host "  Would remove $($eligible.Count) anonymous dangling volume(s):" -ForegroundColor Yellow
            $eligible | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
            Write-Host "  Kept (named / personal / protected): $($kept.Count)" -ForegroundColor Gray
            Register-Phase "volume-cleanup" -Skipped -Reason "dry-run"
        } else {
            $removedVolumes = 0
            $failedVolumes = 0
            foreach ($volume in $eligible) {
                $rmOutput = docker volume rm $volume 2>&1
                if ($LASTEXITCODE -eq 0) {
                    $removedVolumes++
                } else {
                    $failedVolumes++
                    $rmOutput | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
                }
            }
            Write-Host "  Removed $removedVolumes anonymous dangling volume(s)$(if ($failedVolumes) { ", $failedVolumes failed" })" -ForegroundColor Gray
            Register-Phase "volume-cleanup" -ExitCode $(if ($failedVolumes -gt 0) { 1 } else { 0 })
        }
        }
    }
}

# 7. Full cleanup mode. NOTE: step 1 stops every container and step 3
# removes all stopped ones, so by this point NO container remains and
# `docker system prune -af` treats every image on the machine as unused —
# the price is the whole local image store, not just project layers.
if ($Full) {
    Write-Host "`n7. Full cleanup mode (removing ALL unused images, not volumes)..." -ForegroundColor Cyan
    $images = @(docker images -q 2>&1)
    $imgCode = $LASTEXITCODE
    if ($imgCode -ne 0) {
        $images | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
        Register-Phase "full-cleanup" -ExitCode $imgCode
    } else {
    $dfBefore = docker system df 2>$null
    Write-Host "  Cost: $($images.Count) image(s) will be removed (no containers remain after steps 1+3)" -ForegroundColor Yellow
    $dfBefore | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
    if ($DryRun) {
        Register-Phase "full-cleanup" -Skipped -Reason "dry-run"
    } else {
        & $BackupCheck
        if ($LASTEXITCODE -ne 0) {
            Register-Phase "full-cleanup" -ExitCode 1
        } else {
            $code = Invoke-DockerStep "full-cleanup" { docker system prune -af }
        }
    }
    }
} else {
    Register-Phase "full-cleanup" -Skipped -Reason "not requested"
}

# 8. WSL2 optimization (optional). Uses diskpart to compact the VHD —
# requires admin rights, but NOT Hyper-V.
if ($WslOptimize) {
    Write-Host "`n8. Optimizing WSL2 disk..." -ForegroundColor Cyan

    $wsl_check = wsl --list 2>$null
    if (-not $wsl_check) {
        Register-Phase "optimize-disk" -Skipped -Reason "WSL not found"
    } else {
        # Only the Docker Desktop VHD is a valid target; the disk lives
        # under %LOCALAPPDATA%\Docker\wsl. Anything else is not ours.
        $vhdx_file = Get-ChildItem -Path (Join-Path $env:LOCALAPPDATA 'Docker\wsl') -Filter *.vhdx -Recurse -ErrorAction SilentlyContinue |
            Sort-Object Length -Descending | Select-Object -First 1

        if (-not $vhdx_file) {
            Register-Phase "optimize-disk" -Skipped -Reason "Docker WSL2 VHD not found under $env:LOCALAPPDATA\Docker\wsl"
        } else {
            $old_size = [math]::Round($vhdx_file.Length / 1GB, 2)
            Write-Host "  Target VHD: $($vhdx_file.FullName) ($old_size GB)" -ForegroundColor Gray

            if ($DryRun) {
                Write-Host "  Would shut down WSL and compact this disk" -ForegroundColor Yellow
                Register-Phase "optimize-disk" -Skipped -Reason "dry-run"
            } else {
                Write-Host "  Shutting down WSL..." -ForegroundColor Gray
                wsl --shutdown 2>$null

                # Wait for the VHD to be released. Never kill processes by
                # name: the Personal stack's volumes live inside this VM,
                # and force-killing it leaves the filesystem dirty.
                $fileLocked = $true
                for ($i = 0; $i -lt 30 -and $fileLocked; $i++) {
                    Start-Sleep -Seconds 2
                    try {
                        $stream = [System.IO.File]::Open($vhdx_file.FullName, 'Open', 'ReadWrite', 'None')
                        $stream.Close()
                        $fileLocked = $false
                    } catch {
                        $fileLocked = $true
                    }
                }

                if ($fileLocked) {
                    Register-Phase "optimize-disk" -ExitCode 1
                    Write-Host "    VHD still locked after 60s — refusing to compact. Close Docker Desktop and retry." -ForegroundColor Red
                } else {
                    $diskpartScript = @"
select vdisk file="$($vhdx_file.FullName)"
attach vdisk readonly
compact vdisk
detach vdisk
exit
"@
                    $scriptPath = Join-Path $env:TEMP 'kg_diskpart_compress.txt'
                    $diskpartScript | Out-File -FilePath $scriptPath -Encoding ASCII
                    $output = & diskpart /s $scriptPath 2>&1
                    $code = $LASTEXITCODE
                    $output | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
                    Remove-Item $scriptPath -ErrorAction SilentlyContinue
                    if ($code -eq 0) {
                        $new_size = [math]::Round((Get-Item $vhdx_file.FullName).Length / 1GB, 2)
                        Write-Host "  VHD: $old_size GB -> $new_size GB" -ForegroundColor Gray
                    }
                    Register-Phase "optimize-disk" -ExitCode $code

                    Write-Host "  Restarting WSL..." -ForegroundColor Gray
                    wsl -e ls /home 2>$null | Out-Null
                }
            }
        }
    }
} else {
    Register-Phase "optimize-disk" -Skipped -Reason "not requested"
}

# Show status
Write-Host "`nDocker system status:" -ForegroundColor Cyan
docker system df 2>$null | Out-String | ForEach-Object { Write-Host $_ -ForegroundColor Gray }

$scriptFailed = Test-AnyFailed
Write-FinalSummary -Success (-not $scriptFailed)

Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  .\cleanup-docker.ps1                    # Basic cleanup; preserves all volumes" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -DryRun            # Preview every step, change nothing" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -Full              # Remove all unused images; still preserves volumes" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -RemoveVolumes     # Remove only anonymous dangling volumes" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -WslOptimize       # Include WSL2 disk optimization" -ForegroundColor Gray
Write-Host "  .\cleanup-docker.ps1 -Full -WslOptimize # Full image cleanup + WSL optimization" -ForegroundColor Gray

if ($scriptFailed) { exit 1 }
exit 0
