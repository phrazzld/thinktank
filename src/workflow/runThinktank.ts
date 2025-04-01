/**
 * Main orchestration for the thinktank application
 * 
 * This module brings together all the specialized components and orchestrates the 
 * workflow, acting as the primary integration point for the application.
 */
import { processInput, InputResult } from './inputHandler';
import { loadConfig } from '../core/configManager';
import { selectModels, ModelSelectionError, ModelSelectionResult } from './modelSelector';
import { executeQueries, QueryExecutionResult } from './queryExecutor';
import { createOutputDirectory, formatForConsole, writeResponsesToFiles, FileOutputResult } from './outputHandler';
// No need to import LLMResponse as it's not directly used in this file
import { 
  styleHeader, 
  styleDim, 
  styleSuccess, 
  styleError, 
  styleWarning, 
  styleInfo,
  formatError, 
  formatErrorWithTip,
  errorCategories,
  colors
} from '../utils/consoleUtils';
import ora from 'ora';

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
   * Whether to include thinking output (for Claude models that support it)
   */
  includeThinking?: boolean;
  
  /**
   * Whether to enable Claude's thinking capability
   */
  enableThinking?: boolean;
}

/**
 * Error class for thinktank runtime errors
 * Provides additional context like category and helpful suggestions
 */
export class ThinktankError extends Error {
  /**
   * The category of error (e.g., "File System", "API", etc.)
   */
  category?: string;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands related to this error context
   */
  examples?: string[];
  
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'ThinktankError';
  }
}

/**
 * Creates a nice tree-style summary of the execution results 
 * 
 * @param results - The query execution results
 * @param options - The run options for context
 * @returns Formatted console output
 */
function formatResultsSummary(
  results: QueryExecutionResult,
  options?: RunOptions
): string {
  const { responses, statuses } = results;
  const successCount = Object.values(statuses).filter(s => s.status === 'success').length;
  const errorCount = Object.values(statuses).filter(s => s.status === 'error').length;
  
  // Create mode-specific completion message
  let completionMessage = '';
  if (options?.specificModel) {
    completionMessage = options.specificModel;
  } else if (options?.groupName) {
    completionMessage = `${options.groupName} group (${responses.length} model${responses.length === 1 ? '' : 's'})`;
  } else {
    completionMessage = `${responses.length} model${responses.length === 1 ? '' : 's'}`;
  }
  
  // Format completion time
  const elapsedTime = results.timing.durationMs;
  const completionTimeText = elapsedTime > 1000 
    ? `${(elapsedTime / 1000).toFixed(2)}s` 
    : `${elapsedTime}ms`;
  
  // Calculate success percentage
  const percentage = successCount > 0 ? Math.round((successCount / responses.length) * 100) : 0;
  
  let summaryOutput = '';
  
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
    summaryOutput += `\n${colors.blue('📊')} Results Summary:\n`;
    summaryOutput += `${colors.dim('│')}\n`;
    
    // First show successful models
    if (successCount > 0) {
      summaryOutput += `${colors.dim('├')} ${colors.green('✓')} Successful Models (${successCount}):\n`;
      const successModels = Object.entries(statuses)
        .filter(([_, status]) => status.status === 'success')
        .map(([model]) => model);
        
      successModels.forEach((model, i) => {
        const isLast = i === successModels.length - 1;
        const prefix = isLast ? `${colors.dim('│  └')}` : `${colors.dim('│  ├')}`;
        summaryOutput += `${prefix} ${model}\n`;
      });
    }
    
    // Then show failed models by category
    summaryOutput += `${colors.dim('├')} ${colors.red('✖')} Failed Models (${errorCount}):\n`;
    
    // Display errors by category
    Object.entries(errorsByCategory).forEach(([category, errors], categoryIndex, categories) => {
      const isLastCategory = categoryIndex === categories.length - 1;
      const categoryPrefix = isLastCategory ? `${colors.dim('│  └')}` : `${colors.dim('│  ├')}`;
      
      summaryOutput += `${categoryPrefix} ${colors.yellow(category)} errors (${errors.length}):\n`;
      
      errors.forEach(({ model, message }, errorIndex) => {
        const isLastError = errorIndex === errors.length - 1;
        const errorPrefix = isLastCategory 
          ? (isLastError ? `${colors.dim('│     └')}` : `${colors.dim('│     ├')}`)
          : (isLastError ? `${colors.dim('│  │  └')}` : `${colors.dim('│  │  ├')}`);
          
        summaryOutput += `${errorPrefix} ${colors.red(model)}\n`;
        
        // Add indented error message
        const messagePrefix = isLastCategory
          ? (isLastError ? `${colors.dim('│      ')}` : `${colors.dim('│     │')}`)
          : (isLastError ? `${colors.dim('│  │   ')}` : `${colors.dim('│  │  │')}`);
          
        summaryOutput += `${messagePrefix} ${colors.dim('→')} ${message}\n`;
      });
    });
    
    summaryOutput += `${colors.dim('└')} Completed in ${completionTimeText}\n`;
  } else {
    // Format complete success message
    summaryOutput += styleSuccess(`Successfully completed ${completionMessage} in ${completionTimeText}\n`);
    
    // Display a nice tree-style summary for successful models
    summaryOutput += `\n${colors.blue('📊')} Results Summary:\n`;
    summaryOutput += `${colors.dim('│')}\n`;
    
    const successModels = Object.keys(statuses);
    successModels.forEach((model, i) => {
      const isLast = i === successModels.length - 1;
      const prefix = isLast ? `${colors.dim('├')}` : `${colors.dim('├')}`;
      summaryOutput += `${prefix} ${i+1}. ${model} - ${colors.green('✓')} Success\n`;
    });
    
    summaryOutput += `${colors.dim('└')} Complete.\n`;
  }
  
  return summaryOutput;
}

