#!/bin/bash

# performance-regression-check.sh
# Compares current performance against baseline benchmarks
# Fails if performance regresses more than 5%

set -euo pipefail

# Configuration
BASELINE_DIR="benchmarks/20250707_200027"
REGRESSION_THRESHOLD=5  # 5% regression threshold
TEMP_DIR=$(mktemp -d)
CURRENT_RESULTS_DIR="$TEMP_DIR/current"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "=== Performance Regression Check ==="
echo "Baseline: $BASELINE_DIR"
echo "Regression threshold: ${REGRESSION_THRESHOLD}%"
echo "Temp directory: $TEMP_DIR"
echo ""

# Check if baseline exists
if [[ ! -d "$BASELINE_DIR" ]]; then
    echo -e "${RED}ERROR: Baseline directory not found: $BASELINE_DIR${NC}"
    echo "Run ./scripts/benchmark-before.sh first to create baseline"
    exit 1
fi

# Create current results directory
mkdir -p "$CURRENT_RESULTS_DIR"

# Function to run current benchmarks
run_current_benchmarks() {
    echo "=== Running Current Benchmarks ==="

    # Run the same benchmarks as baseline
    echo "1. Benchmarking Main() function..."
    go test -bench=. -benchmem -benchtime=10s ./internal/cli > "$CURRENT_RESULTS_DIR/main_function.txt" 2>&1 || echo "No benchmarks found for cli"

    echo "2. Benchmarking Execute() function..."
    go test -bench=. -benchmem -benchtime=10s ./internal/thinktank > "$CURRENT_RESULTS_DIR/execute_function.txt" 2>&1 || echo "No benchmarks found for thinktank"

    echo "3. Benchmarking console_writer.go functions..."
    go test -bench=. -benchmem -benchtime=10s ./internal/logutil > "$CURRENT_RESULTS_DIR/console_writer.txt" 2>&1 || echo "No benchmarks found for logutil"

    echo "4. Benchmarking GatherProjectContextWithContext() function..."
    go test -bench=. -benchmem -benchtime=10s ./internal/fileutil > "$CURRENT_RESULTS_DIR/gather_project_context.txt" 2>&1 || echo "No benchmarks found for fileutil"

    echo "Current benchmarks completed."
    echo ""
}

# Function to parse benchmark results
parse_benchmark_result() {
    local file="$1"
    local benchmark_name="$2"

    if [[ ! -f "$file" ]]; then
        echo "0 0 0"  # default values if file doesn't exist
        return
    fi

    # Extract ns/op, B/op, allocs/op from benchmark line
    # Format: BenchmarkName-CPUs    iterations    ns/op    B/op    allocs/op
    local line=$(grep "^$benchmark_name" "$file" | head -1)

    if [[ -z "$line" ]]; then
        echo "0 0 0"  # default values if benchmark not found
        return
    fi

    # Extract numeric values, removing units
    # Field 3 is ns/op, Field 4 is B/op, Field 5 is allocs/op
    local ns_op=$(echo "$line" | awk '{print $3}' | tr -d '\t' | sed 's/ns\/op//')
    local b_op=$(echo "$line" | awk '{print $4}' | tr -d '\t' | sed 's/B\/op//')
    local allocs_op=$(echo "$line" | awk '{print $5}' | tr -d '\t' | sed 's/allocs\/op//')

    echo "$ns_op $b_op $allocs_op"
}

# Function to calculate regression percentage
calculate_regression() {
    local baseline="$1"
    local current="$2"

    # Clean up values (remove whitespace, ensure they're numeric)
    baseline=$(echo "$baseline" | xargs)
    current=$(echo "$current" | xargs)

    # Check if values are empty or zero
    if [[ -z "$baseline" ]] || [[ -z "$current" ]] || [[ "$baseline" == "0" ]] || [[ "$current" == "0" ]]; then
        echo "0"
        return
    fi

    # Validate that values are numeric
    if ! [[ "$baseline" =~ ^[0-9]+\.?[0-9]*$ ]] || ! [[ "$current" =~ ^[0-9]+\.?[0-9]*$ ]]; then
        echo "0"
        return
    fi

    # Calculate percentage change: ((current - baseline) / baseline) * 100
    local regression=$(echo "scale=2; (($current - $baseline) / $baseline) * 100" | bc -l)
    echo "$regression"
}

