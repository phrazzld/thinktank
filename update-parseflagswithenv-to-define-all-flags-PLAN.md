# Update `ParseFlagsWithEnv` to Define All Flags

## Task Title
Update `ParseFlagsWithEnv` to define all flags

## Implementation Approach
Verify that the `ParseFlagsWithEnv` function in `cmd/architect/cli.go` correctly defines all flags for all configurable options in the `CliConfig` struct, using default values from `internal/config/config.go` where appropriate. Ensure every field in `CliConfig` has a corresponding flag definition.