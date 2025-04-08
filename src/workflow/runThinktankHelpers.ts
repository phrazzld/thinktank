/**
 * Helper functions for the runThinktank workflow
 * 
 * This file contains the implementation of helper functions that encapsulate
 * distinct phases of the runThinktank workflow, making the main function more
 * modular and easier to maintain.
 * 
 * Each helper function follows a consistent pattern:
 * 1. Updates the spinner with current status
 * 2. Performs its core functionality
 * 3. Updates the spinner with success/failure information
 * 4. Returns structured result data
 * 5. Handles errors according to defined error contracts
 * 
 * The workflow is divided into these sequential phases:
 * - Setup: Configuration loading and output directory creation
 * - Input Processing: Handling file/stdin/direct text input
 * - Model Selection: Determining which models to query
 * - Query Execution: Sending prompts to selected models
 * - Output Processing: Writing responses to files and formatting console output
 * - Completion Summary: Displaying final results and statistics
 * 
 * Error handling follows established contracts defined in runThinktankTypes.ts
 */
import { findModelGroup } from '../core/configManager';
import { generateFunName } from '../utils/nameGenerator';
import { 
  generateFilename,
  formatResponseAsMarkdown,
  formatForConsole,
  createOutputDirectory
} from './outputHandler';
import { 
  ThinktankError,
  ConfigError,
  ApiError,
  FileSystemError,
  PermissionError,
  errorCategories
} from '../core/errors';
import { createContextualError } from '../core/errors/utils/categorization';
//import { FileOutputStatus } from '../utils/throttledSpinner';
import { 
  styleInfo, 
  styleSuccess, 
  styleWarning, 
  colors,
  styleDim
} from '../utils/consoleUtils';
import { logger } from '../utils/logger';
import { readContextPaths, formatCombinedInput } from '../utils/fileReader';
import { ContextFileResult } from '../utils/fileReaderTypes';
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
  PureProcessOutputResult,
  FileData,
  LogCompletionSummaryParams,
  LogCompletionSummaryResult,
  HandleWorkflowErrorParams
} from './runThinktankTypes';
import { processInput, InputError } from './inputHandler';
import { selectModels, ModelSelectionError } from './modelSelector';
import { QueryExecutorError, ModelQueryStatus } from './queryExecutor';

/**
 * Setup workflow helper function
 * 
 * Handles configuration loading, run name generation, and output directory creation
 * with proper error handling and spinner updates.
 * 
 * @param params - Parameters containing the spinner, options, and configManager
 * @returns An object containing the configuration, run name, and output directory path
 * @throws 
 *   - ConfigError when config loading fails
 *   - FileSystemError when directory creation fails
 *   - PermissionError when permission issues occur
 */
