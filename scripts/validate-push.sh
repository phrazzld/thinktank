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

# Check if commitlint is available
if ! command -v commitlint &> /dev/null; then
    echo -e "${RED}Error: commitlint is not installed${NC}"
    echo "Install with: npm install -g @commitlint/cli @commitlint/config-conventional"
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

    # Get the commits to validate
    COMMITS=$(git rev-list ${RANGE})

    if [ -z "${COMMITS}" ]; then
        echo -e "${YELLOW}No commits to validate after baseline.${NC}"
        echo "This could mean:"
        echo "  1. All commits in the push were made before the baseline date"
        echo "  2. There are no new commits to push"
        echo "  3. This branch is based on a version before the baseline"
        echo ""
        echo -e "${GREEN}Validation completed: No issues found (no applicable commits)${NC}"
        continue
    fi

    # Count commits to validate
    COMMIT_COUNT=$(echo "${COMMITS}" | wc -l | tr -d ' ')
    echo -e "${BOLD}Found ${COMMIT_COUNT} commits to validate${NC}"

    # Validate each commit
    FAILED=0
    for COMMIT in ${COMMITS}; do
        COMMIT_MSG=$(git log --format=%B -n 1 ${COMMIT})
        COMMIT_SHORT=$(git log --format=%h -n 1 ${COMMIT})
        COMMIT_DATE=$(git log --format=%ci -n 1 ${COMMIT})
        COMMIT_AUTHOR=$(git log --format="%an <%ae>" -n 1 ${COMMIT})

        echo ""
        echo -e "${BOLD}Checking commit ${COMMIT_SHORT}${NC}"
        echo "Date:   ${COMMIT_DATE}"
        echo "Author: ${COMMIT_AUTHOR}"

        # Use echo to pipe the commit message to commitlint
        if echo "${COMMIT_MSG}" | commitlint; then
            echo -e "${GREEN}✓ Valid conventional commit${NC}"
        else
            echo -e "${RED}✗ Invalid commit format${NC}"
            echo ""
            echo "Commit message:"
            echo "--------------"
            echo "${COMMIT_MSG}"
            echo "--------------"
            echo ""
            echo "Fix tips:"
            echo "  1. Format should be: <type>[optional scope]: <description>"
            echo "  2. Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf"
            echo "  3. Use 'git commit --amend' to fix the most recent commit"
            echo "  4. Use 'git rebase -i' to fix older commits"
            echo ""
            FAILED=1
        fi
    done

    if [ ${FAILED} -eq 1 ]; then
        echo ""
        echo -e "${RED}Push validation failed: Some commits do not follow conventional commit format${NC}"
        echo "Please fix the commit messages before pushing."
        echo "For more details, see docs/conventional-commits.md"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}Validation successful: All commits follow conventional commit format${NC}"
echo "Push will proceed."
exit 0
