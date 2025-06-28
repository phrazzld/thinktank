#!/bin/bash
set -e

# check-coverage-detailed.sh - Enhanced coverage analysis with actionable recommendations
# Usage: ./scripts/check-coverage-detailed.sh [threshold]

THRESHOLD=${1:-90}

echo "🔍 Detailed Coverage Analysis (Target: ${THRESHOLD}%)"
echo "========================================================"

# Generate coverage if needed
if [ ! -f coverage.out ]; then
  echo "Generating coverage data..."
  ./scripts/check-coverage.sh "$THRESHOLD" >/dev/null 2>&1 || true
fi

# Get overall coverage
OVERALL=$(go tool cover -func=coverage.out | grep "total:" | grep -v "_total:" | awk '{print $3}' | tr -d '%')
echo "Overall Coverage: ${OVERALL}%"

# Calculate gap to target
GAP=$(awk -v overall="$OVERALL" -v target="$THRESHOLD" 'BEGIN { printf "%.1f", target - overall }')
echo "Gap to Target: ${GAP}%"
echo ""

# Analyze packages by coverage level
echo "📊 Package Coverage Analysis:"
echo "-----------------------------"

# Use the test output to get package coverage (more reliable than parsing coverage.out)
echo "Running package-level coverage analysis..."
PACKAGES_COVERAGE=$(go test -cover ./... 2>&1 | grep -E "^ok.*coverage:" | \
  awk '{
    # Extract package name and coverage percentage
    package = $2;
    for(i=3; i<=NF; i++) {
      if($i ~ /coverage:/) {
        coverage = $(i+1);
        gsub(/%/, "", coverage);
        if(coverage != "") {
          print coverage " " package;
        }
        break;
      }
    }
  }' | sort -nr)

# High performers (>= 90%)
echo "🟢 High Performers (≥90%):"
echo "$PACKAGES_COVERAGE" | awk '$1 >= 90 { printf "  ✅ %-50s %6.1f%%\n", $2, $1 }' | head -10

# Medium performers (80-89%)
echo ""
echo "🟡 Medium Performers (80-89%):"
echo "$PACKAGES_COVERAGE" | awk '$1 >= 80 && $1 < 90 { printf "  📈 %-50s %6.1f%%\n", $2, $1 }'

# Low performers (<80%)
echo ""
echo "🔴 Needs Attention (<80%):"
echo "$PACKAGES_COVERAGE" | awk '$1 < 80 { printf "  ⚠️  %-50s %6.1f%%\n", $2, $1 }'

echo ""
echo "🎯 Priority Recommendations:"
echo "----------------------------"

# Find the 3 lowest coverage packages
echo "Focus on these packages for biggest impact:"
echo "$PACKAGES_COVERAGE" | tail -3 | awk '{
  printf "  📊 %-50s %6.1f%% - Target: +5%%\n", $2, $1;
}'

echo ""
echo "💡 Next Steps:"
echo "--------------"
echo "1. Pick the lowest coverage package from above"
echo "2. Generate detailed report: go test -coverprofile=pkg.out ./internal/PACKAGE"
echo "3. View in browser: go tool cover -html=pkg.out -o pkg.html && open pkg.html"
echo "4. Write tests for uncovered functions (focus on error paths)"
echo "5. Run: ./scripts/update-coverage-threshold.sh when ready to increase"

echo ""
echo "⚡ Quick Commands:"
echo "-----------------"
echo "# Check specific package:"
echo "go test -cover ./internal/PACKAGE"
echo ""
echo "# Generate HTML report for specific package:"
echo "go test -coverprofile=pkg.out ./internal/PACKAGE && go tool cover -html=pkg.out"
echo ""
echo "# Update threshold when coverage improves:"
echo "./scripts/update-coverage-threshold.sh"
