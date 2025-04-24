# Todo

## CLI & Configuration
- [x] **T001 · Feature · P2: add `--synthesis-model` cli flag**
    - **Context:** PLAN.md § 2.1.1 Add CLI Flag
    - **Action:**
        1. Add `synthesisModelFlag` using `flagSet.String` in `cmd/thinktank/cli.go`.
        2. Update CLI help/usage text to include the new flag description.
    - **Done‑when:**
        1. CLI accepts `--synthesis-model <model_name>`.
        2. CLI help output displays the new flag.
    - **Depends‑on:** none

- [x] **T002 · Feature · P2: add `SynthesisModel` field to config struct**
    - **Context:** PLAN.md § 2.1.2 Update Configuration
    - **Action:**
        1. Add `SynthesisModel string` field to `CliConfig` in `internal/config/config.go`.
    - **Done‑when:**
        1. `CliConfig` struct includes the `SynthesisModel` field.
    - **Depends‑on:** none

- [x] **T003 · Feature · P2: parse and store synthesis model flag in config**
    - **Context:** PLAN.md § 2.1.2 Update Configuration
    - **Action:**
        1. Update `ParseFlagsWithEnv` to assign `*synthesisModelFlag` to `cfg.SynthesisModel`.
    - **Done‑when:**
        1. The value passed via `--synthesis-model` is correctly stored in the `CliConfig` instance.
    - **Depends‑on:** [T001, T002]

- [x] **T004 · Feature · P2: add validation for synthesis model name**
    - **Context:** PLAN.md § 2.1.2 Update Configuration
    - **Action:**
        1. Update `ValidateInputsWithEnv` to check if `cfg.SynthesisModel` (if not empty) exists in the model registry.
    - **Done‑when:**
        1. `ValidateInputsWithEnv` returns an error if a non-empty `SynthesisModel` is provided but not found in the registry.
    - **Depends‑on:** [T003]

## Model Processing
- [x] **T005 · Refactor · P2: update `Processor.Process` method signature to return content**
    - **Context:** PLAN.md § 2.2.1 Update Process Method Signature
    - **Action:**
        1. Change `Process` method signature in `internal/thinktank/modelproc/processor.go` from `(...) error` to `(...) (string, error)`.
    - **Done‑when:**
        1. Method signature is updated in the interface and implementation.
        2. Code compiles.
    - **Depends‑on:** none

- [x] **T006 · Refactor · P2: implement content return in `Processor.Process`**
    - **Context:** PLAN.md § 2.2.1 Update Process Method Signature
    - **Action:**
        1. Modify the implementation of `Process` to return the `generatedOutput` string on success.
        2. Return an empty string along with the error if processing fails.
    - **Done‑when:**
        1. `Process` method returns the generated content string when successful.
        2. `Process` method returns an empty string when an error occurs.
    - **Depends‑on:** [T005]

## Orchestration
- [x] **T007 · Refactor · P2: define `modelResult` struct for orchestrator**
    - **Context:** PLAN.md § 2.3.1 Modify `processModelWithRateLimit`
    - **Action:**
        1. Define `type modelResult struct { modelName string; content string; err error }` within `internal/thinktank/orchestrator/orchestrator.go`.
    - **Done‑when:**
        1. The `modelResult` struct is defined.
    - **Depends‑on:** none

- [x] **T008 · Refactor · P2: update `processModelWithRateLimit` to use result struct**
    - **Context:** PLAN.md § 2.3.1 Modify `processModelWithRateLimit`
    - **Action:**
        1. Change the result channel from `chan error` to `chan modelResult`.
        2. Update goroutines to send a `modelResult` (containing name, content, or error) on the channel.
    - **Done‑when:**
        1. Channel type is updated and goroutines send `modelResult` structs.
        2. Correctly captures model outputs or errors.
    - **Depends‑on:** [T006, T007]

- [x] **T009 · Refactor · P2: update `processModels` to collect and return outputs/errors**
    - **Context:** PLAN.md § 2.3.2 Update `processModels`
    - **Action:**
        1. Change signature to return `(map[string]string, []error)`.
        2. Update implementation to read from the `modelResult` channel and collect content/errors.
    - **Done‑when:**
        1. Method correctly aggregates content and errors from the result channel.
        2. Returns both model outputs and errors.
    - **Depends‑on:** [T008]

- [x] **T010 · Refactor · P2: update `Orchestrator.Run` to call new `processModels`**
    - **Context:** PLAN.md § 2.3.3 Modify `Run` Method
    - **Action:**
        1. Modify the call to `o.processModels` to handle the new return signature.
        2. Store both outputs map and errors slice.
    - **Done‑when:**
        1. `Run` method correctly calls the updated `processModels` and receives both outputs and errors.
    - **Depends‑on:** [T009]

