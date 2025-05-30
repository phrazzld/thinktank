#!/bin/bash
# Script to verify developer environment is properly configured
# This can be run at any time to ensure hooks and tools are set up correctly

set -euo pipefail

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BOLD}$1${NC}"
    echo "=================================================="
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

ERRORS=0
WARNINGS=0

print_header "Developer Environment Verification"

# Check Go installation
print_header "Checking Go Installation"
if command -v go >/dev/null 2>&1; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go is installed: $GO_VERSION"
else
    print_error "Go is not installed"
    echo "  Install from: https://golang.org/dl/"
    ERRORS=$((ERRORS + 1))
fi

# Check pre-commit installation
print_header "Checking Pre-commit Installation"
if command -v pre-commit >/dev/null 2>&1; then
    PRECOMMIT_VERSION=$(pre-commit --version)
    print_success "pre-commit is installed: $PRECOMMIT_VERSION"

    # Check version
    MIN_VERSION="3.0.0"
    CURRENT_VERSION=$(pre-commit --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    if [ "$(printf '%s\n' "$MIN_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]; then
        print_warning "pre-commit version $CURRENT_VERSION is older than recommended $MIN_VERSION"
        echo "  Upgrade with: pip install --upgrade pre-commit"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    print_error "pre-commit is not installed"
    echo "  Install with: pip install pre-commit"
    ERRORS=$((ERRORS + 1))
fi

# Check Git hooks
print_header "Checking Git Hooks Installation"
REQUIRED_HOOKS=("pre-commit" "commit-msg" "pre-push" "post-commit")
MISSING_HOOKS=()
NON_EXEC_HOOKS=()

for hook in "${REQUIRED_HOOKS[@]}"; do
    if [ -f ".git/hooks/$hook" ]; then
        if [ -x ".git/hooks/$hook" ]; then
            print_success "Hook installed and executable: $hook"
        else
            print_warning "Hook installed but not executable: $hook"
            NON_EXEC_HOOKS+=("$hook")
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        print_error "Missing hook: $hook"
        MISSING_HOOKS+=("$hook")
        ERRORS=$((ERRORS + 1))
    fi
done

# Check pre-commit configuration
print_header "Checking Pre-commit Configuration"
if [ -f ".pre-commit-config.yaml" ]; then
    print_success "Pre-commit configuration file exists"

    # Validate with pre-commit if available
    if command -v pre-commit >/dev/null 2>&1; then
        if pre-commit validate-config >/dev/null 2>&1; then
            print_success "Pre-commit configuration is valid"
        else
            print_error "Pre-commit configuration is invalid"
            echo "  Run: pre-commit validate-config"
            ERRORS=$((ERRORS + 1))
        fi
    fi
else
    print_error "Missing .pre-commit-config.yaml"
    ERRORS=$((ERRORS + 1))
fi

# Check development tools
print_header "Checking Development Tools"
TOOLS=(
    "golangci-lint:golangci-lint --version"
    "svu:svu --version"
    "git-chglog:git-chglog --version"
    "govulncheck:govulncheck --version || govulncheck --help"
    "glance:glance --version || echo 'glance installed'"
)

for tool_check in "${TOOLS[@]}"; do
    IFS=':' read -r tool cmd <<< "$tool_check"
    if command -v "$tool" >/dev/null 2>&1; then
        VERSION=$(eval "$cmd" 2>&1 | head -1 || echo "version unknown")
        print_success "$tool is installed: $VERSION"
    else
        if [ "$tool" = "glance" ]; then
            print_warning "$tool is not installed (required for post-commit hook)"
            echo "  Install with: go install github.com/phaedrus-dev/glance@latest"
            WARNINGS=$((WARNINGS + 1))
        else
            print_warning "$tool is not installed (optional but recommended)"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
done

# Check baseline commit file
print_header "Checking Commit Validation Configuration"
if [ -f ".git/BASELINE_COMMIT" ]; then
    BASELINE=$(cat .git/BASELINE_COMMIT)
    print_success "Baseline commit configured: ${BASELINE:0:7}"
else
    print_warning "No baseline commit configured"
    echo "  Baseline will be created on first validation"
    WARNINGS=$((WARNINGS + 1))
fi

# Summary
print_header "Environment Verification Summary"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    print_success "Your development environment is perfectly configured!"
    echo "You're ready to contribute to the project."
elif [ $ERRORS -eq 0 ]; then
    print_success "Your environment meets all requirements"
    print_warning "There are $WARNINGS warning(s) that you may want to address"
    echo "You can contribute, but consider fixing the warnings."
else
    print_error "Your environment has $ERRORS critical error(s)"
    [ $WARNINGS -gt 0 ] && print_warning "Additionally, there are $WARNINGS warning(s)"
    echo ""
    echo "❌ YOUR ENVIRONMENT IS NOT READY FOR DEVELOPMENT"
    echo ""
    echo "Please fix the critical errors before attempting to contribute."
    echo "Run './scripts/setup.sh' to automatically fix most issues."
fi

# Provide fix commands if there are issues
if [ ${#MISSING_HOOKS[@]} -gt 0 ]; then
    echo ""
    echo "To install missing hooks, run:"
    echo "  make hooks"
    echo "OR"
    echo "  ./scripts/setup.sh"
fi

if [ ${#NON_EXEC_HOOKS[@]} -gt 0 ]; then
    echo ""
    echo "To fix non-executable hooks, run:"
    for hook in "${NON_EXEC_HOOKS[@]}"; do
        echo "  chmod +x .git/hooks/$hook"
    done
fi

exit $ERRORS
