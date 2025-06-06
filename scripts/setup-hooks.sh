#!/bin/bash

set -e

echo "Setting up Git hooks for thinktank..."

# Install pre-commit if not available
if ! command -v pre-commit &> /dev/null; then
    echo "Installing pre-commit..."
    pip install pre-commit
fi

# Install pre-commit hooks
echo "Installing pre-commit hooks..."
pre-commit install

# Set up pre-push hook
echo "Setting up pre-push hook..."
cat > .git/hooks/pre-push << 'EOF'
#!/bin/bash

set -e

echo "Running pre-push checks..."

# Run comprehensive linting
echo "🔍 Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout=10m
else
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
    $(go env GOPATH)/bin/golangci-lint run --timeout=10m
fi

# Run build check
echo "🔨 Building project..."
go build ./...

# Run comprehensive tests
echo "🧪 Running tests with race detection..."
go test -race -short $(go list ./... | grep -v 'internal/e2e')

# Check coverage
echo "📊 Checking test coverage..."
if [ -f "./scripts/check-coverage.sh" ]; then
    ./scripts/check-coverage.sh 75
else
    echo "Coverage check script not found, skipping..."
fi

# Check for vulnerabilities
echo "🔒 Checking for vulnerabilities..."
if command -v govulncheck &> /dev/null; then
    govulncheck ./...
else
    echo "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
fi

echo "✅ All pre-push checks passed!"
EOF

chmod +x .git/hooks/pre-push

echo "✅ Git hooks setup complete!"
echo ""
echo "Hooks configured:"
echo "  - Pre-commit: Fast checks (formatting, basic validation)"
echo "  - Pre-push: Comprehensive checks (tests, linting, coverage)"
echo ""
echo "To skip pre-commit hooks: git commit --no-verify"
echo "To skip pre-push hooks: git push --no-verify"
