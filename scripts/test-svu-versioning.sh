#!/bin/bash
# Test script for verifying svu version calculation behavior

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Output file for results
RESULTS_FILE="docs/svu-version-test-results.md"

# Initialize results file
init_results() {
    cat > "$RESULTS_FILE" << 'EOF'
# SVU Version Calculation Test Results

This document contains test results for verifying `svu next` version calculation behavior across different commit scenarios.

Generated on: $(date)

## Test Results

| Starting Version | Commit(s) | Expected Version | Actual Version | Status |
|-----------------|-----------|------------------|----------------|--------|
EOF
}

# Create a test repository
create_test_repo() {
    local test_name=$1
    local test_dir="test-repo-$test_name"

    rm -rf "$test_dir"
    mkdir "$test_dir"
    cd "$test_dir"

    git init
    git config user.email "test@example.com"
    git config user.name "Test User"

    # Create initial commit
    echo "initial" > file.txt
    git add file.txt
    git commit -m "chore: initial commit"

    cd ..
}

# Run a test case
run_test() {
    local test_name=$1
    local start_version=$2
    local commits=$3
    local expected_version=$4
    local description=$5

    echo -n "Testing: $description... "

    cd "test-repo-$test_name"

    # Set starting version if provided
    if [ "$start_version" != "none" ]; then
        git tag "$start_version"
    fi

    # Apply commits
    IFS='|' read -ra COMMIT_ARRAY <<< "$commits"
    for commit in "${COMMIT_ARRAY[@]}"; do
        echo "change" >> file.txt
        git add file.txt
        git commit -m "$commit"
    done

    # Get version from svu
    actual_version=$(svu next)

    # Compare with expected
    if [ "$actual_version" = "$expected_version" ]; then
        status="✅ PASS"
        echo -e "${GREEN}PASS${NC}"
    else
        status="❌ FAIL"
        echo -e "${RED}FAIL${NC} (expected: $expected_version, got: $actual_version)"
    fi

    # Write to results file
    echo "| $start_version | $commits | $expected_version | $actual_version | $status |" >> "../$RESULTS_FILE"

    cd ..
}

# Main test execution
main() {
    echo "Starting SVU version calculation tests..."
    echo

    # Initialize results file
    init_results

    # Test 1: No tags, patch commit
    create_test_repo "test1"
    run_test "test1" "none" "fix: resolve bug" "v0.0.1" "No tags, patch commit"

    # Test 2: No tags, minor commit
    create_test_repo "test2"
    run_test "test2" "none" "feat: add new feature" "v0.1.0" "No tags, minor commit"

    # Test 3: No tags, breaking change
    create_test_repo "test3"
    run_test "test3" "none" "feat!: breaking change" "v1.0.0" "No tags, breaking change"

    # Test 4: Existing version, patch commit
    create_test_repo "test4"
    run_test "test4" "v1.2.3" "fix: bug fix" "v1.2.4" "Existing version, patch commit"

    # Test 5: Existing version, minor commit
    create_test_repo "test5"
    run_test "test5" "v1.2.3" "feat: new feature" "v1.3.0" "Existing version, minor commit"

    # Test 6: Existing version, breaking change
    create_test_repo "test6"
    run_test "test6" "v1.2.3" "fix!: breaking fix" "v2.0.0" "Existing version, breaking change"

    # Test 7: Multiple commits, highest precedence wins
    create_test_repo "test7"
    run_test "test7" "v1.0.0" "fix: bug fix|feat: new feature" "v1.1.0" "Multiple commits (patch + minor)"

    # Test 8: Multiple commits with breaking change
    create_test_repo "test8"
    run_test "test8" "v1.0.0" "fix: bug fix|feat: new feature|refactor!: breaking refactor" "v2.0.0" "Multiple commits with breaking"

    # Test 9: Non-version-changing commit
    create_test_repo "test9"
    run_test "test9" "v1.0.0" "chore: update deps" "v1.0.1" "Non-version-changing commit"

    # Test 10: Scoped commits
    create_test_repo "test10"
    run_test "test10" "v1.0.0" "feat(api): add endpoint" "v1.1.0" "Scoped feature commit"

    # Test 11: Pre-release version
    create_test_repo "test11"
    run_test "test11" "v1.0.0-alpha.1" "feat: new feature" "v1.0.0-alpha.2" "Pre-release version"

    # Test 12: Multiple non-version commits
    create_test_repo "test12"
    run_test "test12" "v1.0.0" "docs: update readme|style: format code|chore: cleanup" "v1.0.1" "Multiple non-version commits"

    # Cleanup
    rm -rf test-repo-*

    echo
    echo "Tests complete! Results written to $RESULTS_FILE"
}

# Run tests
main
