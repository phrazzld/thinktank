# TODO

## Phase 1: Core Type & CLI Flag Refactoring

- [x] **Task: Update Core CliConfig Types**
    - **Action:** Modify the `CliConfig` struct in both `internal/architect/types.go` and `cmd/architect/cli.go`. Remove fields: `TaskDescription`, `Template`, `ListExamples`, `ShowExample`. Add field: `InstructionsFile string`.
    - **Depends On:** None
    - **AC Ref:** Plan Section 2 (Type Definitions), Plan Section 4 (CLIConfig Structs)

- [x] **Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)**
    - **Action:** Update `ParseFlagsWithEnv` function: Remove definitions for `--task-file`, `--prompt-template`, `--list-examples`, `--show-example`. Add definition for a required `--instructions` string flag. Update flag parsing logic to populate `InstructionsFile` in `CliConfig`. Remove population of deleted fields.
    - **Depends On:** Task: Update Core CliConfig Types
    - **AC Ref:** Plan Section 3 (Task 1), Plan Section 4 (CLIConfig Structs)

- [x] **Task: Update CLI Usage Message (`cmd/architect/cli.go`)**
    - **Action:** Modify the `flagSet.Usage` function to accurately reflect the new command structure (`architect --instructions <file> [context_paths...]`) and remove documentation for deleted flags.
    - **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    - **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update CLI Input Validation (`cmd/architect/cli.go`)**
    * **Action:** Modify `ValidateInputs` function. Remove the check for `TaskFile`. Add a check ensuring `InstructionsFile` is provided (unless in `--dry-run` mode, if that distinction is kept). Remove validation bypass logic related to example commands.
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update Config-Flag Mapping (`cmd/architect/cli.go`)**
    * **Action:** Modify `ConvertConfigToMap`. Remove map keys related to removed flags (`promptTemplate`, `listExamples`, `showExample`, `taskFile`). Add `instructionsFile` key if necessary for config merging (likely not needed).
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 1)

- [x] **Task: Update Main Entry Point Logic (`cmd/architect/main.go`)**
    * **Action:** Modify `convertToArchitectConfig` function to correctly map the fields between `cmd/architect/cli.go::CliConfig` and `internal/architect/types.go::CliConfig`, reflecting the removal/addition of fields. Remove call to `architect.HandleSpecialCommands`.
    * **Depends On:** Task: Update Core CliConfig Types
    * **AC Ref:** Plan Section 3 (Task 1)

- [ ] **Task: Update CLI Tests (`cmd/architect/cli_test.go`)**
    * **Action:** Rewrite tests to verify the new flag parsing (`--instructions`), validation logic, usage message, and removal of old flags/features. Ensure tests cover required flag errors and positional argument handling for context.
    * **Depends On:** Task: Update CLI Input Validation (`cmd/architect/cli.go`), Task: Update CLI Usage Message (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

## Phase 2: Remove Templating System

- [ ] **Task: Delete Prompt Package (`internal/prompt/`)**
    * **Action:** Delete the entire `internal/prompt/` directory, including all `.go` files, `.tmpl` files, and tests within it.
    * **Depends On:** None (Can be done early, but code won't compile until usages are removed)
    * **AC Ref:** Plan Section 2 (Templating Removal), Plan Section 3 (Task 4)

- [ ] **Task: Remove Template Fields from Config (`internal/config/`)**
    * **Action:** Edit `internal/config/config.go`: Remove the `Templates TemplateConfig` field from the `AppConfig` struct. Update the `DefaultConfig` function accordingly.
    * **Depends On:** None (Can be done early, but code won't compile until usages are removed)
    * **AC Ref:** Plan Section 2 (Configuration Changes), Plan Section 3 (Task 7)

- [ ] **Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)**
    * **Action:** Edit `internal/config/loader.go`: Remove the `GetTemplatePath` function. Remove any logic related to `templates.dir` or loading specific template names (e.g., `default`, `test`, `custom`) from the configuration within `LoadFromFiles` or `MergeWithFlags`. Remove `setViperDefaults` entries for templates.
    * **Depends On:** Task: Remove Template Fields from Config (`internal/config/`)
    * **AC Ref:** Plan Section 2 (Configuration Changes), Plan Section 3 (Task 7)

- [ ] **Task: Update Config Tests (`internal/config/loader_test.go`, `legacy_config_test.go`)**
    * **Action:** Rewrite config loader tests to remove checks for template loading, `GetTemplatePath`, and template-related configuration fields. Ensure legacy config tests still handle potentially ignored fields correctly if applicable, but without checking for template fields specifically.
    * **Depends On:** Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)
    * **AC Ref:** Plan Section 3 (Task 7), Plan Section 6 (Unit Tests)

## Phase 3: Refactor Core Logic & Context Handling

- [ ] **Task: Define FileMeta Struct (`internal/fileutil/fileutil.go`)**
    * **Action:** Define `type FileMeta struct { Path string; Content string }` within the `internal/fileutil` package (or reuse if a similar suitable type already exists).
    * **Depends On:** None
    * **AC Ref:** Plan Section 4 (FileUtil)

- [ ] **Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)**
    * **Action:** Change the signature of `GatherProjectContext` to `func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error)`. Modify the implementation to collect processed files into a `[]FileMeta` slice instead of building a formatted string. Remove internal usage of `config.Format`.
    * **Depends On:** Task: Define FileMeta Struct (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 2 (Structured Context), Plan Section 3 (Task 3), Plan Section 4 (FileUtil)

- [ ] **Task: Update `fileutil` Tests (`internal/fileutil/context_test.go`)**
    * **Action:** Rewrite tests for `GatherProjectContext` to assert that the returned slice (`[]FileMeta`) contains the expected files (correct paths and content) in the expected order (if applicable, e.g., based on argument order or FS walk order) and that the file count is correct. Verify filtering logic still works.
    * **Depends On:** Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

- [ ] **Task: Implement Prompt Stitching Logic (`internal/architect/app.go` or `output.go`)**
    * **Action:** Add code (likely within the `Execute` or `RunInternal` function, or potentially refactored into `output.go`) that:
        1. Takes the instructions string and the `[]FileMeta` context slice.
        2. Formats each `FileMeta` entry using the `<{path}>...</{path}>` style (including potential escaping via an `escapeContent` helper).
        3. Concatenates the formatted file contexts into a single string dump.
        4. Combines the instructions and the context dump using the `<instructions>...</instructions>` and `<context>...</context>` XML tags, including appropriate newlines.
    * **Depends On:** Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)
    * **AC Ref:** Plan Section 2 (Prompt Stitching), Plan Section 4 (Prompt Stitching Logic)

