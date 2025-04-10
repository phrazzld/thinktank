# TODO

## Configuration and Flag Removal
- [x] **Remove ClarifyTask field from Configuration struct in main.go**
  - **Action:** Edit main.go to remove the ClarifyTask field from the Configuration struct declaration.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Configuration Structures), Step 2.2.

- [x] **Remove ClarifyTask field from AppConfig struct in config.go**
  - **Action:** Edit internal/config/config.go to remove the ClarifyTask field including mapstructure and toml tags.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Configuration Structures), Step 2.3.

- [x] **Remove Clarify and Refine fields from TemplateConfig struct in config.go**
  - **Action:** Edit internal/config/config.go to remove the Clarify and Refine string fields from the TemplateConfig struct.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Configuration Structures), Step 2.4.

- [x] **Remove clarify flag declaration in parseFlags()**
  - **Action:** Edit main.go to remove the clarifyTaskFlag declaration in the parseFlags() function.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Flag Declaration and Parsing), Step 2.1.

- [x] **Remove ClarifyTask assignment in parseFlags()**
  - **Action:** Edit main.go to remove the assignment to config.ClarifyTask in the parseFlags() function.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Flag Declaration and Parsing), Step 2.1.

- [x] **Remove clarify_task entry in convertConfigToMap()**
  - **Action:** Edit main.go to remove the clarify_task entry in the convertConfigToMap() function.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Flag Declaration and Parsing).

- [x] **Remove clarify-related logic in backfillConfigFromAppConfig()**
  - **Action:** Edit main.go to remove the code that handles ClarifyTask in the backfillConfigFromAppConfig() function.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Flag Declaration and Parsing).

- [x] **Update DefaultConfig() to remove clarify template references**
  - **Action:** Edit internal/config/config.go to remove Clarify and Refine template references from the DefaultConfig() function.
  - **Depends On:** "Remove Clarify and Refine fields from TemplateConfig struct in config.go"
  - **AC Ref:** Implementation Steps 1 (Configuration Structures), Step 2.5.

## Function and Code Path Removal
- [x] **Remove clarifyTaskDescription() function from main.go**
  - **Action:** Delete the entire clarifyTaskDescription() function from main.go.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.1.

- [x] **Remove clarifyTaskDescriptionWithConfig() function from main.go**
  - **Action:** Delete the entire clarifyTaskDescriptionWithConfig() function from main.go.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.1.

- [x] **Remove clarifyTaskDescriptionWithPromptManager() function from main.go**
  - **Action:** Delete the entire clarifyTaskDescriptionWithPromptManager() function from main.go.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.1.

- [ ] **Remove clarify condition in the main function**
  - **Action:** Delete the if block `if config.ClarifyTask && !config.DryRun {...}` from the main function.
  - **Depends On:** "Remove clarifyTaskDescriptionWithConfig() function from main.go"
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.2.

- [x] **Update SetupPromptManagerWithConfig to remove clarify/refine templates**
  - **Action:** Modify internal/prompt/integration.go to only load default.tmpl and remove clarify.tmpl and refine.tmpl from the template loading loop.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.3.

- [ ] **Remove simulateClarifyTaskDescription function in main_adapter.go**
  - **Action:** Delete the entire simulateClarifyTaskDescription function from internal/integration/main_adapter.go.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.4.

- [ ] **Remove clarify condition in main_adapter.go**
  - **Action:** Delete the if block that calls simulateClarifyTaskDescription in internal/integration/main_adapter.go.
  - **Depends On:** "Remove simulateClarifyTaskDescription function in main_adapter.go"
  - **AC Ref:** Implementation Steps 1 (Functions and Code Paths), Step 3.4.

## Template File Removal
- [ ] **Delete clarify.tmpl**
  - **Action:** Delete the file internal/prompt/templates/clarify.tmpl.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Template Files), Step 4.

- [ ] **Delete refine.tmpl**
  - **Action:** Delete the file internal/prompt/templates/refine.tmpl.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Template Files), Step 4.

## Documentation Updates
- [ ] **Remove clarify flag from README.md configuration table**
  - **Action:** Edit README.md to remove the `--clarify` entry from the configuration options table.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Documentation), Step 5.1.

- [ ] **Remove Task Clarification section from README.md**
  - **Action:** Delete the entire "Task Clarification" section from README.md.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Documentation), Step 5.1.

- [ ] **Remove clarify examples from README.md**
  - **Action:** Remove any examples in README.md that use the `--clarify` flag.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Documentation), Step 5.1.

- [ ] **Check and update other documentation files**
  - **Action:** Search for and update any other documentation files that reference the clarify functionality.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Documentation), Step 5.2.

## Test Updates
- [ ] **Update integration_test.go to remove clarify-related tests**
  - **Action:** Modify internal/integration/integration_test.go to remove any tests specifically for the clarify feature.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Test Code), Step 6.1.

- [ ] **Update main_test.go to remove clarify-related tests**
  - **Action:** Modify main_test.go to remove any tests that involve the clarify flag.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Test Code), Step 6.1.

- [ ] **Update other test files if needed**
  - **Action:** Check and update any other test files that might contain clarify-related test cases.
  - **Depends On:** None.
  - **AC Ref:** Implementation Steps 1 (Test Code), Step 6.1.

## Verification
- [ ] **Compile and build the application**
  - **Action:** Run `go build` to verify the application compiles without errors after all changes.
  - **Depends On:** All previous tasks.
  - **AC Ref:** Verification Steps 1.

- [ ] **Test CLI functionality**
  - **Action:** Run the application with various flag combinations to ensure it still works correctly.
  - **Depends On:** "Compile and build the application"
  - **AC Ref:** Verification Steps 2.

- [ ] **Check help text**
  - **Action:** Run the application with `--help` to verify the clarify flag is no longer shown.
  - **Depends On:** "Compile and build the application"
  - **AC Ref:** Verification Steps 3.

- [ ] **Run tests**
  - **Action:** Execute `go test ./...` to ensure all tests pass after the changes.
  - **Depends On:** "Compile and build the application"
  - **AC Ref:** Verification Steps 4.

- [ ] **Search for remaining clarify references**
  - **Action:** Use grep or similar to search for any remaining references to "clarify" in the codebase.
  - **Depends On:** All code modification tasks.
  - **AC Ref:** Verification Steps 5.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Assumption: No UI components affected**
  - **Context:** The plan doesn't mention any UI components (other than CLI) that might be affected by removing the clarify feature. I'm assuming there are no UI interfaces that need updating beyond the CLI help text.

- [ ] **Clarification: Extent of test coverage**
  - **Context:** The plan states to "Update test files to remove tests specific to the clarify feature" but doesn't specify which test files contain clarify-related tests. We should verify all test files that might reference the clarify functionality.

- [ ] **Assumption: No runtime feature toggles**
  - **Context:** I'm assuming there are no runtime feature toggles or configuration files that control the clarify feature's availability that would need to be updated.

- [ ] **Assumption: No database impacts**
  - **Context:** The plan doesn't mention any database schema or data changes. I'm assuming the clarify feature doesn't store any data that would need to be migrated or cleaned up.

- [ ] **Clarification: Embedded template handling**
  - **Context:** The code references embedded templates in prompt.go. We need to ensure that removing the clarify.tmpl and refine.tmpl files doesn't cause issues with the template embedding system.