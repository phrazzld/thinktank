# Test Refactoring Roadmap

This document provides a roadmap for the "Update integration tests if needed" task, based on the findings from running component tests after extraction.

## Test Issues Categories

Tests were categorized by failure reason:

1. **Implementation Coupling (35%)**: Tests tightly coupled to implementation details
2. **Test Setup/Environment (35%)**: Issues with test setup, directories, etc.
3. **Behavioral Changes (15%)**: Behavior that changed during refactoring
4. **Dependency Issues (15%)**: Problems with package references after refactoring

## Priority Issues to Address

### 1. Implementation Coupling Tests

These tests violate our "Behavior Over Implementation" principle from TESTING_STRATEGY.md:

- **Prompt Component**:
  - `TestBuildPromptWithConfig/SuccessWithCustomTemplate`: Relies on exact template content
  - `TestBuildPromptWithConfig/ErrorFromSetupPromptManager`: Uses specific error message format

**Recommended Action**: Refactor these tests to focus on meaningful behavior verification rather than exact string matching. Use pattern matching or semantic validation instead of exact string comparisons.

### 2. Behavioral Changes 

These tests indicate potential changes in component behavior:

- **Context Component**:
  - `TestGatherContext/FileAccessError`: Behavior changed to return empty result for non-existent paths rather than an error

**Recommended Action**: Verify if the new behavior is intentional and correct, then update tests to match intended behavior.

### 3. Test Setup Issues

These tests have issues with their setup/environment:

- **Output Component**:
  - `TestGenerateAndSavePlan/Happy_path_-_generates_content_successfully`: Directory creation issue
  - `TestListExampleTemplates`: Nil pointer dereference due to not restoring original function

**Recommended Action**: Create shared test utilities for directory creation/cleanup and improve setup/teardown patterns.

### 4. Dependency Issues

These tests have issues with imports or package references:

- **Output Component**:
  - `TestGenerateAndSavePlanWithConfig`: Skipped due to package reference issues

**Recommended Action**: Fix import paths and update references to work with the new package structure.

## Integration Test Improvements

After addressing existing test issues, consider these improvements:

1. **Add Component Integration Tests**: 
   - Create tests that verify components work together correctly
   - Focus on workflow tests that exercise public APIs

2. **Review Mocking Strategy**:
   - Ensure tests only mock true external boundaries
   - Replace internal mocks with real implementations where possible

3. **Standardize Test Patterns**:
   - Create consistent test setup/teardown helpers
   - Standardize directory handling for file-based tests

## Approach for Implementation

1. Start with fixing structural issues (test setup, nil pointers)
2. Verify and update tests for behavioral changes
3. Refactor implementation-coupled tests
4. Add missing integration tests
5. Create shared test utilities to standardize future tests

This approach will result in a more robust test suite that aligns with our testing principles and is less prone to breaking with future refactoring.