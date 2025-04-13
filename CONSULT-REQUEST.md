# Consultation Request: End-to-End CLI Tests

## Goal
Implement end-to-end (CLI) tests that execute the compiled architect binary with various command-line arguments and fixtures, verify exit codes, output files, audit logs, and behavior with key flags as described in the "Implement End-to-End (CLI) Tests" task in TODO.md.

## Problem/Blocker
I've implemented the infrastructure for end-to-end CLI tests in a new internal/e2e package, including a test environment with mock HTTP server for the Gemini API. However, I'm encountering a blocker where the architect binary doesn't seem to connect to the mock API server properly. The tests fail because the architect binary is still trying to reach the real Gemini API instead of using our mock.

This creates testing issues that conflict with our TESTING_STRATEGY.md principles:
- We need to mock the external Gemini API (a true external dependency) for reproducible tests
- Tests are currently failing and unreliable (violating "Repeatable / Reliable" principle)
- The mock setup may need architectural modifications to the main code to support testing

## Context/History
I've created a comprehensive test infrastructure including:
1. A TestEnv structure for isolated test environments
2. Mock HTTP server for the Gemini API
3. Test cases for basic CLI functionality, file filtering, multiple models, etc.
4. Functions for executing the binary with different arguments

The implementation plan is documented in implement-end-to-end-cli-tests-PLAN.md, and I've marked the task as in progress in TODO.md with a status note about the current blocker.

I tried setting environment variables (GEMINI_API_KEY and GEMINI_API_URL) when executing the binary, but it appears that the architect binary is not respecting the API URL environment variable to use the mock server instead of the real API.

## Key Files/Code Sections
- `/Users/phaedrus/Development/architect/internal/e2e/e2e_test.go`: Main test infrastructure
- `/Users/phaedrus/Development/architect/internal/e2e/cli_basic_test.go`: Basic CLI tests
- `/Users/phaedrus/Development/architect/internal/gemini/client.go`: Gemini API client that needs to be mocked
- `/Users/phaedrus/Development/architect/cmd/architect/main.go`: CLI entry point
- `/Users/phaedrus/Development/architect/internal/architect/api.go`: API service implementation

## Error Messages
When running the tests, the architect binary fails with errors indicating it's still trying to use the real Gemini API:
```
Failed to count tokens in prompt: googleapi: Error 400: API key not valid. Please pass a valid API key.
error details: name = ErrorInfo reason = API_KEY_INVALID domain = googleapis.com metadata = map[service:generativelanguage.googleapis.com]
error details: name = LocalizedMessage locale = en-US msg = API key not valid. Please pass a valid API key.
```

This occurs despite setting the GEMINI_API_URL environment variable to point to our mock server.

## Desired Outcome
1. A way to modify or configure the architect binary to use the mock API server in test mode
2. Understanding of how the Gemini API client should be configured to respect the mock server URL
3. Potentially, an approach to make the code more testable without excessive modifications to the production code
4. Clean, passing end-to-end tests that reliably verify the CLI behavior without depending on the actual Gemini API