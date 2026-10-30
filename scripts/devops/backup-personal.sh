#!/bin/bash

# Backup script for the personal Knowledge Graph instance.
# Supports two modes:
#   daily  - runs at 02:00 every day, default 7 days local retention
#   weekly - runs at 23:00 every Sunday, default 90 days local retention
#
# Usage:
#   ./scripts/devops/backup-personal.sh [daily|weekly]
#   BACKUP_MODE=weekly ./scripts/devops/backup-personal.sh

set -e

MODE="${1:-${BACKUP_MODE:-daily}}"

resolve_backup_dir() {
  # Container callers set BACKUP_DIR to the in-container path (/backups).
  # Host callers read the canonical default from backup-policy.env and may
  # override it with KG_BACKUP_DIR.
  if [ -n "${BACKUP_DIR:-}" ]; then
    echo "$BACKUP_DIR"
    return
  fi
  local policy_file
  policy_file="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/backup-policy.env"
  local kg_backup_dir="${KG_BACKUP_DIR:-}"
  if [ -z "$kg_backup_dir" ] && [ -f "$policy_file" ]; then
    kg_backup_dir="$(grep -E '^[[:space:]]*KG_BACKUP_DIR[[:space:]]*=' "$policy_file" | head -n1 | cut -d= -f2-)"
    # trim leading/trailing whitespace
    kg_backup_dir="$(printf '%s' "$kg_backup_dir" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
  fi
  if [ -z "$kg_backup_dir" ]; then
    echo "  [ERROR] KG_BACKUP_DIR not set and not found in $policy_file" >&2
    exit 1
  fi
  # Expand a leading ~ to the home directory.
  if [ "${kg_backup_dir:0:1}" = "~" ]; then
    kg_backup_dir="${HOME}${kg_backup_dir#\~}"
  fi
  case "$kg_backup_dir" in
    /*|[A-Za-z]:\\*|[A-Za-z]:/*|\\*) echo "$kg_backup_dir" ;;
    *) echo "$HOME/$kg_backup_dir" ;;
  esac
}
case "$MODE" in
  daily|weekly) ;;
  *) echo "Usage: $0 {daily|weekly}" >&2; exit 1 ;;
esac

# Configuration
BACKUP_DIR="$(resolve_backup_dir)"
TIMESTAMP=$(date +%Y-%m-%d-%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/backup-personal-${MODE}-${TIMESTAMP}.sql"

# Database connection (can be overridden by environment variables)
DB_HOST="${PERSONAL_POSTGRES_HOST:-localhost}"
DB_PORT="${PERSONAL_POSTGRES_PORT:-5433}"
DB_USER="${PERSONAL_POSTGRES_USER:-personal}"
DB_PASSWORD="${PERSONAL_POSTGRES_PASSWORD:-personal_password}"
DB_NAME="${PERSONAL_POSTGRES_DB:-knowledge_personal}"

# Local retention (only local files are deleted, cloud backups are kept forever)
if [ "$MODE" = "weekly" ]; then
  RETENTION_DAYS="${BACKUP_WEEKLY_RETENTION_DAYS:-${BACKUP_RETENTION_DAYS:-90}}"
else
  RETENTION_DAYS="${BACKUP_DAILY_RETENTION_DAYS:-${BACKUP_RETENTION_DAYS:-7}}"
fi

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

echo "Starting ${MODE} backup of personal Knowledge Graph database..."
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

# Optional: Upload to Yandex.Disk via REST API
if [ "${BACKUP_CLOUD_ENABLED:-false}" = "true" ]; then
  echo "Uploading to Yandex.Disk..."
  TOKEN="${BACKUP_YANDEX_OAUTH_TOKEN:-${BACKUP_YANDEX_TOKEN:-}}"
  BACKUP_FOLDER="${BACKUP_YANDEX_FOLDER:-/KnowledgeGraphBackups}"

  if [ -z "$TOKEN" ]; then
    echo "Warning: BACKUP_YANDEX_OAUTH_TOKEN or BACKUP_YANDEX_TOKEN not set"
  else
    REMOTE_PATH="${BACKUP_FOLDER}/$(basename "${BACKUP_FILE}")"
    BASE_URL="https://cloud-api.yandex.net/v1/disk"

    # Ensure the backup folder exists (ignore 409 - already exists)
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
      -X PUT \
      -H "Authorization: OAuth ${TOKEN}" \
      "${BASE_URL}/resources?path=${BACKUP_FOLDER}")

    if [ "$HTTP_STATUS" != "201" ] && [ "$HTTP_STATUS" != "409" ]; then
      echo "Warning: failed to ensure backup folder on Yandex.Disk (HTTP ${HTTP_STATUS})"
    fi

    # Get pre-signed upload URL
    UPLOAD_RESPONSE=$(curl -s -X GET \
      -H "Authorization: OAuth ${TOKEN}" \
      "${BASE_URL}/resources/upload?path=${REMOTE_PATH}&overwrite=true")

    UPLOAD_URL=$(echo "$UPLOAD_RESPONSE" | sed -n 's/.*"href":"\([^"]*\)".*/\1/p')

    if [ -z "$UPLOAD_URL" ]; then
      echo "Warning: failed to get Yandex.Disk upload URL. Response: ${UPLOAD_RESPONSE}"
    else
      # Upload file to the pre-signed URL
      UPLOAD_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -X PUT \
        -T "${BACKUP_FILE}" \
        -H "Content-Type: application/octet-stream" \
        "${UPLOAD_URL}")

      if [ "$UPLOAD_STATUS" = "201" ] || [ "$UPLOAD_STATUS" = "202" ]; then
        echo "Successfully uploaded to Yandex.Disk: ${REMOTE_PATH}"
      else
        echo "Warning: Yandex.Disk upload failed with HTTP ${UPLOAD_STATUS}"
      fi
    fi
  fi
fi

# Clean up old local backups (cloud backups are never deleted)
if [ "${CLEANUP_OLD_BACKUPS:-true}" = "true" ] && [ "$RETENTION_DAYS" -gt 0 ]; then
  echo "Cleaning up old ${MODE} local backups (keeping last ${RETENTION_DAYS} days)..."
  find "${BACKUP_DIR}" -name "backup-personal-${MODE}-*.sql.gz" -mtime +"${RETENTION_DAYS}" -delete
fi

echo "Backup process completed."
