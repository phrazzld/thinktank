//go:build tools
// +build tools

package tools

import (
	// golangci-lint for running linters
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"

	// go-formatted hooks for pre-commit
	// Note: dnephin/pre-commit-golang doesn't have a Go package itself, but it's referenced in .pre-commit-config.yaml
	// The pre-commit hooks operate as shell scripts, not Go packages

	// govulncheck for vulnerability scanning
	_ "golang.org/x/vuln/cmd/govulncheck"

	// svu for semantic versioning
	_ "github.com/caarlos0/svu/v3"

	// git-chglog for changelog generation
	_ "github.com/git-chglog/git-chglog/cmd/git-chglog"

	// commitlint for validating conventional commit messages
	_ "github.com/conventionalcommit/commitlint"

	// go-conventionalcommits for validating commit messages in pre-commit hooks
	_ "github.com/leodido/go-conventionalcommits"

	// goreleaser for automated releases
	_ "github.com/goreleaser/goreleaser"
)

// This file ensures tool dependencies are tracked in go.mod.
// The go.mod file will track these dependencies even though they're not being used
// in the actual code of the project.
//
// To install a tool with a specific version, use: go install <import-path>@<version>
//
// For example:
// go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.5
// go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
// go install github.com/caarlos0/svu/v3@v3.2.3
// go install github.com/git-chglog/git-chglog/cmd/git-chglog@v0.15.4
// go install github.com/leodido/go-conventionalcommits@v0.12.0
// go install github.com/goreleaser/goreleaser@v1.26.2
//
// Note: Some tools mentioned in the codebase are either:
// 1. Not Go packages (pre-commit-golang)
// 2. Private repositories (phaedrus-dev/glance)
// These should be installed through alternative means like pip, brew, or directly from their repositories.
