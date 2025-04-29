#!/bin/bash
# Test script for analyze-claude-warnings.sh
# This script tests the basic functionality of the warning analysis tool

set -e

# Setup test environment
TEST_DIR="/tmp/claude-warnings-test"
TEST_LOG="$TEST_DIR/.claude-warnings.log"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ANALYZE_SCRIPT="$SCRIPT_DIR/analyze-claude-warnings.sh"

# Create test directory and log file
mkdir -p "$TEST_DIR"

# Generate test log entries
cat > "$TEST_LOG" << EOF
----------------------------------------
DATE: 2024-04-29 10:00:00 UTC
STATUS: WARN
COMMIT: test-commit-1
BRANCH: feature/test
COMMIT_MSG: test: first test commit
CORRELATION_ID: claude_test_123456
----------------------------------------
WARN
- Test warning 1
- Test warning 2

----------------------------------------
DATE: 2024-04-29 11:00:00 UTC
STATUS: FAIL
COMMIT: test-commit-2
BRANCH: feature/test
COMMIT_MSG: feat: second test commit
CORRELATION_ID: claude_test_789012
----------------------------------------
FAIL
- Test failure 1
- Test failure 2
- Test failure 3

EOF

# Run tests with more robust assertions
run_test() {
    local test_name="$1"
    local command="$2"
    local expected="$3"
    local output
    local failed=0
    local total_assertions=0
    local passed_assertions=0

    echo "Running test: $test_name"
    output=$(cd "$TEST_DIR" && LOG_FILE="$TEST_LOG" bash -c "$command")

    # Split expected patterns by comma for multiple assertions
    IFS=',' read -ra ASSERTIONS <<< "$expected"
    for assertion in "${ASSERTIONS[@]}"; do
        total_assertions=$((total_assertions + 1))
        if echo "$output" | grep -q "$assertion"; then
            passed_assertions=$((passed_assertions + 1))
            echo "  ✓ Assertion passed: '$assertion' found in output"
        else
            failed=1
            echo "  ✗ Assertion failed: '$assertion' not found in output"
        fi
    done

    # Check for unexpected content that shouldn't be there
    if [ $# -eq 4 ] && [ -n "$4" ]; then
        total_assertions=$((total_assertions + 1))
        if echo "$output" | grep -q "$4"; then
            failed=1
            echo "  ✗ Negative assertion failed: '$4' was found in output but shouldn't be"
        else
            passed_assertions=$((passed_assertions + 1))
            echo "  ✓ Negative assertion passed: '$4' correctly not found in output"
        fi
    fi

    # Print test summary
    if [ $failed -eq 0 ]; then
        echo "✅ Test passed: $passed_assertions/$total_assertions assertions successful"
        return 0
    else
        echo "❌ Test failed: $passed_assertions/$total_assertions assertions successful"
        echo "Output was:"
        echo "$output"
        return 1
    fi
}

# Test list functionality
run_test "List entries" \
    "$ANALYZE_SCRIPT --list" \
    "test-commit-1,test-commit-2,WARN,FAIL,CORRELATION_ID" \
    "NOT_A_VALID_COMMIT"

# Test commit details
run_test "Commit details" \
    "$ANALYZE_SCRIPT -c test-commit-1" \
    "Test warning 1,Test warning 2,BRANCH: feature/test,CORRELATION_ID: claude_test_123456" \
    "Test failure"

# Test filter by status
run_test "Filter by status" \
    "$ANALYZE_SCRIPT -s FAIL" \
    "COMMIT: test-commit-2,Test failure 1,Test failure 2,Test failure 3,CORRELATION_ID: claude_test_789012" \
    "Test warning"

# Test filter by branch
run_test "Filter by branch" \
    "$ANALYZE_SCRIPT -b feature/test" \
    "COMMIT: test-commit-1,COMMIT: test-commit-2,WARN,FAIL,CORRELATION_ID" \
    "NOT_A_VALID_BRANCH"

# Test summary
run_test "Summary" \
    "$ANALYZE_SCRIPT --summary" \
    "Total warnings: 1,Total failures: 1,feature/test"

# Clean up
rm -rf "$TEST_DIR"

echo "All tests completed!"
