# Refactor _processInput to use FileSystem interface

## Goal
Modify the `_processInput` function in `runThinktankHelpers.ts` to use the `FileSystem` interface instead of making direct file system operations. This will improve testability and maintain consistent dependency injection patterns throughout the application.

## Implementation Approach

I'll use the Direct Parameter Injection approach, following the same pattern used in the recently completed `_setupWorkflow` and `_executeQueries` implementations:

1. Modify the `ProcessInputParams` interface in `runThinktankTypes.ts` to add a `fileSystem` parameter of type `FileSystem`.
2. Update the `_processInput` function in `runThinktankHelpers.ts` to use the injected `fileSystem` instead of direct imports.
3. Update the `runThinktank.ts` file to instantiate a `ConcreteFileSystem` once and pass it to the `_processInput` function.
4. Make the necessary adjustments to the underlying `inputHandler.processInput` and `readContextPaths` functions to accept and use the FileSystem interface.

### Key Changes:

1. In `runThinktankTypes.ts`:
   - Add `fileSystem: FileSystem` parameter to the `ProcessInputParams` interface
   - Add the corresponding import for the `FileSystem` interface

2. In `runThinktankHelpers.ts`:
   - Update the `_processInput` function to use the injected `fileSystem`
   - Pass the `fileSystem` to `processInput` and `readContextPaths` functions
   - Consider refactoring to have less direct file system operations

3. In the `inputHandler.ts` and `fileReader.ts` modules:
   - Modify `processInput` to optionally accept a `fileSystem` parameter
   - Update `readContextPaths` to optionally accept a `fileSystem` parameter

4. In `runThinktank.ts`:
   - Instantiate a `ConcreteFileSystem` for injection
   - Pass the `fileSystem` to `_processInput`

## Reasoning for Approach

1. **Consistency**: This approach maintains consistency with the other recently refactored functions (`_setupWorkflow` and `_executeQueries`), following the same dependency injection pattern.

2. **Minimal Changes**: By modifying the parameter interface and updating the function, we make minimal changes to the codebase while achieving the desired separation of concerns.

3. **Testability**: Adding the FileSystem interface as a parameter makes the function much easier to test with mock implementations.

4. **Direct Dependencies**: Making the dependencies explicit in the function signature clarifies the function's requirements.

5. **Error Handling Pattern**: The existing error handling pattern in `_processInput` can be preserved, with adjustments to leverage the FileSystem interface's error handling.

This approach follows the same pattern used in previous refactorings, maintaining consistency while improving the testability and maintainability of the codebase.