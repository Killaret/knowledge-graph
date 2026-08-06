# Knowledge Graph Backup System

Backup system for personal Knowledge Graph instance with support for local storage and cloud backup to Yandex.Disk.

## 📋 System Overview

The backup system includes:

- **Local scripts**: `scripts/devops/backup-personal.sh` (Linux/Mac) and `scripts/devops/backup-personal.ps1` (Windows)
- **Go service**: `backend/internal/infrastructure/cloud/yandex_disk.go` for Yandex.Disk REST API integration
- **Asynq tasks**: `TypeBackupToCloud` for asynchronous backup uploads, `TypeDatabaseBackup` for full database dumps
- **Docker service**: `backup_scheduler` in `docker-compose.personal.yml` for scheduled daily backups
- **Worker-driven backup**: `backend/cmd/worker` performs local `pg_dump`, compresses the result and uploads it to Yandex.Disk after note changes
- **Configuration**: `knowledge-graph.config.json` backup section

### Backup Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Backup Scheduler (Docker)                   │
│              - Daily scheduled execution                      │
│              - Executes backup-personal.sh                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Backup Script (backup-personal.sh/ps1)          │
│              - PostgreSQL pg_dump                             │
│              - Gzip compression                               │
│              - Upload to Yandex.Disk (optional)               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│               Local Storage                                   │
│               ./backups/backup-personal-YYYY-MM-DD.sql.gz    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼ (if cloud backup enabled)
┌─────────────────────────────────────────────────────────────┐
│               Yandex.Disk (REST API)                         │
│               /KnowledgeGraphBackups/                        │
└─────────────────────────────────────────────────────────────┘

### Event-driven backup flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Note changes (Create/Update/Delete)        │
│                    Import, Bookmarklet, Batch restore          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                 Asynq queue (TypeDatabaseBackup)              │
│                 - 5-minute unique window                      │
│                 - 30-second delay                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                 Worker backup runner                          │
│                 - pg_dump                                     │
│                 - gzip                                        │
│                 - local retention                             │
│                 - Yandex.Disk upload (REST API)               │
└─────────────────────────────────────────────────────────────┘
```

## ⚙️ Backup Configuration

### Step 1: Configure `knowledge-graph.config.json`

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": true,
      "provider": "yandex",
      "yandex": {
        "oauth_token": "your_oauth_token_here",
        "backup_folder": "/KnowledgeGraphBackups",
        "max_backups": 10
      }
    },
    "schedule": "0 2 * * *",
    "retention_days": 7,
    "draft_ttl_hours": 168
  }
}
```

**Configuration Parameters:**

- `local_path` — local directory for backup storage
- `cloud.enabled` — enable cloud backup
- `cloud.provider` — cloud storage provider (only `yandex`)
- `cloud.yandex.oauth_token` — Yandex.Disk OAuth token
- `cloud.yandex.backup_folder` — folder on Yandex.Disk for backups
- `cloud.yandex.max_backups` — maximum number of backups to keep in cloud
- `schedule` — cron schedule (default `0 2 * * *` — daily at 2:00 AM)
- `retention_days` — number of days to keep local backups
- `draft_ttl_hours` — draft TTL in MongoDB in hours

### Step 2: Get Yandex.Disk OAuth Token

#### Creating the Application

