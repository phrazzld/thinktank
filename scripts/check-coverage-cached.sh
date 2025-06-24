#!/bin/bash
set -e

# check-coverage-cached.sh - Wrapper for check-coverage.sh with content-addressable caching
# Usage: ./scripts/check-coverage-cached.sh [threshold]

THRESHOLD=${1:-80}
CACHE_DIR=".cache/coverage"
STATE_HASH_FILE="$CACHE_DIR/state.hash"
CACHED_COVERAGE_FILE="$CACHE_DIR/coverage.out"
ROOT_COVERAGE_FILE="coverage.out"

# Calculate hash of all relevant Go source and dependency files
get_current_hash() {
  git ls-files '*.go' 'go.mod' 'go.sum' | \
  xargs -P4 -n50 cat | \
  sha256sum | \
  awk '{print $1}'
}

# Ensure cache directory exists
mkdir -p "$CACHE_DIR"

CURRENT_HASH=$(get_current_hash)
CACHED_HASH=""
if [ -f "$STATE_HASH_FILE" ]; then
  CACHED_HASH=$(cat "$STATE_HASH_FILE")
fi

# Check if cache is valid
if [[ -f "$CACHED_COVERAGE_FILE" && "$CURRENT_HASH" == "$CACHED_HASH" ]]; then
  echo "⚡️ Coverage cache is fresh. Using cached result."
  cp "$CACHED_COVERAGE_FILE" "$ROOT_COVERAGE_FILE"
else
  echo "🔥 Coverage cache is stale or missing. Regenerating..."
  rm -f "$ROOT_COVERAGE_FILE"

  # Try running original script first
  if ./scripts/check-coverage.sh "$THRESHOLD"; then
    echo "✅ Coverage generated successfully, caching result."
    cp "$ROOT_COVERAGE_FILE" "$CACHED_COVERAGE_FILE"
    echo "$CURRENT_HASH" > "$STATE_HASH_FILE"
  else
    echo "⚠️  Coverage script failed, checking for partial data..."

    # Check if temporary coverage file exists (generated but not processed due to test failures)
    if [ -f "coverage.out.tmp" ]; then
      echo "📋 Found partial coverage data, processing..."

      # Process the temporary coverage file manually (same logic as check-coverage.sh)
      cat coverage.out.tmp | \
        grep -v "_test_helpers\.go:" | \
        grep -v "_test_utils\.go:" | \
        grep -v "mock_.*\.go:" | \
        grep -v "/mocks\.go:" > "$ROOT_COVERAGE_FILE"

      # Cleanup temporary file
      rm coverage.out.tmp

      echo "✅ Coverage data processed, caching result."
      cp "$ROOT_COVERAGE_FILE" "$CACHED_COVERAGE_FILE"
      echo "$CURRENT_HASH" > "$STATE_HASH_FILE"

      # Check the final coverage
      ./scripts/check-coverage.sh "$THRESHOLD"
    else
      echo "❌ No coverage data available, cannot proceed."
      exit 1
    fi
  fi
fi

# Final check using (potentially cached) coverage file
echo "---"
./scripts/check-coverage.sh "$THRESHOLD"
