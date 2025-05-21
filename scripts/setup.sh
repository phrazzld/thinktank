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
print_header "Installing Git hooks (MANDATORY)"
if command -v pre-commit >/dev/null 2>&1; then
    echo "Git hooks are MANDATORY for this project. Installing all required hooks..."

    # Install pre-commit hooks
    echo "Installing pre-commit hooks..."
    pre-commit install --install-hooks

    if [ $? -eq 0 ]; then
        print_success "Pre-commit hooks installed successfully"
    else
        print_error "Failed to install pre-commit hooks"
        echo "Please try running 'pre-commit install --install-hooks' manually"
    fi

    # Install commit-msg hooks for conventional commit validation
    echo "Installing commit-msg hooks..."
    pre-commit install --hook-type commit-msg

    if [ $? -eq 0 ]; then
        print_success "Commit-msg hooks installed successfully"
    else
        print_error "Failed to install commit-msg hooks"
        echo "Please try running 'pre-commit install --hook-type commit-msg' manually"
    fi

    # Set up baseline-aware commit validation
    echo "Setting up baseline-aware commit validation..."
    if [ -f "./scripts/setup-commitlint.sh" ]; then
        ./scripts/setup-commitlint.sh
        if [ $? -eq 0 ]; then
            print_success "Baseline-aware commit validation configured successfully"
            echo "Only commits made after the baseline commit (May 18, 2025) will be validated"
            echo "This matches the CI workflow validation approach"
        else
            print_error "Failed to configure baseline-aware commit validation"
            echo "Commit message validation will still work but without baseline awareness"
        fi
    else
        print_warning "setup-commitlint.sh not found, skipping baseline configuration"
        echo "Commit message validation will work but without baseline awareness"
    fi

    # Install pre-push hooks
    echo "Installing pre-push hooks..."
    pre-commit install --hook-type pre-push

    if [ $? -eq 0 ]; then
        print_success "Pre-push hooks installed successfully"
    else
        print_error "Failed to install pre-push hooks"
        echo "Please try running 'pre-commit install --hook-type pre-push' manually"
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
    if [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/post-commit" ] && [ -f ".git/hooks/commit-msg" ] && [ -f ".git/hooks/pre-push" ]; then
        print_success "All Git hooks verified in .git/hooks/"
        print_success "Code formatting, commit message validation, and pre-push validation are now active"
    else
        print_warning "Some hook files not found in .git/hooks/ directory"
        echo "There may be an issue with your git configuration"
        echo "Expected hooks: pre-commit, commit-msg, pre-push, post-commit"
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
    pre-commit install --install-hooks

    echo "Installing commit-msg hooks..."
    pre-commit install --hook-type commit-msg

    # Set up baseline-aware commit validation on reinstall
    if [ -f "./scripts/setup-commitlint.sh" ]; then
        echo "Reconfiguring baseline-aware commit validation..."
        ./scripts/setup-commitlint.sh
    fi

    echo "Installing pre-push hooks..."
    pre-commit install --hook-type pre-push

    echo "Installing post-commit hooks..."
    pre-commit install --hook-type post-commit

    # Verify installation
    if [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/post-commit" ] && [ -f ".git/hooks/commit-msg" ] && [ -f ".git/hooks/pre-push" ]; then
        print_success "All hooks reinstalled successfully"
    else
        print_error "Hook reinstallation may have failed - some hook files not found"
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

# Set up commit template
print_header "Setting up Git commit template"
if [ -f "./scripts/setup-commit-template.sh" ]; then
    echo "Configuring Git commit template..."
    ./scripts/setup-commit-template.sh
    if [ $? -eq 0 ]; then
        print_success "Git commit template configured successfully"
        echo "The template will be used for all new commits in this repository"
    else
        print_warning "Failed to configure Git commit template"
        echo "You can set it up manually later by running:"
        echo "  ./scripts/setup-commit-template.sh"
    fi
else
    print_warning "setup-commit-template.sh not found, skipping commit template setup"
    echo "You can set it up manually later if it becomes available"
fi

# Set up Commitizen for guided commit creation
print_header "Setting up Commitizen (optional)"
if [ -f "./package.json" ]; then
    if command -v npm >/dev/null 2>&1; then
        echo "Commitizen is available for guided commit creation."
        echo "Would you like to install the Node.js dependencies for Commitizen? (y/n)"
        read -r INSTALL_COMMITIZEN

        if [[ $INSTALL_COMMITIZEN =~ ^[Yy]$ ]] || [[ -z $INSTALL_COMMITIZEN ]]; then
            echo "Installing Commitizen dependencies..."
            npm install

            if [ $? -eq 0 ]; then
                print_success "Commitizen dependencies installed successfully"
                echo "You can now use guided commit creation with:"
                echo "  ./scripts/commit.sh  or  make commit  or  npm run commit"
            else
                print_error "Failed to install Commitizen dependencies"
                echo "You can try installing them manually with 'npm install'"
            fi
        else
            print_warning "Commitizen dependency installation skipped"
            echo "You can install them later with 'npm install' if you want to use guided commit creation"
        fi
    else
        print_warning "Node.js/npm is not installed"
        echo "If you want to use Commitizen for guided commit creation, you'll need to:"
        echo "1. Install Node.js from https://nodejs.org/"
        echo "2. Run 'npm install' in this directory"
    fi
else
    print_warning "package.json not found, skipping Commitizen setup"
    echo "Commitizen configuration may not be properly set up in this repository"
fi

print_header "Setup Complete"
echo "Your environment is now ready for development."
echo "For more information, see the project README.md and CONTRIBUTING.md"
echo ""
echo "If you encounter issues with Git hooks, you can run this script again"
echo "or follow the manual reinstallation steps in hooks/README.md"