1. Go to [Yandex OAuth](https://oauth.yandex.ru/client/new)
2. Fill out the form:
   - **Application Name**: Knowledge Graph Backup
   - **Description**: Database backup system
   - **Website URL**: `http://localhost`
   - **Redirect URI**: `https://oauth.yandex.ru/verification_code`
   - **Permissions (DLS)**: Select "Yandex.Disk REST API"
3. Get your **Client ID**

#### Getting Token via Browser

1. Open in browser:
   ```
   https://oauth.yandex.ru/authorize?response_type=token&client_id=YOUR_CLIENT_ID
   ```
   Replace `YOUR_CLIENT_ID` with your Client ID
2. Log in to Yandex
3. Grant access to Yandex.Disk
4. The token will be in the URL after redirect:
   ```
   https://oauth.yandex.ru/verification_code#access_token=YOUR_TOKEN&...
   ```
5. Copy the token (part after `access_token=`)

#### Getting Token via curl

```bash
curl -X POST "https://oauth.yandex.ru/token" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

### Step 3: Set Environment Variables (Optional)

Add to `.env` file:

```bash
# Enable cloud backup
BACKUP_CLOUD_ENABLED=true
BACKUP_CLOUD_PROVIDER=yandex

# Yandex.Disk OAuth token (used by the worker and Go scheduler)
BACKUP_YANDEX_OAUTH_TOKEN=your_oauth_token_here

# Folder on Yandex.Disk
BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups

# Local backup directory
BACKUP_DIR=./backups

# Clean old backups (true/false)
CLEANUP_OLD_BACKUPS=true
```

The legacy scripts (`scripts/devops/backup-personal.sh` and `backup-personal.ps1`) also accept `BACKUP_YANDEX_TOKEN` as a fallback alias.

## � Event-driven Backup

The worker automatically schedules a full database backup after any note change:

- `POST /api/v1/notes` (create)
- `PUT /api/v1/notes/:id` (update)
- `DELETE /api/v1/notes/:id` (delete)
- `POST /api/v1/notes/batch` (batch delete)
- `POST /api/v1/notes/:id/restore` (restore)
- `POST /api/v1/import/bookmarklet` (bookmarklet)
- `POST /api/v1/import/bookmarks` async batch import

Multiple changes within 5 minutes are deduplicated by Asynq's `Unique` option. The actual dump is delayed by 30 seconds to avoid backing up in the middle of a burst of edits. The worker produces a timestamped file like:

```
backups/backup-personal-YYYY-MM-DD-HHMMSS.sql.gz
```

and uploads it to Yandex.Disk if `BACKUP_CLOUD_ENABLED=true`.

## �🚀 Manual Backup Execution

### Windows (PowerShell)

```powershell
# Set environment variables
$env:BACKUP_CLOUD_ENABLED = "true"
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token_here"
$env:BACKUP_YANDEX_FOLDER = "/KnowledgeGraphBackups"

# Run backup script
.\scripts\devops\backup-personal.ps1
```

### Linux/Mac

```bash
# Set environment variables
export BACKUP_CLOUD_ENABLED=true
export BACKUP_YANDEX_TOKEN="your_oauth_token_here"
export BACKUP_YANDEX_FOLDER="/KnowledgeGraphBackups"

# Run backup script
./scripts/devops/backup-personal.sh
```

### Via Docker Compose (personal instance)

```bash
# Start backup_scheduler service
docker-compose -f docker-compose.personal.yml up backup_scheduler

# Or one-time backup execution
docker-compose -f docker-compose.personal.yml run --rm backup_scheduler
```

## 📅 Automatic Backups

### Via Docker Compose (Recommended)

The `backup_scheduler` service in `docker-compose.personal.yml` automatically runs backups every 24 hours:

```yaml
backup_scheduler:
  image: postgres:16-alpine
  container_name: kg-backup-scheduler
  volumes:
    - ./backups:/backups
    - ./scripts/devops:/scripts:ro
    - ./knowledge-graph.config.json:/config/config.json:ro
  environment:
    BACKUP_DIR: /backups
    BACKUP_CLOUD_ENABLED: ${BACKUP_CLOUD_ENABLED:-false}
    BACKUP_YANDEX_TOKEN: ${BACKUP_YANDEX_TOKEN:-}
    BACKUP_YANDEX_FOLDER: ${BACKUP_YANDEX_FOLDER:-/KnowledgeGraphBackups}
    PERSONAL_POSTGRES_HOST: postgres_personal
    PERSONAL_POSTGRES_PORT: 5432
    PERSONAL_POSTGRES_USER: ${PERSONAL_POSTGRES_USER:-personal}
    PERSONAL_POSTGRES_PASSWORD: ${PERSONAL_POSTGRES_PASSWORD:-personal_password}
    PERSONAL_POSTGRES_DB: ${PERSONAL_POSTGRES_DB:-knowledge_personal}
  command: >
    sh -c "
      apk add --no-cache curl gzip &&
      while true; do
        echo Running scheduled backup at $$(date) &&
        /scripts/devops/backup-personal.sh &&
        echo Next backup in 24 hours &&
        sleep 86400
      done
    "
  depends_on:
    postgres_personal:
      condition: service_healthy
  restart: unless-stopped
```

### Via cron (Linux/Mac)

Add to crontab (`crontab -e`):

```bash
# Daily backup at 2:00 AM
0 2 * * * cd /path/to/knowledge-graph && BACKUP_CLOUD_ENABLED=true BACKUP_YANDEX_TOKEN=your_token ./scripts/devops/backup-personal.sh >> /var/log/kg-backup.log 2>&1
```

### Via Task Scheduler (Windows)

1. Open Task Scheduler
2. Create Task → Triggers → Daily at 2:00 AM
3. Action: Start a program
   - Program: `powershell.exe`
   - Arguments: `-ExecutionPolicy Bypass -File "D:\knowledge-graph\scripts\devops\backup-personal.ps1"`
   - Add environment variables:
     - `BACKUP_CLOUD_ENABLED=true`
     - `BACKUP_YANDEX_TOKEN=your_token`
     - `BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups`

## 🔄 Backup Restoration

### Step 1: Download Backup

**From Yandex.Disk (web interface):**

1. Open [Yandex.Disk](https://disk.yandex.ru)
2. Navigate to `KnowledgeGraphBackups` folder
3. Download the desired backup: `backup-personal-YYYY-MM-DD.sql.gz`

**From Yandex.Disk (via WebDAV):**

```bash
# Download via curl
TOKEN="your_oauth_token"
BACKUP_DATE="2024-05-17"
curl -X GET "https://webdav.yandex.ru/KnowledgeGraphBackups/backup-personal-${BACKUP_DATE}.sql.gz" \
  -H "Authorization: OAuth ${TOKEN}" \
  --output backup-personal-${BACKUP_DATE}.sql.gz
```

**From local storage:**

```bash
# Backups are in ./backups/ folder
ls ./backups/
```

### Step 2: Extract Backup

```bash
gunzip backup-personal-2024-05-17.sql.gz
```

### Step 3: Restore to PostgreSQL

```bash
# Restore to personal database
psql -h localhost -p 5433 -U personal -d knowledge_personal < backup-personal-2024-05-17.sql

# Or via Docker
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal < backup-personal-2024-05-17.sql
```

### Step 4: Verify Restoration

```bash
# Connect to database and verify data
docker exec -it kg-postgres-personal psql -U personal -d knowledge_personal

# Inside psql
SELECT COUNT(*) FROM notes;
SELECT COUNT(*) FROM links;
\q
```

## 📊 Backup Management

### View Backup List

**Local backups:**

```bash
ls -lh ./backups/
```

**Yandex.Disk backups (web interface):**

1. Open [Yandex.Disk](https://disk.yandex.ru)
2. Navigate to `KnowledgeGraphBackups` folder

**Yandex.Disk backups (via API):**

Go service `YandexBackupService` supports `ListBackups` method to get file list:

```go
files, err := service.ListBackups(ctx, "")
// files contains list of backup filenames
```

### Delete Old Backups

**Automatic cleanup:**

Backup scripts automatically delete local backups older than 7 days (configurable via `CLEANUP_OLD_BACKUPS` and `retention_days`).

**Manual deletion of local backups:**

```bash
# Delete backups older than 30 days
find ./backups -name "backup-personal-*.sql.gz" -mtime +30 -delete
```

**Delete backups from Yandex.Disk:**

1. Open [Yandex.Disk](https://disk.yandex.ru)
2. Navigate to `KnowledgeGraphBackups` folder
3. Delete unwanted files

Go service supports `DeleteBackup` method for programmatic deletion:

```go
err := service.DeleteBackup(ctx, "backup-personal-2024-05-17.sql.gz")
```

## 🔍 Troubleshooting

### Error "BACKUP_YANDEX_TOKEN not set"

**Cause:** Environment variable `BACKUP_YANDEX_TOKEN` is not set.

**Solution:**

```bash
# Linux/Mac
export BACKUP_YANDEX_TOKEN="your_token"
echo $BACKUP_YANDEX_TOKEN

# Windows PowerShell
$env:BACKUP_YANDEX_TOKEN = "your_token"
echo $env:BACKUP_YANDEX_TOKEN
```

### Error "Yandex.Disk upload failed"

**Possible causes:**

1. Token is invalid or expired
2. Not enough space on Yandex.Disk
3. Internet connection issues
4. Incorrect folder path on Yandex.Disk

**Solution:**

1. Check token and get new one if necessary
2. Check available space on Yandex.Disk (free tier - 10 GB)
3. Check internet connection
4. Ensure `KnowledgeGraphBackups` folder exists or can be created

### Error "pg_dump: command not found"

**Cause:** `pg_dump` utility is not installed or not in PATH.

**Solution:**

```bash
# Install PostgreSQL (includes pg_dump)
# Ubuntu/Debian
sudo apt-get install postgresql-client

# macOS
brew install postgresql

# Windows
# Download PostgreSQL from https://www.postgresql.org/download/windows/
```

### Error "Permission denied" when writing backup

**Cause:** Insufficient permissions to write to backup directory.

**Solution:**

```bash
# Create backup directory with proper permissions
mkdir -p ./backups
chmod 755 ./backups
```

### Error 401 Unauthorized when uploading to Yandex.Disk

**Cause:** Token is invalid or expired.

**Solution:**

1. Get new OAuth token following instructions above
2. Update `BACKUP_YANDEX_TOKEN` in `.env` file
3. Restart backup service

### Backup created locally but not uploaded to cloud

**Cause:** Cloud backup is disabled or misconfigured.

**Solution:**

```bash
# Check if cloud backup is enabled
echo $BACKUP_CLOUD_ENABLED  # should be "true"

# Check token
echo $BACKUP_YANDEX_TOKEN

# Check configuration in knowledge-graph.config.json
cat knowledge-graph.config.json | grep -A 10 backup
```

## 🔒 Security

- **Never commit** `.env` file with real tokens
- Store token in a secure location (password manager)
- Regularly rotate OAuth token
- Limit application permissions to Yandex.Disk only
- Use separate tokens for different environments (dev, staging, prod)
- Encrypt backups when storing in sensitive environments
- Regularly test backup restoration

## 📈 Monitoring & Logging

### View Backup Logs

**Docker service:**

```bash
docker logs kg-backup-scheduler
```

**Local execution:**

```bash
# Redirect output to file
./scripts/devops/backup-personal.sh >> /var/log/kg-backup.log 2>&1

# View logs
tail -f /var/log/kg-backup.log
```

### Backup Metrics

Recommended to monitor:

- Frequency of successful backups
- Backup size
- Backup execution time
- Available space on Yandex.Disk
- Number of stored backups

## 📚 Additional Documentation

- [Yandex.Disk Backup Setup (detailed)](docs/YANDEX_DISK_BACKUP.md) — detailed Yandex.Disk configuration guide
- [Cloud Backup Setup](docs/CLOUD_BACKUP_SETUP.md) — general cloud backup information
- [System Configuration](docs/CONFIGURATION_EN.md) — complete configuration guide
- [Architecture](docs/ARCHITECTURE.md) — backup system architecture

## 🆘 Support

If you encounter issues with backup system setup or usage:

1. Check this document for solutions
2. Review [docs/YANDEX_DISK_BACKUP.md](docs/YANDEX_DISK_BACKUP.md)
3. Check system logs
4. Create issue in project repository with problem description