/**
 * Main function to run thinktank
 * 
 * @param options - Options for running thinktank
 * @returns The formatted results
 * @throws {ThinktankError} If an error occurs during execution
 */
export async function runThinktank(options: RunOptions): Promise<string> {
  // Initialize the ora spinner for basic loading states
  const spinner = ora('Starting thinktank...').start();
  
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    spinner.text = 'Configuration loaded successfully';
    
    // 2. Process input (file, stdin, or direct text)
    spinner.text = 'Processing input...';
    const inputResult: InputResult = await processInput({ input: options.input });
    spinner.text = `Input processed from ${inputResult.sourceType} (${inputResult.content.length} characters)`;
    
    // 3. Create output directory
    spinner.text = 'Creating output directory...';
    // Determine directory identifier based on options
    const directoryIdentifier = options.specificModel || options.groupName;
    const outputDirectoryPath = await createOutputDirectory({
      outputDirectory: options.output,
      directoryIdentifier
    });
    spinner.info(styleInfo(`Output directory created: ${outputDirectoryPath}`));
    
    // 4. Select models using ModelSelector
    spinner.text = 'Selecting models to query...';
    let modelSelectionResult: ModelSelectionResult;
    
    try {
      modelSelectionResult = selectModels(config, {
        models: options.models,
        specificModel: options.specificModel,
        groupName: options.groupName,
        groups: options.groups,
        includeDisabled: true, // Include disabled models if explicitly requested
        validateApiKeys: true,
        throwOnError: true
      });
      
      // Display any warnings from model selection
      modelSelectionResult.warnings.forEach(warning => {
        spinner.warn(styleWarning(warning));
      });
      
      // Display information about disabled models that will be used
      if (modelSelectionResult.disabledModels.length > 0) {
        const disabledModelNames = modelSelectionResult.disabledModels
          .map(model => `${model.provider}:${model.modelId}`)
          .join(', ');
        
        spinner.info(styleInfo(`Disabled models that will be used: ${disabledModelNames}`));
      }
      
      // If we don't have any models after filtering, show warning and return early
      if (modelSelectionResult.models.length === 0) {
        const message = 'No models available after filtering.';
        spinner.warn(styleWarning(message));
        return message;
      }
      
      // Show selected models count
      spinner.text = `Selected ${modelSelectionResult.models.length} model(s) to query`;
    } catch (error) {
      // Handle errors from model selection
      if (error instanceof ModelSelectionError) {
        // Convert ModelSelectionError to ThinktankError
        const thinktankError = new ThinktankError(error.message, error);
        
        if (error.category) {
          thinktankError.category = error.category;
        }
        
        if (error.suggestions) {
          thinktankError.suggestions = error.suggestions;
        }
        
        if (error.examples) {
          thinktankError.examples = error.examples;
        }
        
        // Display detailed error with spinner
        spinner.fail(formatError(
          error.message, 
          error.category || errorCategories.CONFIG, 
          'Check your model specifications and configuration'
        ));
        
        throw thinktankError;
      } else {
        // Rethrow other errors
        throw error;
      }
    }
    
    // 5. Display model information based on selection mode
    // Provide CLI mode specific description
    let modeDescription: string;
    if (options.specificModel) {
      modeDescription = `Using specific model: ${styleInfo(options.specificModel)}`;
    } else if (options.groupName) {
      modeDescription = `Using models from group: ${styleInfo(options.groupName)} (${modelSelectionResult.models.length} models)`;
    } else if (options.models && options.models.length > 0) {
      modeDescription = `Using specified models: ${options.models.map(m => styleInfo(m)).join(', ')}`;
    } else if (options.groups && options.groups.length > 0) {
      modeDescription = `Using models from groups: ${options.groups.map(g => styleInfo(g)).join(', ')} (${modelSelectionResult.models.length} models)`;
    } else {
      modeDescription = `Using all enabled models (${modelSelectionResult.models.length} models)`;
    }
    spinner.info(modeDescription);
    
    // Show list of models that will be queried
    const modelList = modelSelectionResult.models
      .map((model, index) => `  ${index + 1}. ${model.provider}:${model.modelId}${!model.enabled ? ' (disabled)' : ''}`)
      .join('\n');
    
    // eslint-disable-next-line no-console
    console.log(modelList);
    
    // 6. Execute queries using QueryExecutor
    spinner.text = `Querying ${modelSelectionResult.models.length} model(s)...`;
    
    // Set up status update callback to update spinner
    const onStatusUpdate = (modelKey: string, status: { status: string; message?: string }): void => {
      if (status.status === 'running') {
        spinner.text = `Querying ${modelKey}...`;
      } else if (status.status === 'success') {
        spinner.succeed(`Model ${modelKey} completed successfully`);
        spinner.start(); // Restart spinner for next model
      } else if (status.status === 'error') {
        spinner.fail(`Model ${modelKey} failed: ${status.message}`);
        spinner.start(); // Restart spinner for next model
      }
    };
    
    // Execute queries with the QueryExecutor
    const queryResults = await executeQueries(config, modelSelectionResult.models, {
      prompt: inputResult.content,
      systemPrompt: options.systemPrompt,
      enableThinking: options.enableThinking,
      timeoutMs: 120000, // 2 minute timeout
      onStatusUpdate
    });
    
    // Display execution summary
    spinner.stop(); // Stop any active spinner
    // eslint-disable-next-line no-console
    console.log(formatResultsSummary(queryResults, options));
    
    // 7. Write responses to files
    spinner.start();
    spinner.text = `Writing ${queryResults.responses.length} model responses to files...`;
    
    // Set up status update callback for file writing
    const onFileWriteStatusUpdate = (fileDetail: { status: string; filename: string; error?: string }): void => {
      if (fileDetail.status === 'success') {
        spinner.succeed(`Wrote file: ${fileDetail.filename}`);
        spinner.start(); // Restart spinner for next file
      } else if (fileDetail.status === 'error') {
        spinner.fail(`Failed to write file: ${fileDetail.filename} - ${fileDetail.error}`);
        spinner.start(); // Restart spinner for next file
      }
    };
    
    // Write responses to files using OutputHandler
    const fileOutputResult: FileOutputResult = await writeResponsesToFiles(
      queryResults.responses,
      outputDirectoryPath,
      {
        includeMetadata: options.includeMetadata,
        throwOnError: false, // Don't throw on file write errors to ensure we continue with remaining files
        onStatusUpdate: onFileWriteStatusUpdate
      }
    );
    
    // Format completion message based on file writing results
    if (fileOutputResult.failedWrites === 0) {
      spinner.succeed(styleSuccess(
        `All ${fileOutputResult.succeededWrites} model responses saved to ${styleInfo(outputDirectoryPath)}`
      ));
    } else {
      spinner.warn(styleWarning(
        `Completed with issues: ${fileOutputResult.succeededWrites} successful, ${fileOutputResult.failedWrites} failed writes`
      ));
      
      // Show files with errors
      const failedFiles = fileOutputResult.files.filter(file => file.status === 'error');
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('❌ Files with errors:'));
      
      failedFiles.forEach(file => {
        // eslint-disable-next-line no-console
        console.log(styleError(`  - ${file.filename}: ${file.error || 'Unknown error'}`));
      });
    }
    
    // Show output directory
    // eslint-disable-next-line no-console
    console.log(`\n${styleInfo(`📁 Output directory: ${outputDirectoryPath}`)}`);
    
    // 8. Format model responses for console output
    const consoleOutput = formatForConsole(queryResults.responses, {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors !== false, // Default to true
      includeThinking: options.includeThinking,
      useTable: process.env.NODE_ENV !== 'test' // Only use table format in real CLI usage
    });
    
    // Show execution metadata if requested
    if (options.includeMetadata) {
      // Display timing information
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('Execution timing:'));
      
      const totalTime = queryResults.timing.durationMs + fileOutputResult.timing.durationMs;
      
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Total API calls:    ${queryResults.timing.durationMs}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  File writing:       ${fileOutputResult.timing.durationMs}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Total execution:    ${totalTime}ms`));
      
      // Additional model-specific timing information
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('Model timing:'));
      
      Object.entries(queryResults.statuses)
        .sort((a, b) => (a[1].durationMs || 0) - (b[1].durationMs || 0))
        .forEach(([model, status]) => {
          if (status.durationMs) {
            const statusIcon = status.status === 'success' ? '✅' : '❌';
            // eslint-disable-next-line no-console
            console.log(styleDim(`  ${statusIcon} ${model}: ${status.durationMs}ms`));
          }
        });
    }
    
    // Return the formatted results for CLI display
    return consoleOutput;
  } catch (error) {
    spinner.fail(formatErrorWithTip(error instanceof Error ? error : 'An unknown error occurred'));
    
    if (error instanceof Error) {
      throw new ThinktankError(`Error running thinktank: ${error.message}`, error);
    }
    
    throw new ThinktankError('Unknown error running thinktank');
  }
}