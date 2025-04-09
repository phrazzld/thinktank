# Implementation Plan: Simplify runThinktank orchestration

## Task Overview
Rewrite `runThinktank` in `src/workflow/runThinktank.ts` to compose pure functions and perform I/O at boundaries, improving testability by separating I/O operations from data transformation.

## Chosen Approach
After analyzing the three approaches generated, I've selected a hybrid approach that combines elements from Approach 3 (Hybrid - Dedicated File Writing Helper) from the first model and Approach 1 (Centralized I/O within `runThinktank` using DI) from the second model.

### Implementation Summary
1. Extract the file writing logic from `runThinktank` into a dedicated helper function `_writeOutputFiles` that will handle all file I/O operations.
2. Ensure `runThinktank` orchestrates the workflow by calling pure functions for data transformation and using the injected dependencies for I/O operations.
3. Keep console logging within `runThinktank` since it's simpler than file I/O and already uses the injected logger.
4. Update type definitions and tests to reflect this new structure.

## Implementation Details

### 1. Create a Dedicated File Writing Helper Function

Create a new function `_writeOutputFiles` in `src/workflow/runThinktankHelpers.ts`:

```typescript
/**
 * Writes file data to the specified output directory using the injected FileSystem.
 * 
 * This function handles all file I/O operations, including directory creation, 
 * file writing, error handling, and status tracking.
 * 
 * @param files - Array of FileData objects containing files to write
 * @param outputDirectoryPath - Directory where files should be written
 * @param fileSystem - Injected FileSystem interface for I/O operations
 * @param options - Optional settings for file writing
 * @returns Promise resolving to a FileOutputResult with success/failure statistics
 */
export async function _writeOutputFiles(
  files: FileData[],
  outputDirectoryPath: string,
  fileSystem: FileSystem,
  options?: Partial<RunOptions>
): Promise<FileOutputResult> {
  // Start timing for file operations
  const fileWriteStartTime = Date.now();
  
  // Track file write stats
  let succeededWrites = 0;
  let failedWrites = 0;
  const fileDetails: FileWriteDetail[] = [];
  
  // Ensure output directory exists
  try {
    await fileSystem.mkdir(outputDirectoryPath, { recursive: true });
  } catch (error) {
    throw new FileSystemError(`Failed to create output directory: ${outputDirectoryPath}`, {
      cause: error instanceof Error ? error : undefined,
      filePath: outputDirectoryPath
    });
  }
  
  // Process each file
  for (const file of files) {
    const filePath = path.join(outputDirectoryPath, file.filename);
    
    // Create file detail for tracking
    const fileDetail: FileWriteDetail = {
      modelKey: file.modelKey,
      filename: file.filename,
      filePath,
      status: 'pending',
      startTime: Date.now()
    };
    
    fileDetails.push(fileDetail);
    
    try {
      // Create parent directory if needed (for nested paths)
      const parentDir = path.dirname(filePath);
      await fileSystem.mkdir(parentDir, { recursive: true });
      
      // Write the file
      await fileSystem.writeFile(filePath, file.content);
      
      // Update stats
      succeededWrites++;
      
      // Mark as success
      fileDetail.status = 'success';
      fileDetail.endTime = Date.now();
      fileDetail.durationMs = fileDetail.endTime - (fileDetail.startTime || fileDetail.endTime);
    } catch (error) {
      // Update stats
      failedWrites++;
      
      // Mark as error
      fileDetail.status = 'error';
      fileDetail.error = error instanceof Error ? error.message : String(error);
      fileDetail.endTime = Date.now();
      fileDetail.durationMs = fileDetail.endTime - (fileDetail.startTime || fileDetail.endTime);
    }
  }
  
  // Calculate overall timing
  const fileWriteEndTime = Date.now();
  const fileWriteDurationMs = fileWriteEndTime - fileWriteStartTime;
  
  // Create file output result object
  return {
    outputDirectory: outputDirectoryPath,
    files: fileDetails,
    succeededWrites,
    failedWrites,
    timing: {
      startTime: fileWriteStartTime,
      endTime: fileWriteEndTime,
      durationMs: fileWriteDurationMs
    }
  };
}
```

