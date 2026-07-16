#!/usr/bin/env bash
# Clean up generated log/replay/output files older than retention thresholds.
# Run from repo root.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LOG_RETENTION_DAYS="${1:-30}"
SNAPSHOT_RETENTION_DAYS="${2:-7}"
DRY_RUN="${DRY_RUN:-false}"

remove_if_old() {
    local path="$1"
    local days="$2"
    local label="$3"
    if [[ ! -e "$path" ]]; then
        return
    fi
    local mtime
    mtime=$(stat -c %Y "$path" 2>/dev/null || stat -f %m "$path" 2>/dev/null || echo 0)
    local now
    now=$(date +%s)
    local age_days=$(( (now - mtime) / 86400 ))
    if [[ "$days" -eq 0 ]] || [[ "$age_days" -ge "$days" ]]; then
        if [[ "$DRY_RUN" == "true" ]]; then
            echo "[DRY-RUN] Would delete $label: $path (${age_days} days old)"
        else
            rm -rf "$path"
            echo "Deleted $label: $path"
        fi
    fi
}

echo "Cleaning ad-hoc log files in project root..."
find "$REPO_ROOT" -maxdepth 1 -type f \( \
    -name '*.log' -o -name 'hs_err_pid*.log' -o -name 'replay_pid*.log' -o -name '*.results.txt' \
) -print0 | while IFS= read -r -d '' file; do
    remove_if_old "$file" 0 'root log'
done

echo "Cleaning frontend generated logs..."
find "$REPO_ROOT/frontend" -maxdepth 1 -type f -name '*.log' -print0 2>/dev/null | while IFS= read -r -d '' file; do
    remove_if_old "$file" 0 'frontend log'
done

echo "Cleaning backend generated logs and test result files..."
find "$REPO_ROOT/backend" -maxdepth 1 -type f \( -name '*.log' -o -name '*-test-results.txt' \) -print0 2>/dev/null | while IFS= read -r -d '' file; do
    remove_if_old "$file" 0 'backend artifact'
done

echo "Cleaning old logs in logs/ (keeping .gitkeep and README)..."
find "$REPO_ROOT/logs" -type f ! -name '.gitkeep' ! -name 'README*' -print0 2>/dev/null | while IFS= read -r -d '' file; do
    remove_if_old "$file" "$LOG_RETENTION_DAYS" 'log file'
done

echo "Cleaning old test snapshots in scripts/testing/temp/snapshots/..."
SNAPSHOTS_DIR="$REPO_ROOT/scripts/testing/temp/snapshots"
if [[ -d "$SNAPSHOTS_DIR" ]]; then
    find "$SNAPSHOTS_DIR" -mindepth 1 -maxdepth 1 -type d -print0 | while IFS= read -r -d '' dir; do
        remove_if_old "$dir" "$SNAPSHOT_RETENTION_DAYS" 'snapshot dir'
    done
fi

echo "Done."
