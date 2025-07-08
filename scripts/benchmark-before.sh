#!/bin/bash

# benchmark-before.sh
# Establishes performance baselines for refactoring target functions
# Run this before starting the refactoring process

set -euo pipefail

echo "=== Performance Baseline Measurement ==="
echo "Date: $(date)"
echo "Commit: $(git rev-parse HEAD)"
echo "Branch: $(git branch --show-current)"
echo ""

# Create output directory for benchmarks
BENCHMARK_DIR="benchmarks/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BENCHMARK_DIR"

# Function to run benchmark and save results
run_benchmark() {
    local test_name="$1"
    local test_path="$2"
    local output_file="$BENCHMARK_DIR/$test_name.txt"

    echo "Running benchmark: $test_name"
    echo "Test path: $test_path"

    # Run benchmark with memory profiling
    go test -bench=. -benchmem -memprofile="$BENCHMARK_DIR/$test_name.memprofile" \
        -cpuprofile="$BENCHMARK_DIR/$test_name.cpuprofile" \
        -benchtime=10s \
        "$test_path" > "$output_file" 2>&1

    echo "Results saved to: $output_file"
    echo ""
}

# Function to run CPU profile for specific function
run_cpu_profile() {
    local function_name="$1"
    local test_path="$2"
    local output_file="$BENCHMARK_DIR/${function_name}_cpu_profile.txt"

    echo "Running CPU profile for: $function_name"

    # Create a temporary profile and analyze it
    go test -cpuprofile="$BENCHMARK_DIR/${function_name}.cpuprofile" \
        -bench="Benchmark.*${function_name}" \
        -benchtime=10s \
        "$test_path" > /dev/null 2>&1 || echo "No specific benchmark found for $function_name"

    if [ -f "$BENCHMARK_DIR/${function_name}.cpuprofile" ]; then
        go tool pprof -text -cum "$BENCHMARK_DIR/${function_name}.cpuprofile" > "$output_file" 2>&1
        echo "CPU profile saved to: $output_file"
    fi
    echo ""
}

# Benchmark target functions
echo "1. Benchmarking Main() function..."
run_benchmark "main_function" "./internal/cli"

echo "2. Benchmarking Execute() function..."
run_benchmark "execute_function" "./internal/thinktank"

echo "3. Benchmarking console_writer.go functions..."
run_benchmark "console_writer" "./internal/logutil"

echo "4. Benchmarking GatherProjectContextWithContext() function..."
run_benchmark "gather_project_context" "./internal/fileutil"

# Generate CPU profiles for key functions
echo "5. Generating CPU profiles..."
run_cpu_profile "Main" "./internal/cli"
run_cpu_profile "Execute" "./internal/thinktank"
run_cpu_profile "GatherProjectContext" "./internal/fileutil"

# Generate memory profiles for key functions
echo "6. Generating memory profiles..."
echo "Running memory profile for main components..."
go test -bench=. -benchmem -memprofile="$BENCHMARK_DIR/memory_baseline.memprofile" \
    -benchtime=10s \
    ./internal/cli ./internal/thinktank ./internal/fileutil ./internal/logutil \
    > "$BENCHMARK_DIR/memory_baseline.txt" 2>&1

if [ -f "$BENCHMARK_DIR/memory_baseline.memprofile" ]; then
    go tool pprof -text -cum "$BENCHMARK_DIR/memory_baseline.memprofile" > "$BENCHMARK_DIR/memory_analysis.txt" 2>&1
    echo "Memory analysis saved to: $BENCHMARK_DIR/memory_analysis.txt"
fi

# Create summary report
echo "7. Creating summary report..."
SUMMARY_FILE="$BENCHMARK_DIR/baseline_summary.md"
cat > "$SUMMARY_FILE" << EOF
# Performance Baseline Summary

Generated: $(date)
Commit: $(git rev-parse HEAD)
Branch: $(git branch --show-current)

## Test Results

### Main() Function
- Test path: ./internal/cli
- Results: [main_function.txt](main_function.txt)

### Execute() Function
- Test path: ./internal/thinktank
- Results: [execute_function.txt](execute_function.txt)

### Console Writer Functions
- Test path: ./internal/logutil
- Results: [console_writer.txt](console_writer.txt)

### GatherProjectContextWithContext() Function
- Test path: ./internal/fileutil
- Results: [gather_project_context.txt](gather_project_context.txt)

## Profile Analysis

### CPU Profiles
- Main function: [Main_cpu_profile.txt](Main_cpu_profile.txt)
- Execute function: [Execute_cpu_profile.txt](Execute_cpu_profile.txt)
- GatherProjectContext: [GatherProjectContext_cpu_profile.txt](GatherProjectContext_cpu_profile.txt)

### Memory Profiles
- Overall memory baseline: [memory_baseline.txt](memory_baseline.txt)
- Memory analysis: [memory_analysis.txt](memory_analysis.txt)

## Usage

Compare future benchmark results against these baseline measurements to detect performance regressions during refactoring.

Run comparison with:
\`\`\`bash
./scripts/benchmark-compare.sh $BENCHMARK_DIR
\`\`\`
EOF

echo "=== Benchmark Complete ==="
echo "Baseline results saved to: $BENCHMARK_DIR"
echo "Summary report: $SUMMARY_FILE"
echo ""
echo "Next steps:"
echo "1. Review the baseline measurements in $BENCHMARK_DIR"
echo "2. Use these baselines to detect performance regressions during refactoring"
echo "3. Run './scripts/benchmark-compare.sh $BENCHMARK_DIR' after refactoring changes"
