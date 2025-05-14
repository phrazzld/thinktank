# Makefile for thinktank project
# This file contains common development tasks

# Ensure we use bash for shell commands
SHELL := /bin/bash

# Project variables
BINARY_NAME := thinktank
GO_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

.PHONY: help tools build test lint fmt clean vendor coverage cover-report

# Display help information about available targets
help:
	@echo "Thinktank Makefile"
	@echo "=================="
	@echo "Available targets:"
	@echo "  help       - Display this help message"
	@echo "  tools      - Install required development tools"
	@echo "  build      - Build the project"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  fmt        - Format Go code"
	@echo "  clean      - Remove build artifacts"
	@echo "  vendor     - Update vendor directory"
	@echo "  coverage   - Run tests with coverage"
	@echo "  cover-report - Generate HTML coverage report"

# Install required Go tools from tools.go
tools:
	@echo "Installing development tools..."
	@grep -E "_ \".*\"" tools.go | sed -E 's/.*_ "(.*)"$$/\1/' | xargs -I{} go install {}@latest
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

# Default target is help
.DEFAULT_GOAL := help
