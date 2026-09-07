# Shared phase tracking for the full test cycle scripts.
# Source this file before using register_phase / write_final_summary / test_any_failed.
#
# PHASE_RESULTS is an associative array:
#   Name -> "skip" or numeric exit code
#
# The main script sets SNAPSHOT_DIR so write_final_summary can report the
# snapshot location without each call site passing it.

if [[ -z ${PHASE_RESULTS+x} ]]; then
    declare -A PHASE_RESULTS
fi

# Associative arrays do not preserve insertion order, so the registration
# order is kept in a parallel indexed array for the final report.
if [[ -z ${PHASE_ORDER+x} ]]; then
    PHASE_ORDER=()
fi

register_phase() {
    local name="$1"
    local code="${2:-0}"
    local skipped="${3:-0}"

    if [[ -z ${PHASE_RESULTS[$name]+_} ]]; then
        PHASE_ORDER+=("$name")
    fi

    if [[ "$skipped" -eq 1 && "$code" -eq 0 ]]; then
        PHASE_RESULTS["$name"]="skip"
        echo "  [SKIP] $name"
        return 0
    fi

    PHASE_RESULTS["$name"]="$code"
    if [[ "$code" -ne 0 ]]; then
        echo "  [FAIL] $name (exit $code)"
    else
        echo "  [PASS] $name"
    fi
}

write_final_summary() {
    local success="$1"

    echo ""
    echo "[Final Summary] Test cycle summary"
    echo "========================================"
    if [[ "$success" == "true" ]]; then
        echo "  TEST CYCLE COMPLETE"
    else
        echo "  TEST CYCLE FAILED"
    fi
    echo "========================================"
    echo ""

    for name in "${PHASE_ORDER[@]}"; do
        local value="${PHASE_RESULTS[$name]}"
        if [[ "$value" == "skip" ]]; then
            echo "  [SKIP] $name"
        elif [[ "$value" =~ ^[0-9]+$ ]]; then
            if [[ "$value" -eq 0 ]]; then
                echo "  [PASS] $name"
            else
                echo "  [FAIL] $name (exit $value)"
            fi
        else
            echo "  [????] $name (value: $value)"
        fi
    done
    echo ""

    if [[ -n "${SNAPSHOT_DIR:-}" ]]; then
        echo "Snapshots saved to: $SNAPSHOT_DIR"
        echo ""
    fi

    if [[ "$success" == "true" ]]; then
        echo "All stacks are stable and isolated testing completed successfully."
    else
        echo "One or more phases failed. See details above."
    fi
}

test_any_failed() {
    local value
    for value in "${PHASE_RESULTS[@]}"; do
        if [[ "$value" != "skip" && "$value" -ne 0 ]]; then
            return 0
        fi
    done
    return 1
}
