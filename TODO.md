# TODO

## Phase 0: Code Refactoring (P0 - Immediate)

- [x] **R001:** Create openai_test_utils.go Utility File
    - **Action:** Create new file `internal/openai/openai_test_utils.go` and move all shared mock implementations, helper functions, and test utilities from the large test file.
    - **Depends On:** None
    - **AC Ref:** Code Cleanup

- [x] **R002:** Create and Populate openai_interface_test.go
    - **Action:** Create new file `internal/openai/openai_interface_test.go` and move the TestOpenAIClientImplementsLLMClient test from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [x] **R003:** Create and Populate openai_client_creation_test.go
    - **Action:** Create new file `internal/openai/openai_client_creation_test.go` and move all client creation tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R004:** Create and Populate openai_parameters_test.go
    - **Action:** Create new file `internal/openai/openai_parameters_test.go` and move all parameter handling tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R005:** Create and Populate openai_content_test.go
    - **Action:** Create new file `internal/openai/openai_content_test.go` and move all content generation tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R006:** Create and Populate openai_errors_test.go
    - **Action:** Create new file `internal/openai/openai_errors_test.go` and move all error handling tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R007:** Create and Populate openai_tokens_test.go
    - **Action:** Create new file `internal/openai/openai_tokens_test.go` and move all token counting tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R008:** Create and Populate openai_model_info_test.go
    - **Action:** Create new file `internal/openai/openai_model_info_test.go` and move all model info tests from the original test file.
    - **Depends On:** [R001]
    - **AC Ref:** Code Cleanup

- [ ] **R009:** Verify All Tests Pass After Refactoring
    - **Action:** Run `go test ./internal/openai/...` to verify all tests still pass after the refactoring.
    - **Depends On:** [R001, R002, R003, R004, R005, R006, R007, R008]
    - **AC Ref:** Code Cleanup

- [ ] **R010:** Remove Original Large Test File
    - **Action:** Once all tests have been successfully moved and are passing, delete the original `internal/openai/openai_client_test.go` file.
    - **Depends On:** [R009]
    - **AC Ref:** Code Cleanup

## Phase 1: Core Functionality and API Key Handling (P0 - Immediate)

