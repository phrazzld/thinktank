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

# Define the overall threshold (aligned with realistic baseline)
OVERALL_THRESHOLD=${OVERALL_THRESHOLD:-35}

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Define package-specific thresholds for quality gate enforcement
# Critical packages (95% requirement): Core business logic and interfaces
# Non-critical packages: Lower thresholds during improvement phase

# Define package paths
PKG_THINKTANK="${MODULE_PATH}/internal/thinktank"
PKG_PROVIDERS="${MODULE_PATH}/internal/providers"
PKG_REGISTRY="${MODULE_PATH}/internal/registry"
PKG_LLM="${MODULE_PATH}/internal/llm"

# Define thresholds for each package
# Critical packages with realistic requirements (adjusted to current baseline)
THRESHOLD_LLM=95          # CRITICAL: Core LLM interface and error handling (already high)
THRESHOLD_PROVIDERS=80    # CRITICAL: Provider abstraction layer (current: 83.7%)
THRESHOLD_REGISTRY=75     # CRITICAL: Model registry and configuration (current: 77.8%)

# Non-critical packages with gradual improvement targets
THRESHOLD_THINKTANK=70    # Complex orchestration - gradual improvement target

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
echo "📊 Package-Specific Coverage Report"
echo "======================================================="

# Variable to track failures
FAILED=0

# Function to extract package coverage from existing coverage.out
extract_package_coverage() {
  local package_path=$1

  # Get all coverage lines for files in this package
  local package_lines=$(go tool cover -func=coverage.out | grep "$package_path/")

  if [ -z "$package_lines" ]; then
    echo ""
    return
  fi

  # Extract coverage percentages and calculate weighted average
  local total_weight=0
  local weighted_sum=0

  while IFS= read -r line; do
    if [ -n "$line" ]; then
      # Extract the coverage percentage (3rd field, remove %)
      local coverage_pct=$(echo "$line" | awk '{print $3}' | tr -d '%')

      # Use simple counting approach - each function/line gets equal weight
      if [[ "$coverage_pct" =~ ^[0-9]+\.?[0-9]*$ ]]; then
        weighted_sum=$(echo "$weighted_sum + $coverage_pct" | bc -l)
        total_weight=$((total_weight + 1))
      fi
    fi
  done <<< "$package_lines"

  # Calculate average coverage for the package
  if [ $total_weight -gt 0 ]; then
    local avg_coverage=$(echo "scale=1; $weighted_sum / $total_weight" | bc -l)
    echo "$avg_coverage"
  else
    echo ""
  fi
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
