# Makefile for thinktank project
# This file contains common development tasks

# Ensure we use bash for shell commands
SHELL := /bin/bash

# Project variables
BINARY_NAME := thinktank
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

.PHONY: help tools hooks hooks-validate hooks-clean hooks-status build test lint fmt clean vendor coverage cover-report commit vuln-check vuln-check-ci

# Display help information about available targets
help:
	@echo "Thinktank Makefile"
	@echo "=================="
	@echo "Available targets:"
	@echo "  help         - Display this help message"
	@echo "  tools        - Install required development tools and git hooks"
	@echo "  hooks        - Install pre-commit hooks (all types)"
	@echo "  hooks-validate - Validate hook configuration without installing"
	@echo "  hooks-clean  - Remove all hooks for fresh installation"
	@echo "  hooks-status - Show current hook installation status"
	@echo "  commit       - Run guided commit creation with Commitizen"
	@echo "  build      - Build the project"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  fmt        - Format Go code"
	@echo "  clean      - Remove build artifacts"
	@echo "  vendor     - Update vendor directory"
	@echo "  coverage   - Run tests with coverage"
	@echo "  cover-report - Generate HTML coverage report"
	@echo "  vuln-check - Run vulnerability scan"
	@echo "  vuln-check-ci - Run vulnerability scan with CI output format"

# Install pre-commit hooks
hooks:
	@echo "Installing pre-commit hooks..."
	@# Check pre-commit installation and version
	@if ! command -v pre-commit &> /dev/null; then \
		echo "ERROR: pre-commit is not installed."; \
		echo ""; \
		echo "Install pre-commit using one of these methods:"; \
		echo "  pip install pre-commit"; \
		echo "  pipx install pre-commit"; \
		echo "  brew install pre-commit"; \
		echo "  conda install -c conda-forge pre-commit"; \
		echo ""; \
		echo "For more options: https://pre-commit.com/#install"; \
		exit 1; \
	fi
	@# Check pre-commit version
	@MIN_VERSION="3.0.0"; \
	CURRENT_VERSION=$$(pre-commit --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ "$$(printf '%s\n' "$$MIN_VERSION" "$$CURRENT_VERSION" | sort -V | head -n1)" != "$$MIN_VERSION" ]; then \
		echo "ERROR: pre-commit version $$CURRENT_VERSION is older than required $$MIN_VERSION"; \
		echo "Please upgrade: pip install --upgrade pre-commit"; \
		exit 1; \
	fi
	@# Handle custom hooks path if set
	@if git config core.hooksPath > /dev/null 2>&1; then \
		echo "Removing custom hooks path to allow pre-commit installation..."; \
		git config --unset-all core.hooksPath; \
	fi
	@# Install hooks with error handling
	@echo "Installing pre-commit hooks..."
	@if ! pre-commit install --install-hooks; then \
		echo "ERROR: Failed to install pre-commit hooks"; \
		exit 1; \
	fi
	@echo "Installing commit-msg hooks..."
	@if ! pre-commit install --hook-type commit-msg; then \
		echo "ERROR: Failed to install commit-msg hooks"; \
		exit 1; \
	fi
	@echo "Installing pre-push hooks..."
	@if ! pre-commit install --hook-type pre-push; then \
		echo "ERROR: Failed to install pre-push hooks"; \
		exit 1; \
	fi
	@echo "Installing post-commit hooks..."
	@if ! pre-commit install --hook-type post-commit; then \
		echo "ERROR: Failed to install post-commit hooks"; \
		exit 1; \
	fi
	@# Verify all hooks are installed and executable
	@echo "Verifying hook installation..."
	@MISSING_HOOKS=""; \
	for hook in pre-commit commit-msg pre-push post-commit; do \
		if [ ! -f .git/hooks/$$hook ]; then \
			MISSING_HOOKS="$$MISSING_HOOKS $$hook"; \
		elif [ ! -x .git/hooks/$$hook ]; then \
			chmod +x .git/hooks/$$hook; \
			echo "Made .git/hooks/$$hook executable"; \
		fi; \
	done; \
	if [ -n "$$MISSING_HOOKS" ]; then \
		echo "ERROR: Failed to install hooks:$$MISSING_HOOKS"; \
		exit 1; \
	fi
	@echo ""
	@echo "Pre-commit hooks installed successfully!"
	@echo "✓ pre-commit: Code formatting and quality checks"
	@echo "✓ commit-msg: Conventional commit validation"
	@echo "✓ pre-push: Commit validation before push"
	@echo "✓ post-commit: Documentation generation"

