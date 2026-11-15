# Checks that a fresh, non-empty Personal-stack backup exists before a
# destructive Docker operation. Mirrors the rule in
# scripts/devops/guard-personal-data.py - the backup globs and the
# freshness threshold come from the same policy, so the two cannot drift.
#
# Exit code 0 - a usable backup exists (prints its path and age).
# Exit code 1 - no usable backup (prints the reason and where to put one).

$ErrorActionPreference = "Stop"

Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem

$policyFile = Join-Path $PSScriptRoot "backup-policy.env"

# A backup counts only if it provably holds data (BACKUP-4): a 913-byte pg
# envelope with zero tables used to pass the freshness check. Mirrors
# scripts/devops/guard-personal-data.py.
function Test-SqlDumpHasNotes([string]$Path) {
    $fs = $null
    $reader = $null
    try {
        $fs = [System.IO.File]::OpenRead($Path)
        $magic = New-Object byte[] 2
        [void]$fs.Read($magic, 0, 2)
        $fs.Position = 0
        $stream = $fs
        if ($magic[0] -eq 0x1f -and $magic[1] -eq 0x8b) {
            $stream = New-Object System.IO.Compression.GzipStream($fs, [System.IO.Compression.CompressionMode]::Decompress)
        }
        $reader = New-Object System.IO.StreamReader($stream)
        while (($line = $reader.ReadLine()) -ne $null) {
            if (($line.StartsWith("COPY ") -or $line.StartsWith("INSERT INTO ")) -and ($line -match '\bnotes\b')) {
                return $true
            }
        }
    } catch {
        return $false
    } finally {
        if ($reader) { $reader.Close() }
        if ($fs) { $fs.Close() }
    }
    return $false
}

function Test-CustomDumpHasTables([string]$Path) {
    # No pg_restore means the .dump cannot be verified - refuse, not guess.
    if (-not (Get-Command pg_restore -ErrorAction SilentlyContinue)) { return $false }
    $listing = & pg_restore --list $Path 2>$null
    if ($LASTEXITCODE -ne 0) { return $false }
    return [bool]($listing | Select-String -Pattern 'TABLE\s+\S+\s+notes\b' -Quiet)
}

function Test-MongoDumpHasNamespace([string]$Path) {
    try {
        $bytes = [System.IO.File]::ReadAllBytes($Path)
        if ([System.Text.Encoding]::ASCII.GetString($bytes).Contains("knowledge_personal.")) { return $true }
        if ($bytes.Length -ge 2 -and $bytes[0] -eq 0x1f -and $bytes[1] -eq 0x8b) {
            $fs = [System.IO.File]::OpenRead($Path)
            try {
                $stream = New-Object System.IO.Compression.GzipStream($fs, [System.IO.Compression.CompressionMode]::Decompress)
                $buffer = New-Object byte[] (8MB)
                $read = $stream.Read($buffer, 0, $buffer.Length)
                $inner = [System.Text.Encoding]::ASCII.GetString($buffer, 0, $read)
                return $inner.Contains("knowledge_personal.")
            } finally {
                $fs.Close()
            }
        }
    } catch {
        return $false
    }
    return $false
}

function Test-BackupHasData([System.IO.FileInfo]$File) {
    if (-not $File -or $File.Length -eq 0) { return $false }
    $name = $File.Name
    if ($name -match '\.sql(\.gz)?$') { return Test-SqlDumpHasNotes $File.FullName }
    if ($name.EndsWith(".dump")) { return Test-CustomDumpHasTables $File.FullName }
    if ($name.EndsWith(".mongo.gz")) { return Test-MongoDumpHasNamespace $File.FullName }
    if ($name.StartsWith("personal-volumes-raw-")) { return $true }
    return $false
}

# Resolve the backup directory: $env:KG_BACKUP_DIR overrides backup-policy.env.
# A relative tail is expanded from the user's home directory.
$KG_BACKUP_DIR = $env:KG_BACKUP_DIR
if (-not $KG_BACKUP_DIR -and (Test-Path $policyFile)) {
    $line = Get-Content $policyFile | Where-Object { $_ -match '^\s*KG_BACKUP_DIR\s*=' } | Select-Object -First 1
    if ($line) { $KG_BACKUP_DIR = ($line -split '=', 2)[1].Trim() }
}
if (-not $KG_BACKUP_DIR) {
    Write-Host "  [ERROR] KG_BACKUP_DIR not set and not found in $policyFile" -ForegroundColor Red
    exit 1
}
if ($KG_BACKUP_DIR -match '^~') {
    $KG_BACKUP_DIR = $KG_BACKUP_DIR -replace '^~', $env:USERPROFILE
}
if ([System.IO.Path]::IsPathRooted($KG_BACKUP_DIR)) {
    $backupDir = $KG_BACKUP_DIR
} else {
    $backupDir = Join-Path $env:USERPROFILE $KG_BACKUP_DIR
}

