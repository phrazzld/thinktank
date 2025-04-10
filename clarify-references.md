# Clarify Feature References

This document catalogs all references to the "clarify" feature found in the codebase, organized by component.

## CLI Components

### `cmd/architect/cli.go`
- Line ~39: `ClarifyTask bool` in the `CliConfig` struct
- Line ~98: `clarifyTaskFlag := flagSet.Bool("clarify", false, "Enable interactive task clarification to refine your task description")`
- Line ~139: `config.ClarifyTask = *clarifyTaskFlag` in flag parsing logic
- Line ~209: `"clarify": cliConfig.ClarifyTask` in the `ConvertConfigToMap` function

## Tests

### `cmd/architect/flags_test.go`
- Lines 7-23: `TestConvertConfigNoClarity` test which verifies that `clarify_task` is not included in the config map
- Line 19: `_, hasClarifyTask := configMap["clarify_task"]` 
- Line 20-22: Assertion that `hasClarifyTask` should be false

### `internal/config/legacy_config_test.go`
- Contains tests for legacy config handling with clarify-related fields
- Line 31: `clarify_task = true` in test config file content
- Line 35: `clarify = "clarify.tmpl"` in test template configuration
- Lines 76-87: Test to verify clarify template isn't included
- Lines 90-92: Test to verify clarify template path isn't retrievable 

### `internal/integration/integration_test.go`
- Line 443: Comment about replaced `TestTaskClarification` test

## Configuration

### `internal/config/example_config.toml`
- Line 8: `clarify_task = false  # Whether to enable task clarification`
- Line 23: `clarify = "clarify.tmpl"  # Template for task clarification`

## Documentation

### `README.md`
- Line ~33: "Task Clarification" feature in the Features list
- Line ~87-88: `architect --task-file task.txt --clarify ./` example
- Line ~115: `--clarify` row in the Configuration Options table

### `BACKLOG.md`
- Line 5: `purge the program of the clarify flag and feature and related code and tests` as a backlog item

## Templates
- No actual template files named `clarify.tmpl` were found, but references exist in configuration

## Core Logic
- No explicit references to `cliConfig.ClarifyTask` or clarify-related function calls were found in the `internal/architect` package Go files.

## Summary

The clarify feature appears to be primarily defined at the CLI level with the `--clarify` flag and `ClarifyTask` field in the `CliConfig` struct. The feature is referenced in configuration files and tests, but no actual implementation code (conditional blocks, function calls) was found in the core logic. 

This suggests that while the flag and configuration infrastructure exists, the actual feature implementation may have been removed already, leaving behind these references. This is consistent with the backlog item to "purge the program of the clarify flag and feature and related code and tests".

The main components to address are:
1. CLI flag and struct field removal in `cmd/architect/cli.go`
2. Test cleanup in multiple files
3. Configuration cleanup in `internal/config/example_config.toml`
4. Documentation updates in README.md
5. Backlog update once complete