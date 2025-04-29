#!/bin/bash
# Pre-commit hook to run Claude as a sanity check for commits
# This checks if the changes adhere to our development philosophy

set -e

echo "Running Claude sanity check on commit changes..."

# Get the commit message from the staged commit
COMMIT_MSG_FILE=$(git rev-parse --git-dir)/COMMIT_EDITMSG
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Get the diff of the staged changes
DIFF=$(git diff --cached)

# Skip if there's no diff (empty commit)
if [ -z "$DIFF" ]; then
  echo "No changes to analyze."
  exit 0
fi

# Fetch the development philosophy files for context
DEV_PHILOSOPHY=$(cat docs/DEVELOPMENT_PHILOSOPHY.md)
DEV_PHILOSOPHY_GO=$(cat docs/DEVELOPMENT_PHILOSOPHY_APPENDIX_GO.md)

# Create a temporary file for the Claude output
CLAUDE_OUTPUT_FILE=$(mktemp)

# Run Claude with the prepared prompt
claude -p "You are a code review assistant for pre-commit hooks. Your task is to analyze the code changes in the given diff and determine if they adhere to our development philosophy.

---DEVELOPMENT PHILOSOPHY SUMMARY---
- Simplicity First: Avoid unnecessary complexity
- Modularity is Mandatory: Small, focused components
- Design for Testability: Code must be testable
- Maintainability Over Premature Optimization
- Explicit is Better than Implicit
- Automate Everything
- Document Decisions, Not Mechanics
- Strict Package Structure: Organize by feature
- No Mocking Internal Collaborators
- Structured Logging with correlation_id
- Error Handling: Return errors, add context
- No Secrets in Code
- Conventional Commits

---COMMIT MESSAGE---
$COMMIT_MSG

---DIFF---
$DIFF

Based ONLY on these changes:
1. Does this commit adhere to our development philosophy?
2. Is the code maintainable, testable, and follows Go best practices?
3. Is the commit message following the Conventional Commits spec?
4. Are there any potential issues or improvements needed?

Respond with:
- PASS: If everything looks good
- WARN: If there are minor issues that should be fixed (with bulleted list)
- FAIL: If there are major issues that must be fixed before committing (with bulleted list)

Be concise but specific about any issues found. Only include actionable feedback." > "$CLAUDE_OUTPUT_FILE"

RESULT=$(cat "$CLAUDE_OUTPUT_FILE")

# Check if Claude found any issues
if echo "$RESULT" | grep -q "^FAIL"; then
  echo -e "\033[0;31m[CLAUDE CHECK FAILED]\033[0m"
  echo "$RESULT"
  echo
  echo -e "\033[0;31mPlease fix the issues before committing.\033[0m"
  rm "$CLAUDE_OUTPUT_FILE"
  exit 1
elif echo "$RESULT" | grep -q "^WARN"; then
  echo -e "\033[0;33m[CLAUDE CHECK WARNING]\033[0m"
  echo "$RESULT"
  echo
  echo -e "\033[0;33mConsider fixing these issues, or commit with --no-verify to bypass.\033[0m"
fi

echo "Claude sanity check passed."
rm "$CLAUDE_OUTPUT_FILE"
exit 0
