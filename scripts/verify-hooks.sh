#!/bin/bash

# Script to verify pre-commit hooks are properly installed
# Usage: ./scripts/verify-hooks.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "Checking pre-commit hook installation..."
echo

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not in a git repository root directory${NC}"
    echo "Please run this script from the repository root."
    exit 1
fi

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo -e "${RED}✗ pre-commit is not installed${NC}"
    echo
    echo "To install pre-commit, run one of the following:"
    echo "  pip install pre-commit"
    echo "  brew install pre-commit  # on macOS"
    echo
    echo "For more information, see CONTRIBUTING.md"
    exit 1
else
    echo -e "${GREEN}✓ pre-commit is installed${NC}"
fi

# Check if pre-commit hooks are installed
if [ -f ".git/hooks/pre-commit" ] && grep -q "pre-commit" ".git/hooks/pre-commit"; then
    echo -e "${GREEN}✓ pre-commit hooks are installed${NC}"
else
    echo -e "${YELLOW}⚠ pre-commit hooks are not installed${NC}"
    echo
    echo "To install pre-commit hooks, run:"
    echo "  pre-commit install"
    echo
fi

# Check if commit-msg hook is installed
if [ -f ".git/hooks/commit-msg" ] && grep -q "pre-commit" ".git/hooks/commit-msg"; then
    echo -e "${GREEN}✓ commit-msg hook is installed${NC}"
else
    echo -e "${YELLOW}⚠ commit-msg hook is not installed${NC}"
    echo
    echo "To install the commit-msg hook, run:"
    echo "  pre-commit install --hook-type commit-msg"
    echo
fi

# Check for custom hooks path
HOOKS_PATH=$(git config core.hooksPath || echo "")
if [ -n "$HOOKS_PATH" ]; then
    echo -e "${BLUE}ℹ Custom hooks path configured: $HOOKS_PATH${NC}"
    # Check if custom hooks exist
    if [ -d "$HOOKS_PATH" ]; then
        echo -e "${GREEN}✓ Custom hooks directory exists${NC}"
        # Check for commit-msg hook in custom path
        if [ -f "$HOOKS_PATH/commit-msg" ]; then
            echo -e "${GREEN}✓ commit-msg hook exists in custom path${NC}"
        else
            echo -e "${YELLOW}⚠ commit-msg hook not found in custom path${NC}"
        fi
    else
        echo -e "${RED}✗ Custom hooks directory not found: $HOOKS_PATH${NC}"
    fi
fi

# Check pre-commit config exists
if [ -f ".pre-commit-config.yaml" ] || [ -f ".pre-commit-config.yml" ]; then
    echo -e "${GREEN}✓ pre-commit configuration file exists${NC}"
else
    echo -e "${RED}✗ pre-commit configuration file not found${NC}"
    echo "Expected: .pre-commit-config.yaml or .pre-commit-config.yml"
fi

# Check commitlint config exists
if [ -f ".commitlintrc.yml" ] || [ -f ".commitlintrc.yaml" ] || [ -f ".commitlintrc.json" ]; then
    echo -e "${GREEN}✓ commitlint configuration file exists${NC}"
else
    echo -e "${RED}✗ commitlint configuration file not found${NC}"
    echo "Expected: .commitlintrc.yml, .commitlintrc.yaml, or .commitlintrc.json"
fi

echo
# Summary based on custom hooks path
if [ -n "$HOOKS_PATH" ]; then
    if [ -f "$HOOKS_PATH/commit-msg" ]; then
        echo -e "${GREEN}Git hooks are properly configured with custom path!${NC}"
    else
        echo -e "${YELLOW}Custom hooks path is set but some hooks may be missing.${NC}"
        echo "Check the documentation in CONTRIBUTING.md for custom hook setup."
    fi
elif [ -f ".git/hooks/pre-commit" ] && [ -f ".git/hooks/commit-msg" ]; then
    echo -e "${GREEN}All pre-commit hooks are properly installed!${NC}"
else
    echo -e "${YELLOW}Some pre-commit hooks are missing.${NC}"
    echo "Please follow the instructions above or refer to CONTRIBUTING.md"
fi
