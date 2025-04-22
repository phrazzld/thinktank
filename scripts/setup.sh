#!/bin/bash
# setup.sh - Project setup script for thinktank
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

print_header "Thinktank Project Setup"
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
    echo "The pre-commit framework is REQUIRED for this project's development workflow."
    echo "Without pre-commit, commits will not be properly validated, which may lead to CI failures."

    # Offer automatic installation with stronger messaging
    echo "Would you like to install pre-commit automatically? (y/n) [y recommended]"
    read -r INSTALL_PRECOMMIT

    if [[ $INSTALL_PRECOMMIT =~ ^[Yy]$ ]] || [[ -z $INSTALL_PRECOMMIT ]]; then
        echo "Installing pre-commit..."

        if command -v pip >/dev/null 2>&1; then
            pip install pre-commit
        elif command -v pip3 >/dev/null 2>&1; then
            pip3 install pre-commit
        elif command -v brew >/dev/null 2>&1; then
            brew install pre-commit
        else
            print_error "Could not find pip, pip3, or brew to install pre-commit"
            echo "Please install pre-commit manually following these steps:"
            echo "  1. Visit https://pre-commit.com/#install"
            echo "  2. Follow the installation instructions for your platform"
            echo "  3. Run this setup script again"
            exit 1
        fi

        if command -v pre-commit >/dev/null 2>&1; then
            print_success "pre-commit installed successfully"
        else
            print_error "Failed to install pre-commit"
            echo "Please install pre-commit manually following these steps:"
            echo "  1. Visit https://pre-commit.com/#install"
            echo "  2. Follow the installation instructions for your platform"
            echo "  3. Run this setup script again"
            exit 1
        fi
    else
        print_error "pre-commit installation skipped, but it is required for development"
        echo "Please be aware that:"
        echo "  - Your commits may not meet project standards"
        echo "  - CI checks may fail on your pull requests"
        echo "  - Other developers may need to fix issues in your code"
        echo ""
        echo "To install pre-commit later, see hooks/README.md or visit https://pre-commit.com/#install"

        # Ask if they want to continue without pre-commit
        echo "Do you want to continue setup without pre-commit? (y/n) [n recommended]"
        read -r CONTINUE_WITHOUT_PRECOMMIT

        if [[ ! $CONTINUE_WITHOUT_PRECOMMIT =~ ^[Yy]$ ]]; then
            print_error "Setup aborted. Please install pre-commit and run setup again."
            exit 1
        fi

        print_warning "Continuing setup without pre-commit (not recommended)"
    fi
fi

# Install git hooks
print_header "Installing Git hooks"
if command -v pre-commit >/dev/null 2>&1; then
    # Install pre-commit hooks
    echo "Installing pre-commit hooks..."
    pre-commit install

    if [ $? -eq 0 ]; then
        print_success "Pre-commit hooks installed successfully"
    else
        print_error "Failed to install pre-commit hooks"
        echo "Please try running 'pre-commit install' manually"
    fi

    # Install post-commit hooks
    echo "Installing post-commit hooks..."
    pre-commit install --hook-type post-commit

    if [ $? -eq 0 ]; then
        print_success "Post-commit hooks installed successfully"
    else
        print_error "Failed to install post-commit hooks"
        echo "Please try running 'pre-commit install --hook-type post-commit' manually"
    fi

    # Verify hook installation
    if [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/post-commit" ]; then
        print_success "Git hooks verified in .git/hooks/"
    else
        print_warning "Hook files not found in .git/hooks/ directory"
        echo "There may be an issue with your git configuration"
    fi

    # If glance is not installed, warn about post-commit hook potentially failing
    if ! command -v glance >/dev/null 2>&1; then
        print_warning "Post-commit hook requires glance, which is not installed"
        echo "The post-commit hook will fail until glance is installed"
    fi
else
    print_warning "Skipping Git hook installation (pre-commit not available)"
    echo "Please install pre-commit and run the following commands later:"
    echo "  pre-commit install"
    echo "  pre-commit install --hook-type post-commit"
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

# Check for glance installation
print_header "Checking glance installation"
if command -v glance >/dev/null 2>&1; then
    GLANCE_VERSION=$(glance --version 2>&1 || echo "version unknown")
    print_success "glance is installed: $GLANCE_VERSION"
else
    print_warning "glance is not installed"
    echo "glance is required for generating directory overview documentation."
    echo "It is used by the post-commit hook to automatically update documentation."
    echo ""
    echo "Would you like to install glance automatically? (y/n) [y recommended]"
    read -r INSTALL_GLANCE

    if [[ $INSTALL_GLANCE =~ ^[Yy]$ ]] || [[ -z $INSTALL_GLANCE ]]; then
        echo "Installing glance..."

        if command -v go >/dev/null 2>&1; then
            go install github.com/phaedrus-dev/glance@latest

            # Check if installation was successful
            if command -v glance >/dev/null 2>&1; then
                print_success "glance installed successfully"
            else
                print_warning "glance was installed but is not in PATH"
                echo "Please ensure your Go bin directory is in your PATH."
                echo "For example, add this to your shell profile:"
                echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
            fi
        else
            print_error "Could not install glance - Go is not available"
            echo "Please install glance manually:"
            echo "  go install github.com/phaedrus-dev/glance@latest"
        fi
    else
        print_warning "glance installation skipped, but it is required for directory documentation"
        echo "The post-commit hook will not function properly without glance."
        echo "To install glance later, run: go install github.com/phaedrus-dev/glance@latest"
    fi
fi

# Building the project
print_header "Building the project"
go build -v
if [ $? -eq 0 ]; then
    print_success "Build successful"
else
    print_error "Build failed"
fi

# Function to reinstall hooks if requested
reinstall_hooks() {
    print_header "Reinstalling Git Hooks"

    if ! command -v pre-commit >/dev/null 2>&1; then
        print_error "pre-commit is not installed"
        echo "Please install pre-commit first"
        return 1
    fi

    echo "Cleaning existing hooks..."
    pre-commit clean

    echo "Installing pre-commit hooks..."
    pre-commit install

    echo "Installing post-commit hooks..."
    pre-commit install --hook-type post-commit

    # Verify installation
    if [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/post-commit" ]; then
        print_success "Hooks reinstalled successfully"
    else
        print_error "Hook reinstallation may have failed - hook files not found"
    fi
}

# Offer to reinstall hooks
if [ -f ".git/hooks/pre-commit" ] || [ -f ".git/hooks/post-commit" ]; then
    echo ""
    echo "Would you like to reinstall all Git hooks? (y/n)"
    echo "This can help fix hook issues by ensuring the latest configuration is used."
    read -r REINSTALL_HOOKS

    if [[ $REINSTALL_HOOKS =~ ^[Yy]$ ]]; then
        reinstall_hooks
    fi
fi

print_header "Setup Complete"
echo "Your environment is now ready for development."
echo "For more information, see the project README.md and hooks/README.md"
echo ""
echo "If you encounter issues with Git hooks, you can run this script again"
echo "or follow the manual reinstallation steps in hooks/README.md"
