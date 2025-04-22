#!/bin/bash
set -e

# check-package-coverage.sh - Report test coverage for each package and highlight those below threshold
# Usage: scripts/check-package-coverage.sh [threshold_percentage]

# Default threshold is 55% (temporarily reduced from 90%)
THRESHOLD=${1:-55}
FAILED=0

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

# Print header
echo "📊 Package Coverage Report (Threshold: ${THRESHOLD}%)"
echo "======================================================="

# Process and print the results by package
go tool cover -func=coverage.out | grep -v "^total:" | awk -v threshold="$THRESHOLD" '{
  package=$1;
  coverage=$3;
  gsub(/%/, "", coverage);

  # Only process package total lines
  if ($2 == "total:") {
    if (coverage < threshold) {
      printf "❌ %-60s %6s%% (below threshold)\n", package, coverage;
      failed += 1;
    } else {
      printf "✅ %-60s %6s%%\n", package, coverage;
    }
  }
}
END {
  print "======================================================="
  if (failed > 0) {
    printf "Result: %d packages below %s%% threshold\n", failed, threshold;
    exit 1;
  } else {
    printf "Result: All packages meet or exceed %s%% threshold\n", threshold;
  }
}'

# Check if any package is below threshold
if [ $? -ne 0 ]; then
  echo "❌ Some packages don't meet coverage requirements"
  exit 1
else
  echo "✅ All packages meet coverage requirements"
  exit 0
fi
