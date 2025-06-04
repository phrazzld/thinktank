#!/bin/bash
# validate-pr-commits.sh - Validate commits on current branch against conventional commits standard
#
# This script allows developers to validate their branch commit history locally
# before creating a PR or pushing, using the same validation rules as the CI workflow.
#
# USAGE:
#   ./scripts/validate-pr-commits.sh [base_branch]
#
# EXAMPLES:
#   ./scripts/validate-pr-commits.sh          # Validates against master
#   ./scripts/validate-pr-commits.sh main     # Validates against main branch
#
# PURPOSE:
#   - Enables local validation before pushing/creating PRs
#   - Uses identical validation logic as CI workflow
#   - Provides early feedback to prevent CI failures
#   - Reduces iteration time in development workflow

set -e

# Configuration
BASELINE_COMMIT="1300e4d675ac087783199f1e608409e6853e589f"
DEFAULT_BASE_BRANCH="master"

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_header() {
    echo -e "${BOLD}$1${NC}"
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

print_info() {
    echo -e "${BLUE}→ $1${NC}"
}

# Usage information
show_usage() {
    echo "USAGE: $0 [base_branch]"
    echo ""
    echo "Validates commit messages on the current branch against conventional commits standard."
    echo "Uses the same validation logic as the CI workflow to prevent CI failures."
    echo ""
    echo "ARGUMENTS:"
    echo "  base_branch    Base branch to compare against (default: master)"
    echo ""
    echo "EXAMPLES:"
    echo "  $0              # Validate against master branch"
    echo "  $0 main         # Validate against main branch"
    echo "  $0 develop      # Validate against develop branch"
    echo ""
    echo "NOTES:"
    echo "  - Only validates commits made after baseline (May 18, 2025)"
    echo "  - Uses Go-based validator (cmd/commitvalidate)"
    echo "  - Provides same validation as CI workflow"
}

# Parse arguments
BASE_BRANCH="${1:-$DEFAULT_BASE_BRANCH}"

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

# Print header
print_header "Branch Commit Validation - Conventional Commits"
echo "Base branch: ${BASE_BRANCH}"
echo "Baseline commit: ${BASELINE_COMMIT} (May 18, 2025)"
echo "Only commits made after the baseline will be validated."
echo "========================================================"

# Verify we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Verify base branch exists
if ! git rev-parse --verify "${BASE_BRANCH}" > /dev/null 2>&1; then
    print_error "Base branch '${BASE_BRANCH}' does not exist"
    print_info "Available branches:"
    git branch -a | grep -E "(master|main|develop)" | head -5
    exit 1
fi

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
print_info "Current branch: ${CURRENT_BRANCH}"

# Check if we're on the base branch
if [[ "$CURRENT_BRANCH" == "$BASE_BRANCH" ]]; then
    print_warning "You are currently on the base branch '${BASE_BRANCH}'"
    print_info "This script is intended to validate feature branches before merging"
    print_info "Continuing with validation of recent commits..."
fi

# Get the merge base between current branch and base branch
MERGE_BASE=$(git merge-base "${BASE_BRANCH}" HEAD)
print_info "Merge base: ${MERGE_BASE}"

# Determine commits to validate (same logic as CI workflow)
HEAD_SHA=$(git rev-parse HEAD)
BASE_SHA=$(git rev-parse "${BASE_BRANCH}")

print_info "Comparing ${BASE_SHA:0:8}..${HEAD_SHA:0:8}"

# Get commits to validate (from base branch to current HEAD)
COMMITS=$(git rev-list "${BASE_SHA}..${HEAD_SHA}")

if [ -z "$COMMITS" ]; then
    print_warning "No commits found between ${BASE_BRANCH} and current branch"
    print_success "Validation completed: No commits to validate"
    exit 0
fi

# Count commits
COMMIT_COUNT=$(echo "$COMMITS" | wc -l | tr -d ' ')
print_info "Found ${COMMIT_COUNT} commit(s) to validate"

# Validate each commit individually (same logic as CI)
FAILED=0
VALIDATED_COUNT=0

for commit in $COMMITS; do
    # Check if commit is after baseline (same as CI logic)
    if git merge-base --is-ancestor "${BASELINE_COMMIT}" "${commit}" && [ "${commit}" != "${BASELINE_COMMIT}" ]; then
        VALIDATED_COUNT=$((VALIDATED_COUNT + 1))

        # Get commit details
        COMMIT_SHORT=$(git log --format=%h -n 1 "${commit}")
        COMMIT_DATE=$(git log --format=%ci -n 1 "${commit}")
        COMMIT_AUTHOR=$(git log --format="%an" -n 1 "${commit}")
        COMMIT_SUBJECT=$(git log --format=%s -n 1 "${commit}")

        echo ""
        print_info "Validating commit ${COMMIT_SHORT}: ${COMMIT_SUBJECT}"
        echo "  Date: ${COMMIT_DATE}"
        echo "  Author: ${COMMIT_AUTHOR}"

        # Validate using the same command as CI
        COMMIT_MSG=$(git show -s --format=%B "${commit}")
        if echo "${COMMIT_MSG}" | go run ./cmd/commitvalidate --stdin 2>/dev/null; then
            print_success "Valid conventional commit"
        else
            print_error "Invalid commit format"
            echo ""
            echo "Commit message:"
            echo "---------------"
            echo "${COMMIT_MSG}"
            echo "---------------"
            echo ""
            echo "Fix suggestions:"
            echo "  1. Format: <type>[optional scope]: <description>"
            echo "  2. Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf"
            echo "  3. Use 'git commit --amend' to fix the most recent commit"
            echo "  4. Use 'git rebase -i ${BASE_BRANCH}' to fix older commits"
            echo ""
            FAILED=1
        fi
    else
        COMMIT_SHORT=$(git log --format=%h -n 1 "${commit}")
        print_info "Skipping commit ${COMMIT_SHORT} (before baseline)"
    fi
done

# Summary
echo ""
echo "========================================================"
if [ $VALIDATED_COUNT -eq 0 ]; then
    print_warning "No commits validated (all commits were before baseline)"
    print_info "This means all commits on this branch were made before May 18, 2025"
    print_success "Validation completed: No issues found"
elif [ $FAILED -eq 1 ]; then
    print_error "Validation failed: ${VALIDATED_COUNT} commit(s) checked, some have invalid messages"
    echo ""
    echo "Next steps:"
    echo "  1. Fix the commit messages using the suggestions above"
    echo "  2. Re-run this script to verify fixes"
    echo "  3. For detailed guidance, see: docs/conventional-commits.md"
    echo ""
    print_info "This script uses the same validation as CI - fixing these issues will prevent CI failures"
    exit 1
else
    print_success "All ${VALIDATED_COUNT} commit(s) validated successfully!"
    print_info "Your branch is ready for push/PR - CI validation should pass"
fi

exit 0
