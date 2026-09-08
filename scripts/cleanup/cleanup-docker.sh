#!/bin/bash
# Docker Cleanup Script for Knowledge Graph (Linux/macOS)
# Removes dangling images, stopped containers, unused networks, and build cache
# Usage: bash cleanup-docker.sh [-f|--full] [--remove-volumes] [-o|--optimize-docker] [-n|--dry-run]
#
# Statuses and the exit code are honest: every step registers its real exit
# code, the script exits non-zero when any step failed, and a failed step
# prints the command's stderr. --dry-run previews each step and changes
# nothing. Steps that can touch volumes (6, 7) refuse without a fresh
# non-empty Personal-stack backup — the rule lives in
# scripts/devops/check-personal-backup.sh and backup-policy.env, the same
# policy the guard-personal-data.py hook enforces.

FULL_CLEANUP=false
OPTIMIZE_DOCKER=false
REMOVE_VOLUMES=false
DRY_RUN=false

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
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            echo "Usage: bash cleanup-docker.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -f, --full              Full system cleanup (removes ALL unused images; preserves volumes)"
            echo "  -o, --optimize-docker   Optimize Docker disk space (Linux only)"
            echo "      --remove-volumes    Remove only anonymous dangling volumes"
            echo "  -n, --dry-run           Preview every step, change nothing"
            echo "  -h, --help              Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../testing/lib/phase-tracking.sh
. "$SCRIPT_DIR/../testing/lib/phase-tracking.sh"
SNAPSHOT_DIR=""
BACKUP_CHECK="$SCRIPT_DIR/../devops/check-personal-backup.sh"

echo "🧹 Knowledge Graph Docker Cleanup"
if [ "$DRY_RUN" = true ]; then
    echo "$(date '+%H:%M:%S') Starting cleanup (dry-run)..."
else
    echo "$(date '+%H:%M:%S') Starting cleanup..."
fi
echo ""

# Runs one docker command, registers its real exit code, and prints the
# captured output on failure so there is something to fix.
run_docker_step() {
    local name="$1"
    shift
    local output
    output=$("$@" 2>&1)
    local code=$?
    if [ "$code" -ne 0 ] && [ -n "$output" ]; then
        printf '%s\n' "$output" | sed 's/^/    /'
    fi
    register_phase "$name" "$code"
    return "$code"
}

# 1. Stop running containers
echo "1️⃣  Stopping containers..."
running=$(docker ps -q 2>&1)
code=$?
if [ "$code" -ne 0 ]; then
    printf '%s\n' "$running" | sed 's/^/    /'
    register_phase "stop-containers" "$code"
elif [ -z "$running" ]; then
    register_phase "stop-containers" 0 1 "no running containers"
elif [ "$DRY_RUN" = true ]; then
    echo "  Would stop $(echo "$running" | wc -l) container(s):"
    printf '%s\n' "$running" | sed 's/^/    /'
    register_phase "stop-containers" 0 1 "dry-run"
else
    run_docker_step "stop-containers" docker stop $running
fi

# 2. Remove dangling images
echo ""
echo "2️⃣  Removing dangling images..."
if [ "$DRY_RUN" = true ]; then
    dangling=$(docker images -f "dangling=true" -q 2>&1)
    code=$?
    if [ "$code" -ne 0 ]; then
        printf '%s\n' "$dangling" | sed 's/^/    /'
        register_phase "prune-dangling-images" "$code"
    else
        count=0; [ -n "$dangling" ] && count=$(echo "$dangling" | wc -l)
        echo "  Would remove $count dangling image(s)"
        [ -n "$dangling" ] && printf '%s\n' "$dangling" | sed 's/^/    /'
        register_phase "prune-dangling-images" 0 1 "dry-run"
    fi
else
    run_docker_step "prune-dangling-images" docker image prune -f
fi

# 3. Remove stopped containers
echo ""
echo "3️⃣  Removing stopped containers..."
if [ "$DRY_RUN" = true ]; then
    stopped=$(docker ps -aq --filter "status=exited" --filter "status=created" --filter "status=dead" 2>&1)
    code=$?
    if [ "$code" -ne 0 ]; then
        printf '%s\n' "$stopped" | sed 's/^/    /'
        register_phase "prune-stopped-containers" "$code"
    else
        count=0; [ -n "$stopped" ] && count=$(echo "$stopped" | wc -l)
        echo "  Would remove $count stopped container(s)"
        [ -n "$stopped" ] && printf '%s\n' "$stopped" | sed 's/^/    /'
        register_phase "prune-stopped-containers" 0 1 "dry-run"
    fi
