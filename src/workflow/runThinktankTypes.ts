/**
 * TypeScript interfaces for the runThinktank workflow helper functions
 * 
 * This file defines the input parameters, return types, and error handling contracts
 * for each helper function in the runThinktank workflow.
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
 */
export interface WorkflowState {
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * The loaded application configuration
   */
  config?: AppConfig;
  
  /**
   * Friendly run name for display and reference
   */
  friendlyRunName?: string;
  
  /**
   * Processed input result
   */
  inputResult?: InputResult;
  
  /**
   * Output directory path for writing files
   */
  outputDirectoryPath?: string;
  
  /**
   * Model selection result
   */
  modelSelectionResult?: ModelSelectionResult;
  
  /**
   * Query execution result
   */
  queryResults?: QueryExecutionResult;
  
  /**
   * File output result
   */
  fileOutputResult?: FileOutputResult;
  
  /**
   * Formatted console output
   */
  consoleOutput?: string;
}

// ----------------------------------------
// 1. Setup Workflow Helper
// ----------------------------------------

/**
 * Parameters for the _setupWorkflow helper function
 */
export interface SetupWorkflowParams extends SpinnerContext {
  /**
   * The user-provided run options
   */
  options: RunOptions;
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
 * Parameters for the _processInput helper function
 */
export interface ProcessInputParams extends SpinnerContext {
  /**
   * The user-provided input (file path, raw text, or stdin indicator)
   */
  input: string;
}

/**
 * Result of the _processInput helper function
 */
export interface ProcessInputResult {
  /**
   * Processed input with content and metadata
   */
  inputResult: InputResult;
}

// ----------------------------------------
// 3. Select Models Helper
// ----------------------------------------

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
 */
export interface SelectModelsResult {
  /**
   * The result of model selection, including selected models, warnings, and disabled models
   */
  modelSelectionResult: ModelSelectionResult;
  
  /**
   * Description of the model selection mode (group, specific model, etc.)
   */
  modeDescription: string;
}

// ----------------------------------------
// 4. Execute Queries Helper
// ----------------------------------------

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
   * The processed input content
   */
  prompt: string;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
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
 * Parameters for the _processOutput helper function
 */
export interface ProcessOutputParams extends SpinnerContext {
  /**
   * The query execution results to process
   */
  queryResults: QueryExecutionResult;
  
  /**
   * The output directory path to write files to
   */
  outputDirectoryPath: string;
  
  /**
   * The user-provided run options
   */
  options: RunOptions;
  
  /**
   * Friendly run name for display and reference
   */
  friendlyRunName: string;
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
 * Parameters for the _logCompletionSummary helper function
 */
export interface LogCompletionSummaryParams {
  /**
   * The query execution results
   */
  queryResults: QueryExecutionResult;
  
  /**
   * The file output result
   */
  fileOutputResult: FileOutputResult;
  
  /**
   * The user-provided run options with run name
   */
  options: RunOptions & { friendlyRunName: string };
  
  /**
   * The output directory path
   */
  outputDirectoryPath: string;
}

/**
 * Result of the _logCompletionSummary helper function
 * No specific return value since this function primarily logs to the console
 */
export interface LogCompletionSummaryResult {
  // Empty - function primarily produces side effects (logging)
}

// ----------------------------------------
// 7. Handle Workflow Error Helper
// ----------------------------------------

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
 * Error handling responsibility matrix for helper functions
 * Defines which functions are responsible for handling and wrapping specific error types
 */
export const errorHandlingContracts = {
  _setupWorkflow: {
    handles: ['ConfigError', 'FileSystemError', 'PermissionError'],
    wraps: ['Error', 'NodeJS.ErrnoException'],
    throws: ['ConfigError', 'FileSystemError', 'PermissionError', 'ThinktankError']
  },
  _processInput: {
    handles: ['FileSystemError'],
    wraps: ['Error', 'NodeJS.ErrnoException', 'InputError'],
    throws: ['FileSystemError', 'ThinktankError']
  },
  _selectModels: {
    handles: ['ModelSelectionError', 'ConfigError'],
    wraps: ['Error', 'ModelSelectionError'],
    throws: ['ConfigError', 'ApiError', 'ThinktankError']
  },
  _executeQueries: {
    handles: ['ApiError', 'Error'],
    wraps: ['Error'],
    throws: ['ApiError', 'ThinktankError']
  },
  _processOutput: {
    handles: ['FileSystemError', 'PermissionError'],
    wraps: ['Error', 'NodeJS.ErrnoException'],
    throws: ['FileSystemError', 'PermissionError', 'ThinktankError']
  },
  _handleWorkflowError: {
    handles: ['Error', 'unknown'],
    wraps: ['Error', 'unknown'],
    throws: ['ThinktankError', 'ConfigError', 'ApiError', 'FileSystemError', 'PermissionError']
  }
};