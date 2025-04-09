# TODO

## Spinner Removal

- [x] **Identify all spinner usage in main.go**
  - **Action:** Scan main.go for all instances of spinnerInstance and document their locations and usage patterns to ensure complete replacement.
  - **Depends On:** None.
  - **AC Ref:** Key Step 1.

- [x] **Replace spinner start/stop messages with logger in main.go**
  - **Action:** Replace all `spinnerInstance.Start(msg)` and `spinnerInstance.Stop(msg)` calls with equivalent `logger.Info(msg)` calls in main.go, preserving the informational content.
  - **Depends On:** Identify all spinner usage in main.go.
  - **AC Ref:** Key Step 1.

- [x] **Replace spinner update message calls with logger in main.go**
  - **Action:** Replace all `spinnerInstance.UpdateMessage(msg)` calls with equivalent `logger.Info(msg)` calls in main.go, maintaining the same level of user feedback.
  - **Depends On:** Identify all spinner usage in main.go.
  - **AC Ref:** Key Step 1.

- [x] **Replace spinner error handling with logger in main.go**
  - **Action:** Replace all `spinnerInstance.StopFail(msg)` calls with equivalent `logger.Error(msg)` calls in main.go to preserve error reporting.
  - **Depends On:** Identify all spinner usage in main.go.
  - **AC Ref:** Key Step 1.

- [x] **Ensure debug-level logging is preserved**
  - **Action:** Review spinner code to identify any debug-level logging and ensure equivalent `logger.Debug(msg)` calls are maintained in the replacement code.
  - **Depends On:** Identify all spinner usage in main.go.
  - **AC Ref:** Key Step 1.

- [x] **Remove initSpinner function from main.go**
  - **Action:** Delete the entire `initSpinner` function implementation from main.go.
  - **Depends On:** Replace spinner start/stop messages with logger in main.go, Replace spinner update message calls with logger in main.go, Replace spinner error handling with logger in main.go.
  - **AC Ref:** Key Step 2.

- [x] **Remove no-spinner flag from parseFlags in main.go**
  - **Action:** Delete the `--no-spinner` flag definition and related code from the `parseFlags` function in main.go.
  - **Depends On:** Remove initSpinner function from main.go.
  - **AC Ref:** Key Step 2.

- [ ] **Remove NoSpinner field from Configuration struct in main.go**
  - **Action:** Remove the `NoSpinner` boolean field from the `Configuration` struct definition in main.go.
  - **Depends On:** Remove no-spinner flag from parseFlags in main.go.
  - **AC Ref:** Key Step 2.

- [ ] **Remove NoSpinner field from AppConfig struct**
  - **Action:** Remove the `NoSpinner` boolean field from the `AppConfig` struct definition in internal/config/config.go.
  - **Depends On:** Remove NoSpinner field from Configuration struct in main.go.
  - **AC Ref:** Key Step 2.

- [ ] **Remove NoSpinner handling in convertConfigToMap**
  - **Action:** Remove the `"no_spinner": cliConfig.NoSpinner,` entry from the map returned by the `convertConfigToMap` function in main.go.
  - **Depends On:** Remove NoSpinner field from AppConfig struct.
  - **AC Ref:** Key Step 2.

- [ ] **Remove NoSpinner handling in backfillConfigFromAppConfig**
  - **Action:** Remove the conditional block that checks and assigns NoSpinner in the `backfillConfigFromAppConfig` function in main.go.
  - **Depends On:** Remove NoSpinner handling in convertConfigToMap.
  - **AC Ref:** Key Step 2.

- [ ] **Delete internal/spinner directory**
  - **Action:** Remove the entire internal/spinner directory, including spinner.go and any other files within it.
  - **Depends On:** Replace spinner start/stop messages with logger in main.go, Replace spinner update message calls with logger in main.go, Replace spinner error handling with logger in main.go, Remove initSpinner function from main.go.
  - **AC Ref:** Key Step 3.

- [ ] **Remove NoSpinner field from main_adapter.go**
  - **Action:** Remove the `NoSpinner` boolean field from the `Configuration` struct definition in internal/integration/main_adapter.go.
  - **Depends On:** Remove NoSpinner field from Configuration struct in main.go.
  - **AC Ref:** Key Step 4.

- [ ] **Update integration tests related to spinner functionality**
  - **Action:** Identify and update any integration tests that use or reference spinner functionality to ensure they continue to work with the logging-based approach.
  - **Depends On:** Remove NoSpinner field from main_adapter.go.
  - **AC Ref:** Key Step 4.

- [ ] **Remove spinner dependency from go.mod and go.sum**
  - **Action:** Run `go mod tidy` to remove the `github.com/briandowns/spinner` dependency from go.mod and go.sum files.
  - **Depends On:** Delete internal/spinner directory.
  - **AC Ref:** Key Step 5.

- [ ] **Remove spinner task from BACKLOG.md**
  - **Action:** Delete the "completely remove the spinner" entry from BACKLOG.md.
  - **Depends On:** None (can be done independently).
  - **AC Ref:** Key Step 6.

- [ ] **Verify all tests pass after spinner removal**
  - **Action:** Run the full test suite to ensure all functionality works correctly without the spinner.
  - **Depends On:** Replace spinner start/stop messages with logger in main.go, Replace spinner update message calls with logger in main.go, Replace spinner error handling with logger in main.go, Remove NoSpinner field from main_adapter.go, Update integration tests related to spinner functionality.
  - **AC Ref:** Testing Strategy.

- [ ] **Validate logging coverage**
  - **Action:** Manually test the application to ensure log messages provide adequate user feedback during operations that previously used the spinner.
  - **Depends On:** Replace spinner start/stop messages with logger in main.go, Replace spinner update message calls with logger in main.go, Replace spinner error handling with logger in main.go.
  - **AC Ref:** Testing Strategy, Implementation Notes.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Assumption: No direct import of spinner package outside main.go and main_adapter.go**
  - **Context:** The plan assumes that the spinner package is only used in main.go and main_adapter.go. If other files also import and use it, they would need to be modified as well.

- [ ] **Assumption: No custom test mocks for spinner**
  - **Context:** The plan does not explicitly mention checking for or adjusting any test mocks that might be related to the spinner. It's assumed there are no test-specific mocks for the spinner that would need updating.

- [ ] **Assumption: Debug logging equivalence**
  - **Context:** Key Step 1 mentions preserving debug level messages from the spinner. It's assumed that these messages should map 1:1 to logger.Debug calls without modification.

- [ ] **Assumption: No UI/UX changes needed outside of direct replacement**
  - **Context:** The plan assumes that directly replacing spinner messages with log messages will be sufficient for user feedback. No additional UI/UX enhancements are mentioned to compensate for the loss of the visual spinner.