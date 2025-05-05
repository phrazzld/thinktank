#!/bin/bash
set -e

# check-package-coverage.sh - Report test coverage for each package and highlight those below threshold
# Usage: scripts/check-package-coverage.sh [threshold_percentage] [show_registry_api]

# Default threshold is 75% (increased from 55%, target is 90%)
THRESHOLD=${1:-75}
SHOW_REGISTRY_API=${2:-"false"}
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
go tool cover -func=coverage.out | grep "total:" | grep -v "^total:" | awk -v threshold="$THRESHOLD" '{
  package=$1;
  coverage=$3;
  gsub(/%/, "", coverage);

  if (coverage < threshold) {
    printf "❌ %-60s %6s%% (below threshold)\n", package, coverage;
    failed += 1;
  } else {
    printf "✅ %-60s %6s%%\n", package, coverage;
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
