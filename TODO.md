```markdown
# TODO

## CLI Flag Handling (Plan B)
- [x] **Define `stringSliceFlag` Type:**
    - **Action:** In `cmd/architect/cli.go` or a new `flags.go`, define a type `stringSliceFlag` as `[]string`.
    - **Depends On:** None
    - **AC Ref:** AC1, AC5
- [x] **Implement `flag.Value` Interface for `stringSliceFlag`:**
    - **Action:** Implement the `String() string` and `Set(value string) error` methods for `stringSliceFlag`. `Set` should append the value to the slice.
    - **Depends On:** Define `stringSliceFlag` Type
    - **AC Ref:** AC1, AC5
- [x] **Update `CliConfig` Struct:**
    - **Action:** In `cmd/architect/cli.go`, modify the `CliConfig` struct: change `ModelName string` to `ModelNames []string`, add `OutputDir string`, and remove `OutputFile string`. Also update the corresponding struct in `internal/architect/types.go` and the `convertToArchitectConfig` function in `cmd/architect/main.go`.
    - **Depends On:** None
    - **AC Ref:** AC1, AC3
- [x] **Register Repeatable `--model` Flag:**
    - **Action:** In `ParseFlagsWithEnv` (`cmd/architect/cli.go`), use `flagSet.Var()` to register the `stringSliceFlag` instance for the `--model` flag. Update the usage description to indicate it's repeatable. Remove the `defaultModel` constant usage for this flag's definition.
    - **Depends On:** Implement `flag.Value` Interface for `stringSliceFlag`, Update `CliConfig` Struct
    - **AC Ref:** AC1, AC5
- [ ] **Register `--output-dir` Flag:**
    - **Action:** In `ParseFlagsWithEnv` (`cmd/architect/cli.go`), add a new `flagSet.String()` for `--output-dir`, making it optional with a default empty string `""`.
    - **Depends On:** Update `CliConfig` Struct
    - **AC Ref:** AC3, AC5
- [ ] **Remove Old `--output` Flag:**
    - **Action:** In `ParseFlagsWithEnv` (`cmd/architect/cli.go`), remove the definition and parsing logic for the old `--output` flag.
    - **Depends On:** Update `CliConfig` Struct
    - **AC Ref:** AC4 (implicitly replaces single output file)
- [ ] **Update `ParseFlagsWithEnv` Population:**
    - **Action:** Modify `ParseFlagsWithEnv` (`cmd/architect/cli.go`) to correctly populate the `CliConfig.ModelNames` slice and `CliConfig.OutputDir` string from the parsed flags.
    - **Depends On:** Register Repeatable `--model` Flag, Register `--output-dir` Flag
    - **AC Ref:** AC1, AC3
- [ ] **Update `ValidateInputs` Logic:**
    - **Action:** In `ValidateInputs` (`cmd/architect/cli.go` and `internal/architect/app.go`), update the validation logic: check `len(config.ModelNames) > 0` (unless `DryRun` is true), return an error if empty. Remove any checks related to the old `config.OutputFile`.
    - **Depends On:** Update `CliConfig` Struct, Remove Old `--output` Flag
    - **AC Ref:** AC1
- [ ] **Update CLI Usage Message:**
    - **Action:** In `ParseFlagsWithEnv` (`cmd/architect/cli.go`), update the `flagSet.Usage` message to reflect the repeatable `--model` flag, the new `--output-dir` option, and remove references to the old `--output` flag.
    - **Depends On:** Register Repeatable `--model` Flag, Register `--output-dir` Flag, Remove Old `--output` Flag
    - **AC Ref:** AC1, AC3, AC5

## Run Name Generation
- [ ] **Implement `GenerateRunName` Utility:**
    - **Action:** Create a function `GenerateRunName()` (e.g., in a new `internal/runutil` package or a suitable existing utility location) that returns a random "adjective-noun" string using simple internal lists. Ensure hyphens are used.
    - **Depends On:** None
    - **AC Ref:** AC2

## Core Application Logic
- [ ] **Modify `Execute`/`RunInternal` Signatures:**
    - **Action:** Update the function signatures for `Execute` and `RunInternal` in `internal/architect/app.go` to accept the modified `CliConfig` struct (with `ModelNames` and `OutputDir`).
    - **Depends On:** Update `CliConfig` Struct
    - **AC Ref:** AC1, AC3
- [ ] **Implement Output Directory Determination:**
    - **Action:** In `Execute`/`RunInternal`, add logic at the beginning to determine the final output directory. If `cliConfig.OutputDir` is set, use it. Otherwise, call `GenerateRunName()`, construct the path in the CWD, and store it. Ensure the determined directory exists using `os.MkdirAll`. Log the determined path.
    - **Depends On:** Modify `Execute`/`RunInternal` Signatures, Implement `GenerateRunName` Utility
    - **AC Ref:** AC2, AC3
- [ ] **Implement Main Execution Loop:**
    - **Action:** In `Execute`/`RunInternal`, replace the single execution flow with a loop iterating over `cliConfig.ModelNames`.
    - **Depends On:** Modify `Execute`/`RunInternal` Signatures
    - **AC Ref:** AC1, AC4
- [ ] **Refactor `geminiClient` Initialization:**
    - **Action:** Move the `apiService.InitClient` call *inside* the main execution loop. Initialize a new `geminiClient` for *each* `modelName` in the loop. Ensure client resources (`Close()`) are managed correctly (e.g., defer close at the end of each iteration or manage collectively after the loop).
    - **Depends On:** Implement Main Execution Loop
    - **AC Ref:** AC1, AC4
- [ ] **Adapt Context Gathering:**
    - **Action:** Verify that context gathering (`contextGatherer.GatherContext`) can run once *before* the main loop. Ensure the gathered context (`contextFiles`, `stitchedPrompt`) is reused within the loop for each model.
    - **Depends On:** Implement Main Execution Loop
    - **AC Ref:** AC1, AC4
- [ ] **Adapt Core Steps within Loop:**
    - **Action:** Inside the loop, ensure prompt stitching (if needed per model, unlikely), token checking (`tokenManager.GetTokenInfo`), content generation (`geminiClient.GenerateContent`), and response processing (`apiService.ProcessResponse`) use the *current* model's `geminiClient` instance.
    - **Depends On:** Refactor `geminiClient` Initialization, Adapt Context Gathering
    - **AC Ref:** AC1, AC4
- [ ] **Implement Per-Model Output Path Construction:**
    - **Action:** Inside the loop, construct the specific output file path for the current model using `filepath.Join(determinedOutputDir, modelName+".md")`. Include basic sanitization for `modelName` (e.g., replace `/` with `-`).
    - **Depends On:** Implement Output Directory Determination, Implement Main Execution Loop
    - **AC Ref:** AC4
- [ ] **Call `savePlanToFile` within Loop:**
    - **Action:** Inside the loop, call the existing `savePlanToFile` helper function (or the underlying `FileWriter.SaveToFile`) with the generated content and the specific output path constructed for the current model.
    - **Depends On:** Implement Per-Model Output Path Construction, Adapt Core Steps within Loop
    - **AC Ref:** AC4
- [ ] **Implement Per-Model Error Handling:**
    - **Action:** Inside the loop, wrap the processing steps for a single model in error handling. If an error occurs (API error, token limit, etc.), log it clearly (including the model name) but allow the loop to *continue* to the next model. Consider collecting errors.
    - **Depends On:** Implement Main Execution Loop
    - **AC Ref:** AC1, AC4
- [ ] **Remove Single `OutputFile` Logic:**
    - **Action:** Remove all code related to handling the single `cliConfig.OutputFile` from `Execute`/`RunInternal`, as output is now handled per-model within the loop.
    - **Depends On:** Call `savePlanToFile` within Loop
    - **AC Ref:** AC4
- [ ] **Update Audit Logging:**
    - **Action:** Review and update all `auditLogger.Log` calls within `Execute`/`RunInternal` to include the current `modelName` and specific `outputFile` path where relevant, especially for operations happening inside the loop (e.g., GenerateContent, SaveOutput). Adjust `ExecuteStart`/`ExecuteEnd` logs to reflect the multi-model nature.
    - **Depends On:** Implement Main Execution Loop
    - **AC Ref:** AC1, AC4

## Output Saving
- [ ] **Verify `FileWriter.SaveToFile` Compatibility:**
    - **Action:** Review the implementation of `savePlanToFile` and `FileWriter.SaveToFile` in `internal/architect/app.go` and `internal/architect/output.go`. Confirm that it correctly handles creating nested directories (e.g., `run-name/model.md`) using `os.MkdirAll` or similar. Make adjustments if necessary.
    - **Depends On:** Call `savePlanToFile` within Loop
    - **AC Ref:** AC3, AC4

## Testing
- [ ] **Write Unit Tests for `stringSliceFlag`:**
    - **Action:** Create unit tests for the `stringSliceFlag` type, covering its `String()` and `Set()` methods.
    - **Depends On:** Implement `flag.Value` Interface for `stringSliceFlag`
    - **AC Ref:** AC1, AC5
- [ ] **Write Unit Tests for `GenerateRunName`:**
    - **Action:** Create unit tests for the `GenerateRunName` utility, ensuring it produces output in the expected format.
    - **Depends On:** Implement `GenerateRunName` Utility
    - **AC Ref:** AC2
- [ ] **Update CLI Parsing Unit Tests:**
    - **Action:** Update existing unit tests for `ParseFlagsWithEnv` and `ValidateInputs` to cover the new `--model` (repeatable), `--output-dir` flags, removal of `--output`, and updated validation rules.
    - **Depends On:** Update `ParseFlagsWithEnv` Population, Update `ValidateInputs` Logic
    - **AC Ref:** AC1, AC3, AC5
- [ ] **Write/Update Integration Tests for `Execute`/`RunInternal`:**
    - **Action:** Add or modify integration tests for `Execute`/`RunInternal` to cover:
        - Running with a single `--model`.
        - Running with multiple `--model` flags.
        - Outputting to a specified `--output-dir`.
        - Outputting to an auto-generated run name directory.
        - Correct naming of output files (`modelname.md`).
        - Scenario where one model fails but others succeed.
    - **Depends On:** Implement Per-Model Error Handling, Call `savePlanToFile` within Loop
    - **AC Ref:** AC1, AC2, AC3, AC4

## Documentation
- [ ] **Update `README.md`:**
    - **Action:** Modify `README.md` to update usage examples, showing the repeatable `--model` flag and the `--output-dir` option. Explain the new output directory structure and per-model file naming. Remove references to the old `--output` flag.
    - **Depends On:** Write/Update Integration Tests for `Execute`/`RunInternal` (to confirm final behavior)
    - **AC Ref:** AC1, AC2, AC3, AC4, AC5
- [ ] **Create Multi-Model ADR:**
    - **Action:** Create a new ADR file (e.g., `docs/adrs/00X-support-multiple-models.md`) detailing the decision to use Plan B (repeated flags), the introduction of the run name/output directory concept, and the rationale based on project principles.
    - **Depends On:** Update `README.md`
    - **AC Ref:** AC1, AC2, AC3, AC4, AC5

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** Assumed context gathering (`GatherContext`) is model-independent and can run once before the main loop.
    - **Context:** PLAN.md Actionable Plan Step 4 suggests this possibility. Needs confirmation based on `GatherContext` implementation details (e.g., does it use the client in a model-specific way *during* gathering?).
- [ ] **Issue/Assumption:** Assumed `gemini.Client` must be re-initialized per model inside the loop.
    - **Context:** PLAN.md Actionable Plan Step 4 explicitly recommends this approach for correctness and isolation.
- [ ] **Issue/Assumption:** Assumed model names from flags are suitable for filenames after basic sanitization (e.g., replacing `/` with `-`).
    - **Context:** PLAN.md Actionable Plan Step 4 mentions sanitization. Need to define the exact rules.
- [ ] **Issue/Assumption:** Assumed `FileWriter.SaveToFile` already handles nested directory creation correctly.
    - **Context:** PLAN.md Actionable Plan Step 5 mentions reviewing this. Requires code verification.
- [ ] **Issue/Assumption:** Error handling within the loop should log the error for the specific model and continue processing other models. A summary error mechanism is mentioned but not defined.
    - **Context:** PLAN.md Actionable Plan Step 4 describes logging and continuing. The exact behavior for summarizing multiple errors at the end needs clarification if required beyond individual logging.
```