- [ ] **Task: Add Prompt Stitching Unit Tests**
    * **Action:** Create unit tests specifically for the prompt stitching logic implemented in the previous task. Verify correct formatting of individual context files, correct usage of XML tags, and proper combination of instructions and context parts. Test edge cases like empty context or empty instructions.
    * **Depends On:** Task: Implement Prompt Stitching Logic (`internal/architect/app.go` or `output.go`)
    * **AC Ref:** Plan Section 6 (Unit Tests)

- [ ] **Task: Refactor Core Application Flow (`internal/architect/app.go`)**
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

- [ ] **Task: Refactor Output Writing (`internal/architect/output.go`)**
    * **Action:** Remove the `GenerateAndSavePlan` and `GenerateAndSavePlanWithConfig` methods. The responsibility of prompt generation moves to `app.go`. Ensure `SaveToFile` remains functional, accepting the raw response string from the LLM. Remove dependencies on prompt-related types and interfaces.
    * **Depends On:** Task: Remove Prompt Package (`internal/prompt/`)
    * **AC Ref:** Plan Section 3 (Task 6)

- [ ] **Task: Update Output Tests (`internal/architect/output_test.go`)**
    * **Action:** Remove tests for `GenerateAndSavePlan*`. Update or add tests for `SaveToFile` if its core logic changed (likely minimal changes needed). Remove mocks related to prompt managers.
    * **Depends On:** Task: Refactor Output Writing (`internal/architect/output.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Unit Tests)

## Phase 4: Integration Testing & Documentation

- [ ] **Task: Update Integration Tests (`internal/integration/`)**
    * **Action:** Rewrite integration tests:
        * Use the new `--instructions` flag.
        * Provide sample static instruction files.
        * Provide sample context files/directories.
        * Mock the `GenerateContent` method of the Gemini client.
        * Within the mock, assert that the received `prompt` string has the correct `<instructions>...</instructions><context>...</context>` structure and includes correctly formatted context from the sample files.
        * Assert that the final output file written by `SaveToFile` contains exactly the mocked response content.
        * Remove tests for `--list-examples` and `--show-example`.
    * **Depends On:** Task: Refactor Core Application Flow (`internal/architect/app.go`), Task: Refactor Output Writing (`internal/architect/output.go`), Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 8), Plan Section 6 (Integration Tests)

- [ ] **Task: Update README.md**
    * **Action:** Thoroughly revise `README.md`: Update Usage section with new command examples (`architect --instructions ...`). Remove references to `--task-file`, `--prompt-template`, example templates. Update Configuration Options table. Explain the new "Instructions + Context Dump" philosophy.
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 9)

- [ ] **Task: Update CLAUDE.md**
    * **Action:** Update the Run command example in `CLAUDE.md` to reflect the new CLI structure.
    * **Depends On:** Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)
    * **AC Ref:** Plan Section 3 (Task 9)

- [ ] **Task: Verify CI Pipeline (`ci.yml`)**
    * **Action:** Run the CI pipeline. Ensure all linting, building, and rewritten tests pass. Verify test coverage meets requirements.
    * **Depends On:** Task: Update Integration Tests (`internal/integration/`), Task: Update CLI Tests (`cmd/architect/cli_test.go`), Task: Update `fileutil` Tests (`internal/fileutil/context_test.go`), Task: Update Config Tests (`internal/config/loader_test.go`), Task: Add Prompt Stitching Unit Tests
    * **AC Ref:** Plan Section 6 (CI Pipeline)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

* **(Assumption Confirmed)** Assumed that *no* dynamic substitution or templating whatsoever is required in the instructions file content.
* **(Assumption Confirmed)** Assumed `<instructions>...</instructions>` and `<context>...</context>` XML tags are the desired final prompt structure delimiters.
* **(Assumption Confirmed)** Assumed `[]FileMeta{Path, Content}` is the preferred structure for internal context representation.
* **Issue:** Need final decision on whether the `Format` field in `internal/fileutil/Config` is still needed/useful for formatting the context *during the stitching phase* or if that formatting logic should be hardcoded in the stitching function. (Current plan assumes stitching logic handles formatting using `<{path}>...</{path}>`).
