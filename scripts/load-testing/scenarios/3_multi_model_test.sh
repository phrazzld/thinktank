#!/bin/bash
#
# Scenario 3: Multi-Model & Synthesis Test
# Tests realistic multi-provider and synthesis workflows
#

set -e

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$SCRIPT_DIR/../lib/common.sh"

echo "========================================="
echo "  Scenario 3: Multi-Model & Synthesis Test"
echo "========================================="

# Initialize test environment
init_test_env

# Test configuration
WORKLOAD_DIR_LARGE="$SCRIPT_DIR/../lib/workloads/large-project"
WORKLOAD_DIR_SMALL="$SCRIPT_DIR/../lib/workloads/small-project"
COMPLEX_INSTRUCTIONS="$SCRIPT_DIR/../lib/workloads/instructions/complex-prompt.md"
SIMPLE_INSTRUCTIONS="$SCRIPT_DIR/../lib/workloads/instructions/simple-prompt.md"

SUCCESS_COUNT=0
TOTAL_TESTS=3

echo ""
log_info "========== Test 1: Multi-Provider Workload ==========\n"
log_info "Testing models from different providers (Gemini, OpenAI, OpenRouter)"

if run_thinktank \
    "Multi-provider workload (Gemini, OpenAI, OpenRouter)" \
    "tmp/load-testing/multi-model/test1_multi_provider" \
    --instructions "$COMPLEX_INSTRUCTIONS" \
    --model "gemini-2.5-pro" \
    --model "gpt-4.1" \
    --model "openrouter/deepseek/deepseek-chat-v3-0324" \
    "$WORKLOAD_DIR_LARGE"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "Multi-provider test completed successfully"
    count_outputs "tmp/load-testing/multi-model/test1_multi_provider"
else
    log_fail "Multi-provider test failed"
fi

sleep "${LOAD_TEST_SLEEP:-2}"

echo ""
log_info "========== Test 2: Multi-Model with Synthesis ==========\n"
log_info "Testing multiple models with synthesis using a different model"

if run_thinktank \
    "Multi-model with synthesis" \
    "tmp/load-testing/multi-model/test2_synthesis" \
    --instructions "$COMPLEX_INSTRUCTIONS" \
    --model "gemini-2.5-flash" \
    --model "o4-mini" \
    --synthesis-model "gpt-4.1" \
    "$WORKLOAD_DIR_LARGE"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "Multi-model synthesis test completed successfully"
    count_outputs "tmp/load-testing/multi-model/test2_synthesis"

    # Check for synthesis output
    if ls tmp/load-testing/multi-model/test2_synthesis/synthesis*.md 1> /dev/null 2>&1; then
        log_pass "Synthesis output detected"
    else
        log_warn "No synthesis output file found (this may be normal depending on configuration)"
    fi
else
    log_fail "Multi-model synthesis test failed"
fi

sleep "${LOAD_TEST_SLEEP:-2}"

echo ""
log_info "========== Test 3: Cross-Provider Performance Comparison ==========\n"
log_info "Testing same task across different providers for performance comparison"

if run_thinktank \
    "Cross-provider performance comparison" \
    "tmp/load-testing/multi-model/test3_comparison" \
    --instructions "$SIMPLE_INSTRUCTIONS" \
    --model "gemini-2.5-flash" \
    --model "o4-mini" \
    --model "openrouter/deepseek/deepseek-chat-v3-0324:free" \
    --max-concurrent 3 \
    "$WORKLOAD_DIR_SMALL"; then

    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    log_pass "Cross-provider comparison test completed successfully"
    count_outputs "tmp/load-testing/multi-model/test3_comparison"
else
    log_fail "Cross-provider comparison test failed"
fi

echo ""
echo "========================================="
echo "  Multi-Model Test Summary"
echo "========================================="
log_info "Total multi-model tests: $TOTAL_TESTS"
log_info "Successful tests: $SUCCESS_COUNT"
log_info "Failed tests: $((TOTAL_TESTS - SUCCESS_COUNT))"

# Detailed analysis of outputs
echo ""
log_info "Detailed Output Analysis:"

# Test 1 Analysis
test1_dir="tmp/load-testing/multi-model/test1_multi_provider"
if [ -d "$test1_dir" ]; then
    test1_outputs=$(find "$test1_dir" -name "*.md" 2>/dev/null | wc -l)
    log_info "Test 1 (Multi-provider): $test1_outputs output files"
    if [ "$test1_outputs" -eq 3 ]; then
        log_pass "  ✅ All 3 providers generated outputs"
    elif [ "$test1_outputs" -gt 0 ]; then
        log_warn "  ⚠️  Partial success: $test1_outputs/3 providers succeeded"
    else
        log_fail "  ❌ No outputs generated"
    fi
fi

# Test 2 Analysis
test2_dir="tmp/load-testing/multi-model/test2_synthesis"
if [ -d "$test2_dir" ]; then
    test2_outputs=$(find "$test2_dir" -name "*.md" 2>/dev/null | wc -l)
    test2_individual=$(find "$test2_dir" -name "*.md" ! -name "synthesis*" 2>/dev/null | wc -l)
    test2_synthesis=$(find "$test2_dir" -name "synthesis*.md" 2>/dev/null | wc -l)

    log_info "Test 2 (Synthesis): $test2_outputs total outputs"
    log_info "  Individual model outputs: $test2_individual"
    log_info "  Synthesis outputs: $test2_synthesis"

    if [ "$test2_individual" -eq 2 ] && [ "$test2_synthesis" -gt 0 ]; then
        log_pass "  ✅ Complete synthesis workflow succeeded"
    elif [ "$test2_individual" -gt 0 ]; then
        log_warn "  ⚠️  Partial synthesis success"
    else
        log_fail "  ❌ Synthesis workflow failed"
    fi
fi

# Test 3 Analysis
test3_dir="tmp/load-testing/multi-model/test3_comparison"
if [ -d "$test3_dir" ]; then
    test3_outputs=$(find "$test3_dir" -name "*.md" 2>/dev/null | wc -l)
    log_info "Test 3 (Comparison): $test3_outputs output files"
    if [ "$test3_outputs" -eq 3 ]; then
        log_pass "  ✅ All providers completed comparison test"
    elif [ "$test3_outputs" -gt 0 ]; then
        log_warn "  ⚠️  Partial comparison: $test3_outputs/3 providers succeeded"
    else
        log_fail "  ❌ Comparison test failed"
    fi
fi

echo ""
log_info "Multi-Model Testing Insights:"
echo "• Cross-provider compatibility validates thinktank's provider abstraction"
echo "• Synthesis functionality tests the tool's ability to combine multiple perspectives"
echo "• Performance comparison helps identify optimal provider combinations"
echo "• Review individual outputs to assess quality consistency across providers"

echo ""
if [ "$SUCCESS_COUNT" -eq "$TOTAL_TESTS" ]; then
    log_pass "✅ MULTI-MODEL TEST PASSED: All multi-model scenarios completed successfully"
    echo "The system demonstrates excellent multi-provider and synthesis capabilities."
    exit 0
elif [ "$SUCCESS_COUNT" -gt 0 ]; then
    log_warn "⚠️  MULTI-MODEL TEST PARTIAL: $SUCCESS_COUNT out of $TOTAL_TESTS scenarios succeeded"
    echo "The system shows good multi-model support but may have issues with specific configurations."
    exit 1
else
    log_fail "❌ MULTI-MODEL TEST FAILED: All multi-model scenarios failed"
    echo "The system has significant issues with multi-model functionality."
    exit 1
fi
