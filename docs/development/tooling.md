# Go Tool Installation Guide

This guide clarifies the distinction between Go libraries and CLI tools, and explains the correct installation methods for each type.

## Understanding Go Packages

In Go, there are two main types of packages:

1. **Libraries**: Packages that provide functionality to be imported and used by other Go code
2. **CLI Tools**: Executable programs with a `main` package that can be run from the command line

## When to Use `go install`

The `go install` command is specifically designed for installing **executable programs** (CLI tools). It:

- Compiles the Go source code
- Creates an executable binary
- Places the binary in `$GOPATH/bin` (or `$GOBIN` if set)

### Requirements for `go install`

For a package to be installable with `go install`, it must:

1. Have a `main` package
2. Contain a `main()` function
3. Be intended as an executable program

## CLI Tools vs Libraries

### CLI Tools (Installable with `go install`)

Examples of Go CLI tools that can be installed:

```bash
# These packages contain main() functions and produce executables
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/caarlos0/svu@latest
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
go install github.com/goreleaser/goreleaser@latest
```

### Libraries (NOT installable with `go install`)

Examples of Go libraries that cannot be installed as CLI tools:

```bash
# These will fail because they are libraries, not executables
go install github.com/stretchr/testify@latest  # Testing library
go install github.com/go-playground/validator/v10@latest  # Validation library
go install github.com/leodido/go-conventionalcommits@latest  # Parsing library
```

## Common Pitfalls and Solutions

### Pitfall 1: Trying to Install a Library

**Problem**: Attempting to use `go install` on a library package.

```bash
# This will fail
go install github.com/leodido/go-conventionalcommits@v0.12.0
# Error: package github.com/leodido/go-conventionalcommits is not a main package
```

**Solution**: Libraries are managed through Go modules. Add them to your project with:

```bash
go get github.com/leodido/go-conventionalcommits@v0.12.0
```

### Pitfall 2: Installing from Wrong Path

**Problem**: Some tools have their `main` package in a subdirectory.

```bash
# Wrong - this is the library root
go install github.com/golangci/golangci-lint@latest

# Correct - this is where the main package lives
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Solution**: Check the repository structure and look for directories containing `main.go` files.

### Pitfall 3: Version Conflicts

**Problem**: Installing tools globally can lead to version conflicts between projects.

**Solution**: Use tools.go to pin tool versions per project:

```go
//go:build tools
// +build tools

package tools

import (
    _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
    _ "github.com/git-chglog/git-chglog/cmd/git-chglog"
)
```

Then install with:

```bash
go install -modfile=tools.go github.com/golangci/golangci-lint/cmd/golangci-lint
```

## Best Practices

1. **Check Package Type**: Before using `go install`, verify the package has a `main` function
2. **Use Module Management**: For libraries, use `go get` and let Go modules handle dependencies
3. **Pin Tool Versions**: Use a `tools.go` file to ensure consistent tool versions across your team
4. **Document Tool Requirements**: List all required CLI tools in your project documentation
5. **Automate Tool Installation**: Create scripts or Makefile targets to install all required tools

## Quick Reference

| Package Type | Installation Method | Example |
|-------------|-------------------|---------|
| CLI Tool | `go install` | `go install github.com/caarlos0/svu@latest` |
| Library | `go get` (for adding to project) | `go get github.com/stretchr/testify@latest` |
| Library | Import in code | `import "github.com/stretchr/testify/assert"` |

## Troubleshooting

If you encounter the error "is not a main package" when running `go install`:

1. Verify you're installing a CLI tool, not a library
2. Check if the tool has a different installation path (often in a `cmd/` subdirectory)
3. Consult the project's documentation for correct installation instructions
4. Consider if you actually need to install it, or just import it as a dependency