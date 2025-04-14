#!/bin/bash
set -e
LINE_THRESHOLD=1000
# Get top-level directory relative to script location or CWD
GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$GIT_ROOT"

# Find Go files excluding vendor
GO_FILES=$(find . -path "./vendor" -prune -o -name "*.go" -print)
LARGE_FILES=()

for file in $GO_FILES; do
  # Ensure file exists before checking size
  if [ -f "$file" ]; then
    LINES=$(wc -l < "$file" | tr -d ' ') # Ensure no extra spaces
    if [ "$LINES" -gt "$LINE_THRESHOLD" ]; then
      # Get relative path for cleaner output
      RELATIVE_PATH="${file#./}"
      LARGE_FILES+=("$RELATIVE_PATH: $LINES lines")
    fi
  fi
done

if [ ${#LARGE_FILES[@]} -gt 0 ]; then
  echo "⚠️  WARNING: The following files exceed $LINE_THRESHOLD lines:" >&2 # Output warning to stderr
  printf "  %s\n" "${LARGE_FILES[@]}" >&2
  echo "  Consider refactoring these files for better maintainability." >&2
  echo "  Check TODO.md for refactoring guidelines." >&2
  # Allowing commit to proceed with just a warning
fi
exit 0
