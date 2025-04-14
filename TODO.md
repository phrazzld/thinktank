# TODO

## 1. Analyze Current Coverage in Detail
- [x] **Task Title:** Run Detailed Coverage Analysis
  - **Action:** Execute `go test -coverprofile=coverage.out ./...`, `go tool cover -func=coverage.out`, and `go tool cover -html=coverage.out` to generate baseline coverage reports. Analyze the HTML report to understand current coverage distribution.
  - **Depends On:** None
  - **AC Ref:** Plan Section 1

- [x] **Task Title:** Identify Specific Low-Coverage Functions
  - **Action:** Review the generated coverage reports (function and HTML) to pinpoint specific functions and code paths with low or zero coverage, especially within `internal/gemini`, `internal/architect/adapters.go`, `internal/config`, `internal/fileutil`, and `cmd/architect`. Document these areas.
  - **Depends On:** Run Detailed Coverage Analysis
  - **AC Ref:** Plan Section 1

## 2. Add Tests for Gemini Package
- [x] **Task Title:** Create `internal/gemini/errors_test.go` File
  - **Action:** Create a new test file `errors_test.go` within the `internal/gemini` package.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 2.1

- [x] **Task Title:** Implement Basic Error Tests
  - **Action:** Add unit tests for `APIError.Error()`, `APIError.Unwrap()`, `APIError.UserFacingError()`, and `APIError.DebugInfo()` to verify they return the expected outputs.
  - **Depends On:** Create `internal/gemini/errors_test.go` File
  - **AC Ref:** Plan Section 2.1

- [x] **Task Title:** Implement Error Classification Tests
  - **Action:** Add unit tests for `IsAPIError()` and `GetErrorType()` to verify they correctly identify API errors and error types based on different error messages and status codes.
  - **Depends On:** Create `internal/gemini/errors_test.go` File
  - **AC Ref:** Plan Section 2.1

- [x] **Task Title:** Implement Table-Driven Tests for `FormatAPIError()`
  - **Action:** Create comprehensive table-driven tests for `FormatAPIError()` covering different error types, status codes, and verifying the resulting `APIError` fields (Type, Message, Suggestion).
  - **Depends On:** Implement Error Classification Tests
  - **AC Ref:** Plan Section 2.1

- [x] **Task Title:** Extend `client_test.go` for Constructor and Defaults
  - **Action:** Add or improve tests in `internal/gemini/client_test.go` to cover `NewClient()` initialization logic (including error cases like missing API key/model name) and `DefaultModelConfig()` correctness.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 2.2

- [x] **Task Title:** Enhance Mock Client Tests
  - **Action:** Review and enhance existing tests for any mock client implementation to ensure comprehensive coverage of its methods and behaviors.
  - **Depends On:** Extend `client_test.go` for Constructor and Defaults
  - **AC Ref:** Plan Section 2.2

- [x] **Task Title:** Refactor `geminiClient` for HTTP Client Injection
  - **Action:** Modify `internal/gemini/gemini_client.go` struct to accept an `httpClient` interface (e.g., `interface{ Do(*http.Request) (*http.Response, error) }`) via its constructor. Update internal methods like `fetchModelInfo` to use this injected client. Ensure the default `http.Client` is used when no custom client is provided.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Create `internal/gemini/gemini_client_test.go` File
  - **Action:** Create a new test file `gemini_client_test.go` within the `internal/gemini` package.
  - **Depends On:** Refactor `geminiClient` for HTTP Client Injection
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Implement Mock HTTP Transport for Tests
  - **Action:** Create a mock HTTP client or transport that allows simulating various HTTP responses (success, errors, different status codes, specific body content) for testing `geminiClient` methods that make HTTP calls.
  - **Depends On:** Create `internal/gemini/gemini_client_test.go` File
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Implement `GenerateContent` Tests
  - **Action:** Add tests for the `GenerateContent` method using the mock HTTP client to simulate successful responses, API errors (e.g., rate limits, auth errors), empty responses, and safety filter responses. Verify the returned `GenerationResult` or `APIError`.
  - **Depends On:** Implement Mock HTTP Transport for Tests
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Implement `CountTokens` Tests
  - **Action:** Add tests for the `CountTokens` method using the mock HTTP client to simulate successful responses and API errors. Verify the returned `TokenCount` or `APIError`. Test the empty prompt case.
  - **Depends On:** Implement Mock HTTP Transport for Tests
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Implement `GetModelInfo` Tests
  - **Action:** Add tests for the `GetModelInfo` method using the mock HTTP client to simulate successful responses, API errors, and invalid JSON responses. Test the caching mechanism. Verify the returned `ModelInfo` or `APIError`.
  - **Depends On:** Implement Mock HTTP Transport for Tests
  - **AC Ref:** Plan Section 2.3

