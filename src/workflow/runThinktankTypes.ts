/**
 * TypeScript interfaces for the runThinktank workflow helper functions
 * 
 * This file defines the input parameters, return types, and error handling contracts
 * for each helper function in the runThinktank workflow.
 * 
 * The workflow is organized as a sequence of helper functions that each handle
 * a specific phase of the overall process:
 * 
 * 1. Setup (_setupWorkflow): Loads configuration and creates output directory
 * 2. Input Processing (_processInput): Handles file, stdin, or direct text input
 * 3. Model Selection (_selectModels): Determines which models to query
 * 4. Query Execution (_executeQueries): Sends prompts to selected models
 * 5. Output Processing (_processOutput): Writes responses to files
 * 6. Completion Summary (_logCompletionSummary): Displays results and statistics
 * 
 * Each helper function has:
 * - Params interface: Defines what input is required for the function
 * - Result interface: Defines what output the function produces
 * - Error handling contract: Defines which errors the function handles and throws
 * 
 * The WorkflowState interface serves as a central container for passing data
 * between helper functions while maintaining type safety.
 */
import { RunOptions } from './runThinktank';
import { AppConfig, ModelConfig } from '../core/types';
import { ThinktankError } from '../core/errors';
import { InputResult } from './inputHandler';
import { ModelSelectionResult } from './modelSelector';
import { QueryExecutionResult } from './queryExecutor';
import { FileOutputResult } from './outputHandler';
import type { Ora } from 'ora';
import { ThrottledSpinner } from '../utils/throttledSpinner';
import { LLMClient, ConfigManagerInterface, FileSystem } from '../core/interfaces';

/**
 * Enhanced spinner type for use in the workflow
 * This allows either a regular Ora spinner or our enhanced ThrottledSpinner
 */
export type EnhancedSpinner = Ora | ThrottledSpinner;

export interface SpinnerContext {
  /**
   * Spinner instance for providing visual feedback
   * This can be either a regular Ora spinner or our enhanced ThrottledSpinner
   */
  spinner: EnhancedSpinner;
}

/**
 * State container for sharing workflow state between helper functions
 * 
 * This interface represents the complete state of the workflow at any point in time.
 * It accumulates data as the workflow progresses through its phases, with each helper
 * function adding its results to this state container.
 * 
 * Properties marked with '?' are optional because they may not be available at all
 * stages of the workflow. For example, early in the workflow, only 'options' will be
 * populated, but as the workflow progresses, more fields will be populated.
 * 
 * This state container is also used for error handling, allowing the error handler
 * to provide context-specific information based on which phase of the workflow
 * the error occurred in.
 */
export interface WorkflowState {
  /**
   * The user-provided run options from the command line
   * This is the only field that's guaranteed to be present at all workflow stages
   */
  options: RunOptions;
  
  /**
   * The loaded application configuration from the config file
   * Populated during the setup phase
   */
  config?: AppConfig;
  
  /**
   * Human-readable run identifier generated during setup
   * Used for naming output files and directories, and in console output
   */
  friendlyRunName?: string;
  
  /**
   * The processed input data containing the prompt and source metadata
   * Populated during the input processing phase
   */
  inputResult?: InputResult;
  
  /**
   * Absolute path to the directory where output files will be written
   * Created during the setup phase
   */
  outputDirectoryPath?: string;
  
  /**
   * Results of model selection, including the list of selected models
   * and any warnings or errors encountered during selection
   * Populated during the model selection phase
   */
  modelSelectionResult?: ModelSelectionResult;
  
  /**
   * Results of executing queries against the selected models
   * Includes the responses from each model and timing information
   * Populated during the query execution phase
   */
  queryResults?: QueryExecutionResult;
  
  /**
   * Results of writing output files to the filesystem
   * Includes success/failure information for each file
   * Populated during the output processing phase
   */
  fileOutputResult?: FileOutputResult;
  
