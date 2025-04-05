# Update ModelSelectionError Handling

## Goal
Convert the ModelSelectionError class in src/workflow/modelSelector.ts to fully integrate with the new error hierarchy system, using the specialized error types and factory functions from src/core/errors.ts.

## Current State Analysis
Currently, the ModelSelectionError extends ThinktankError but doesn't fully leverage the new error system. It sets the category correctly to CONFIG but doesn't make use of the factory functions that provide better error messages, suggestions, and examples. The current ModelSelectionError is also not consistently using the name property in a way that ensures it's correctly set for error type checking in tests.

## Implementation Approach
After analyzing the code and considering different approaches, I've decided to implement the following strategy:

1. Refactor ModelSelectionError to specifically extend ConfigError
   - This is more appropriate since model selection errors are configuration-related
   - It will inherit all the ConfigError functionality and properties

2. Update error creation in model selector functions
   - Replace direct ModelSelectionError instantiations with appropriate factory functions where possible
   - Use createModelFormatError and createModelNotFoundError for common error patterns
   - Add consistent setting of error name property to ensure tests pass

3. Update error handling patterns in selectModels
   - Ensure error cause chaining for better debugging
   - Provide more contextual suggestions and examples
   - Maintain the same error categories for backward compatibility

## Reasoning
I chose this approach over alternatives for these reasons:

1. **Extension vs. Replacement**: I considered completely replacing ModelSelectionError with direct usage of ConfigError, but this would break backward compatibility. The chosen approach maintains the ModelSelectionError class while improving its integration with the error system.

2. **Factory Functions**: Using the factory functions where appropriate provides richer error messages and consistent formatting, while reducing code duplication.

3. **Error Hierarchy**: Having ModelSelectionError extend ConfigError rather than base ThinktankError better represents the domain-specific nature of the error and leverages the specialized functionality in ConfigError.

4. **Test Compatibility**: The implementation ensures that existing tests which check for ModelSelectionError type will continue to pass while also supporting the new error hierarchy.

This approach provides the best balance between improving the error handling system and maintaining compatibility with existing code.