### 2. Refactor the `runThinktank` Function

Update `runThinktank` in `src/workflow/runThinktank.ts` to use the new helper function:

```typescript
export async function runThinktank(options: RunOptions): Promise<string> {
  // Initialize the workflow state to track progress and share context
  const workflowState: Partial<WorkflowState> = {
    options
  };
  
  // Configure spinner throttling based on options
  configureSpinnerFactory({
    useThrottledSpinner: !options.disableSpinnerThrottling
  });
  
  // Initialize the spinner for user feedback
  const spinner = ora('Starting thinktank...') as EnhancedSpinner;
  spinner.start();
  
  try {
    // Create dependency instances for injection in the correct order
    const fileSystem: FileSystem = new ConcreteFileSystem();
    const configManager: ConfigManagerInterface = new ConcreteConfigManager();
    const llmClient: LLMClient = new ConcreteLLMClient(configManager);
    
    // 1. Setup workflow
    const setupResult = await _setupWorkflow({
      spinner,
      options,
      configManager,
      fileSystem
    });
    
    // Update workflow state with setup results
    workflowState.config = setupResult.config;
    workflowState.friendlyRunName = setupResult.friendlyRunName;
    workflowState.outputDirectoryPath = setupResult.outputDirectoryPath;
    
    // 2. Process input
    const inputResult = await _processInput({
      spinner,
      input: options.input,
      contextPaths: options.contextPaths,
      fileSystem
    });
    
    // Update workflow state with input result
    workflowState.inputResult = inputResult.inputResult;
    
    // 3. Select models
    const modelSelectionResult = _selectModels({
      spinner,
      config: setupResult.config,
      options
    });
    
    // Update workflow state with model selection result
    workflowState.modelSelectionResult = modelSelectionResult;
    
    // Early return if no models were selected
    if (modelSelectionResult.models.length === 0) {
      const message = 'No models available after filtering.';
      spinner.warn(styleWarning(message));
      return message;
    }
    
    // Display list of selected models
    spinner.stop();
    const modelList = modelSelectionResult.models
      .map((model, index) => `  ${index + 1}. ${model.provider}:${model.modelId}${!model.enabled ? ' (disabled)' : ''}`)
      .join('\n');
    
    logger.plain(modelList);
    spinner.start();
    
    // 4. Execute queries
    const queryResults = await _executeQueries({
      spinner,
      config: setupResult.config,
      models: modelSelectionResult.models,
      combinedContent: inputResult.combinedContent,
      options,
      llmClient
    });
    
    // Update workflow state with query results
    workflowState.queryResults = queryResults.queryResults;
    
    // 5. Process output: Generate structured data for file and console output
    const processedOutput = _processOutput({
      spinner,
      queryResults: queryResults.queryResults,
      options,
      friendlyRunName: setupResult.friendlyRunName
    });
    
    // 6. Write files to disk using the dedicated helper
    spinner.text = 'Writing files to disk...';
    
    // Call the new helper function to handle file I/O
    const fileOutputResult = await _writeOutputFiles(
      processedOutput.files,
      setupResult.outputDirectoryPath,
      fileSystem,
      options
    );
    
    // Update workflow state with file output results
    workflowState.fileOutputResult = fileOutputResult;
    
    // Update spinner with final file writing status
    spinner.text = `Files written: ${fileOutputResult.succeededWrites} succeeded, ${fileOutputResult.failedWrites} failed`;
    
    // 7. Prepare and log completion summary
    spinner.stop();
    
    // Prepare data for the completion summary
    const summaryData: CompletionSummaryData = {
      totalModels: queryResults.queryResults.responses.length,
      successCount: Object.values(queryResults.queryResults.statuses).filter(s => s.status === 'success').length,
      failureCount: Object.values(queryResults.queryResults.statuses).filter(s => s.status === 'error').length,
      errors: Object.entries(queryResults.queryResults.statuses)
        .filter(([_, status]) => status.status === 'error')
        .map(([modelKey, status]) => ({
          modelKey,
          message: status.message || 'Unknown error',
          category: (status.detailedError && 'category' in status.detailedError) 
            ? (status.detailedError as { category?: string }).category 
            : undefined
        })),
      runName: setupResult.friendlyRunName,
      outputDirectoryPath: setupResult.outputDirectoryPath,
      totalExecutionTimeMs: queryResults.queryResults.timing.durationMs + fileOutputResult.timing.durationMs
    };
    
    // Format the summary data
    const formattedSummary = formatCompletionSummary(summaryData, { 
      useColors: options.useColors !== false 
    });
    
    // Log the formatted summary
    logger.plain(formattedSummary.summaryText);
    if (formattedSummary.errorDetails) {
      formattedSummary.errorDetails.forEach(line => logger.plain(line));
    }
    
    // 8. Show additional metadata if requested
    if (options.includeMetadata) {
      logger.plain('\n' + styleHeader('Execution timing:'));
      
      const totalTime = queryResults.queryResults.timing.durationMs + fileOutputResult.timing.durationMs;
      
      logger.plain(styleDim(`  Total API calls:    ${queryResults.queryResults.timing.durationMs}ms`));
      logger.plain(styleDim(`  File writing:       ${fileOutputResult.timing.durationMs}ms`));
      logger.plain(styleDim(`  Total execution:    ${totalTime}ms`));
      
      logger.plain('\n' + styleHeader('Model timing:'));
      
      Object.entries(queryResults.queryResults.statuses)
        .sort((a, b) => (a[1].durationMs || 0) - (b[1].durationMs || 0))
        .forEach(([model, status]) => {
          if (status.durationMs) {
            const statusIcon = status.status === 'success' ? '+' : 'x';
            logger.plain(styleDim(`  ${statusIcon} ${model}: ${status.durationMs}ms`));
          }
        });
    }
    
    // Return the formatted console output for display
    return processedOutput.consoleOutput;
  } catch (error) {
    // Use the error handling helper to process and rethrow the error
    return _handleWorkflowError({
      spinner,
      error,
      options,
      workflowState
    });
  }
}
```

