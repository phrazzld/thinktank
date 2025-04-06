/**
 * Helper functions for the runThinktank workflow
 * 
 * This file contains the implementation of helper functions that encapsulate
 * distinct phases of the runThinktank workflow, making the main function more
 * modular and easier to maintain.
 */
import { loadConfig } from '../core/configManager';
import { generateFunName } from '../utils/nameGenerator';
import { createOutputDirectory } from './outputHandler';
import { 
  ThinktankError,
  ConfigError,
  ApiError,
  FileSystemError,
  PermissionError,
  errorCategories
} from '../core/errors';
import { styleInfo, styleSuccess, styleWarning } from '../utils/consoleUtils';
import {
  SetupWorkflowParams,
  SetupWorkflowResult,
  ProcessInputParams,
  ProcessInputResult,
  SelectModelsParams,
  SelectModelsResult,
  ExecuteQueriesParams,
  ExecuteQueriesResult
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
    } else if (successCount > 0) {
      // Partial success
      spinner.text = `Queries completed with some failures`;
      spinner.warn(styleWarning(`Some models failed to respond (${failureCount} of ${models.length})`));
      
      // Display counts of successes and failures
      spinner.info(styleInfo(`Results: ${successCount} succeeded, ${failureCount} failed`));
    } else {
      // All models failed
      spinner.text = `All queries failed`;
      spinner.warn(styleWarning(`All models failed to respond (${failureCount} ${failureCount === 1 ? 'model' : 'models'})`));
      
      // Display failure details
      const errorSummary = queryResults.responses
        .map(r => `${r.configKey}: ${r.error}`)
        .join(', ');
        
      spinner.info(styleInfo(`Errors: ${errorSummary}`));
    }
    
    // Restart spinner for next step
    spinner.start();
    
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