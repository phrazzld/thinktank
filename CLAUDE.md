# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Architect Development Guidelines

## Commands
- **Build**: `go build`
- **Run**: `go run main.go --instructions instructions.md PATH/TO/FILES/OR/DIRS`
- **Test**: `go test ./...`
- **Test Single File**: `go test ./PACKAGE_PATH/FILE_test.go`
- **Lint/Format**: `go fmt ./...`
- **Verify**: `go vet ./...`

## Style Guidelines
- **Imports**: Group standard library imports first, followed by third-party
- **Formatting**: Use `gofmt` standards (4-space indentation)
- **Error Handling**: Always check errors, log with context
- **Naming**:
  - Functions: camelCase for unexported, PascalCase for exported
  - Variables: descriptive, self-documenting names
  - Packages: short, lowercase, no underscores
- **Types**: Use strong typing, avoid empty interfaces when possible
- **Comments**: Document exported functions and types with godoc style
- **Documentation**: Update README.md when adding new features
- **Testing**: Write tests for new functionality, maintain test coverage
