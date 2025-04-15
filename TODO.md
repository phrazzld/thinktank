# TODO

## Code Review Implementation Tasks

- [x] **Task Title:** Refactor config validation to use injectable getenv
  - **Action:** Modify `internal/config/config.go:ValidateConfig` to accept a `getenv func(string) string` parameter instead of calling `os.Getenv` directly. Update call sites and tests accordingly. This improves testability by allowing environment variables to be mocked during tests.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: Direct `os.Getenv` use in config validation)

- [ ] **Task Title:** Evaluate necessity of deprecated API methods
  - **Action:** Analyze the codebase (`internal/architect/api.go`) to determine if the deprecated `InitClient` and `ProcessResponse` methods, along with the `llmToGeminiClientAdapter`, are still required internally or by external consumers. Document findings on whether they can be safely removed.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: Complexity of adapter for deprecated methods)

- [ ] **Task Title:** Remove deprecated API methods if unnecessary
  - **Action:** Based on the evaluation, if the deprecated methods (`InitClient`, `ProcessResponse`) and the adapter (`llmToGeminiClientAdapter`) in `internal/architect/api.go` are no longer needed, remove them from the codebase. Update all call sites to use the new provider-agnostic methods.
  - **Depends On:** Evaluate necessity of deprecated API methods
  - **AC Ref:** N/A (Refers to Issue: Complexity of adapter for deprecated methods)

- [ ] **Task Title:** Add comments explaining error helper string matching
  - **Action:** Add explicit comments to the error helper methods (`IsEmptyResponseError`, `IsSafetyBlockedError`) in `internal/architect/api.go` explaining why string matching is used for provider-agnostic checks and the potential risks if error messages change.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: String matching in error helper methods)

- [ ] **Task Title:** Enhance tests for error helper string matching variations
  - **Action:** Update the tests for `IsEmptyResponseError` and `IsSafetyBlockedError` in `internal/architect/api_test.go` to cover potential variations in error message strings from different providers to improve robustness against future changes.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: String matching in error helper methods)

- [ ] **Task Title:** Consider adding error categorization to LLMClient interface
  - **Action:** Evaluate the feasibility and benefit of adding methods to the `llm.LLMClient` interface to expose categorized errors (e.g., rate limit, auth, server error) if underlying provider clients support it. This could improve error logging specificity in `internal/architect/modelproc/processor.go`.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: Less specific error logging in modelproc)

- [ ] **Task Title:** Update app.Execute to use InitLLMClient directly
  - **Action:** Modify `internal/architect/app.go:Execute` to initialize the reference client using `apiService.InitLLMClient` directly, removing the use of the deprecated `apiService.InitClient` and the subsequent `gemini.AsLLMClient` adaptation.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: Deprecated client initialization)

- [ ] **Task Title:** Remove duplicate TestProviderDetection implementation
  - **Action:** Identify the duplicate `TestProviderDetection` function definitions in `internal/architect/api_test.go` and `internal/architect/api_provider_test.go`. Consolidate the test logic into one file (preferably `api_provider_test.go`) and remove the duplicate implementation.
  - **Depends On:** None
  - **AC Ref:** N/A (Refers to Issue: Duplicated test function)

- [ ] **Task Title:** Reinstate minimal test coverage for deprecated InitClient method
  - **Action:** Add a minimal test case back to the appropriate test file to cover the basic functionality of the deprecated `InitClient` method, ensuring it doesn't break unexpectedly while it still exists. This test can be removed if/when the deprecated method is removed.
  - **Depends On:** Evaluate necessity of deprecated API methods
  - **AC Ref:** N/A (Refers to Issue: Removed test for deprecated method)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** The tasks above are based on the code review and don't have formal Acceptance Criteria references since they are remediation tasks identified from a PR review.
  - **Context:** The tasks address specific issues identified in the code review of the OpenAI provider support PR, focusing on improving code quality and following project standards.
