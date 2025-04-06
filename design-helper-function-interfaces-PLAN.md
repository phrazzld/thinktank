# Design Helper Function Interfaces

## Task Goal
Define the inputs, outputs, and contracts for each helper function in the runThinktank workflow refactoring. This task focuses on creating TypeScript interfaces for seven helper functions to ensure consistent, maintainable, and properly typed code.

## Implementation Approach

My approach will be to create a new TypeScript file in the workflow directory called `runThinktankTypes.ts` that will contain all interfaces for the helper functions. This dedicated type file will:

1. **Define Common Types**: Create shared types used across multiple helper functions, including spinner management, error handling, and state propagation.

2. **Design Function-Specific Interfaces**: Define input and output interfaces for each helper function based on the existing operational phases identified in the flow analysis.

3. **Establish Error Handling Contracts**: Specify which error types each function is responsible for handling and wrapping.

4. **Document Interface Purposes**: Add thorough JSDoc comments for each interface to explain its purpose, usage constraints, and relationship to the overall workflow.

The interfaces will be organized by their position in the workflow, starting with setup and progressing through to completion handling and error management:

### Helper Functions to Interface
1. `_setupWorkflow`: Configuration loading, run name generation, output directory creation
2. `_processInput`: Input handling from file, stdin, or direct text
3. `_selectModels`: Model selection with proper error handling and warning display
4. `_executeQueries`: Query execution with spinner updates and status tracking
5. `_processOutput`: File writing and console output formatting
6. `_logCompletionSummary`: Result summary formatting and display
7. `_handleWorkflowError`: Error categorization and proper error type creation

## Reasoning

I'm choosing this approach for the following reasons:

1. **Type Separation**: Creating a dedicated types file keeps the interfaces separate from implementation, allowing for better organization and maintainability. It follows the pattern used elsewhere in the codebase (e.g., `src/core/types.ts`).

2. **Clear Contracts**: Defining explicit interfaces establishes strong contracts between functions, making it easier to understand each function's responsibilities, inputs, and outputs.

3. **Incremental Implementation**: This approach allows the helper functions to be implemented incrementally in subsequent tasks, with each function having a clear contract to fulfill.

4. **Type Safety**: Detailed TypeScript interfaces will provide compile-time checks, ensuring that the implementation properly handles all required parameters and returns expected results.

5. **Documentation Value**: The interfaces serve as self-documenting code, making it easier for developers to understand the workflow without having to trace through implementation details.

Alternative approaches I considered but rejected:

- **Inline types within implementation**: This would mix types and implementation, making the code harder to maintain and obscuring the interface contracts.
- **Using separate files for each function's types**: While this would provide good separation, it would result in too many small files and make it harder to understand the relationships between the helper functions.
- **Using simple function signatures without detailed interfaces**: This would be less verbose but would miss the opportunity to clearly document the purpose and constraints of each parameter.