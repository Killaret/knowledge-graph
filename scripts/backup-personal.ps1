# Backup script for personal Knowledge Graph instance (Windows)
# Usage: .\scripts\backup-personal.ps1

$ErrorActionPreference = "Stop"

# Configuration
$BackupDir = if ($env:BACKUP_DIR) { $env:BACKUP_DIR } else { ".\backups" }
$Timestamp = Get-Date -Format "yyyy-MM-dd"
$BackupFile = Join-Path $BackupDir "backup-personal-${Timestamp}.sql"

# Database connection (can be overridden by environment variables)
$DbHost = if ($env:PERSONAL_POSTGRES_HOST) { $env:PERSONAL_POSTGRES_HOST } else { "localhost" }
$DbPort = if ($env:PERSONAL_POSTGRES_PORT) { $env:PERSONAL_POSTGRES_PORT } else { "5433" }
$DbUser = if ($env:PERSONAL_POSTGRES_USER) { $env:PERSONAL_POSTGRES_USER } else { "personal" }
$DbPassword = if ($env:PERSONAL_POSTGRES_PASSWORD) { $env:PERSONAL_POSTGRES_PASSWORD } else { "personal_password" }
$DbName = if ($env:PERSONAL_POSTGRES_DB) { $env:PERSONAL_POSTGRES_DB } else { "knowledge_personal" }

# Create backup directory if it doesn't exist
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

Write-Host "Starting backup of personal Knowledge Graph database..."
Write-Host "Backup file: $BackupFile"

# Set PGPASSWORD environment variable for pg_dump
$env:PGPASSWORD = $DbPassword

# Perform pg_dump
& pg_dump -h $DbHost -p $DbPort -U $DbUser -d $DbName `
    --format=plain `
    --no-owner `
    --no-acl `
    --verbose `
    --file=$BackupFile

# Unset PGPASSWORD
$env:PGPASSWORD = $null

# Compress the backup
$CompressedFile = "${BackupFile}.gz"
& gzip $BackupFile
$BackupFile = $CompressedFile

Write-Host "Backup completed successfully: $BackupFile"

# Optional: Upload to Yandex.Disk via WebDAV
if ($env:BACKUP_CLOUD_ENABLED -eq "true" -and $env:BACKUP_CLOUD_PROVIDER -eq "yandex") {
    Write-Host "Uploading to Yandex.Disk..."
    try {
        $Token = $env:BACKUP_YANDEX_TOKEN
        if (-not $Token) {
            Write-Host "Warning: BACKUP_YANDEX_TOKEN not set"
        } else {
            $RemotePath = "/knowledge-graph-backups/$(Split-Path $BackupFile -Leaf)"
            $Url = "https://webdav.yandex.ru$RemotePath"

            $Headers = @{
                Authorization = "OAuth $Token"
            }

            Invoke-RestMethod -Uri $Url `
                -Method PUT `
                -Headers $Headers `
                -InFile $BackupFile `
                -ContentType "application/octet-stream"

            Write-Host "Successfully uploaded to Yandex.Disk: $RemotePath"
        }
    } catch {
        Write-Host "Warning: Yandex.Disk upload failed: $_"
    }
}

# Optional: Trigger cloud backup if backend is running (for R2)
if ($env:CLOUD_BACKUP_ENABLED -eq "true" -and $env:BACKUP_CLOUD_PROVIDER -eq "r2") {
    Write-Host "Triggering cloud backup..."
    try {
        $Body = @{
            local_path = $BackupFile
        } | ConvertTo-Json

        Invoke-RestMethod -Uri "http://localhost:8081/api/v1/admin/backup/cloud" `
            -Method POST `
            -ContentType "application/json" `
            -Body $Body
    } catch {
        Write-Host "Warning: Cloud backup trigger failed: $_"
    }
}

# Optional: Clean up old backups (keep last 7 days)
if ($env:CLEANUP_OLD_BACKUPS -ne "false") {
    Write-Host "Cleaning up old backups (keeping last 7 days)..."
    Get-ChildItem -Path $BackupDir -Filter "backup-personal-*.sql.gz" | 
        Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-7) } | 
        Remove-Item -Force
}

Write-Host "Backup process completed."
