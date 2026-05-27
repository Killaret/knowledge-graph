#!/usr/bin/env bash
# Clean and compress 'lunix' images (Linux / WSL)
# Usage: ./clean_and_compress_lunix.sh [-c|--compress] [-p|--path PATH] [-s|--search] [--dry-run] [-f|--force]

set -euo pipefail
COMPRESS=false
IMAGE_PATH=""
SEARCH=false
DRY_RUN=false
FORCE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    -c|--compress)
      COMPRESS=true; shift;;
    -p|--path)
      IMAGE_PATH="$2"; shift 2;;
    -s|--search)
      SEARCH=true; shift;;
    -h|--help)
      echo "Usage: $0 [-c|--compress] [-p|--path PATH] [-s|--search] [--dry-run] [-f|--force]"
      echo "  -c, --compress    Compress found images"
      echo "  -p, --path PATH   Search in specific path (file or directory)"
      echo "  -s, --search      Search common locations for lunix images"
      echo "  --dry-run         Show what would be done without executing"
      echo "  -f, --force       Compress without confirmation prompts"
      exit 0;;
    --dry-run)
      DRY_RUN=true; shift;;
    -f|--force)
      FORCE=true; shift;;
    *)
      echo "Unknown arg: $1"; exit 1;;
  esac
done

echo "🧰 Clean & Compress lunix (bash)"
echo "$(date '+%H:%M:%S') Starting..."

# Function to print status
print_status() {
    local status="$1"
    local message="$2"
    local color=""
    case "$status" in
        "SUCCESS") color="\033[32m";;
        "WARNING") color="\033[33m";;
        "ERROR") color="\033[31m";;
        *) color="\033[36m";;
    esac
    echo -e "${color}  [${status}] ${message}\033[0m"
}

# Lightweight Docker cleanup
echo ""
echo "1️⃣ Running lightweight Docker cleanup (dangling images, stopped containers)"
if command -v docker >/dev/null 2>&1; then
  if [[ "$DRY_RUN" == "true" ]]; then
    print_status "INFO" "DRY RUN: would run 'docker image prune -f'"
    print_status "INFO" "DRY RUN: would run 'docker container prune -f'"
  else
    echo "-> docker image prune -f"
    docker image prune -f || true
    echo "-> docker container prune -f"
    docker container prune -f || true
    print_status "SUCCESS" "Docker prune executed"
  fi
else
  print_status "WARNING" "docker not found, skipping docker cleanup"
fi

# Locate image(s) if not provided
FOUND_IMAGES=()

if [[ -n "$IMAGE_PATH" ]]; then
    if [[ -f "$IMAGE_PATH" ]]; then
        FOUND_IMAGES+=("$IMAGE_PATH")
        print_status "INFO" "Using provided file: $IMAGE_PATH"
    elif [[ -d "$IMAGE_PATH" ]]; then
        print_status "INFO" "Searching in directory: $IMAGE_PATH"
        while IFS= read -r -d '' file; do
            FOUND_IMAGES+=("$file")
        done < <(find "$IMAGE_PATH" -type f -iname '*lunix*' -print0 2>/dev/null)
        print_status "INFO" "Found ${#FOUND_IMAGES[@]} file(s) in directory"
    else
        print_status "ERROR" "Provided path not found: $IMAGE_PATH"
    fi
