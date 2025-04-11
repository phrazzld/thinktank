# Update `ParseFlagsWithEnv` to Populate `CliConfig` Directly

## Task Title
Update `ParseFlagsWithEnv` to populate `CliConfig` directly

## Implementation Approach
Analyze the current `ParseFlagsWithEnv` function in `cmd/architect/cli.go` to verify it already populates the `CliConfig` struct directly using values parsed from flags and the `GEMINI_API_KEY` environment variable. Confirm there's no logic in this function related to `ConvertConfigToMap`.