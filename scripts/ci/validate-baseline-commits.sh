#!/bin/bash
# validate-baseline-commits.sh
#
# This script validates commit messages according to conventional commits standard,
# but only for commits that occur AFTER a specified baseline commit.
# This allows preserving git history while ensuring all new commits follow conventions.
#
# Usage: ./validate-baseline-commits.sh [baseline-commit-sha]
#
# If no baseline commit is provided, it uses the default specified in this script.

set -eo pipefail

# Default baseline commit (May 18, 2025)
DEFAULT_BASELINE_COMMIT="1300e4d675ac087783199f1e608409e6853e589f"
BASELINE_COMMIT=${1:-$DEFAULT_BASELINE_COMMIT}

# Validation status
VALIDATION_PASSED=true
INVALID_COMMITS=0
TOTAL_COMMITS=0

echo "Baseline Commit Validation"
echo "=========================="
echo "Only validating commits after baseline: ${BASELINE_COMMIT}"

# Check if we're in a CI environment, specifically a PR
if [ -n "$GITHUB_BASE_REF" ]; then
  echo "CI environment detected (Pull Request)"

  # For pull requests, we need to:
  # 1. Get the base branch ref
  # 2. Find common ancestor between baseline and the current branch
  # 3. Validate commits between that ancestor and HEAD

  # Get the base ref
  BASE_REF="${GITHUB_BASE_REF:-master}"
  echo "Base branch: $BASE_REF"

  # Create tracking branch for base ref if it doesn't exist
  git fetch origin $BASE_REF --depth=1000 || true

  # Check if the baseline commit exists in the history
  if ! git cat-file -e "$BASELINE_COMMIT^{commit}" 2>/dev/null; then
    echo "Baseline commit $BASELINE_COMMIT not found in current history."
    echo "This might be due to shallow clone or the commit doesn't exist."
    echo "Falling back to checking PR commits only."

    # In this case, we'll only check the commits in the PR
    COMMITS_TO_CHECK=$(git log --format="%H" "origin/$BASE_REF..HEAD")
  else
    # Find the most recent common ancestor between baseline and current HEAD
    COMMON_ANCESTOR=$(git merge-base $BASELINE_COMMIT HEAD)
    echo "Common ancestor with baseline: $COMMON_ANCESTOR"

    # Get all commits after the common ancestor that are in the current PR
    # This ensures we only validate:
    # 1. Commits after the baseline
    # 2. Commits that are part of this PR
    COMMITS_TO_CHECK=$(git log --format="%H" "$COMMON_ANCESTOR..HEAD")
  fi
else
  # For local validation, simply validate all commits after the baseline
  echo "Local environment detected"

  # Check if baseline commit exists
  if ! git cat-file -e "$BASELINE_COMMIT^{commit}" 2>/dev/null; then
    echo "Error: Baseline commit $BASELINE_COMMIT not found in current history."
    echo "Please ensure the baseline commit is valid or use a different commit."
    exit 1
  fi

  # Get all commits after the baseline
  COMMITS_TO_CHECK=$(git log --format="%H" "$BASELINE_COMMIT..HEAD")
fi

# No commits to check
if [ -z "$COMMITS_TO_CHECK" ]; then
  echo "No commits to validate after the baseline."
  exit 0
fi

# Count total commits to check
TOTAL_COMMITS=$(echo "$COMMITS_TO_CHECK" | wc -l | tr -d ' ')
echo "Found $TOTAL_COMMITS commit(s) to validate"

# Setup commitlint command
# First check if npx is available
if command -v npx &> /dev/null; then
  COMMITLINT_CMD="npx commitlint"
else
  # Fall back to direct node_modules path if npx isn't available
  COMMITLINT_CMD="./node_modules/.bin/commitlint"

  # Check if commitlint is installed
  if [ ! -f "$COMMITLINT_CMD" ]; then
    echo "Error: commitlint not found. Please ensure @commitlint/cli is installed."
    echo "You can install it with: npm install --save-dev @commitlint/cli @commitlint/config-conventional"
    exit 1
  fi
fi

# Process each commit
echo "Validating commits..."
for commit in $COMMITS_TO_CHECK; do
  commit_msg=$(git log -1 --format="%B" $commit)
  commit_short=$(git log -1 --format="%h" $commit)
  commit_subject=$(echo "$commit_msg" | head -n 1)

  echo -n "Checking commit $commit_short: \"$commit_subject\"... "

  # Check commit message against commitlint rules
  if echo "$commit_msg" | $COMMITLINT_CMD > /dev/null 2>&1; then
    echo "✓ VALID"
  else
    echo "✗ INVALID"
    echo "=========================="
    echo "Invalid commit: $commit_short"
    echo "------------------------"
    echo "$commit_msg"
    echo "------------------------"
    echo "Validation errors:"
    echo "$commit_msg" | $COMMITLINT_CMD || true
    echo "=========================="
    VALIDATION_PASSED=false
    INVALID_COMMITS=$((INVALID_COMMITS + 1))
  fi
done

# Summary
echo ""
echo "Validation Summary"
echo "================="
echo "Total commits checked: $TOTAL_COMMITS"
echo "Invalid commits found: $INVALID_COMMITS"

if [ "$VALIDATION_PASSED" = true ]; then
  echo "✅ Validation passed! All commits after the baseline follow conventional commits format."
  exit 0
else
  echo "❌ Validation failed! Some commits after the baseline do not follow conventional commits format."
  echo ""
  echo "Note: This project has a baseline commit policy that only requires conventional"
  echo "commit messages for commits made AFTER $BASELINE_COMMIT (May 18, 2025)."
  echo ""
  echo "Options to fix this:"
  echo "1. Add new commits that follow the conventional format"
  echo "2. See docs/conventional-commits.md for format guidelines"
  exit 1
fi
