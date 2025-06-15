#!/bin/bash

# License compliance checking script for local validation
# Mirrors the logic from .github/workflows/security-gates.yml license-scan job

set -eo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Color definitions for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to display help
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Check license compliance for Go dependencies

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    --report-only          Generate report without failing on violations
    --output-file FILE     Save license report to specified file (default: licenses.csv)

EXAMPLES:
    $0                      # Basic license check
    $0 -v                   # Verbose license check
    $0 --report-only        # Generate report without failing
    $0 --output-file deps.csv  # Save report to custom file

DESCRIPTION:
    This script checks the licenses of all Go dependencies in the current project
    against a predefined allowlist of acceptable licenses. It mimics the behavior
    of the CI license scanning job to catch license violations early during
    local development.

    Allowed licenses:
      - Apache-2.0
      - BSD-2-Clause
      - BSD-3-Clause
      - MIT
      - ISC
      - Unlicense

EXIT CODES:
    0 - All licenses are compliant
    1 - License violations found or tool error
    2 - Invalid arguments or missing dependencies
EOF
}

# Parse command line arguments
VERBOSE=false
REPORT_ONLY=false
OUTPUT_FILE="licenses.csv"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --report-only)
            REPORT_ONLY=true
            shift
            ;;
        --output-file)
            if [[ -n $2 && $2 != -* ]]; then
                OUTPUT_FILE="$2"
                shift 2
            else
                log_error "Option --output-file requires a filename argument"
                exit 2
            fi
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 2
            ;;
    esac
done

# Check if we're in a Go project
if [[ ! -f "go.mod" ]]; then
    log_error "No go.mod found. Please run this script from the root of a Go project."
    exit 2
fi

# Check if go-licenses is installed
if ! command -v go-licenses >/dev/null 2>&1; then
    log_warning "go-licenses not found. Installing..."

    # Install go-licenses with version matching CI configuration
    if ! go install github.com/google/go-licenses@v1.6.0; then
        log_error "Failed to install go-licenses. Please install it manually:"
        log_error "  go install github.com/google/go-licenses@v1.6.0"
        exit 2
    fi

    # Verify installation
    if ! command -v go-licenses >/dev/null 2>&1; then
        log_error "go-licenses installation failed or not in PATH"
        log_error "Make sure \$GOPATH/bin or \$GOBIN is in your PATH"
        exit 2
    fi

    log_success "go-licenses installed successfully"
fi

# Define allowed licenses (must match CI configuration)
ALLOWED_LICENSES=(
    "Apache-2.0"
    "BSD-2-Clause"
    "BSD-3-Clause"
    "MIT"
    "ISC"
    "Unlicense"
)

log_info "Checking dependency license compliance..."

if [[ "$VERBOSE" == "true" ]]; then
    log_info "Project root: $PROJECT_ROOT"
    log_info "Output file: $OUTPUT_FILE"
    log_info "Allowed licenses: ${ALLOWED_LICENSES[*]}"
    log_info "go-licenses version: $(go-licenses --version 2>/dev/null || echo 'version unknown')"
fi

# Generate license report
log_info "Generating license report using go-licenses..."

if ! go-licenses csv . > "$OUTPUT_FILE" 2>/dev/null; then
    log_error "Failed to generate license report"
    log_error "This could be due to:"
    log_error "  - Network connectivity issues"
    log_error "  - Invalid go.mod or go.sum"
    log_error "  - Missing or private dependencies"
    log_error ""
    log_error "Try running: go mod tidy && go mod download"
    exit 1
fi

# Check if the report is empty or only contains header
if [[ ! -s "$OUTPUT_FILE" ]] || [[ $(wc -l < "$OUTPUT_FILE") -le 1 ]]; then
    log_warning "License report is empty or contains no dependencies"
    log_info "This might be expected for projects with no external dependencies"
    exit 0
fi

log_success "License report generated: $OUTPUT_FILE"

# Initialize violation tracking
VIOLATIONS_FOUND=false
VIOLATION_COUNT=0
TOTAL_PACKAGES=0

# Process each line in the license report
log_info "Analyzing license compliance..."

while IFS=, read -r package license_url license_type; do
    # Skip header line and empty lines
    if [[ "$license_type" == "license" ]] || [[ -z "$license_type" ]] || [[ -z "$package" ]]; then
        continue
    fi

    TOTAL_PACKAGES=$((TOTAL_PACKAGES + 1))

    # Check if license is in the allowlist
    allowed=false
    for allowed_license in "${ALLOWED_LICENSES[@]}"; do
        if [[ "$license_type" == "$allowed_license" ]]; then
            allowed=true
            break
        fi
    done

    if [[ "$allowed" == "false" ]]; then
        VIOLATIONS_FOUND=true
        VIOLATION_COUNT=$((VIOLATION_COUNT + 1))
        log_error "FORBIDDEN LICENSE: $package uses $license_type"

        if [[ "$VERBOSE" == "true" ]]; then
            log_error "  License URL: $license_url"
        fi
    else
        if [[ "$VERBOSE" == "true" ]]; then
            log_success "$package: $license_type (allowed)"
        fi
    fi
done < "$OUTPUT_FILE"

# Report results
echo ""
log_info "License compliance check complete"
log_info "Total packages analyzed: $TOTAL_PACKAGES"

if [[ "$VIOLATIONS_FOUND" == "true" ]]; then
    log_error "License violations found: $VIOLATION_COUNT"
    echo ""
    log_error "Only the following licenses are allowed:"
    for license in "${ALLOWED_LICENSES[@]}"; do
        log_error "  - $license"
    done
    echo ""
    log_error "Please replace packages with forbidden licenses or get approval to add"
    log_error "the license to the allowlist in .github/workflows/security-gates.yml"

    if [[ "$REPORT_ONLY" == "true" ]]; then
        log_warning "Report-only mode: exiting with success despite violations"
        exit 0
    else
        exit 1
    fi
else
    log_success "All dependency licenses are compliant!"
    log_success "No forbidden licenses detected in $TOTAL_PACKAGES packages"
fi

# Clean up if verbose mode is off and we succeeded
if [[ "$VERBOSE" == "false" ]] && [[ "$VIOLATIONS_FOUND" == "false" ]] && [[ "$OUTPUT_FILE" == "licenses.csv" ]]; then
    rm -f "$OUTPUT_FILE"
fi
