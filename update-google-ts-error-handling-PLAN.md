# Google Provider Error Handling Update Plan

## Goal
Update the error handling in the Google provider to use the centralized error system while maintaining backward compatibility. This will ensure consistent error reporting and improved error messages for users.

## Current State
- The `GoogleProviderError` class currently extends the standard JavaScript `Error` class
- There are three main error handling points in the code:
  1. Missing API key in the `getClient()` method
  2. Various errors in the `generate()` method 
  3. API and network errors in the `listModels()` method
- There are no specific tests for the Google provider's error handling

## Implementation Plan

### 1. Update GoogleProviderError class
- Extend `ApiError` instead of `Error`
- Ensure backward compatibility (still catchable as `GoogleProviderError`)
- Set the `providerId` to 'google'
- Ensure proper error inheritance chain using `Object.defineProperty`

### 2. Enhance error handling in getClient()
- Add specific suggestions for the missing API key error
- Include examples of how to set the API key

### 3. Improve error handling in generate()
- Add specific error detection for common Google API errors
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
- This approach follows the same pattern used for updating the Anthropic and OpenAI providers
- It maintains backward compatibility while adding new features
- It improves the user experience with more helpful error messages
- It makes the error handling consistent across all providers

## Expected Outcomes
- `GoogleProviderError` will be an instance of `ApiError` and `ThinktankError`
- Errors will include helpful suggestions and examples
- Error messages will be better formatted with the provider ID prefix
- All existing code that catches `GoogleProviderError` will continue to work