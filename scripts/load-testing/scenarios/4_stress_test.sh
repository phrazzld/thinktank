#!/bin/bash
#
# Scenario 4: Rate Limiting & Error Handling Stress Test
# Stress tests rate limiting and error handling for partial success
#

set -e

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$SCRIPT_DIR/../lib/common.sh"

echo "========================================="
echo "  Scenario 4: Rate Limiting & Error Handling Stress Test"
echo "========================================="

# Initialize test environment
init_test_env

# Test configuration
WORKLOAD_DIR_SMALL="$SCRIPT_DIR/../lib/workloads/small-project"
SIMPLE_INSTRUCTIONS="$SCRIPT_DIR/../lib/workloads/instructions/simple-prompt.md"

SUCCESS_COUNT=0
TOTAL_TESTS=3

echo ""
log_info "========== Test 1: Rate Limiting Stress Test ==========\n"
log_info "Testing system behavior under intentionally low rate limits"
log_warn "This test may take several minutes due to rate limiting..."

if run_thinktank \
    "Rate limit stress test (--rate-limit 10)" \
    "tmp/load-testing/stress/test1_rate_limit" \
    --instructions "$SIMPLE_INSTRUCTIONS" \
    --model "gemini-3-flash" --model "gemini-3-flash" --model "gemini-3-flash" \
    --model "o3" --model "o3" --model "o3" \
    --max-concurrent 6 \
    --rate-limit 10 \
    "$WORKLOAD_DIR_SMALL"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "Rate limiting stress test completed successfully"
    count_outputs "tmp/load-testing/stress/test1_rate_limit"
    log_info "The system gracefully handled rate limiting constraints"
else
    log_warn "Rate limiting stress test failed - this may be expected under extreme constraints"
    log_info "Check if the system failed gracefully or crashed unexpectedly"
fi

sleep "${LOAD_TEST_SLEEP:-2}"

echo ""
log_info "========== Test 2: Partial Failure with Recovery ==========\n"
log_info "Testing partial failure scenarios with --partial-success-ok flag"

if run_thinktank_partial \
    "Partial failure with --partial-success-ok flag" \
    "tmp/load-testing/stress/test2_partial_failure" \
    --instructions "$SIMPLE_INSTRUCTIONS" \
    --model "gemini-3-flash" \
    --model "this-is-not-a-real-model-name" \
    --model "gpt-5.2" \
    --partial-success-ok \
    "$WORKLOAD_DIR_SMALL"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "Partial failure test completed successfully"
    count_outputs "tmp/load-testing/stress/test2_partial_failure"
    log_info "The system handled partial failures gracefully"
else
    log_fail "Partial failure test failed unexpectedly"
fi

sleep "${LOAD_TEST_SLEEP:-2}"

echo ""
log_info "========== Test 3: High Concurrency with Resource Constraints ==========\n"
log_info "Testing high concurrency with many models to stress system resources"

if run_thinktank \
    "High concurrency resource stress test" \
    "tmp/load-testing/stress/test3_resource_stress" \
    --instructions "$SIMPLE_INSTRUCTIONS" \
    --model "gemini-3-flash" --model "gemini-3-flash" --model "gemini-3-flash" \
    --model "o3" --model "o3" --model "o3" \
    --model "openrouter/deepseek/deepseek-chat-v3-0324:free" \
    --model "openrouter/deepseek/deepseek-chat-v3-0324:free" \
    --max-concurrent 8 \
    --timeout 600 \
    "$WORKLOAD_DIR_SMALL"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "High concurrency stress test completed successfully"
    count_outputs "tmp/load-testing/stress/test3_resource_stress"
    log_info "The system handled high concurrency well"
else
    log_warn "High concurrency stress test failed - system may have resource limitations"
fi

echo ""
echo "========================================="
echo "  Stress Test Summary"
echo "========================================="
log_info "Total stress tests: $TOTAL_TESTS"
log_info "Successful tests: $SUCCESS_COUNT"
log_info "Failed tests: $((TOTAL_TESTS - SUCCESS_COUNT))"

# Detailed analysis of stress test results
echo ""
log_info "Detailed Stress Test Analysis:"

