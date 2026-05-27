#!/usr/bin/env bash
# Simple lint for repo scripts: shellcheck for bash scripts (if available)
set -euo pipefail
if command -v shellcheck >/dev/null 2>&1; then
  echo "Running shellcheck on scripts/*.sh"
  shellcheck scripts/*.sh || true
else
  echo "shellcheck not installed; skipping shell lint"
fi