# Function to compare benchmark results
compare_benchmarks() {
    local baseline_file="$1"
    local current_file="$2"
    local test_name="$3"

    echo "=== Comparing $test_name ==="

    # Key benchmarks to check
    local benchmarks=(
        "BenchmarkEstimateTokenCount/Small-11"
        "BenchmarkEstimateTokenCount/Medium-11"
        "BenchmarkEstimateTokenCount/Large-11"
        "BenchmarkShouldProcess/Simple_Path_No_Filters-11"
        "BenchmarkShouldProcess/With_Include_Filters-11"
        "BenchmarkShouldProcess/With_Exclude_Filters-11"
        "BenchmarkShouldProcess/With_All_Filters-11"
        "BenchmarkIsBinaryFile/Small_Text-11"
    )

    local has_regression=false

    for benchmark in "${benchmarks[@]}"; do
        # Parse baseline results
        local baseline_result=$(parse_benchmark_result "$baseline_file" "$benchmark")
        local baseline_ns=$(echo "$baseline_result" | awk '{print $1}')
        local baseline_bytes=$(echo "$baseline_result" | awk '{print $2}')
        local baseline_allocs=$(echo "$baseline_result" | awk '{print $3}')

        # Parse current results
        local current_result=$(parse_benchmark_result "$current_file" "$benchmark")
        local current_ns=$(echo "$current_result" | awk '{print $1}')
        local current_bytes=$(echo "$current_result" | awk '{print $2}')
        local current_allocs=$(echo "$current_result" | awk '{print $3}')

        # Skip if benchmark not found in either file
        if [[ "$baseline_ns" == "0" ]] && [[ "$current_ns" == "0" ]]; then
            continue
        fi

        # Calculate regressions
        local ns_regression=$(calculate_regression "$baseline_ns" "$current_ns")
        local bytes_regression=$(calculate_regression "$baseline_bytes" "$current_bytes")
        local allocs_regression=$(calculate_regression "$baseline_allocs" "$current_allocs")

        # Check for regressions
        local ns_failed=$(echo "$ns_regression > $REGRESSION_THRESHOLD" | bc -l)
        local bytes_failed=$(echo "$bytes_regression > $REGRESSION_THRESHOLD" | bc -l)
        local allocs_failed=$(echo "$allocs_regression > $REGRESSION_THRESHOLD" | bc -l)

        # Report results
        echo "  $benchmark:"

        if [[ "$ns_failed" == "1" ]]; then
            echo -e "    ${RED}✗ ns/op: $baseline_ns → $current_ns (${ns_regression}% regression)${NC}"
            has_regression=true
        elif [[ "$baseline_ns" != "0" ]] && [[ "$current_ns" != "0" ]]; then
            echo -e "    ${GREEN}✓ ns/op: $baseline_ns → $current_ns (${ns_regression}% change)${NC}"
        elif [[ "$baseline_ns" != "0" ]] || [[ "$current_ns" != "0" ]]; then
            echo -e "    ${YELLOW}- ns/op: $baseline_ns → $current_ns (one missing)${NC}"
        fi

        if [[ "$bytes_failed" == "1" ]]; then
            echo -e "    ${RED}✗ B/op: $baseline_bytes → $current_bytes (${bytes_regression}% regression)${NC}"
            has_regression=true
        elif [[ "$baseline_bytes" != "0" ]] && [[ "$current_bytes" != "0" ]]; then
            echo -e "    ${GREEN}✓ B/op: $baseline_bytes → $current_bytes (${bytes_regression}% change)${NC}"
        elif [[ "$baseline_bytes" != "0" ]] || [[ "$current_bytes" != "0" ]]; then
            echo -e "    ${YELLOW}- B/op: $baseline_bytes → $current_bytes (one missing)${NC}"
        fi

        if [[ "$allocs_failed" == "1" ]]; then
            echo -e "    ${RED}✗ allocs/op: $baseline_allocs → $current_allocs (${allocs_regression}% regression)${NC}"
            has_regression=true
        elif [[ "$baseline_allocs" != "0" ]] && [[ "$current_allocs" != "0" ]]; then
            echo -e "    ${GREEN}✓ allocs/op: $baseline_allocs → $current_allocs (${allocs_regression}% change)${NC}"
        elif [[ "$baseline_allocs" != "0" ]] || [[ "$current_allocs" != "0" ]]; then
            echo -e "    ${YELLOW}- allocs/op: $baseline_allocs → $current_allocs (one missing)${NC}"
        fi

        echo ""
    done

    if [[ "$has_regression" == "true" ]]; then
        return 1
    else
        return 0
    fi
}

# Main execution
main() {
    # Run current benchmarks
    run_current_benchmarks

    # Compare results
    local overall_regression=false

    # Check each benchmark file
    local test_files=(
        "gather_project_context.txt"
        "console_writer.txt"
        "main_function.txt"
        "execute_function.txt"
    )

    for test_file in "${test_files[@]}"; do
        local baseline_file="$BASELINE_DIR/$test_file"
        local current_file="$CURRENT_RESULTS_DIR/$test_file"

        if [[ -f "$baseline_file" ]]; then
            if ! compare_benchmarks "$baseline_file" "$current_file" "$test_file"; then
                overall_regression=true
            fi
        fi
    done

    # Final report
    echo "=== Performance Regression Check Results ==="
    if [[ "$overall_regression" == "true" ]]; then
        echo -e "${RED}FAILED: Performance regression detected (>${REGRESSION_THRESHOLD}%)${NC}"
        echo "Review the results above for specific regressions."
        echo ""
        echo "To update baseline if this is expected:"
        echo "  rm -rf $BASELINE_DIR"
        echo "  ./scripts/benchmark-before.sh"
        exit 1
    else
        echo -e "${GREEN}PASSED: No performance regression detected${NC}"
        echo "All benchmarks are within ${REGRESSION_THRESHOLD}% of baseline performance."
        exit 0
    fi
}

# Check dependencies
if ! command -v bc &> /dev/null; then
    echo -e "${RED}ERROR: bc calculator not found. Please install bc.${NC}"
    exit 1
fi

# Run main function
main "$@"