  /**
   * Formatted console output text for display
   * Generated during the output processing phase
   */
  consoleOutput?: string;
}

// ----------------------------------------
// 1. Setup Workflow Helper
// ----------------------------------------
/**
 * The Setup phase initializes the workflow by loading configuration
 * and creating the output directory. This is the first step in 
 * the workflow sequence and establishes the foundation for all
 * subsequent operations.
 */

/**
 * Parameters for the _setupWorkflow helper function
 */
export interface SetupWorkflowParams extends SpinnerContext {
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * Configuration manager for loading and saving application configuration
   * Used for dependency injection to improve testability
   */
  configManager: ConfigManagerInterface;
  
  /**
   * File system interface for file operations
   * Used for dependency injection to improve testability
   */
  fileSystem: FileSystem;
}

/**
 * Result of the _setupWorkflow helper function
 */
export interface SetupWorkflowResult {
  /**
   * The loaded application configuration
   */
  config: AppConfig;
  
  /**
   * Friendly run name for display and reference
   */
  friendlyRunName: string;
  
  /**
   * Output directory path for writing files
   */
  outputDirectoryPath: string;
}

// ----------------------------------------
// 2. Process Input Helper
// ----------------------------------------
/**
 * The Input Processing phase handles retrieving and parsing the prompt
 * content from various sources (file, stdin, or direct text input).
 * This phase ensures that valid input content is available for the
 * subsequent model queries.
 */

/**
 * Parameters for the _processInput helper function
 */
export interface ProcessInputParams extends SpinnerContext {
  /**
   * The user-provided input (file path, raw text, or stdin indicator)
   */
  input: string;
  
  /**
   * Optional array of paths to files or directories to include as context
   * If provided, these will be read and combined with the prompt
   */
  contextPaths?: string[];
  
  /**
   * File system interface for file operations
   * Used for dependency injection to improve testability
   * Optional for backward compatibility with existing tests
   */
  fileSystem?: FileSystem;
}

/**
 * Extended metadata for input processing including context information
 */
export interface ExtendedInputMetadata {
  /**
   * Processing time in milliseconds
   */
  processingTimeMs: number;
  
  /**
   * Original content length before any processing
   */
  originalLength: number;
  
  /**
   * Final content length after processing
   */
  finalLength: number;
  
  /**
   * Whether the content was normalized
   */
  normalized: boolean;
  
  /**
   * Number of context files successfully processed (if applicable)
   */
  contextFilesCount?: number;
  
  /**
   * Number of context files that had errors during processing (if applicable)
   */
  contextFilesWithErrors?: number;
  
  /**
   * Whether the content includes context files
   */
  hasContextFiles?: boolean;
}

/**
 * Extended input result with context information
 */
export interface ExtendedInputResult extends Omit<InputResult, 'metadata'> {
  /**
   * Extended metadata including context information
   */
  metadata: ExtendedInputMetadata;
}

/**
 * Result of the _processInput helper function
 * 
 * This interface provides both the full ExtendedInputResult with all metadata
 * and a direct accessor for the combined content (the content field already
 * contains prompt+context when context files are provided)
 */
export interface ProcessInputResult {
  /**
   * Processed input with content and metadata
   * When context files are provided, the content field contains
   * the combined prompt+context content
   */
  inputResult: ExtendedInputResult;
  
  /**
   * Array of context file results if contextPaths were provided
   * Includes both successful and failed context files
   */
  contextFiles?: Array<import('../utils/fileReaderTypes').ContextFileResult>;
  
  /**
   * Direct accessor for the combined prompt+context content
   * This is the same as inputResult.content, provided for convenience
   * and more explicit access by other components like executeQueries
   */
  combinedContent: string;
}

// ----------------------------------------
// 3. Select Models Helper
// ----------------------------------------
/**
 * The Model Selection phase determines which LLM models to query based on
 * the user's options (specific model, group, or explicit list). It validates
 * model availability, API key presence, and filter settings, then produces
 * a filtered list of models with appropriate warnings.
 */

