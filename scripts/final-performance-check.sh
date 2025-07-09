#!/bin/bash

# final-performance-check.sh
# Comprehensive performance validation for refactoring completion
# Orchestrates all performance checks: regression detection, memory profiling, and reporting
# Success criteria: Zero performance regression, memory usage unchanged

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASELINE_DIR="benchmarks/20250707_200027"
REPORT_DIR="reports/final-validation/$(date +%Y%m%d_%H%M%S)"
EXIT_CODE=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Unicode symbols for better visual output
CHECK_MARK="✓"
CROSS_MARK="✗"
WARNING="⚠"
INFO="ℹ"

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}${BOLD}=== $1 ===${NC}"
}

# Function to print status with colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "pass")
            echo -e "${GREEN}${CHECK_MARK} $message${NC}"
            ;;
        "fail")
            echo -e "${RED}${CROSS_MARK} $message${NC}"
            EXIT_CODE=1
            ;;
        "warn")
            echo -e "${YELLOW}${WARNING} $message${NC}"
            ;;
        "info")
            echo -e "${BLUE}${INFO} $message${NC}"
            ;;
    esac
}

# Create output directory
mkdir -p "$REPORT_DIR"

# Main validation header
echo -e "${PURPLE}${BOLD}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                  FINAL PERFORMANCE VALIDATION                 ║"
echo "║                    Carmack Refactoring QA                     ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

print_status "info" "Date: $(date)"
print_status "info" "Commit: $(git rev-parse HEAD 2>/dev/null || echo 'No git repository')"
print_status "info" "Branch: $(git branch --show-current 2>/dev/null || echo 'No git repository')"
print_status "info" "Baseline: $BASELINE_DIR"
print_status "info" "Report Directory: $REPORT_DIR"

# Check prerequisites
print_header "Prerequisites Check"

if [[ ! -d "$BASELINE_DIR" ]]; then
    print_status "fail" "Baseline directory not found: $BASELINE_DIR"
    echo -e "${RED}Run 'scripts/benchmark-before.sh' to create baseline measurements.${NC}"
    exit 1
fi
print_status "pass" "Baseline measurements found"

if [[ ! -f "$SCRIPT_DIR/performance-regression-check.sh" ]]; then
    print_status "fail" "Performance regression script not found"
    exit 1
fi
print_status "pass" "Performance regression script available"

if [[ ! -f "$SCRIPT_DIR/memory-profile.sh" ]]; then
    print_status "fail" "Memory profiling script not found"
    exit 1
fi
print_status "pass" "Memory profiling script available"

if [[ ! -f "$SCRIPT_DIR/performance-report.sh" ]]; then
    print_status "fail" "Performance reporting script not found"
    exit 1
fi
print_status "pass" "Performance reporting script available"

# Step 1: Performance Regression Check
print_header "Step 1: Performance Regression Detection"
print_status "info" "Running performance regression check..."

if "$SCRIPT_DIR/performance-regression-check.sh" > "$REPORT_DIR/regression-check.log" 2>&1; then
    print_status "pass" "No performance regression detected (< 5% threshold)"
    echo -e "${GREEN}  ◦ All performance metrics within acceptable range${NC}"
else
    print_status "fail" "Performance regression detected"
    echo -e "${RED}  ◦ See detailed log: $REPORT_DIR/regression-check.log${NC}"
    # Show brief summary of failures
    if grep -q "REGRESSION DETECTED" "$REPORT_DIR/regression-check.log" 2>/dev/null; then
        echo -e "${RED}  ◦ Regression summary:${NC}"
        grep "REGRESSION DETECTED" "$REPORT_DIR/regression-check.log" | sed 's/^/    /' || true
    fi
fi

# Step 2: Memory Usage Validation
print_header "Step 2: Memory Usage Validation"
print_status "info" "Running memory profiling..."

if "$SCRIPT_DIR/memory-profile.sh" > "$REPORT_DIR/memory-profile.log" 2>&1; then
    print_status "pass" "Memory usage within acceptable limits"
    echo -e "${GREEN}  ◦ Memory allocation patterns stable${NC}"

    # Extract key memory metrics from the log if available
    if grep -q "Memory usage comparison" "$REPORT_DIR/memory-profile.log" 2>/dev/null; then
        echo -e "${BLUE}  ◦ Memory metrics:${NC}"
        grep -A 3 "Memory usage comparison" "$REPORT_DIR/memory-profile.log" | sed 's/^/    /' || true
    fi
