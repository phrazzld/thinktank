/**
 * Helper functions for the runThinktank workflow
 * 
 * This file contains the implementation of helper functions that encapsulate
 * distinct phases of the runThinktank workflow, making the main function more
 * modular and easier to maintain.
 */
import { loadConfig } from '../core/configManager';
import { generateFunName } from '../utils/nameGenerator';
import { createOutputDirectory, processOutput, OutputHandlerError } from './outputHandler';
import { 
  ThinktankError,
  ConfigError,
  ApiError,
  FileSystemError,
  PermissionError,
  ValidationError,
  NetworkError,
  errorCategories
} from '../core/errors';
import { categorizeError } from '../core/errors/utils/categorization';
import { 
  styleInfo, 
  styleSuccess, 
  styleWarning, 
  colors,
  styleDim
} from '../utils/consoleUtils';
import {
  SetupWorkflowParams,
  SetupWorkflowResult,
  ProcessInputParams,
  ProcessInputResult,
  SelectModelsParams,
  SelectModelsResult,
  ExecuteQueriesParams,
  ExecuteQueriesResult,
  ProcessOutputParams,
  ProcessOutputResult,
  LogCompletionSummaryParams,
  LogCompletionSummaryResult,
  HandleWorkflowErrorParams
} from './runThinktankTypes';
import { processInput, InputError } from './inputHandler';
import { selectModels, ModelSelectionError } from './modelSelector';
import { executeQueries, QueryExecutorError, ModelQueryStatus } from './queryExecutor';

/**
 * Setup workflow helper function
 * 
 * Handles configuration loading, run name generation, and output directory creation
 * with proper error handling and spinner updates.
 * 
 * @param params - Parameters containing the spinner and options
 * @returns An object containing the configuration, run name, and output directory path
 * @throws 
 *   - ConfigError when config loading fails
 *   - FileSystemError when directory creation fails
 *   - PermissionError when permission issues occur
 */
export async function _setupWorkflow({ 
  spinner, 
  options 
}: SetupWorkflowParams): Promise<SetupWorkflowResult> {
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    spinner.text = 'Configuration loaded successfully';
    
    // 2. Generate a friendly run name
    spinner.text = 'Generating run identifier...';
    const friendlyRunName = generateFunName();
    spinner.info(styleInfo(`Run name: ${styleSuccess(friendlyRunName)}`));
    spinner.start(); // Restart spinner for next step
    
    // 3. Create output directory
    spinner.text = 'Creating output directory...';
    // Determine directory identifier based on options
    const directoryIdentifier = options.specificModel || options.groupName;
    const outputDirectoryPath = await createOutputDirectory({
      outputDirectory: options.output,
      directoryIdentifier,
      friendlyRunName
    });
    spinner.info(styleInfo(`Output directory: ${outputDirectoryPath} (Run: ${friendlyRunName})`));
    spinner.start(); // Restart spinner for next step
    
    // Update spinner for final state
    spinner.text = 'Setup completed successfully';
    
    // Return the result object with all required properties
    return {
      config,
      friendlyRunName,
      outputDirectoryPath
    };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's already a ConfigError, just rethrow it
    if (error instanceof ConfigError) {
      throw error;
    }
    
    // If it's a FileSystemError or PermissionError, rethrow it
    if (error instanceof FileSystemError || error instanceof PermissionError) {
      throw error;
    }
    
    // Handle NodeJS.ErrnoException for file system errors
    if (
      error instanceof Error && 
      'code' in error && 
      typeof (error as NodeJS.ErrnoException).code === 'string'
    ) {
      const nodeError = error as NodeJS.ErrnoException;
      
      // Permission errors
      if (nodeError.code === 'EACCES' || nodeError.code === 'EPERM') {
        throw new PermissionError(`Permission denied: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check that you have sufficient permissions for the directory',
            'Try specifying a different output directory with --output'
          ]
        });
      }
      
      // Directory/file not found
      if (nodeError.code === 'ENOENT') {
        throw new FileSystemError(`File or directory not found: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check that the file exists at the specified path',
            `Current working directory: ${process.cwd()}`
          ]
        });
      }
      
      // Other file system errors
      throw new FileSystemError(`File system error: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check disk space and permissions',
          'Verify the path is valid'
        ]
      });
    }
    
    // For config-related errors during setup, wrap in ConfigError
    if (spinner.text.includes('Loading configuration') || spinner.text.includes('configuration')) {
      throw new ConfigError(`Configuration error: ${error instanceof Error ? error.message : String(error)}`, {
        cause: error instanceof Error ? error : undefined,
        suggestions: [
          'Check that your configuration file is valid JSON',
          'Verify the configuration path is correct'
        ]
      });
    }
    
    // For directory creation errors, wrap in FileSystemError
    if (spinner.text.includes('Creating output directory') || spinner.text.includes('directory')) {
      throw new FileSystemError(`Error creating output directory: ${error instanceof Error ? error.message : String(error)}`, {
        cause: error instanceof Error ? error : undefined,
        suggestions: [
          'Check that the parent directory exists and is writable',
          'Verify that there is sufficient disk space'
        ]
      });
    }
    
    // Generic ThinktankError for other cases
    throw new ThinktankError(`Error during workflow setup: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      category: errorCategories.UNKNOWN
    });
  }
}

