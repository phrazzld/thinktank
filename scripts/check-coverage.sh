#!/bin/bash
set -e

# check-coverage.sh - Verify that test coverage meets or exceeds the threshold
# Usage: scripts/check-coverage.sh [threshold_percentage] [show_registry_api]

# Threshold configuration with gradual rollout support
DEFAULT_THRESHOLD=75  # Target threshold
CURRENT_THRESHOLD=55  # Current achievable threshold

# Support for gradual rollout via environment variables
if [ "$COVERAGE_GRADUAL_ROLLOUT" = "true" ]; then
    # Use current threshold during gradual rollout
    EFFECTIVE_THRESHOLD=${COVERAGE_THRESHOLD_CURRENT:-$CURRENT_THRESHOLD}
    echo "🔄 Gradual rollout enabled: using current threshold $EFFECTIVE_THRESHOLD% (target: ${COVERAGE_THRESHOLD_TARGET:-$DEFAULT_THRESHOLD}%)"
else
    # Use override if provided, otherwise command line argument, otherwise default
    EFFECTIVE_THRESHOLD=${COVERAGE_THRESHOLD_OVERRIDE:-${1:-$DEFAULT_THRESHOLD}}
fi

THRESHOLD=$EFFECTIVE_THRESHOLD
SHOW_REGISTRY_API=${2:-"false"}

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Coverage file configuration (for testing support)
COVERAGE_OUT=${COVERAGE_FILE:-"coverage.out"}

# Function to generate coverage data if not already present
generate_coverage() {
  echo "Generating coverage data..."
  # Skip packages and files that shouldn't be included in coverage metrics:
  # - integration and e2e tests (slow and have different coverage characteristics)
  # - disabled code
  # - test helper packages and files (testutil, mock implementations, test utilities)
  PACKAGES=$(go list ./... | \
    grep -v "${MODULE_PATH}/internal/integration" | \
    grep -v "${MODULE_PATH}/internal/e2e" | \
    grep -v "/disabled/" | \
    grep -v "${MODULE_PATH}/internal/testutil")

  # Run tests and generate coverage profile
  go test -short -coverprofile=${COVERAGE_OUT}.tmp -covermode=atomic $PACKAGES

  # Process coverage file to exclude test helper files by pattern
  cat ${COVERAGE_OUT}.tmp | \
    grep -v "_test_helpers\.go:" | \
    grep -v "_test_utils\.go:" | \
    grep -v "mock_.*\.go:" | \
    grep -v "/mocks\.go:" > $COVERAGE_OUT

  # Cleanup temporary file
  rm ${COVERAGE_OUT}.tmp
}

# Generate coverage if file doesn't exist
if [ ! -f $COVERAGE_OUT ]; then
  generate_coverage
fi

# Extract the total coverage percentage
COVERAGE=$(go tool cover -func=$COVERAGE_OUT | grep "total:" | grep -v "_total:" | awk '{print $3}' | tr -d '%')
echo "Total code coverage: $COVERAGE%"

# Show registry API coverage if requested
if [ "$SHOW_REGISTRY_API" = "true" ]; then
  echo -e "\n📊 Registry API Coverage:"
  echo "======================================================="
  go tool cover -func=$COVERAGE_OUT | grep "registry_api.go" | awk '{
    printf "  %-60s %6s\n", $1, $3
  }'
  echo "======================================================="

  # Calculate average registry API coverage
  REGISTRY_API_COVERAGE=$(go tool cover -func=$COVERAGE_OUT | grep "registry_api.go" | awk '
    BEGIN { total=0; count=0; }
    {
      coverage=$3;
      gsub(/%/, "", coverage);
      total += coverage;
      count++;
    }
    END {
      if (count > 0) printf "%.1f", total/count;
      else print "0";
    }
  ')
  echo "Average Registry API coverage: ${REGISTRY_API_COVERAGE}%"
fi

# Compare with threshold using awk for proper float comparison
PASS=$(awk -v coverage="$COVERAGE" -v threshold="$THRESHOLD" 'BEGIN { print (coverage >= threshold) }')

if [ "$PASS" -eq 1 ]; then
  echo "✅ Coverage check passed ($COVERAGE% >= $THRESHOLD%)"
  exit 0
else
  echo "❌ Coverage check failed ($COVERAGE% < $THRESHOLD%)"

  # Show packages with coverage below threshold for easier debugging
  echo -e "\nPackages below the $THRESHOLD% threshold:"
  go tool cover -func=$COVERAGE_OUT | grep "total:" | grep -v "^total:" | awk -v threshold="$THRESHOLD" '{
    package=$1;
    coverage=$3;
    gsub(/%/, "", coverage);
    if (coverage < threshold) {
      printf "  ❌ %-60s %6s\n", package, coverage "%";
    }
  }'

  exit 1
fi