# Validate hook configuration without installing
hooks-validate:
	@echo "Validating pre-commit configuration..."
	@# Check if .pre-commit-config.yaml exists
	@if [ ! -f .pre-commit-config.yaml ]; then \
		echo "ERROR: .pre-commit-config.yaml not found"; \
		exit 1; \
	fi
	@# Validate YAML syntax
	@if command -v python3 &> /dev/null; then \
		if python3 -c "import yaml" 2>/dev/null; then \
			python3 -c "import yaml; yaml.safe_load(open('.pre-commit-config.yaml'))" 2>&1 || { \
				echo "ERROR: .pre-commit-config.yaml has invalid YAML syntax"; \
				exit 1; \
			}; \
			echo "✓ .pre-commit-config.yaml syntax is valid"; \
		else \
			echo "Note: Python yaml module not available, skipping YAML syntax check"; \
		fi; \
	fi
	@# Check if pre-commit is installed
	@if command -v pre-commit &> /dev/null; then \
		pre-commit validate-config || { \
			echo "ERROR: pre-commit configuration validation failed"; \
			exit 1; \
		}; \
		echo "✓ pre-commit configuration is valid"; \
		pre-commit validate-manifest || { \
			echo "ERROR: pre-commit manifest validation failed"; \
			exit 1; \
		}; \
		echo "✓ pre-commit manifest is valid"; \
	else \
		echo "WARNING: pre-commit not installed, skipping advanced validation"; \
	fi

# Remove all hooks for fresh installation
hooks-clean:
	@echo "Removing all git hooks..."
	@if command -v pre-commit &> /dev/null; then \
		pre-commit uninstall || true; \
		pre-commit uninstall --hook-type commit-msg || true; \
		pre-commit uninstall --hook-type pre-push || true; \
		pre-commit uninstall --hook-type post-commit || true; \
		pre-commit clean || true; \
	fi
	@# Manually remove hook files as backup
	@for hook in pre-commit commit-msg pre-push post-commit; do \
		if [ -f .git/hooks/$$hook ]; then \
			rm -f .git/hooks/$$hook; \
			echo "Removed .git/hooks/$$hook"; \
		fi; \
	done
	@echo "All hooks removed successfully"

# Show current hook installation status
hooks-status:
	@echo "Pre-commit Hook Status"
	@echo "======================"
	@# Check pre-commit installation
	@if command -v pre-commit &> /dev/null; then \
		echo "✓ pre-commit installed: $$(pre-commit --version)"; \
	else \
		echo "✗ pre-commit not installed"; \
	fi
	@# Check each hook file
	@echo ""
	@echo "Git hooks in .git/hooks/:"
	@for hook in pre-commit commit-msg pre-push post-commit; do \
		if [ -f .git/hooks/$$hook ]; then \
			if [ -x .git/hooks/$$hook ]; then \
				echo "✓ $$hook (executable)"; \
			else \
				echo "⚠ $$hook (not executable)"; \
			fi; \
		else \
			echo "✗ $$hook (missing)"; \
		fi; \
	done
	@# Check configuration file
	@echo ""
	@echo "Configuration:"
	@if [ -f .pre-commit-config.yaml ]; then \
		echo "✓ .pre-commit-config.yaml exists"; \
	else \
		echo "✗ .pre-commit-config.yaml missing"; \
	fi

# Install required Go tools from tools.go with specified versions
tools: hooks
	@echo "Installing development tools..."
	@grep -E "_ \".*\"" tools.go | sed -E 's/.*_ \"(.*)[@v].*\"$$/\1/' | while read pkg; do \
		if [[ "$$pkg" == "github.com/leodido/go-conventionalcommits" ]]; then \
			echo "Note: $$pkg is a library, not a CLI tool, skipping go install"; \
		else \
			echo "Installing $$pkg..."; \
			full_pkg=$$(grep -E "_ \"$$pkg.*\"" tools.go | sed -E 's/.*_ \"(.*)"$$/\1/'); \
			go install "$$full_pkg"; \
		fi \
	done
	@echo "Tools installed successfully."
	@echo "Note: Ensure $(shell go env GOPATH)/bin is in your PATH to use these tools."

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./...
	@echo "Build complete."

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with short flag
test-short:
	@echo "Running short tests..."
	@go test -short ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Generate HTML coverage report
cover-report: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: make tools"; \
		exit 1; \
	fi

# Format Go code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html

# Update vendor directory
vendor:
	@echo "Updating vendor directory..."
	@go mod vendor
	@go mod tidy

# Run Commitizen for guided commit creation
commit:
	@echo "Starting guided commit creation..."
	@./scripts/commit.sh

# Run vulnerability scan
vuln-check:
	@echo "Running vulnerability scan..."
	@if command -v govulncheck &> /dev/null; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@v1.1.4"; \
		exit 1; \
	fi

# Run vulnerability scan with CI output format
vuln-check-ci:
	@echo "Running vulnerability scan (CI mode)..."
	@if command -v govulncheck &> /dev/null; then \
		govulncheck -json ./... | ./scripts/ci/parse-govulncheck.sh; \
	else \
		echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@v1.1.4"; \
		exit 1; \
	fi

# Default target is help
.DEFAULT_GOAL := help
