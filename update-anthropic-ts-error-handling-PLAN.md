# Update anthropic.ts error handling

## Goal

The goal of this task is to update the error handling in the Anthropic provider implementation to use the new centralized error system, specifically replacing the custom `AnthropicProviderError` with the new `ApiError` class from `src/core/errors.ts`.

## Current State Analysis

Currently, the Anthropic provider (`src/providers/anthropic.ts`) implements its own error class:

```typescript
export class AnthropicProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'AnthropicProviderError';
  }
}
```

This error class is used in the following places:
1. In the `getClient()` method when the API key is missing
2. In the `generate()` method's catch block to handle API errors
3. In the `listModels()` method's catch block for API errors during model listing

The new error system provides a more comprehensive `ApiError` class with better formatting, suggestions, and examples. We need to replace `AnthropicProviderError` with `ApiError` while ensuring backward compatibility with existing code that might be catching the old error type.

## Implementation Approach

After analyzing three different approaches, I've decided to implement the following strategy:

1. **Replace and Extend**: Keep the `AnthropicProviderError` class but refactor it to extend `ApiError` instead of `Error`
   - This maintains backward compatibility with existing code that expects `AnthropicProviderError`
   - It inherits all the functionality of `ApiError` including formatting, categorization, and suggestions
   - It automatically sets the `providerId` to 'anthropic' for consistent error reporting

2. **Enhance Error Contexts**: Add more detailed error information using the options in `ApiError`
   - Provide specific suggestions for different error types (missing API key, rate limiting, etc.)
   - Add helpful examples where appropriate
   - Include cause chaining for better debugging and error reporting

3. **Update Error Creation**: Update the catch blocks to use the new error class with appropriate parameters
   - Create convenience factory functions for common error patterns specific to Anthropic
   - Add more specific error handling for different Anthropic API error types

4. **Update Tests**: Ensure tests continue to pass with the new error handling
   - Update test expectations to account for the new error properties
   - Add new tests for any new error handling functionality

## Reasoning

I chose this approach over alternatives for the following reasons:

1. **Why not completely remove `AnthropicProviderError`?**
   - Complete removal would break backward compatibility with existing code that relies on catching `AnthropicProviderError`
   - The chosen approach provides a smooth transition while allowing code to benefit from the new error system immediately

2. **Why not use factory functions exclusively?**
   - While factory functions are useful for common error patterns, maintaining a dedicated error class provides more flexibility
   - A provider-specific error class allows for provider-specific error handling logic that might be needed in the future
   - The class approach follows the pattern established in other provider implementations

3. **Why not create separate error classes for different error types?**
   - The `ApiError` class with its `providerId` property already provides sufficient categorization
   - Additional error classes would add complexity without significant benefits
   - The chosen approach is more maintainable and follows the project's pattern of simplicity

This approach provides the best balance between backward compatibility, code clarity, and adherence to the new error system's design principles.