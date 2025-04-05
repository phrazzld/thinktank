# Update CLI command error handling

## Task Goal
Update try/catch blocks in CLI command files to use the new error class hierarchy instead of generic error handling.

## Implementation Approach
After analyzing the code, I'll use the following approach to update the error handling in the CLI command files:

1. Replace generic `catch (error)` blocks with more specific error handling that:
   - Detects specific error conditions and creates appropriate error instances
   - Adds helpful suggestions for common errors in each command context
   - Preserves the existing error handling flow (passing to the central `handleError` function)

2. For each command file (`models.ts`, `run.ts`, and `config.ts`):
   - Update error handling in `action` functions
   - Add specific error detection and wrapping
   - Add command-specific suggestions and examples
   - Convert standard errors to the appropriate ThinktankError subclass

3. Update dependencies:
   - Import error classes and factory functions from `src/core/errors.ts`
   - Remove obsolete error handling code (such as the `createFileNotFoundError` import in `run.ts`)

4. Ensure backward compatibility:
   - Use the centralized error format instead of custom formatting
   - Maintain existing functionality while enhancing error messages

This approach ensures consistent error handling throughout the CLI commands while providing more specific guidance based on the context of each command.

## Reasoning for This Approach
I selected this approach because:

1. It maintains consistency with the existing error handling system already implemented in the providers and core components.
2. It leverages the specialized error classes and factory functions from the centralized error system.
3. It improves user experience by providing more specific error messages and suggestions based on the command context.
4. It allows for more granular error categorization and handling.
5. It removes the redundant error handling code and consolidates the logic in the centralized error system.

Working with the existing pattern in the main CLI error handler, I'll use the specialized error classes like `ConfigError`, `ApiError`, and `ValidationError` to provide consistent error handling across all CLI commands.