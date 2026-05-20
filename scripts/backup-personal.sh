#!/bin/bash

# Backup script for personal Knowledge Graph instance
# Usage: ./scripts/backup-personal.sh

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +%Y-%m-%d)
BACKUP_FILE="${BACKUP_DIR}/backup-personal-${TIMESTAMP}.sql"

# Database connection (can be overridden by environment variables)
DB_HOST="${PERSONAL_POSTGRES_HOST:-localhost}"
DB_PORT="${PERSONAL_POSTGRES_PORT:-5433}"
DB_USER="${PERSONAL_POSTGRES_USER:-personal}"
DB_PASSWORD="${PERSONAL_POSTGRES_PASSWORD:-personal_password}"
DB_NAME="${PERSONAL_POSTGRES_DB:-knowledge_personal}"

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

echo "Starting backup of personal Knowledge Graph database..."
echo "Backup file: ${BACKUP_FILE}"

# Set PGPASSWORD for pg_dump
export PGPASSWORD="${DB_PASSWORD}"

# Perform pg_dump
pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-acl \
    --verbose \
    > "${BACKUP_FILE}"

# Unset PGPASSWORD
unset PGPASSWORD

# Compress the backup
gzip "${BACKUP_FILE}"
BACKUP_FILE="${BACKUP_FILE}.gz"

echo "Backup completed successfully: ${BACKUP_FILE}"

# Optional: Upload to Yandex.Disk via WebDAV
if [ "${BACKUP_CLOUD_ENABLED:-false}" = "true" ] && [ "${BACKUP_CLOUD_PROVIDER:-r2}" = "yandex" ]; then
    echo "Uploading to Yandex.Disk..."
    TOKEN="${BACKUP_YANDEX_TOKEN:-}"
    if [ -z "$TOKEN" ]; then
        echo "Warning: BACKUP_YANDEX_TOKEN not set"
    else
        REMOTE_PATH="/knowledge-graph-backups/$(basename "${BACKUP_FILE}")"
        URL="https://webdav.yandex.ru${REMOTE_PATH}"

        curl -X PUT "${URL}" \
            -H "Authorization: OAuth ${TOKEN}" \
            --upload-file "${BACKUP_FILE}" \
            --connect-timeout 30 \
            --max-time 300 \
            && echo "Successfully uploaded to Yandex.Disk: ${REMOTE_PATH}" \
            || echo "Warning: Yandex.Disk upload failed"
    fi
fi

# Optional: Trigger cloud backup if backend is running (for R2)
if [ "${CLOUD_BACKUP_ENABLED:-false}" = "true" ] && [ "${BACKUP_CLOUD_PROVIDER:-r2}" = "r2" ]; then
    echo "Triggering cloud backup..."
    curl -X POST http://localhost:8081/api/v1/admin/backup/cloud \
        -H "Content-Type: application/json" \
        -d "{\"local_path\": \"${BACKUP_FILE}\"}" \
        || echo "Warning: Cloud backup trigger failed"
fi

# Optional: Clean up old backups (keep last 7 days)
if [ "${CLEANUP_OLD_BACKUPS:-true}" = "true" ]; then
    echo "Cleaning up old backups (keeping last 7 days)..."
    find "${BACKUP_DIR}" -name "backup-personal-*.sql.gz" -mtime +7 -delete
fi

echo "Backup process completed."
