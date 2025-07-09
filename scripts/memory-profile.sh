#!/bin/bash

# memory-profile.sh
# Tracks memory usage before/after each refactoring step
# Compares current memory allocations against baseline measurements

set -euo pipefail

# Configuration
BASELINE_DIR="benchmarks/20250707_200027"
CURRENT_DIR="benchmarks/$(date +%Y%m%d_%H%M%S)_memory"
MEMORY_THRESHOLD_INCREASE=10  # Allow 10% increase in memory usage
PROFILE_TIME="5s"             # Profile duration for stability

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Memory Allocation Profiling ===${NC}"
echo "Date: $(date)"
echo "Commit: $(git rev-parse HEAD)"
echo "Branch: $(git branch --show-current)"
echo "Baseline: $BASELINE_DIR"
echo "Current: $CURRENT_DIR"
echo ""

# Create output directory
mkdir -p "$CURRENT_DIR"

# Function to run memory benchmark for a package
run_memory_benchmark() {
    local package_name="$1"
    local test_path="$2"
    local output_file="$CURRENT_DIR/${package_name}_memory.txt"
    local profile_file="$CURRENT_DIR/${package_name}.memprofile"

    echo -e "${BLUE}Profiling memory for: $package_name${NC}"
    echo "Test path: $test_path"

    # Run benchmark with memory profiling
    if go test -bench=. -benchmem -memprofile="$profile_file" \
        -benchtime="$PROFILE_TIME" \
        "$test_path" > "$output_file" 2>&1; then
        echo -e "${GREEN}✓${NC} Memory profile saved to: $output_file"
    else
        echo -e "${YELLOW}⚠${NC} Some benchmarks failed, but memory data captured"
    fi

    # Generate memory analysis if profile exists
    if [ -f "$profile_file" ]; then
        local analysis_file="$CURRENT_DIR/${package_name}_memory_analysis.txt"
        go tool pprof -text -cum "$profile_file" > "$analysis_file" 2>&1
        echo -e "${GREEN}✓${NC} Memory analysis saved to: $analysis_file"
    fi
    echo ""
}