/**
 * Parameters for the _selectModels helper function
 */
export interface SelectModelsParams extends SpinnerContext {
  /**
   * The loaded application configuration
   */
  config: AppConfig;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
}

/**
 * Result of the _selectModels helper function
 * 
 * This is an intersection type that combines the ModelSelectionResult with
 * additional properties specific to the _selectModels helper function.
 * Using an intersection type provides a flatter structure than nested objects
 * while maintaining full type safety.
 */
export type SelectModelsResult = ModelSelectionResult & {
  /**
   * Description of the model selection mode (group, specific model, etc.)
   */
  modeDescription: string;
  
  /**
   * Self-reference for backward compatibility
   * Allows existing code to access properties via modelSelectionResult.property
   * @deprecated Access properties directly instead (result.models vs result.modelSelectionResult.models)
   */
  modelSelectionResult: ModelSelectionResult;
}

// ----------------------------------------
// 4. Execute Queries Helper
// ----------------------------------------
/**
 * The Query Execution phase sends prompts to all selected models in parallel
 * and collects their responses. This is the core phase of the workflow that
 * handles communication with LLM provider APIs, status tracking, response
 * parsing, and error handling for each model query.
 */

/**
 * Parameters for the _executeQueries helper function
 */
export interface ExecuteQueriesParams extends SpinnerContext {
  /**
   * The loaded application configuration
   */
  config: AppConfig;
  
  /**
   * The models to query
   */
  models: ModelConfig[];
  
  /**
   * The combined prompt and context content
   * This is the formatted content that will be sent to the LLM,
   * potentially containing both the user prompt and additional context files
   */
  combinedContent: string;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * The LLMClient instance to use for API calls
   * This implements the dependency injection pattern for better testability
   */
  llmClient: LLMClient;
}

/**
 * Result of the _executeQueries helper function
 */
export interface ExecuteQueriesResult {
  /**
   * The results of query execution, including responses and statuses
   */
  queryResults: QueryExecutionResult;
}

// ----------------------------------------
// 5. Process Output Helper
// ----------------------------------------
/**
 * The Output Processing phase takes the model responses and writes them to
 * output files, as well as formatting console output. It handles file creation,
 * content formatting, metadata inclusion, and error handling for file system
 * operations.
 */

/**
 * Parameters for the _processOutput helper function
 */
export interface ProcessOutputParams extends SpinnerContext {
  /**
   * The query execution results to process
   */
  queryResults: QueryExecutionResult;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * Friendly run name for display and reference
   * This property is optional because it's not directly used in the pure implementation
   */
  friendlyRunName?: string;
}

/**
 * Represents data for a file to be written
 */
export interface FileData {
  /**
   * The filename for this output file
   */
  filename: string;
  
  /**
   * The content to be written to the file
   */
  content: string;
  
  /**
   * The model key (provider:modelId) associated with this file
   */
  modelKey: string;
}

/**
 * Pure result of the refactored _processOutput helper function
 * with no I/O operations
 */
export interface PureProcessOutputResult {
  /**
   * Array of file data objects that can be written to disk
   */
  files: FileData[];
  
  /**
   * Formatted console output string
   */
  consoleOutput: string;
}

/**
 * Result of the _processOutput helper function
 */
export interface ProcessOutputResult {
  /**
   * The file output result, including successes and failures
   */
  fileOutputResult: FileOutputResult;
  
  /**
   * Formatted console output
   */
  consoleOutput: string;
}

// ----------------------------------------
// 6. Log Completion Summary Helper
// ----------------------------------------
/**
 * The Completion Summary phase generates and displays a human-readable report
 * of the workflow's execution, including success/failure counts, timing information,
 * and model-specific details. This phase is responsible for providing clear feedback
 * to the user about the outcome of their request.
 */

// _logCompletionSummary has been refactored out; these interfaces are no longer needed

