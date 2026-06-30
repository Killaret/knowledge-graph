# Yandex.Disk Backup Setup

This project supports database backup to Yandex.Disk via REST API.

## Overview

The Yandex.Disk backup system includes:
- **Local scripts**: `scripts/backup-personal.ps1` (Windows) and `scripts/backup-personal.sh` (Linux/Mac)
- **REST API**: direct file upload to Yandex.Disk via OAuth token
- **Automatic cleanup**: deletion of old backups (default 7 days)

## Storage Requirements

For Knowledge Graph, you typically need:
- **Text data (notes)**: a few MB
- **Embeddings**: may take more space, usually up to 100-500 MB
- **Daily backups for a week**: ~1-3 GB

**Yandex.Disk free tier**: 10 GB — more than enough for personal use.

## Configuration

### Method 1: Via knowledge-graph.config.json (recommended)

Edit the `knowledge-graph.config.json` file:

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": true,
      "provider": "yandex",
      "yandex": {
        "oauth_token": "your_oauth_token_here"
      }
    },
    "schedule": "0 2 * * *",
    "retention_days": 7,
    "draft_ttl_hours": 168
  }
}
```

### Method 2: Via environment variables

Add to `.env` file:

```bash
# Enable cloud backup
BACKUP_CLOUD_ENABLED=true

# Provider - Yandex.Disk
BACKUP_CLOUD_PROVIDER=yandex

# Local path for backups
BACKUP_LOCAL_PATH=./backups

# Schedule in cron format (0 2 * * * = every day at 2:00)
BACKUP_SCHEDULE=0 2 * * *

# Keep backups for 7 days
BACKUP_RETENTION_DAYS=7

# Draft TTL in hours (168 = 7 days)
BACKUP_DRAFT_TTL_HOURS=168

# Yandex.Disk OAuth token
BACKUP_YANDEX_TOKEN=your_oauth_token_here
```

## Configuring Backup Frequency

Backup frequency is configured via the `schedule` parameter in cron format.

### Cron Format

```
minute hour day_of_month month day_of_week
*      *    *           *      *
```

- **minute**: 0-59
- **hour**: 0-23
- **day_of_month**: 1-31
- **month**: 1-12
- **day_of_week**: 0-7 (0 or 7 = Sunday)

### Schedule Examples

```json
"schedule": "0 2 * * *"           // Every day at 2:00
"schedule": "0 */6 * * *"         // Every 6 hours
"schedule": "0 0 * * 0"           // Every Sunday at midnight
"schedule": "0 0 1 * *"           // 1st of every month at midnight
"schedule": "*/30 * * * *"        // Every 30 minutes
"schedule": "0 12,18 * * *"       // Every day at 12:00 and 18:00
```

### Configuration via config file

```json
{
  "backup": {
    "schedule": "0 2 * * *",
    "retention_days": 7
  }
}
```

### Configuration via environment variables

```bash
BACKUP_SCHEDULE=0 2 * * *
BACKUP_RETENTION_DAYS=7
```

### Retention parameters

- **retention_days**: Number of days to keep backups (default 7)
- **draft_ttl_hours**: Draft TTL in hours (default 168 = 7 days)

### Backup integrity verification

The system automatically verifies integrity of uploaded backups using SHA256 hashes:

1. Hash is computed for the local file before upload
2. After upload, the file is downloaded to a temporary directory
3. Hash of the uploaded file is computed
4. Hashes are compared to verify integrity

If hashes do not match, the upload is considered failed and an error is returned.

## Getting Yandex.Disk OAuth Token

### 1. Create an application

1. Go to [Yandex OAuth](https://oauth.yandex.ru/client/new)
2. Fill in the form:
   - **Application name**: Knowledge Graph Backup
   - **Description**: Database backup
   - **Website URL**: `http://localhost`
   - **Redirect URI**: `https://oauth.yandex.ru/verification_code`
   - **Permissions (DLS)**: select:
     - `cloud_api:disk.write` - Write anywhere on Disk
     - `cloud_api:disk.read` - Read entire Disk
3. Get the **Client ID**

### 2. Get the token

**Method A: Via browser (easier)**

