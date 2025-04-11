# DONE

## Phase 1: Core Type & CLI Flag Refactoring

- [x] **Task: Update Core CliConfig Types** - 2025-04-10
    - **Action:** Modify the `CliConfig` struct in both `internal/architect/types.go` and `cmd/architect/cli.go`. Remove fields: `TaskDescription`, `Template`, `ListExamples`, `ShowExample`. Add field: `InstructionsFile string`.
    - **Depends On:** None
    - **AC Ref:** Plan Section 2 (Type Definitions), Plan Section 4 (CLIConfig Structs)

- [x] **Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)** - 2025-04-10
    - **Action:** Update `ParseFlagsWithEnv` function: Remove definitions for `--task-file`, `--prompt-template`, `--list-examples`, `--show-example`. Add definition for a required `--instructions` string flag. Update flag parsing logic to populate `InstructionsFile` in `CliConfig`. Remove population of deleted fields.
    - **Depends On:** Task: Update Core CliConfig Types
    - **AC Ref:** Plan Section 3 (Task 1), Plan Section 4 (CLIConfig Structs)

- [x] **Task: Update CLI Usage Message (`cmd/architect/cli.go`)** - 2025-04-10
    - **Action:** Modify the `flagSet.Usage` function to accurately reflect the new command structure (`architect --instructions <file> [context_paths...]`) and remove documentation for deleted flags.
    - **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    - **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update CLI Input Validation (`cmd/architect/cli.go`)** - 2025-04-10
    * **Action:** Modify `ValidateInputs` function. Remove the check for `TaskFile`. Add a check ensuring `InstructionsFile` is provided (unless in `--dry-run` mode, if that distinction is kept). Remove validation bypass logic related to example commands.
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update Config-Flag Mapping (`cmd/architect/cli.go`)** - 2025-04-10
    * **Action:** Modify `ConvertConfigToMap`. Remove map keys related to removed flags (`promptTemplate`, `listExamples`, `showExample`, `taskFile`). Add `instructionsFile` key if necessary for config merging (likely not needed).
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update Main Entry Point Logic (`cmd/architect/main.go`)** - 2025-04-10
    * **Action:** Modify `convertToArchitectConfig` function to correctly map the fields between `cmd/architect/cli.go::CliConfig` and `internal/architect/types.go::CliConfig`, reflecting the removal/addition of fields. Remove call to `architect.HandleSpecialCommands`.
    * **Depends On:** Task: Update Core CliConfig Types
    * **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update CLI Tests (`cmd/architect/cli_test.go`)** - 2025-04-10
    * **Action:** Rewrite tests to verify the new flag parsing (`--instructions`), validation logic, usage message, and removal of old flags/features. Ensure tests cover required flag errors and positional argument handling for context.
    * **Depends On:** Task: Update CLI Input Validation (`cmd/architect/cli.go`), Task: Update CLI Usage Message (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

## Phase 2: Remove Templating System

- [x] **Task: Delete Prompt Package (`internal/prompt/`)** - 2025-04-10
    * **Action:** Delete the entire `internal/prompt/` directory, including all `.go` files, `.tmpl` files, and tests within it.
    * **Depends On:** None (Can be done early, but code won't compile until usages are removed)
    * **AC Ref:** Plan Section 2 (Templating Removal), Plan Section 3 (Task 4)

- [x] **Task: Remove Template Fields from Config (`internal/config/`)** - 2025-04-10
    * **Action:** Edit `internal/config/config.go`: Remove the `Templates TemplateConfig` field from the `AppConfig` struct. Update the `DefaultConfig` function accordingly.
    * **Depends On:** None (Can be done early, but code won't compile until usages are removed)
    * **AC Ref:** Plan Section 2 (Configuration Changes), Plan Section 3 (Task 7)

- [x] **Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)** - 2025-04-10
    * **Action:** Edit `internal/config/loader.go`: Remove the `GetTemplatePath` function. Remove any logic related to `templates.dir` or loading specific template names (e.g., `default`, `test`, `custom`) from the configuration within `LoadFromFiles` or `MergeWithFlags`. Remove `setViperDefaults` entries for templates.
    * **Depends On:** Task: Remove Template Fields from Config (`internal/config/`)
    * **AC Ref:** Plan Section 2 (Configuration Changes), Plan Section 3 (Task 7)

- [x] **Task: Update Config Tests (`internal/config/loader_test.go`, `legacy_config_test.go`)** - 2025-04-10
    * **Action:** Rewrite config loader tests to remove checks for template loading, `GetTemplatePath`, and template-related configuration fields. Ensure legacy config tests still handle potentially ignored fields correctly if applicable, but without checking for template fields specifically.
    * **Depends On:** Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)
    * **AC Ref:** Plan Section 3 (Task 7), Plan Section 6 (Unit Tests)

## Phase 3: Refactor Core Logic & Context Handling

- [x] **Task: Define FileMeta Struct (`internal/fileutil/fileutil.go`)** - 2025-04-10
    * **Action:** Define `type FileMeta struct { Path string; Content string }` within the `internal/fileutil` package (or reuse if a similar suitable type already exists).
    * **Depends On:** None
    * **AC Ref:** Plan Section 4 (FileUtil)

- [x] **Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)** - 2025-04-10
    * **Action:** Change the signature of `GatherProjectContext` to `func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error)`. Modify the implementation to collect processed files into a `[]FileMeta` slice instead of building a formatted string. Remove internal usage of `config.Format`.
    * **Depends On:** Task: Define FileMeta Struct (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 2 (Structured Context), Plan Section 3 (Task 3), Plan Section 4 (FileUtil)

- [x] **Task: Update `fileutil` Tests (`internal/fileutil/context_test.go`)** - 2025-04-10
    * **Action:** Rewrite tests for `GatherProjectContext` to assert that the returned slice (`[]FileMeta`) contains the expected files (correct paths and content) in the expected order (if applicable, e.g., based on argument order or FS walk order) and that the file count is correct. Verify filtering logic still works.
    * **Depends On:** Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

- [x] **Task: Implement Prompt Stitching Logic (`internal/architect/app.go` or `output.go`)** - 2025-04-10
    * **Action:** Add code (likely within the `Execute` or `RunInternal` function, or potentially refactored into `output.go`) that:
        1. Takes the instructions string and the `[]FileMeta` context slice.
        2. Formats each `FileMeta` entry using the `<{path}>...</{path}>` style (including potential escaping via an `escapeContent` helper).
        3. Concatenates the formatted file contexts into a single string dump.
        4. Combines the instructions and the context dump using the `<instructions>...</instructions>` and `<context>...</context>` XML tags, including appropriate newlines.
    * **Depends On:** Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 2 (Prompt Stitching), Plan Section 4 (Prompt Stitching Logic)

- [x] **Task: Add Prompt Stitching Unit Tests** - 2025-04-10
    * **Action:** Create unit tests specifically for the prompt stitching logic implemented in the previous task. Verify correct formatting of individual context files, correct usage of XML tags, and proper combination of instructions and context parts. Test edge cases like empty context or empty instructions.
    * **Depends On:** Task: Implement Prompt Stitching Logic (`internal/architect/app.go` or `output.go`)
    * **AC Ref:** Plan Section 6 (Unit Tests)

- [x] **Task: Refactor Core Application Flow (`internal/architect/app.go`)** - 2025-04-10
    * **Action:** Modify `Execute` and `RunInternal`:
        * Remove call to `processTaskInput`. Read instructions directly from `cliConfig.InstructionsFile` using `os.ReadFile`.
        * Update call to `contextGatherer.GatherContext` to receive `[]FileMeta`.
        * Call the new Prompt Stitching logic to get the final prompt string.
        * Remove logic related to `prompt.ManagerInterface`, `TemplateData`.
        * Update the call to `geminiClient.GenerateContent` to use the final stitched prompt string.
        * Update the call to `outputWriter.SaveToFile` (or equivalent) to pass the raw API response.
        * Remove `HandleSpecialCommands`.
    * **Depends On:** Task: Implement Prompt Stitching Logic (`internal/architect/app.go` or `output.go`), Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`), Task: Remove Prompt Package (`internal/prompt/`)
    * **AC Ref:** Plan Section 3 (Task 5)

- [x] **Task: Refactor Output Writing (`internal/architect/output.go`)** - 2025-04-10
    * **Action:** Removed the `GenerateAndSavePlan` and `GenerateAndSavePlanWithConfig` methods. Renamed the interface from `OutputWriter` to `FileWriter` to better reflect its focused responsibility. Simplified the implementation by removing token management and API service dependencies, focusing solely on file writing. Updated all references in app.go to use the new interface.
    * **Depends On:** Task: Remove Prompt Package (`internal/prompt/`)
    * **AC Ref:** Plan Section 3 (Task 6)

- [x] **Task: Update Output Tests (`internal/architect/output_test.go`)** - 2025-04-10
    * **Action:** Verified that tests for `GenerateAndSavePlan*` methods were already removed during previous refactoring. Confirmed all tests for the remaining functionality (`SaveToFile`, `EscapeContent`, and `StitchPrompt`) were working correctly. Removed unused files (prompt.go and prompt_test.go) from the cmd/architect package as they depended on the deleted internal/prompt package.
    * **Depends On:** Task: Refactor Output Writing (`internal/architect/output.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

- [x] **Task: Update Integration Tests (`internal/integration/`)** - 2025-04-10
    * **Action:** Rewrote integration tests to use the new `--instructions` flag instead of the removed `--task-file` flag. Updated references in test cases from `TaskFile` to `InstructionsFile`. Created helper methods to validate the XML structure of prompts, including proper XML escaping in file content. Added tests for complex XML content handling, empty instructions, and proper structure validation. Created mock methods to replace template-related functions that were removed from the config package.
    * **Depends On:** Task: Refactor Core Application Flow (`internal/architect/app.go`), Task: Refactor Output Writing (`internal/architect/output.go`), Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Integration Tests)