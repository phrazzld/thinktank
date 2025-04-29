#!/usr/bin/env bash
# Pre-commit hook to run Claude as a sanity check for commits
# This checks if the changes adhere to our development philosophy

set -eo pipefail

# Error handling function
handle_error() {
  echo "Error occurred in pre-commit hook at line $1"
  exit 1
}

# Set up trap to catch errors
trap 'handle_error $LINENO' ERR

# Skip in CI environments
if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ] || [ -n "$TRAVIS" ] || [ -n "$GITLAB_CI" ] || [ -n "$JENKINS_URL" ]; then
  echo "Skipping Claude sanity check in CI environment"
  exit 0
fi

echo "Running Claude sanity check on commit changes..."

# Get the commit message from the staged commit with error handling
if ! COMMIT_MSG_FILE=$(git rev-parse --git-dir 2>/dev/null)/COMMIT_EDITMSG; then
  echo "Error: Could not locate commit message file"
  exit 1
fi

if [ ! -f "$COMMIT_MSG_FILE" ]; then
  echo "Error: Commit message file does not exist: $COMMIT_MSG_FILE"
  exit 1
fi

if ! COMMIT_MSG=$(cat "$COMMIT_MSG_FILE" 2>/dev/null); then
  echo "Error: Could not read commit message file"
  exit 1
fi

# Get the diff of the staged changes with error handling
if ! DIFF=$(git diff --cached 2>/dev/null); then
  echo "Error: Could not get staged diff"
  exit 1
fi

# Skip if there's no diff (empty commit)
if [ -z "$DIFF" ]; then
  echo "No changes to analyze."
  exit 0
fi

# Get current branch with error handling
if ! BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null); then
  echo "Warning: Could not determine current branch, using 'unknown'"
  BRANCH="unknown"
fi

# We're in a pre-commit hook, so we don't have a commit hash yet
# We'll use a placeholder that will be updated in post-commit
COMMIT_HASH="PRE_COMMIT_$(date +%s)"

# Fetch the development philosophy files for context
if ! DEV_PHILOSOPHY=$(cat docs/DEVELOPMENT_PHILOSOPHY.md 2>/dev/null); then
  echo "Warning: Could not read development philosophy file"
  DEV_PHILOSOPHY="File not found"
fi

if ! DEV_PHILOSOPHY_GO=$(cat docs/DEVELOPMENT_PHILOSOPHY_APPENDIX_GO.md 2>/dev/null); then
  echo "Warning: Could not read Go development philosophy appendix"
  DEV_PHILOSOPHY_GO="File not found"
fi

# Create a temporary file for the Claude output with error handling
if ! CLAUDE_OUTPUT_FILE=$(mktemp 2>/dev/null); then
  echo "Error: Could not create temporary file for Claude output"
  exit 1
fi

# Define log file path
LOG_FILE=".claude-warnings.log"

# Run Claude with the prepared prompt and handle errors
if ! claude -p "You are a code review assistant for pre-commit hooks. Your task is to analyze the code changes in the given diff and determine if they adhere to our development philosophy.

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

Be concise but specific about any issues found. Only include actionable feedback." > "$CLAUDE_OUTPUT_FILE" 2>/dev/null; then
  echo "Error: Failed to run Claude CLI. Is it installed and properly configured?"
  rm -f "$CLAUDE_OUTPUT_FILE" 2>/dev/null || true
  exit 1
fi

# Read Claude's response with error handling
if ! RESULT=$(cat "$CLAUDE_OUTPUT_FILE" 2>/dev/null); then
  echo "Error: Could not read Claude output"
  rm -f "$CLAUDE_OUTPUT_FILE" 2>/dev/null || true
  exit 1
fi

# Log warnings and failures to the log file
log_to_file() {
  local status=$1
  local message=$2

  # Create timestamp and correlation ID with guaranteed uniqueness
  # Using more portable methods for generating random values
  local timestamp=""
  local random_suffix=""

  # Use ISO-8601 format for better cross-platform compatibility
  if ! timestamp=$(date -u "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null); then
    timestamp="$(date "+%Y-%m-%d")-unknown-time"
  fi

  # Generate random string in a cross-platform way
  if command -v openssl >/dev/null 2>&1; then
    # Use openssl if available (most systems have this)
    if ! random_suffix=$(openssl rand -hex 4 2>/dev/null); then
      random_suffix="$RANDOM$$"
    fi
  else
    # Fallback to a timestamp + pid based approach
    random_suffix="${RANDOM}${RANDOM}$$"
  fi

  local correlation_id="claude_${COMMIT_HASH}_$(date +%s)_${random_suffix}"

  # Ensure the directory exists with error handling
  if ! mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null; then
    echo "Warning: Could not create directory for log file"
    return 1
  fi

  # Write to log file with commit hash, branch, timestamp, and correlation_id
  {
    echo "----------------------------------------"
    echo "DATE: $timestamp"
    echo "STATUS: $status"
    echo "COMMIT: $COMMIT_HASH"
    echo "BRANCH: $BRANCH"
    echo "COMMIT_MSG: $(echo "$COMMIT_MSG" | head -1)"
    echo "CORRELATION_ID: $correlation_id"
    echo "----------------------------------------"
    echo "$message"
    echo ""
  } >> "$LOG_FILE" 2>/dev/null

  # Check if writing to log file failed
  if [ $? -ne 0 ]; then
    echo "Warning: Could not write to log file: $LOG_FILE"
    return 1
  fi

  # Write commit hash to temp file for post-commit hook with error handling
  if ! echo "$COMMIT_HASH" > "./.git/claude_last_check" 2>/dev/null; then
    echo "Warning: Could not create temp file for post-commit hook"
    return 1
  fi

  return 0
}

# Check if Claude found any issues with error handling
if echo "$RESULT" | grep -q "^FAIL" 2>/dev/null; then
  echo -e "\033[0;31m[CLAUDE CHECK FAILED]\033[0m"
  echo "$RESULT"
  echo
  echo -e "\033[0;31mPlease fix the issues before committing.\033[0m"

  # Log the failure
  log_to_file "FAIL" "$RESULT"
  if [ $? -ne 0 ]; then
    echo "Warning: Failed to log results"
  fi

  rm -f "$CLAUDE_OUTPUT_FILE" 2>/dev/null || true
  exit 1
elif echo "$RESULT" | grep -q "^WARN" 2>/dev/null; then
  echo -e "\033[0;33m[CLAUDE CHECK WARNING]\033[0m"
  echo "$RESULT"
  echo
  echo -e "\033[0;33mConsider fixing these issues, or commit with --no-verify to bypass.\033[0m"

  # Log the warning
  log_to_file "WARN" "$RESULT"
  if [ $? -ne 0 ]; then
    echo "Warning: Failed to log results"
  fi
else
  # If it's a PASS, we don't log by default
  # Uncomment these lines if you want to log successful checks too
  # log_to_file "PASS" "$RESULT"
  # if [ $? -ne 0 ]; then
  #   echo "Warning: Failed to log results"
  # fi

  echo "Claude sanity check passed."
fi

# Clean up
rm -f "$CLAUDE_OUTPUT_FILE" 2>/dev/null || true
exit 0
