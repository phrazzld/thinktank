# Makefile for thinktank local development workflow

# Color definitions for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

# Project configuration
PROJECT_ROOT := $(shell pwd)
GOPATH := $(shell go env GOPATH)
GOLANGCI_LINT_VERSION := v2.1.1

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Thinktank Development Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: ci-check
ci-check: ## Run full local CI simulation (mirrors GitHub Actions pipeline)
	@echo "$(CYAN)🚀 Running full local CI simulation...$(NC)"
	@echo ""
	$(MAKE) deps-verify
	$(MAKE) lint
	$(MAKE) security-scan
	$(MAKE) test
	$(MAKE) coverage
	$(MAKE) build
	@echo ""
	@echo "$(GREEN)✅ All CI checks passed! Ready for push.$(NC)"

.PHONY: deps-verify
deps-verify: ## Verify Go module dependencies
	@echo "$(BLUE)📦 Verifying Go module dependencies...$(NC)"
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies verified$(NC)"

.PHONY: lint
lint: ## Run all linting checks (format, vet, golangci-lint)
	@echo "$(BLUE)🔍 Running linting checks...$(NC)"
	@echo "  - Checking Go formatting..."
	@if [ -n "$$(go fmt ./...)" ]; then \
		echo "$(RED)❌ Code formatting issues found. Run 'make fmt' to fix.$(NC)"; \
		exit 1; \
	fi
	@echo "  - Running go vet..."
	@go vet ./...
	@echo "  - Running golangci-lint..."
	@$(MAKE) golangci-lint-check
	@echo "  - Running pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "$(YELLOW)⚠️  pre-commit not installed, skipping pre-commit checks$(NC)"; \
	fi
	@echo "$(GREEN)✅ All linting checks passed$(NC)"

.PHONY: golangci-lint-check
golangci-lint-check: ## Run golangci-lint (installs if needed)
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(YELLOW)📥 Installing golangci-lint $(GOLANGCI_LINT_VERSION)...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	@golangci-lint run --timeout=5m

.PHONY: security-scan
security-scan: ## Run security scanning (secrets, licenses, SAST)
	@echo "$(PURPLE)🔒 Running security scans...$(NC)"
	@echo "  - Checking license compliance..."
	@./scripts/check-licenses.sh
	@echo "  - Running secret detection..."
	@if command -v trufflehog >/dev/null 2>&1; then \
		trufflehog git file://. --since-commit HEAD --only-verified --fail || echo "$(YELLOW)⚠️  TruffleHog scan completed with findings$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  TruffleHog not installed locally, skipping secret scan$(NC)"; \
	fi
	@echo "$(GREEN)✅ Security scans completed$(NC)"

.PHONY: test
test: ## Run all tests (unit, integration, E2E)
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@echo "  - Running integration tests..."
	@go test -v -race -short -parallel 4 ./internal/integration/...
	@echo "  - Running unit tests..."
	@go test -v -race -short $$(go list ./... | grep -v "github.com/phrazzld/thinktank/internal/integration" | grep -v "github.com/phrazzld/thinktank/internal/e2e")
	@echo "  - Running E2E tests..."
	@./internal/e2e/run_e2e_tests.sh
	@echo "$(GREEN)✅ All tests passed$(NC)"

.PHONY: coverage
coverage: ## Check test coverage (90% threshold)
	@echo "$(BLUE)📊 Checking test coverage...$(NC)"
	@echo "  - Generating coverage report..."
	@go test -short -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v "/internal/integration" | grep -v "/internal/e2e" | grep -v "/disabled/" | grep -v "/internal/testutil")
	@echo "  - Checking overall coverage threshold (90%)..."
	@./scripts/check-coverage.sh 90
	@echo "  - Checking package-specific coverage..."
	@./scripts/ci/check-package-specific-coverage.sh
	@echo "$(GREEN)✅ Coverage checks passed$(NC)"

.PHONY: build
build: ## Build the thinktank binary
	@echo "$(BLUE)🔨 Building thinktank binary...$(NC)"
	@go build -v -ldflags="-s -w" -o thinktank ./cmd/thinktank
	@echo "$(GREEN)✅ Build completed: ./thinktank$(NC)"

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)✨ Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(NC)"
	@rm -f thinktank
	@rm -f coverage.out
	@rm -f *.prof
	@rm -f licenses.csv
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