else
    run_docker_step "prune-stopped-containers" docker container prune -f
fi

# 4. Remove unused networks
echo ""
echo "4️⃣  Removing unused networks..."
if [ "$DRY_RUN" = true ]; then
    networks=$(docker network ls --format "{{.Name}}" 2>&1)
    code=$?
    if [ "$code" -ne 0 ]; then
        printf '%s\n' "$networks" | sed 's/^/    /'
        register_phase "prune-networks" "$code"
    else
        echo "  Would remove unused networks. Current networks:"
        printf '%s\n' "$networks" | sed 's/^/    /'
        register_phase "prune-networks" 0 1 "dry-run"
    fi
else
    run_docker_step "prune-networks" docker network prune -f
fi

# 5. Clear build cache
echo ""
echo "5️⃣  Clearing Docker build cache..."
if [ "$DRY_RUN" = true ]; then
    df_out=$(docker system df 2>&1)
    code=$?
    if [ "$code" -ne 0 ]; then
        printf '%s\n' "$df_out" | sed 's/^/    /'
        register_phase "prune-build-cache" "$code"
    else
        echo "$df_out" | grep "Build Cache" | sed 's/^/  Would clear: /'
        register_phase "prune-build-cache" 0 1 "dry-run"
    fi
else
    run_docker_step "prune-build-cache" docker builder prune -f
fi

# 6. Volumes are preserved by default. With --remove-volumes only anonymous
# dangling volumes are eligible; personal-named and protected-labeled
# volumes are always skipped. A fresh non-empty backup is required first.
echo ""
echo "6️⃣  Volume cleanup..."
if [ "$REMOVE_VOLUMES" != true ]; then
    register_phase "volume-cleanup" 0 1 "default safe mode"
else
    backup_ok=true
    if [ "$DRY_RUN" = true ]; then
        echo "  (dry-run: nothing will be removed)"
        if ! "$BACKUP_CHECK"; then
            echo "  A real --remove-volumes run would stop here"
            backup_ok=false
        fi
    else
        if ! "$BACKUP_CHECK"; then
            register_phase "volume-cleanup" 1
            backup_ok=false
        fi
    fi
    if [ "$backup_ok" = true ]; then
        protected_volumes=$(docker volume ls --filter "label=com.knowledgegraph.protected=true" --format "{{.Name}}" 2>&1)
        ls_code=$?
        dangling_volumes=$(docker volume ls --filter "dangling=true" --format "{{.Name}}" 2>&1)
        [ $? -ne 0 ] && ls_code=1
        if [ "$ls_code" -ne 0 ]; then
            printf '%s\n%s\n' "$protected_volumes" "$dangling_volumes" | sed 's/^/    /'
            register_phase "volume-cleanup" "$ls_code"
        else
            eligible=()
            kept=0
            while IFS= read -r volume; do
                [ -z "$volume" ] && continue
                case "$volume" in
                    *personal*) kept=$((kept + 1)); continue ;;
                esac
                if ! printf '%s' "$volume" | grep -Eq '^[0-9a-f]{64}$'; then
                    kept=$((kept + 1)); continue
                fi
                if printf '%s\n' "$protected_volumes" | grep -qxF "$volume"; then
                    kept=$((kept + 1)); continue
                fi
                eligible+=("$volume")
            done <<EOF
