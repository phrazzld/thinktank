# TODO - OpenRouter Authentication Fix

## OpenRouter Integration Authentication Issue

- [x] **T100 · fix · p0: Fix API Key Isolation in Registry API Service**
    - **context:** Different provider API keys are being mixed up, causing authentication failures with OpenRouter
    - **action:**
        1. Update `internal/architect/registry_api.go` to properly isolate API keys per provider
        2. Add debug logging to show which API key is being used for each provider (first 5 chars only)
        3. Ensure the OpenRouter provider uses the correct OPENROUTER_API_KEY environment variable
    - **done-when:**
        1. Debug logs show correct API key for OpenRouter (starting with "sk-or")
        2. Authentication works correctly for OpenRouter models
    - **depends-on:** []

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