# Function to extract memory allocation data from benchmark output
extract_memory_metrics() {
    local benchmark_file="$1"
    local metric_type="$2"  # "B/op" or "allocs/op"

    if [ ! -f "$benchmark_file" ]; then
        echo "0"
        return
    fi

    # Extract memory metrics, handling both integer and decimal values
    local metrics=$(grep -E "Benchmark.*-[0-9]+" "$benchmark_file" | \
        awk -v metric="$metric_type" '
        {
            for (i = 1; i <= NF; i++) {
                if ($i == metric && i > 1) {
                    # Remove any non-numeric characters except decimal point
                    gsub(/[^0-9.]/, "", $(i-1))
                    print $(i-1)
                }
            }
        }' | \
        awk '{sum += $1; count++} END {if (count > 0) printf "%.2f", sum/count; else print "0"}')

    echo "${metrics:-0}"
}

# Function to compare memory usage against baseline
compare_memory_metrics() {
    local package_name="$1"
    local current_file="$CURRENT_DIR/${package_name}_memory.txt"
    local baseline_file="$BASELINE_DIR/${package_name}.txt"

    echo -e "${BLUE}Memory Comparison for $package_name:${NC}"

    if [ ! -f "$baseline_file" ]; then
        echo -e "${YELLOW}⚠ No baseline data for $package_name${NC}"
        return 0
    fi

    # Extract current memory metrics
    local current_bytes=$(extract_memory_metrics "$current_file" "B/op")
    local current_allocs=$(extract_memory_metrics "$current_file" "allocs/op")

    # Extract baseline memory metrics
    local baseline_bytes=$(extract_memory_metrics "$baseline_file" "B/op")
    local baseline_allocs=$(extract_memory_metrics "$baseline_file" "allocs/op")

    echo "  Memory Usage (B/op):"
    echo "    Baseline: $baseline_bytes B/op"
    echo "    Current:  $current_bytes B/op"

    echo "  Allocations (allocs/op):"
    echo "    Baseline: $baseline_allocs allocs/op"
    echo "    Current:  $current_allocs allocs/op"

    # Calculate percentage changes
    local bytes_change=""
    local allocs_change=""
    local regression_detected=false

    if [ "$(echo "$baseline_bytes > 0" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
        bytes_change=$(echo "scale=2; (($current_bytes - $baseline_bytes) / $baseline_bytes) * 100" | bc -l 2>/dev/null || echo "0")
        if [ "$(echo "$bytes_change > $MEMORY_THRESHOLD_INCREASE" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
            echo -e "    ${RED}✗ Memory regression: +${bytes_change}% (threshold: +${MEMORY_THRESHOLD_INCREASE}%)${NC}"
            regression_detected=true
        else
            echo -e "    ${GREEN}✓ Memory change: ${bytes_change:+${bytes_change}%}${NC}"
        fi
    fi

    if [ "$(echo "$baseline_allocs > 0" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
        allocs_change=$(echo "scale=2; (($current_allocs - $baseline_allocs) / $baseline_allocs) * 100" | bc -l 2>/dev/null || echo "0")
        if [ "$(echo "$allocs_change > $MEMORY_THRESHOLD_INCREASE" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
            echo -e "    ${RED}✗ Allocation regression: +${allocs_change}% (threshold: +${MEMORY_THRESHOLD_INCREASE}%)${NC}"
            regression_detected=true
        else
            echo -e "    ${GREEN}✓ Allocation change: ${allocs_change:+${allocs_change}%}${NC}"
        fi
    fi

    echo ""
    return $([ "$regression_detected" = true ] && echo 1 || echo 0)
}

# Function to analyze memory hotspots
analyze_memory_hotspots() {
    local package_name="$1"
    local profile_file="$CURRENT_DIR/${package_name}.memprofile"
    local hotspot_file="$CURRENT_DIR/${package_name}_hotspots.txt"

    if [ ! -f "$profile_file" ]; then
        return
    fi

    echo -e "${BLUE}Analyzing memory hotspots for $package_name...${NC}"

    # Generate top memory allocation functions
    go tool pprof -text -nodecount=10 "$profile_file" > "$hotspot_file" 2>&1

    # Extract top 5 functions
    echo "  Top memory allocating functions:"
    grep -E "^[[:space:]]*[0-9]" "$hotspot_file" | head -5 | while read line; do
        echo "    $line"
    done
    echo ""
}

# Function to generate comprehensive memory report
generate_memory_report() {
    local report_file="$CURRENT_DIR/memory_analysis_report.md"

    cat > "$report_file" << EOF
# Memory Allocation Analysis Report

Generated: $(date)
Commit: $(git rev-parse HEAD)
Branch: $(git branch --show-current)
Baseline: $BASELINE_DIR

## Summary

This report tracks memory allocation changes across refactored packages.

## Memory Metrics Comparison

EOF

    # Add comparison data for each package
    for package in "main_function" "execute_function" "console_writer" "gather_project_context"; do
        local current_file="$CURRENT_DIR/${package}_memory.txt"
        local baseline_file="$BASELINE_DIR/${package}.txt"

        if [ -f "$current_file" ] && [ -f "$baseline_file" ]; then
            echo "### $package" >> "$report_file"
            echo "" >> "$report_file"

            local current_bytes=$(extract_memory_metrics "$current_file" "B/op")
            local current_allocs=$(extract_memory_metrics "$current_file" "allocs/op")
            local baseline_bytes=$(extract_memory_metrics "$baseline_file" "B/op")
            local baseline_allocs=$(extract_memory_metrics "$baseline_file" "allocs/op")

            echo "| Metric | Baseline | Current | Change |" >> "$report_file"
            echo "|--------|----------|---------|--------|" >> "$report_file"
            echo "| Memory (B/op) | $baseline_bytes | $current_bytes | $([ "$(echo "$baseline_bytes > 0" | bc -l 2>/dev/null || echo "0")" = "1" ] && echo "scale=2; (($current_bytes - $baseline_bytes) / $baseline_bytes) * 100" | bc -l 2>/dev/null || echo "N/A")% |" >> "$report_file"
            echo "| Allocations (allocs/op) | $baseline_allocs | $current_allocs | $([ "$(echo "$baseline_allocs > 0" | bc -l 2>/dev/null || echo "0")" = "1" ] && echo "scale=2; (($current_allocs - $baseline_allocs) / $baseline_allocs) * 100" | bc -l 2>/dev/null || echo "N/A")% |" >> "$report_file"
            echo "" >> "$report_file"
        fi
    done

    cat >> "$report_file" << EOF

## Memory Hotspots

The following functions show the highest memory allocations:

EOF

    # Add hotspot data for each package
    for package in "main_function" "execute_function" "console_writer" "gather_project_context"; do
        local hotspot_file="$CURRENT_DIR/${package}_hotspots.txt"
        if [ -f "$hotspot_file" ]; then
            echo "### $package Top Allocators" >> "$report_file"
            echo "" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            grep -E "^[[:space:]]*[0-9]" "$hotspot_file" | head -5 >> "$report_file" 2>/dev/null || echo "No allocation data available" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            echo "" >> "$report_file"
        fi
    done

    cat >> "$report_file" << EOF

## Recommendations

1. **Memory Efficiency**: Monitor functions with increasing allocation patterns
2. **Performance**: Focus optimization on high-allocation functions
3. **Regression Tracking**: Investigate any memory increases >$MEMORY_THRESHOLD_INCREASE%

## Files Generated

- Memory benchmarks: \`${CURRENT_DIR}/*_memory.txt\`
- Memory profiles: \`${CURRENT_DIR}/*.memprofile\`
- Memory analysis: \`${CURRENT_DIR}/*_memory_analysis.txt\`
- Hotspot analysis: \`${CURRENT_DIR}/*_hotspots.txt\`

EOF

    echo -e "${GREEN}✓ Comprehensive report saved to: $report_file${NC}"
}

# Main execution
echo -e "${BLUE}1. Running memory benchmarks...${NC}"

# Run memory profiling for each target package
run_memory_benchmark "main_function" "./internal/cli"
run_memory_benchmark "execute_function" "./internal/thinktank"
run_memory_benchmark "console_writer" "./internal/logutil"
run_memory_benchmark "gather_project_context" "./internal/fileutil"

echo -e "${BLUE}2. Comparing against baseline...${NC}"

# Compare memory metrics
regression_count=0
for package in "main_function" "execute_function" "console_writer" "gather_project_context"; do
    if ! compare_memory_metrics "$package"; then
        ((regression_count++))
    fi
done

echo -e "${BLUE}3. Analyzing memory hotspots...${NC}"

# Analyze memory hotspots
for package in "main_function" "execute_function" "console_writer" "gather_project_context"; do
    analyze_memory_hotspots "$package"
done

echo -e "${BLUE}4. Generating comprehensive report...${NC}"
generate_memory_report

echo -e "${BLUE}=== Memory Profiling Complete ===${NC}"
echo "Results saved to: $CURRENT_DIR"

if [ $regression_count -gt 0 ]; then
    echo -e "${RED}❌ Memory regression detected in $regression_count package(s)${NC}"
    echo "Review the analysis and consider optimizing memory allocations."
    exit 1
else
    echo -e "${GREEN}✅ No significant memory regressions detected${NC}"
    echo "Memory usage is within acceptable thresholds."
fi

echo ""
echo "Next steps:"
echo "1. Review memory analysis in $CURRENT_DIR"
echo "2. Investigate any functions with high memory allocations"
echo "3. Use memory profiles for optimization: go tool pprof <profile>"