### 3. Update Type Definitions in `runThinktankTypes.ts`

Make sure the type definitions are updated if needed to reflect the new function signature of `_writeOutputFiles`.

## Justification for Chosen Approach

I selected this hybrid approach for several key reasons:

1. **Targeted Complexity Extraction**: The file writing logic is the most complex I/O operation in `runThinktank`, involving multiple steps (directory creation, file writing, error handling, status tracking). By isolating this complexity into a dedicated helper, we significantly improve both the readability of `runThinktank` and the testability of the file writing logic.

2. **Balanced Implementation**: It provides a good balance between code cleanliness and implementation complexity. Unlike extracting all I/O (including simple console logging) to a separate orchestration layer, this approach focuses on the most significant I/O block while keeping the overall workflow structure intact.

3. **Testability Benefits**: This approach directly addresses the primary goal of improving testability:
   - The file writing logic becomes independently testable by mocking just the `FileSystem` interface.
   - The main `runThinktank` function's tests become simpler as the complex file I/O is delegated to the helper.
   - The approach aligns perfectly with the project's testing philosophy of mocking at external system boundaries.

4. **Incremental Improvement**: It offers a substantial improvement over the current state without requiring a complete restructuring of the workflow. The refactoring is focused and targeted toward the most complex I/O operations.

5. **Maintainability**: By encapsulating the file writing logic in a dedicated function with a clear signature, the code becomes more modular and easier to maintain or extend in the future.

## Testing Strategy

1. Create unit tests for the new `_writeOutputFiles` function:
   - Test successful file writing
   - Test error handling for directory creation failure
   - Test error handling for file writing failures
   - Test statistics tracking for mixed success/failure scenarios

2. Update existing tests for `runThinktank` to:
   - Mock the new `_writeOutputFiles` function
   - Verify it's called with the correct parameters
   - Ensure the workflow state is properly updated based on the helper's return value

3. Run the full test suite to ensure the refactoring doesn't break any existing functionality.