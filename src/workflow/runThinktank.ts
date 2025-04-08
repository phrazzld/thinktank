/**
 * Main orchestration for the thinktank application
 * 
 * This module brings together all the specialized components and orchestrates the 
 * workflow, acting as the primary integration point for the application.
 */
import { 
  styleHeader,
  styleDim,
  styleWarning
} from '../utils/consoleUtils';
import { logger } from '../utils/logger';
import ora, { configureSpinnerFactory } from '../utils/spinnerFactory';
import { WorkflowState, EnhancedSpinner } from './runThinktankTypes';
import {
  _setupWorkflow,
  _processInput,
  _selectModels,
  _executeQueries,
  _processOutput,
  _logCompletionSummary,
  _handleWorkflowError
} from './runThinktankHelpers';

// Import the concrete implementations and interfaces
import { ConcreteLLMClient } from '../core/LLMClient';
import { ConcreteConfigManager } from '../core/ConcreteConfigManager';
import { ConcreteFileSystem } from '../core/FileSystem';
import { LLMClient, ConfigManagerInterface, FileSystem } from '../core/interfaces';

// Import provider modules to ensure they're registered
import '../providers/openai';
import '../providers/anthropic';
import '../providers/google';
import '../providers/openrouter';
// Future providers will be imported here

/**
 * Options for running thinktank
 * 
 * The model selection hierarchy is as follows (highest to lowest priority):
 * 1. Multiple specific models (models array) - use exactly these models
 * 2. Single specific model (specificModel) - use just this one model
 * 3. Group name (groupName) - use all models in this group
 * 4. Groups array (groups) - use all models in these groups
 * 5. Default - use all enabled models
 * 
 * When both models and groupName are specified, models are filtered to only include
 * those that are both explicitly requested and in the specified group.
 */
export interface RunOptions {
  /**
   * Path to the input prompt file, content directly, or '-' for stdin
   */
  input: string;
  
  /**
   * Array of paths to files or directories to include as context (optional)
   * If provided, these will be read and combined with the prompt
   */
  contextPaths?: string[];
  
  /**
   * Path to the configuration file (optional)
   */
  configPath?: string;
  
  /**
   * Path to the output directory (optional)
   * If provided, this will be used as the parent directory for the run-specific output folder
   * If not provided, './thinktank-reports/' in the current working directory will be used
   */
  output?: string;
  
  /**
   * Array of specific model identifiers to use in provider:modelId format (e.g., ["openai:gpt-4o", "anthropic:claude-3-opus"])
   * If provided, only these models will be used
   * Takes highest precedence in the model selection hierarchy
   * Can be combined with groupName to filter models by both criteria
   */
  models?: string[];
  
  /**
   * Array of group names to use (optional)
   * If not provided, all groups will be used
   * If provided, only models in the specified groups will be used
   * Lower precedence than models, specificModel, and groupName
   */
  groups?: string[];
  
  /**
   * A specific model to use in provider:modelId format (e.g., "openai:gpt-4o")
   * If provided, only this model will be used
   * Takes precedence over groups and groupName parameters
   * Kept for backward compatibility - the models array is recommended for new code
   */
  specificModel?: string;
  
  /**
   * A single group name to use
   * If provided, only models in this group will be used
   * Takes precedence over groups parameter (array)
   * Can be combined with models to filter by both criteria
   */
  groupName?: string;
  
  /**
   * System prompt override (optional)
   * If provided, this system prompt will be used for all models, regardless of their group's system prompt
   */
  systemPrompt?: string;
  
  /**
   * Whether to include metadata in the output
   */
  includeMetadata?: boolean;
  
  /**
   * Whether to use colors in the output
   */
  useColors?: boolean;
  
  /**
   * Whether to disable spinner throttling
   * If true, spinner updates will not be throttled, which may cause more terminal flicker
   * If false or undefined, spinner updates will be throttled to reduce flicker
   */
  disableSpinnerThrottling?: boolean;
  
  /**
   * Whether to include thinking output (for Claude models that support it)
   */
  includeThinking?: boolean;
  
  /**
   * Whether to enable Claude's thinking capability
   */
  enableThinking?: boolean;
  
