#!/bin/bash
# Cleanup temporary test artifacts
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
removed=0

remove_if_exists() {
    local path="$1"
    if [ -e "$path" ]; then
        rm -rf "$path"
        echo "  Removed $path"
        removed=$((removed + 1))
    fi
}

remove_if_exists "$PROJECT_ROOT/backend/coverage.out"
remove_if_exists "$PROJECT_ROOT/backend/.coverage_tmp"

for f in "$PROJECT_ROOT/backend/"*.cov; do
    [ -e "$f" ] || continue
    rm -f "$f"
    echo "  Removed $f"
    removed=$((removed + 1))
done

remove_if_exists "$PROJECT_ROOT/frontend/coverage"

for f in "$PROJECT_ROOT/logs/test-outputs/"*.log; do
    [ -e "$f" ] || continue
    rm -f "$f"
    echo "  Removed $f"
    removed=$((removed + 1))
done

if [ -d "$PROJECT_ROOT/scripts/testing/temp/snapshots" ]; then
    find "$PROJECT_ROOT/scripts/testing/temp/snapshots" -mindepth 1 -delete
fi

echo "  ✓ Temporary test artifacts cleaned ($removed items)"
