#!/bin/bash
set -e

# check-coverage-fast.sh - Optimized coverage checking for pre-commit hooks
# Usage: ./scripts/check-coverage-fast.sh [threshold]

THRESHOLD=${1:-79}
CACHE_DIR=".cache/coverage"
STATE_HASH_FILE="$CACHE_DIR/state.hash"
PACKAGE_HASH_FILE="$CACHE_DIR/package.hash"
CACHED_COVERAGE_FILE="$CACHE_DIR/coverage.out"
ROOT_COVERAGE_FILE="coverage.out"
TIMEOUT_FILE="$CACHE_DIR/timeout.flag"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[coverage-fast]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[coverage-fast]${NC} ⚠️  $1"
}

error() {
    echo -e "${RED}[coverage-fast]${NC} ❌ $1"
}

success() {
    echo -e "${GREEN}[coverage-fast]${NC} ✅ $1"
}

# Ensure cache directory exists
mkdir -p "$CACHE_DIR"

# Check if this is a documentation-only change
get_changed_files() {
    if git rev-parse --verify HEAD >/dev/null 2>&1; then
        git diff --name-only --cached
    else
        # Initial commit - check all tracked files
        git ls-files
    fi
}

is_docs_only_change() {
    local changed_files=$(get_changed_files)
    local non_docs_files=$(echo "$changed_files" | grep -v -E '\.(md|txt|yaml|yml|json)$|^docs/|^\.github/|LICENSE|README' || true)

    if [ -z "$non_docs_files" ]; then
        return 0  # docs-only change
    else
        return 1  # has non-docs changes
    fi
}

# Quick skip for documentation-only changes
if is_docs_only_change; then
    success "Documentation-only change detected. Skipping coverage check."
    exit 0
fi

# Calculate hash of Go source files and dependencies (optimized)
get_current_hash() {
    # Only hash Go files that actually affect coverage
    local go_files=$(git ls-files '*.go' | grep -v '_test\.go$' || true)
    local mod_files=$(git ls-files 'go.mod' 'go.sum' || true)

    if [ -n "$go_files" ] || [ -n "$mod_files" ]; then
        (echo "$go_files $mod_files" | tr ' ' '\n' | xargs -P8 -n25 cat 2>/dev/null || true) | \
        sha256sum | awk '{print $1}'
    else
        echo "no-go-files"
    fi
}

# Calculate hash of packages to test (for more granular caching)
get_package_hash() {
    go list ./... 2>/dev/null | \
    grep -v "/internal/integration" | \
    grep -v "/internal/e2e" | \
    grep -v "/disabled/" | \
    grep -v "/internal/testutil" | \
    sort | sha256sum | awk '{print $1}'
}

# Check if we're in a timeout recovery situation
check_timeout_recovery() {
    if [ -f "$TIMEOUT_FILE" ]; then
        local timeout_age=$(( $(date +%s) - $(stat -f %m "$TIMEOUT_FILE" 2>/dev/null || stat -c %Y "$TIMEOUT_FILE" 2>/dev/null || echo 0) ))
        if [ $timeout_age -lt 3600 ]; then  # Less than 1 hour old
            warn "Recent timeout detected. Using cached coverage if available."
            return 0
        else
            rm -f "$TIMEOUT_FILE"
        fi
    fi
    return 1
}

# Trap timeout and save state
handle_timeout() {
    touch "$TIMEOUT_FILE"
    error "Coverage check timed out. Using cached result if available."

    if [ -f "$CACHED_COVERAGE_FILE" ]; then
        warn "Using stale cached coverage data due to timeout."
        cp "$CACHED_COVERAGE_FILE" "$ROOT_COVERAGE_FILE"
        ./scripts/check-coverage.sh "$THRESHOLD" || {
            error "Even cached coverage fails threshold. Please run 'make test-coverage' manually."
            exit 1
        }
        exit 0
    else
        error "No cached coverage available. Please run 'make test-coverage' manually."
        exit 1
    fi
}

# Set up timeout handling
trap 'handle_timeout' TERM

CURRENT_HASH=$(get_current_hash)
PACKAGE_HASH=$(get_package_hash)
CACHED_HASH=""
CACHED_PACKAGE_HASH=""