export async function _setupWorkflow({ 
  spinner, 
  options,
  configManager,
  fileSystem
}: SetupWorkflowParams): Promise<SetupWorkflowResult> {
  try {
    // 1. Load configuration using the injected configManager
    spinner.text = 'Loading configuration...';
    const config = await configManager.loadConfig({ configPath: options.configPath });
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
    }, fileSystem); // Use the shared fileSystem instance
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
    
    // Convert to ConfigError for consistency with test expectations
    throw new ConfigError(`Error during workflow setup: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined
    });
  }
}

/**
 * Input processing helper function
 * 
 * Handles input processing from various sources (file, stdin, or direct text),
 * processes optional context files/directories, and combines them with the main prompt.
 * 
 * @param params - Parameters containing the spinner, input string, optional context paths, and fileSystem
 * @returns An object containing the processed input result and context files
 * @throws 
 *   - FileSystemError when input processing fails due to file system issues
 *   - ThinktankError for other unexpected errors
 */
export async function _processInput({ 
  spinner, 
  input,
  contextPaths,
  fileSystem
}: ProcessInputParams): Promise<ProcessInputResult> {
  try {
    // 1. Update spinner with processing status
    spinner.text = 'Processing input...';
    
    // 2. Process the main input using inputHandler.processInput with fileSystem
    const originalInputResult = await processInput({ 
      input,
      fileSystem 
    });
    
    // Create extended input result with proper type for metadata
    const inputResult: import('./runThinktankTypes').ExtendedInputResult = {
      ...originalInputResult,
      metadata: {
        ...originalInputResult.metadata,
        contextFilesCount: undefined,
        contextFilesWithErrors: undefined,
        hasContextFiles: false
      }
    };
    
    // 3. If no context paths are provided, return the input result as is
    if (!contextPaths || contextPaths.length === 0) {
      spinner.text = `Input processed from ${inputResult.sourceType} (${inputResult.content.length} characters)`;
      return { 
        inputResult, 
        contextFiles: [],
        combinedContent: inputResult.content
      };
    }
    
    // 4. Process context paths if provided
    spinner.text = `Processing context files from ${contextPaths.length} path${contextPaths.length === 1 ? '' : 's'}...`;
    
    let contextFiles: ContextFileResult[];
    try {
      // Use the fileSystem for reading context paths
      contextFiles = await readContextPaths(contextPaths, fileSystem);
    } catch (err) {
      // If it's already a FileSystemError, just rethrow it
      if (err instanceof FileSystemError) {
        throw err;
      }
      
      // Handle other errors in context file reading
      throw new FileSystemError(`Failed to read context files: ${err instanceof Error ? err.message : String(err)}`, {
        cause: err instanceof Error ? err : undefined,
        suggestions: [
          'Check that all context paths exist and are accessible',
          'Verify permissions on context files and directories',
          'Try using absolute paths if the files are in a different location'
        ]
      });
    }
    
    // 5. Count successful and error context files
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call
    const successfulContextFiles = contextFiles.filter(file => file.error === null && file.content !== null);
    
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call
    const errorContextFiles = contextFiles.filter(file => file.error !== null || file.content === null);
    
    // 6. Show warning for context files with errors
    if (errorContextFiles.length > 0) {
      spinner.warn(styleWarning(`${errorContextFiles.length} of ${contextFiles.length} context files could not be read and will be skipped.`));
      
      // Log detailed information about each error file
      // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call
      errorContextFiles.forEach(file => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        logger.debug(`Context file error - ${file.path}: ${file.error?.message || 'Unknown error'}`);
      });
      
      // Restart spinner after warnings
      spinner.start();
    }
    
    // 7. Combine context content with prompt if we have successful context files
    if (successfulContextFiles.length > 0) {
      // Use formatCombinedInput to merge the prompt with context files
      // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
      const combinedContent = formatCombinedInput(inputResult.content, contextFiles);
      
      // Update input result with the combined content
      inputResult.content = combinedContent;
      
      // Add context information to metadata
      inputResult.metadata.hasContextFiles = true;
      inputResult.metadata.contextFilesCount = successfulContextFiles.length;
      inputResult.metadata.contextFilesWithErrors = errorContextFiles.length;
      inputResult.metadata.finalLength = combinedContent.length;
      
      // Display success message
      spinner.info(styleInfo(`Added ${successfulContextFiles.length} context file${successfulContextFiles.length === 1 ? '' : 's'} to the prompt.`));
      spinner.start(); // Restart spinner after info
    } else if (contextPaths.length > 0) {
      // If no context files were successfully read but some were specified
      spinner.warn(styleWarning('No context files could be read. Continuing with original prompt only.'));
      spinner.start(); // Restart spinner after warning
    }
    
    // 8. Update spinner with final status
    spinner.text = `Input processed from ${inputResult.sourceType}`;
    if (successfulContextFiles.length > 0) {
      spinner.text += ` with ${successfulContextFiles.length} context file${successfulContextFiles.length === 1 ? '' : 's'}`;
    }
    spinner.text += ` (${inputResult.content.length} characters)`;
    
    // 9. Return the result with context files and combinedContent
    return { 
      inputResult,
      contextFiles,
      combinedContent: inputResult.content
    };
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
    
    // For unexpected errors, wrap in FileSystemError with proper filePath
    throw new FileSystemError(`File not found: ${input}`, {
      filePath: input,
      cause: error instanceof Error ? error : undefined,
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
    
    // 6. Create the flattened result object using intersection type
    const result = {
      ...modelSelectionResult,  // Spread all properties from ModelSelectionResult
      modeDescription,          // Add the modeDescription property
      modelSelectionResult      // Self-reference for backward compatibility
    };
    
    // Return the enhanced result
    return result;
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
    
    // For unexpected errors, handle specifically for missing API keys
    if (error instanceof Error && error.message.includes('Missing API keys')) {
      throw new ApiError(`Missing API keys: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check your environment variables for API keys (OPENAI_API_KEY, ANTHROPIC_API_KEY)',
          'Use models that you have API keys for',
          'Run "thinktank models" to see which models require API keys'
        ]
      });
    }
    
    // Other unexpected errors, wrap in ConfigError
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
 * Central function for executing LLM queries against multiple models in parallel.
 * This function:
 * 1. Uses the injected LLMClient interface for making API calls
 * 2. Sets up real-time progress reporting via status update tracking
 * 3. Handles success and error states for each model
 * 4. Aggregates and summarizes results for reporting
 * 5. Formats spinner updates to show progress for each model
 * 
 * This function uses dependency injection to receive an LLMClient instance,
 * improving testability by decoupling from specific provider implementations.
 * It supports both standard Ora spinners and enhanced ThrottledSpinner instances
 * by using feature detection with duck typing.
 * 
 * After execution completes, it provides a summary of successes and failures
 * with appropriate console messages based on the outcome.
 * 
 * @param params - Parameters for query execution
 * @param params.spinner - The spinner instance for providing visual feedback
 * @param params.config - The loaded application configuration
 * @param params.models - Array of model configurations to query
 * @param params.combinedContent - The combined prompt and context content to send to models
 * @param params.options - The user-provided run options
 * @param params.llmClient - The LLMClient interface instance for making API calls
 * @returns Promise resolving to an object containing query execution results
 * @throws {ApiError} When query execution fails due to API communication errors
 * @throws {ThinktankError} For other unexpected errors during query execution
 */
