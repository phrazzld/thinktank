# Implement Input Processing Helper

## Task Goal
Create the `_processInput` helper function to handle input processing with appropriate spinner text updates and error wrapping using `FileSystemError`.

## Implementation Approach

I will create a function in the `runThinktankHelpers.ts` file that processes user input from various sources (file, stdin, or direct text) while providing visual feedback through spinner updates and properly handling errors. This function will:

1. **Leverage Existing Logic**: Utilize the `processInput` function from `inputHandler.ts` which already contains robust input processing functionality.

2. **Add Spinner Management**: Enhance the user experience by providing clear updates about the input processing status.

3. **Implement Proper Error Handling**: Catch and convert errors to the appropriate error types following the defined error handling contract, emphasizing `FileSystemError` for input-related issues.

The implementation will follow the established pattern from the `_setupWorkflow` helper, maintaining consistency in style and error handling approach.

### Function Structure:

```typescript
export async function _processInput({ 
  spinner, 
  input 
}: ProcessInputParams): Promise<ProcessInputResult> {
  try {
    // 1. Update spinner with processing status
    spinner.text = 'Processing input...';
    
    // 2. Process the input using inputHandler.processInput
    const inputResult = await processInput({ input });
    
    // 3. Update spinner with success information
    spinner.text = `Input processed from ${inputResult.sourceType} (${inputResult.content.length} characters)`;
    
    // 4. Return the properly structured result
    return { inputResult };
  } catch (error) {
    // 5. Handle errors according to error contract
    // ... Error handling logic here ...
  }
}
```

## Reasoning

This approach was chosen for several key reasons:

1. **Separation of Concerns**: The implementation clearly separates spinner management and error handling from the core input processing logic, which remains in the `inputHandler` module. This maintains the single responsibility principle.

2. **Reuse of Existing Functionality**: The approach builds upon the well-developed `processInput` function in `inputHandler.ts`, avoiding code duplication.

3. **Consistent Error Handling**: By following the same error handling pattern as the `_setupWorkflow` function, we maintain consistency throughout the codebase.

4. **Clear User Communication**: The spinner updates provide clear feedback to users about what's happening, improving the user experience.

5. **Adherence to Interfaces**: The implementation strictly follows the type definitions in `runThinktankTypes.ts`, ensuring type safety and contract compliance.

I considered these alternative approaches:

- **Reimplementing Input Processing Logic**: I could rewrite the input processing logic directly in this helper function instead of calling `processInput`, but this would lead to significant code duplication and potential inconsistencies.

- **Using a Different Error Wrapping Strategy**: I could use `InputError` from the `inputHandler` module instead of converting to `FileSystemError`, but this would break the established error handling contract which specifies `FileSystemError` for input-related issues in this part of the workflow.

- **Handling Error Conversion in Main Function**: I could let the main `runThinktank` function handle all error conversion, but this would break the encapsulation and make the helper functions less useful as standalone components.