/**
 * Input processing helper function
 * 
 * Handles input processing from various sources (file, stdin, or direct text)
 * with proper error handling and spinner updates.
 * 
 * @param params - Parameters containing the spinner and input string
 * @returns An object containing the processed input result
 * @throws 
 *   - FileSystemError when input processing fails due to file system issues
 *   - ThinktankError for other unexpected errors
 */
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
    // Handle specific error types according to the error handling contract
    
    // If it's an InputError, convert to FileSystemError
    if (error instanceof InputError) {
      const errorMessage = error.message.toLowerCase();
      
      // File not found errors
      if (errorMessage.includes('not found') || errorMessage.includes('enoent')) {
        throw new FileSystemError(`File not found: ${input}`, {
          cause: error,
          filePath: input,
          suggestions: [
            'Check that the file exists at the specified path',
            `Current working directory: ${process.cwd()}`,
            'Use an absolute path if necessary'
          ]
        });
      }
      
      // Permission errors
      if (errorMessage.includes('permission') || errorMessage.includes('access') || errorMessage.includes('eacces')) {
        throw new FileSystemError(`Permission denied: ${error.message}`, {
          cause: error,
          filePath: input,
          suggestions: [
            'Check that you have read permissions for the file',
            'Try using a different input source with proper permissions'
          ]
        });
      }
      
      // Empty input error
      if (errorMessage.includes('input is required') || errorMessage.includes('empty')) {
        throw new FileSystemError(`Input is required`, {
          cause: error,
          suggestions: [
            'Provide a file path, stdin indicator (-), or a direct text prompt',
            'Example: thinktank run prompt.txt',
            'Example: cat prompt.txt | thinktank run -',
            'Example: thinktank run "Your prompt text here"'
          ]
        });
      }
      
      // Stdin errors
      if (errorMessage.includes('stdin') || errorMessage.includes('timeout')) {
        throw new FileSystemError(`Input processing error: ${error.message}`, {
          cause: error,
          suggestions: [
            'Make sure you are piping content to stdin when using "-" as input',
            'Check that the input is not empty',
            'Consider using a file instead if stdin is not working'
          ]
        });
      }
      
      // Pass through the original error message to match test expectations
      throw new FileSystemError(error.message, {
        cause: error,
        suggestions: [
          'Check that your input is valid and accessible',
          'Try using a different input method (file, stdin, or direct text)'
        ]
      });
    }
    
    // Handle NodeJS.ErrnoException for file system errors
    if (
      error instanceof Error && 
      'code' in error && 
      typeof (error as NodeJS.ErrnoException).code === 'string'
    ) {
      const nodeError = error as NodeJS.ErrnoException;
      
      // Permission errors
      if (nodeError.code === 'EACCES' || nodeError.code === 'EPERM') {
        throw new FileSystemError(`Permission denied when accessing input: ${error.message}`, {
          cause: error,
          filePath: input,
          suggestions: [
            'Check file permissions',
            'Try using a different input file'
          ]
        });
      }
      
      // File not found
      if (nodeError.code === 'ENOENT') {
        throw new FileSystemError(`File not found: ${input}`, {
          cause: error,
          filePath: input,
          suggestions: [
            'Check that the file exists at the specified path',
            `Current working directory: ${process.cwd()}`
          ]
        });
      }
      
      // Other file system errors
      throw new FileSystemError(`File system error while processing input: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check file permissions and path',
          'Verify that the file exists and is readable'
        ]
      });
    }
    
    // For unexpected errors, wrap in ThinktankError
    throw new ThinktankError(`Error processing input: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      category: errorCategories.INPUT,
      suggestions: [
        'This is an unexpected error',
        'Try a different input method or file'
      ]
    });
  }
}

/**
 * Model selection helper function
 * 
 * Handles model selection with warnings display, error handling, and appropriate spinner updates.
 * Uses ConfigError for wrapping errors.
 * 
 * @param params - Parameters containing the spinner, config, and options
 * @returns An object containing the model selection result and a mode description
 * @throws 
 *   - ConfigError when model selection fails
 */
