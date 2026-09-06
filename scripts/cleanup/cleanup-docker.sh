#!/bin/bash
# Docker Cleanup Script for Knowledge Graph (Linux/macOS)
# Removes dangling images, stopped containers, unused networks, and build cache
# Usage: bash cleanup-docker.sh [-f|--full] [--remove-volumes] [-o|--optimize-docker]

FULL_CLEANUP=false
OPTIMIZE_DOCKER=false
REMOVE_VOLUMES=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--full)
            FULL_CLEANUP=true
            shift
            ;;
        -o|--optimize-docker)
            OPTIMIZE_DOCKER=true
            shift
            ;;
        --remove-volumes)
            REMOVE_VOLUMES=true
            shift
            ;;
        -h|--help)
            echo "Usage: bash cleanup-docker.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -f, --full              Full system cleanup (aggressive, removes all dangling)"
            echo "  -o, --optimize-docker   Optimize Docker disk space"
            echo "      --remove-volumes    Remove only anonymous dangling volumes"
            echo "  -h, --help              Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "🧹 Knowledge Graph Docker Cleanup"
echo "$(date '+%H:%M:%S') Starting cleanup..."
echo ""

# Function to print status
print_status() {
    local message="$1"
    local status="${2:-INFO}"
    local color=""
    
    case $status in
        SUCCESS)
            color="\033[32m"  # Green
            ;;
        WARNING)
            color="\033[33m"  # Yellow
            ;;
        ERROR)
            color="\033[31m"  # Red
            ;;
        *)
            color="\033[36m"  # Cyan
            ;;
    esac
    
    echo -e "${color}  [$status] $message\033[0m"
}

# 1. Stop running containers
echo "1️⃣  Stopping containers..."
if docker ps -q >/dev/null 2>&1; then
    running=$(docker ps -q 2>/dev/null)
    if [ -n "$running" ]; then
        docker stop $running >/dev/null 2>&1
        print_status "Stopped running containers" "SUCCESS"
    else
        print_status "No running containers" "INFO"
    fi
fi

# 2. Remove dangling images
echo ""
echo "2️⃣  Removing dangling images..."
if docker image prune -f >/dev/null 2>&1; then
    print_status "Dangling images removed" "SUCCESS"
else
    print_status "Failed to remove dangling images" "WARNING"
fi

# 3. Remove stopped containers
echo ""
echo "3️⃣  Removing stopped containers..."
if docker container prune -f >/dev/null 2>&1; then
    print_status "Stopped containers removed" "SUCCESS"
else
    print_status "Failed to remove containers" "WARNING"
fi

# 4. Remove unused networks
echo ""
echo "4️⃣  Removing unused networks..."
if docker network prune -f >/dev/null 2>&1; then
    print_status "Unused networks removed" "SUCCESS"
else
    print_status "Failed to remove networks" "WARNING"
fi

# 5. Clear build cache
echo ""
echo "5️⃣  Clearing Docker build cache..."
if docker builder prune -f >/dev/null 2>&1; then
    print_status "Build cache cleared" "SUCCESS"
else
    print_status "Failed to clear build cache" "WARNING"
fi

# 6. Volumes are preserved by default. They are only removed with --remove-volumes,
# and even then only anonymous dangling volumes are eligible.
echo ""
echo "6️⃣  Volume cleanup..."
if [ "$REMOVE_VOLUMES" != true ]; then
    print_status "Volume cleanup skipped (default safe mode)" "INFO"
else
    removed_volumes=0
    while IFS= read -r volume; do
        case "$volume" in
            *personal*)
                continue
                ;;
        esac
        if ! printf '%s' "$volume" | grep -Eq '^[0-9a-f]{64}$'; then
            continue
        fi
        if docker volume inspect "$volume" --format '{{index .Labels "com.knowledgegraph.protected"}}' 2>/dev/null | grep -q '^true$'; then
            continue
        fi
        if docker volume rm "$volume" >/dev/null 2>&1; then
            removed_volumes=$((removed_volumes + 1))
        fi
    done <<EOF
$(docker volume ls --filter dangling=true --format '{{.Name}}' 2>/dev/null)
EOF
    print_status "Removed $removed_volumes anonymous dangling volumes" "SUCCESS"
fi

# 7. Full cleanup (optional; volumes are still not removed)
if [ "$FULL_CLEANUP" = true ]; then
    echo ""
    echo "7️⃣  Full cleanup mode (removing ALL unused images, not volumes)..."
    if docker system prune -af >/dev/null 2>&1; then
        print_status "Full system cleanup completed" "SUCCESS"
    else
        print_status "Full cleanup completed with warnings" "WARNING"
    fi
fi

# 8. Optimize Docker disk (optional, Linux only)
if [ "$OPTIMIZE_DOCKER" = true ]; then
    echo ""
    echo "8️⃣  Optimizing Docker disk space..."
    
    if command -v docker info &> /dev/null; then
        if [ "$(uname)" = "Linux" ]; then
            print_status "Stopping Docker daemon..." "INFO"
            sudo systemctl stop docker 2>/dev/null || sudo service docker stop 2>/dev/null || true
            sleep 2
            
            # Find Docker data directory (usually /var/lib/docker)
            docker_dir="/var/lib/docker"
            if [ -d "$docker_dir" ]; then
                print_status "Compacting Docker data directory: $docker_dir" "INFO"
                sudo fstrim "$docker_dir" 2>/dev/null || print_status "fstrim not supported or failed" "WARNING"
            fi
            
            print_status "Starting Docker daemon..." "INFO"
            sudo systemctl start docker 2>/dev/null || sudo service docker start 2>/dev/null || true
            sleep 2
            print_status "Docker daemon restarted" "SUCCESS"
        else
            print_status "Docker optimization only supported on Linux" "WARNING"
        fi
    fi
fi

# Show status
echo ""
echo "📊 Docker system status:"
docker system df 2>/dev/null | sed 's/^/  /'

echo ""
echo "✅ Cleanup completed!"
echo "$(date '+%H:%M:%S') Done."
echo ""
echo "ℹ️  Usage:"
echo "  bash cleanup-docker.sh                  # Basic cleanup; preserves all volumes"
echo "  bash cleanup-docker.sh --full           # Remove all unused images; still preserves volumes"
echo "  bash cleanup-docker.sh --remove-volumes # Remove only anonymous dangling volumes"
echo "  bash cleanup-docker.sh --optimize       # Include Docker disk optimization"
echo "  bash cleanup-docker.sh -f -o            # Full image cleanup + disk optimization"
