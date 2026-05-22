#!/usr/bin/env bash
# Clean and compress 'lunix' image (Linux / WSL)
# Usage: ./clean_and_compress_lunix.sh [-c] [-p path_to_image]

set -euo pipefail
COMPRESS=false
IMAGE_PATH=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case $1 in
    -c|--compress)
      COMPRESS=true; shift;;
    -p|--path)
      IMAGE_PATH="$2"; shift 2;;
    -s|--search)
      IMAGE_PATH=""; shift;;
    -h|--help)
      echo "Usage: $0 [-c|--compress] [-p|--path IMAGE] [--dry-run]"; exit 0;;
    --dry-run)
      DRY_RUN=true; shift;;
    *) echo "Unknown arg: $1"; exit 1;;
  esac
done

echo "🧰 Clean & Compress lunix (bash)"

# Lightweight Docker cleanup
if command -v docker >/dev/null 2>&1; then
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY RUN: would run 'docker image prune -f'"
    echo "DRY RUN: would run 'docker container prune -f'"
  else
    echo "-> docker image prune -f"
    docker image prune -f || true
    echo "-> docker container prune -f"
    docker container prune -f || true
  fi
else
  echo "docker not found, skipping docker cleanup"
fi

# Locate image if not provided
if [[ -z "$IMAGE_PATH" ]]; then
  echo "Searching for lunix image in common locations..."
  candid="$(pwd)/lunix.vhdx"
  if [[ -f "$candid" ]]; then
    IMAGE_PATH="$candid"
  fi
  if [[ -z "$IMAGE_PATH" ]]; then
    # try ~/lunix.*
    found=$(find "$HOME" -maxdepth 3 -type f -iname '*lunix*' | head -n 1 || true)
    if [[ -n "$found" ]]; then
      IMAGE_PATH="$found"
    fi
  fi
fi

if [[ -z "$IMAGE_PATH" ]]; then
  echo "No lunix image found. Use -p /path/to/image to specify."
  exit 0
fi

echo "Found image: $IMAGE_PATH"

if [[ "$COMPRESS" == "true" ]]; then
  if command -v qemu-img >/dev/null 2>&1; then
    OUT="$IMAGE_PATH.compressed.qcow2"
    if [[ "$DRY_RUN" == "true" ]]; then
      echo "DRY RUN: would run 'qemu-img convert -O qcow2 -c $IMAGE_PATH $OUT'"
    else
      echo "Compressing with qemu-img -> $OUT"
      qemu-img convert -O qcow2 -c "$IMAGE_PATH" "$OUT"
      echo "Compressed saved to $OUT"
    fi
  else
    echo "qemu-img not found. Install qemu-utils or use host tools for compression."
    exit 1
  fi
fi

echo "✅ Done."