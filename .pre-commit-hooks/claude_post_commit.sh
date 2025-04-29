#!/usr/bin/env bash
# Post-commit hook to update Claude warning logs with the actual commit hash

set -eo pipefail

# Error handling function
handle_error() {
  echo "Error occurred in post-commit hook at line $1"
  exit 1
}

# Set up trap to catch errors
trap 'handle_error $LINENO' ERR

# Skip in CI environments
if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ] || [ -n "$TRAVIS" ] || [ -n "$GITLAB_CI" ] || [ -n "$JENKINS_URL" ]; then
  echo "Skipping Claude post-commit hook in CI environment"
  exit 0
fi

# Check if we have a temporary file with the last Claude check
if [ -f "./.git/claude_last_check" ]; then
  if ! PLACEHOLDER_HASH=$(cat "./.git/claude_last_check" 2>/dev/null); then
    echo "Warning: Could not read placeholder hash from .git/claude_last_check"
    rm -f "./.git/claude_last_check" 2>/dev/null || true
    exit 0
  fi

  if ! ACTUAL_HASH=$(git rev-parse HEAD 2>/dev/null); then
    echo "Warning: Could not get actual commit hash"
    rm -f "./.git/claude_last_check" 2>/dev/null || true
    exit 0
  fi

  LOG_FILE=".claude-warnings.log"

  if [ -f "$LOG_FILE" ]; then
    # Replace the placeholder hash with the actual commit hash in the log file
    # Using a portable approach that works across different sed versions

    # Create a temporary file with proper error handling
    if ! TMP_LOG=$(mktemp 2>/dev/null); then
      echo "Warning: Could not create temporary file"
      rm -f "./.git/claude_last_check" 2>/dev/null || true
      exit 0
    fi

    # Perform the replacement in a way that works on both GNU and BSD sed
    sed "s|$PLACEHOLDER_HASH|$ACTUAL_HASH|g" "$LOG_FILE" > "$TMP_LOG" 2>/dev/null

    # Check if sed operation failed
    if [ $? -ne 0 ]; then
      echo "Warning: Failed to perform sed replacement"
      rm -f "$TMP_LOG" 2>/dev/null || true
      rm -f "./.git/claude_last_check" 2>/dev/null || true
      exit 0
    fi

    # Move the temporary file back to the original
    if ! mv "$TMP_LOG" "$LOG_FILE" 2>/dev/null; then
      echo "Warning: Failed to update log file"
      rm -f "$TMP_LOG" 2>/dev/null || true
    fi
  fi

  # Clean up
  rm -f "./.git/claude_last_check" 2>/dev/null || true
fi

exit 0
