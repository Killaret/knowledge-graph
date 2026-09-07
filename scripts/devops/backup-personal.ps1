# Backup script for personal Knowledge Graph instance (Windows)
# Supports two modes:
#   daily  - default, 7 days local retention
#   weekly - Sunday snapshot, 90 days local retention
#
# Usage:
#   .\scripts\devops\backup-personal.ps1 [-Mode daily]
#   $env:BACKUP_MODE = "weekly"; .\scripts\devops\backup-personal.ps1

[CmdletBinding()]
param(
    [ValidateSet("daily", "weekly")]
    [string]$Mode = $(if ($env:BACKUP_MODE) { $env:BACKUP_MODE } else { "daily" })
)

$ErrorActionPreference = "Stop"

# Configuration
$BackupDir = if ($env:BACKUP_DIR) { $env:BACKUP_DIR } else { ".\backups" }
$Timestamp = Get-Date -Format "yyyy-MM-dd-HHmmss"
$BackupFile = Join-Path $BackupDir "backup-personal-${Mode}-${Timestamp}.sql"

# Database connection (can be overridden by environment variables)
$DbHost = if ($env:PERSONAL_POSTGRES_HOST) { $env:PERSONAL_POSTGRES_HOST } else { "localhost" }
$DbPort = if ($env:PERSONAL_POSTGRES_PORT) { $env:PERSONAL_POSTGRES_PORT } else { "5433" }
$DbUser = if ($env:PERSONAL_POSTGRES_USER) { $env:PERSONAL_POSTGRES_USER } else { "personal" }
$DbPassword = if ($env:PERSONAL_POSTGRES_PASSWORD) { $env:PERSONAL_POSTGRES_PASSWORD } else { "personal_password" }
$DbName = if ($env:PERSONAL_POSTGRES_DB) { $env:PERSONAL_POSTGRES_DB } else { "knowledge_personal" }

# Local retention: cloud backups are never deleted
$RetentionDays = if ($Mode -eq "weekly") {
    if ($env:BACKUP_WEEKLY_RETENTION_DAYS) { [int]$env:BACKUP_WEEKLY_RETENTION_DAYS } else { 90 }
} else {
    if ($env:BACKUP_DAILY_RETENTION_DAYS) { [int]$env:BACKUP_DAILY_RETENTION_DAYS } else { 7 }
}

# Create backup directory if it doesn't exist
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

Write-Host "Starting ${Mode} backup of personal Knowledge Graph database..."
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

# Optional: Upload to Yandex.Disk via REST API
if ($env:BACKUP_CLOUD_ENABLED -eq "true") {
    Write-Host "Uploading to Yandex.Disk..."
    try {
        $Token = if ($env:BACKUP_YANDEX_OAUTH_TOKEN) { $env:BACKUP_YANDEX_OAUTH_TOKEN } else { $env:BACKUP_YANDEX_TOKEN }
        $BackupFolder = if ($env:BACKUP_YANDEX_FOLDER) { $env:BACKUP_YANDEX_FOLDER } else { "/KnowledgeGraphBackups" }

        if (-not $Token) {
            Write-Host "Warning: BACKUP_YANDEX_OAUTH_TOKEN or BACKUP_YANDEX_TOKEN not set"
        } else {
            $BaseUrl = "https://cloud-api.yandex.net/v1/disk"
            $RemotePath = "${BackupFolder}/$(Split-Path $BackupFile -Leaf)"
            $Headers = @{ Authorization = "OAuth $Token" }

            # Ensure the backup folder exists (ignore 409 - already exists)
            try {
                Invoke-RestMethod -Uri "${BaseUrl}/resources?path=${BackupFolder}" `
                    -Method PUT `
                    -Headers $Headers `
                    -ErrorAction SilentlyContinue | Out-Null
            } catch {
                $status = $_.Exception.Response.StatusCode.value__
                if ($status -ne 409) {
                    Write-Host "Warning: failed to ensure backup folder on Yandex.Disk (HTTP $status)"
                }
            }

            # Get pre-signed upload URL
            $UploadLink = Invoke-RestMethod -Uri "${BaseUrl}/resources/upload?path=${RemotePath}&overwrite=true" `
                -Method GET `
                -Headers $Headers

            # Upload file to the pre-signed URL
            Invoke-RestMethod -Uri $UploadLink.href `
                -Method PUT `
                -InFile $BackupFile `
                -ContentType "application/octet-stream" | Out-Null

            Write-Host "Successfully uploaded to Yandex.Disk: $RemotePath"
        }
    } catch {
        Write-Host "Warning: Yandex.Disk upload failed: $_"
    }
}

# Clean up old local backups (cloud backups are never deleted)
if ($env:CLEANUP_OLD_BACKUPS -ne "false" -and $RetentionDays -gt 0) {
    Write-Host "Cleaning up old ${Mode} local backups (keeping last ${RetentionDays} days)..."
    Get-ChildItem -Path $BackupDir -Filter "backup-personal-${Mode}-*.sql.gz" |
        Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$RetentionDays) } |
        Remove-Item -Force
}

Write-Host "Backup process completed."
