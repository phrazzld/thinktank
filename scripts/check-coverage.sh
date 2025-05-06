#!/bin/bash
set -e

# check-coverage.sh - Verify that test coverage meets or exceeds the threshold
# Usage: scripts/check-coverage.sh [threshold_percentage] [show_registry_api]

# Default threshold is 75% (increased from 55%, target is 90%)
THRESHOLD=${1:-75}
SHOW_REGISTRY_API=${2:-"false"}

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

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
  go test -short -coverprofile=coverage.out.tmp -covermode=atomic $PACKAGES

  # Process coverage file to exclude test helper files by pattern
  cat coverage.out.tmp | \
    grep -v "_test_helpers\.go:" | \
    grep -v "_test_utils\.go:" | \
    grep -v "mock_.*\.go:" | \
    grep -v "/mocks\.go:" > coverage.out

  # Cleanup temporary file
  rm coverage.out.tmp
}

# Generate coverage if file doesn't exist
if [ ! -f coverage.out ]; then
  generate_coverage
fi

# Extract the total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | grep -v "_total:" | awk '{print $3}' | tr -d '%')
echo "Total code coverage: $COVERAGE%"

# Show registry API coverage if requested
if [ "$SHOW_REGISTRY_API" = "true" ]; then
  echo -e "\n📊 Registry API Coverage:"
  echo "======================================================="
  go tool cover -func=coverage.out | grep "registry_api.go" | awk '{
    printf "  %-60s %6s\n", $1, $3
  }'
  echo "======================================================="

  # Calculate average registry API coverage
  REGISTRY_API_COVERAGE=$(go tool cover -func=coverage.out | grep "registry_api.go" | awk '
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
  go tool cover -func=coverage.out | grep "total:" | grep -v "^total:" | awk -v threshold="$THRESHOLD" '{
    package=$1;
    coverage=$3;
    gsub(/%/, "", coverage);
    if (coverage < threshold) {
      printf "  ❌ %-60s %6s\n", package, coverage "%";
    }
  }'

  exit 1
fi
