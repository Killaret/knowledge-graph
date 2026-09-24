# check-disk-layout.ps1 - DISK-1 guard: heavy tool caches must not live on the system drive.
# Prints each path's size; FAILS if a configured path points at the system drive.
# Not part of check-all: the GitHub runner has its own disk layout.

$ErrorActionPreference = 'Continue'
$fail = @()
# The OS reports its own system drive (usually C:) - do not hardcode the letter.
$sysDrive = $env:SystemDrive
if ([string]::IsNullOrWhiteSpace($sysDrive)) { $sysDrive = 'C:' }

function Test-PathOnSystemDrive([string]$label, [string]$path, [string]$how) {
    if ([string]::IsNullOrWhiteSpace($path)) {
        Write-Host ("  {0,-28} (not set) - {1}" -f $label, $how)
        return
    }
    $p = $path.TrimEnd('\')
    if ($p -match '^[a-zA-Z]:') {
        if ($p.StartsWith($sysDrive, [StringComparison]::OrdinalIgnoreCase)) {
            $script:fail += "$label -> $p ($how)"
            Write-Host ("  {0,-28} {1}  <-- ON $sysDrive" -f $label, $p)
        } else {
            Write-Host ("  {0,-28} {1}" -f $label, $p)
        }
    } else {
        Write-Host ("  {0,-28} {1} (unresolved)" -f $label, $p)
    }
}

function Show-Size([string]$label, [string]$path) {
    if (Test-Path $path) {
        $item = Get-Item $path -Force
        # A junction lives on this drive but its data is elsewhere - do not walk it.
        if ($item.LinkType -in @('Junction', 'SymbolicLink')) {
            Write-Host ("  {0,-28} junction  {1} -> {2}" -f $label, $path, ($item.Target -join ', '))
            return
        }
        $gb = (Get-ChildItem $path -Recurse -File -Force -ErrorAction SilentlyContinue |
               Measure-Object Length -Sum).Sum / 1GB
        Write-Host ("  {0,-28} {1,7:N2} GB  {2}" -f $label, $gb, $path)
    } else {
        Write-Host ("  {0,-28}    absent  {1}" -f $label, $path)
    }
}

Write-Host "== Configured tool paths (must be off $sysDrive) =="
# go env prints bare values (no key=value) - read positionally
$goEnv = @{}
try {
    $vals = (go env GOCACHE GOMODCACHE GOTMPDIR 2>$null) -split "`n" |
            ForEach-Object { $_.Trim() } | Where-Object { $_ }
    if ($vals.Count -ge 1) { $goEnv['GOCACHE'] = $vals[0] }
    if ($vals.Count -ge 2) { $goEnv['GOMODCACHE'] = $vals[1] }
    if ($vals.Count -ge 3) { $goEnv['GOTMPDIR'] = $vals[2] }
} catch {}
Test-PathOnSystemDrive 'go GOCACHE'      $goEnv['GOCACHE']   'go env -w GOCACHE=D:\kg-build-cache\go-build'
Test-PathOnSystemDrive 'go GOMODCACHE'   $goEnv['GOMODCACHE'] 'go env -w GOMODCACHE=D:\kg-build-cache\go-mod'
Test-PathOnSystemDrive 'go GOTMPDIR'     $goEnv['GOTMPDIR']   'go env -w GOTMPDIR=D:\kg-build-cache\tmp'

$npmCache = (npm config get cache 2>$null | Select-Object -First 1)
Test-PathOnSystemDrive 'npm cache' $npmCache 'npm config set cache D:\kg-build-cache\npm --location=user'

$pipCache = (pip cache dir 2>$null | Select-Object -First 1)
Test-PathOnSystemDrive 'pip cache' $pipCache 'pip config set global.cache-dir D:\kg-build-cache\pip'

# User env vars (setx): read from the registry, not this session
$reg = 'HKCU:\Environment'
function Get-UserEnv([string]$name) {
    (Get-ItemProperty $reg -Name $name -ErrorAction SilentlyContinue).$name
}
Test-PathOnSystemDrive 'PLAYWRIGHT_BROWSERS_PATH' (Get-UserEnv 'PLAYWRIGHT_BROWSERS_PATH') 'setx PLAYWRIGHT_BROWSERS_PATH D:\kg-build-cache\ms-playwright'
Test-PathOnSystemDrive 'PUPPETEER_CACHE_DIR'      (Get-UserEnv 'PUPPETEER_CACHE_DIR')      'setx PUPPETEER_CACHE_DIR D:\kg-build-cache\puppeteer'
Test-PathOnSystemDrive 'TORCH_HOME'               (Get-UserEnv 'TORCH_HOME')               'setx TORCH_HOME D:\kg-build-cache\torch'
Test-PathOnSystemDrive 'HF_HOME'                  (Get-UserEnv 'HF_HOME')                  'setx HF_HOME D:\kg-hf-cache'

# HF_CACHE_DIR from .env next to the repo (compose mount source)
$envFile = Join-Path $PSScriptRoot '..\..\.env'
if (Test-Path $envFile) {
    $hfLine = Get-Content $envFile | Where-Object { $_ -match '^\s*HF_CACHE_DIR\s*=' } | Select-Object -First 1
    $hfDir = ($hfLine -split '=', 2)[1]
    Test-PathOnSystemDrive '.env HF_CACHE_DIR' $hfDir '.env: HF_CACHE_DIR=D:/kg-hf-cache'
} else {
    Write-Host "  .env HF_CACHE_DIR            (.env absent - cp .env.example .env)"
}

Write-Host ""
Write-Host "== Known heavy dirs still on the system drive (sizes) =="
$home2 = $env:USERPROFILE
Show-Size 'go modules'        "$home2\go\pkg\mod"
Show-Size 'go build cache'    "$env:LOCALAPPDATA\go-build"
Show-Size 'npm cache'         "$env:LOCALAPPDATA\npm-cache"
Show-Size 'pip cache'         "$env:LOCALAPPDATA\pip\cache"
Show-Size 'playwright'        "$env:LOCALAPPDATA\ms-playwright"
Show-Size '.cache leftovers'  "$home2\.cache"
Show-Size 'devin cli data'    "$env:APPDATA\devin\cli"
Show-Size 'claude vm_bundles' "$env:APPDATA\Claude\vm_bundles"
Show-Size 'wsl vhdx'          "$env:LOCALAPPDATA\wsl"

Write-Host ""
if ($fail.Count -gt 0) {
    Write-Host "FAIL: $($fail.Count) configured path(s) still point at ${sysDrive}:"
    $fail | ForEach-Object { Write-Host "  - $_" }
    exit 1
}
Write-Host "OK: all configured tool caches are off $sysDrive"
exit 0