if [ -f "$STATE_HASH_FILE" ]; then
    CACHED_HASH=$(cat "$STATE_HASH_FILE")
fi

if [ -f "$PACKAGE_HASH_FILE" ]; then
    CACHED_PACKAGE_HASH=$(cat "$PACKAGE_HASH_FILE")
fi

# Check for timeout recovery first
if check_timeout_recovery && [ -f "$CACHED_COVERAGE_FILE" ]; then
    log "Using cached coverage due to recent timeout."
    cp "$CACHED_COVERAGE_FILE" "$ROOT_COVERAGE_FILE"
    ./scripts/check-coverage.sh "$THRESHOLD"
    exit $?
fi

# Check if cache is valid (both source and package structure unchanged)
if [[ -f "$CACHED_COVERAGE_FILE" && "$CURRENT_HASH" == "$CACHED_HASH" && "$PACKAGE_HASH" == "$CACHED_PACKAGE_HASH" ]]; then
    log "Coverage cache is fresh. Using cached result."
    cp "$CACHED_COVERAGE_FILE" "$ROOT_COVERAGE_FILE"
    ./scripts/check-coverage.sh "$THRESHOLD"
    exit $?
fi

# Cache is stale, need to regenerate
log "Coverage cache is stale or missing. Running optimized coverage generation..."
rm -f "$ROOT_COVERAGE_FILE" "$TIMEOUT_FILE"

# Get packages to test (exclude slow integration tests)
PACKAGES=$(go list ./... | \
    grep -v "/internal/integration" | \
    grep -v "/internal/e2e" | \
    grep -v "/disabled/" | \
    grep -v "/internal/testutil")

PACKAGE_COUNT=$(echo "$PACKAGES" | wc -l)
log "Running coverage tests on $PACKAGE_COUNT packages (parallel execution)..."

# Run tests with optimized settings for speed
start_time=$(date +%s)

# Use parallel execution and shorter timeout per package
if go test -short -parallel 8 -timeout=6m -coverprofile=coverage.out.tmp -covermode=atomic $PACKAGES; then
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    log "Tests completed successfully in ${duration}s"

    # Process coverage file to exclude test helper files
    cat coverage.out.tmp | \
        grep -v "_test_helpers\.go:" | \
        grep -v "_test_utils\.go:" | \
        grep -v "mock_.*\.go:" | \
        grep -v "/mocks\.go:" > "$ROOT_COVERAGE_FILE"

    # Cache the results
    cp "$ROOT_COVERAGE_FILE" "$CACHED_COVERAGE_FILE"
    echo "$CURRENT_HASH" > "$STATE_HASH_FILE"
    echo "$PACKAGE_HASH" > "$PACKAGE_HASH_FILE"

    # Clean up
    rm -f coverage.out.tmp "$TIMEOUT_FILE"

    success "Coverage generated and cached successfully."
else
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    warn "Some tests failed after ${duration}s, checking for partial coverage data..."

    # Check if we have partial coverage data
    if [ -f "coverage.out.tmp" ]; then
        log "Processing partial coverage data..."

        cat coverage.out.tmp | \
            grep -v "_test_helpers\.go:" | \
            grep -v "_test_utils\.go:" | \
            grep -v "mock_.*\.go:" | \
            grep -v "/mocks\.go:" > "$ROOT_COVERAGE_FILE"

        # Cache even partial results
        cp "$ROOT_COVERAGE_FILE" "$CACHED_COVERAGE_FILE"
        echo "$CURRENT_HASH" > "$STATE_HASH_FILE"
        echo "$PACKAGE_HASH" > "$PACKAGE_HASH_FILE"

        rm -f coverage.out.tmp

        warn "Using partial coverage data. Some packages may have failed tests."
    else
        error "No coverage data generated. Tests may have failed early."

        # Try to use cached data if available
        if [ -f "$CACHED_COVERAGE_FILE" ]; then
            warn "Falling back to cached coverage data."
            cp "$CACHED_COVERAGE_FILE" "$ROOT_COVERAGE_FILE"
        else
            error "No coverage data available. Please fix test failures and try again."
            exit 1
        fi
    fi
fi

# Final coverage check
log "Checking coverage against ${THRESHOLD}% threshold..."
./scripts/check-coverage.sh "$THRESHOLD"
