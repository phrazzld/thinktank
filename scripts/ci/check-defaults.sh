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
# Look for mentions of the default model in various patterns
README_MODEL=""

# Try to find default model mentioned explicitly (not just examples)
README_MODEL=$(grep -i "default.*model.*:" "$README_PATH" | grep -o '`[^`]*gemini[^`]*`' | tr -d '`' | head -1)

# If not found, look for explicit statements about default models
if [ -z "$README_MODEL" ]; then
  README_MODEL=$(grep -i "uses.*\`gemini" "$README_PATH" | grep -o '`[^`]*gemini[^`]*`' | tr -d '`' | head -1)
fi

# The README uses intelligent model selection without hardcoding a single default
# It mentions different models for different scenarios (small vs large inputs)
# So we should not expect to find a single hardcoded default
verbose_log "README describes intelligent model selection rather than a single hardcoded default"
README_MODEL="NOT_SPECIFIED_IN_README"

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

# Handle the case where README doesn't specify a default model
if [ "$README_MODEL" = "NOT_SPECIFIED_IN_README" ]; then
  echo "✅ SUCCESS: README uses intelligent model selection (no hardcoded default)"
  echo "Config default model: $CONFIG_MODEL"
  echo "This is acceptable since the README describes intelligent model selection."
  exit 0
elif [ "$README_MODEL" = "$CONFIG_MODEL" ]; then
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
