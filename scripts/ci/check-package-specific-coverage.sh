#!/bin/bash
set -e

# check-package-specific-coverage.sh - Check coverage for specific packages against individualized thresholds
# Usage: scripts/ci/check-package-specific-coverage.sh
#
# This script checks test coverage for specific packages against individualized thresholds.
# It allows setting different coverage requirements for different packages based on
# their current state and importance. This is useful during a transition period when
# some packages have lower coverage that's being gradually improved.
#
# Package thresholds are defined in this script and documented in coverage-analysis.md.
# The CI workflow enforces these thresholds by running this script.

# Define the overall threshold (default 75%)
OVERALL_THRESHOLD=${OVERALL_THRESHOLD:-75}

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Define package-specific thresholds
# NOTE: These thresholds are based on the coverage-analysis.md document

# Define package paths
PKG_THINKTANK="${MODULE_PATH}/internal/thinktank"
PKG_PROVIDERS="${MODULE_PATH}/internal/providers"
PKG_REGISTRY="${MODULE_PATH}/internal/registry"
PKG_LLM="${MODULE_PATH}/internal/llm"

# Define thresholds for each package
THRESHOLD_THINKTANK=50    # Lower initial target due to current 18.3%
THRESHOLD_PROVIDERS=85    # Higher due to current 86.2%
THRESHOLD_REGISTRY=80     # Current is 80.9%
THRESHOLD_LLM=85          # Current is 87.6%

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
echo "📊 Package-Specific Coverage Report"
echo "======================================================="

# Variable to track failures
FAILED=0

# Function to extract package coverage
extract_package_coverage() {
  local package_path=$1
  # Get the total coverage for the specified package
  coverage=$(go tool cover -func=coverage.out | grep "$package_path" | grep "total:" | awk '{print $3}' | tr -d '%')
  echo $coverage
}

# Check specific packages first
echo "Checking specific packages against their thresholds:"

# Check internal/thinktank
coverage=$(extract_package_coverage "$PKG_THINKTANK")
if [ -n "$coverage" ]; then
  if (( $(echo "$coverage < $THRESHOLD_THINKTANK" | bc -l) )); then
    printf "❌ %-60s %6.1f%% (below threshold of %s%%)\n" "$PKG_THINKTANK" "$coverage" "$THRESHOLD_THINKTANK"
    FAILED=$((FAILED+1))
  else
    printf "✅ %-60s %6.1f%% (threshold: %s%%)\n" "$PKG_THINKTANK" "$coverage" "$THRESHOLD_THINKTANK"
  fi
fi

# Check internal/providers
coverage=$(extract_package_coverage "$PKG_PROVIDERS")
if [ -n "$coverage" ]; then
  if (( $(echo "$coverage < $THRESHOLD_PROVIDERS" | bc -l) )); then
    printf "❌ %-60s %6.1f%% (below threshold of %s%%)\n" "$PKG_PROVIDERS" "$coverage" "$THRESHOLD_PROVIDERS"
    FAILED=$((FAILED+1))
  else
    printf "✅ %-60s %6.1f%% (threshold: %s%%)\n" "$PKG_PROVIDERS" "$coverage" "$THRESHOLD_PROVIDERS"
  fi
fi

# Check internal/registry
coverage=$(extract_package_coverage "$PKG_REGISTRY")
if [ -n "$coverage" ]; then
  if (( $(echo "$coverage < $THRESHOLD_REGISTRY" | bc -l) )); then
    printf "❌ %-60s %6.1f%% (below threshold of %s%%)\n" "$PKG_REGISTRY" "$coverage" "$THRESHOLD_REGISTRY"
    FAILED=$((FAILED+1))
  else
    printf "✅ %-60s %6.1f%% (threshold: %s%%)\n" "$PKG_REGISTRY" "$coverage" "$THRESHOLD_REGISTRY"
  fi
fi

# Check internal/llm
coverage=$(extract_package_coverage "$PKG_LLM")
if [ -n "$coverage" ]; then
  if (( $(echo "$coverage < $THRESHOLD_LLM" | bc -l) )); then
    printf "❌ %-60s %6.1f%% (below threshold of %s%%)\n" "$PKG_LLM" "$coverage" "$THRESHOLD_LLM"
    FAILED=$((FAILED+1))
  else
    printf "✅ %-60s %6.1f%% (threshold: %s%%)\n" "$PKG_LLM" "$coverage" "$THRESHOLD_LLM"
  fi
fi

echo "======================================================="

# Get the total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | tr -d '%')
TOTAL_COVERAGE_ROUNDED=$(printf "%.0f" "$TOTAL_COVERAGE")

echo "Total code coverage: $TOTAL_COVERAGE% (threshold: $OVERALL_THRESHOLD%)"

# Check if total coverage meets the threshold
if (( $(echo "$TOTAL_COVERAGE < $OVERALL_THRESHOLD" | bc -l) )); then
  echo "❌ Overall coverage check failed ($TOTAL_COVERAGE% < $OVERALL_THRESHOLD%)"
  FAILED=$((FAILED+1))
else
  echo "✅ Overall coverage check passed ($TOTAL_COVERAGE% >= $OVERALL_THRESHOLD%)"
fi

# Exit with failure if any check failed
if [ $FAILED -gt 0 ]; then
  echo "❌ Coverage check failed: $FAILED package(s) don't meet their threshold"
  exit 1
else
  echo "✅ All packages meet their coverage thresholds"
  exit 0
fi
