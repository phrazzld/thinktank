# TODO - OpenRouter Authentication Fix

## OpenRouter Integration Authentication Issue

- [~] **T100 · fix · p0: Fix API Key Isolation in Registry API Service**
    - **context:** Different provider API keys are being mixed up, causing authentication failures with OpenRouter
    - **action:**
        1. Update `internal/architect/registry_api.go` to properly isolate API keys per provider
        2. Add debug logging to show which API key is being used for each provider (first 5 chars only)
        3. Ensure the OpenRouter provider uses the correct OPENROUTER_API_KEY environment variable
    - **done-when:**
        1. Debug logs show correct API key for OpenRouter (starting with "sk-or")
        2. Authentication works correctly for OpenRouter models
    - **depends-on:** []
    - **status:** Initial implementation was incomplete. Additional tasks T105-T111 added to address the root cause.

- [ ] **T101 · fix · p1: Add API Key Validation in OpenRouter Provider**
    - **context:** The OpenRouter provider should validate the API key format before attempting to use it
    - **action:**
        1. Update `internal/providers/openrouter/provider.go` to check if the API key starts with "sk-or"
        2. Add warning log if key has incorrect format
        3. Improve error message if authentication fails
    - **done-when:**
        1. Client creation fails with clear error message if key format is incorrect
        2. Debug logs show API key validation is being performed
    - **depends-on:** [T100]

- [ ] **T102 · refactor · p1: Improve API Key Documentation and Error Messages**
    - **context:** Users need clear guidance on API key setup and better error messages
    - **action:**
        1. Add comments in code explaining API key retrieval logic
        2. Improve error messages for auth failures to suggest checking environment variables
        3. Update API key handling documentation
    - **done-when:**
        1. Code has clear comments about API key handling
        2. Error messages provide helpful troubleshooting guidance
    - **depends-on:** [T100]

- [ ] **T103 · test · p1: Create Multi-Provider Integration Test**
    - **context:** We need to verify API key isolation works with multiple providers
    - **action:**
        1. Create test that uses both Gemini and OpenRouter models in a single run
        2. Verify each provider uses the correct API key
        3. Add detailed logging of which keys are used
    - **done-when:**
        1. Test successfully authenticates with both providers
        2. Each provider uses its own correct API key
    - **depends-on:** [T100, T101]

- [ ] **T104 · test · p0: Validate OpenRouter Authentication End-to-End**
    - **context:** Final verification that the authentication issue is fixed
    - **action:**
        1. Run the architect tool with an OpenRouter model
        2. Verify authentication succeeds
        3. Check logs to confirm correct API key is used
    - **done-when:**
        1. The application successfully authenticates with OpenRouter
        2. API call succeeds with 200 OK response
        3. Generated output shows model response
    - **depends-on:** [T103]

## API Key Management Fix (OpenRouter Authentication Fix)

- [x] **T105 · fix · p0: Remove Hardcoded Gemini API Key Assignment**
    - **context:** In `cmd/architect/cli.go`, the `ParseFlagsWithEnv` function always sets `cfg.APIKey` from `GEMINI_API_KEY`
    - **action:**
        1. Remove or comment out the line that assigns `GEMINI_API_KEY` to `cfg.APIKey`
        2. Verify this doesn't break existing CLI validation logic
        3. Update any associated tests if needed
    - **done-when:**
        1. Code no longer automatically assigns `GEMINI_API_KEY` to `cfg.APIKey`
        2. Existing tests still pass
    - **depends-on:** []

- [x] **T106 · fix · p0: Pass Empty API Key to InitLLMClient**
    - **context:** The app passes `cliConfig.APIKey` to `InitLLMClient` which sets incorrect precedence
    - **action:**
        1. In `internal/architect/app.go`, find the call to `apiService.InitLLMClient`
        2. Change the first argument from `cliConfig.APIKey` to an empty string `""`
        3. Update any associated documentation
    - **done-when:**
        1. The application passes an empty string as the API key to force environment lookup
        2. All tests still pass
    - **depends-on:** [T105]

- [x] **T107 · fix · p0: Prioritize Environment Variables in InitLLMClient**
    - **context:** The `InitLLMClient` method prioritizes the passed API key over environment variables
    - **action:**
        1. Modify `InitLLMClient` in `registry_api.go` to always check the environment variable first
        2. Only fall back to the passed `apiKey` parameter if the environment variable is not set
        3. Add clear logging to show which API key source is being used
    - **done-when:**
        1. The method always checks environment variables first before using the provided apiKey
        2. Unit tests pass with the modified logic
    - **depends-on:** [T106]

- [x] **T108 · test · p1: Add Unit Tests for API Key Precedence Logic**
    - **context:** Need comprehensive test coverage for the key selection logic
    - **action:**
        1. Create test cases for when environment variables are set/unset
        2. Test with different providers (Gemini, OpenAI, OpenRouter)
        3. Verify correct precedence of env vars over passed parameters
    - **done-when:**
        1. Tests verify the environment variable is properly used when available
        2. Tests verify fallback behavior when environment variable is not set
        3. Tests cover multiple providers
    - **depends-on:** [T107]

- [x] **T109 · test · p1: Create Multi-Provider Test for API Key Isolation**
    - **context:** Need an integration test that verifies API keys aren't mixed between providers
    - **action:**
        1. Set up test environment with distinct API keys for different providers
        2. Create a test that uses models from multiple providers in one run
        3. Verify that each provider uses its correct API key
    - **done-when:**
        1. Test can successfully call multiple providers in one run
        2. Each provider uses the correct API key from its environment variable
    - **depends-on:** [T107]

- [ ] **T110 · test · p0: Manual End-to-End Testing**
    - **context:** Need to verify fixes work in a real environment
    - **action:**
        1. Build the application with all fixes
        2. Test with OpenRouter model using valid OPENROUTER_API_KEY
        3. Test with Gemini model using valid GEMINI_API_KEY
        4. Test with both models in sequence or parallel
    - **done-when:**
        1. OpenRouter authentication works correctly
        2. Gemini authentication works correctly
        3. Both can work in the same session without key leakage
    - **depends-on:** [T107]

- [ ] **T111 · task · p1: Update T100 Status and Documentation**
    - **context:** Original task T100 was marked complete but the fix was incomplete
    - **action:**
        1. Update T100 status to indicate partial implementation
        2. Add references to T105-T110 as the complete fix
        3. Update documentation to explain the root cause that was missed
    - **done-when:**
        1. T100 status accurately reflects partial implementation
        2. Documentation clearly explains the issue and complete fix
    - **depends-on:** [T105, T106, T107, T108, T109, T110]