- [x] **T011 · Feature · P2: add conditional synthesis logic in `Orchestrator.Run`**
    - **Context:** PLAN.md § 2.3.3 Modify `Run` Method
    - **Action:**
        1. Add an `if o.config.SynthesisModel == ""` check after calling `processModels`.
        2. Define the structure for both "no synthesis" and "synthesis" branches.
    - **Done‑when:**
        1. Conditional logic exists based on `SynthesisModel` config value.
        2. Basic structure for both paths is in place.
    - **Depends‑on:** [T004, T010]

- [x] **T012 · Feature · P2: implement file writing without synthesis**
    - **Context:** PLAN.md § 2.3.3 Modify `Run` Method (No synthesis case)
    - **Action:**
        1. Implement the logic within the "no synthesis" branch.
        2. Iterate over `modelOutputs` map and save each model's content to its individual file.
    - **Done‑when:**
        1. Without synthesis, individual model output files are saved correctly.
    - **Depends‑on:** [T011]

## Synthesis Logic
- [x] **T013 · Feature · P2: create `StitchSynthesisPrompt` function**
    - **Context:** PLAN.md § 2.4.1 Add Prompt Function
    - **Action:**
        1. Implement `StitchSynthesisPrompt` in `internal/thinktank/prompt/prompt.go`.
        2. Format instructions and model outputs with proper delimiters.
    - **Done‑when:**
        1. Function correctly combines instructions and model outputs into a single prompt string.
    - **Depends‑on:** none

- [x] **T014 · Feature · P2: implement `synthesizeResults` method**
    - **Context:** PLAN.md § 2.4.2 Implement Synthesis Method
    - **Action:**
        1. Add `synthesizeResults` method to `Orchestrator`.
        2. Implement logic to build prompt, get client for synthesis model, call API, and process response.
    - **Done‑when:**
        1. Method successfully calls synthesis model and returns result or error.
    - **Depends‑on:** [T013]

- [x] **T015 · Feature · P2: implement synthesis file saving in `Orchestrator.Run`**
    - **Context:** PLAN.md § 2.3.3 Modify `Run` Method (Synthesis case)
    - **Action:**
        1. Implement the logic within the synthesis branch after calling `synthesizeResults`.
        2. Save synthesis output to a file named `<model-name>-synthesis.md`.
    - **Done‑when:**
        1. With synthesis, synthesized output is saved to a file with the correct naming convention.
    - **Depends‑on:** [T011, T014]

- [x] **T016 · Feature · P2: add error handling for synthesis failure**
    - **Context:** PLAN.md § 2.3.3 Modify `Run` Method (Error handling)
    - **Action:**
        1. Add error handling for synthesis API failures.
        2. Ensure errors are logged and propagated appropriately.
    - **Done‑when:**
        1. Errors from synthesis are handled gracefully and reported to the user.
    - **Depends‑on:** [T015]

- [x] **T017 · Feature · P2: add audit logging for synthesis process**
    - **Context:** PLAN.md § 2.4.3 Add Audit Logging
    - **Action:**
        1. Add structured log entries within `synthesizeResults` (start/end, model used, success/failure).
    - **Done‑when:**
        1. Appropriate log messages are generated during the synthesis process.
    - **Depends‑on:** [T014]

## Unit Testing
- [x] **T018 · Test · P2: add unit tests for CLI flag and config parsing**
    - **Context:** PLAN.md § 3.1.1 CLI and Config Tests
    - **Action:**
        1. Write tests for `--synthesis-model` flag parsing and validation.
        2. Test various scenarios (valid model, invalid model, empty).
    - **Done‑when:**
        1. Tests pass, confirming correct flag parsing and validation.
    - **Depends‑on:** [T004]

- [x] **T019 · Test · P2: add unit tests for `StitchSynthesisPrompt`**
    - **Context:** PLAN.md § 3.1.2 Prompt Tests
    - **Action:**
        1. Write tests for different input combinations (empty, single model, multiple models).
        2. Verify correct prompt structure is generated.
    - **Done‑when:**
        1. Tests pass, verifying correct prompt formatting.
    - **Depends‑on:** [T013]

- [ ] **T020 · Test · P2: add unit tests for `synthesizeResults`**
    - **Context:** PLAN.md § 3.1.3 Orchestrator Tests
    - **Action:**
        1. Write tests with mocked `APIService` to verify synthesis API calls.
        2. Test error handling for various failure points.
    - **Done‑when:**
        1. Tests pass, covering success and error scenarios.
    - **Depends‑on:** [T014]

