#!/bin/bash
# validate-baseline-commits.sh
#
# This script provides a forward-only approach to commit message validation,
# preserving git history while ensuring all future commits follow standards.
#
# It uses pure bash for validation without external dependencies.
#
# It creates a "baseline commit ID" file on first run that marks the starting point
# for validation. Only commits made AFTER this file was created will be validated.
#
# Usage: ./validate-baseline-commits.sh [baseline-file-path]
#
# If no baseline file path is provided, it uses the default in .git/BASELINE_COMMIT

set -eo pipefail

# Determine baseline file path
DEFAULT_BASELINE_FILE=".git/BASELINE_COMMIT"
BASELINE_FILE=${1:-$DEFAULT_BASELINE_FILE}

# Current timestamp for new baseline creation
CURRENT_DATE=$(date +"%Y-%m-%d %H:%M:%S")
CURRENT_COMMIT=$(git rev-parse HEAD)

# Check if baseline file exists
if [ ! -f "$BASELINE_FILE" ]; then
  echo "No baseline commit file found. Creating one to mark the starting point for validation."
  echo "# Commit validation baseline" > "$BASELINE_FILE"
  echo "# All commits made AFTER this baseline will be validated against conventional commit standards" >> "$BASELINE_FILE"
  echo "# Created: $CURRENT_DATE" >> "$BASELINE_FILE"
  echo "$CURRENT_COMMIT" >> "$BASELINE_FILE"
  echo ""
  echo "✅ Created baseline at current commit ($CURRENT_COMMIT)"
  echo "✅ Future commits will be validated against conventional commit standards"
  echo "✅ Historical commits are preserved and exempt from validation"
  echo ""
  echo "No validation needed - baseline just created."
  exit 0
fi

# Read the baseline commit from file
BASELINE_COMMIT=$(tail -n 1 "$BASELINE_FILE")

# Validation status
VALIDATION_PASSED=true
INVALID_COMMITS=0
TOTAL_COMMITS=0

echo "Baseline Commit Validation"
echo "=========================="
echo "Only validating commits made after baseline: ${BASELINE_COMMIT}"
echo "Baseline created: $(grep "Created:" "$BASELINE_FILE" | sed 's/# Created: //')"

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

# Valid commit types
VALID_TYPES="feat|fix|docs|style|refactor|perf|test|chore|ci|build|revert"

# Process each commit
echo "Validating commits..."
for commit in $COMMITS_TO_CHECK; do
  commit_msg=$(git log -1 --format="%B" $commit)
  commit_short=$(git log -1 --format="%h" $commit)
  commit_subject=$(echo "$commit_msg" | head -n 1)
  commit_date=$(git log -1 --format="%ad" --date=iso $commit)

  echo -n "Checking commit $commit_short ($commit_date): \"$commit_subject\"... "

  # Pure bash validation of commit message format using regex
  # Match pattern: <type>[optional scope]: <description>
  # Example: feat(api): add new endpoint
  if [[ "$commit_subject" =~ ^($VALID_TYPES)(\([a-z0-9/-]+\))?!?:\ [a-z] ]]; then
    echo "✓ VALID"
  else
    echo "✗ INVALID"
    echo "=========================="
    echo "Invalid commit: $commit_short"
    echo "------------------------"
    echo "$commit_msg"
    echo "------------------------"
    echo "Validation errors:"
    echo "Commit message must follow the conventional commit format:"
    echo "<type>[optional scope]: <description>"
    echo ""
    echo "Valid types: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert"
    echo "Scope is optional and should be lowercase"
    echo "Description should start with lowercase letter"
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
  echo "Note: This project has a baseline validation policy that only requires conventional"
  echo "commit messages for commits made AFTER the baseline file was created."
  echo ""
  echo "Options to fix this:"
  echo "1. Add new commits that follow the conventional format"
  echo "2. See docs/conventional-commits.md for format guidelines"
  exit 1
fi
