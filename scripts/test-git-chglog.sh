#!/bin/bash
# Test script for verifying git-chglog output for all commit types

set -e

# Output file
RESULTS_FILE="docs/git-chglog-test-results.md"

# Create test repository
create_test_repo() {
    echo "Creating test repository..."
    TEST_DIR="test-repo-chglog"

    rm -rf "$TEST_DIR"
    mkdir "$TEST_DIR"
    cd "$TEST_DIR"

    git init
    git config user.email "test@example.com"
    git config user.name "Test User"

    # Copy the git-chglog config from parent
    mkdir -p .chglog
    cp ../.chglog/config.yml .chglog/
    cp ../.chglog/CHANGELOG.tpl.md .chglog/

    # Create initial commit
    echo "initial" > file.txt
    git add .
    git commit -m "chore: initial commit"

    # Create initial tag
    git tag v0.0.0
}

# Create commits for all types
create_test_commits() {
    echo "Creating test commits..."

    # Feature commits
    echo "feature1" >> file.txt
    git add file.txt
    git commit -m "feat: add user authentication"

    echo "feature2" >> file.txt
    git add file.txt
    git commit -m "feat(api): add new REST endpoint"

    # Fix commits
    echo "fix1" >> file.txt
    git add file.txt
    git commit -m "fix: resolve memory leak"

    echo "fix2" >> file.txt
    git add file.txt
    git commit -m "fix(auth): correct token validation"

    # Performance improvements
    echo "perf1" >> file.txt
    git add file.txt
    git commit -m "perf: optimize database queries"

    echo "perf2" >> file.txt
    git add file.txt
    git commit -m "perf(cache): improve cache hit ratio"

    # Refactoring
    echo "refactor1" >> file.txt
    git add file.txt
    git commit -m "refactor: extract common utilities"

    echo "refactor2" >> file.txt
    git add file.txt
    git commit -m "refactor(core): reorganize module structure"

    # Documentation
    echo "docs1" >> file.txt
    git add file.txt
    git commit -m "docs: update API documentation"

    echo "docs2" >> file.txt
    git add file.txt
    git commit -m "docs(readme): add installation guide"

    # Style changes
    echo "style1" >> file.txt
    git add file.txt
    git commit -m "style: format code with prettier"

    echo "style2" >> file.txt
    git add file.txt
    git commit -m "style(css): update color scheme"

    # Test changes
    echo "test1" >> file.txt
    git add file.txt
    git commit -m "test: add unit tests for auth module"

    echo "test2" >> file.txt
    git add file.txt
    git commit -m "test(e2e): add integration tests"

    # Build system
    echo "build1" >> file.txt
    git add file.txt
    git commit -m "build: update webpack configuration"

    echo "build2" >> file.txt
    git add file.txt
    git commit -m "build(deps): upgrade dependencies"

    # CI/CD
    echo "ci1" >> file.txt
    git add file.txt
    git commit -m "ci: add GitHub Actions workflow"

    echo "ci2" >> file.txt
    git add file.txt
    git commit -m "ci(docker): optimize Docker build"

    # Chores
    echo "chore1" >> file.txt
    git add file.txt
    git commit -m "chore: update .gitignore"

    echo "chore2" >> file.txt
    git add file.txt
    git commit -m "chore(deps): bump lodash version"

    # Breaking change
    echo "breaking1" >> file.txt
    git add file.txt
    git commit -m "feat!: redesign API authentication

BREAKING CHANGE: API keys are now required for all endpoints"

    # Tag a version
    git tag v1.0.0

    # Regular commit after tag
    echo "post-tag" >> file.txt
    git add file.txt
    git commit -m "fix: post-release bug fix"

    # Revert commit
    git revert HEAD --no-edit

    # Another version tag
    git tag v1.0.1
}

# Generate changelog
generate_changelog() {
    echo "Generating changelog..."

    # Run git-chglog for all versions
    git-chglog > CHANGELOG_FULL.md

    # Run git-chglog for latest version only
    git-chglog v1.0.0..v1.0.1 > CHANGELOG_LATEST.md

    # Copy to results directory
    cp CHANGELOG_FULL.md "../$RESULTS_FILE"

    # Add extra documentation to results
    cat >> "../$RESULTS_FILE" << 'EOF'

## Test Commit Summary

This changelog was generated from a test repository with the following commit types:

| Commit Type | Count | Example |
|-------------|-------|---------|
| feat | 3 | feat: add user authentication |
| fix | 3 | fix: resolve memory leak |
| perf | 2 | perf: optimize database queries |
| refactor | 2 | refactor: extract common utilities |
| docs | 2 | docs: update API documentation |
| style | 2 | style: format code with prettier |
| test | 2 | test: add unit tests for auth module |
| build | 2 | build: update webpack configuration |
| ci | 2 | ci: add GitHub Actions workflow |
| chore | 3 | chore: update .gitignore |
| revert | 1 | Revert "fix: post-release bug fix" |

## Breaking Changes

One breaking change was included:
- `feat!: redesign API authentication` with BREAKING CHANGE note

## Version Tags

- v0.0.0 (initial)
- v1.0.0 (after all test commits)
- v1.0.1 (after revert)

## Latest Version Changelog

EOF

    cat CHANGELOG_LATEST.md >> "../$RESULTS_FILE"
}

# Main execution
main() {
    echo "Starting git-chglog test..."

    # Create test repo and commits
    create_test_repo
    create_test_commits
    generate_changelog

    # Cleanup
    cd ..
    rm -rf test-repo-chglog

    echo "Test complete! Results written to $RESULTS_FILE"
}

# Run the test
main
