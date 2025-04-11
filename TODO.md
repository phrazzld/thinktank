# TODO

## Preparation
- [x] **Conduct comprehensive search for clarify references**
  - **Action:** Run `grep -rE 'clarify|ClarifyTask|clarifyTaskFlag'` across the entire project and document all relevant files, functions, and code blocks.
  - **Depends On:** None
  - **AC Ref:** AC 1.7

## CLI Flag Removal
- [x] **Remove clarify flag definition**
  - **Action:** In `cmd/architect/cli.go`, delete the `clarifyTaskFlag` variable definition (line ~98).
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.1

- [x] **Remove ClarifyTask field from CliConfig struct**
  - **Action:** In `cmd/architect/cli.go`, delete the `ClarifyTask bool` field from the CliConfig struct (line ~39).
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.2

- [x] **Remove clarify key-value pair from ConvertConfigToMap**
  - **Action:** In `cmd/architect/cli.go`, delete the `"clarify": cliConfig.ClarifyTask` entry from the ConvertConfigToMap function (line ~209).
  - **Depends On:** Remove ClarifyTask field from CliConfig struct
  - **AC Ref:** AC 1.2

- [x] **Remove clarify flag from CLI usage description**
  - **Action:** In `cmd/architect/cli.go`, delete the `--clarify` flag description from the flagSet.Usage function.
  - **Depends On:** Remove clarify flag definition
  - **AC Ref:** AC 1.1, AC 1.6

## Internal Configuration Removal
- [x] **Remove ClarifyTask field from internal CliConfig**
  - **Action:** If present in `internal/architect/types.go`, delete the `ClarifyTask` field from any config struct.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.2

- [x] **Remove fields from AppConfig in config.go**
  - **Action:** In `internal/config/config.go`, check for and remove any `ClarifyTask` field from the AppConfig struct if present.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.2

- [x] **Remove fields from TemplateConfig**
  - **Action:** In `internal/config/config.go`, delete the `Clarify` and `Refine` fields from the TemplateConfig struct if present.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.2

- [x] **Update example config file**
  - **Action:** In `internal/config/example_config.toml`, remove the `clarify_task = false` line and the `clarify = "clarify.tmpl"` and `refine = "refine.tmpl"` lines under the [templates] section.
  - **Depends On:** Remove fields from TemplateConfig
  - **AC Ref:** AC 1.2

## Core Logic Removal
- [x] **Remove conditional clarify logic from execution flow**
  - **Action:** In `internal/architect/app.go`, locate and remove any `if cliConfig.ClarifyTask` blocks or function calls related to the clarify feature. Ensure the standard execution path remains intact.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.3

## Template Removal
- [x] **Delete clarify template files**
  - **Action:** Check `internal/prompt/templates/` for `clarify.tmpl` and `refine.tmpl` and delete these files if found.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.4

- [x] **Check and delete example templates**
  - **Action:** Check `internal/prompt/templates/examples/` for any templates related to clarify and delete if found.
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.4

- [x] **Update embed.go if necessary**
  - **Action:** If needed, update `internal/prompt/embed.go` to remove references to deleted templates (though go:embed should handle missing files gracefully).
  - **Depends On:** Delete clarify template files, Check and delete example templates
  - **AC Ref:** AC 1.4

## Documentation Updates
- [x] **Remove clarify feature from README**
  - **Action:** Edit `README.md` to remove the "Task Clarification" feature from the Features list (line ~33).
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.6

- [x] **Remove clarify example from README**
  - **Action:** Edit `README.md` to remove the `--clarify` usage example (lines ~87-88).
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.6

- [x] **Remove clarify from options table in README**
  - **Action:** Edit `README.md` to remove the `--clarify` row from the Configuration Options table (line ~115).
  - **Depends On:** Conduct comprehensive search for clarify references
  - **AC Ref:** AC 1.6

