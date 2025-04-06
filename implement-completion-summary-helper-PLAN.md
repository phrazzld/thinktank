# Implement Completion Summary Helper

## Task Goal
Create the `_logCompletionSummary` helper function that formats and logs the completion summary, handling both success and partial failure scenarios.

## Implementation Approach

I'll implement a dedicated helper function that handles the final summary presentation to the user. This function will:

1. Generate a comprehensive completion summary that includes:
   - Overview of the operation (which models were used, run name, etc.)
   - Success/failure statistics
   - Execution timing information
   - Tree-style breakdown of results by model and error category

2. Format the output using the appropriate styling utilities from consoleUtils to ensure consistency with the rest of the application.

3. Display the formatted summary to the user through the console while maintaining the existing tree-style structure currently used in the runThinktank function.

The implementation will extract and refactor the existing summary functionality from the runThinktank function's formatResultsSummary, adapting it to work with the new helper function structure. This approach ensures consistency with the existing behavior while making the code more modular and maintainable.

### Key Components:
- Tree-style hierarchical display showing successful and failed models
- Error grouping by category
- Timing information
- Success percentage calculation
- Context-aware messaging based on run options

## Alternatives Considered

1. **Table-Based Summary**: Instead of the current tree-style output, implement a more compact table format using a library like cli-table3. This would be more space-efficient but less visually distinctive and might lose some of the hierarchical organization of errors by category.

2. **Customizable Verbosity Levels**: Implement different verbosity levels for the summary (e.g., minimal, standard, detailed) controlled by a new user option. This would allow users to choose how much detail they want in the output, but would add complexity and another option to the interface.

3. **Delegating to outputFormatter**: Move all the logic to the outputFormatter module and just call a function from there. This would centralize all formatting logic but would make the formatter module aware of specific workflow concepts, reducing separation of concerns.

## Reasoning for Selected Approach

I've chosen to refactor the existing tree-style summary approach because:

1. **Consistency**: It maintains the existing user experience and visual style that users are already familiar with, minimizing disruption.

2. **Information Hierarchy**: The tree structure effectively communicates the relationship between models, error categories, and specific error messages in a way that's easy to scan and understand.

3. **Code Reuse**: Much of the logic already exists in the current implementation and can be refactored into a dedicated helper function without reinventing the wheel.

4. **Clear Separation of Concerns**: By moving this logic into a dedicated helper function, we improve the modularity of the codebase without adding unnecessary complexity or dependencies.

This approach provides the best balance between maintaining a consistent user experience and improving the code organization through proper modularization.