# Test 1 Analysis - Rate Limiting
test1_dir="tmp/load-testing/stress/test1_rate_limit"
if [ -d "$test1_dir" ]; then
    test1_outputs=$(find "$test1_dir" -name "*.md" 2>/dev/null | wc -l)
    expected_outputs=6
    log_info "Test 1 (Rate Limiting): $test1_outputs/$expected_outputs outputs"

    if [ "$test1_outputs" -eq "$expected_outputs" ]; then
        log_pass "  ✅ All models completed despite rate limiting"
        log_info "  → System demonstrates excellent rate limit handling"
    elif [ "$test1_outputs" -gt $((expected_outputs / 2)) ]; then
        log_warn "  ⚠️  Partial success under rate limiting ($test1_outputs/$expected_outputs)"
        log_info "  → System shows resilience but may need rate limit tuning"
    else
        log_warn "  ⚠️  Significant impact from rate limiting ($test1_outputs/$expected_outputs)"
        log_info "  → Consider adjusting rate limits for this configuration"
    fi
fi

# Test 2 Analysis - Partial Failure
test2_dir="tmp/load-testing/stress/test2_partial_failure"
if [ -d "$test2_dir" ]; then
    test2_outputs=$(find "$test2_dir" -name "*.md" 2>/dev/null | wc -l)
    expected_valid_outputs=2  # Only 2 valid models

    log_info "Test 2 (Partial Failure): $test2_outputs outputs generated"

    if [ "$test2_outputs" -eq "$expected_valid_outputs" ]; then
        log_pass "  ✅ Perfect partial failure handling - only valid models produced outputs"
        log_info "  → System correctly filtered failed models and continued with successful ones"
    elif [ "$test2_outputs" -gt 0 ] && [ "$test2_outputs" -lt 3 ]; then
        log_pass "  ✅ Good partial failure handling - some valid outputs generated"
        log_info "  → System demonstrated resilience to individual model failures"
    else
        log_fail "  ❌ Unexpected partial failure behavior"
        log_info "  → System may not be handling model failures correctly"
    fi
fi

# Test 3 Analysis - Resource Stress
test3_dir="tmp/load-testing/stress/test3_resource_stress"
if [ -d "$test3_dir" ]; then
    test3_outputs=$(find "$test3_dir" -name "*.md" 2>/dev/null | wc -l)
    expected_outputs=8

    log_info "Test 3 (Resource Stress): $test3_outputs/$expected_outputs outputs"

    if [ "$test3_outputs" -eq "$expected_outputs" ]; then
        log_pass "  ✅ Excellent performance under high concurrency and resource stress"
        log_info "  → System scales well and handles resource pressure effectively"
    elif [ "$test3_outputs" -gt $((expected_outputs * 3 / 4)) ]; then
        log_pass "  ✅ Good performance under stress ($test3_outputs/$expected_outputs)"
        log_info "  → System shows strong resilience with minor resource limitations"
    elif [ "$test3_outputs" -gt $((expected_outputs / 2)) ]; then
        log_warn "  ⚠️  Moderate performance under stress ($test3_outputs/$expected_outputs)"
        log_info "  → System experiences resource pressure but continues functioning"
    else
        log_warn "  ⚠️  Significant performance degradation under stress ($test3_outputs/$expected_outputs)"
        log_info "  → System may need resource optimization or concurrency limits"
    fi
fi

echo ""
log_info "Stress Testing Insights:"
echo "• Rate limiting tests validate graceful degradation under API constraints"
echo "• Partial failure tests ensure system resilience when some models fail"
echo "• Resource stress tests identify system scalability limits"
echo "• These tests help establish operational boundaries for production use"

echo ""
if [ "$SUCCESS_COUNT" -eq "$TOTAL_TESTS" ]; then
    log_pass "✅ STRESS TEST PASSED: All stress scenarios completed successfully"
    echo "The system demonstrates excellent resilience under various stress conditions."
    exit 0
elif [ "$SUCCESS_COUNT" -eq 2 ]; then
    log_warn "⚠️  STRESS TEST MOSTLY PASSED: $SUCCESS_COUNT out of $TOTAL_TESTS scenarios succeeded"
    echo "The system shows good stress resistance with minor limitations under extreme conditions."
    exit 0
elif [ "$SUCCESS_COUNT" -eq 1 ]; then
    log_warn "⚠️  STRESS TEST PARTIAL: $SUCCESS_COUNT out of $TOTAL_TESTS scenarios succeeded"
    echo "The system has moderate stress resistance but may need optimization."
    exit 1
else
    log_fail "❌ STRESS TEST FAILED: All stress scenarios failed"
    echo "The system has significant issues under stress conditions."
    exit 1
fi
