#!/bin/bash

# Script: check-defaults.sh
# Purpose: Extract and compare the default model name from README.md and internal/config/config.go
# Exit code: 0 if they match, 1 if they don't match
# Created as part of T008 to ensure consistent default model naming

set -e

# Display usage instructions
function show_usage {
  echo "Usage: $0 [options]"
  echo
  echo "Options:"
  echo "  -h, --help     Show this help message and exit"
  echo "  -v, --verbose  Show detailed output while executing"
}

# Setup variables
VERBOSE=false
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
README_PATH="$PROJECT_ROOT/README.md"
CONFIG_PATH="$PROJECT_ROOT/internal/config/config.go"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      show_usage
      exit 0
      ;;
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      show_usage
      exit 1
      ;;
  esac
done

# Validate that both files exist
if [ ! -f "$README_PATH" ]; then
  echo "ERROR: README.md not found at $README_PATH"
  exit 1
fi

if [ ! -f "$CONFIG_PATH" ]; then
  echo "ERROR: config.go not found at $CONFIG_PATH"
  exit 1
fi

# Function to log only if verbose is enabled
function verbose_log {
  if [ "$VERBOSE" = true ]; then
    echo "$1"
  fi
}

# Print header
echo "=== Checking Default Model Consistency ==="
verbose_log "README path: $README_PATH"
verbose_log "Config path: $CONFIG_PATH"
echo

# Extract model name from README.md
verbose_log "Extracting model name from README.md..."
README_MODEL=$(grep -A 3 "| \`--model\`" "$README_PATH" | grep "gemini" | sed 's/.*| `\([^`]*\)`.*/\1/')

if [ -z "$README_MODEL" ]; then
  # Fallback method
  README_MODEL=$(grep -A 5 "Common Options" "$README_PATH" | grep -o '`gemini[^`]*`' | tr -d '`' | head -1)

  if [ -z "$README_MODEL" ]; then
    echo "ERROR: Failed to extract model name from README.md"
    echo "The script expects a table row with the default model in backticks"
    exit 1
  fi
fi

# Extract model name from config.go
verbose_log "Extracting model name from config.go..."
CONFIG_MODEL=$(grep "DefaultModel.*=.*\".*\"" "$CONFIG_PATH" | sed -n 's/.*DefaultModel[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p')

if [ -z "$CONFIG_MODEL" ]; then
  echo "ERROR: Failed to extract model name from config.go"
  echo "The script expects a line containing 'DefaultModel = \"model-name\"'"
  exit 1
fi

# Compare the values
echo "README Model: $README_MODEL"
echo "Config Model: $CONFIG_MODEL"
echo

if [ "$README_MODEL" = "$CONFIG_MODEL" ]; then
  echo "✅ SUCCESS: Default model names match!"
  exit 0
else
  echo "❌ ERROR: Default model names do not match!"
  echo "README.md specifies: '$README_MODEL'"
  echo "config.go specifies: '$CONFIG_MODEL'"
  echo
  echo "Please update the files to ensure consistency."
  exit 1
fi
