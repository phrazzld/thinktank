#!/bin/bash
# setup-commitlint.sh - Configure commitlint with baseline validation
# This script sets up commitlint to use baseline validation

# The baseline commit SHA when conventional commits were established
BASELINE_COMMIT="1300e4d675ac087783199f1e608409e6853e589f"

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}! $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

echo -e "${BOLD}Setting up baseline-aware commit validation${NC}"
echo "Using baseline commit: ${BASELINE_COMMIT} (May 18, 2025)"

# Create a local commitlint configuration file
cat > .commitlint-with-baseline.js << EOL
// Baseline-aware commitlint configuration
// Only validates commits made after policy establishment (${BASELINE_COMMIT})
module.exports = {
  extends: ['./.commitlintrc.yml'],
  // Additional custom rules can be added here
  rules: {},
  // Only validate commits after the baseline commit
  ignores: [
    (commit) => {
      // Always validate during pre-commit hooks
      // This is a placeholder for the actual git history check
      // The full validation with proper baseline comparison is done in CI
      // For local validation, we just use the commit message without checking history
      return false;
    }
  ]
};
EOL

# Create a helper script for validating PRs with baseline
cat > scripts/validate-pr-commits.sh << EOL
#!/bin/bash
# validate-pr-commits.sh - Validate PR commits against conventional commits standard
# Only validates commits made after the baseline commit

BASELINE_COMMIT="${BASELINE_COMMIT}"
BRANCH=\${1:-HEAD}
BASE_BRANCH=\${2:-master}

# Format text
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "\${BOLD}Validating commits in \${BRANCH} against \${BASE_BRANCH}${NC}"
echo "Using baseline commit: \${BASELINE_COMMIT} (May 18, 2025)"
echo "Only commits made after the baseline will be validated"

# Check if commitlint is available
if ! command -v commitlint &> /dev/null; then
    echo -e "\${RED}Error: commitlint is not installed${NC}"
    echo "Install with: npm install -g @commitlint/cli @commitlint/config-conventional"
    exit 1
fi

# Get the commits to validate (only those after the baseline)
COMMITS=\$(git rev-list \${BASELINE_COMMIT}..\${BRANCH})

if [ -z "\${COMMITS}" ]; then
    echo -e "\${GREEN}No commits to validate after baseline${NC}"
    exit 0
fi

# Count commits to validate
COMMIT_COUNT=\$(echo "\${COMMITS}" | wc -l | tr -d ' ')
echo "Found \${COMMIT_COUNT} commits to validate"

# Validate each commit
FAILED=0
for COMMIT in \${COMMITS}; do
    COMMIT_MSG=\$(git log --format=%B -n 1 \${COMMIT})
    COMMIT_SHORT=\$(git log --format=%h -n 1 \${COMMIT})
    COMMIT_DATE=\$(git log --format=%ci -n 1 \${COMMIT})

    echo -e "\${BOLD}Checking commit \${COMMIT_SHORT} (\${COMMIT_DATE})${NC}"

    # Use echo to pipe the commit message to commitlint
    if echo "\${COMMIT_MSG}" | commitlint --config .commitlintrc.yml; then
        echo -e "\${GREEN}✓ Valid conventional commit${NC}"
    else
        echo -e "\${RED}✗ Invalid commit format${NC}"
        echo "Commit message:"
        echo "--------------"
        echo "\${COMMIT_MSG}"
        echo "--------------"
        FAILED=1
    fi
    echo ""
done

if [ \${FAILED} -eq 0 ]; then
    echo -e "\${GREEN}All commits passed validation${NC}"
    exit 0
else
    echo -e "\${RED}Some commits failed validation${NC}"
    echo "Please fix the commit messages or refer to docs/conventional-commits.md"
    exit 1
fi
EOL

# Make the script executable
chmod +x scripts/validate-pr-commits.sh

print_success "Created baseline-aware commitlint configuration"
print_success "Created PR validation script"
echo ""
echo "You can now validate your PR commits with:"
echo "  ./scripts/validate-pr-commits.sh"
