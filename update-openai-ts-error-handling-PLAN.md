# Update openai.ts error handling

## Goal

The goal of this task is to update the error handling in the OpenAI provider implementation to use the new centralized error system, specifically replacing the custom `OpenAIProviderError` with the new `ApiError` class from `src/core/errors.ts`.

## Current State Analysis

Currently, the OpenAI provider (`src/providers/openai.ts`) implements its own error class:

```typescript
export class OpenAIProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'OpenAIProviderError';
  }
}
```

This error class is used in the following places:
1. In the `getClient()` method when the API key is missing
2. In the `generate()` method's catch block to handle API errors
3. In the `listModels()` method's catch block for API errors during model listing

The new error handling system provides an `ApiError` class with improved formatting, additional context (suggestions and examples), and built-in prefixing of provider ID to error messages. We've already refactored the `AnthropicProviderError` class in `anthropic.ts` to use this system, and we'll follow a similar approach.

## Implementation Approach

After considering multiple approaches, I've chosen to implement the following strategy:

1. **Extend ApiError**: Refactor the `OpenAIProviderError` class to extend `ApiError` instead of `Error`
   - This maintains backward compatibility with existing code that catches `OpenAIProviderError`
   - It inherits the formatting and categorization functionality from `ApiError`
   - It will automatically include the provider ID in error messages

2. **Enhance Error Context**: Add detailed suggestions and examples to error messages
   - Provide specific suggestions for common error cases (missing API key, rate limiting, authentication issues)
   - Include examples of how to fix the issues where applicable
   - Add appropriate error categorization for better user feedback

3. **Improve Error Detection**: Add more specific error handling based on error message content
   - Detect rate limiting errors through status codes or error messages
   - Identify authentication issues and provide targeted suggestions
   - Handle model-specific errors with appropriate guidance

4. **Update Tests**: Expand test coverage to verify the new error handling
   - Update existing tests to check for new error properties and inheritance
   - Add tests for specific error types and their suggestions
   - Ensure backward compatibility with code expecting `OpenAIProviderError`

## Reasoning

I selected this approach for the following reasons:

1. **Consistency with Anthropic Provider**: Following the same pattern used in the Anthropic provider maintains consistency across the codebase, making it easier to understand and maintain.

2. **Backward Compatibility**: By extending `ApiError` but keeping the `OpenAIProviderError` class, we ensure existing code that catches these errors will continue to work, while getting the benefits of the new error system.

3. **Enhanced Error Information**: The OpenAI API can return various error types, and providing specific suggestions for each improves the developer experience by making it clearer how to fix issues.

4. **Code Reuse**: The `ApiError` class already implements prefixing of provider IDs and formatting, so we can leverage this functionality without duplicating code.

Alternative approaches I considered but rejected:

- **Direct use of ApiError without a custom class**: This would break backward compatibility with code that catches `OpenAIProviderError`.

- **Creating error factory functions only**: While factory functions are useful for common patterns, keeping the provider-specific error class provides more flexibility for future provider-specific error handling.

- **Complete rewrite of error handling**: A more radical approach would require more extensive testing and risk introducing bugs for minimal added benefit.

The chosen approach provides a good balance of improved error handling, backward compatibility, and code clarity.