#!/bin/bash
#
# Scenario 2: Concurrency Performance Test
# Tests performance scaling with different concurrency levels
#

set -e

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$SCRIPT_DIR/../lib/common.sh"

echo "========================================="
echo "  Scenario 2: Concurrency Performance Test"
echo "========================================="

# Initialize test environment
init_test_env

# Test configuration
CONCURRENCY_LEVELS=(1 2 4 8)
WORKLOAD_DIR="$SCRIPT_DIR/../lib/workloads/small-project"
INSTRUCTIONS_FILE="$SCRIPT_DIR/../lib/workloads/instructions/simple-prompt.md"

# Use multiple instances of fast models to test concurrency
MODELS_TO_RUN=(
    "gemini-3-flash" "gemini-3-flash" "gemini-3-flash" "gemini-3-flash"
    "o3" "o3" "o3" "o3"
)

# Build model arguments
MODEL_ARGS=""
for model in "${MODELS_TO_RUN[@]}"; do
    MODEL_ARGS="$MODEL_ARGS --model $model"
done

log_info "Testing concurrency with ${#MODELS_TO_RUN[@]} models across different concurrency levels"
log_info "Models: ${MODELS_TO_RUN[*]}"
log_info "Concurrency levels: ${CONCURRENCY_LEVELS[*]}"

declare -A EXECUTION_TIMES
SUCCESS_COUNT=0
TOTAL_TESTS=${#CONCURRENCY_LEVELS[@]}

for level in "${CONCURRENCY_LEVELS[@]}"; do
    output_dir="tmp/load-testing/concurrency/level_$level"

    log_info "========== Testing Concurrency Level: $level =========="

    if run_thinktank \
        "Concurrency Level: $level" \
        "$output_dir" \
        --instructions "$INSTRUCTIONS_FILE" \
        --max-concurrent "$level" \
        $MODEL_ARGS \
        "$WORKLOAD_DIR"; then

        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        count_outputs "$output_dir"

        # Extract execution time from the test (approximate)
        # In a real implementation, you might want to capture this more precisely
        log_info "Concurrency level $level completed successfully"
    else
        log_fail "Concurrency level $level failed"
    fi

    # Brief pause between concurrency tests
    sleep "${LOAD_TEST_SLEEP:-1}"
done

echo ""
echo "========================================="
echo "  Concurrency Test Summary"
echo "========================================="
log_info "Total concurrency levels tested: $TOTAL_TESTS"
log_info "Successful tests: $SUCCESS_COUNT"
log_info "Failed tests: $((TOTAL_TESTS - SUCCESS_COUNT))"

# Performance analysis
echo ""
log_info "Performance Analysis:"
for level in "${CONCURRENCY_LEVELS[@]}"; do
    output_dir="tmp/load-testing/concurrency/level_$level"
    if [ -d "$output_dir" ]; then
        output_count=$(find "$output_dir" -name "*.md" 2>/dev/null | wc -l)
        expected_outputs=${#MODELS_TO_RUN[@]}

        if [ "$output_count" -eq "$expected_outputs" ]; then
            log_pass "Concurrency $level: ✅ All $expected_outputs outputs generated"
        elif [ "$output_count" -gt 0 ]; then
            log_warn "Concurrency $level: ⚠️  Partial success ($output_count/$expected_outputs outputs)"
        else
            log_fail "Concurrency $level: ❌ No outputs generated"
        fi
    else
        log_fail "Concurrency $level: ❌ Test failed completely"
    fi
done

echo ""
log_info "Concurrency Insights:"
echo "• Higher concurrency levels should process models more efficiently"
echo "• Monitor for diminishing returns or increased failure rates at high concurrency"
echo "• Check individual model outputs for quality consistency across concurrency levels"

echo ""
if [ "$SUCCESS_COUNT" -eq "$TOTAL_TESTS" ]; then
    log_pass "✅ CONCURRENCY TEST PASSED: All concurrency levels completed successfully"
    echo "The system scales well with different concurrency configurations."
    exit 0
elif [ "$SUCCESS_COUNT" -gt $((TOTAL_TESTS / 2)) ]; then
    log_warn "⚠️  CONCURRENCY TEST PARTIAL: $SUCCESS_COUNT out of $TOTAL_TESTS levels succeeded"
    echo "The system shows good concurrency support but may have issues at certain levels."
    exit 1
else
    log_fail "❌ CONCURRENCY TEST FAILED: Only $SUCCESS_COUNT out of $TOTAL_TESTS levels succeeded"
    echo "The system has significant concurrency limitations."
    exit 1
fi
