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

# Create validate-pr.sh if it doesn't exist
if [ ! -f "scripts/validate-pr.sh" ]; then
    # Create the validation script
    cat > scripts/validate-pr.sh << EOL
#!/bin/bash
# validate-pr.sh - Validate PR commits against conventional commits standard
# Only validates commits made after the baseline commit
#
# Usage:
#   ./scripts/validate-pr.sh [branch] [base-branch]
#
# Examples:
#   ./scripts/validate-pr.sh                  # Check current branch against master
#   ./scripts/validate-pr.sh feature/foo      # Check feature/foo branch against master
#   ./scripts/validate-pr.sh feature/foo main # Check feature/foo branch against main

# The baseline commit SHA when conventional commits were established
BASELINE_COMMIT="${BASELINE_COMMIT}"
BRANCH=\${1:-HEAD}
BASE_BRANCH=\${2:-master}

# Format text
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Print header
echo -e "\${BOLD}PR Validation - Conventional Commits\${NC}"
echo "Branch to validate: \${BRANCH}"
echo "Base branch: \${BASE_BRANCH}"
echo "Baseline commit: \${BASELINE_COMMIT} (May 18, 2025)"
echo "Only commits made after the baseline will be validated."
echo "--------------------------------------------------------"

# Check if commitlint is available
if ! command -v commitlint &> /dev/null; then
    echo -e "\${RED}Error: commitlint is not installed\${NC}"
    echo "Install with: npm install -g @commitlint/cli @commitlint/config-conventional"
    exit 1
fi

# Check that the branch exists
if ! git rev-parse --verify \${BRANCH} &> /dev/null; then
    echo -e "\${RED}Error: Branch '\${BRANCH}' does not exist\${NC}"
    exit 1
fi

# Check that the base branch exists
if ! git rev-parse --verify \${BASE_BRANCH} &> /dev/null; then
    echo -e "\${RED}Error: Base branch '\${BASE_BRANCH}' does not exist\${NC}"
    exit 1
fi

# Get the commits to validate (only those after the baseline)
echo "Getting commits to validate..."
COMMITS=\$(git rev-list \${BASELINE_COMMIT}..\${BRANCH} --not \${BASE_BRANCH})

if [ -z "\${COMMITS}" ]; then
    echo -e "\${YELLOW}No commits to validate after baseline.\${NC}"
    echo "This could mean:"
    echo "  1. All commits in the PR were made before the baseline date"
    echo "  2. There are no commits unique to this branch"
    echo "  3. This branch is based on a version before the baseline"
    echo ""
    echo -e "\${GREEN}Validation completed: No issues found (no applicable commits)\${NC}"
    exit 0
fi

# Count commits to validate
COMMIT_COUNT=\$(echo "\${COMMITS}" | wc -l | tr -d ' ')
echo -e "\${BOLD}Found \${COMMIT_COUNT} commits to validate\${NC}"

# Validate each commit
FAILED=0
for COMMIT in \${COMMITS}; do
    COMMIT_MSG=\$(git log --format=%B -n 1 \${COMMIT})
    COMMIT_SHORT=\$(git log --format=%h -n 1 \${COMMIT})
    COMMIT_DATE=\$(git log --format=%ci -n 1 \${COMMIT})
    COMMIT_AUTHOR=\$(git log --format="%an <%ae>" -n 1 \${COMMIT})

    echo ""
    echo -e "\${BOLD}Checking commit \${COMMIT_SHORT}\${NC}"
    echo "Date:   \${COMMIT_DATE}"
    echo "Author: \${COMMIT_AUTHOR}"

    # Use echo to pipe the commit message to commitlint
    if echo "\${COMMIT_MSG}" | commitlint; then
        echo -e "\${GREEN}✓ Valid conventional commit\${NC}"
    else
        echo -e "\${RED}✗ Invalid commit format\${NC}"
        echo ""
        echo "Commit message:"
        echo "--------------"
        echo "\${COMMIT_MSG}"
        echo "--------------"
        echo ""
        echo "Fix tips:"
        echo "  1. Format should be: <type>[optional scope]: <description>"
        echo "  2. Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf"
        echo "  3. Use 'git commit --amend' to fix the most recent commit"
        echo "  4. Use 'git rebase -i \${BASE_BRANCH}' to fix older commits"
        echo ""
        FAILED=1
    fi
done

echo ""
if [ \${FAILED} -eq 0 ]; then
    echo -e "\${GREEN}Validation successful: All commits follow conventional commit format\${NC}"
    echo "Your PR is ready to be submitted."
    exit 0
else
    echo -e "\${RED}Validation failed: Some commits do not follow conventional commit format\${NC}"
    echo "Please fix the commit messages before submitting your PR."
    echo "For more details, see docs/conventional-commits.md"
    exit 1
fi
EOL
    chmod +x scripts/validate-pr.sh
    print_success "Created PR validation script (scripts/validate-pr.sh)"
else
    print_success "PR validation script (scripts/validate-pr.sh) already exists"
fi

print_success "Created baseline-aware commitlint configuration"
echo ""
echo "You can now validate your commits with:"
echo "  ./scripts/validate-pr.sh           # Validate PR commits"
echo ""
echo "The pre-push hook will automatically validate commits before pushing"
echo "using the same baseline-aware validation approach."
