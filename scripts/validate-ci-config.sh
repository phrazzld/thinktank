#!/bin/bash
# validate-ci-config.sh - Validate GitHub Actions workflow configuration
# Prevents common CI configuration errors before they reach CI

set -euo pipefail

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

error() { echo -e "${RED}❌ ERROR: $1${NC}" >&2; }
warning() { echo -e "${YELLOW}⚠️  WARNING: $1${NC}" >&2; }
info() { echo -e "${GREEN}✅ $1${NC}"; }

[ ! -d ".github/workflows" ] && { info "No .github/workflows directory found"; exit 0; }

echo "🔍 Validating GitHub Actions workflow configuration..."

has_errors=false
has_warnings=false

for file in .github/workflows/*.{yml,yaml}; do
    [ ! -f "$file" ] && continue
    echo "📁 Checking $file"

    # YAML syntax check
    yq eval '.' "$file" >/dev/null 2>&1 || { error "Invalid YAML: $file"; has_errors=true; continue; }

    # TruffleHog duplicate flag check
    grep -q "trufflesecurity/trufflehog" "$file" && grep -q "extra_args:.*--fail" "$file" && {
        error "TruffleHog duplicate --fail flag in $file"
        has_errors=true
    }

    # Missing Dockerfile check
    dockerfiles=$(grep -o "docker/[^[:space:]]*\.Dockerfile" "$file" 2>/dev/null || true)
    for dockerfile in $dockerfiles; do
        [ ! -f "$dockerfile" ] && { error "Missing file: $dockerfile (referenced in $file)"; has_errors=true; }
    done

    # @latest usage check
    grep -q "@latest" "$file" && {
        count=$(grep -c "@latest" "$file")
        warning "File $file uses @latest ($count times) - pin versions for security"
        has_warnings=true
    }
done

echo -e "\n📊 Validation Summary:"
if [ "$has_errors" = false ] && [ "$has_warnings" = false ]; then
    info "All workflow configurations are valid"
elif [ "$has_errors" = false ]; then
    warning "Warnings found - consider fixing for better reliability"
    exit 0
else
    error "Errors found that may cause CI failures"
    exit 1
fi
