# Project Rename: architect → thinktank

This document outlines tasks to complete the project rename from "architect" to "thinktank".

## Repository Management

- [ ] Rename the repository on GitHub (phrazzld/architect → phrazzld/thinktank)
- [ ] Update git remote URLs on all development machines
- [ ] Update any CI/CD configurations that reference the old repository name

## Code Structure

- [x] Rename directory structure:
  - [x] `cmd/architect/` → `cmd/thinktank/`
  - [x] `internal/architect/` → `internal/thinktank/`
  - [x] `internal/architect/thinktank-output/` (check if this needs adjustment)

## Module and Import Paths

- [x] Update Go module name in go.mod (github.com/phrazzld/architect → github.com/phrazzld/thinktank)
- [x] Fix all import paths referencing the old module name
- [x] Update any hardcoded import paths in test files

## Command Line Interface

- [x] Update binary name reference in build scripts
- [x] Update CLI help text and command name references
- [x] Update CLI documentation in README.md and other documentation
- [x] Ensure the built binary is named "thinktank" instead of "architect"

## Documentation

- [x] Update README.md to use "thinktank" instead of "architect"
- [x] Update CLAUDE.md reference to using the CLI (section "Using the `architect` CLI")
- [x] Update any examples in documentation that use "architect" command
- [x] Update any API documentation that references "architect"
- [x] Check for references in all markdown files
## Code References

- [x] Update comments in code that refer to "architect"
  Action: This task is superseded by T107-T115. It will be marked complete by T115.
  Depends On: []
  AC Ref: None

- [x] T107: Verify current branch state
  Action: Confirm we are on the rebrand branch with `git branch` and ensure we have a clean working state with `git status`. All subsequent tasks in this sequence (T108-T115) will be performed on this existing branch.
  Depends On: []
  AC Ref: None

- [x] T108: Verify clean build state
  Action: Run the full build process (`go build ./...`) and ensure it completes successfully without errors or warnings on the current branch *before* making any changes.
  Depends On: [T107]
  AC Ref: None

- [x] T109: Identify all 'architect' references in comments (excluding vendored code)
  Action: Use search tools (e.g., `grep -ri --exclude-dir=vendor 'architect'`, IDE search) to find all occurrences of the word 'architect' within code comments across the entire repository, specifically *excluding* any `vendor/` or third-party library directories. Document the list of files containing relevant comments.
  Depends On: [T108]
  AC Ref: None

- [x] T110: Update identified comments from 'architect' to 'thinktank'
  Action: Carefully review each comment identified in T109. Update occurrences of 'architect' to 'thinktank' where appropriate, ensuring the context still makes sense. Avoid modifying comments within vendored/third-party code directories. Commit these changes.
  Depends On: [T109]
  AC Ref: None

- [x] T111: Verify build integrity after comment updates
  Action: Run the full build process again (`go build ./...`). Ensure it still completes successfully without any new errors or warnings introduced by the comment changes.
  Depends On: [T110]
  AC Ref: None

- [x] T112: Run automated tests
  Action: Execute the project's full automated test suite (`go test ./...`). Ensure all tests pass.
  Depends On: [T111]
  AC Ref: None

- [x] T113: Run linters and static analysis
  Action: Run any configured linters or static analysis tools (`go vet ./...`). Ensure they pass without reporting new issues related to the comment changes.
  Depends On: [T112]
  AC Ref: None

- [x] T114: Final commit review
  Action: Perform a final manual review of the changed files, confirming only comments were modified as intended and build/test/lint steps (T111, T112, T113) were successful.
  Depends On: [T113]
  AC Ref: None

- [x] T115: Mark original task as complete
  Action: Once T107 through T114 are completed, mark this task and the original task as complete in `TODO.md` and push changes.
  Depends On: [T107, T108, T109, T110, T111, T112, T113, T114]
  AC Ref: None

- [ ] Update any error messages that include "architect"
- [ ] Update any log messages that include "architect"
- [ ] Update any test fixtures or test data with "architect" references

## Configuration

- [ ] Update any config files that might contain "architect" references
- [x] Update scripts/setup.sh if it contains "architect" references
- [x] Check config/ directory for any configuration referencing "architect"

## Testing

- [x] Update end-to-end tests in `internal/e2e/` to use "thinktank" instead of "architect"
- [x] Make sure all tests pass with the new naming

## Package Names

- [x] Rename all Go package declarations from "architect" to "thinktank"
- [x] Update any package documentation containing "architect" references

## Verification

- [ ] Grep for any remaining instances of "architect" or "Architect"
- [ ] Run all tests to ensure everything works with the new name
