# Update openrouter.ts error handling

## Goal
Update the error handling in the OpenRouter provider to use the centralized error system while maintaining backward compatibility. This will ensure consistent error reporting and improved error messages for users.

## Current State
- The `OpenRouterProviderError` class currently extends the standard JavaScript `Error` class
- There are three main error handling points in the code:
  1. Missing API key in the `getClient()` method
  2. Various errors in the `generate()` method 
  3. API and network errors in the `listModels()` method
- Error handling is relatively simple with limited contextual suggestions

## Implementation Plan

### 1. Update OpenRouterProviderError class
- Extend `ApiError` instead of `Error`
- Ensure backward compatibility (still catchable as `OpenRouterProviderError`)
- Set the `providerId` to 'openrouter'
- Ensure proper error inheritance chain using `Object.defineProperty`

### 2. Enhance error handling in getClient()
- Add specific suggestions for the missing API key error
- Include examples of how to set the API key

### 3. Improve error handling in generate()
- Add specific error detection for common OpenRouter API errors
- Enhance error messages with suggestions based on error type
- Add examples where appropriate

### 4. Enhance error handling in listModels()
- Improve axios error handling with detailed information
- Add specific error detection for rate limiting, authentication, etc.
- Add helpful suggestions for common errors

### 5. Maintain all existing functionality
- Ensure all existing code paths continue to work
- Preserve backward compatibility for error catching

## Reasoning
- This approach follows the same pattern used for updating the other providers (Anthropic, OpenAI, Google)
- It maintains backward compatibility while adding new features
- It improves the user experience with more helpful error messages
- It makes the error handling consistent across all providers
- OpenRouter is a proxy to multiple models, so specific error handling for model selection and availability is valuable

## Expected Outcomes
- `OpenRouterProviderError` will be an instance of `ApiError` and `ThinktankError`
- Errors will include helpful suggestions and examples
- Error messages will be better formatted with the provider ID prefix
- All existing code that catches `OpenRouterProviderError` will continue to work