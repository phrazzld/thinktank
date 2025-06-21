#!/bin/bash
#
# Main orchestrator for thinktank load tests.
# Runs all load testing scenarios and provides comprehensive summary.
#

set -e

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "========================================="
echo "  Thinktank Load Testing Suite"
echo "========================================="
echo "Project: $(basename "$PROJECT_ROOT")"
echo "Test suite: $(basename "$SCRIPT_DIR")"
echo "Timestamp: $(date)"
echo ""

# Source config and helpers
if [ -f "$SCRIPT_DIR/config.sh" ]; then
    source "$SCRIPT_DIR/config.sh"
    echo "✅ Configuration loaded from config.sh"
else
    echo "❌ ERROR: Configuration file not found."
    echo ""
    echo "Please copy config.sh.example to config.sh and fill it out:"
    echo "  cp $SCRIPT_DIR/config.sh.example $SCRIPT_DIR/config.sh"
    echo "  # Edit config.sh with your API keys and settings"
    echo ""
    exit 1
fi

# Change to project root for consistent execution
cd "$PROJECT_ROOT"

# Check for prerequisites
echo ""
echo "Checking prerequisites..."

if ! command -v "$THINKTANK_BINARY" &> /dev/null; then
    echo "❌ ERROR: thinktank binary not found at '$THINKTANK_BINARY'"
    echo ""
    echo "Please build the binary first:"
    echo "  go build -o thinktank cmd/thinktank/main.go"
    echo ""
    exit 1
fi
echo "✅ thinktank binary found: $THINKTANK_BINARY"

if ! command -v bc &> /dev/null; then
    echo "❌ ERROR: 'bc' command not found. Please install it for timing calculations."
    exit 1
fi
echo "✅ bc (calculator) available"

# Check API keys (warn if missing, but allow continuation for dry-run testing)
api_key_warnings=0
if [ -z "$GEMINI_API_KEY" ] || [ "$GEMINI_API_KEY" = "your-gemini-api-key" ]; then
    echo "⚠️  WARNING: GEMINI_API_KEY not properly set in config.sh"
    ((api_key_warnings++))
fi

if [ -z "$OPENAI_API_KEY" ] || [ "$OPENAI_API_KEY" = "your-openai-api-key" ]; then
    echo "⚠️  WARNING: OPENAI_API_KEY not properly set in config.sh"
    ((api_key_warnings++))
fi

if [ -z "$OPENROUTER_API_KEY" ] || [ "$OPENROUTER_API_KEY" = "your-openrouter-api-key" ]; then
    echo "⚠️  WARNING: OPENROUTER_API_KEY not properly set in config.sh"
    ((api_key_warnings++))
fi

if [ $api_key_warnings -gt 0 ]; then
    echo ""
    echo "⚠️  API Key Warning: $api_key_warnings provider API keys are not properly configured."
    echo "Some tests may fail. Please update config.sh with valid API keys."
    echo ""
    read -p "Do you want to continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Exiting. Please configure API keys and try again."
        exit 1
    fi
    echo "Continuing with partial API key configuration..."
fi

echo "✅ Prerequisites check completed"

# Clean up any previous test results
echo ""
echo "Preparing test environment..."
rm -rf tmp/load-testing
mkdir -p tmp/load-testing
echo "✅ Test environment prepared"

# Track overall results
declare -a SCENARIO_RESULTS
declare -a SCENARIO_NAMES
declare -a SCENARIO_DURATIONS
TOTAL_SCENARIOS=4
PASSED_SCENARIOS=0

# Get overall start time
overall_start_time=$(date +%s.%N)

echo ""
echo "========================================="
echo "  Running Load Test Scenarios"
echo "========================================="

# Scenario 1: Reliability Test
echo ""
scenario_start_time=$(date +%s.%N)
echo "🔄 Running Scenario 1/4: Model Reliability Test..."
if "$SCRIPT_DIR/scenarios/1_reliability_test.sh"; then
    SCENARIO_RESULTS[0]="PASS"
    PASSED_SCENARIOS=$((PASSED_SCENARIOS + 1))
    echo "✅ Scenario 1 PASSED"
else
    SCENARIO_RESULTS[0]="FAIL"
    echo "❌ Scenario 1 FAILED"
fi
scenario_end_time=$(date +%s.%N)
SCENARIO_DURATIONS[0]=$(echo "$scenario_end_time - $scenario_start_time" | bc -l)
SCENARIO_NAMES[0]="Model Reliability Test"

# Scenario 2: Concurrency Test
echo ""
scenario_start_time=$(date +%s.%N)
echo "🔄 Running Scenario 2/4: Concurrency Performance Test..."
if "$SCRIPT_DIR/scenarios/2_concurrency_test.sh"; then
    SCENARIO_RESULTS[1]="PASS"
    PASSED_SCENARIOS=$((PASSED_SCENARIOS + 1))
    echo "✅ Scenario 2 PASSED"
else
    SCENARIO_RESULTS[1]="FAIL"
    echo "❌ Scenario 2 FAILED"
fi
scenario_end_time=$(date +%s.%N)
SCENARIO_DURATIONS[1]=$(echo "$scenario_end_time - $scenario_start_time" | bc -l)
SCENARIO_NAMES[1]="Concurrency Performance Test"

