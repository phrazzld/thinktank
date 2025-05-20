#!/bin/bash
# setup-commit-template.sh - Configure Git to use the repository's commit template
# This script configures Git to use the repository's conventional commit template

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Path to commit template
TEMPLATE_PATH=".github/commit-template.txt"
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

echo -e "${BOLD}Setting up Git commit template${NC}"

# Check if we're in a git repository
if [ -z "$REPO_ROOT" ]; then
    print_error "Not in a git repository"
    echo "This script must be run from within a git repository."
    exit 1
fi

# Check if template file exists
TEMPLATE_FULL_PATH="$REPO_ROOT/$TEMPLATE_PATH"
if [ ! -f "$TEMPLATE_FULL_PATH" ]; then
    print_error "Commit template not found at $TEMPLATE_FULL_PATH"
    echo "Make sure the template file exists in the repository."
    exit 1
fi

# Configure git to use the template
echo "Configuring Git to use commit template..."
git config commit.template "$TEMPLATE_PATH"

if [ $? -eq 0 ]; then
    print_success "Git commit template configured successfully"
    echo "The template will be used for all new commits in this repository."
    echo "To use it, simply run 'git commit' (without -m flag) to open your editor"
    echo "with the template pre-filled."
else
    print_error "Failed to configure Git commit template"
    echo "Try running the following command manually:"
    echo "  git config commit.template $TEMPLATE_PATH"
    exit 1
fi

# Show template content preview
echo ""
echo -e "${BOLD}Commit Template Preview:${NC}"
echo "---------------------------------------------"
head -n 20 "$TEMPLATE_FULL_PATH" | sed 's/^/  /'
echo "  ..."
echo "---------------------------------------------"
echo ""
echo "Template includes:"
echo "- Conventional commit format examples"
echo "- Type definitions and use cases"
echo "- Guidelines for commit message body and footer"
echo "- Breaking change notation examples"
echo ""
print_success "Setup complete"