export function _selectModels({
  spinner,
  config,
  options
}: SelectModelsParams): SelectModelsResult {
  try {
    // 1. Update spinner with selection status
    spinner.text = 'Selecting models...';
    
    // 2. Determine selection mode description based on options
    let modeDescription: string;
    if (options.specificModel) {
      modeDescription = `Specific model: ${options.specificModel}`;
    } else if (options.groupName) {
      modeDescription = `Group: ${options.groupName}`;
    } else if (options.models && options.models.length > 0) {
      modeDescription = `Selected models: ${options.models.join(', ')}`;
    } else {
      modeDescription = 'All enabled models';
    }
    
    // 3. Select models using the modelSelector
    const modelSelectionResult = selectModels(config, {
      specificModel: options.specificModel,
      groupName: options.groupName,
      models: options.models,
      groups: options.groups,
      includeDisabled: true,  // Always include disabled models in results
      validateApiKeys: true   // Always validate API keys
    });
    
    // 4. Display warnings if any
    if (modelSelectionResult.warnings.length > 0) {
      for (const warning of modelSelectionResult.warnings) {
        spinner.warn(styleWarning(warning));
      }
      // Restart spinner after displaying warnings
      spinner.start();
    }
    
    // 5. Update spinner with success information
    const modelCount = modelSelectionResult.models.length;
    spinner.text = `Models selected: ${modelCount} ${modelCount === 1 ? 'model' : 'models'}`;
    
    // Display model list in spinner info
    const modelList = modelSelectionResult.models
      .map(model => `${model.provider}:${model.modelId}`)
      .join(', ');
    
    spinner.info(styleInfo(`Using ${modelCount} ${modelCount === 1 ? 'model' : 'models'}: ${styleSuccess(modelList)}`));
    spinner.start(); // Restart spinner for next step
    
    // Ensure spinner has final status text for consistent state
    spinner.text = `Model selection complete: ${modelCount} ${modelCount === 1 ? 'model' : 'models'} selected`;
    
    // 6. Return the result
    return {
      modelSelectionResult,
      modeDescription
    };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's a ModelSelectionError or an object with name 'ModelSelectionError', wrap it in ConfigError
    if (error instanceof ModelSelectionError) {
      // These errors come from the modelSelector module and need to be wrapped
      // in ConfigError to maintain consistent error types
      throw new ConfigError(error.message, {
        cause: error,
        suggestions: error.suggestions,
        examples: error.examples
      });
    } 
    
    // For test mocks that aren't proper Error instances but have the right structure
    if (
      error && 
      typeof error === 'object' && 
      'name' in error && 
      error.name === 'ModelSelectionError' && 
      'message' in error
    ) {
      const errorObj = error as { 
        message: string; 
        suggestions?: string[]; 
        examples?: string[];
      };
      
      throw new ConfigError(errorObj.message, {
        cause: undefined,
        suggestions: errorObj.suggestions,
        examples: errorObj.examples
      });
    }
    
    // If it's already a ConfigError, just rethrow it
    if (error instanceof ConfigError) {
      throw error;
    }
    
    // For unexpected errors, wrap in ConfigError
    throw new ConfigError(`Error selecting models: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      suggestions: [
        'Check your configuration file for valid model definitions',
        'Verify the model or group name is spelled correctly',
        'Use "thinktank models" to list available models'
      ]
    });
  }
}

/**
 * Query execution helper function
 * 
 * Handles execution of queries to LLM providers with appropriate spinner updates
 * and error handling. Uses ApiError for categorizing and propagating errors.
 * 
 * @param params - Parameters containing spinner, config, models, prompt, and options
 * @returns Promise resolving to an object containing query execution results
 * @throws
 *   - ApiError when query execution fails due to API errors
 *   - ThinktankError for other unexpected errors
 */
export async function _executeQueries({
  spinner,
  config,
  models,
  prompt,
  options
}: ExecuteQueriesParams): Promise<ExecuteQueriesResult> {
  try {
    // 1. Update spinner with execution status
    spinner.text = `Executing queries to ${models.length} ${models.length === 1 ? 'model' : 'models'}...`;
    
    // 2. Extract relevant options for the query executor
    const queryOptions = {
      prompt,
      systemPrompt: options.systemPrompt,
      enableThinking: options.enableThinking,
      // Only pass timeoutMs if it's defined in options
      ...(options.timeoutMs !== undefined && { timeoutMs: options.timeoutMs }),
      
      // Define status update callback to update spinner with model progress
      onStatusUpdate: (modelKey: string, status: ModelQueryStatus) => {
        if (status.status === 'running') {
          spinner.text = `Querying ${modelKey}...`;
        } else if (status.status === 'success') {
          spinner.text = `Received response from ${modelKey} (${status.durationMs}ms)`;
        } else if (status.status === 'error') {
          spinner.text = `Error from ${modelKey}: ${status.message || 'Unknown error'}`;
        }
      }
    };
    
    // 3. Execute queries to models
    const queryResults = await executeQueries(config, models, queryOptions);
    
    // 4. Count successes and failures
    const successCount = queryResults.responses.filter(r => !r.error).length;
    const failureCount = queryResults.responses.filter(r => !!r.error).length;
    
    // 5. Update spinner with appropriate success/warning message
    if (failureCount === 0) {
      // All models succeeded
      spinner.text = `Queries completed successfully: ${successCount} ${successCount === 1 ? 'model' : 'models'}`;
      
      // Display timing information
      const totalTime = queryResults.timing.durationMs / 1000;
      spinner.info(styleInfo(`Queried ${successCount} ${successCount === 1 ? 'model' : 'models'} in ${totalTime.toFixed(2)}s`));
      spinner.start(); // Restart spinner after info message
    } else if (successCount > 0) {
      // Partial success
      spinner.text = `Queries completed with some failures`;
      spinner.warn(styleWarning(`Some models failed to respond (${failureCount} of ${models.length})`));
      
      // Display counts of successes and failures
      spinner.info(styleInfo(`Results: ${successCount} succeeded, ${failureCount} failed`));
      spinner.start(); // Restart spinner after info message
    } else {
      // All models failed
      spinner.text = `All queries failed`;
      spinner.warn(styleWarning(`All models failed to respond (${failureCount} ${failureCount === 1 ? 'model' : 'models'})`));
      
      // Display failure details
      const errorSummary = queryResults.responses
        .map(r => `${r.configKey}: ${r.error}`)
        .join(', ');
        
      spinner.info(styleInfo(`Errors: ${errorSummary}`));
      spinner.start(); // Restart spinner after info message
    }
    
    // Set final spinner text for consistent state
    spinner.text = `Query execution complete: ${successCount} succeeded, ${failureCount} failed`;
    
    // 6. Return the results
    return { queryResults };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's already an ApiError, just rethrow it
    if (error instanceof ApiError) {
      throw error;
    }
    
    // If it's a QueryExecutorError, convert to ApiError 
    if (error instanceof QueryExecutorError) {
      throw new ApiError(error.message, {
        cause: error,
        suggestions: error.suggestions
      });
    }
    
    // For unexpected errors, wrap in ApiError
    throw new ApiError(`Error executing queries: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      suggestions: [
        'Check your network connection',
        'Verify API keys for the models are correctly set',
        'Check if the models are available and not experiencing downtime'
      ]
    });
  }
}

