# Backup script for personal Knowledge Graph instance (Windows)
# Usage: .\scripts\devops\backup-personal.ps1

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
if ($env:BACKUP_CLOUD_ENABLED -eq "true") {
    Write-Host "Uploading to Yandex.Disk..."
    try {
        $Token = $env:BACKUP_YANDEX_TOKEN
        $BackupFolder = if ($env:BACKUP_YANDEX_FOLDER) { $env:BACKUP_YANDEX_FOLDER } else { "/KnowledgeGraphBackups" }
        if (-not $Token) {
            Write-Host "Warning: BACKUP_YANDEX_TOKEN not set"
        } else {
            # Create backup folder if it doesn't exist
            $FolderUrl = "https://webdav.yandex.ru$BackupFolder"
            $Headers = @{
                Authorization = "OAuth $Token"
            }

            try {
                Invoke-RestMethod -Uri $FolderUrl -Method MKCOL -Headers $Headers -ErrorAction SilentlyContinue
            } catch {
                # Folder might already exist, ignore error
            }

            # Upload file
            $RemotePath = "${BackupFolder}/$(Split-Path $BackupFile -Leaf)"
            $Url = "https://webdav.yandex.ru$RemotePath"

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

# Optional: Clean up old backups (keep last 7 days)
if ($env:CLEANUP_OLD_BACKUPS -ne "false") {
    Write-Host "Cleaning up old backups (keeping last 7 days)..."
    Get-ChildItem -Path $BackupDir -Filter "backup-personal-*.sql.gz" | 
        Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-7) } | 
        Remove-Item -Force
}

Write-Host "Backup process completed."