- [x] **Task Title:** Implement Helper Method Tests
  - **Action:** Add tests for helper methods like `mapSafetyRatings`, `GetModelName`, `GetTemperature`, `GetMaxOutputTokens`, and `GetTopP`. These might not require the mock HTTP client.
  - **Depends On:** Create `internal/gemini/gemini_client_test.go` File
  - **AC Ref:** Plan Section 2.3

## 3. Implement Tests for Adapter Code
- [x] **Task Title:** Create `internal/architect/adapters_test.go` File
  - **Action:** Create a new test file `adapters_test.go` within the `internal/architect` package.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 3.1

- [x] **Task Title:** Implement Client Initialization Adapter Tests
  - **Action:** Add tests for `APIServiceAdapter.InitClient`. Use simple mocks/fakes for the underlying `APIService` to verify the adapter correctly passes arguments and returns values/errors.
  - **Depends On:** Create `internal/architect/adapters_test.go` File
  - **AC Ref:** Plan Section 3.1

- [x] **Task Title:** Implement Response Processing Adapter Tests
  - **Action:** Add tests for `APIServiceAdapter.ProcessResponse`. Use mocks/fakes for `APIService` and provide sample `gemini.GenerationResult` inputs to verify the adapter's behavior.
  - **Depends On:** Create `internal/architect/adapters_test.go` File
  - **AC Ref:** Plan Section 3.1

- [x] **Task Title:** Implement Error Handling Adapter Tests
  - **Action:** Add tests for `APIServiceAdapter.IsEmptyResponseError`, `IsSafetyBlockedError`, and `GetErrorDetails`. Use mocks/fakes for `APIService` and sample errors to verify correct delegation and return values.
  - **Depends On:** Create `internal/architect/adapters_test.go` File
  - **AC Ref:** Plan Section 3.1

- [x] **Task Title:** Implement Token-Related Adapter Tests
  - **Action:** Add tests for `TokenResultAdapter` (verify correct field mapping) and `TokenManagerAdapter` methods (`CheckTokenLimit`, `GetTokenInfo`, `PromptForConfirmation`). Use mocks/fakes for the underlying `TokenManager` to verify argument passing, return value conversion, and delegation.
  - **Depends On:** Create `internal/architect/adapters_test.go` File
  - **AC Ref:** Plan Section 3.1

- [x] **Task Title:** Implement Context/File Handling Adapter Tests
  - **Action:** Add tests for `ContextGathererAdapter` methods (`GatherContext`, `DisplayDryRunInfo`) and `FileWriterAdapter.SaveToFile`. Use mocks/fakes for the underlying interfaces, verify config/stats conversion, argument passing, and return values/errors.
  - **Depends On:** Create `internal/architect/adapters_test.go` File
  - **AC Ref:** Plan Section 3.1

## 4. Improve Config and File Utility Testing
- [x] **Task Title:** Extend `config_test.go` for `DefaultConfig`
  - **Action:** Add tests in `internal/config/config_test.go` to verify that `DefaultConfig()` returns a struct with the expected default values for all fields, including nested structures.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 4.1

- [x] **Task Title:** Implement `ValidateConfig` Tests
  - **Action:** Add tests in `internal/config/config_test.go` for `ValidateConfig()`. Provide various valid and invalid `CliConfig` inputs (e.g., missing required fields, invalid log levels) and verify it returns nil or an appropriate error.
  - **Depends On:** Extend `config_test.go` for `DefaultConfig`
  - **AC Ref:** Plan Section 4.1

- [x] **Task Title:** Add Edge Case/Error Handling Tests for Config
  - **Action:** Review `internal/config/config.go` for any potential edge cases or error conditions not covered by previous tests and add specific tests for them (e.g., handling of empty strings).
  - **Depends On:** Implement `ValidateConfig` Tests
  - **AC Ref:** Plan Section 4.1

