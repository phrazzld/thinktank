#!/bin/bash
set -e
WARNING_THRESHOLD=500
ERROR_THRESHOLD=1000
# Get top-level directory relative to script location or CWD
GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$GIT_ROOT"

# Find Go files excluding vendor
GO_FILES=$(find . -path "./vendor" -prune -o -name "*.go" -print)
WARNING_FILES=()
ERROR_FILES=()

for file in $GO_FILES; do
  # Ensure file exists before checking size
  if [ -f "$file" ]; then
    LINES=$(wc -l < "$file" | tr -d ' ') # Ensure no extra spaces
    RELATIVE_PATH="${file#./}"
    if [ "$LINES" -gt "$ERROR_THRESHOLD" ]; then
      ERROR_FILES+=("$RELATIVE_PATH: $LINES lines")
    elif [ "$LINES" -gt "$WARNING_THRESHOLD" ]; then
      WARNING_FILES+=("$RELATIVE_PATH: $LINES lines")
    fi
  fi
done

HAS_ERROR=0

if [ ${#WARNING_FILES[@]} -gt 0 ]; then
  echo "⚠️  WARNING: The following files exceed $WARNING_THRESHOLD lines:" >&2
  printf "  %s\n" "${WARNING_FILES[@]}" >&2
  echo "  Consider refactoring these files for better maintainability." >&2
fi

if [ ${#ERROR_FILES[@]} -gt 0 ]; then
  echo "❌ ERROR: The following files exceed $ERROR_THRESHOLD lines:" >&2
  printf "  %s\n" "${ERROR_FILES[@]}" >&2
  echo "  Commit blocked - you must refactor these files before committing." >&2
  echo "  See DEVELOPMENT_PHILOSOPHY.md for code organization guidelines." >&2
  HAS_ERROR=1
fi

exit $HAS_ERROR
