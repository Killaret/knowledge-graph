#!/usr/bin/env bash
# Register a crontab entry for running clean_and_compress_lunix.sh
# Usage:
#   ./register_cron.sh --daily 03:00 --path /abs/path/to/repo
#   ./register_cron.sh --weekly Sun 03:00

set -euo pipefail
SCHEDULE="daily"
TIME="03:00"
DAY="Sun"
REPO_PATH="$(pwd)"
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --daily)
      SCHEDULE="daily"; TIME="$2"; shift 2;;
    --weekly)
      SCHEDULE="weekly"; DAY="$2"; TIME="$3"; shift 3;;
    --path)
      REPO_PATH="$2"; shift 2;;
    --dry-run)
      DRY_RUN=true; shift;;
    -h|--help)
      echo "Usage: $0 [--daily HH:MM] [--weekly DOW HH:MM] [--path /path/to/repo] [--dry-run]"; exit 0;;
    *) echo "Unknown arg: $1"; exit 1;;
  esac
done

# compute minute and hour
HH=${TIME%%:*}
MM=${TIME##*:}

SCRIPT="$REPO_PATH/scripts/clean_and_compress_lunix.sh"
if [[ ! -f "$SCRIPT" ]]; then
  echo "Script not found: $SCRIPT" >&2
  exit 1
fi

if [[ "$SCHEDULE" == "daily" ]]; then
  CRON_ENTRY="$MM $HH * * * \"$SCRIPT\" -c --path \"$REPO_PATH/lunix.vhdx\" >> $REPO_PATH/scripts/cleanup_cron.log 2>&1"
else
  # map day name to crontab number (Sun=0)
  case $DAY in
    Sun) D=0;; Mon) D=1;; Tue) D=2;; Wed) D=3;; Thu) D=4;; Fri) D=5;; Sat) D=6;; *) D=0;;
  esac
  CRON_ENTRY="$MM $HH * * $D \"$SCRIPT\" -c --path \"$REPO_PATH/lunix.vhdx\" >> $REPO_PATH/scripts/cleanup_cron.log 2>&1"
fi

if [[ "$DRY_RUN" == "true" ]]; then
  echo "DRY RUN: would add crontab entry:"
  echo "$CRON_ENTRY"
  exit 0
fi

# Install crontab entry (idempotent)
TMPFILE=$(mktemp)
crontab -l 2>/dev/null | grep -v "$SCRIPT" > $TMPFILE || true
echo "$CRON_ENTRY" >> $TMPFILE
crontab $TMPFILE
rm -f $TMPFILE

echo "Crontab entry registered."
