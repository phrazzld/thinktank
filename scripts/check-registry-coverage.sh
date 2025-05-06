#!/bin/bash
set -e

# check-registry-coverage.sh - Report detailed coverage stats for the registry API
# Usage: scripts/check-registry-coverage.sh [threshold_percentage]

# Default threshold is 75% (increased from 55%, target is 90%)
THRESHOLD=${1:-75}

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

# Print header
echo "📊 Registry API Coverage Report (Threshold: ${THRESHOLD}%)"
echo "======================================================="

# Calculate average registry API coverage
REGISTRY_API_COVERAGE=$(go tool cover -func=coverage.out | grep "registry_api.*\.go" | awk '
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

echo "Average Registry API Coverage: ${REGISTRY_API_COVERAGE}%"

# Compare with threshold using awk for proper float comparison
PASS=$(awk -v coverage="$REGISTRY_API_COVERAGE" -v threshold="$THRESHOLD" 'BEGIN { print (coverage >= threshold) }')

# Show registry API file details
echo -e "\n📋 Registry API Files:"
echo "======================================================="
go tool cover -func=coverage.out | grep "registry_api.*\.go" | awk '{
  fn=$2;
  coverage=$3;
  file=$1;
  gsub(".*/", "", file);
  printf "  %-30s %-30s %6s\n", file, fn, coverage;
}'
echo "======================================================="

# Show registry adapters coverage
echo -e "\n📋 Adapter Methods:"
echo "======================================================="
go tool cover -func=coverage.out | grep "/adapters\.go" | awk '{
  fn=$2;
  coverage=$3;
  printf "  %-45s %6s\n", fn, coverage;
}'
echo "======================================================="

# Check if registry API coverage meets threshold
if [ "$PASS" -eq 1 ]; then
  echo "✅ Registry API coverage check passed (${REGISTRY_API_COVERAGE}% >= ${THRESHOLD}%)"
  exit 0
else
  echo "❌ Registry API coverage check failed (${REGISTRY_API_COVERAGE}% < ${THRESHOLD}%)"

  # Show methods below threshold for easier debugging
  echo -e "\nRegistry API methods below the ${THRESHOLD}% threshold:"
  go tool cover -func=coverage.out | grep "registry_api.*\.go" | awk -v threshold="$THRESHOLD" '{
    fn=$2;
    coverage=$3;
    gsub(/%/, "", coverage);
    if (coverage < threshold) {
      printf "  ❌ %-45s %6s%%\n", fn, coverage;
    }
  }'

  exit 1
fi