# Threshold: env var first, then the shared policy file. Refuse rather than
# guess - a cleanup that cannot tell how fresh the backup must be is not safe.
$maxAgeHours = $env:KG_BACKUP_MAX_AGE_HOURS
if (-not $maxAgeHours) {
    if (-not (Test-Path $policyFile)) {
        Write-Host "  [ERROR] Backup policy file missing: $policyFile" -ForegroundColor Red
        exit 1
    }
    $line = Get-Content $policyFile | Where-Object { $_ -match '^\s*KG_BACKUP_MAX_AGE_HOURS\s*=' } | Select-Object -First 1
    $maxAgeHours = ($line -split '=', 2)[1].Trim()
}
if (-not $maxAgeHours) {
    Write-Host "  [ERROR] KG_BACKUP_MAX_AGE_HOURS not set in $policyFile" -ForegroundColor Red
    exit 1
}
$maxAgeHours = [double]$maxAgeHours

if (-not (Test-Path $backupDir)) {
    Write-Host "  [ERROR] No backup directory: $backupDir" -ForegroundColor Red
    Write-Host "          Run scripts/devops/backup-personal.ps1 first." -ForegroundColor Yellow
    exit 1
}

# A usable backup exists per datastore: Postgres AND MongoDB (BACKUP-4
# item 2). A single fresh file must not cover the other store.
function Get-NewestUsableBackup([string[]]$Patterns, [string]$ExcludeSuffix) {
    $candidates = @()
    foreach ($pattern in $Patterns) {
        $candidates += Get-ChildItem -Path $backupDir -Filter $pattern -File -ErrorAction SilentlyContinue
    }
    return @($candidates |
        Where-Object { (-not $ExcludeSuffix -or -not $_.Name.EndsWith($ExcludeSuffix)) -and (Test-BackupHasData $_) } |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1)
}

$latestPg = @(Get-NewestUsableBackup @("backup-personal-*", "personal-volumes-raw-*pgdata*") ".mongo.gz") | Select-Object -First 1
$latestMongo = @(Get-NewestUsableBackup @("backup-personal-*.mongo.gz", "personal-volumes-raw-*mongodbdata*") "") | Select-Object -First 1

$missing = @()
if (-not $latestPg) { $missing += "Postgres" }
if (-not $latestMongo) { $missing += "MongoDB" }
if ($missing) {
    Write-Host "  [ERROR] No usable backup in $backupDir : $($missing -join ', ') - a dump must hold real content (notes table / knowledge_personal namespace) or be a raw volume archive" -ForegroundColor Red
    Write-Host "          Run scripts/devops/backup-personal.ps1 first." -ForegroundColor Yellow
    exit 1
}

foreach ($latest in @($latestPg, $latestMongo)) {
    $ageHours = ((Get-Date) - $latest.LastWriteTime).TotalHours
    if ($ageHours -gt $maxAgeHours) {
        Write-Host ("  [ERROR] Newest backup {0} is {1:N1} h old; allowed: {2:N0} h" -f $latest.Name, $ageHours, $maxAgeHours) -ForegroundColor Red
        Write-Host "          Make a fresh backup via scripts/devops/backup-personal.ps1." -ForegroundColor Yellow
        exit 1
    }
}

$pgAge = ((Get-Date) - $latestPg.LastWriteTime).TotalHours
$mongoAge = ((Get-Date) - $latestMongo.LastWriteTime).TotalHours
Write-Host ("  [PASS] Backups fresh enough: {0} ({1:N0} KB, {2:N1} h), {3} ({4:N0} KB, {5:N1} h)" -f $latestPg.Name, ($latestPg.Length / 1KB), $pgAge, $latestMongo.Name, ($latestMongo.Length / 1KB), $mongoAge) -ForegroundColor Green
exit 0
