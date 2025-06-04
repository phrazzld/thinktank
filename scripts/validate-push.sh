#!/bin/bash
# validate-push.sh - Validate commits being pushed against conventional commits standard
# Only validates commits made after the baseline commit
#
# This script is used by the pre-push hook in .pre-commit-config.yaml
# It reads stdin for the pre-push hook data and validates the commits in the push range

# The baseline commit SHA when conventional commits were established
BASELINE_COMMIT="1300e4d675ac087783199f1e608409e6853e589f"

# Format text
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Print header
echo -e "${BOLD}Pre-Push Validation - Conventional Commits${NC}"
echo "Baseline commit: ${BASELINE_COMMIT} (May 18, 2025)"
echo "Only commits made after the baseline will be validated."
echo "--------------------------------------------------------"

# Verify Go validator is available
if ! go run ./cmd/commitvalidate --help >/dev/null 2>&1; then
    echo -e "${RED}Error: Go commit validator is not available${NC}"
    echo "Ensure you're in the project root and Go is installed"
    exit 1
fi

# Read standard input (pre-push hook data)
# Format of stdin for pre-push: <local_ref> <local_sha> <remote_ref> <remote_sha>
while read local_ref local_sha remote_ref remote_sha
do
    # Skip if pushing branch deletion
    if [[ "$local_sha" == "0000000000000000000000000000000000000000" ]]; then
        echo -e "${YELLOW}Branch deletion detected. Skipping validation.${NC}"
        continue
    fi

    # Determine the range of commits to check
    # If there's a remote SHA, check from that to our local SHA
    # If it's a new branch, use the baseline commit
    if [[ "$remote_sha" == "0000000000000000000000000000000000000000" ]]; then
        # New branch, check all commits after baseline
        RANGE="${BASELINE_COMMIT}..${local_sha}"
        echo "New branch detected. Checking commits from baseline: ${RANGE}"
    else
        # Existing branch, check only commits we're pushing (that are after baseline)
        # Use the more recent of remote_sha and baseline
        # First check if baseline is an ancestor of remote_sha
        if git merge-base --is-ancestor ${BASELINE_COMMIT} ${remote_sha} 2>/dev/null; then
            # Baseline is an ancestor, use remote_sha as the start
            RANGE="${remote_sha}..${local_sha}"
            echo "Pushing to existing branch. Checking new commits only: ${RANGE}"
        else
            # Baseline is not an ancestor or there was an error, use baseline
            RANGE="${BASELINE_COMMIT}..${local_sha}"
            echo "Remote branch predates baseline. Checking commits from baseline: ${RANGE}"
        fi
    fi

    # Use the Go validator's range validation feature for efficiency
    # The validator will handle baseline checking internally
    echo -e "${BOLD}Validating commit range: ${RANGE}${NC}"

    if go run ./cmd/commitvalidate --from "${RANGE%%..*}" --to "${RANGE##*..}" --verbose; then
        echo ""
        echo -e "${GREEN}Validation successful: All commits follow conventional commit format${NC}"
    else
        echo ""
        echo -e "${RED}Push validation failed: Some commits do not follow conventional commit format${NC}"
        echo ""
        echo "Please fix the invalid commit messages before pushing."
        echo -e "For detailed guidance, see: ${BOLD}docs/conventional-commits.md${NC}"
        echo ""
        echo "To fix recent commits:"
        echo "  • Use 'git commit --amend' to fix the most recent commit"
        echo "  • Use 'git rebase -i ${RANGE%%..*}' to fix older commits"
        echo ""
        echo "To validate locally before pushing:"
        echo "  • Run: ./scripts/validate-pr-commits.sh"
        echo ""
        exit 1
    fi
done

# If we've reached here, all validations passed
echo "Push will proceed."
exit 0
