#!/bin/bash
set -e

# check-coverage.sh - Verify that test coverage meets or exceeds the threshold
# Usage: scripts/check-coverage.sh [threshold_percentage]

# Default threshold is 75% (increased from 55%, target is 90%)
THRESHOLD=${1:-75}

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Function to generate coverage data if not already present
generate_coverage() {
  echo "Generating coverage data..."
  # Skip integration and e2e tests as they're slow and have different coverage characteristics
  PACKAGES=$(go list ./... | grep -v "${MODULE_PATH}/internal/integration" | grep -v "${MODULE_PATH}/internal/e2e" | grep -v "/disabled/")
  go test -short -coverprofile=coverage.out -covermode=atomic $PACKAGES
}

# Generate coverage if file doesn't exist
if [ ! -f coverage.out ]; then
  generate_coverage
fi

# Extract the total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
echo "Total code coverage: $COVERAGE%"

# Compare with threshold using awk for proper float comparison
PASS=$(awk -v coverage="$COVERAGE" -v threshold="$THRESHOLD" 'BEGIN { print (coverage >= threshold) }')

if [ "$PASS" -eq 1 ]; then
  echo "✅ Coverage check passed ($COVERAGE% >= $THRESHOLD%)"
  exit 0
else
  echo "❌ Coverage check failed ($COVERAGE% < $THRESHOLD%)"

  # Show packages with coverage below threshold for easier debugging
  echo -e "\nPackages below the $THRESHOLD% threshold:"
  go tool cover -func=coverage.out | grep -v "total:" | awk -v threshold="$THRESHOLD" '{
    coverage=$3;
    gsub(/%/, "", coverage);
    if (coverage < threshold) {
      printf "  ❌ %-60s %6s\n", $1, $3
    }
  }'

  exit 1
fi
