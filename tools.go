//go:build tools
// +build tools

package tools

import (
	// golangci-lint for running linters
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"

	// go-formatted hooks for pre-commit
	// Note: dnephin/pre-commit-golang doesn't have a Go package itself, but it's referenced in .pre-commit-config.yaml
	// The pre-commit hooks operate as shell scripts, not Go packages

	// govulncheck for vulnerability scanning (mentioned in documentation)
	_ "golang.org/x/vuln/cmd/govulncheck"

	// svu for semantic versioning (mentioned in TODO.md)
	_ "github.com/caarlos0/svu"

	// git-chglog for changelog generation (mentioned in TODO.md)
	_ "github.com/git-chglog/git-chglog/cmd/git-chglog"

	// commitlint for validating conventional commit messages
	_ "github.com/conventionalcommit/commitlint"

	// goreleaser for automated releases
	_ "github.com/goreleaser/goreleaser"
)

// This file ensures tool dependencies are tracked in go.mod.
// The go.mod file will track these dependencies even though they're not being used
// in the actual code of the project.
//
// To install a tool, use: go install <import-path>@<version>
//
// For example:
// go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
// go install golang.org/x/vuln/cmd/govulncheck@latest
// go install github.com/caarlos0/svu@latest
// go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
//
// Note: Some tools mentioned in the codebase are either:
// 1. Not Go packages (pre-commit-golang)
// 2. Private repositories (phaedrus-dev/glance)
// 3. Repositories that couldn't be found (commitdev/go-conventionalcommits)
// These should be installed through alternative means like pip, brew, or directly from their repositories.
