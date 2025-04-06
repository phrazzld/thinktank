# Improve Error Factory Accessibility

## Goal
Improve discoverability and ease of use of error factory functions by re-exporting them from `core/errors/index.ts`.

## Implementation Approach
Currently, the error factory functions are exported from `core/errors/factories/index.ts`, but not all of them are re-exported from the main `core/errors/index.ts` file. This requires users to either:

1. Import from both `core/errors` and `core/errors/factories` separately, or
2. Import deeply from `core/errors/factories/provider` etc.

The chosen approach will:
1. Re-export ALL error factory functions from the main `core/errors/index.ts` file
2. Maintain the existing exports to ensure backward compatibility 
3. Update the JSDoc comments in the index file to document the available factory functions

## Reasoning
This approach provides several advantages:

- **Simplicity**: Developers only need to import from a single location (`core/errors`) to access both error classes and factory functions
- **Discoverability**: All error-related functionality is available from one central location
- **Consistency**: Follows the pattern of having a single entry point for a module's functionality
- **Backward Compatibility**: Maintains existing import paths to avoid breaking changes

Alternative approaches like creating a dedicated `factories.ts` file or restructuring the entire error system would be more invasive and risk breaking existing code.