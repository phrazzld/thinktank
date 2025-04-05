# Update CLI error handler in index.ts

## Goal
Update the CLI error handler function in index.ts to fully utilize the new error formatting capabilities provided by the centralized error system, specifically the `format()` method from `ThinktankError`.

## Current State
- The CLI has a `handleError` function that already uses `ThinktankError.format()` for basic formatting
- It manually adds some additional information for causes
- It has special handling for filesystem errors with a help message
- It handles standard `Error` objects and unknown errors separately

## Implementation Plan

### 1. Enhance Error Cause Handling
- Modify the handling of error causes to leverage the `ThinktankError` properties more effectively
- Consider if we can remove the manual cause handling since the formatting might already include it

### 2. Improve Category-Specific Guidance
- Expand the special handling beyond just "File System" errors
- Add specialized guidance for API errors, configuration errors, and other common error categories
- Ensure each category has appropriate tips for resolving issues

### 3. Implement Enhanced Error Type Detection
- Check for specific error subclasses (ApiError, ConfigError, etc.) to provide more precise help
- Add provider-specific guidance for ApiError instances with different providerIds

### 4. Ensure Consistent Formatting for Non-ThinktankErrors
- Improve handling of standard Error objects by wrapping them in appropriate ThinktankError subclasses
- Convert unknown errors into a consistent format that aligns with ThinktankError formatting

## Reasoning
- This approach maintains backward compatibility while improving error messages
- It leverages the existing `format()` method on ThinktankError while enhancing it with context-specific guidance
- Breaking error handling into category-specific guidance makes errors more actionable for users
- Converting non-ThinktankErrors to the new format ensures a consistent user experience
- Special handling for ApiErrors allows for provider-specific guidance, which is especially helpful

## Implementation Details
The updated `handleError` will:
1. First check if the error is a ThinktankError and use its format() method
2. For specific error categories, it will add additional help text based on the error type
3. For standard Errors, it will wrap them in an appropriate ThinktankError subclass
4. For unknown errors, it will create a new ThinktankError with a generic message

This implementation will provide more consistent, helpful error messages to users while maintaining the existing functionality.