1. Open in browser:
   ```
   https://oauth.yandex.ru/authorize?response_type=token&client_id=YOUR_CLIENT_ID
   ```
   Replace `YOUR_CLIENT_ID` with your Client ID
2. Log in to Yandex
3. Grant access to Yandex.Disk
4. The URL after redirect will contain the token in format:
   ```
   https://oauth.yandex.ru/verification_code#access_token=YOUR_TOKEN&...
   ```
5. Copy the token (part after `access_token=`)

**Method B: Via curl**

```bash
curl -X POST "https://oauth.yandex.ru/token" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

## Usage

### Manual backup run (Windows)

```powershell
# Set environment variables
$env:BACKUP_CLOUD_ENABLED = "true"
$env:BACKUP_CLOUD_PROVIDER = "yandex"
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token_here"

# Run backup script
.\scripts\backup-personal.ps1
```

### Manual backup run (Linux/Mac)

```bash
# Set environment variables
export BACKUP_CLOUD_ENABLED=true
export BACKUP_CLOUD_PROVIDER=yandex
export BACKUP_YANDEX_TOKEN=your_oauth_token_here

# Run backup script
./scripts/backup-personal.sh
```

## Automation via cron

### Linux/Mac

Add to crontab (`crontab -e`):

```bash
# Daily backup at 2:00 AM
0 2 * * * cd /path/to/knowledge-graph && BACKUP_CLOUD_ENABLED=true BACKUP_CLOUD_PROVIDER=yandex BACKUP_YANDEX_TOKEN=your_token ./scripts/backup-personal.sh
```

### Windows (Task Scheduler)

1. Open Task Scheduler
2. Create Task → Triggers → Daily at 2:00 AM
3. Action: Start a program
   - Program: `powershell.exe`
   - Arguments: `-ExecutionPolicy Bypass -File "D:\knowledge-graph\scripts\backup-personal.ps1"`
   - Add environment variables:
     - `BACKUP_CLOUD_ENABLED=true`
     - `BACKUP_CLOUD_PROVIDER=yandex`
     - `BACKUP_YANDEX_TOKEN=your_token`

## Verifying Backups

### View backups on Yandex.Disk

1. Open [Yandex.Disk](https://disk.yandex.ru)
2. Navigate to `knowledge-graph-backups` folder
3. Backups will be named: `backup-personal-YYYY-MM-DD.sql.gz`

### Restore from backup

```bash
# Download backup from Yandex.Disk (via web interface or WebDAV)
# Unzip
gunzip backup-personal-2024-05-17.sql.gz

# Restore to PostgreSQL
psql -h localhost -p 5433 -U personal -d knowledge_personal < backup-personal-2024-05-17.sql
```

## Troubleshooting

### Error "BACKUP_YANDEX_TOKEN not set"

Ensure the `BACKUP_YANDEX_TOKEN` environment variable is set:
```bash
echo $BACKUP_YANDEX_TOKEN  # Linux/Mac
echo $env:BACKUP_YANDEX_TOKEN  # PowerShell
```

### Error "Yandex.Disk upload failed"

- Check that the token is correct and not expired
- Ensure you have enough space on Yandex.Disk
- Check internet connection

### Token expired

Yandex.Disk OAuth tokens may expire. Get a new token using the instructions above.

### Error 401 Unauthorized

- Token is incorrect or expired
- Check that the application has access to Yandex.Disk

## Security

- **Never commit** `.env` file with real tokens
- Store token in a secure place (password manager)
- Rotate tokens regularly
- Restrict application permissions to Yandex.Disk only
- Use separate tokens for different environments

## Additional Features

### Changing backup folder

Default backup folder: `/KnowledgeGraphBackups`

To change, edit the configuration in `knowledge-graph.config.json` or use the `BACKUP_YANDEX_FOLDER` environment variable.

### Changing retention period

Change `BACKUP_RETENTION_DAYS` in configuration (default 7 days).

### Backup encryption

For additional security, you can encrypt backups before upload:

```bash
# Encrypt before upload
gpg --symmetric --cipher-algo AES256 backup-personal-2024-05-17.sql.gz

# Decrypt after download
gpg --decrypt backup-personal-2024-05-17.sql.gz.gpg > backup-personal-2024-05-17.sql.gz
```