$dangling_volumes
EOF
            if [ "$DRY_RUN" = true ]; then
                echo "  Would remove ${#eligible[@]} anonymous dangling volume(s):"
                printf '    %s\n' "${eligible[@]}"
                echo "  Kept (named / personal / protected): $kept"
                register_phase "volume-cleanup" 0 1 "dry-run"
            else
                removed_volumes=0
                failed_volumes=0
                for volume in "${eligible[@]}"; do
                    rm_output=$(docker volume rm "$volume" 2>&1)
                    if [ $? -eq 0 ]; then
                        removed_volumes=$((removed_volumes + 1))
                    else
                        failed_volumes=$((failed_volumes + 1))
                        printf '%s\n' "$rm_output" | sed 's/^/    /'
                    fi
                done
                echo "  Removed $removed_volumes anonymous dangling volume(s)$([ "$failed_volumes" -gt 0 ] && echo ", $failed_volumes failed")"
                [ "$failed_volumes" -gt 0 ] && register_phase "volume-cleanup" 1 || register_phase "volume-cleanup" 0
            fi
        fi
    fi
fi

# 7. Full cleanup mode. NOTE: step 1 stops every container and step 3
# removes all stopped ones, so by this point NO container remains and
# `docker system prune -af` treats every image on the machine as unused —
# the price is the whole local image store, not just project layers.
if [ "$FULL_CLEANUP" = true ]; then
    echo ""
    echo "7️⃣  Full cleanup mode (removing ALL unused images, not volumes)..."
    images=$(docker images -q 2>&1)
    code=$?
    if [ "$code" -ne 0 ]; then
        printf '%s\n' "$images" | sed 's/^/    /'
        register_phase "full-cleanup" "$code"
    else
        image_count=0; [ -n "$images" ] && image_count=$(echo "$images" | wc -l)
        echo "  Cost: $image_count image(s) will be removed (no containers remain after steps 1+3)"
        docker system df 2>/dev/null | sed 's/^/    /'
        if [ "$DRY_RUN" = true ]; then
            register_phase "full-cleanup" 0 1 "dry-run"
        else
            if ! "$BACKUP_CHECK"; then
                register_phase "full-cleanup" 1
            else
                run_docker_step "full-cleanup" docker system prune -af
            fi
        fi
    fi
else
    register_phase "full-cleanup" 0 1 "not requested"
fi

# 8. Optimize Docker disk (optional, Linux only)
if [ "$OPTIMIZE_DOCKER" = true ]; then
    echo ""
    echo "8️⃣  Optimizing Docker disk space..."

    if [ "$DRY_RUN" = true ]; then
        register_phase "optimize-disk" 0 1 "dry-run"
    elif ! docker info >/dev/null 2>&1; then
        register_phase "optimize-disk" 1
    elif [ "$(uname)" != "Linux" ]; then
        register_phase "optimize-disk" 0 1 "only supported on Linux"
    else
        echo "  Stopping Docker daemon..."
        sudo systemctl stop docker 2>/dev/null || sudo service docker stop 2>/dev/null || true
        sleep 2

        docker_dir="/var/lib/docker"
        if [ -d "$docker_dir" ]; then
            echo "  Compacting Docker data directory: $docker_dir"
            if sudo fstrim "$docker_dir" 2>/dev/null; then
                opt_code=0
            else
                echo "    fstrim not supported or failed"
                opt_code=1
            fi
        else
            echo "    Docker data directory not found: $docker_dir"
            opt_code=1
        fi

        echo "  Starting Docker daemon..."
        sudo systemctl start docker 2>/dev/null || sudo service docker start 2>/dev/null || true
        sleep 2
        register_phase "optimize-disk" "${opt_code:-1}"
    fi
else
    register_phase "optimize-disk" 0 1 "not requested"
fi

# Show status
echo ""
echo "📊 Docker system status:"
docker system df 2>/dev/null | sed 's/^/  /'

if test_any_failed; then
    script_failed=true
else
    script_failed=false
fi
write_final_summary "$([ "$script_failed" = true ] && echo false || echo true)"

echo "ℹ️  Usage:"
echo "  bash cleanup-docker.sh                    # Basic cleanup; preserves all volumes"
echo "  bash cleanup-docker.sh -n|--dry-run       # Preview every step, change nothing"
echo "  bash cleanup-docker.sh -f|--full          # Remove all unused images; still preserves volumes"
echo "  bash cleanup-docker.sh --remove-volumes   # Remove only anonymous dangling volumes"
echo "  bash cleanup-docker.sh -o|--optimize-docker # Include Docker disk optimization"
echo "  bash cleanup-docker.sh -f -o              # Full image cleanup + disk optimization"

if [ "$script_failed" = true ]; then
    exit 1
fi
exit 0
