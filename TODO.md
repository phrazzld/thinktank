# Project Rename: architect → thinktank

This document outlines tasks to complete the project rename from "architect" to "thinktank".

## Repository Management

- [ ] Rename the repository on GitHub (phrazzld/architect → phrazzld/thinktank)
- [ ] Update git remote URLs on all development machines
- [ ] Update any CI/CD configurations that reference the old repository name

## Code Structure

- [ ] Rename directory structure:
  - [ ] `cmd/architect/` → `cmd/thinktank/`
  - [ ] `internal/architect/` → `internal/thinktank/`
  - [ ] `internal/architect/thinktank-output/` (check if this needs adjustment)

## Module and Import Paths

- [ ] Update Go module name in go.mod (github.com/phrazzld/architect → github.com/phrazzld/thinktank)
- [ ] Fix all import paths referencing the old module name
- [ ] Update any hardcoded import paths in test files

## Command Line Interface

- [ ] Update binary name reference in build scripts
- [ ] Update CLI help text and command name references
- [ ] Update CLI documentation in README.md and other documentation
- [ ] Ensure the built binary is named "thinktank" instead of "architect"

## Documentation

- [x] Update README.md to use "thinktank" instead of "architect"
- [x] Update CLAUDE.md reference to using the CLI (section "Using the `architect` CLI")
- [x] Update any examples in documentation that use "architect" command
- [x] Update any API documentation that references "architect"
- [ ] Check for references in all markdown files
## Code References

- [ ] Update comments in code that refer to "architect"
- [ ] Update any error messages that include "architect"
- [ ] Update any log messages that include "architect"
- [ ] Update any test fixtures or test data with "architect" references

## Configuration

- [ ] Update any config files that might contain "architect" references
- [ ] Update scripts/setup.sh if it contains "architect" references
- [ ] Check config/ directory for any configuration referencing "architect"

## Testing

- [ ] Update end-to-end tests in `internal/e2e/` to use "thinktank" instead of "architect"
- [ ] Make sure all tests pass with the new naming

## Package Names

- [ ] Rename all Go package declarations from "architect" to "thinktank"
- [ ] Update any package documentation containing "architect" references

## Verification

- [ ] Grep for any remaining instances of "architect" or "Architect"
- [ ] Run all tests to ensure everything works with the new name
