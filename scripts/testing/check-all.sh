#!/bin/bash

QUICK=0
STRICT=0
MANIFEST_ARG=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --quick) QUICK=1 ;;
        --strict) STRICT=1 ;;
        # Test hook: run a subset manifest (e.g. one docker phase). The
        # workflow sync check is skipped — it validates the real manifest.
        --manifest) MANIFEST_ARG="${2:?--manifest requires a path}"; shift ;;
        *) echo "Usage: $0 [--quick] [--strict] [--manifest PATH]"; exit 2 ;;
    esac
    shift
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MANIFEST_PATH="${MANIFEST_ARG:-$SCRIPT_DIR/core-checks.tsv}"
WORKFLOW_PATH="$PROJECT_ROOT/.github/workflows/_core-checks.yml"
SYNC_SCRIPT="$SCRIPT_DIR/check-core-workflow-sync.mjs"

. "$SCRIPT_DIR/lib/phase-tracking.sh"
PHASE_RESULTS=()
PHASE_ORDER=()
SNAPSHOT_DIR=""
cover_created=0

register_skip() {
    register_phase "$1" 0 1 "$2"
}

tool_unavailable_reason() {
    local tool="$1"
    local name
    IFS='+' read -ra names <<< "$tool"
    for name in "${names[@]}"; do
        if ! command -v "$name" >/dev/null 2>&1; then
            echo "$name is not installed or not in PATH"
            return 0
        fi
    done
    if [[ "$tool" == docker* ]]; then
        # CHECK-ALL-2: `docker info` can fail transiently on a live daemon
        # under load — retry once before believing it, and the skip reason
        # must name the command and show its output.
        local docker_output docker_exit=0 attempt
        for attempt in 1 2; do
            docker_output="$(docker info 2>&1)"
            docker_exit=$?
            [[ $docker_exit -eq 0 ]] && break
            [[ $attempt -eq 1 ]] && sleep 3
        done
        if [[ $docker_exit -ne 0 ]]; then
            local tail
            tail="$(printf '%s\n' "$docker_output" | grep -v '^[[:space:]]*$' | tail -3 | tr '\n' '|' | cut -c1-300)"
            echo "docker info failed (exit $docker_exit): $tail"
            return 0
        fi
    fi
    return 1
}

run_check() {
    local id="$1"
    local cwd="$2"
    local command="$3"
    local old_jwt="${JWT_SECRET-}"
    local had_jwt="${JWT_SECRET+x}"
    local code=0

    cd "$PROJECT_ROOT/$cwd" || return 1
    if [[ "$id" == "backend-coverage" ]]; then
        if [[ ! -f cover.out ]]; then
            echo "cover.out was not produced by backend tests"
            cd "$PROJECT_ROOT" || true
            return 1
        fi
        local required="${command#@coverage:}"
        python3 ../scripts/testing/backend-coverage-total.py cover.out "$required"
        code=$?
    elif [[ "$id" == "frontend-config" ]]; then
        npm run build-config </dev/null
        code=$?
        if [[ $code -eq 0 ]]; then
            git diff --exit-code -- knowledge-graph.config.json
            code=$?
        fi
    else
        if [[ "$id" == "backend-config" && -z "${JWT_SECRET:-}" ]]; then
            export JWT_SECRET=local-check-secret-32-characters-long
        fi
        eval "$command" </dev/null
        code=$?
    fi

    if [[ -n "$had_jwt" ]]; then
        export JWT_SECRET="$old_jwt"
    else
        unset JWT_SECRET
    fi
    cd "$PROJECT_ROOT" || return 1
    return "$code"
}

printf '%s\n' '========================================' '  Knowledge Graph Local Core Checks' '========================================'

if [[ -n "$MANIFEST_ARG" ]]; then
    : # fixture manifest — the sync check applies to the real one only
elif command -v node >/dev/null 2>&1; then
    node "$SYNC_SCRIPT" "$MANIFEST_PATH" "$WORKFLOW_PATH"
    register_phase "Core workflow sync" "$?"
else
    register_skip "Core workflow sync" "node is not installed or not in PATH"
fi

while IFS=$'\t' read -r id name job workflow_step cwd command signature quick_skip tool; do
    [[ "$id" == "id" || -z "$id" ]] && continue
    if [[ $QUICK -eq 1 && "$quick_skip" == "1" ]]; then
        register_skip "$name" "quick mode skips integration checks"
        continue
    fi
    if reason=$(tool_unavailable_reason "$tool"); then
        register_skip "$name" "$reason"
        continue
    fi
    run_check "$id" "$cwd" "$command"
    code=$?
    [[ "$id" == "backend-unit" ]] && cover_created=1
    register_phase "$name" "$code"
done < "$MANIFEST_PATH"

if [[ $cover_created -eq 1 ]]; then
    rm -f "$PROJECT_ROOT/backend/cover.out"
fi

if test_any_failed; then
    write_final_summary false
    exit 1
fi
skipped_count=0
for value in "${PHASE_RESULTS[@]}"; do
    [[ "$value" == skip\|* ]] && skipped_count=$((skipped_count + 1))
done
write_final_summary true
if [[ $STRICT -eq 1 && $skipped_count -gt 0 ]]; then
    echo "Strict mode: $skipped_count skipped phase(s) treated as failure."
    exit 1
fi
exit 0