# Scenario 3: Multi-Model Test
echo ""
scenario_start_time=$(date +%s.%N)
echo "🔄 Running Scenario 3/4: Multi-Model & Synthesis Test..."
if "$SCRIPT_DIR/scenarios/3_multi_model_test.sh"; then
    SCENARIO_RESULTS[2]="PASS"
    PASSED_SCENARIOS=$((PASSED_SCENARIOS + 1))
    echo "✅ Scenario 3 PASSED"
else
    SCENARIO_RESULTS[2]="FAIL"
    echo "❌ Scenario 3 FAILED"
fi
scenario_end_time=$(date +%s.%N)
SCENARIO_DURATIONS[2]=$(echo "$scenario_end_time - $scenario_start_time" | bc -l)
SCENARIO_NAMES[2]="Multi-Model & Synthesis Test"

# Scenario 4: Stress Test
echo ""
scenario_start_time=$(date +%s.%N)
echo "🔄 Running Scenario 4/4: Rate Limiting & Error Handling Stress Test..."
if "$SCRIPT_DIR/scenarios/4_stress_test.sh"; then
    SCENARIO_RESULTS[3]="PASS"
    PASSED_SCENARIOS=$((PASSED_SCENARIOS + 1))
    echo "✅ Scenario 4 PASSED"
else
    SCENARIO_RESULTS[3]="FAIL"
    echo "❌ Scenario 4 FAILED"
fi
scenario_end_time=$(date +%s.%N)
SCENARIO_DURATIONS[3]=$(echo "$scenario_end_time - $scenario_start_time" | bc -l)
SCENARIO_NAMES[3]="Rate Limiting & Error Handling Stress Test"

# Calculate overall duration
overall_end_time=$(date +%s.%N)
overall_duration=$(echo "$overall_end_time - $overall_start_time" | bc -l)

# Format duration helper function
format_duration() {
    local duration=$1
    if (( $(echo "$duration < 1" | bc -l) )); then
        printf "%.3fs" "$duration"
    elif (( $(echo "$duration < 60" | bc -l) )); then
        printf "%.1fs" "$duration"
    else
        local minutes=$(echo "$duration / 60" | bc)
        local seconds=$(echo "$duration % 60" | bc)
        printf "%dm%.1fs" "$minutes" "$seconds"
    fi
}

echo ""
echo ""
echo "========================================="
echo "  LOAD TEST SUITE SUMMARY"
echo "========================================="
echo "📊 Overall Results:"
echo "   Total scenarios: $TOTAL_SCENARIOS"
echo "   Passed scenarios: $PASSED_SCENARIOS"
echo "   Failed scenarios: $((TOTAL_SCENARIOS - PASSED_SCENARIOS))"
echo "   Success rate: $(echo "scale=1; $PASSED_SCENARIOS * 100 / $TOTAL_SCENARIOS" | bc)%"
echo "   Total execution time: $(format_duration "$overall_duration")"
echo ""

echo "📋 Individual Scenario Results:"
for i in $(seq 0 $((TOTAL_SCENARIOS - 1))); do
    status_icon="❌"
    if [ "${SCENARIO_RESULTS[$i]}" = "PASS" ]; then
        status_icon="✅"
    fi
    printf "   %s Scenario %d: %s (%s)\n" \
        "$status_icon" \
        $((i + 1)) \
        "${SCENARIO_NAMES[$i]}" \
        "$(format_duration "${SCENARIO_DURATIONS[$i]}")"
done

echo ""
echo "📁 Test Output Locations:"
echo "   All test outputs: tmp/load-testing/"
echo "   Reliability test: tmp/load-testing/reliability/"
echo "   Concurrency test: tmp/load-testing/concurrency/"
echo "   Multi-model test: tmp/load-testing/multi-model/"
echo "   Stress test: tmp/load-testing/stress/"

echo ""
echo "🔍 Quick Analysis:"
total_outputs=$(find tmp/load-testing -name "*.md" 2>/dev/null | wc -l)
echo "   Total output files generated: $total_outputs"

if [ "$PASSED_SCENARIOS" -eq "$TOTAL_SCENARIOS" ]; then
    echo "   🎉 Excellent! All scenarios passed. The system shows strong reliability and performance."
elif [ "$PASSED_SCENARIOS" -ge $((TOTAL_SCENARIOS * 3 / 4)) ]; then
    echo "   👍 Good! Most scenarios passed. Minor issues detected but overall system health is strong."
elif [ "$PASSED_SCENARIOS" -ge $((TOTAL_SCENARIOS / 2)) ]; then
    echo "   ⚠️  Moderate! Some scenarios failed. System functional but may need optimization."
else
    echo "   🚨 Concerning! Many scenarios failed. System may have significant reliability issues."
fi

echo ""
echo "📚 Recommendations:"
echo "   • Review individual scenario outputs for detailed insights"
echo "   • Check API key configuration if many tests failed"
echo "   • Monitor system resources during high-concurrency tests"
echo "   • Consider adjusting rate limits based on stress test results"
echo "   • Use synthesis outputs to understand multi-model quality differences"

echo ""
echo "========================================="
echo "  Load Test Suite Completed"
echo "========================================="
echo "Timestamp: $(date)"

# Exit with appropriate code
if [ "$PASSED_SCENARIOS" -eq "$TOTAL_SCENARIOS" ]; then
    echo "🎉 ALL TESTS PASSED"
    exit 0
elif [ "$PASSED_SCENARIOS" -gt 0 ]; then
    echo "⚠️  PARTIAL SUCCESS ($PASSED_SCENARIOS/$TOTAL_SCENARIOS passed)"
    exit 1
else
    echo "❌ ALL TESTS FAILED"
    exit 2
fi
