#!/bin/bash

# verify-function-length-simple.sh
# Quick verification that key refactored functions are under 100 LOC

set -euo pipefail

MAX_LENGTH=100
VIOLATIONS=0

echo "=== Function Length Verification (Carmack Principle) ==="
echo "Maximum function length: $MAX_LENGTH LOC"
echo ""

# Key files to check (the ones that were refactored)
KEY_FILES=(
    "internal/thinktank/app.go"
    "internal/cli/main.go"
    "internal/logutil/formatting.go"
    "internal/fileutil/filtering.go"
)

check_file() {
    local file="$1"
    echo "Checking: $file"

    if [ ! -f "$file" ]; then
        echo "  ❌ File not found"
        return 1
    fi

    # Simple function length check using grep and awk
    awk '
    /^func / {
        func_name = $0
        gsub(/func /, "", func_name)
        gsub(/\(.*/, "", func_name)
        func_start = NR
        brace_count = 0
        # Count braces on this line
        for (i = 1; i <= length($0); i++) {
            if (substr($0, i, 1) == "{") brace_count++
            if (substr($0, i, 1) == "}") brace_count--
        }
        if (brace_count > 0) {
            in_function = 1
        } else {
            # Single line function
            print "  ✅ " func_name " (1 LOC)"
        }
        next
    }

    in_function {
        for (i = 1; i <= length($0); i++) {
            if (substr($0, i, 1) == "{") brace_count++
            if (substr($0, i, 1) == "}") {
                brace_count--
                if (brace_count == 0) {
                    func_length = NR - func_start
                    if (func_length > '"$MAX_LENGTH"') {
                        print "  ❌ " func_name " (" func_length " LOC > '"$MAX_LENGTH"')"
                        exit 1
                    } else {
                        print "  ✅ " func_name " (" func_length " LOC)"
                    }
                    in_function = 0
                    break
                }
            }
        }
    }
    ' "$file"

    return $?
}

# Check each key file
for file in "${KEY_FILES[@]}"; do
    if ! check_file "$file"; then
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
    echo ""
done

# Quick check of extracted functions specifically
echo "=== Checking Specific Extracted Functions ==="

# Check app.go extracted functions
if [ -f "internal/thinktank/app.go" ]; then
    echo "App.go extracted functions:"
    grep -n "^func.*gatherProjectFiles\|^func.*processFiles\|^func.*generateOutput\|^func.*writeResults" internal/thinktank/app.go | while read line; do
        func_line=$(echo "$line" | cut -d: -f1)
        func_name=$(echo "$line" | cut -d: -f2- | sed 's/func //' | sed 's/(.*//')

        # Count lines until next function or end of file
        next_func=$(awk 'NR > '"$func_line"' && /^func / { print NR; exit }' internal/thinktank/app.go)
        if [ -z "$next_func" ]; then
            next_func=$(wc -l < internal/thinktank/app.go)
        fi

        func_length=$((next_func - func_line))
        if [ "$func_length" -gt "$MAX_LENGTH" ]; then
            echo "  ❌ $func_name ($func_length LOC > $MAX_LENGTH)"
            VIOLATIONS=$((VIOLATIONS + 1))
        else
            echo "  ✅ $func_name ($func_length LOC)"
        fi
    done
fi

# Check formatting.go functions
if [ -f "internal/logutil/formatting.go" ]; then
    echo ""
    echo "Formatting.go extracted functions:"
    grep -c "^func " internal/logutil/formatting.go | while read count; do
        echo "  Found $count functions in formatting.go"
    done

    # Check specific key functions
    for func in "FormatDuration" "FormatToWidth" "ColorizeStatus" "DetectInteractiveEnvironment"; do
        if grep -q "func $func" internal/logutil/formatting.go; then
            echo "  ✅ Found $func function"
        fi
    done
fi

# Check filtering.go functions
if [ -f "internal/fileutil/filtering.go" ]; then
    echo ""
    echo "Filtering.go extracted functions:"
    grep -c "^func " internal/fileutil/filtering.go | while read count; do
        echo "  Found $count functions in filtering.go"
    done

    # Check specific key functions
    for func in "ShouldProcessFile" "CalculateFileStatistics" "ValidateFilePath"; do
        if grep -q "func $func" internal/fileutil/filtering.go; then
            echo "  ✅ Found $func function"
        fi
    done
fi

echo ""
echo "=== Summary ==="
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "✅ SUCCESS: All checked functions appear to be under $MAX_LENGTH LOC"
    echo "✅ Carmack principle compliance verified for key refactored functions"
    exit 0
else
    echo "❌ VIOLATIONS: Found $VIOLATIONS functions exceeding $MAX_LENGTH LOC"
    echo "⚠️  Further refactoring may be needed"
    exit 1
fi