- [ ] **T021 · Test · P2: add unit tests for orchestrator without synthesis**
    - **Context:** PLAN.md § 3.1.3 Orchestrator Tests
    - **Action:**
        1. Write tests for `Run` method without synthesis model.
        2. Verify individual files are created as expected.
    - **Done‑when:**
        1. Tests pass, confirming correct behavior without synthesis.
    - **Depends‑on:** [T012]

- [ ] **T022 · Test · P2: add unit tests for orchestrator with synthesis**
    - **Context:** PLAN.md § 3.1.3 Orchestrator Tests
    - **Action:**
        1. Write tests for `Run` method with synthesis model.
        2. Verify synthesis is called and output is saved properly.
    - **Done‑when:**
        1. Tests pass, confirming correct behavior with synthesis.
    - **Depends‑on:** [T015, T016]

## Integration Testing
- [ ] **T023 · Test · P1: add integration test for flow without synthesis**
    - **Context:** PLAN.md § 3.2.1 End-to-End Flow Tests
    - **Action:**
        1. Create test running without `--synthesis-model`.
        2. Verify multiple output files are created.
    - **Done‑when:**
        1. Test passes, confirming expected file outputs.
    - **Depends‑on:** [T012]

- [ ] **T024 · Test · P1: add integration test for flow with synthesis**
    - **Context:** PLAN.md § 3.2.1 End-to-End Flow Tests
    - **Action:**
        1. Create test running with `--synthesis-model`.
        2. Verify single synthesis output file is created.
    - **Done‑when:**
        1. Test passes, confirming synthesis output file.
    - **Depends‑on:** [T015]

- [ ] **T025 · Test · P2: add integration test for synthesis with model failures**
    - **Context:** PLAN.md § 3.2.1 End-to-End Flow Tests
    - **Action:**
        1. Create test with one primary model configured to fail.
        2. Verify synthesis proceeds with available outputs.
    - **Done‑when:**
        1. Test passes, confirming synthesis robustness to model failures.
    - **Depends‑on:** [T016]

- [ ] **T026 · Test · P2: add integration test for invalid synthesis model**
    - **Context:** PLAN.md § 3.2.2 Edge Case Tests
    - **Action:**
        1. Create test with invalid synthesis model name.
        2. Verify validation prevents execution.
    - **Done‑when:**
        1. Test passes, confirming validation rejects invalid model.
    - **Depends‑on:** [T004]

## E2E Testing
- [ ] **T027 · Test · P1: create E2E test file for synthesis feature**
    - **Context:** PLAN.md § 3.3.1 Create New E2E Test File
    - **Action:**
        1. Create `internal/e2e/cli_synthesis_test.go`.
        2. Set up basic structure and mock API config.
    - **Done‑when:**
        1. File exists with basic test structure.
    - **Depends‑on:** [T015]

- [ ] **T028 · Test · P1: add E2E test for basic synthesis workflow**
    - **Context:** PLAN.md § 3.3.2 Test Cases
    - **Action:**
        1. Implement test for full workflow with multiple models and synthesis.
        2. Configure mock APIs for all models.
    - **Done‑when:**
        1. Test passes, validating the complete synthesis workflow.
    - **Depends‑on:** [T027]

- [ ] **T029 · Test · P2: add E2E test for primary model failures**
    - **Context:** PLAN.md § 3.3.2 Test Cases
    - **Action:**
        1. Create test where some primary models fail.
        2. Verify synthesis handles partial successes.
    - **Done‑when:**
        1. Test passes, validating synthesis with partial model results.
    - **Depends‑on:** [T027]

## Documentation
- [ ] **T030 · Chore · P2: update README with synthesis flag info**
    - **Context:** PLAN.md § 4.1.1 Update README.md
    - **Action:**
        1. Add `--synthesis-model` flag to options table in README.
        2. Add explanation and example command.
    - **Done‑when:**
        1. README includes complete documentation for the new flag.
    - **Depends‑on:** [T001]

- [ ] **T031 · Chore · P2: update code documentation with GoDoc comments**
    - **Context:** PLAN.md § 4.2.1 Add/Update GoDoc Comments
    - **Action:**
        1. Add GoDoc comments for all new/modified components.
        2. Document the synthesis workflow in orchestrator.
    - **Done‑when:**
        1. All new methods, functions, and fields have clear documentation.
    - **Depends‑on:** [T002, T005, T009, T013, T014]

### Clarifications & Assumptions
- [x] **Issue:** decide on naming convention for the synthesis output file
    - **Context:** Plan doesn't specify exact filename format for synthesis output
    - **Resolution:** Use `<model-name>-synthesis.md` format for the synthesized output file
    - **Blocking?:** no