elif [[ "$SEARCH" == "true" ]]; then
    echo ""
    echo "2️⃣ Searching common locations for 'lunix' images..."
    
    # Common locations to check
    declare -a CANDIDATES=(
        "$HOME/lunix.vhdx"
        "$HOME/lunix.img"
        "$HOME/.docker/wsl/lunix.vhdx"
        "/mnt/wsl/docker-desktop/lunix.vhdx"
        "/mnt/c/Users/$USER/lunix.vhdx"
        "/d/lunix.vhdx"
        "/c/lunix.vhdx"
    )
    
    # Check specific files first
    for cand in "${CANDIDATES[@]}"; do
        if [[ -f "$cand" ]]; then
            FOUND_IMAGES+=("$cand")
            print_status "INFO" "Found: $cand"
        fi
    done
    
    # Check directories for lunix files
    declare -a DIR_CANDIDATES=(
        "$HOME/.docker/wsl"
        "/mnt/wsl/docker-desktop"
        "/mnt/c/Users/$USER/AppData/Local/Docker/wsl"
        "/mnt/c/Users/$USER/AppData/Local/Packages"
        "/d/images"
        "/c/images"
    )
    
    for dir_cand in "${DIR_CANDIDATES[@]}"; do
        if [[ -d "$dir_cand" ]]; then
            while IFS= read -r -d '' file; do
                # Avoid duplicates
                if [[ ! " ${FOUND_IMAGES[@]} " =~ " ${file} " ]]; then
                    FOUND_IMAGES+=("$file")
                    print_status "INFO" "Found in $dir_cand: $file"
                fi
            done < <(find "$dir_cand" -type f -iname '*lunix*' -print0 2>/dev/null)
        fi
    done
    
    # Fallback: recursive search in home
    if [[ ${#FOUND_IMAGES[@]} -eq 0 ]]; then
        print_status "INFO" "No images in common locations, searching home directory..."
        while IFS= read -r -d '' file; do
            if [[ ! " ${FOUND_IMAGES[@]} " =~ " ${file} " ]]; then
                FOUND_IMAGES+=("$file")
                print_status "INFO" "Found via search: $file"
            fi
        done < <(find "$HOME" -maxdepth 4 -type f -iname '*lunix*' -print0 2>/dev/null)
    fi
    
    if [[ ${#FOUND_IMAGES[@]} -gt 0 ]]; then
        print_status "SUCCESS" "Total found: ${#FOUND_IMAGES[@]} lunix image(s)"
    else
        print_status "WARNING" "No lunix images found in common locations"
    fi
fi

# Remove duplicates and sort
IFS=$'\n' FOUND_IMAGES=($(sort -u <<<"${FOUND_IMAGES[*]}"))

# Compress if requested
if [[ "$COMPRESS" == "true" && ${#FOUND_IMAGES[@]} -gt 0 ]]; then
    echo ""
    echo "3️⃣ Compressing found images..."
    
    for img in "${FOUND_IMAGES[@]}"; do
        echo ""
        echo "Processing: $img"
        
        if [[ ! -f "$img" ]]; then
            print_status "WARNING" "File not accessible: $img"
            continue
        fi
        
        OLD_SIZE=$(stat -f%z "$img" 2>/dev/null || stat -c%s "$img" 2>/dev/null || echo "0")
        OLD_SIZE_GB=$(echo "scale=2; $OLD_SIZE / 1024^3" | bc 2>/dev/null || echo "unknown")
        print_status "INFO" "Original size: ${OLD_SIZE_GB} GB"
        
        if [[ "$DRY_RUN" == "true" ]]; then
            print_status "INFO" "DRY RUN: would compress $img"
            continue
        fi
        
        # Check for qemu-img
        if command -v qemu-img >/dev/null 2>&1; then
            COMPRESSED_FILE="${img}.compressed.qcow2"
            
            if [[ "$FORCE" == "true" ]] || read -p "Proceed with qemu-img compression on $img? (y/N) " -r; then
                if [[ "$REPLY" =~ ^[Yy]$ ]] || [[ "$FORCE" == "true" ]]; then
                    print_status "INFO" "Compressing with qemu-img -> $COMPRESSED_FILE"
                    
                    if qemu-img convert -O qcow2 -c "$img" "$COMPRESSED_FILE" 2>/dev/null; then
                        NEW_SIZE=$(stat -f%z "$COMPRESSED_FILE" 2>/dev/null || stat -c%s "$COMPRESSED_FILE" 2>/dev/null || echo "0")
                        NEW_SIZE_GB=$(echo "scale=2; $NEW_SIZE / 1024^3" | bc 2>/dev/null || echo "unknown")
                        SAVED_GB=$(echo "scale=2; $OLD_SIZE_GB - $NEW_SIZE_GB" | bc 2>/dev/null || echo "unknown")
                        
                        print_status "SUCCESS" "Compression complete: ${OLD_SIZE_GB} GB -> ${NEW_SIZE_GB} GB (saved: ${SAVED_GB} GB)"
                        print_status "INFO" "Compressed file: $COMPRESSED_FILE"
                    else
                        print_status "ERROR" "qemu-img compression failed"
                    fi
                else
                    print_status "WARNING" "Compression aborted by user"
                fi
            fi
        else
            print_status "WARNING" "qemu-img not found. Install qemu-utils for compression"
            print_status "INFO" "Try: sudo apt-get install qemu-utils (Ubuntu/Debian)"
            print_status "INFO" "Or use Windows Compact.exe on Windows host"
        fi
        
        # Try to enable sparse files (Linux)
        if command -v chattr >/dev/null 2>&1; then
            print_status "INFO" "Attempting to enable sparse attribute..."
            if chattr +s "$img" 2>/dev/null; then
                print_status "SUCCESS" "Enabled sparse attribute on $img"
            else
                print_status "WARNING" "Could not set sparse attribute (may require sudo)"
            fi
        fi
    done
elif [[ "$COMPRESS" == "true" && ${#FOUND_IMAGES[@]} -eq 0 ]]; then
    print_status "ERROR" "No images available to compress"
fi

echo ""
print_status "SUCCESS" "Done."
echo "$(date '+%H:%M:%S') Finished."

echo ""
echo "Usage examples:"
echo "  ./clean_and_compress_lunix.sh -s -c              # Search and compress all lunix images"
echo "  ./clean_and_compress_lunix.sh -p /path/to/dir -c -f  # Compress all in directory without prompts"
echo "  ./clean_and_compress_lunix.sh -s -c --dry-run     # Dry run to see what would be compressed"