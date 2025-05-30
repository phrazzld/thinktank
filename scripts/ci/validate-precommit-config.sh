#!/bin/bash
# CI script to validate pre-commit configuration
# This ensures the .pre-commit-config.yaml is valid and all required hooks are defined

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

print_header "Pre-commit Configuration Validation"

# Check if .pre-commit-config.yaml exists
if [ ! -f ".pre-commit-config.yaml" ]; then
    print_error ".pre-commit-config.yaml not found!"
    echo "This file is MANDATORY for project development."
    exit 1
fi

print_success ".pre-commit-config.yaml found"

# Validate YAML syntax
echo "Validating YAML syntax..."
if command -v python3 >/dev/null 2>&1; then
    if python3 -c "import yaml; yaml.safe_load(open('.pre-commit-config.yaml'))" 2>/dev/null; then
        print_success "YAML syntax is valid"
    else
        print_error "Invalid YAML syntax in .pre-commit-config.yaml"
        exit 1
    fi
else
    print_warning "Python3 not available, skipping YAML validation"
fi

# Check for required hook stages
print_header "Checking Required Hook Stages"

REQUIRED_STAGES=(
    "pre-commit"
    "commit-msg"
    "pre-push"
    "post-commit"
)

MISSING_STAGES=()

for stage in "${REQUIRED_STAGES[@]}"; do
    if grep -q "stages:.*\[$stage\]" .pre-commit-config.yaml || grep -q "stages:.*\[.*$stage.*\]" .pre-commit-config.yaml; then
        print_success "Found hook for stage: $stage"
    else
        MISSING_STAGES+=("$stage")
        print_warning "No hook found for stage: $stage"
    fi
done

# Check for specific required hooks
print_header "Checking Required Hooks"

REQUIRED_HOOKS=(
    "trailing-whitespace"
    "end-of-file-fixer"
    "golangci-lint"
    "go-fmt"
    "conventional-commit-check"
    "conventional-commits-push-check"
    "run-glance"
)

MISSING_HOOKS=()

for hook in "${REQUIRED_HOOKS[@]}"; do
    if grep -q "id: $hook" .pre-commit-config.yaml; then
        print_success "Found required hook: $hook"
    else
        MISSING_HOOKS+=("$hook")
        print_error "Missing required hook: $hook"
    fi
done

# Summary
print_header "Validation Summary"

ERRORS=0

if [ ${#MISSING_STAGES[@]} -gt 0 ]; then
    print_error "Missing hook stages: ${MISSING_STAGES[*]}"
    ERRORS=$((ERRORS + 1))
fi

if [ ${#MISSING_HOOKS[@]} -gt 0 ]; then
    print_error "Missing required hooks: ${MISSING_HOOKS[*]}"
    ERRORS=$((ERRORS + 1))
fi

# Check if pre-commit is available in CI
if command -v pre-commit >/dev/null 2>&1; then
    print_success "pre-commit CLI is available"

    # Validate configuration with pre-commit
    echo "Running pre-commit configuration validation..."
    if pre-commit validate-config; then
        print_success "pre-commit configuration is valid"
    else
        print_error "pre-commit configuration validation failed"
        ERRORS=$((ERRORS + 1))
    fi
else
    print_warning "pre-commit CLI not available in CI, skipping advanced validation"
fi

if [ $ERRORS -eq 0 ]; then
    print_success "All pre-commit configuration checks passed!"
    exit 0
else
    print_error "Pre-commit configuration validation failed with $ERRORS error(s)"
    echo ""
    echo "Pre-commit hooks are MANDATORY for this project."
    echo "Please ensure .pre-commit-config.yaml includes all required hooks."
    exit 1
fi
