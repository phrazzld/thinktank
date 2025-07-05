#!/bin/bash

# precommit-troubleshoot.sh - Help developers troubleshoot pre-commit hook timeouts
# Usage: ./scripts/precommit-troubleshoot.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[precommit-troubleshoot]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[precommit-troubleshoot]${NC} ⚠️  $1"
}

error() {
    echo -e "${RED}[precommit-troubleshoot]${NC} ❌ $1"
}

success() {
    echo -e "${GREEN}[precommit-troubleshoot]${NC} ✅ $1"
}

echo "🔧 Pre-commit Hook Troubleshooting Guide"
echo "======================================="
echo ""

# Check if pre-commit is installed
if ! command -v pre-commit >/dev/null 2>&1; then
    error "pre-commit is not installed. Install it with: pip install pre-commit"
    exit 1
fi

# Check cache directory status
CACHE_DIR=".cache/coverage"
if [ -d "$CACHE_DIR" ]; then
    CACHE_SIZE=$(du -sh "$CACHE_DIR" 2>/dev/null | cut -f1)
    log "Coverage cache directory exists: $CACHE_SIZE"

    if [ -f "$CACHE_DIR/timeout.flag" ]; then
        TIMEOUT_AGE=$(( $(date +%s) - $(stat -f %m "$CACHE_DIR/timeout.flag" 2>/dev/null || stat -c %Y "$CACHE_DIR/timeout.flag" 2>/dev/null || echo 0) ))
        warn "Timeout flag detected (${TIMEOUT_AGE}s old). Previous coverage check timed out."

        echo ""
        echo "🔄 Recovery Options:"
        echo "1. Clear timeout flag: rm '$CACHE_DIR/timeout.flag'"
        echo "2. Clear entire cache: rm -rf '$CACHE_DIR'"
        echo "3. Run coverage manually: make test-coverage"
    fi
else
    log "No coverage cache directory found. First run will be slower."
fi

echo ""
echo "⚡ Performance Optimization Tips:"
echo "================================"
echo ""

echo "1. 📊 Coverage Check Optimization:"
echo "   - Current timeout: 8 minutes"
echo "   - Skips documentation-only changes automatically"
echo "   - Uses aggressive caching based on file content hashes"
echo "   - Run 'make test-coverage' manually to pre-warm cache"
echo ""

echo "2. 🔍 Linting Optimization:"
echo "   - golangci-lint timeout: 4 minutes"
echo "   - Skips integration/e2e directories for speed"
echo "   - Uses --fast flag for quicker analysis"
echo "   - Run 'golangci-lint run --fast' manually to test"
echo ""

echo "3. 🏗️  Build Check Optimization:"
echo "   - Build timeout: 2 minutes"
echo "   - Only verifies compilation, doesn't run tests"
echo "   - Run 'go build ./...' manually to test"
echo ""

echo "4. 🧪 Test Optimization:"
echo "   - Fast tokenizer tests: 1 minute timeout"
echo "   - Parallel execution enabled"
echo "   - Only runs critical performance regression tests"
echo ""

echo "🚀 Quick Commands:"
echo "=================="
echo ""
echo "Clear all caches:           rm -rf .cache/"
echo "Skip hooks for emergency:   git commit --no-verify"
echo "Run hooks manually:         pre-commit run --all-files"
echo "Update hook repos:          pre-commit autoupdate"
echo "Test coverage manually:     make test-coverage"
echo "Test specific hook:         pre-commit run go-coverage-check"
echo ""

echo "⏱️  Timeout Summary:"
echo "==================="
echo "- golangci-lint:      4 minutes"
echo "- go-build-check:     2 minutes"
echo "- tokenizer-tests:    1 minute"
echo "- go-coverage-check:  8 minutes"
echo "- Total max time:     ~15 minutes (if all hooks run from scratch)"
echo ""

echo "💡 For documentation-only changes, coverage check is automatically skipped!"
echo ""

# Check if user wants to run a specific troubleshooting action
if [ "$1" = "--clear-cache" ]; then
    log "Clearing coverage cache..."
    rm -rf "$CACHE_DIR"
    success "Coverage cache cleared. Next run will regenerate from scratch."
elif [ "$1" = "--clear-timeout" ]; then
    if [ -f "$CACHE_DIR/timeout.flag" ]; then
        rm -f "$CACHE_DIR/timeout.flag"
        success "Timeout flag cleared."
    else
        log "No timeout flag found."
    fi
elif [ "$1" = "--test-hooks" ]; then
    log "Testing individual hooks (dry run)..."
    echo ""

    echo "Testing go-fmt..."
    timeout 30s pre-commit run go-fmt --all-files || warn "go-fmt had issues"

    echo "Testing go-build-check..."
    timeout 2m pre-commit run go-build-check --all-files || warn "go-build-check had issues"

    echo "Testing golangci-lint..."
    timeout 4m pre-commit run golangci-lint --all-files || warn "golangci-lint had issues"

    success "Hook testing completed."
elif [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  --clear-cache     Clear coverage cache"
    echo "  --clear-timeout   Clear timeout flag"
    echo "  --test-hooks      Test individual hooks"
    echo "  --help, -h        Show this help"
else
    echo "Run '$0 --help' for additional troubleshooting options."
fi