- [ ] **Update BACKLOG.md**
  - **Action:** Edit `BACKLOG.md` to move the "purge the program of the clarify flag and feature and related code and tests" line to a "Done" section or remove it entirely.
  - **Depends On:** All other tasks completed
  - **AC Ref:** AC 1.6

## Test Updates
- [x] **Update CLI flag tests**
  - **Action:** In `cmd/architect/cli_test.go`, remove any test cases specifically validating the parsing or behavior of the `--clarify` flag.
  - **Depends On:** Remove ClarifyTask field from CliConfig struct
  - **AC Ref:** AC 1.5, AC 1.8

- [x] **Remove TestConvertConfigNoClarity**
  - **Action:** In `cmd/architect/flags_test.go`, delete the `TestConvertConfigNoClarity` test as it tests a field that will no longer exist.
  - **Depends On:** Remove clarify key-value pair from ConvertConfigToMap
  - **AC Ref:** AC 1.5, AC 1.8

- [ ] **Update legacy config tests**
  - **Action:** In `internal/config/legacy_config_test.go`, review and update tests to ensure they still pass when loading legacy config files with clarify-related fields. The test should verify these fields don't cause errors and are ignored.
  - **Depends On:** Remove fields from TemplateConfig
  - **AC Ref:** AC 1.5, AC 1.8

- [ ] **Update integration tests**
  - **Action:** In `internal/integration/integration_test.go`, rename `TestTaskClarification` to `TestTaskExecution` if it exists and update its logic to test standard execution flow without clarification. Remove any clarify-specific setup or assertions.
  - **Depends On:** Remove conditional clarify logic from execution flow
  - **AC Ref:** AC 1.5, AC 1.8

## Validation and Final Cleanup
- [ ] **Run unit tests**
  - **Action:** Run `go test ./...` and fix any test failures resulting from the removed code.
  - **Depends On:** All test update tasks
  - **AC Ref:** AC 1.8, AC 1.9

- [ ] **Clean up Go dependencies**
  - **Action:** Run `go mod tidy` to ensure no unused dependencies remain.
  - **Depends On:** All code removal tasks
  - **AC Ref:** AC 1.7

- [ ] **Run go vet**
  - **Action:** Run `go vet ./...` to check for any issues introduced by the code changes.
  - **Depends On:** Run unit tests
  - **AC Ref:** AC 1.8, AC 1.9

- [ ] **Run linter**
  - **Action:** Run `golangci-lint run` (or use `hooks/pre-commit`) to ensure code quality and catch any issues.
  - **Depends On:** Run go vet
  - **AC Ref:** AC 1.8, AC 1.9

- [ ] **Verify flag removal in CLI help**
  - **Action:** Run `go run main.go --help` and verify the `--clarify` flag no longer appears in the output.
  - **Depends On:** Remove clarify flag definition
  - **AC Ref:** AC 1.1, AC 1.9

- [ ] **Perform final grep search**
  - **Action:** Run a final grep search for any remaining references to "clarify", "ClarifyTask", etc., to ensure complete removal.
  - **Depends On:** All code, documentation, and test update tasks
  - **AC Ref:** AC 1.7

- [ ] **Test basic application functionality**
  - **Action:** Run `go run main.go --task-file task.txt ./` to verify that plan generation works correctly without the clarify feature.
  - **Depends On:** All validation tasks
  - **AC Ref:** AC 1.9

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Assumption: No external dependencies on clarify**
  - **Context:** The plan assumes that no external components or integrations rely on the clarify feature. If there are external clients or systems using this feature, additional communication might be needed.

- [ ] **Assumption: TestTaskClarification exists**
  - **Context:** The plan mentions renaming `TestTaskClarification` to `TestTaskExecution` in integration_test.go, but we need to confirm this test actually exists before trying to modify it.

- [ ] **Assumption: Limited refactoring needed**
  - **Context:** The plan assumes that removing the clarify feature is mostly about deletion rather than significant refactoring. If the clarify logic is deeply intertwined with core functionality, more careful refactoring might be required.