// ----------------------------------------
// 7. Handle Workflow Error Helper
// ----------------------------------------
/**
 * The Error Handling phase provides centralized error processing for all
 * workflow phases. It categorizes unknown errors, adds contextual information
 * and suggestions, formats error messages for users, and ensures consistent
 * error handling throughout the workflow.
 */

/**
 * Parameters for the _handleWorkflowError helper function
 */
export interface HandleWorkflowErrorParams extends SpinnerContext {
  /**
   * The error to handle
   */
  error: unknown;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * The current state of the workflow when the error occurred
   */
  workflowState: Partial<WorkflowState>;
}

/**
 * Result of the _handleWorkflowError helper function
 * Helper will always throw, so no return value is defined
 */
export interface HandleWorkflowErrorResult {
  /**
   * The categorized error that will be thrown
   */
  error: ThinktankError;
}

// ----------------------------------------
// Error Handling Contracts
// ----------------------------------------
/**
 * This section defines the error handling contracts for the workflow helper functions.
 * These contracts specify which error types each function is responsible for handling,
 * which types it wraps, and which types it throws. This ensures consistent error
 * handling behavior throughout the workflow pipeline.
 */

/**
 * Error handling responsibility matrix for helper functions
 * 
 * This contract defines which functions are responsible for handling and wrapping
 * specific error types, ensuring a consistent approach to error handling across
 * the entire workflow.
 * 
 * Each helper function has three categories of error handling responsibilities:
 * 
 * - `handles`: Error types that the function is responsible for recognizing and responding to.
 *   When these errors occur, the function should handle them directly.
 * 
 * - `wraps`: Error types from other modules that need to be converted into standard 
 *   ThinktankError types. This allows us to normalize errors from different sources.
 * 
 * - `throws`: Error types that the function is expected to throw in response to various
 *   failure scenarios. These are the error types that callers should anticipate.
 * 
 * This explicit contract makes error handling expectations clear and helps ensure
 * that all errors are properly categorized, enriched with context, and presented
 * to users in a consistent way.
 */
export const errorHandlingContracts = {
  /**
   * Setup workflow error handling contract
   * Primarily deals with configuration and file system errors during initialization
   */
  _setupWorkflow: {
    handles: ['ConfigError', 'FileSystemError', 'PermissionError'],
    wraps: ['Error', 'NodeJS.ErrnoException'],
    throws: ['ConfigError', 'FileSystemError', 'PermissionError', 'ThinktankError']
  },
  
  /**
   * Input processing error handling contract
   * Focuses on file access and content parsing errors
   */
  _processInput: {
    handles: ['FileSystemError'],
    wraps: ['Error', 'NodeJS.ErrnoException', 'InputError'],
    throws: ['FileSystemError', 'ThinktankError']
  },
  
  /**
   * Model selection error handling contract
   * Deals with configuration validation and model selection errors
   */
  _selectModels: {
    handles: ['ModelSelectionError', 'ConfigError'],
    wraps: ['Error', 'ModelSelectionError'],
    throws: ['ConfigError', 'ApiError', 'ThinktankError']
  },
  
  /**
   * Query execution error handling contract
   * Handles API communication and response parsing errors
   */
  _executeQueries: {
    handles: ['ApiError', 'Error'],
    wraps: ['Error'],
    throws: ['ApiError', 'ThinktankError']
  },
  
  /**
   * Output processing error handling contract
   * Manages file system and permission errors during output writing
   */
  _processOutput: {
    handles: ['FileSystemError', 'PermissionError'],
    wraps: ['Error', 'NodeJS.ErrnoException'],
    throws: ['FileSystemError', 'PermissionError', 'ThinktankError']
  },
  
  /**
   * Centralized error handler contract
   * Can handle any error type and converts it to a standardized format
   */
  _handleWorkflowError: {
    handles: ['Error', 'unknown'],
    wraps: ['Error', 'unknown'],
    throws: ['ThinktankError', 'ConfigError', 'ApiError', 'FileSystemError', 'PermissionError']
  }
};
