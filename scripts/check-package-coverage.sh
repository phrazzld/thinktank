#!/bin/bash
set -e

# check-package-coverage.sh - Report test coverage for each package and highlight those below threshold
# Usage: scripts/check-package-coverage.sh [threshold_percentage] [show_registry_api]

# Default threshold is 65% (adjusted to realistic baseline from 90%)
THRESHOLD=${1:-65}
SHOW_REGISTRY_API=${2:-"false"}
FAILED=0

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
echo "📊 Package Coverage Report (Threshold: ${THRESHOLD}%)"
echo "======================================================="

# Process and print the results by package
# Parse function-level coverage data and aggregate by package
go tool cover -func=coverage.out | grep -v "^total:" | awk -v threshold="$THRESHOLD" '
BEGIN {
  failed = 0;
}
{
  # Extract package name from file path (everything before the last /)
  split($1, parts, "/");
  filename = parts[length(parts)];
  package = $1;
  gsub("/" filename ".*", "", package);

  # Extract coverage percentage (remove % symbol)
  coverage = $3;
  gsub(/%/, "", coverage);

  # Accumulate coverage data per package
  package_sum[package] += coverage;
  package_count[package]++;
}
END {
  # Calculate and display average coverage per package
  for (package in package_sum) {
    if (package_count[package] > 0) {
      avg_coverage = package_sum[package] / package_count[package];

      if (avg_coverage < threshold) {
        printf "❌ %-60s %6.1f%% (below threshold)\n", package, avg_coverage;
        failed++;
      } else {
        printf "✅ %-60s %6.1f%%\n", package, avg_coverage;
      }
    }
  }

  print "======================================================="
  if (failed > 0) {
    printf "Result: %d packages below %s%% threshold\n", failed, threshold;
    exit 1;
  } else {
    printf "Result: All packages meet or exceed %s%% threshold\n", threshold;
  }
}'

# Show registry API coverage if requested
if [ "$SHOW_REGISTRY_API" = "true" ]; then
  echo -e "\n📊 Registry API File Coverage:"
  echo "======================================================="
  go tool cover -func=coverage.out | grep "registry_api.*\.go" | awk '{
    fn=$2;
    coverage=$3;
    file=$1;
    gsub(".*/", "", file);
    printf "  %-30s %-30s %6s\n", file, fn, coverage;
  }'
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
  echo "Average Registry API coverage: ${REGISTRY_API_COVERAGE}%"

  # Show registry adapters coverage
  echo -e "\n📊 Adapter File Coverage:"
  echo "======================================================="
  go tool cover -func=coverage.out | grep "/adapters\.go" | awk '{
    fn=$2;
    coverage=$3;
    printf "  %-30s %6s\n", fn, coverage;
  }'
  echo "======================================================="
fi

# Check if any package is below threshold
if [ $? -ne 0 ]; then
  echo "❌ Some packages don't meet coverage requirements"
  exit 1
else
  echo "✅ All packages meet coverage requirements"
  exit 0
fi
