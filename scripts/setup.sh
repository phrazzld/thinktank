#!/bin/bash
# setup.sh - Project setup script for architect
# This script ensures necessary dependencies are installed and sets up development tools

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BOLD}$1${NC}"
    echo -e "=================================================="
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

print_header "Architect Project Setup"
echo "This script will help set up your development environment."

# Check for Go installation
print_header "Checking Go installation"
if command -v go >/dev/null 2>&1; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go is installed: $GO_VERSION"
else
    print_error "Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check for pre-commit installation
print_header "Checking pre-commit framework installation"
if command -v pre-commit >/dev/null 2>&1; then
    PRECOMMIT_VERSION=$(pre-commit --version)
    print_success "pre-commit is installed: $PRECOMMIT_VERSION"
else
    print_warning "pre-commit framework is not installed"
    echo "The pre-commit framework is required for our Git hooks."
    echo "Would you like to install pre-commit now? (y/n)"
    read -r INSTALL_PRECOMMIT

    if [[ $INSTALL_PRECOMMIT =~ ^[Yy]$ ]]; then
        echo "Installing pre-commit..."

        if command -v pip >/dev/null 2>&1; then
            pip install pre-commit
        elif command -v pip3 >/dev/null 2>&1; then
            pip3 install pre-commit
        elif command -v brew >/dev/null 2>&1; then
            brew install pre-commit
        else
            print_error "Could not find pip, pip3, or brew to install pre-commit"
            echo "Please install pre-commit manually: https://pre-commit.com/#install"
            exit 1
        fi

        if command -v pre-commit >/dev/null 2>&1; then
            print_success "pre-commit installed successfully"
        else
            print_error "Failed to install pre-commit"
            echo "Please install pre-commit manually: https://pre-commit.com/#install"
            exit 1
        fi
    else
        print_warning "Skipping pre-commit installation"
        echo "Note: pre-commit is required for development. Please install it later."
        echo "See hooks/README.md for installation instructions."
    fi
fi

# Install git hooks
print_header "Installing Git hooks"
if command -v pre-commit >/dev/null 2>&1; then
    pre-commit install
    print_success "Git hooks installed successfully"
else
    print_warning "Skipping Git hook installation (pre-commit not available)"
    echo "Please install pre-commit and run 'pre-commit install' later."
fi

# Check for golangci-lint
print_header "Checking golangci-lint installation"
if command -v golangci-lint >/dev/null 2>&1; then
    GOLANGCI_VERSION=$(golangci-lint --version | head -n 1)
    print_success "golangci-lint is installed: $GOLANGCI_VERSION"
else
    print_warning "golangci-lint is not installed"
    echo "golangci-lint is recommended for development."
    echo "Install instructions: https://golangci-lint.run/usage/install/"
fi

# Building the project
print_header "Building the project"
go build -v
if [ $? -eq 0 ]; then
    print_success "Build successful"
else
    print_error "Build failed"
fi

print_header "Setup Complete"
echo "Your environment is now ready for development."
echo "For more information, see the project README.md and hooks/README.md"
