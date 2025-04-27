#!/bin/bash

# Script to detect manual correlation_id formatting in Go code
# This linter checks for strings containing "correlation_id=" in .go files
# and returns an error if any are found.

# Set colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Search directories with Go files
DIRS="."
if [ $# -gt 0 ]; then
  DIRS="$@"
fi

# Find all Go files, excluding vendor, .git, test files, and logger implementations
FILES=$(find $DIRS -type f -name "*.go" \
  ! -path "*/vendor/*" \
  ! -path "*/.git/*" \
  ! -path "*/internal/lint/*" \
  ! -name "*_test.go" \
  ! -path "*/internal/logutil/logutil.go" \
  ! -path "*/internal/logutil/test_logger.go" \
  ! -path "*/internal/fileutil/mock_logger.go" \
  ! -path "*/internal/test_lint.go")

# Check for the pattern in each file
FOUND=0
for FILE in $FILES; do
  # Use grep to find the pattern, excluding comments
  MATCHES=$(grep -n "correlation_id=" "$FILE" | grep -v "//")

  if [ -n "$MATCHES" ]; then
    echo -e "${RED}Linter Error:${NC} ${FILE} contains forbidden pattern 'correlation_id=':"
    echo "$MATCHES"
    echo "Use logger.WithContext(ctx) instead of manual correlation ID formatting."
    echo ""
    FOUND=1
  fi
done

if [ $FOUND -eq 0 ]; then
  echo -e "${GREEN}Linter passed:${NC} No manual correlation_id formatting found."
  exit 0
else
  echo -e "${RED}Linter failed:${NC} Manual correlation_id formatting found in the code."
  exit 1
fi