- [x] **Task Title:** Enhance `fileutil` Tests for `shouldProcess`
  - **Action:** Add more comprehensive tests in `internal/fileutil/fileutil_test.go` for the `shouldProcess` function. Test various combinations of file paths, extensions, include patterns, exclude patterns, and exclude names to ensure correct filtering behavior.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 4.2

- [x] **Task Title:** Enhance `fileutil` Tests for `isGitIgnored`
  - **Action:** Add tests in `internal/fileutil/fileutil_test.go` for the `isGitIgnored` function. Set up mock `.gitignore` files with different patterns and test various file paths to ensure correct identification of ignored files.
  - **Depends On:** Enhance `fileutil` Tests for `shouldProcess`
  - **AC Ref:** Plan Section 4.2

- [x] **Task Title:** Add Error Handling Tests for File Operations
  - **Action:** Review `internal/fileutil` functions for file system interactions. Add tests that simulate errors like missing files, permission denied errors, or invalid paths, ensuring these errors are handled or propagated correctly.
  - **Depends On:** Enhance `fileutil` Tests for `isGitIgnored`
  - **AC Ref:** Plan Section 4.2

## 5. Complete Command Package Testing
- [x] **Task Title:** Improve Flag Parsing Tests
  - **Action:** Add or enhance tests in `cmd/architect` package for CLI flag parsing functions (`ParseFlags`, `ParseFlagsWithEnv`). Test various combinations of flags, environment variables, default values, and invalid inputs.
  - **Depends On:** Identify Specific Low-Coverage Functions
  - **AC Ref:** Plan Section 5.1

- [x] **Task Title:** Implement Logging Setup Tests
  - **Action:** Add tests in `cmd/architect` package for `SetupLogging` or equivalent logging initialization logic. Verify that the correct log level and output are configured based on flags/config.
  - **Depends On:** Improve Flag Parsing Tests
  - **AC Ref:** Plan Section 5.1

- [x] **Task Title:** Implement File Writing Tests
  - **Action:** Add tests in `cmd/architect` package for `SaveToFile` or equivalent output file writing logic. Test successful writing, handling of existing files, and error conditions (e.g., invalid path, permissions). Use temporary files/directories for testing.
  - **Depends On:** Implement Logging Setup Tests
  - **AC Ref:** Plan Section 5.1

- [x] **Task Title:** Implement Token Management Tests
  - **Action:** Add tests in `cmd/architect` package for token management CLI interactions, such as handling the `--confirm-tokens` flag, interacting with the `TokenManager`, and potentially simulating user confirmation prompts.
  - **Depends On:** Implement File Writing Tests
  - **AC Ref:** Plan Section 5.1

## 6. Update CI Configuration
- [x] **Task Title:** Update Coverage Threshold in CI Workflow
  - **Action:** Modify the `Check coverage threshold` step in `.github/workflows/ci.yml`. Change the `THRESHOLD` variable from `30` to `80`. Ensure the comparison logic remains correct.
  - **Depends On:** Run Full Test Suite and Verify Coverage Exceeds 80%
  - **AC Ref:** Plan Section 6.1

## 7. Verify Final Coverage
- [x] **Task Title:** Run Full Test Suite and Verify Coverage Exceeds 80%
  - **Action:** After implementing the above tests, run the full test suite with coverage analysis (`go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`). Check if the total coverage meets or exceeds 80%.
  - **Depends On:** All test implementation tasks from Sections 2-5
  - **AC Ref:** Plan Section 7

- [x] **Task Title:** Add Supplementary Tests if Needed
  - **Action:** If the coverage is below 80%, re-analyze the coverage report (`go tool cover -html=coverage.out`) to identify remaining significant gaps and add targeted tests to reach the threshold.
  - **Depends On:** Run Full Test Suite and Verify Coverage Exceeds 80%
  - **AC Ref:** Plan Section 7

- [x] **Task Title:** Commit Changes and Verify CI Pipeline Pass
  - **Action:** Commit all test additions, refactoring, and the CI configuration change. Push the changes and monitor the GitHub Actions CI pipeline to ensure all jobs (lint, test, build) pass, including the new 80% coverage check.
  - **Depends On:** Update Coverage Threshold in CI Workflow, Add Supplementary Tests if Needed (if required)
  - **AC Ref:** Plan Section 7