- [x] **T001:** Implement Test for Empty API Key Handling (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` to verify that client creation or relevant functions fail appropriately when an empty API key is provided or configured. Ensure the error returned is specific and informative.
    - **Depends On:** None
    - **AC Ref:** Plan Item 1

- [x] **T002:** Implement Test for Valid API Key Format Detection (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` to verify that the client correctly identifies and accepts API keys matching the expected format (e.g., starts with "sk-"). Use mocked validation logic.
    - **Depends On:** None
    - **AC Ref:** Plan Item 1

- [x] **T003:** Implement Test for Invalid API Key Format Handling (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` to verify that the client rejects API keys that do not match the expected format and returns an appropriate error.
    - **Depends On:** None
    - **AC Ref:** Plan Item 1

- [x] **T004:** Implement Test for API Key Environment Variable Fallback (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` to verify that the client correctly falls back to using the `OPENAI_API_KEY` environment variable when no key is explicitly provided. Use environment variable mocking techniques.
    - **Depends On:** None
    - **AC Ref:** Plan Item 1

- [x] **T005:** Implement Test for Mocked API Key Permission/Validation Logic (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` simulating scenarios where an API key is syntactically valid but lacks permissions or fails validation with the (mocked) API. Ensure appropriate errors are handled.
    - **Depends On:** None
    - **AC Ref:** Plan Item 1

- [x] **T006:** Implement Test for Client Creation with Default Configuration (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` to verify successful creation of an OpenAI client using default settings (e.g., default model, default parameters).
    - **Depends On:** [T001, T002, T003, T004, T005]
    - **AC Ref:** Plan Item 2

- [x] **T007:** Implement Test for Client Creation with Custom Configuration (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` to verify successful creation of an OpenAI client with various custom configurations (e.g., different model specified, custom base URL if applicable).
    - **Depends On:** [T001, T002, T003, T004, T005]
    - **AC Ref:** Plan Item 2

- [x] **T008:** Implement Test for `GenerateContent` with Valid Parameters (openai_client_test.go)
    - **Action:** Write a unit test in `internal/openai/openai_client_test.go` that calls `GenerateContent` with valid inputs and parameters, mocking the API response to return a successful completion. Verify the expected content is returned.
    - **Depends On:** [T006, T007]
    - **AC Ref:** Plan Item 2

- [x] **T009:** Implement Test for Parameter Conversion and Validation in Client (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` to verify that input parameters (e.g., temperature, max_tokens) are correctly converted to the format expected by the OpenAI API client library and that invalid parameter values are handled correctly before making an API call.
    - **Depends On:** [T006, T007]
    - **AC Ref:** Plan Item 2

- [x] **T010:** Implement Mock for OpenAI API Error Responses (openai_client_test.go)
    - **Action:** Create mock implementations or use a mocking library to simulate various OpenAI API error responses (e.g., 401 Unauthorized, 429 Rate Limit Exceeded, 500 Server Error) for use in error handling tests.
    - **Depends On:** None
    - **AC Ref:** Plan Item 2

- [x] **T011:** Implement Test for Client Error Handling (API Errors) (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` using the error mocks (T010) to verify that the client correctly handles different API error scenarios returned by `GenerateContent` and other methods, propagating or converting errors appropriately.
    - **Depends On:** [T006, T007, T010]
    - **AC Ref:** Plan Item 2

- [x] **T012:** Implement Mock for Token Counting Mechanism/API (openai_client_test.go)
    - **Action:** Create a mock implementation for the token counting mechanism (e.g., `tiktoken`) used by the OpenAI client, allowing tests to control the returned token count for specific inputs.
    - **Depends On:** None
    - **AC Ref:** Plan Item 2

- [x] **T013:** Implement Test for Token Counting Accuracy in Client (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` using the token counting mock (T012) to verify that the client's `CountTokens` method (or equivalent) returns the expected token count for various inputs.
    - **Depends On:** [T006, T007, T012]
    - **AC Ref:** Plan Item 2

- [x] **T014:** Implement Mock for Model Information Retrieval API (openai_client_test.go)
    - **Action:** Create mock implementations or use a mocking library to simulate the OpenAI API endpoint for retrieving model information (e.g., token limits).
    - **Depends On:** None
    - **AC Ref:** Plan Item 2

- [ ] **T015:** Implement Test for Model Information Retrieval in Client (openai_client_test.go)
    - **Action:** Write unit tests in `internal/openai/openai_client_test.go` using the model info mock (T014) to verify that the client's `GetModelInfo` method (or equivalent) correctly retrieves and parses model details like token limits.
    - **Depends On:** [T006, T007, T014]
    - **AC Ref:** Plan Item 2

- [ ] **T016:** Implement Test for OpenAI Provider's `Provider` Interface Compliance (provider_test.go)
    - **Action:** Write compile-time or runtime tests in `internal/providers/openai/provider_test.go` to ensure the `OpenAIProvider` struct correctly implements all methods required by the `providers.Provider` interface.
    - **Depends On:** None
    - **AC Ref:** Plan Item 3

- [ ] **T017:** Implement Test for Provider's `CreateClient` Method (Successful Case) (provider_test.go)
    - **Action:** Write a unit test in `internal/providers/openai/provider_test.go` for the `CreateClient` method, mocking dependencies as needed, to verify that it successfully returns an initialized OpenAI client when provided with valid inputs (API key, model ID).
    - **Depends On:** None
    - **AC Ref:** Plan Item 3

- [ ] **T018:** Implement Test for Provider's `CreateClient` Method (API Key Error Cases) (provider_test.go)
    - **Action:** Write unit tests in `internal/providers/openai/provider_test.go` for the `CreateClient` method, verifying that it returns appropriate errors when provided with invalid or missing API keys, leveraging scenarios tested in T001, T003.
    - **Depends On:** [T001, T003]
    - **AC Ref:** Plan Item 3

- [ ] **T019:** Implement Test for Provider's Environment Variable Handling for API Keys (provider_test.go)
    - **Action:** Write unit tests in `internal/providers/openai/provider_test.go` for the `CreateClient` method, verifying that it correctly reads the API key from the `OPENAI_API_KEY` environment variable when no key is explicitly passed. Use environment variable mocking.
    - **Depends On:** [T004]
    - **AC Ref:** Plan Item 3

- [ ] **T020:** Implement Test for Provider Initialization with Different Loggers (provider_test.go)
    - **Action:** Write unit tests in `internal/providers/openai/provider_test.go` to verify that the `NewProvider` function correctly accepts and uses different `logutil.LoggerInterface` implementations (including nil, which should default).
    - **Depends On:** None
    - **AC Ref:** Plan Item 3

## Phase 2: Parameter Handling and Integration (P1 - Higher Priority)

- [ ] **T021:** Design Table-Driven Test Structure for Parameter Combinations (openai_client_test.go)
    - **Action:** Refactor or design the test structure in `internal/openai/openai_client_test.go` to use table-driven tests for efficiently testing various parameter combinations passed to `GenerateContent` or client creation.
    - **Depends On:** [T006, T007]
    - **AC Ref:** Plan Item 4

- [ ] **T022:** Implement Test Cases for All Valid Parameter Combinations (openai_client_test.go)
    - **Action:** Add test cases to the table-driven structure (T021) covering all valid combinations of supported OpenAI parameters (temperature, top_p, max_tokens, etc.). Verify parameters are correctly passed to the mocked API call.
    - **Depends On:** [T021]
    - **AC Ref:** Plan Item 4

- [ ] **T023:** Implement Test Cases for Parameter Type Conversions (openai_client_test.go)
    - **Action:** Add test cases to the table-driven structure (T021) verifying that parameters passed as different numeric types (e.g., int, float32, float64) are correctly converted to the types expected by the OpenAI client library.
    - **Depends On:** [T021]
    - **AC Ref:** Plan Item 4

- [ ] **T024:** Implement Test Cases for Parameter Validation Logic (Invalid Values) (openai_client_test.go)
    - **Action:** Add test cases to the table-driven structure (T021) providing invalid parameter values (e.g., temperature out of range, negative max_tokens) and verify that appropriate errors are returned *before* an API call is made.
    - **Depends On:** [T021]
    - **AC Ref:** Plan Item 4

- [ ] **T025:** Implement Test for `createChatCompletionWithParams` Method (openai_client_test.go)
    - **Action:** Write specific unit tests in `internal/openai/openai_client_test.go` targeting the `createChatCompletionWithParams` method (or equivalent parameter-handling logic) to ensure it correctly builds the API request payload based on provided parameters.
    - **Depends On:** [T006, T007]
    - **AC Ref:** Plan Item 4

- [ ] **T026:** Implement Test Cases for Parameter Inheritance/Overriding (if applicable) (openai_client_test.go)
    - **Action:** If the client/provider supports parameter inheritance (e.g., default parameters overridden by call-specific parameters), add test cases to the table-driven structure (T021) to verify this behavior.
    - **Depends On:** [T021]
    - **AC Ref:** Plan Item 4

- [ ] **T027:** Implement Test for `OpenAIClientAdapter` Construction and Basic Methods (provider_test.go)
    - **Action:** Write unit tests in `internal/providers/openai/provider_test.go` to verify the `NewOpenAIClientAdapter` function and ensure basic methods of the adapter correctly delegate to the wrapped client.
    - **Depends On:** None
    - **AC Ref:** Plan Item 5

- [ ] **T028:** Implement Test for Parameter Passing through `OpenAIClientAdapter` (provider_test.go)
    - **Action:** Write unit tests in `internal/providers/openai/provider_test.go` verifying that parameters set via `SetParameters` on the adapter are correctly passed through to the underlying client's `GenerateContent` method.
    - **Depends On:** [T027]
    - **AC Ref:** Plan Item 5

- [ ] **T029:** Implement Test for Provider-Specific Parameter Handling in Adapter (if any) (provider_test.go)
    - **Action:** If the `OpenAIClientAdapter` performs any provider-specific parameter mapping or validation beyond simple pass-through, write unit tests in `internal/providers/openai/provider_test.go` to cover this logic.
    - **Depends On:** [T027]
    - **AC Ref:** Plan Item 5

## Phase 3: Registry Integration (P2 - Medium Priority)

- [ ] **T030:** Implement Test for Loading OpenAI Provider/Model Config from Registry (registry_test.go)
    - **Action:** Write unit tests in `internal/registry/registry_test.go` to verify that the registry correctly loads and parses configuration details for the OpenAI provider and its associated models (e.g., gpt-4, gpt-3.5-turbo) from a sample `models.yaml`.
    - **Depends On:** None
    - **AC Ref:** Plan Item 6

- [ ] **T031:** Implement Test for Retrieving OpenAI Model Info via Registry (registry_test.go)
    - **Action:** Write unit tests in `internal/registry/registry_test.go` to verify that `registry.GetModel` correctly returns the `ModelDefinition` (including token limits, parameters) for registered OpenAI models.
    - **Depends On:** [T030]
    - **AC Ref:** Plan Item 6

- [ ] **T032:** Implement Test for Selecting OpenAI Provider Based on Model Name via Registry (registry_test.go)
    - **Action:** Write unit tests in `internal/registry/registry_test.go` verifying that the registry can correctly identify "openai" as the provider for models like "gpt-4-turbo" based on the loaded configuration.
    - **Depends On:** [T030]
    - **AC Ref:** Plan Item 6

- [ ] **T033:** Design Integration Test Setup for OpenAI Provider (internal/integration/)
    - **Action:** Define and set up the necessary mocks (e.g., mock HTTP server simulating OpenAI API) and test environment configurations required for OpenAI provider integration tests within the `internal/integration/` directory.
    - **Depends On:** None
    - **AC Ref:** Plan Item 7

- [ ] **T034:** Implement Integration Test for End-to-End OpenAI Provider Selection and Client Instantiation (internal/integration/)
    - **Action:** Write an integration test using the setup from T033 that simulates the application flow: selecting the OpenAI provider based on a model name (e.g., "gpt-4"), creating the client via the provider, and making a simple mocked API call.
    - **Depends On:** [T017, T032, T033]
    - **AC Ref:** Plan Item 7

- [ ] **T035:** Implement Integration Test for Multi-Provider Scenario Including OpenAI (internal/integration/)
    - **Action:** Write an integration test using the setup from T033 that involves both the OpenAI provider and another provider (e.g., Gemini), verifying that the application can correctly instantiate and use clients from multiple providers based on model names.
    - **Depends On:** [T033]
    - **AC Ref:** Plan Item 7

- [ ] **T036:** Implement Integration Test for OpenAI Error Handling and Fallbacks (internal/integration/)
    - **Action:** Write integration tests using the setup from T033 that simulate API errors from the mocked OpenAI API and verify that the application handles these errors gracefully (e.g., logs errors, potentially falls back if applicable).
    - **Depends On:** [T011, T033]
    - **AC Ref:** Plan Item 7

## Metrics and Verification

- [ ] **T037:** Configure CI to Run Coverage Check for OpenAI Packages (`internal/openai`, `internal/providers/openai`)
    - **Action:** Update the GitHub Actions workflow (`/.github/workflows/test.yml`) to specifically calculate and report test coverage for the `internal/openai` and `internal/providers/openai` packages. Ensure the job fails if coverage drops below 70% for these packages.
    - **Depends On:** None
    - **AC Ref:** Metrics and Verification

- [ ] **T038:** Run Coverage Check Locally After Each Phase Completion and Update PR/Status
    - **Action:** After completing the tasks for each phase (Phase 1, Phase 2, Phase 3), run `go test -short -coverprofile=coverage.out ./...` locally. Analyze the coverage report (`go tool cover -html=coverage.out`) for `internal/openai` and `internal/providers/openai`. Update the pull request (#14) with the current coverage numbers.
    - **Depends On:** [T020, T029, T036]
    - **AC Ref:** Metrics and Verification

- [ ] **T039:** Run `go vet` and Linters on New Test Code and Fix Issues
    - **Action:** Regularly run `go vet ./...` and the configured linter (e.g., `golangci-lint run`) on the newly added test code. Address any reported issues to maintain code quality.
    - **Depends On:** [T001] # Start vetting early
    - **AC Ref:** Metrics and Verification

- [ ] **T040:** Ensure Final CI Run Passes with >70% Coverage for Relevant Packages
    - **Action:** After all test implementation tasks (T001-T036) are complete, trigger a final CI run. Verify that the coverage checks for `internal/openai` and `internal/providers/openai` pass the >70% threshold and that all other CI checks succeed.
    - **Depends On:** [T036, T037, T038, T039]
    - **AC Ref:** Metrics and Verification
