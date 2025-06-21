#!/bin/bash
#
# Scenario 1: Model Reliability Test
# Tests the reliability of a single model over multiple consecutive runs
#

set -e

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$SCRIPT_DIR/../lib/common.sh"

echo "========================================="
echo "  Scenario 1: Model Reliability Test"
echo "========================================="

# Initialize test environment
init_test_env

# Test configuration
RELIABILITY_RUNS=5
SUCCESS_COUNT=0
FAILED_RUNS=()
TEST_MODEL="gemini-2.5-flash" # A fast and reliable model
WORKLOAD_DIR="$SCRIPT_DIR/../lib/workloads/small-project"
INSTRUCTIONS_FILE="$SCRIPT_DIR/../lib/workloads/instructions/simple-prompt.md"

log_info "Testing $TEST_MODEL reliability over $RELIABILITY_RUNS consecutive runs"
log_info "Workload: small-project with simple prompt"

# Track total time for all runs
total_start_time=$(date +%s.%N)

for i in $(seq 1 $RELIABILITY_RUNS); do
    output_dir="tmp/load-testing/reliability/run_$i"

    log_info "========== Run $i/$RELIABILITY_RUNS =========="

    if run_thinktank \
        "Reliability Run $i/$RELIABILITY_RUNS for $TEST_MODEL" \
        "$output_dir" \
        --instructions "$INSTRUCTIONS_FILE" \
        --model "$TEST_MODEL" \
        "$WORKLOAD_DIR"; then

        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        count_outputs "$output_dir"
    else
        FAILED_RUNS+=("$i")
        log_fail "Run $i failed"
    fi

    # Brief pause between runs to avoid overwhelming the API
    if [ $i -lt $RELIABILITY_RUNS ]; then
        sleep "${LOAD_TEST_SLEEP:-1}"
    fi
done

total_end_time=$(date +%s.%N)
total_duration=$(echo "$total_end_time - $total_start_time" | bc -l)

echo ""
echo "========================================="
echo "  Reliability Test Summary"
echo "========================================="
log_info "Model: $TEST_MODEL"
log_info "Total runs: $RELIABILITY_RUNS"
log_info "Successful runs: $SUCCESS_COUNT"
log_info "Failed runs: $((RELIABILITY_RUNS - SUCCESS_COUNT))"
log_info "Success rate: $(echo "scale=1; $SUCCESS_COUNT * 100 / $RELIABILITY_RUNS" | bc)%"
log_info "Total execution time: $(format_duration "$total_duration")"
log_info "Average time per run: $(format_duration "$(echo "$total_duration / $RELIABILITY_RUNS" | bc -l)")"

if [ ${#FAILED_RUNS[@]} -gt 0 ]; then
    log_warn "Failed runs: ${FAILED_RUNS[*]}"
fi

echo ""
if [ "$SUCCESS_COUNT" -eq "$RELIABILITY_RUNS" ]; then
    log_pass "✅ RELIABILITY TEST PASSED: All $RELIABILITY_RUNS runs succeeded"
    echo "The model demonstrates excellent reliability under sustained load."
    exit 0
elif [ "$SUCCESS_COUNT" -ge $((RELIABILITY_RUNS * 4 / 5)) ]; then
    log_warn "⚠️  RELIABILITY TEST WARNING: $SUCCESS_COUNT out of $RELIABILITY_RUNS runs succeeded ($(echo "scale=1; $SUCCESS_COUNT * 100 / $RELIABILITY_RUNS" | bc)%)"
    echo "The model shows good reliability but may have occasional issues."
    exit 1
else
    log_fail "❌ RELIABILITY TEST FAILED: Only $SUCCESS_COUNT out of $RELIABILITY_RUNS runs succeeded"
    echo "The model reliability is below acceptable threshold."
    exit 1
fi
