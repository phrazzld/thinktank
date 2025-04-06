# Implement Setup Workflow Helper

## Task Goal
Create the `_setupWorkflow` helper function to handle configuration loading, run name generation, and output directory creation with proper error handling and spinner updates.

## Implementation Approach

I'll implement the `_setupWorkflow` helper function in a new TypeScript file called `runThinktankHelpers.ts` in the workflow directory. This function will extract and encapsulate the first three phases of the existing runThinktank workflow:

1. **Configuration Loading**: Load user configuration using the configManager's loadConfig function
2. **Run Name Generation**: Generate a friendly name for the run using nameGenerator 
3. **Output Directory Creation**: Create the output directory using outputHandler's createOutputDirectory function

The implementation will:

1. **Follow the Interface Contract**: Implement the function according to the `SetupWorkflowParams` and `SetupWorkflowResult` interfaces defined in runThinktankTypes.ts
2. **Handle Errors Properly**: Follow the error handling contract to properly catch, categorize, and wrap errors
3. **Manage Spinner Updates**: Update the spinner state at appropriate points during execution
4. **Use Strong Typing**: Leverage TypeScript's type system to ensure type safety and maintainability

### Function Signature and Flow:
```typescript
async function _setupWorkflow({ spinner, options }: SetupWorkflowParams): Promise<SetupWorkflowResult> {
  // 1. Load configuration
  // 2. Generate friendly run name
  // 3. Create output directory
  // 4. Return the result object with all required properties
  // 5. Handle errors according to the contract
}
```

## Reasoning

This approach was selected for the following reasons:

1. **Clean Separation of Concerns**: By extracting the setup logic into a dedicated helper function, we make the code more modular and easier to test.

2. **Consistent Error Handling**: The function will implement standardized error handling patterns defined in the error handling contract, ensuring that errors are properly categorized and wrapped.

3. **Improved Testability**: With well-defined input and output interfaces, the function becomes easier to test in isolation.

4. **Progressive Refactoring**: This approach allows for incremental refactoring of the runThinktank function, with each helper function being implemented and tested individually before integrating them all together.

5. **Spinner Management**: By passing the spinner as a parameter, we maintain consistent visual feedback throughout the workflow while clearly defining responsibility for spinner updates.

Alternative approaches I considered but rejected:

- **Stateful Class**: Creating a class to manage workflow state and operations would add unnecessary complexity for what is essentially a functional pipeline.
- **Global State Management**: Using global state or singletons for managing workflow data would make testing more difficult and introduce tight coupling.
- **Duplicating Logic**: Keeping the setup logic in the main function while adding helper functions would lead to code duplication and maintenance issues.