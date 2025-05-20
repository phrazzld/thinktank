# Makefile for thinktank project
# This file contains common development tasks

# Ensure we use bash for shell commands
SHELL := /bin/bash

# Project variables
BINARY_NAME := thinktank
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

.PHONY: help tools hooks build test lint fmt clean vendor coverage cover-report commit

# Display help information about available targets
help:
	@echo "Thinktank Makefile"
	@echo "=================="
	@echo "Available targets:"
	@echo "  help       - Display this help message"
	@echo "  tools      - Install required development tools and git hooks"
	@echo "  hooks      - Install pre-commit hooks"
	@echo "  commit     - Run guided commit creation with Commitizen"
	@echo "  build      - Build the project"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  fmt        - Format Go code"
	@echo "  clean      - Remove build artifacts"
	@echo "  vendor     - Update vendor directory"
	@echo "  coverage   - Run tests with coverage"
	@echo "  cover-report - Generate HTML coverage report"

# Install pre-commit hooks
hooks:
	@echo "Installing pre-commit hooks..."
	@if ! command -v pre-commit &> /dev/null; then \
		echo "Error: pre-commit is not installed. Install it with: pip install pre-commit"; \
		exit 1; \
	fi
	@# Handle custom hooks path if set
	@if git config core.hooksPath > /dev/null 2>&1; then \
		echo "Removing custom hooks path to allow pre-commit installation..."; \
		git config --unset-all core.hooksPath; \
	fi
	@pre-commit install --install-hooks
	@pre-commit install --hook-type commit-msg
	@pre-commit install --hook-type post-commit
	@echo "Pre-commit hooks installed successfully."
	@echo "All code formatting will be automatically checked and fixed on commit."

# Install required Go tools from tools.go
tools: hooks
	@echo "Installing development tools..."
	@grep -E "_ \".*\"" tools.go | sed -E 's/.*_ "(.*)"$$/\1/' | while read pkg; do \
		if [ "$$pkg" = "github.com/leodido/go-conventionalcommits" ]; then \
			echo "Note: $$pkg is a library, not a CLI tool, skipping go install"; \
		else \
			go install $$pkg@latest; \
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

# Default target is help
.DEFAULT_GOAL := help