/**
 * Output processing helper function
 * 
 * Handles file writing and console output formatting with spinner updates.
 * Catches and wraps errors using `FileSystemError` when needed.
 * 
 * @param params - Parameters containing spinner, query results, output directory path, and options
 * @returns Promise resolving to an object containing file output result and console formatted output
 * @throws 
 *   - FileSystemError when file writing fails
 *   - PermissionError when permission issues occur
 *   - ThinktankError for other unexpected errors
 */
export async function _processOutput({
  spinner,
  queryResults,
  outputDirectoryPath,
  options,
  friendlyRunName
}: ProcessOutputParams): Promise<ProcessOutputResult> {
  try {
    // 1. Update spinner with processing status
    spinner.text = 'Processing output and writing results...';
    
    // 2. Prepare the file write status update callback to update spinner
    const onStatusUpdate = (fileDetail: {
      modelKey: string;
      filename: string;
      status: string;
      error?: string;
    }, _allDetails: Array<{ modelKey: string; filename: string; status: string; error?: string }>): void => {
      if (fileDetail.status === 'pending') {
        spinner.text = `Writing file for ${fileDetail.modelKey}...`;
      } else if (fileDetail.status === 'success') {
        spinner.text = `Wrote results for ${fileDetail.modelKey}`;
      } else if (fileDetail.status === 'error') {
        spinner.text = `Error writing file for ${fileDetail.modelKey}: ${fileDetail.error || 'Unknown error'}`;
      }
    };
    
    // 3. Process output using the outputHandler
    const outputOptions = {
      outputDirectory: outputDirectoryPath,
      directoryIdentifier: undefined, // Already incorporated in the path
      friendlyRunName,
      includeMetadata: options.includeMetadata,
      useColors: true,
      includeThinking: options.enableThinking,
      throwOnError: false, // Handle errors locally
      onStatusUpdate
    };
    
    // Extract responses from query results for output processing
    const responses = queryResults.responses.map(response => ({
      ...response,
      configKey: `${response.provider}:${response.modelId}`
    }));
    
    // 4. Process output (both file writing and console formatting)
    const outputResult = await processOutput(responses, outputOptions);
    
    // 5. Update spinner with success/warning information based on results
    const { succeededWrites, failedWrites } = outputResult.fileOutput;
    const totalFiles = succeededWrites + failedWrites;
    
    if (failedWrites === 0) {
      // All files wrote successfully
      spinner.text = `Output processed successfully: ${succeededWrites} ${succeededWrites === 1 ? 'file' : 'files'} written`;
      
      // Display success message with output directory
      spinner.info(styleInfo(`Results saved to: ${styleSuccess(outputDirectoryPath)}`));
      spinner.start(); // Restart spinner after info message
    } else if (succeededWrites > 0) {
      // Partial success
      spinner.text = `Output processed with some failures`;
      spinner.warn(styleWarning(`Some files failed to write (${failedWrites} of ${totalFiles})`));
      
      // Show success and output directory
      spinner.info(styleInfo(`${succeededWrites} ${succeededWrites === 1 ? 'file' : 'files'} saved to: ${outputDirectoryPath}`));
      spinner.start(); // Restart spinner after info message
    } else {
      // All files failed
      spinner.text = `Failed to write any output files`;
      spinner.warn(styleWarning(`All ${failedWrites} ${failedWrites === 1 ? 'file' : 'files'} failed to write`));
      
      // Show errors summary
      const errorMessages = outputResult.fileOutput.files
        .filter(f => f.error)
        .map(f => `${f.modelKey}: ${f.error}`);
        
      if (errorMessages.length > 0) {
        spinner.info(styleInfo(`Errors: ${errorMessages.join(', ')}`));
        spinner.start(); // Restart spinner after info message
      }
    }
    
    // Set final spinner text for consistent state
    spinner.text = `Output processing complete: ${succeededWrites} files written, ${failedWrites} failed`;
    
    // 6. Return the result
    return {
      fileOutputResult: outputResult.fileOutput,
      consoleOutput: outputResult.consoleOutput
    };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's already a FileSystemError or PermissionError, just rethrow it
    if (error instanceof FileSystemError || error instanceof PermissionError) {
      throw error;
    }
    
    // If it's an OutputHandlerError, convert to FileSystemError
    if (error instanceof OutputHandlerError) {
      throw new FileSystemError(error.message, {
        cause: error,
        suggestions: [
          'Check that you have write permissions for the output directory',
          'Verify that there is sufficient disk space',
          'Try using a different output directory with --output'
        ]
      });
    }
    
    // Handle NodeJS.ErrnoException for file system errors
    if (
      error instanceof Error && 
      'code' in error && 
      typeof (error as NodeJS.ErrnoException).code === 'string'
    ) {
      const nodeError = error as NodeJS.ErrnoException;
      
      // Permission errors
      if (nodeError.code === 'EACCES' || nodeError.code === 'EPERM') {
        throw new PermissionError(`Permission denied when writing output: ${error.message} (path: ${outputDirectoryPath})`, {
          cause: error,
          suggestions: [
            'Check that you have write permissions for the output directory',
            'Try specifying a different output directory with --output'
          ]
        });
      }
      
      // Directory or file not found
      if (nodeError.code === 'ENOENT') {
        throw new FileSystemError(`Directory not found: ${outputDirectoryPath}`, {
          cause: error,
          suggestions: [
            'The output directory may have been deleted during processing',
            'Try specifying a different output directory with --output'
          ]
        });
      }
      
      // Disk space issues
      if (nodeError.code === 'ENOSPC') {
        throw new FileSystemError(`No space left on device when writing output files`, {
          cause: error,
          suggestions: [
            'Free up disk space',
            'Try specifying a different output directory on a drive with more space'
          ]
        });
      }
      
      // Other file system errors
      throw new FileSystemError(`File system error while writing output: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check directory permissions and path',
          'Verify that the disk is not full or write-protected'
        ]
      });
    }
    
    // For unexpected errors, wrap in ThinktankError
    throw new ThinktankError(`Error processing output: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      category: errorCategories.FILESYSTEM,
      suggestions: [
        'This is an unexpected error',
        'Try a different output directory',
        'Check system resources and permissions'
      ]
    });
  }
}

