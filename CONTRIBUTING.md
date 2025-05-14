# Contributing to Thinktank

Thank you for your interest in contributing to Thinktank! This document provides detailed information about the development environment setup, coding standards, testing requirements, and workflow processes.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
  - [Prerequisites](#prerequisites)
  - [Tools Installation](#tools-installation)
  - [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

## Development Environment Setup

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: The project requires Go version 1.21 or later. [Download Go](https://golang.org/dl/)
- **Git**: For version control. [Download Git](https://git-scm.com/downloads)
- **Make**: Required for running Makefile commands

### Tools Installation

This project uses several tools for development, testing, and code quality. We maintain a `tools.go` file to pin tool dependencies and ensure consistent versions across all development environments.

#### Automatic Installation (Recommended)

The simplest way to install all required tools is through our Makefile:

```bash
# Clone the repository if you haven't already
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install all development tools
make tools
```

This command parses the `tools.go` file and installs all required tools automatically.

#### Manual Installation

If you prefer to install tools manually, you can run:

```bash
# Install tools directly using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/caarlos0/svu@latest
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
```

#### PATH Configuration

Ensure that your `$GOPATH/bin` directory is in your PATH to access the installed tools:

```bash
# Add this to your .bashrc, .zshrc, or equivalent configuration file
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Tool Verification

To verify that all tools are installed correctly:

```bash
# Check that the tools are available in your PATH
golangci-lint --version
govulncheck --help
svu --version
git-chglog --version
```

### Development Workflow

Our project provides several Make commands to streamline development:

```bash
make help           # Display all available commands
make build          # Build the project
make test           # Run all tests
make test-short     # Run short tests (faster)
make lint           # Run the linter
make fmt            # Format code
make coverage       # Run tests with coverage
make cover-report   # Generate HTML coverage report
make clean          # Clean build artifacts
make vendor         # Update vendor directory
```

## Coding Standards

We follow strict coding standards in this project, as detailed in our [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md) document. Key points include:

- Follow Go's official [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Ensure proper error handling is implemented
- Write clear, concise, self-documenting code
- Document the "why" (rationale), not just the "how" (mechanics)
- Maintain high test coverage

## Testing Requirements

Testing is a core component of our development process:

- **Test-Driven Development (TDD)**: Write tests before implementing features
- **Coverage Requirements**: Maintain at least 75% overall code coverage, with a 90% target
- **Testing Approaches**:
  - Unit tests for isolated logic
  - Integration tests for component interactions
  - End-to-end tests for critical workflows
- **Mocking Policy**: Only mock true external dependencies, never internal collaborators

See the README section on [Code Coverage Requirements](README.md#code-coverage-requirements) for detailed information.

## Pull Request Process

1. Ensure your code passes all tests and linting checks
2. Update documentation as needed
3. Submit a pull request with a clear description of the changes
4. Wait for reviews and address any feedback

## Project Structure

The project follows a standard Go project layout:

- `/cmd/thinktank`: Main application entry point
- `/internal`: Private application code
- `/docs`: Documentation files
- `/scripts`: Utility scripts for development and CI
- `/config`: Configuration files
- `/vendor`: Vendored dependencies

For more details on our project structure and philosophy, see the [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md) document.

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go FAQ](https://golang.org/doc/faq)
