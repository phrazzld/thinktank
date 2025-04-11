```markdown
# TODO

## Remove Viper Dependency
- [x] **Task Title:** Remove Viper dependency via Go modules
  - **Action:** Execute `go get github.com/spf13/viper@none` in the terminal at the project root. Verify the dependency is removed from `go.mod`.
  - **Depends On:** None
  - **AC Ref:** PLAN.md Step 1

- [x] **Task Title:** Tidy Go module dependencies
  - **Action:** Execute `go mod tidy` in the terminal at the project root. Verify `go.mod` and `go.sum` are updated correctly.
  - **Depends On:** Remove Viper dependency via Go modules
  - **AC Ref:** PLAN.md Step 1

## Delete Config File Loading Code
- [ ] **Task Title:** Delete config loader implementation file
  - **Action:** Delete the file `internal/config/loader.go`.
  - **Depends On:** Tidy Go module dependencies
  - **AC Ref:** PLAN.md Step 2

- [ ] **Task Title:** Delete config manager interface file
  - **Action:** Delete the file `internal/config/interfaces.go`.
  - **Depends On:** Tidy Go module dependencies
  - **AC Ref:** PLAN.md Step 2

- [ ] **Task Title:** Delete example config file
  - **Action:** Delete the file `internal/config/example_config.toml`.
  - **Depends On:** Tidy Go module dependencies
  - **AC Ref:** PLAN.md Step 2

## Simplify `internal/config/config.go`
- [ ] **Task Title:** Remove `ManagerInterface` usage reference
  - **Action:** Remove any code referencing `ManagerInterface` within `internal/config/config.go` (if any exists after deleting `interfaces.go`). This task might be implicitly completed by deleting `interfaces.go`. Verify no compilation errors related to the interface remain in this file.
  - **Depends On:** Delete config manager interface file
  - **AC Ref:** PLAN.md Step 3

- [ ] **Task Title:** Remove config directory structures and constants
  - **Action:** Delete the `ConfigDirectories` struct and related constants (`AppName`, `ConfigFilename`) and functions using them from `internal/config/config.go`.
  - **Depends On:** Delete config loader implementation file
  - **AC Ref:** PLAN.md Step 3

- [ ] **Task Title:** Remove struct tags from `AppConfig`
  - **Action:** Remove all `mapstructure` and `toml` struct tags from the `AppConfig` struct definition in `internal/config/config.go`.
  - **Depends On:** Delete config loader implementation file
  - **AC Ref:** PLAN.md Step 3

- [ ] **Task Title:** Refine `AppConfig` struct fields
  - **Action:** Review the fields in `AppConfig`. Remove fields that are *only* set via flags and have no default value or are not used internally across multiple components *after* flag parsing (e.g., potentially `TaskDescription`, `TaskFile`, `Paths`, `DryRun`, `APIKey` as per PLAN.MD). Keep fields that hold default values (`OutputFile`, `ModelName`, `Format`, `LogLevel`, `ConfirmTokens`, `Excludes`) or are needed post-merge.
  - **Depends On:** Remove struct tags from `AppConfig`
  - **AC Ref:** PLAN.md Step 3

- [ ] **Task Title:** Verify `DefaultConfig()` function correctness
  - **Action:** Ensure the `DefaultConfig()` function correctly initializes the simplified `AppConfig` struct with appropriate default values, reflecting any field removals or changes.
  - **Depends On:** Refine `AppConfig` struct fields
  - **AC Ref:** PLAN.md Step 3

## Refactor `internal/architect/types.go`
- [ ] **Task Title:** Review and update `architect.CliConfig` struct
  - **Action:** Review the `CliConfig` struct in `internal/architect/types.go`. Ensure it contains all necessary configuration fields that the core `architect` logic requires, which will now be derived solely from defaults and flags passed down from the `cmd` layer. Add any missing fields previously obtained via config files but now needed from flags (e.g., `ModelName`, `ConfirmTokens`, `LogLevel` if they weren't already there).
  - **Depends On:** Refine `AppConfig` struct fields
  - **AC Ref:** PLAN.md Step 7

## Refactor `cmd/architect/cli.go`
- [ ] **Task Title:** Update `cmd.CliConfig` struct definition
  - **Action:** Modify the `CliConfig` struct in `cmd/architect/cli.go` to include fields for *all* configurable options previously covered by flags *and* config files (e.g., `OutputFile`, `ModelName`, `Include`, `Exclude`, `ExcludeNames`, `Format`, `ConfirmTokens`, `Verbose`, `LogLevel`). Ensure field names and types are appropriate.
  - **Depends On:** Refine `AppConfig` struct fields
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Update `ParseFlagsWithEnv` to define all flags
  - **Action:** Modify the `ParseFlagsWithEnv` function. Ensure `flagSet.String`, `flagSet.Bool`, `flagSet.Int` etc. are called for *all* configurable options defined in the updated `cmd.CliConfig`. Use default values sourced from `internal/config/config.go` (e.g., `config.DefaultOutputFile`, `config.DefaultModel`).
  - **Depends On:** Update `cmd.CliConfig` struct definition, Verify `DefaultConfig()` function correctness
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Update `ParseFlagsWithEnv` to populate `CliConfig` directly
  - **Action:** Modify `ParseFlagsWithEnv` to populate the local `config *CliConfig` struct directly using the values parsed from flags and the `GEMINI_API_KEY` environment variable (via `getenv`). Remove any logic related to preparing data for `ConvertConfigToMap`.
  - **Depends On:** Update `ParseFlagsWithEnv` to define all flags
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Remove `ConvertConfigToMap` call from `ParseFlagsWithEnv`
  - **Action:** Delete the line calling `ConvertConfigToMap` within the `ParseFlagsWithEnv` function, as it's no longer needed.
  - **Depends On:** Update `ParseFlagsWithEnv` to populate `CliConfig` directly
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Update `SetupLoggingCustom` function
  - **Action:** Ensure the `SetupLoggingCustom` function correctly uses the `LogLevel` field from the parsed `CliConfig` struct passed as an argument.
  - **Depends On:** Update `cmd.CliConfig` struct definition
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Remove `ConvertConfigToMap` function definition
  - **Action:** Delete the entire `ConvertConfigToMap` function from `cmd/architect/cli.go`.
  - **Depends On:** Remove `ConvertConfigToMap` call from `ParseFlagsWithEnv`
  - **AC Ref:** PLAN.md Step 4

- [ ] **Task Title:** Update `ValidateInputs` function (if necessary)
  - **Action:** Review the `ValidateInputs` function. Ensure it correctly validates necessary fields directly from the `CliConfig` struct argument. Modify if it was previously relying on intermediate structures or config manager state.
  - **Depends On:** Update `cmd.CliConfig` struct definition
  - **AC Ref:** PLAN.md Step 4

## Refactor `internal/architect/app.go`
- [ ] **Task Title:** Remove `configManager` parameter from `Execute` signature
  - **Action:** Modify the function signature of `Execute` in `internal/architect/app.go` to remove the `configManager config.ManagerInterface` parameter.
  - **Depends On:** Delete config manager interface file
  - **AC Ref:** PLAN.md Step 6

- [ ] **Task Title:** Remove `configManager` parameter from `RunInternal` signature
  - **Action:** Modify the function signature of `RunInternal` in `internal/architect/app.go` to remove the `configManager config.ManagerInterface` parameter.
  - **Depends On:** Delete config manager interface file
  - **AC Ref:** PLAN.md Step 6

- [ ] **Task Title:** Update `Execute` function body to use `cliConfig`
  - **Action:** Refactor the body of the `Execute` function. Replace any usage of `configManager.GetConfig()` or similar methods with direct access to the fields of the `cliConfig *CliConfig` parameter.
  - **Depends On:** Remove `configManager` parameter from `Execute` signature, Review and update `architect.CliConfig` struct
  - **AC Ref:** PLAN.md Step 6

- [ ] **Task Title:** Update `RunInternal` function body to use `cliConfig`
  - **Action:** Refactor the body of the `RunInternal` function. Replace any usage of `configManager.GetConfig()` or similar methods with direct access to the fields of the `cliConfig *CliConfig` parameter.
  - **Depends On:** Remove `configManager` parameter from `RunInternal` signature, Review and update `architect.CliConfig` struct
  - **AC Ref:** PLAN.md Step 6

- [ ] **Task Title:** Update `validateInputs` function in `app.go`
  - **Action:** Ensure the internal `validateInputs` function within `internal/architect/app.go` uses the passed `cliConfig *CliConfig` parameter for its checks.
  - **Depends On:** Review and update `architect.CliConfig` struct
  - **AC Ref:** PLAN.md Step 6

## Refactor `cmd/architect/main.go`
- [ ] **Task Title:** Remove `configManager` initialization and usage in `main.go`
  - **Action:** Delete the lines related to `config.NewManager`, `configManager.LoadFromFiles`, `configManager.EnsureConfigDirs`, `configManager.MergeWithFlags`, and `configManager.GetConfig` from `cmd/architect/main.go`.
  - **Depends On:** Delete config loader implementation file, Delete config manager interface file
  - **AC Ref:** PLAN.md Step 5

- [ ] **Task Title:** Update `architect.Execute` call in `main.go`
  - **Action:** Modify the call to `architect.Execute` in `main.go`. Pass the `coreConfig` (derived from `cmdConfig`) and `logger`. Remove the `configManager` argument. Ensure the arguments match the updated `Execute` signature.
  - **Depends On:** Remove `configManager` initialization and usage in `main.go`, Remove `configManager` parameter from `Execute` signature
  - **AC Ref:** PLAN.md Step 5

- [ ] **Task Title:** Update `convertToArchitectConfig` function
  - **Action:** Review and update the `convertToArchitectConfig` function in `main.go`. Ensure it correctly maps all necessary fields from the `cmd.CliConfig` struct (populated by flags/env) to the `architect.CliConfig` struct (used by core logic).
  - **Depends On:** Update `cmd.CliConfig` struct definition, Review and update `architect.CliConfig` struct
  - **AC Ref:** PLAN.md Step 5

## Refactor Tests
- [ ] **Task Title:** Delete config file loading tests
  - **Action:** Delete any test files specifically targeting the config file loading mechanism (e.g., `internal/config/loader_test.go` if it exists). Remove related test helper functions or fixtures.
  - **Depends On:** Delete config loader implementation file
  - **AC Ref:** PLAN.md Step 9

- [ ] **Task Title:** Update `cmd/architect/cli.go` tests
  - **Action:** Modify existing tests or add new tests for `cmd/architect/cli.go`. Verify that `ParseFlagsWithEnv` correctly parses all flags, applies defaults, reads the environment variable, and populates the `CliConfig` struct accurately. Test edge cases and error handling for flag parsing. Ensure tests for `SetupLoggingCustom` and `ValidateInputs` are still valid or updated.
  - **Depends On:** Update `ParseFlagsWithEnv` to populate `CliConfig` directly, Remove `ConvertConfigToMap` function definition, Update `ValidateInputs` function (if necessary)
  - **AC Ref:** PLAN.md Step 9

- [ ] **Task Title:** Update integration tests for `main.go` and `app.go`
  - **Action:** Refactor integration tests that involve `cmd/architect/main.go` or `internal/architect/app.go`. Remove any setup related to mocking `ConfigManager` or creating temporary config files. Update test setup to pass configuration directly via the `architect.CliConfig` struct when calling `architect.Execute` or `architect.RunInternal`.
  - **Depends On:** Update `architect.Execute` call in `main.go`, Update `Execute` function body to use `cliConfig`, Update `RunInternal` function body to use `cliConfig`
  - **AC Ref:** PLAN.md Step 9

- [ ] **Task Title:** Ensure all tests pass
  - **Action:** Run the full test suite (`go test ./...`). Verify that all tests pass after completing the refactoring steps. Debug and fix any failing tests.
  - **Depends On:** Update integration tests for `main.go` and `app.go`
  - **AC Ref:** PLAN.md Step 9, Definition of Done

## Update Documentation
- [ ] **Task Title:** Update `README.md` documentation
  - **Action:** Modify `README.md`. Remove any sections discussing configuration files (`config.toml`, XDG paths). Clearly state that configuration is managed *only* via CLI flags and the `GEMINI_API_KEY` environment variable. Update usage examples to reflect this. Ensure the README accurately describes all available flags and the required environment variable.
  - **Depends On:** Ensure all tests pass (implies functionality is stable)
  - **AC Ref:** PLAN.md Step 8

- [ ] **Task Title:** Verify CLI `--help` output
  - **Action:** Run the application with the `--help` flag. Verify that the output accurately reflects all available flags, their descriptions, default values, and mentions the `GEMINI_API_KEY` environment variable. Ensure the usage examples are correct. Update flag descriptions in `cmd/architect/cli.go` if necessary.
  - **Depends On:** Update `ParseFlagsWithEnv` to define all flags
  - **AC Ref:** PLAN.md Step 8

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** Assumed `AppConfig` in `internal/config` is retained but simplified.
  - **Context:** PLAN.md Step 3 includes a decision: "*Decision: Keep `AppConfig` for now to hold merged values, but simplify it.*" This decomposition assumes this decision holds and `AppConfig` is not entirely removed or merged into `cmd.CliConfig`.

- [ ] **Issue/Assumption:** Assumed specific fields should be removed from `internal/config/AppConfig`.
  - **Context:** PLAN.md Step 3 suggests removing fields like `TaskDescription`, `TaskFile`, `Paths`, `DryRun`, `APIKey` from `AppConfig` if they are only set via flags/env and not defaulted. Task "Refine `AppConfig` struct fields" implements this assumption. Confirmation during implementation is advised.

- [ ] **Issue/Assumption:** Assumed `architect.CliConfig` in `internal/architect/types.go` is the designated struct for passing final configuration to the core logic.
  - **Context:** PLAN.md Steps 5, 6, and 7 imply that a config struct (likely `architect.CliConfig`) will be populated from `cmd.CliConfig` and passed to the core `architect.Execute` function. This decomposition follows that structure.
```