/**
 * Completion summary helper function
 * 
 * Formats and logs the completion summary, handling both success and partial failure scenarios.
 * 
 * @param params - Parameters containing query results, file results, options, and output path
 * @returns Empty result object since this function primarily produces console output
 */
export function _logCompletionSummary({
  queryResults,
  fileOutputResult,
  options,
  outputDirectoryPath
}: LogCompletionSummaryParams): LogCompletionSummaryResult {
  // Extract important data for the summary
  const { responses, statuses } = queryResults;
  
  // Count successes and failures
  const successCount = Object.values(statuses).filter(s => s.status === 'success').length;
  const errorCount = Object.values(statuses).filter(s => s.status === 'error').length;
  
  // Create mode-specific completion message based on options
  let completionMessage = '';
  if (options.specificModel) {
    completionMessage = options.specificModel;
  } else if (options.groupName) {
    completionMessage = `${options.groupName} group (${responses.length} model${responses.length === 1 ? '' : 's'})`;
  } else {
    completionMessage = `${responses.length} model${responses.length === 1 ? '' : 's'}`;
  }
  
  // Add the run name if available
  if (options.friendlyRunName) {
    completionMessage = `'${options.friendlyRunName}' (${completionMessage})`;
  }
  
  // Format completion time
  const elapsedTime = queryResults.timing.durationMs;
  const completionTimeText = elapsedTime > 1000 
    ? `${(elapsedTime / 1000).toFixed(2)}s` 
    : `${elapsedTime}ms`;
  
  // Calculate success percentage
  const percentage = successCount > 0 
    ? Math.round((successCount / responses.length) * 100) 
    : 0;
  
  // Create the summary text
  let summaryOutput = '';
  
  // Add file output summary
  const { succeededWrites, failedWrites } = fileOutputResult;
  const totalFiles = succeededWrites + failedWrites;
  
  if (failedWrites > 0) {
    const fileStatusText = failedWrites === totalFiles 
      ? 'No output files were written'
      : `${succeededWrites} of ${totalFiles} output files were written`;
    
    summaryOutput += styleWarning(`${fileStatusText} to: ${outputDirectoryPath}\n`);
  } else {
    summaryOutput += styleSuccess(`Output saved to: ${outputDirectoryPath}\n`);
  }
  
  // Model results summary section
  if (errorCount > 0) {
    // Format partial success message
    summaryOutput += styleWarning(
      `Processing complete for ${completionMessage} - ${successCount} of ${responses.length} models completed successfully (${percentage}%)\n`
    );
    
    // Group errors by category for better display
    const errorsByCategory: Record<string, Array<{ model: string, message: string }>> = {};
    
    Object.entries(statuses)
      .filter(([_, status]) => status.status === 'error')
      .forEach(([model, status]) => {
        // Determine error category
        let category = errorCategories.UNKNOWN;
        const message = status.message || 'Unknown error';
        
        // Try to extract category from the error message or error object
        if (status.detailedError && 'category' in status.detailedError) {
          category = (status.detailedError as { category?: string }).category || errorCategories.UNKNOWN;
        } else {
          // Try to extract from message
          Object.values(errorCategories).forEach(cat => {
            if (message.includes(cat)) {
              category = cat;
            }
          });
        }
        
        if (!errorsByCategory[category]) {
          errorsByCategory[category] = [];
        }
        
        errorsByCategory[category].push({ 
          model, 
          message
        });
      });
    
    // Create a nice tree-style summary
    summaryOutput += `\n${colors.blue('Results Summary:')}\n`;
    summaryOutput += `${styleDim('│')}\n`;
    
    // First show successful models
    if (successCount > 0) {
      summaryOutput += `${styleDim('├')} ${colors.green('+')} Successful Models (${successCount}):\n`;
      const successModels = Object.entries(statuses)
        .filter(([_, status]) => status.status === 'success')
        .map(([model]) => model);
        
      successModels.forEach((model, i) => {
        const isLast = i === successModels.length - 1;
        const prefix = isLast ? `${styleDim('│  └')}` : `${styleDim('│  ├')}`;
        summaryOutput += `${prefix} ${model}\n`;
      });
    }
    
    // Then show failed models by category
    summaryOutput += `${styleDim('├')} ${colors.red('x')} Failed Models (${errorCount}):\n`;
    
    // Display errors by category
    Object.entries(errorsByCategory).forEach(([category, errors], categoryIndex, categories) => {
      const isLastCategory = categoryIndex === categories.length - 1;
      const categoryPrefix = isLastCategory ? `${styleDim('│  └')}` : `${styleDim('│  ├')}`;
      
      summaryOutput += `${categoryPrefix} ${colors.yellow(category)} errors (${errors.length}):\n`;
      
      errors.forEach(({ model, message }, errorIndex) => {
        const isLastError = errorIndex === errors.length - 1;
        const errorPrefix = isLastCategory 
          ? (isLastError ? `${styleDim('│     └')}` : `${styleDim('│     ├')}`)
          : (isLastError ? `${styleDim('│  │  └')}` : `${styleDim('│  │  ├')}`);
          
        summaryOutput += `${errorPrefix} ${colors.red(model)}\n`;
        
        // Add indented error message
        const messagePrefix = isLastCategory
          ? (isLastError ? `${styleDim('│      ')}` : `${styleDim('│     │')}`)
          : (isLastError ? `${styleDim('│  │   ')}` : `${styleDim('│  │  │')}`);
          
        summaryOutput += `${messagePrefix} ${styleDim('→')} ${message}\n`;
      });
    });
    
    summaryOutput += `${styleDim('└')} Completed in ${completionTimeText}\n`;
  } else {
    // Format complete success message
    summaryOutput += styleSuccess(`Successfully completed ${completionMessage} in ${completionTimeText}\n`);
    
    // Display a nice tree-style summary for successful models
    summaryOutput += `\n${colors.blue('Results Summary:')}\n`;
    summaryOutput += `${styleDim('│')}\n`;
    
    const successModels = Object.keys(statuses);
    successModels.forEach((model, i) => {
      const isLast = i === successModels.length - 1;
      const prefix = isLast ? `${styleDim('└')}` : `${styleDim('├')}`;
      summaryOutput += `${prefix} ${i+1}. ${model} - ${colors.green('+')} Success\n`;
    });
    
    if (successModels.length === 0) {
      summaryOutput += `${styleDim('└')} No models were queried.\n`;
    }
  }
  
  // Log the summary to the console
  console.log(summaryOutput);
  
  // Return empty object since this function primarily produces console output
  return {};
}

