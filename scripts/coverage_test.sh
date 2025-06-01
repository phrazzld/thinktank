#!/bin/bash
# Test script for coverage threshold enforcement

set -e

# Test configuration
TEST_DIR="$(mktemp -d)"
TEST_COVERAGE_FILE="$TEST_DIR/test_coverage.out"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Cleanup function
cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Create test coverage file with known coverage percentages
create_test_coverage() {
    local total_coverage=$1
    
    cat > "$TEST_COVERAGE_FILE" << EOF
mode: atomic
github.com/phrazzld/thinktank/cmd/thinktank/main.go:10.1,15.2 5 1
github.com/phrazzld/thinktank/cmd/thinktank/main.go:16.1,20.2 3 1
github.com/phrazzld/thinktank/cmd/thinktank/cli.go:20.1,25.2 2 0
EOF

    # Use existing package paths to avoid module issues
    if [ "$total_coverage" = "80.0" ]; then
        # Add more covered lines for 80%
        for i in $(seq 1 8); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/file$i.go:$i.1,$((i+1)).2 1 1" >> "$TEST_COVERAGE_FILE"
        done
        for i in $(seq 1 2); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/uncovered$i.go:$i.1,$((i+1)).2 1 0" >> "$TEST_COVERAGE_FILE"
        done
    elif [ "$total_coverage" = "70.0" ]; then
        # Add lines for 70%
        for i in $(seq 1 7); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/file$i.go:$i.1,$((i+1)).2 1 1" >> "$TEST_COVERAGE_FILE"
        done
        for i in $(seq 1 3); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/uncovered$i.go:$i.1,$((i+1)).2 1 0" >> "$TEST_COVERAGE_FILE"
        done
    elif [ "$total_coverage" = "60.0" ]; then
        # Add lines for 60%
        for i in $(seq 1 6); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/file$i.go:$i.1,$((i+1)).2 1 1" >> "$TEST_COVERAGE_FILE"
        done
        for i in $(seq 1 4); do
            echo "github.com/phrazzld/thinktank/cmd/thinktank/uncovered$i.go:$i.1,$((i+1)).2 1 0" >> "$TEST_COVERAGE_FILE"
        done
    fi
}

# Test 1: Threshold enforcement with environment variable
test_threshold_enforcement() {
    echo "Testing threshold enforcement..."
    
    # Test passing threshold
    create_test_coverage "80.0"
    export COVERAGE_THRESHOLD_OVERRIDE="75"
    
    # This should pass
    if ! COVERAGE_FILE="$TEST_COVERAGE_FILE" "$SCRIPT_DIR/check-coverage.sh"; then
        echo "❌ Test failed: 80% coverage should pass 75% threshold"
        return 1
    fi
    
    # Test failing threshold  
    create_test_coverage "70.0"
    
    # This should fail
    if COVERAGE_FILE="$TEST_COVERAGE_FILE" "$SCRIPT_DIR/check-coverage.sh" 2>/dev/null; then
        echo "❌ Test failed: 70% coverage should fail 75% threshold"
        return 1
    fi
    
    echo "✅ Threshold enforcement test passed"
    unset COVERAGE_THRESHOLD_OVERRIDE
    return 0
}

# Test 2: Gradual rollout mechanism
test_gradual_rollout() {
    echo "Testing gradual rollout mechanism..."
    
    create_test_coverage "60.0"
    
    # With gradual rollout enabled, should use lower threshold
    export COVERAGE_GRADUAL_ROLLOUT="true"
    export COVERAGE_THRESHOLD_CURRENT="55"
    export COVERAGE_THRESHOLD_TARGET="75"
    
    # Should pass with gradual rollout
    if ! COVERAGE_FILE="$TEST_COVERAGE_FILE" "$SCRIPT_DIR/check-coverage.sh"; then
        echo "❌ Test failed: Gradual rollout should use current threshold"
        return 1
    fi
    
    echo "✅ Gradual rollout test passed"
    unset COVERAGE_GRADUAL_ROLLOUT COVERAGE_THRESHOLD_CURRENT COVERAGE_THRESHOLD_TARGET
    return 0
}

# Test 3: Package-specific threshold validation
test_package_thresholds() {
    echo "Testing package-specific thresholds..."
    
    # Create coverage file with package details
    cat > "$TEST_COVERAGE_FILE" << EOF
mode: atomic
github.com/phrazzld/thinktank/internal/thinktank/file1.go:10.1,15.2 1 1
github.com/phrazzld/thinktank/internal/thinktank/file2.go:20.1,25.2 1 0
github.com/phrazzld/thinktank/internal/thinktank	total:	(statements)	60.0%
github.com/phrazzld/thinktank/internal/llm/file1.go:10.1,15.2 1 1
github.com/phrazzld/thinktank/internal/llm	total:	(statements)	90.0%
total:	(statements)	75.0%
EOF
    
    # Package thresholds: thinktank=50%, llm=85%
    # Coverage: thinktank=60% (pass), llm=90% (pass), overall=75% (pass)
    if ! COVERAGE_FILE="$TEST_COVERAGE_FILE" "$SCRIPT_DIR/ci/check-package-specific-coverage.sh"; then
        echo "❌ Test failed: Package thresholds should pass"
        return 1
    fi
    
    echo "✅ Package threshold test passed"
    return 0
}

# Run all tests
echo "Starting coverage threshold tests..."

test_threshold_enforcement || exit 1
test_gradual_rollout || exit 1
test_package_thresholds || exit 1

echo "✅ All coverage threshold tests passed!"