  /**
   * Timeout in milliseconds for each model query
   */
  timeoutMs?: number;
  
  /**
   * Friendly name for the run
   * Used in console output, but not for actual directory naming
   * Set internally during run execution - not meant to be provided by users
   */
  friendlyRunName?: string;
}

// ThinktankError class is now imported from src/core/errors.ts

// The formatResultsSummary function has been replaced by the _logCompletionSummary helper

/**
 * Main function to run thinktank
 * 
 * This function orchestrates the overall workflow by calling the specialized helper
 * functions in sequence and handling errors appropriately.
 * 
 * @param options - Options for running thinktank
 * @returns The formatted results for display
 * @throws {ThinktankError} If an error occurs during execution
 */
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
    // 1. First create FileSystem (no dependencies)
    const fileSystem: FileSystem = new ConcreteFileSystem();
    // 2. Then create ConfigManager (conceptually uses FileSystem, but doesn't require it in constructor)
    const configManager: ConfigManagerInterface = new ConcreteConfigManager();
    // 3. Finally create LLMClient (depends on ConfigManager)
    const llmClient: LLMClient = new ConcreteLLMClient(configManager);
    
    // 1. Setup workflow: Load configuration, generate run name, create output directory
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
    
    // 2. Process input: Read from file, stdin, or direct text
    const inputResult = await _processInput({
      spinner,
      input: options.input,
      contextPaths: options.contextPaths,
      fileSystem
    });
    
    // Update workflow state with input result
    workflowState.inputResult = inputResult.inputResult;
    
    // 3. Select models: Determine which models to use based on options
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
    
    // Show list of selected models
    // Stop spinner before showing the model list to avoid visual conflicts
    spinner.stop();
    const modelList = modelSelectionResult.models
      .map((model, index) => `  ${index + 1}. ${model.provider}:${model.modelId}${!model.enabled ? ' (disabled)' : ''}`)
      .join('\n');
    
    logger.plain(modelList);
    // Restart spinner for next step
    spinner.start();
    
    // 4. Execute queries: Query the selected models with the processed input (including context if any)
    const queryResults = await _executeQueries({
      spinner,
      config: setupResult.config,
      models: modelSelectionResult.models,
      combinedContent: inputResult.combinedContent,
      options,
      llmClient  // Inject the LLMClient instance
    });
    
    // Update workflow state with query results
    workflowState.queryResults = queryResults.queryResults;
    
    // 5. Process output: Write results to files and format for console
    const outputResult = await _processOutput({
      spinner,
      queryResults: queryResults.queryResults,
      outputDirectoryPath: setupResult.outputDirectoryPath,
      options,
      friendlyRunName: setupResult.friendlyRunName,
      fileSystem  // Pass the FileSystem instance
    });
    
    // Update workflow state with output results
    workflowState.fileOutputResult = outputResult.fileOutputResult;
    workflowState.consoleOutput = outputResult.consoleOutput;
    
    // 6. Log completion summary: Display a summary of the executed queries
    // Stop spinner before showing completion summary to avoid visual conflicts
    spinner.stop();
    _logCompletionSummary({
      queryResults: queryResults.queryResults,
      fileOutputResult: outputResult.fileOutputResult,
      options: { ...options, friendlyRunName: setupResult.friendlyRunName },
      outputDirectoryPath: setupResult.outputDirectoryPath
    });
    
    // 7. Show additional metadata if requested
    if (options.includeMetadata) {
      // Spinner is already stopped after _logCompletionSummary call
      // Display timing information
      logger.plain('\n' + styleHeader('Execution timing:'));
      
      const totalTime = queryResults.queryResults.timing.durationMs + outputResult.fileOutputResult.timing.durationMs;
      
      logger.plain(styleDim(`  Total API calls:    ${queryResults.queryResults.timing.durationMs}ms`));
      logger.plain(styleDim(`  File writing:       ${outputResult.fileOutputResult.timing.durationMs}ms`));
      logger.plain(styleDim(`  Total execution:    ${totalTime}ms`));
      
      // Additional model-specific timing information
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
    return outputResult.consoleOutput;
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
