#!/bin/bash
set -e

# pre-submit-coverage.sh - Comprehensive coverage check script for pre-submission verification
# Usage: scripts/pre-submit-coverage.sh [-t threshold] [-r] [-v]
#   -t, --threshold <value>   Set coverage threshold (default: 75%)
#   -r, --registry           Include registry API coverage checks
#   -v, --verbose            Show detailed coverage information

# Default values
THRESHOLD=75
CHECK_REGISTRY=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -t|--threshold)
      THRESHOLD="$2"
      shift 2
      ;;
    -r|--registry)
      CHECK_REGISTRY=true
      shift
      ;;
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    -h|--help)
      echo "Usage: scripts/pre-submit-coverage.sh [options]"
      echo "Options:"
      echo "  -t, --threshold <value>   Set coverage threshold (default: 75%)"
      echo "  -r, --registry            Include registry API coverage checks"
      echo "  -v, --verbose             Show detailed coverage information"
      echo "  -h, --help                Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help to see available options."
      exit 1
      ;;
  esac
done

# Header
echo "🔍 Pre-Submission Coverage Check (Threshold: ${THRESHOLD}%)"
echo "======================================================="

# Determine the module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Function to generate coverage data for tests
generate_coverage() {
  echo "Generating coverage data..."
  # Skip integration and e2e tests as they're slow and have different coverage characteristics
  PACKAGES=$(go list ./... | grep -v "${MODULE_PATH}/internal/integration" | grep -v "${MODULE_PATH}/internal/e2e" | grep -v "/disabled/")
  go test -short -coverprofile=coverage.out -covermode=atomic $PACKAGES
}

# Always generate fresh coverage for pre-submission
if [ -f coverage.out ]; then
  rm coverage.out
fi
generate_coverage

# Extract the total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | grep -v "_total:" | awk '{print $3}' | tr -d '%')
PASS_TOTAL=$(awk -v coverage="$COVERAGE" -v threshold="$THRESHOLD" 'BEGIN { print (coverage >= threshold) }')

# Store errors
ERRORS=0
ERROR_OUTPUT=""

# Check if total coverage passes
if [ "$PASS_TOTAL" -eq 1 ]; then
  echo "✅ Overall coverage check passed: ${COVERAGE}% (threshold: ${THRESHOLD}%)"
else
  echo "❌ Overall coverage check failed: ${COVERAGE}% (threshold: ${THRESHOLD}%)"
  ERRORS=$((ERRORS + 1))
  ERROR_OUTPUT="${ERROR_OUTPUT}\n- Overall coverage is below the threshold: ${COVERAGE}% < ${THRESHOLD}%"
fi

# Check package coverage
# First find packages below threshold directly
PACKAGES_BELOW_THRESHOLD=$(go tool cover -func=coverage.out | grep "total:" | grep -v "^total:" | awk -v threshold="$THRESHOLD" '
{
  package=$1;
  coverage=$3;
  gsub(/%/, "", coverage);
  if (coverage < threshold) {
    printf "  ❌ %-60s %6s%%\n", package, coverage;
  }
}')

if [ -z "$PACKAGES_BELOW_THRESHOLD" ]; then
  echo "✅ Package coverage check passed: All packages meet the threshold"
else
  echo "❌ Package coverage check failed: Some packages below threshold"
  ERRORS=$((ERRORS + 1))
  ERROR_OUTPUT="${ERROR_OUTPUT}\n- Some packages are below the threshold:\n${PACKAGES_BELOW_THRESHOLD}"
fi

# Check registry API coverage if requested
if [ "$CHECK_REGISTRY" = true ]; then
  REGISTRY_OUTPUT=$(scripts/check-registry-coverage.sh "$THRESHOLD" 2>&1)
  REGISTRY_RESULT=$?

  if [ "$REGISTRY_RESULT" -eq 0 ]; then
    echo "✅ Registry API coverage check passed"
  else
    echo "❌ Registry API coverage check failed"
    ERRORS=$((ERRORS + 1))

    # Extract registry API coverage information
    REGISTRY_COVERAGE=$(echo "$REGISTRY_OUTPUT" | grep "Average Registry API Coverage:" | sed 's/.*: //' | tr -d '%')
    ERROR_OUTPUT="${ERROR_OUTPUT}\n- Registry API coverage is below threshold: ${REGISTRY_COVERAGE}% < ${THRESHOLD}%"
  fi
fi

# Show detailed output if verbose mode is enabled
if [ "$VERBOSE" = true ]; then
  echo -e "\n📋 Detailed Coverage Information:"
  echo "======================================================="

  # Show overall coverage breakdown
  echo -e "\n📊 Overall Coverage Breakdown:"
  echo "======================================================="
  go tool cover -func=coverage.out | grep "total:" | awk '{printf "  %-60s %6s\n", $1, $3}'
  echo "======================================================="

  # Show package coverage breakdown - use direct command output
  echo -e "\n📊 Package Coverage Breakdown:"
  echo "======================================================="
  # Run test command again with coverage flag
  go test -short ./... -cover | grep -v "no test files" | grep -v "no statements" | sort -k5

  # Show registry API coverage if requested
  if [ "$CHECK_REGISTRY" = true ]; then
    echo -e "\n📊 Registry API Coverage:"
    go tool cover -func=coverage.out | grep "registry_api.*\.go" | awk '{printf "  %-60s %6s\n", $1, $3}'
  fi

  echo "======================================================="
fi

# Final result
echo -e "\n📝 Summary:"
if [ "$ERRORS" -eq 0 ]; then
  echo "✅ All coverage checks passed! Your code is ready for submission."
  exit 0
else
  echo "❌ ${ERRORS} coverage check(s) failed. Please fix the following issues:"
  echo -e "${ERROR_OUTPUT}"
  echo "Run with --verbose for more details."
  exit 1
fi
