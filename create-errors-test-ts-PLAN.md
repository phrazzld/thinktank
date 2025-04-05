# Create errors.test.ts

## Goal
Create a comprehensive test suite for the new error classes and factory functions defined in src/core/errors.ts to ensure they work as expected. This will validate that the error hierarchy, error formatting, and specialized error creation logic function correctly.

## Implementation Approach
I'll create a new test file in src/core/__tests__/errors.test.ts with the following test structure:

1. **Error Classes Tests**:
   - Test ThinktankError (base class) construction with various options
   - Test format() method generates correct output with different properties
   - Test each specialized error class (ConfigError, ApiError, etc.)
   - Verify correct inheritance, property setting, and category assignment

2. **Factory Function Tests**:
   - Test createFileNotFoundError with different paths and options
   - Test createModelFormatError with various invalid model specifications
   - Test createMissingApiKeyError with single and multiple missing models
   - Test createModelNotFoundError with different scenarios (model not found, group context)

3. **Integration Tests**:
   - Test how errors work in combination (e.g., error chaining with cause)
   - Test formatting with complex examples

I'll use Jest and follow the project's existing test patterns with focused unit tests that test behavior rather than implementation details.

## Reasoning
The chosen approach provides comprehensive coverage of the error system while maintaining alignment with the project's testing style. 

Benefits of this approach:
- **Completeness**: Will test all error classes and factory functions
- **Behavior-focused**: Tests the actual behavior (error construction, formatting, etc.) rather than implementation details
- **Isolation**: Each test focuses on a specific aspect of the error system for clear failures/debugging
- **Readability**: Organizing tests by error class and factory function makes the test file easy to navigate
- **Maintainability**: Using Jest's describe/it pattern allows easy extension when new error types are added

The test file will provide confidence that the error system is working correctly and serves as documentation for how the different error types should be used.