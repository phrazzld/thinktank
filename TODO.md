# Remove Color Flag

This task involves removing the color flag functionality from the Architect tool. This aligns with the backlog item: "remove user-facing niceties like spinners and excessive color formatting in favor of clean, programmable output".

## Context
- Currently, the tool has a `--color` flag that enables/disables colored log output
- The default is `true` (colors enabled)
- We need to remove this option and make logs consistently uncolored

## Tasks

### 1. Command Line Interface
- [x] Remove `UseColors` field from `CliConfig` struct in `cmd/architect/cli.go`
- [x] Remove color flag definition (`useColorsFlag`) in `cmd/architect/cli.go`
- [x] Remove assignment of flag value to config in `cmd/architect/cli.go`
- [x] Update logger initialization to not use color parameter in `cmd/architect/cli.go`
- [x] Update CLI tests in `cmd/architect/cli_test.go` to remove color flag testing

### 2. Configuration
- [x] Remove `UseColors` field from `AppConfig` struct in `internal/config/config.go`
- [x] Remove default value assignment in `DefaultConfig()` function
- [x] Remove color setting in `internal/config/loader.go`
- [x] Update loader tests in `internal/config/loader_test.go` to remove color flag testing
- [x] Remove color setting in `internal/config/example_config.toml`

### 3. Logger Implementation
- [x] Remove color-related fields in `Logger` struct in `internal/logutil/logutil.go`
- [x] Remove color initialization in `NewLogger` function
- [x] Simplify output handling to always use non-colored output
- [x] Remove `SetUseColors` method from `Logger`
- [x] Update the API of `NewLogger` to remove the color parameter

### 4. Tests Adjustments
- [x] Update any tests using the logger to not specify color parameter (including in `internal/architect/output_test.go`)
- [x] Ensure all tests pass after removing color functionality

### 5. Documentation
- [x] Remove color flag entry from README.md
- [x] Update any other documentation references to colored output or the color flag

## Verification
- [x] Run tests to ensure all functionality works correctly: `go test ./...`
- [x] Verify linting passes: `go fmt ./...` and `go vet ./...`
- [x] Manually test the tool to ensure logs display correctly without color