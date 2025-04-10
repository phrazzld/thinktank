# Implementation Plan: Add integration tests for prompt file usage

## Task
- **Action:** Create integration tests that validate the complete workflow with various prompt file formats.
- **Depends On:** Update generateAndSavePlanWithPromptManager function, Add fileIsTemplate detection mechanism
- **AC Ref:** Integration testing (Testing Strategy)

## Goal
Add comprehensive integration tests to verify that the application correctly handles different types of prompt file formats, specifically focusing on the recently implemented template detection features.

## Analysis of Current Implementation
- The application now uses `IsTemplate` to detect template variables in content
- `FileIsTemplate` function has been added to handle detection based on file extension and content
- Integration tests currently exist but don't specifically test the template detection mechanism

## Approach Selection

After analyzing the codebase, I've chosen to implement integration tests focused on verifying the proper handling of different prompt file formats using the existing integration testing framework.

### Implementation Approach

The integration tests will be implemented in the `internal/integration/integration_test.go` file, extending the existing test cases to cover various template scenarios. This approach leverages the existing test infrastructure while focusing on testing the new functionality.

Key aspects of the implementation:
1. Create test cases for different prompt file formats:
   - Regular content file (no template variables)
   - File with the `.tmpl` extension but no template variables
   - File with template variables but without the `.tmpl` extension
   - File with both template variables and the `.tmpl` extension
   - Invalid template syntax to test error handling

2. For each test case:
   - Create appropriate test files
   - Run the application with those files
   - Verify the output contains the expected content

3. Use the existing test infrastructure:
   - `TestEnv` for managing the test environment
   - `MainAdapter` for controlling application execution
   - Mock Gemini client for controlled responses

This approach is the most suitable because:
1. It builds upon the existing integration test framework
2. It isolates the template detection functionality for targeted testing
3. It covers the full range of possible template formats
4. It follows the project's established testing patterns
5. It maintains good testability by using mocks for external dependencies

## Implementation Details

### Files to Create/Modify:
1. `/Users/phaedrus/Development/architect-context-files/internal/integration/integration_test.go`
   - Add new test functions or extend existing ones

### High-Level Implementation Plan:
1. Create a new test function `TestPromptFileTemplateHandling` in `integration_test.go`
2. Implement sub-tests for each template scenario:
   - `t.Run("RegularTextFile", func(t *testing.T) {...})`
   - `t.Run("TemplateExtensionNoVariables", func(t *testing.T) {...})`
   - `t.Run("TemplateVariablesNoExtension", func(t *testing.T) {...})`
   - `t.Run("TemplateExtensionAndVariables", func(t *testing.T) {...})`
   - `t.Run("InvalidTemplateContent", func(t *testing.T) {...})`

3. For each sub-test:
   - Set up a clean test environment
   - Create the appropriate test files
   - Run the application with specific parameters
   - Verify the output contains expected content based on scenario

4. Additionally, update `generateAndSavePlan` in `main_adapter.go` to handle template detection for testing purposes