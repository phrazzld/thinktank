# End-to-End Tests for thinktank CLI

This directory contains end-to-end tests for the thinktank CLI. These tests verify the functionality of the compiled binary by running it with various arguments and verifying the output.

## Test Structure

The E2E tests in this directory:

1. Build the thinktank binary (if it's not already built)
2. Run the binary with various arguments
3. Verify the output against expected values
4. Verify that output files are created as expected

## Running the Tests

These tests use Go's build tags to control when they run:

- By default, these tests are **skipped** in normal test runs
- To run these tests, use the `manual_api_test` build tag

```bash
# Run just the E2E tests
go test -tags=manual_api_test ./internal/e2e/...

# Run with verbose output
go test -v -tags=manual_api_test ./internal/e2e/...

# Run a specific test
go test -v -tags=manual_api_test ./internal/e2e -run=TestBasicExecution
```

You can also use the provided `run_e2e_tests.sh` script:

```bash
./internal/e2e/run_e2e_tests.sh
```

## Mock Server

The tests use a mock HTTP server to simulate the Gemini API. This allows the tests to run without requiring a real API key or network connection.

The mock server:

1. Simulates the Gemini API endpoints
2. Returns predefined responses for token counting and content generation
3. Can be configured to return errors to test error handling

## Test Environment

Each test creates a `TestEnv` which provides:

1. A temporary directory for test files
2. Helper methods to create test files and directories
3. Methods to run the thinktank binary with various arguments
4. The mock HTTP server

## Assertion Framework

The tests use a flexible assertion framework that:

1. Properly validates outputs while being aware of mock API limitations
2. Clearly distinguishes between required and optional expectations
3. Provides detailed logging for test failures
4. Includes specialized assertion helpers for API-dependent operations

### Assertion Helpers

The following assertion helpers are available:

- `AssertCommandSuccess`: Verifies that a command succeeded with exit code 0 and the expected output
- `AssertCommandFailure`: Verifies that a command failed with a specific exit code and error message
- `AssertAPICommandSuccess`: Like AssertCommandSuccess, but relaxes requirements when API issues are detected
- `AssertFileContent`: Verifies that a file exists and contains expected content
- `AssertFileMayExist`: Checks if a file exists and optionally validates its content (useful for API-dependent tests)

### Required vs. Optional Expectations

The assertion framework distinguishes between:

- **Required expectations**: If not met, the test fails
- **Optional expectations**: If not met, the test logs a message but doesn't fail

This approach allows E2E tests to validate critical behavior while being resilient to the limitations of mock testing.

## Key Files

- `e2e_test.go`: Contains the core test infrastructure including TestMain and the mock server
- `helpers.go`: Contains helper functions for test verification and setup
- `cli_*.go`: Individual test files for different features
- `run_e2e_tests.sh`: Script to easily run the E2E tests

## Design Principles

1. **Self-Validation**: Tests automatically validate their outcomes, not requiring manual inspection
2. **Isolation**: Each test runs in its own environment with its own temporary directory
3. **Clarity**: Tests clearly report what they're testing and why they pass or fail
4. **Resilience**: Tests handle the limitations of mock testing while still providing useful validation

## Mock Server Limitations

Some important limitations to be aware of:

1. The Google Gemini client requires valid authentication even when using a custom endpoint
2. The client expects a specific response format that is challenging to fully mock
3. API key validation happens at the client level, not just at the server level

Our assertion framework is designed to handle these limitations while still providing meaningful test validation.