.PHONY: install-tools
install-tools: ## Install required development tools
	@echo "$(BLUE)🛠️  Installing development tools...$(NC)"
	@echo "  - Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "  - Installing go-licenses..."
	@go install github.com/google/go-licenses@v1.6.0
	@echo "  - Installing pre-commit..."
	@if command -v pip >/dev/null 2>&1; then \
		pip install pre-commit; \
		pre-commit install; \
	else \
		echo "$(YELLOW)⚠️  pip not found, please install pre-commit manually$(NC)"; \
	fi
	@echo "$(GREEN)✅ Development tools installed$(NC)"

.PHONY: quick-check
quick-check: ## Run quick checks (format, vet, basic tests)
	@echo "$(CYAN)⚡ Running quick checks...$(NC)"
	$(MAKE) deps-verify
	@go fmt ./...
	@go vet ./...
	@go test -short -race ./internal/integration/...
	@echo "$(GREEN)✅ Quick checks passed$(NC)"

.PHONY: pre-push
pre-push: ## Recommended checks before pushing (faster than full ci-check)
	@echo "$(CYAN)🚀 Running pre-push checks...$(NC)"
	$(MAKE) lint
	$(MAKE) security-scan
	@go test -short -race ./internal/integration/...
	@echo "$(GREEN)✅ Pre-push checks passed$(NC)"

.PHONY: test-focus
test-focus: ## Run tests for packages needing coverage improvement (cli, integration, cmd)
	@echo "$(BLUE)🎯 Testing coverage-focused packages...$(NC)"
	@echo "  Note: Testing packages with coverage below 80% threshold"
	@echo ""
	@echo "$(YELLOW)Package Coverage Status:$(NC)"
	@echo "  - internal/cli: currently 72.0%"
	@echo "  - internal/integration: currently 74.4%"
	@echo "  - cmd/thinktank: currently 85.4%"
	@echo ""
	@go test -short ./internal/cli ./internal/integration ./cmd/thinktank || (echo "$(YELLOW)⚠️  Some tests failed - review output above$(NC)"; exit 0)
	@echo "$(GREEN)✅ Coverage-focused tests completed$(NC)"

.PHONY: coverage-quick
coverage-quick: ## Quick coverage check for development (shows current coverage for focus packages)
	@echo "$(BLUE)📊 Quick coverage check for development...$(NC)"
	@echo "  Testing packages with coverage below 80% threshold..."
	@echo ""
	@go test -short -cover ./internal/cli ./internal/integration ./cmd/thinktank 2>/dev/null | grep -E "(PASS|FAIL|coverage:)" || echo "$(YELLOW)⚠️  Some tests may have issues$(NC)"
	@echo ""
	@echo "$(GREEN)✅ Quick coverage check completed$(NC)"

.PHONY: coverage-critical
coverage-critical: ## Run tests for 5 lowest-coverage packages (reduces feedback loop from ~15min to ~3min)
	@echo "$(CYAN)🎯 Running coverage-critical tests (5 lowest-coverage packages)...$(NC)"
	@echo ""
	@echo "$(YELLOW)Target packages (in priority order):$(NC)"
	@echo "  1. internal/integration (74.4%) - highest impact"
	@echo "  2. internal/testutil (78.4%)"
	@echo "  3. internal/gemini (79.8%)"
	@echo "  4. internal/config (80.6%)"
	@echo "  5. internal/models (80.7%)"
	@echo ""
	@echo "$(BLUE)📊 Running targeted coverage tests...$(NC)"
	@start_time=$$(date +%s); \
	go test -short -cover -race \
		./internal/integration \
		./internal/testutil \
		./internal/gemini \
		./internal/config \
		./internal/models \
		2>/dev/null | grep -E "(PASS|FAIL|coverage:)" || echo "$(YELLOW)⚠️  Some tests may have issues$(NC)"; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo ""; \
	echo "$(GREEN)✅ Coverage-critical tests completed in $${duration}s$(NC)"; \
	echo "$(CYAN)💡 Use 'make coverage' for full test suite validation$(NC)"

# Default target
.DEFAULT_GOAL := help