/**
 * Error handling helper function
 * 
 * Categorizes unknown errors, ensures proper ThinktankError types, logs contextual
 * information, and rethrows for upstream handling.
 * 
 * @param params - Parameters containing error, spinner, options, and workflow state
 * @returns Never returns normally, always throws an error
 * @throws ThinktankError or one of its specialized subclasses
 */
export function _handleWorkflowError({
  error,
  spinner,
  options,
  workflowState
}: HandleWorkflowErrorParams): never {
  // Type assertions to help TypeScript understand types better
  const typedSpinner = spinner as { 
    fail: (message: string) => void;
    warn: (message: string) => void;
    info: (message: string) => void;
  };
  
  const typedOptions = options as {
    input?: string;
    output?: string;
    specificModel?: string;
    models?: string[];
  };
  
  const typedWorkflowState = workflowState as {
    outputDirectoryPath?: string;
    friendlyRunName?: string;
  };
  // If the error is already a ThinktankError, we may still want to add context
  if (error instanceof ThinktankError) {
    // Update the spinner with the error's formatted message
    typedSpinner.fail(error.format());
    
    // Rethrow the existing error
    throw error;
  }
  
  // Handle standard Error objects
  if (error instanceof Error) {
    // Extract valuable information from the error
    const errorMessage = error.message.toLowerCase();
    const errorStack = error.stack?.toLowerCase() || '';
    
    // Create context information from workflow state to use in error suggestions
    const contextInfo = {
      // Current working directory for file-related errors
      cwd: process.cwd(),
      // Input information for input-related errors
      input: typedOptions.input,
      // Output directory for file system errors
      outputDirectory: typedWorkflowState.outputDirectoryPath || typedOptions.output,
      // Model information for API and model-related errors
      specificModel: typedOptions.specificModel,
      modelsList: typedOptions.models,
      // Run name for general context
      runName: typedWorkflowState.friendlyRunName
    };
    
    // File not found errors
    if (
      (errorMessage.includes('enoent') || errorMessage.includes('file not found')) &&
      typeof typedOptions.input === 'string'
    ) {
      const fileSystemError = new FileSystemError(`File not found: ${typedOptions.input}`, {
        cause: error,
        suggestions: [
          `Check that the file exists at the specified path: ${typedOptions.input}`,
          `Current working directory: ${contextInfo.cwd}`
        ]
      });
      
      // Update spinner and throw the specialized error
      typedSpinner.fail(fileSystemError.format());
      throw fileSystemError;
    }
    
    // Permission errors
    if (
      errorMessage.includes('eacces') || 
      errorMessage.includes('eperm') || 
      errorMessage.includes('permission denied') ||
      errorMessage.includes('access denied')
    ) {
      // Determine if it's related to output directory
      const isOutputDir = 
        (errorStack.includes('directory') || errorStack.includes('mkdir')) && 
        contextInfo.outputDirectory;
      
      let message = '';
      let suggestions: string[] = [];
      
      if (isOutputDir) {
        message = `Permission denied when creating output directory: ${contextInfo.outputDirectory}`;
        suggestions = [
          'Check that you have write permissions for the directory',
          'Try specifying a different output directory with --output'
        ];
      } else if (errorMessage.includes('input') || (errorStack.includes('read') && contextInfo.input)) {
        message = `Permission denied when reading input: ${contextInfo.input}`;
        suggestions = [
          'Check that you have read permissions for the file',
          'Try using a different input source'
        ];
      } else {
        message = `Permission denied: ${error.message}`;
        suggestions = [
          'Check file and directory permissions',
          'Try using a different location or running with elevated privileges if appropriate'
        ];
      }
      
      const permissionError = new PermissionError(message, {
        cause: error,
        suggestions
      });
      
      // Update spinner and throw the specialized error
      typedSpinner.fail(permissionError.format());
      throw permissionError;
    }
    
    // Model-related errors
    if (errorMessage.includes('model')) {
      if (errorMessage.includes('format') || errorMessage.includes('invalid')) {
        // Extract the model specification if possible
        const modelMatch = errorMessage.match(/"([^"]+)"/);
        const modelSpec = modelMatch ? modelMatch[1] : contextInfo.specificModel || 
                         (contextInfo.modelsList && contextInfo.modelsList.length > 0 ? contextInfo.modelsList[0] : 'unknown');
        
        const configError = new ConfigError(`Invalid model format: ${modelSpec}`, {
          cause: error,
          suggestions: [
            'Model specifications must use the format "provider:modelId" (e.g., "openai:gpt-4o")',
            'Check that the model is correctly spelled'
          ],
          examples: [
            'openai:gpt-4o',
            'anthropic:claude-3-7-sonnet-20250219',
            'google:gemini-pro'
          ]
        });
        
        typedSpinner.fail(configError.format());
        throw configError;
      } else if (errorMessage.includes('not found')) {
        // Extract the model specification if possible
        const modelMatch = errorMessage.match(/"([^"]+)"/);
        const modelSpec = modelMatch ? modelMatch[1] : contextInfo.specificModel || 
                         (contextInfo.modelsList && contextInfo.modelsList.length > 0 ? contextInfo.modelsList[0] : 'unknown');
        
        const configError = new ConfigError(`Model "${modelSpec}" not found in configuration`, {
          cause: error,
          suggestions: [
            'Check that the model is correctly spelled and exists in your configuration',
            'Use "thinktank models" to list all available models'
          ]
        });
        
        typedSpinner.fail(configError.format());
        throw configError;
      }
    }
    
    // API key and authentication errors
    if (
      errorMessage.includes('api key') || 
      errorMessage.includes('authentication') ||
      errorMessage.includes('authorization') ||
      errorMessage.includes('auth') ||
      errorMessage.includes('credentials') ||
      errorMessage.includes('unauthorized') ||
      errorMessage.includes('401')
    ) {
      const apiError = new ApiError(`API key error: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check that you have set the correct environment variables for your API keys',
          'You can set them in your .env file or in your environment',
          'Verify that your API keys are valid and have not expired'
        ]
      });
      
      typedSpinner.fail(apiError.format());
      throw apiError;
    }
    
    // Network errors
    if (
      errorMessage.includes('network') ||
      errorMessage.includes('connect') ||
      errorMessage.includes('timeout') ||
      errorMessage.includes('econnrefused') ||
      errorMessage.includes('etimedout') ||
      errorMessage.includes('socket') ||
      errorMessage.includes('dns')
    ) {
      const networkError = new NetworkError(`Network error: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check your internet connection',
          'Verify that the service endpoints are accessible',
          'Try again later if the service might be experiencing downtime'
        ]
      });
      
      typedSpinner.fail(networkError.format());
      throw networkError;
    }
    
    // For all other Error objects, use categorization utility
    const category = categorizeError(error);
    let thinktankError: ThinktankError;
    
    // Create the appropriate error type based on the category
    switch (category) {
      case errorCategories.API:
        thinktankError = new ApiError(`API error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.CONFIG:
        thinktankError = new ConfigError(`Configuration error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.FILESYSTEM:
        thinktankError = new FileSystemError(`File system error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.PERMISSION:
        thinktankError = new PermissionError(`Permission error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.VALIDATION:
        thinktankError = new ValidationError(`Validation error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.INPUT:
        // Use ThinktankError with INPUT category since InputError is already defined elsewhere
        thinktankError = new ThinktankError(`Input error: ${error.message}`, { 
          cause: error,
          category: errorCategories.INPUT
        });
        break;
        
      case errorCategories.NETWORK:
        thinktankError = new NetworkError(`Network error: ${error.message}`, { cause: error });
        break;
        
      case errorCategories.UNKNOWN:
      default:
        // For unknown categories, create a generic ThinktankError
        thinktankError = new ThinktankError(`Error running thinktank: ${error.message}`, {
          cause: error,
          category: errorCategories.UNKNOWN
        });
        break;
    }
    
    // Add workflow state context to the error's suggestions
    const suggestions = thinktankError.suggestions || [];
    
    // Add context-specific suggestions based on the workflow state
    if (typedWorkflowState.friendlyRunName) {
      suggestions.push(`This error occurred during run: ${typedWorkflowState.friendlyRunName}`);
    }
    
    if (category === errorCategories.FILESYSTEM && typedWorkflowState.outputDirectoryPath) {
      suggestions.push(`Output directory: ${typedWorkflowState.outputDirectoryPath}`);
    }
    
    // Update the error with the enhanced suggestions
    thinktankError.suggestions = suggestions;
    
    // Update spinner and throw the categorized error
    typedSpinner.fail(thinktankError.format());
    throw thinktankError;
  }
  
  // Handle non-Error objects (like string messages, numbers, or other values)
  const unknownErrorMessage = error !== null && error !== undefined 
    ? String(error) 
    : 'Unknown error (no error information provided)';
    
  const thinktankError = new ThinktankError(`Unknown error running thinktank: ${unknownErrorMessage}`, {
    category: errorCategories.UNKNOWN,
    suggestions: [
      'This is an unexpected error type',
      'Please report this issue with steps to reproduce'
    ]
  });
  
  // Update spinner with the error message
  typedSpinner.fail(thinktankError.format());
  
  // Throw the error for upstream handling
  throw thinktankError;
}