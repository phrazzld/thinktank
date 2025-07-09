#!/bin/bash

# verify-function-length.sh
# Verifies that all functions are under 100 lines of code (Carmack principle)
# Analyzes Go files and identifies functions exceeding the 100 LOC threshold

set -euo pipefail

# Configuration
MAX_FUNCTION_LENGTH=100
OUTPUT_DIR="reports/function-analysis/$(date +%Y%m%d_%H%M%S)"
DETAILED_REPORT=true

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Function Length Verification (Carmack Principle) ===${NC}"
echo "Date: $(date)"
echo "Commit: $(git rev-parse HEAD)"
echo "Branch: $(git branch --show-current)"
echo "Maximum function length: $MAX_FUNCTION_LENGTH LOC"
echo "Output directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to analyze a Go file for function lengths
analyze_go_file() {
    local file_path="$1"
    local temp_analysis="$OUTPUT_DIR/$(basename "$file_path" .go)_analysis.txt"

    if [ ! -f "$file_path" ]; then
        echo "File not found: $file_path" >&2
        return 1
    fi

    echo "Analyzing: $file_path" > "$temp_analysis"
    echo "===========================================" >> "$temp_analysis"
    echo "" >> "$temp_analysis"

    # Use awk to parse function definitions and calculate lengths
    awk '
    BEGIN {
        in_function = 0
        function_name = ""
        function_start = 0
        brace_count = 0
        max_length = '"$MAX_FUNCTION_LENGTH"'
        violations = 0
        total_functions = 0
    }

    # Match function definitions
    /^func / {
        if (in_function) {
            # End previous function
            length = NR - function_start - 1
            total_functions++
            if (length > max_length) {
                printf "❌ VIOLATION: %s (%d LOC > %d)\n", function_name, length, max_length
                violations++
            } else {
                printf "✅ OK: %s (%d LOC)\n", function_name, length
            }
        }

        # Start new function
        in_function = 1
        function_start = NR
        brace_count = 0

        # Extract function name
        if (match($0, /func\s+(\([^)]*\)\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(/, arr)) {
            function_name = arr[2]
        } else if (match($0, /func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(/, arr)) {
            function_name = arr[1]
        } else {
            function_name = "UNKNOWN"
        }

        # Count opening braces on function line
        for (i = 1; i <= length($0); i++) {
            char = substr($0, i, 1)
            if (char == "{") brace_count++
            if (char == "}") brace_count--
        }
        next
    }

    # Count braces to determine function end
    in_function {
        for (i = 1; i <= length($0); i++) {
            char = substr($0, i, 1)
            if (char == "{") brace_count++
            if (char == "}") {
                brace_count--
                if (brace_count == 0) {
                    # Function ended
                    length = NR - function_start - 1
                    total_functions++
                    if (length > max_length) {
                        printf "❌ VIOLATION: %s (%d LOC > %d)\n", function_name, length, max_length
                        violations++
                    } else {
                        printf "✅ OK: %s (%d LOC)\n", function_name, length
                    }
                    in_function = 0
                    function_name = ""
                    function_start = 0
                }
            }
        }
    }

    END {
        if (in_function && function_name != "") {
            # Handle last function in file
            length = NR - function_start - 1
            total_functions++
            if (length > max_length) {
                printf "❌ VIOLATION: %s (%d LOC > %d)\n", function_name, length, max_length
                violations++
            } else {
                printf "✅ OK: %s (%d LOC)\n", function_name, length
            }
        }

        printf "\nSUMMARY:\n"
        printf "Total functions: %d\n", total_functions
        printf "Violations: %d\n", violations
        printf "Compliance: %.1f%%\n", total_functions > 0 ? (total_functions - violations) * 100 / total_functions : 100
    }
    ' "$file_path" >> "$temp_analysis"

    echo "$temp_analysis"
}

# Function to find all Go files (excluding vendor, test files initially)
find_go_files() {
    find . -name "*.go" \
        -not -path "./vendor/*" \
        -not -path "./.git/*" \
        -not -path "./reports/*" \
        -not -path "./benchmarks/*" \
        -not -path "./internal/thinktank/thinktank_*" \
        -not -path "./*_test.go" \
        -type f 2>/dev/null | sort
}

# Function to find test Go files separately
find_test_files() {
    find . -name "*_test.go" \
        -not -path "./vendor/*" \
        -not -path "./.git/*" \
        -not -path "./reports/*" \
        -not -path "./benchmarks/*" \
        -not -path "./internal/thinktank/thinktank_*" \
        -type f 2>/dev/null | sort
}

echo -e "${BLUE}1. Analyzing production Go files...${NC}"

# Analyze all production Go files
production_files=$(find_go_files)
total_violations=0
total_functions=0
analyzed_files=0

for file in $production_files; do
    if [ -f "$file" ]; then
        echo -e "${BLUE}Analyzing: $file${NC}"
        analysis_file=$(analyze_go_file "$file")

        if [ -f "$analysis_file" ]; then
            # Extract violations and function count from analysis
            violations=$(grep "VIOLATION:" "$analysis_file" | wc -l || echo "0")
            functions=$(grep -E "(✅ OK:|❌ VIOLATION:)" "$analysis_file" | wc -l || echo "0")

            total_violations=$((total_violations + violations))
            total_functions=$((total_functions + functions))
            analyzed_files=$((analyzed_files + 1))

            if [ "$violations" -gt 0 ]; then
                echo -e "${RED}  Found $violations violations in $file${NC}"
                grep "VIOLATION:" "$analysis_file" | while read line; do
                    echo -e "${RED}    $line${NC}"
                done
            else
                echo -e "${GREEN}  ✓ All functions in $file are compliant${NC}"
            fi
        fi
    fi
done

echo ""
echo -e "${BLUE}2. Analyzing test Go files...${NC}"

# Analyze test files separately (often have longer functions due to test data)
test_files=$(find_test_files)
test_violations=0
test_functions=0
test_analyzed_files=0

for file in $test_files; do
    if [ -f "$file" ]; then
        echo -e "${BLUE}Analyzing test file: $file${NC}"
        analysis_file=$(analyze_go_file "$file")

        if [ -f "$analysis_file" ]; then
            violations=$(grep "VIOLATION:" "$analysis_file" | wc -l || echo "0")
            functions=$(grep -E "(✅ OK:|❌ VIOLATION:)" "$analysis_file" | wc -l || echo "0")

            test_violations=$((test_violations + violations))
            test_functions=$((test_functions + functions))
            test_analyzed_files=$((test_analyzed_files + 1))

            if [ "$violations" -gt 0 ]; then
                echo -e "${YELLOW}  Found $violations long test functions in $file${NC}"
                # Note: Test files often have longer functions due to test data - this is acceptable
            else
                echo -e "${GREEN}  ✓ All test functions in $file are compliant${NC}"
            fi
        fi
    fi
done

echo ""
echo -e "${BLUE}3. Generating comprehensive report...${NC}"

# Generate comprehensive report
report_file="$OUTPUT_DIR/function_length_report.md"
cat > "$report_file" << EOF
# Function Length Analysis Report

**Generated:** $(date)
**Commit:** $(git rev-parse HEAD)
**Branch:** $(git branch --show-current)
**Maximum Function Length:** $MAX_FUNCTION_LENGTH LOC

## Executive Summary

### Production Code Analysis
- **Files Analyzed:** $analyzed_files
- **Total Functions:** $total_functions
- **Functions Over $MAX_FUNCTION_LENGTH LOC:** $total_violations
- **Compliance Rate:** $([ "$total_functions" -gt 0 ] && echo "scale=1; ($total_functions - $total_violations) * 100 / $total_functions" | bc -l || echo "100.0")%

### Test Code Analysis
- **Test Files Analyzed:** $test_analyzed_files
- **Total Test Functions:** $test_functions
- **Test Functions Over $MAX_FUNCTION_LENGTH LOC:** $test_violations
- **Test Compliance Rate:** $([ "$test_functions" -gt 0 ] && echo "scale=1; ($test_functions - $test_violations) * 100 / $test_functions" | bc -l || echo "100.0")%

## Detailed Analysis

### Carmack Principle Compliance

The analysis verifies adherence to John Carmack's principle of keeping functions small and focused:
- **Target:** All functions under $MAX_FUNCTION_LENGTH lines of code
- **Rationale:** Small functions are easier to understand, test, and maintain

### Production Code Results

EOF

if [ "$total_violations" -eq 0 ]; then
    cat >> "$report_file" << EOF
✅ **EXCELLENT COMPLIANCE**

All $total_functions production functions are under $MAX_FUNCTION_LENGTH LOC. This demonstrates excellent adherence to Carmack principles and indicates well-structured, maintainable code.

### Key Achievements
- **Function Extraction Success:** Refactoring successfully broke down large functions
- **Maintainability:** All functions are focused and easy to understand
- **Testability:** Small functions are directly testable without complex setup

EOF
else
    cat >> "$report_file" << EOF
⚠️ **VIOLATIONS DETECTED**

Found $total_violations functions exceeding $MAX_FUNCTION_LENGTH LOC out of $total_functions total functions.

### Violation Summary
EOF

    # Add violation details from analysis files
    for file in $production_files; do
        analysis_file="$OUTPUT_DIR/$(basename "$file" .go)_analysis.txt"
        if [ -f "$analysis_file" ] && grep -q "VIOLATION:" "$analysis_file"; then
            echo "" >> "$report_file"
            echo "#### $(basename "$file")" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            grep "VIOLATION:" "$analysis_file" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
        fi
    done
fi

cat >> "$report_file" << EOF

### Test Code Considerations

Test functions often exceed $MAX_FUNCTION_LENGTH LOC due to:
- **Test Data Setup:** Large test case definitions
- **Table-Driven Tests:** Comprehensive test scenarios
- **Integration Tests:** Complex setup and verification logic

**Test Violations:** $test_violations functions (acceptable for test code)

## Recommendations

EOF

if [ "$total_violations" -eq 0 ]; then
    cat >> "$report_file" << EOF
1. **Maintain Excellence:** Continue following Carmack principles in future development
2. **Code Reviews:** Ensure new functions stay under $MAX_FUNCTION_LENGTH LOC
3. **Refactoring Success:** The current refactoring has achieved its goals

EOF
else
    cat >> "$report_file" << EOF
1. **Address Violations:** Refactor functions exceeding $MAX_FUNCTION_LENGTH LOC
2. **Function Extraction:** Break large functions into smaller, focused functions
3. **Review Process:** Implement checks to prevent large functions in future

### Suggested Refactoring Approach
- Extract helper functions for complex logic
- Separate I/O operations from business logic
- Use table-driven patterns to reduce code duplication

EOF
fi

cat >> "$report_file" << EOF
## Files Analyzed

### Production Code
EOF

echo "\`\`\`" >> "$report_file"
find_go_files >> "$report_file"
echo "\`\`\`" >> "$report_file"

cat >> "$report_file" << EOF

### Test Code
EOF

echo "\`\`\`" >> "$report_file"
find_test_files >> "$report_file"
echo "\`\`\`" >> "$report_file"

cat >> "$report_file" << EOF

## Detailed Analysis Files

Individual analysis files are available in the \`$OUTPUT_DIR\` directory:
EOF

for file in "$OUTPUT_DIR"/*_analysis.txt; do
    if [ -f "$file" ]; then
        echo "- [$(basename "$file")]($(basename "$file"))" >> "$report_file"
    fi
done

echo "" >> "$report_file"
echo "---" >> "$report_file"
echo "*Generated by Thinktank Function Length Verification Tool*" >> "$report_file"

echo -e "${GREEN}✓ Comprehensive report saved to: $report_file${NC}"

echo ""
echo -e "${BLUE}=== Function Length Verification Complete ===${NC}"
echo "Analysis saved to: $OUTPUT_DIR"

# Print summary
echo ""
echo -e "${BLUE}📊 SUMMARY${NC}"
echo "Production Code:"
echo -e "  Files analyzed: $analyzed_files"
echo -e "  Total functions: $total_functions"
echo -e "  Violations: $total_violations"

if [ "$total_violations" -eq 0 ]; then
    echo -e "  ${GREEN}✅ ALL FUNCTIONS COMPLIANT${NC}"
    echo -e "  ${GREEN}✅ Carmack principle successfully achieved${NC}"
    exit_code=0
else
    echo -e "  ${RED}❌ $total_violations functions exceed $MAX_FUNCTION_LENGTH LOC${NC}"
    echo -e "  ${YELLOW}⚠ Refactoring needed for full compliance${NC}"
    exit_code=1
fi

echo ""
echo "Test Code:"
echo -e "  Files analyzed: $test_analyzed_files"
echo -e "  Total functions: $test_functions"
echo -e "  Long functions: $test_violations (acceptable for tests)"

echo ""
echo "Next steps:"
echo "1. Review the detailed report: $report_file"
echo "2. Address any violations in production code"
echo "3. Integrate this check into CI/CD pipeline"

exit $exit_code
