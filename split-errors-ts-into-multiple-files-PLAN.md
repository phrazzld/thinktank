# Split errors.ts into multiple files

## Task Goal
Refactor the large `errors.ts` (996 lines) into multiple files organized by error category (e.g., ApiErrors.ts, ConfigErrors.ts) to improve maintainability.

## Current Implementation Analysis
The current `errors.ts` file contains approximately 1000 lines of code including:

1. Error categories constant
2. Base `ThinktankError` class
3. Multiple specialized error classes:
   - `ConfigError`
   - `ApiError`
   - `FileSystemError`
   - `ValidationError`
   - `NetworkError`
   - `PermissionError`
   - `InputError`
4. Factory functions:
   - `createFileNotFoundError`
   - `createModelFormatError`
   - `createMissingApiKeyError`
   - `createModelNotFoundError`

This single large file has several issues:
- Hard to navigate and find specific error types
- Harder to maintain as all error-related code is in one place
- No clear separation of concerns between different types of errors
- Makes it more difficult to extend the error system with new error types

## Implementation Approach

After considering multiple approaches, I've selected a hybrid approach with a well-structured directory organization:

```
src/core/errors/
├── index.ts         # Re-exports everything for backward compatibility
├── base.ts          # Base ThinktankError class and errorCategories
├── types/           # Specialized error classes
│   ├── index.ts     # Re-exports all error types
│   ├── api.ts       # API-related errors
│   ├── config.ts    # Configuration-related errors
│   ├── filesystem.ts # File system errors
│   ├── input.ts     # Input-related errors
│   ├── network.ts   # Network-related errors
│   └── permission.ts # Permission-related errors
└── factories/       # Error factory functions
    ├── index.ts     # Re-exports all factory functions
    ├── api.ts       # API-related factory functions
    ├── config.ts    # Configuration-related factory functions
    └── filesystem.ts # File system factory functions
```

This structure provides:

1. **Clean organization by error domain**: Each error type is grouped with related errors
2. **Separation of concerns**: Base classes, error types, and factory functions are separate
3. **Maintainable import patterns**: Clear imports through index files
4. **Backward compatibility**: The main index.ts preserves the original API
5. **Room for growth**: Easy to add new error types without modifying existing files

## Implementation Steps

1. Create the directory structure
2. Extract the base error class and error categories to base.ts
3. Break up each error class into its own file in the types/ directory
4. Break up the factory functions into their own files in the factories/ directory
5. Set up proper import/export patterns
6. Create index.ts files to re-export everything
7. Update import statements throughout the codebase
8. Run tests to ensure everything works correctly

## Advantages of this Approach

1. **Modularity**: Each error type is in its own file, making it easier to locate and modify
2. **Maintainability**: Smaller files are easier to understand and maintain
3. **Extensibility**: New error types can be added without modifying existing files
4. **Backward compatibility**: The main index.ts preserves the original API, minimizing impact on existing code
5. **Improved organization**: Error types are organized by domain, making the codebase more intuitive

## Potential Challenges

1. **Import complexity**: We'll need to ensure imports are properly set up to avoid circular dependencies
2. **Maintaining backward compatibility**: We must ensure existing code continues to work with the new structure
3. **Testing**: Need to ensure thorough testing to catch any issues with the refactoring

Overall, this refactoring will significantly improve the maintainability of the error system while preserving its functionality and making it easier to extend in the future.