## 8. Test File Refactoring
- [x] **Task Title:** Refactor `adapters_test.go` for Improved Organization
  - **Action:** Break down the large `internal/architect/adapters_test.go` (1764 lines) into multiple smaller test files, each focusing on a specific adapter group or functionality. Create new files like `api_adapter_test.go`, `token_adapter_test.go`, `context_adapter_test.go`.
  - **Depends On:** None
  - **Expected Result:** Smaller, more maintainable test files with clear focus on specific adapter functionality.

- [x] **Task Title:** Refactor `orchestrator_test.go` for Improved Readability
  - **Action:** Break down the large `internal/architect/orchestrator/orchestrator_test.go` (1612 lines) into multiple smaller test files, each focusing on specific orchestrator methods or behaviors. Create new files like `orchestrator_init_test.go`, `orchestrator_process_test.go`, `orchestrator_error_test.go`.
  - **Depends On:** None
  - **Expected Result:** Improved test organization with better separation of concerns.

- [x] **Task Title:** Refactor `fileutil_test.go` for Better Test Structure
  - **Action:** Refactor `internal/fileutil/fileutil_test.go` (1603 lines) by grouping related tests into separate files based on functionality, such as `path_handling_test.go`, `filtering_test.go`, `git_operations_test.go`.
  - **Depends On:** None
  - **Expected Result:** Smaller, more focused test files with better organization.

- [ ] **Task Title:** Refactor `gemini_client_test.go` for Improved Maintainability
  - **Action:** Restructure `internal/gemini/gemini_client_test.go` (1574 lines) into multiple test files organized by API functionality, such as `client_generation_test.go`, `client_tokens_test.go`, `client_model_info_test.go`.
  - **Depends On:** None
  - **Expected Result:** Better test organization with clear separation of different client functionalities.

- [ ] **Task Title:** Refactor `integration_test.go` for Better Readability
  - **Action:** Refactor `internal/integration/integration_test.go` (1011 lines) by splitting tests into focused files based on specific integration scenarios or components being tested.
  - **Depends On:** None
  - **Expected Result:** Improved test structure with better separation of different integration test scenarios.

- [ ] **Task Title:** Refactor `cli_test.go` for Better Organization
  - **Action:** Restructure `cmd/architect/cli_test.go` (1007 lines) into multiple test files based on the CLI functionality being tested, such as `cli_args_test.go`, `cli_config_test.go`, `cli_execution_test.go`.
  - **Depends On:** None
  - **Expected Result:** Better organized CLI tests with clear focus on specific functionality.

- [ ] **Task Title:** Refactor `processor_test.go` for Improved Maintainability
  - **Action:** Reorganize `internal/architect/modelproc/processor_test.go` (959 lines) into multiple test files based on specific processor behaviors or methods being tested.
  - **Depends On:** None
  - **Expected Result:** Smaller, more maintainable test files with better organization.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** The plan assumes that the existing tests are generally valid and well-structured. Significant refactoring of existing tests might affect the estimated effort.
  - **Context:** General plan structure, Testing Principles section.
  
- [ ] **Issue/Assumption:** The refactoring of `geminiClient` involves injecting an HTTP client interface to enable mocking of HTTP requests/responses during testing rather than using the built-in `http.Client`.
  - **Context:** Plan Section 2.3, `internal/gemini/gemini_client.go` code showing internal `http.Client` creation.
  
- [ ] **Issue/Assumption:** Tests for adapters will use simple mock implementations of the underlying services rather than complex mocking frameworks, focusing on testing the adapter logic (delegation, data transformation).
  - **Context:** Plan Section 3, `internal/architect/adapters.go` structure indicating simple delegation pattern.
  
- [ ] **Issue/Assumption:** "Token management functions" in `cmd/architect` refers primarily to logic related to parsing/handling the `--confirm-tokens` flag and orchestrating calls to the `TokenManager`.
  - **Context:** Plan Section 5.1.
  
- [ ] **Clarification Needed:** The plan mentions `internal/architect/interfaces` has no tests, but tasks focus on testing `adapters.go`. Is testing the adapters sufficient to cover the interfaces they implement?
  - **Context:** Current State Analysis mentions `internal/architect/interfaces (no tests)`.