# Update `cmd.CliConfig` Struct Definition

## Task Title
Update `cmd.CliConfig` struct definition

## Implementation Approach
Review the current `CliConfig` struct in `cmd/architect/cli.go` and compare it with both the `AppConfig` struct in `internal/config/config.go` and the `architect.CliConfig` struct in `internal/architect/types.go`. Ensure it contains all necessary fields for configuration options that were previously managed via both flags and config files.