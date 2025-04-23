# TODO: Restore OpenAI Functionality

## Critical Regression
The OpenAI provider is currently completely mocked and returns dummy responses. This is a critical regression that needs immediate attention. The provider previously worked correctly but was disabled during refactoring when token functionality was removed.

## Tasks

- [x] **T1: Restore OpenAI Go Client Dependency**
  - Add `github.com/openai/openai-go v0.1.0-beta.10` to go.mod
  - Run `go mod tidy` to update dependencies
  - Verify dependency is correctly installed

- [ ] **T2: Restore OpenAI Client Implementation** depends-on: T1
  - Update `/internal/openai/openai_client.go` to use the actual OpenAI API client
  - Replace the mock implementation with the original implementation from commit `1db5091`
  - Restore the following key components:
    - `openaiAPI` interface and `realOpenAIAPI` implementation
    - ChatCompletion structs and methods
    - Proper API call implementation in `GenerateContent`
    - Tag extraction helper functions

- [ ] **T3: Update OpenAI Client Constructor** depends-on: T2
  - Modify `NewClient` function to accept API key again
  - Restore OpenAI client initialization code
  - Update error handling for API initialization failures

- [ ] **T4: Update Provider Integration** depends-on: T3
  - Ensure `/internal/providers/openai/provider.go` correctly calls the updated client
  - Verify the `OpenAIClientAdapter` works with the restored client
  - Make sure parameter handling is working correctly

- [ ] **T5: Add Tests** depends-on: T4
  - Restore skipped tests in `openai_content_test.go`
  - Ensure all test mocks are updated to match the new implementation
  - Add specific tests for streaming functionality if implemented

- [ ] **T6: Test End-to-End** depends-on: T5
  - Manually verify that OpenAI models work correctly
  - Check that responses are complete (not truncated)
  - Verify all parameters are correctly passed to the API

## Implementation Notes

- The original implementation used `github.com/openai/openai-go` client library
- Make sure not to disrupt other providers during restoration
- Consult the git history for the complete original implementation
- Key commits to reference:
  - `1db5091` - Before token functionality was removed
  - `7414e1a` - Update to LLM interface that removed token-related methods

## Priority: CRITICAL

This functionality must be restored as soon as possible, as it renders OpenAI models unusable in the current state.
