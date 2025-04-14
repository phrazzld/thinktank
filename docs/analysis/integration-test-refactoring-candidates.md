# Integration Test Refactoring Candidates

This document identifies integration tests that are overly broad and should be refactored to test smaller units directly.

## Identified Tests for Refactoring

### 1. TestFilteringBehaviors

**Current Implementation:**
- Tests file filtering functionality by running the full `architect.Execute` flow
- Creates test files and configures filters in a CliConfig
- Validates that the expected output file is created with correct content

**Refactoring Approach:**
- Should directly test `fileutil.GatherProjectContext` instead
- Create a standalone test that sets up test files with various extensions/names
- Configure a `fileutil.Config` with different include/exclude patterns
- Call `fileutil.GatherProjectContext` directly and verify the returned files match expectations
- No need to run content generation or check output files

**Benefits:**
- Much faster test execution (no API calls, no file generation)
- More focused on the actual filtering functionality
- Better isolation and easier debugging of filtering issues
- Tests the specific unit of code responsible for filtering

### 2. TestUserInteractions

**Current Implementation:**
- Tests user confirmation for token count thresholds using the full `architect.Execute` flow
- Sets up mock token count responses and simulates user input
- Verifies output file existence based on user confirmation

**Refactoring Approach:**
- Should directly test `TokenManager.PromptForConfirmation` instead
- Create a standalone test that sets up a `TokenManager` with different token counts
- Simulate user input responses (yes/no) directly
- Verify the return value of `PromptForConfirmation` without running the full application

**Benefits:**
- Better isolation of the user interaction component
- Avoids unnecessary setup of file contexts and API interactions
- Faster and more reliable tests
- More comprehensive test coverage of different interaction scenarios

### 3. TestModeVariations

**Current Implementation:**
- Tests different execution modes (like dry-run) using the full `architect.Execute` flow
- Sets up complete test environment with files, mocks, etc.
- Verifies behavior based on existence of output files

**Refactoring Approach:**
- Should test the dry-run mode directly at the `ContextGatherer.DisplayDryRunInfo` level
- Create a standalone test that sets up a `ContextGatherer` with test statistics
- Call `DisplayDryRunInfo` directly and capture/verify the output
- No need to run the full application flow

**Benefits:**
- More direct testing of the dry-run functionality
- Avoids overhead of setting up full application context
- Better isolation for debugging issues with dry-run mode

### 4. TestErrorScenarios

**Current Implementation:**
- Tests error handling in various scenarios using the full `architect.Execute` flow
- Sets up mock clients to return specific errors
- Verifies the errors are propagated correctly

**Refactoring Approach:**
- Split into multiple focused tests that target specific components:
  - Test token limit errors directly with `TokenManager.CheckTokenLimit`
  - Test API errors directly with API client error handling
  - Test file access errors with direct file utility calls
- Each test should focus on a specific error case at the component level

**Benefits:**
- More comprehensive testing of error scenarios
- Better isolation for debugging specific error handling
- Faster tests that don't require full application setup

## Implementation Strategy

For each refactoring:

1. Create a new test function focused on the specific component
2. Keep the original integration test to ensure overall system behavior
3. Gradually replace the original test once component tests are verified
4. Ensure the new tests properly mock only external dependencies, not internal components
5. Follow the project's testing philosophy of testing behavior, not implementation

This approach aligns with the project's testing strategy by:
- Testing at component boundaries
- Avoiding unnecessary mocking of internal collaborators
- Focusing on behavior, not implementation details
- Improving test performance and reliability