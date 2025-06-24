#!/bin/bash
#
# Common helper functions for thinktank load testing scripts.
#

# Color codes for logging
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Check if bc is available for timing calculations
check_prerequisites() {
    if ! command -v bc &> /dev/null; then
        log_fail "ERROR: 'bc' command is required for timing calculations. Please install it."
        exit 1
    fi
}

# Format duration for human readability
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

# Wrapper function to run thinktank, measure time, and validate basic output.
# Usage: run_thinktank "Test Description" "output_dir" "thinktank_args..."
run_thinktank() {
    local description="$1"
    local output_dir="$2"
    shift 2
    local args=("$@")

    log_info "Running test: $description"
    log_info "  -> Args: ${args[*]}"

    # Ensure a clean state for the run
    rm -rf "$output_dir"
    mkdir -p "$output_dir"

    # Measure execution time
    start_time=$(date +%s.%N)
    output=$("$THINKTANK_BINARY" --output-dir "$output_dir" "${args[@]}" 2>&1)
    exit_code=$?
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc -l)

    # Basic validation
    if [ $exit_code -eq 0 ]; then
        if ls "$output_dir"/*.md 1> /dev/null 2>&1; then
            log_pass "SUCCESS: Test completed in $(format_duration "$duration") with exit code 0. Output files found."
            return 0
        else
            log_fail "FAILURE: Test exited 0 but NO output files were found in '$output_dir'."
            echo "$output" | head -20  # Show first 20 lines of output
            return 1
        fi
    else
        log_fail "FAILURE: Test failed with exit code $exit_code. Duration: $(format_duration "$duration")."
        echo "$output" | head -20  # Show first 20 lines of output
        return 1
    fi
}

# Wrapper function for tests that expect partial success
# Usage: run_thinktank_partial "Test Description" "output_dir" "thinktank_args..."
run_thinktank_partial() {
    local description="$1"
    local output_dir="$2"
    shift 2
    local args=("$@")

    log_info "Running test: $description (partial success expected)"
    log_info "  -> Args: ${args[*]}"

    # Ensure a clean state for the run
    rm -rf "$output_dir"
    mkdir -p "$output_dir"

    # Measure execution time
    start_time=$(date +%s.%N)
    output=$("$THINKTANK_BINARY" --output-dir "$output_dir" "${args[@]}" 2>&1)
    exit_code=$?
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc -l)

    # For partial success tests, we accept both success and partial failure
    if [ $exit_code -eq 0 ] || [ $exit_code -eq 2 ]; then
        # Check if at least some output files were created
        output_count=$(find "$output_dir" -name "*.md" | wc -l)
        if [ "$output_count" -gt 0 ]; then
            log_pass "PARTIAL SUCCESS: Test completed in $(format_duration "$duration") with exit code $exit_code. $output_count output files found."
            return 0
        else
            log_warn "UNEXPECTED: Test exited $exit_code but no output files found."
            echo "$output" | head -20
            return 1
        fi
    else
        log_fail "FAILURE: Test failed with exit code $exit_code. Duration: $(format_duration "$duration")."
        echo "$output" | head -20
        return 1
    fi
}

# Count output files and log summary
count_outputs() {
    local output_dir="$1"
    local output_count=$(find "$output_dir" -name "*.md" 2>/dev/null | wc -l)
    log_info "Output files generated: $output_count"
    if [ "$output_count" -gt 0 ]; then
        find "$output_dir" -name "*.md" -exec basename {} \; | while read -r file; do
            log_info "  -> $file"
        done
    fi
}

# Initialize test environment
init_test_env() {
    log_info "Initializing test environment..."

    # Check prerequisites
    check_prerequisites

    # Create temporary directory for load testing
    mkdir -p tmp/load-testing

    # Log configuration
    log_info "Binary: $THINKTANK_BINARY"
    log_info "Test timeout: ${LOAD_TEST_TIMEOUT:-300}s"
    log_info "Sleep between tests: ${LOAD_TEST_SLEEP:-1}s"
}
