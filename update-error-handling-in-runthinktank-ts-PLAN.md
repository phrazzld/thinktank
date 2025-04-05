# Update error handling in runThinktank.ts

## Goal
Update the try/catch blocks in runThinktank.ts to use the new error system more effectively.

## Implementation Approach
After analyzing the current implementation in runThinktank.ts, I can see that basic integration with the error system already exists. However, there are opportunities to further enhance error handling by:

1. Using specialized error factory functions for specific error types
2. Ensuring proper error categorization throughout the workflow
3. Improving error cause chaining to preserve context
4. Enhancing error suggestions for better troubleshooting

### Current State
- The file already imports ThinktankError, ConfigError, and errorCategories
- There's a main try/catch block that wraps the entire workflow
- Error conversion happens in lines 518-537:
  - ThinktankError instances are preserved
  - Standard Error objects are converted to ThinktankError
  - Non-Error objects get a generic unknown error message
- There's specific handling for ModelSelectionError (lines 363-378)
- Error categorization exists for display purposes

### Planned Changes
1. **Replace Generic Error Conversions**: Use factory functions instead of direct ThinktankError instantiation
   - Replace the generic error conversion at lines 523-529 with appropriate factory functions
   - Use createFileNotFoundError for file-related issues
   - Use createModelNotFoundError for model selection issues
   - Use createMissingApiKeyError for API key issues

2. **Enhance Error Cause Chaining**: Ensure the original error is preserved in the cause property
   - This will maintain the full stack trace and error context

3. **Improve Error Suggestions**: Provide more specific suggestions based on error types
   - Add relevant troubleshooting tips for common errors

4. **Update Error Tests**: Modify the tests in runThinktank-error-handling.test.ts to verify the new behavior
   - Test that proper error types are thrown
   - Verify error messages contain helpful information

### Reasoning for Approach
This approach takes advantage of the specialized error factory functions that are already implemented in the error system. By using these functions instead of direct ThinktankError instantiation, we'll have more consistent error messages, better categorization, and more helpful suggestions. The factory functions analyze the specific error context to provide tailored help.

The changes will be minimally invasive, maintaining the core logic while enhancing the user experience through better error messages. We'll maintain backward compatibility with any code expecting ThinktankError instances.