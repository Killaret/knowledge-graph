#!/bin/bash
# Checks that a fresh, non-empty Personal-stack backup exists before a
# destructive Docker operation. Mirrors the rule in
# scripts/devops/guard-personal-data.py — the backup globs and the
# freshness threshold come from the same policy, so the two cannot drift.
#
# Exit code 0 — a usable backup exists (prints its path and age).
# Exit code 1 — no usable backup (prints the reason and where to put one).

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || cd "$SCRIPT_DIR/../.." && pwd)"
BACKUP_DIR="$REPO_ROOT/backups"
POLICY_FILE="$SCRIPT_DIR/backup-policy.env"

# Threshold: env var first, then the shared policy file. Refuse rather than
# guess — a cleanup that cannot tell how fresh the backup must be is not safe.
MAX_AGE_HOURS="${KG_BACKUP_MAX_AGE_HOURS:-}"
if [ -z "$MAX_AGE_HOURS" ]; then
    if [ ! -f "$POLICY_FILE" ]; then
        echo "  [ERROR] Backup policy file missing: $POLICY_FILE"
        exit 1
    fi
    MAX_AGE_HOURS="$(grep -E '^[[:space:]]*KG_BACKUP_MAX_AGE_HOURS[[:space:]]*=' "$POLICY_FILE" | head -n1 | cut -d= -f2 | tr -d '[:space:]')"
fi
if [ -z "$MAX_AGE_HOURS" ]; then
    echo "  [ERROR] KG_BACKUP_MAX_AGE_HOURS not set in $POLICY_FILE"
    exit 1
fi

if [ ! -d "$BACKUP_DIR" ]; then
    echo "  [ERROR] No backup directory: $BACKUP_DIR"
    echo "          Run scripts/devops/backup-personal.sh first."
    exit 1
fi

newest=""
newest_mtime=0
for pattern in backup-personal-* personal-volumes-raw-*; do
    for candidate in "$BACKUP_DIR"/$pattern; do
        [ -f "$candidate" ] || continue
        [ -s "$candidate" ] || continue
        mtime=$(stat -c %Y "$candidate" 2>/dev/null || stat -f %m "$candidate" 2>/dev/null || echo 0)
        if [ "$mtime" -gt "$newest_mtime" ]; then
            newest="$candidate"
            newest_mtime=$mtime
        fi
    done
done

if [ -z "$newest" ]; then
    echo "  [ERROR] No non-empty backup matching backup-personal-* or personal-volumes-raw-* in $BACKUP_DIR"
    echo "          Run scripts/devops/backup-personal.sh first."
    exit 1
fi

now=$(date +%s)
age_hours=$(awk -v n="$now" -v m="$newest_mtime" 'BEGIN { printf "%.1f", (n - m) / 3600 }')
too_old=$(awk -v a="$age_hours" -v max="$MAX_AGE_HOURS" 'BEGIN { print (a > max) ? 1 : 0 }')
if [ "$too_old" -eq 1 ]; then
    echo "  [ERROR] Newest backup $(basename "$newest") is ${age_hours} h old; allowed: ${MAX_AGE_HOURS} h"
    echo "          Make a fresh backup via scripts/devops/backup-personal.sh."
    exit 1
fi

size_kb=$(( $(stat -c %s "$newest" 2>/dev/null || stat -f %z "$newest" 2>/dev/null || echo 0) / 1024 ))
echo "  [PASS] Backup $(basename "$newest") (${size_kb} KB, ${age_hours} h old) is fresh enough"
exit 0
