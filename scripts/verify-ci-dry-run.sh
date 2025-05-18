#!/bin/bash
# Script to help verify CI dry-run behavior

set -e

echo "CI Pipeline Dry-Run Verification Script"
echo "======================================"
echo

# Function to check recent git tags
check_recent_tags() {
    echo "Recent Git Tags (last 5):"
    git tag --sort=-version:refname | head -5 || echo "No tags found"
    echo
}

# Function to check for goreleaser artifacts
check_goreleaser_artifacts() {
    echo "Checking for local goreleaser artifacts:"
    if [ -d "dist" ]; then
        echo "Found dist directory:"
        ls -la dist/ | head -10
    else
        echo "No dist directory found (expected if running in CI)"
    fi
    echo
}

# Function to run goreleaser in snapshot mode locally
test_snapshot_locally() {
    echo "Testing goreleaser snapshot locally:"
    echo "Note: This requires goreleaser to be installed"

    if command -v goreleaser &> /dev/null; then
        echo "Running goreleaser check..."
        goreleaser check || echo "Goreleaser check failed"

        echo
        echo "Running goreleaser snapshot (dry-run)..."
        # Don't actually run this - just show the command
        echo "Command that CI runs: goreleaser release --snapshot --clean"
        echo "This would generate artifacts without creating a release"
    else
        echo "Goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"
    fi
    echo
}

# Function to check workflow file
check_workflow() {
    echo "Verifying workflow configuration:"
    if [ -f ".github/workflows/release.yml" ]; then
        echo "✅ Release workflow found"

        # Check for snapshot configuration
        if grep -q "goreleaser release --snapshot" .github/workflows/release.yml; then
            echo "✅ Snapshot mode configured for PR/master"
        else
            echo "❌ Snapshot mode not found"
        fi

        # Check for proper conditions
        if grep -q "github.event_name == 'pull_request'" .github/workflows/release.yml; then
            echo "✅ PR trigger configured"
        fi

        if grep -q "github.ref == 'refs/heads/master'" .github/workflows/release.yml; then
            echo "✅ Master branch trigger configured"
        fi
    else
        echo "❌ Release workflow not found"
    fi
    echo
}

# Main execution
main() {
    check_workflow
    check_recent_tags
    check_goreleaser_artifacts
    test_snapshot_locally

    echo "Verification Complete!"
    echo "===================="
    echo
    echo "To fully verify CI behavior:"
    echo "1. Check GitHub Actions tab for recent PR/master builds"
    echo "2. Verify no releases were created in GitHub Releases"
    echo "3. Check that artifacts are available in Actions"
    echo "4. Review logs to confirm --snapshot flag usage"
}

main
