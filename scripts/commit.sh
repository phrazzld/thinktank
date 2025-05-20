#!/bin/bash
# commit.sh - Run Commitizen for guided commit message creation
# This script helps create conventional commit messages through an interactive interface

# ANSI color codes
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if node/npm is installed
if ! command -v npm &> /dev/null; then
    echo -e "${YELLOW}Node.js/npm is required to run Commitizen.${NC}"
    echo "Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check if dependencies are installed
if [ ! -d "node_modules/commitizen" ]; then
    echo -e "${BLUE}Installing Commitizen dependencies...${NC}"
    npm install
fi

# Print header
echo -e "${BOLD}Thinktank - Guided Commit Creation${NC}"
echo -e "${GREEN}This tool will help you create a properly formatted conventional commit.${NC}"
echo -e "Baseline commit policy is respected - only affects new commits.\n"

# Check if we have staged changes
if [ -z "$(git diff --cached)" ]; then
    echo -e "${YELLOW}No staged changes detected.${NC}"
    read -p "Would you like to stage all changes? (y/n): " stage_all
    if [[ $stage_all == "y" || $stage_all == "Y" ]]; then
        git add .
        echo "All changes staged."
    else
        echo -e "${YELLOW}Please stage your changes first with ${BOLD}git add <files>${NC}"
        echo "Then run this script again to commit them."
        exit 1
    fi
fi

# Run Commitizen
echo -e "\n${BLUE}Starting interactive commit process...${NC}"
echo -e "${YELLOW}You will be prompted to select a commit type and enter details.${NC}\n"

# Use npx to run commitizen
npx cz

# Check if commit was successful
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Commit successfully created!${NC}"
    echo "The commit message follows the Conventional Commits standard."
    echo -e "${BOLD}To push your changes:${NC} git push"
else
    echo -e "\n${YELLOW}Commit process was aborted or failed.${NC}"
    echo "Your changes are still staged and ready to be committed."
fi