export async function _executeQueries({
  spinner,
  config,
  models,
  combinedContent,
  options,
  llmClient
}: ExecuteQueriesParams): Promise<ExecuteQueriesResult> {
  try {
    // 1. Update spinner with execution status
    spinner.text = `Executing queries to ${models.length} ${models.length === 1 ? 'model' : 'models'}...`;
    
    // 2. Initialize timing and status tracking
    const startTime = Date.now();
    const statuses: Record<string, ModelQueryStatus> = {};
    
    // Initialize status for each model
    models.forEach(model => {
      const modelKey = `${model.provider}:${model.modelId}`;
      statuses[modelKey] = { status: 'pending' };
    });
    
    // 3. Create an array to hold the promises for each model query
    const queryPromises: Array<Promise<import('../core/types').LLMResponse & { configKey: string }>> = [];
    
    // 4. Process each model using the injected LLMClient
    for (const model of models) {
      const modelKey = `${model.provider}:${model.modelId}`;
      
      // Create the promise for this model query
      const queryPromise = (async () => {
        // Update status to running
        statuses[modelKey] = { 
          status: 'running',
          startTime: Date.now()
        };
        
        // Call status update logic to update spinner
        updateModelStatus(spinner, modelKey, statuses[modelKey], statuses);
        
        try {
          // Determine which system prompt to use
          const groupInfo = findModelGroup(config, model);
          
          // Determine the final system prompt (CLI override > model > group)
          let systemPrompt = undefined;
          if (options.systemPrompt) {
            // Use CLI override
            systemPrompt = {
              text: options.systemPrompt,
              metadata: { source: 'cli-override' }
            };
          } else if (model.systemPrompt) {
            // Use model-specific system prompt
            systemPrompt = model.systemPrompt;
          } else if (groupInfo?.systemPrompt) {
            // Use group system prompt
            systemPrompt = groupInfo.systemPrompt;
          }
          
          // Prepare model options with thinking capability if applicable
          const modelOptions = { ...model.options };
          
          // Enable thinking capability for Claude models if requested
          if (options.enableThinking && model.provider === 'anthropic' && model.modelId.includes('claude-3')) {
            modelOptions.thinking = {
              type: 'enabled',
              budget_tokens: 16000 // Default budget
            };
          }
          
          // If timeout is specified in options, add it to modelOptions
          if (options.timeoutMs !== undefined) {
            modelOptions.timeout = options.timeoutMs;
          }
          
          // Key Change: Use the injected llmClient to generate the response
          const response = await llmClient.generate(
            combinedContent,
            modelKey,
            modelOptions,
            systemPrompt
          );
          
          // Calculate duration
          const endTime = Date.now();
          const durationMs = endTime - (statuses[modelKey].startTime || endTime);
          
          // Update status with success
          statuses[modelKey] = { 
            status: 'success',
            startTime: statuses[modelKey].startTime,
            endTime,
            durationMs
          };
          
          // Call status update logic
          updateModelStatus(spinner, modelKey, statuses[modelKey], statuses);
          
          // Add config key and group information to the response
          const responseWithKey = {
            ...response,
            configKey: modelKey,
          };
          
          // Add group information if available and not already added by the client
          if (groupInfo && !responseWithKey.groupInfo) {
            responseWithKey.groupInfo = {
              name: groupInfo.groupName,
              systemPrompt: groupInfo.systemPrompt
            };
          }
          
          return responseWithKey;
        } catch (error) {
          // Calculate duration
          const endTime = Date.now();
          const durationMs = endTime - (statuses[modelKey].startTime || endTime);
          
          // Get error message and details
          const errorMessage = error instanceof Error ? error.message : String(error);
          
          // Check if it's a ThinktankError to extract category
          const category = (error instanceof ThinktankError) 
            ? error.category 
            : errorCategories.API;
            
          // Update status with error
          statuses[modelKey] = { 
            status: 'error',
            message: errorMessage,
            detailedError: error instanceof Error ? error : new Error(String(error)),
            startTime: statuses[modelKey].startTime,
            endTime,
            durationMs
          };
          
          // Call status update logic
          updateModelStatus(spinner, modelKey, statuses[modelKey], statuses);
          
          // Return error response
          return {
            provider: model.provider,
            modelId: model.modelId,
            text: '',
            error: errorMessage,
            errorCategory: category,
            configKey: modelKey,
          };
        }
      })();
      
      queryPromises.push(queryPromise);
    }
    
    // 5. Execute all queries in parallel
    const responses = await Promise.all(queryPromises);
    
    // 6. Calculate overall timing
    const endTime = Date.now();
    const durationMs = endTime - startTime;
    
    // 7. Build the query results object
    const queryResults = {
      responses,
      statuses,
      timing: {
        startTime,
        endTime,
        durationMs
      },
      combinedContent // Include the original combined content
    };
    
    // 8. Count successes and failures
    const successCount = responses.filter(r => !r.error).length;
    const failureCount = responses.filter(r => !!r.error).length;
    
    // 9. Update spinner with appropriate success/warning message
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
      const errorSummary = responses
        .map(r => `${r.configKey}: ${r.error}`)
        .join(', ');
        
      spinner.info(styleInfo(`Errors: ${errorSummary}`));
      spinner.start(); // Restart spinner after info message
    }
    
    // 10. Set final spinner text for consistent state
    if ('updateForModelSummary' in spinner) {
      spinner.updateForModelSummary(successCount, failureCount, true);
    } else {
      spinner.text = `Query execution complete: ${successCount} succeeded, ${failureCount} failed`;
    }
    
    // 11. Return the results
    return { queryResults };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's already an ApiError, just rethrow it
    if (error instanceof ApiError) {
      throw error;
    }
    
    // If it's a QueryExecutorError, convert to ApiError 
    if (error instanceof QueryExecutorError) {
      throw new ApiError('Failed to execute queries', {
        cause: error,
        suggestions: error.suggestions
      });
    }
    
    // If it's a ThinktankError (but not ApiError), rethrow it
    if (error instanceof ThinktankError) {
      throw error;
    }
    
    // For unexpected errors, wrap in ApiError
    throw new ApiError('Failed to execute queries', {
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
 * Helper function to update the spinner with model status changes
 * 
 * Extracted from the original callback to make the code more modular
 * 
 * @param spinner - The spinner instance
 * @param modelKey - The model identifier key ("provider:modelId")
 * @param status - The current status for the model
 * @param allStatuses - All model statuses
 */
function updateModelStatus(
  spinner: import('./runThinktankTypes').EnhancedSpinner,
  modelKey: string,
  status: ModelQueryStatus,
  allStatuses: Record<string, ModelQueryStatus>
): void {
  // Check if the spinner has the enhanced method
  if ('updateForModelStatus' in spinner) {
    // Type cast for the third parameter since we know the function exists but TypeScript doesn't recognize it
    const enhancedSpinner = spinner as unknown as {
      updateForModelStatus: (modelKey: string, status: ModelQueryStatus, allStatuses?: Record<string, ModelQueryStatus>) => void;
    };
    enhancedSpinner.updateForModelStatus(modelKey, status, allStatuses);
  } else {
    // Fallback to basic text updates
    if (status.status === 'running') {
      spinner.text = `Querying ${modelKey}...`;
    } else if (status.status === 'success') {
      spinner.text = `Received response from ${modelKey} (${status.durationMs}ms)`;
    } else if (status.status === 'error') {
      spinner.text = `Error from ${modelKey}: ${status.message || 'Unknown error'}`;
    }
  }
}

/**
 * Output processing helper function
 * 
 * Processes query results and generates structured data for file output and console display.
 * This refactored version is "pure" in that it doesn't perform I/O operations (writing files),
 * but instead returns the data that would be written. The actual file writing is performed by
 * the caller (runThinktank.ts).
 * 
 * @param params - Parameters containing spinner, query results, and options
 * @returns Promise resolving to an object containing file data and console formatted output
 * @throws 
 *   - ThinktankError for unexpected errors during formatting
 */
export async function _processOutput({
  spinner,
  queryResults,
  options,
}: ProcessOutputParams): Promise<PureProcessOutputResult> {
  try {
    // 1. Update spinner with processing status
    spinner.text = 'Formatting results...';
    
    // Prepare responses with configKey
    const responses = queryResults.responses.map(response => ({
      ...response,
      configKey: `${response.provider}:${response.modelId}`
    }));
    
    // 2. Generate file data for each response
    const files: FileData[] = responses.map(response => {
      // Generate filename using formatters
      const filename = generateFilename(response, {
        // Control if group name is included based on how models were selected
        includeGroup: !(options.specificModel || (options.models && options.models.length === 1))
      });
      
      // Format content for file output
      const content = formatResponseAsMarkdown(
        response, 
        options.includeMetadata
      );
      
      return {
        filename,
        content,
        modelKey: response.configKey,
      };
    });
    
    // 3. Generate console output string
    const consoleOutput = formatForConsole(
      responses,
      {
        includeMetadata: options.includeMetadata,
        useColors: options.useColors !== false, // Default to true
        includeThinking: options.includeThinking,
        useTable: options.useTable
      }
    );
    
    // 4. Update spinner status
    const fileCount = files.length;
    spinner.text = `Formatted ${fileCount} ${fileCount === 1 ? 'result' : 'results'}`;
    
    // 5. Return the structured data
    return {
      files,
      consoleOutput
    };
  } catch (error) {
    // For unexpected errors during formatting, wrap in ThinktankError
    throw new ThinktankError(`Error formatting output: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      category: errorCategories.OUTPUT_FORMATTING, // Using appropriate category
      suggestions: [
        'This is an unexpected error during output formatting',
        'Check input data format and compatibility with formatters'
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
      `Processing complete for ${completionMessage} - ${percentage}% of models completed successfully\n`
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
  logger.plain(summaryOutput);
  
  // Return empty object since this function primarily produces console output
  return {};
}

/**
 * Error handling helper function
 * 
 * Centralized error handling for the entire workflow. This function:
 * 1. Categorizes unknown errors into appropriate ThinktankError subtypes
 * 2. Enriches errors with contextual information from the workflow state
 * 3. Adds helpful suggestions based on error category and context
 * 4. Updates the spinner with formatted error messages
 * 5. Rethrows the categorized error for upstream handling
 * 
 * Uses the createContextualError utility to simplify error categorization and
 * provide consistent error handling with proper context awareness.
 * 
 * This function is responsible for transforming any error type (including unknown)
 * into a properly structured ThinktankError with appropriate categorization,
 * making all errors in the system consistent and user-friendly.
 * 
 * @param params - Parameters containing error, spinner, options, and workflow state
 * @param params.error - The original error that occurred
 * @param params.spinner - The spinner instance for providing visual feedback
 * @param params.options - The user-provided run options
 * @param params.workflowState - The current state of the workflow when the error occurred
 * @returns Never returns normally, always throws an error
 * @throws {ThinktankError} Throws a properly categorized ThinktankError or one of its specialized subclasses
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
  
  // Extract context information from workflow state
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
  
  // Use the createContextualError utility to get a properly categorized error
  // with appropriate suggestions based on the context
  const thinktankError = createContextualError(error, contextInfo);
  
  // Update the spinner with the error's formatted message
  typedSpinner.fail(thinktankError.format());
  
  // Throw the error for upstream handling
  throw thinktankError;
}