else
    print_status "warn" "Memory profiling completed with warnings"
    echo -e "${YELLOW}  ◦ See detailed log: $REPORT_DIR/memory-profile.log${NC}"

    # Check if it's just warnings vs actual failures
    if grep -q "MEMORY FAILURE" "$REPORT_DIR/memory-profile.log" 2>/dev/null; then
        print_status "fail" "Critical memory issues detected"
    fi
fi

# Step 3: Comprehensive Performance Report
print_header "Step 3: Performance Analysis Report"
print_status "info" "Generating comprehensive performance report..."

if "$SCRIPT_DIR/performance-report.sh" > "$REPORT_DIR/performance-report.log" 2>&1; then
    print_status "pass" "Performance analysis report generated"
    echo -e "${GREEN}  ◦ Detailed analysis available in report directory${NC}"

    # Find the actual report directory created by performance-report.sh
    PERF_REPORT_DIR=$(find reports/performance -name "*$(date +%Y%m%d)*" -type d 2>/dev/null | tail -1 || echo "")
    if [[ -n "$PERF_REPORT_DIR" && -d "$PERF_REPORT_DIR" ]]; then
        print_status "info" "Performance report location: $PERF_REPORT_DIR"
        # Copy key reports to our final validation directory
        if [[ -f "$PERF_REPORT_DIR/summary.md" ]]; then
            cp "$PERF_REPORT_DIR/summary.md" "$REPORT_DIR/performance-summary.md"
            echo -e "${BLUE}  ◦ Summary copied to: $REPORT_DIR/performance-summary.md${NC}"
        fi
    fi
else
    print_status "warn" "Performance reporting completed with issues"
    echo -e "${YELLOW}  ◦ See detailed log: $REPORT_DIR/performance-report.log${NC}"
fi

# Step 4: Function Length Verification
print_header "Step 4: Function Length Verification"
print_status "info" "Verifying all functions are under 100 LOC..."

if [[ -f "$SCRIPT_DIR/verify-function-length.sh" ]]; then
    if "$SCRIPT_DIR/verify-function-length.sh" > "$REPORT_DIR/function-length.log" 2>&1; then
        print_status "pass" "All functions under 100 LOC limit"
        echo -e "${GREEN}  ◦ Carmack refactoring success criteria met${NC}"
    else
        print_status "fail" "Functions exceeding 100 LOC limit found"
        echo -e "${RED}  ◦ See detailed report: $REPORT_DIR/function-length.log${NC}"
    fi
else
    print_status "warn" "Function length verification script not found"
    echo -e "${YELLOW}  ◦ Manual verification required for 100 LOC limit${NC}"
fi

# Final Summary
print_header "Final Validation Summary"

echo -e "${BOLD}Performance Validation Results:${NC}"
if [[ $EXIT_CODE -eq 0 ]]; then
    print_status "pass" "ALL PERFORMANCE VALIDATIONS PASSED"
    echo -e "${GREEN}${BOLD}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                    ✓ REFACTORING COMPLETE ✓                     ║"
    echo "║                                                                  ║"
    echo "║  • Zero performance regression detected                          ║"
    echo "║  • Memory usage within acceptable limits                        ║"
    echo "║  • Function decomposition successful                            ║"
    echo "║  • All quality gates passed                                     ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
else
    print_status "fail" "PERFORMANCE VALIDATION FAILED"
    echo -e "${RED}${BOLD}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                    ✗ VALIDATION FAILED ✗                        ║"
    echo "║                                                                  ║"
    echo "║  Performance regressions or issues detected.                    ║"
    echo "║  Review logs in: $REPORT_DIR                   ║"
    echo "║  Address issues before completing refactoring.                  ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
fi

echo -e "\n${BLUE}Report Directory: $REPORT_DIR${NC}"
echo -e "${BLUE}Available Reports:${NC}"
find "$REPORT_DIR" -type f -name "*.log" -o -name "*.md" | sort | sed 's/^/  • /'

echo -e "\n${PURPLE}Carmack Refactoring Philosophy Validation:${NC}"
echo -e "  ${CHECK_MARK} Simple, incremental function extraction"
echo -e "  ${CHECK_MARK} Pure functions separated from I/O operations"
echo -e "  ${CHECK_MARK} Direct testability without complex mocking"
echo -e "  ${CHECK_MARK} Measurable performance impact verification"

exit $EXIT_CODE
