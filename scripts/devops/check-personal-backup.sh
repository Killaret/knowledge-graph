#!/bin/bash
# Checks that a fresh, non-empty Personal-stack backup exists before a
# destructive Docker operation. Mirrors the rule in
# scripts/devops/guard-personal-data.py - the backup globs and the
# freshness threshold come from the same policy, so the two cannot drift.
#
# Exit code 0 - a usable backup exists (prints its path and age).
# Exit code 1 - no usable backup (prints the reason and where to put one).

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POLICY_FILE="$SCRIPT_DIR/backup-policy.env"

# Resolve the backup directory: $KG_BACKUP_DIR overrides backup-policy.env.
# A relative tail is expanded from the user's home directory.
KG_BACKUP_DIR="${KG_BACKUP_DIR:-}"
if [ -z "$KG_BACKUP_DIR" ] && [ -f "$POLICY_FILE" ]; then
  KG_BACKUP_DIR="$(grep -E '^[[:space:]]*KG_BACKUP_DIR[[:space:]]*=' "$POLICY_FILE" | head -n1 | cut -d= -f2-)"
  # trim leading/trailing whitespace
  KG_BACKUP_DIR="$(printf '%s' "$KG_BACKUP_DIR" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
fi
if [ -z "$KG_BACKUP_DIR" ]; then
  echo "  [ERROR] KG_BACKUP_DIR not set and not found in $POLICY_FILE"
  exit 1
fi
if [ "${KG_BACKUP_DIR:0:1}" = "~" ]; then
  KG_BACKUP_DIR="${HOME}${KG_BACKUP_DIR#\~}"
fi
case "$KG_BACKUP_DIR" in
  /*|[A-Za-z]:\\*|[A-Za-z]:/*|\\*) BACKUP_DIR="$KG_BACKUP_DIR" ;;
  *) BACKUP_DIR="$HOME/$KG_BACKUP_DIR" ;;
esac

# Threshold: env var first, then the shared policy file. Refuse rather than
# guess - a cleanup that cannot tell how fresh the backup must be is not safe.
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

# A backup counts only if it provably holds data (BACKUP-4): a 913-byte pg
# envelope with zero tables used to pass the freshness check. Mirrors
# scripts/devops/guard-personal-data.py.
is_gzip() { [ "$(od -An -tx1 -N2 "$1" 2>/dev/null | tr -d ' \n')" = "1f8b" ]; }

sql_dump_has_notes() {
    if is_gzip "$1"; then
        gzip -dc "$1" 2>/dev/null | grep -Eq '^(COPY[[:space:]]|INSERT INTO[[:space:]]).*\bnotes\b'
    else
        grep -Eq '^(COPY[[:space:]]|INSERT INTO[[:space:]]).*\bnotes\b' "$1"
    fi
}

custom_dump_has_tables() {
    # No pg_restore means the .dump cannot be verified - refuse, not guess.
    command -v pg_restore >/dev/null 2>&1 || return 1
    pg_restore --list "$1" 2>/dev/null | grep -Eq 'TABLE[[:space:]]+[^[:space:]]+[[:space:]]+notes\b'
}

mongo_dump_has_namespace() {
    grep -aq 'knowledge_personal\.' "$1" && return 0
    if is_gzip "$1"; then
        gzip -dc "$1" 2>/dev/null | head -c 8388608 | grep -aq 'knowledge_personal\.'
        return $?
    fi
    return 1
}

backup_has_data() {
    local f="$1"
    [ -f "$f" ] && [ -s "$f" ] || return 1
    case "$f" in
        *.sql|*.sql.gz) sql_dump_has_notes "$f" ;;
        *.dump)         custom_dump_has_tables "$f" ;;
        *.mongo.gz)     mongo_dump_has_namespace "$f" ;;
        personal-volumes-raw-*) return 0 ;;
        *) return 1 ;;
    esac
}

# A usable backup exists per datastore: Postgres AND MongoDB (BACKUP-4
# item 2). A single fresh file must not cover the other store.
newest_usable() {
    local exclude_suffix="$1"
    shift
    local newest="" newest_mtime=0 candidate mtime
    for pattern in "$@"; do
        for candidate in "$BACKUP_DIR"/$pattern; do
            [ -f "$candidate" ] || continue
            if [ -n "$exclude_suffix" ]; then
                case "$candidate" in *"$exclude_suffix") continue ;; esac
            fi
            backup_has_data "$candidate" || continue
            mtime=$(stat -c %Y "$candidate" 2>/dev/null || stat -f %m "$candidate" 2>/dev/null || echo 0)
            if [ "$mtime" -gt "$newest_mtime" ]; then
                newest="$candidate"
                newest_mtime=$mtime
            fi
        done
    done
    [ -n "$newest" ] && printf '%s %s\n' "$newest" "$newest_mtime"
}

pg_latest=$(newest_usable ".mongo.gz" "backup-personal-*" "personal-volumes-raw-*pgdata*")
mongo_latest=$(newest_usable "" "backup-personal-*.mongo.gz" "personal-volumes-raw-*mongodbdata*")

missing=""
[ -z "$pg_latest" ] && missing="Postgres"
[ -z "$mongo_latest" ] && missing="${missing:+$missing, }MongoDB"
if [ -n "$missing" ]; then
    echo "  [ERROR] No usable backup in $BACKUP_DIR: $missing - a dump must hold real content (notes table / knowledge_personal namespace) or be a raw volume archive"
    echo "          Run scripts/devops/backup-personal.sh first."
    exit 1
fi

now=$(date +%s)
report=""
for entry in "$pg_latest" "$mongo_latest"; do
    newest="${entry% *}"
    newest_mtime="${entry##* }"
    age_hours=$(awk -v n="$now" -v m="$newest_mtime" 'BEGIN { printf "%.1f", (n - m) / 3600 }')
    too_old=$(awk -v a="$age_hours" -v max="$MAX_AGE_HOURS" 'BEGIN { print (a > max) ? 1 : 0 }')
    if [ "$too_old" -eq 1 ]; then
        echo "  [ERROR] Newest backup $(basename "$newest") is ${age_hours} h old; allowed: ${MAX_AGE_HOURS} h"
        echo "          Make a fresh backup via scripts/devops/backup-personal.sh."
        exit 1
    fi
    size_kb=$(( $(stat -c %s "$newest" 2>/dev/null || stat -f %z "$newest" 2>/dev/null || echo 0) / 1024 ))
    report="${report:+$report; }$(basename "$newest") (${size_kb} KB, ${age_hours} h)"
done

echo "  [PASS] Backups fresh enough: $report"
exit 0
