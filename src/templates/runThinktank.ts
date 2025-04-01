/**
 * Main orchestration for the thinktank application
 * 
 * This template connects all the components and orchestrates the workflow.
 */
import { readFileContent } from '../molecules/fileReader';
import { 
  loadConfig, 
  filterModels, 
  getEnabledModels, 
  validateModelApiKeys,
  getEnabledModelsFromGroups,
  findModelGroup 
} from '../organisms/configManager';
import { getProvider } from '../organisms/llmRegistry';
import { formatResults } from '../molecules/outputFormatter';
import { LLMResponse, ModelConfig, SystemPrompt } from '../atoms/types';
import { getModelConfigKey, generateOutputDirectoryPath, sanitizeFilename } from '../atoms/helpers';
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
  categorizeError,
  getTroubleshootingTip,
  createMissingApiKeyError
} from '../atoms/consoleUtils';
import ora from 'ora';
import fs from 'fs/promises';
import path from 'path';

// Import provider modules to ensure they're registered
import '../molecules/llmProviders/openai';
import '../molecules/llmProviders/anthropic';
import '../molecules/llmProviders/google';
import '../molecules/llmProviders/openrouter';
// Future providers will be imported here

/**
 * Options for running thinktank
 */
export interface RunOptions {
  /**
   * Path to the input prompt file
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
   * Note: Model responses are always written to files in a timestamped subdirectory
   */
  output?: string;
  
  /**
   * Array of model identifiers to use (optional)
   * If not provided, all enabled models will be used
   * Note: When groups parameter is used, this filters models within those groups
   */
  models?: string[];
  
  /**
   * Array of group names to use (optional)
   * If not provided, all groups will be used
   * If provided, only models in the specified groups will be used
   */
  groups?: string[];
  
  /**
   * A specific model to use in provider:modelId format (e.g., "openai:gpt-4o")
   * If provided, only this model will be used
   * Takes precedence over models and groups parameters
   */
  specificModel?: string;
  
  /**
   * A single group name to use
   * If provided, only models in this group will be used
   * Takes precedence over groups parameter (array)
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
 * Formats an LLM response as Markdown
 * 
 * @param response - The LLM response to format
 * @param includeMetadata - Whether to include metadata in the output
 * @returns The formatted Markdown
 */
function formatResponseAsMarkdown(
  response: LLMResponse & { configKey: string },
  includeMetadata = false
): string {
  const { text, error, metadata, configKey, groupInfo } = response;
  
  // Start with a header including group information if available
  let markdown = `# ${configKey}`;
  if (groupInfo && groupInfo.name !== 'default') {
    markdown += ` (${groupInfo.name} group)`;
  }
  markdown += '\n\n';
  
  // Add timestamp
  const timestamp = new Date().toISOString();
  markdown += `Generated: ${timestamp}\n\n`;
  
  // Add group information if available and not default
  if (groupInfo && groupInfo.name !== 'default') {
    markdown += `Group: ${groupInfo.name}\n`;
    if (groupInfo.systemPrompt && includeMetadata) {
      markdown += `System Prompt: "${groupInfo.systemPrompt.text}"\n`;
    }
    markdown += '\n';
  }
  
  // Add error if present
  if (error) {
    markdown += `## Error\n\n\`\`\`\n${error}\n\`\`\`\n\n`;
  }
  
  // Add the response text (if available)
  if (text) {
    markdown += `## Response\n\n${text}\n\n`;
  }
  
  // Include metadata if requested
  if (includeMetadata && metadata) {
    markdown += '## Metadata\n\n```json\n';
    markdown += JSON.stringify(metadata, null, 2);
    markdown += '\n```\n';
  }
  
  return markdown;
}

/**
 * Main function to run thinktank
 * 
 * @param options - Options for running thinktank
 * @returns The formatted results
 * @throws {ThinktankError} If an error occurs during execution
 */
/**
 * Helper function to get the current time in milliseconds
 * Uses performance.now() when available, otherwise falls back to Date.now()
 */
function getCurrentTime(): number {
  return typeof performance !== 'undefined' ? performance.now() : Date.now();
}

/**
 * Helper function to calculate elapsed time between a start time and now
 * @param startTime - The starting time in milliseconds
 * @returns Elapsed time in milliseconds
 */
function getElapsedTime(startTime: number): number {
  return Math.round(getCurrentTime() - startTime);
}

/**
 * Formats a string for the model processing status
 * 
 * @param models - All models being processed
 * @param modelStatuses - Current status of each model
 * @param startTime - Start time of API calls
 * @param currentModel - Current model being processed (optional)
 * @param options - The RunOptions for context
 * @returns Formatted status string
 */

/**
 * Formats a string for the file writing status
 * 
 * @param totalFiles - Total number of files to write
 * @param succeededWrites - Number of successful writes
 * @param failedWrites - Number of failed writes
 * @param startTime - Start time of file writing
 * @param currentFile - Current file being written (optional)
 * @returns Formatted status string
 */
function formatFileWritingStatus(
  totalFiles: number,
  succeededWrites: number,
  failedWrites: number,
  startTime: number,
  currentFile?: string
): string {
  const completedWrites = succeededWrites + failedWrites;
  const percentComplete = Math.round((completedWrites / totalFiles) * 100);
  
  // Format elapsed time in seconds
  const elapsedTime = Math.round((getCurrentTime() - startTime) / 1000);
  
  // Calculate estimated time remaining (if we have at least 10% progress)
  let etaString = '';
  if (percentComplete >= 10 && elapsedTime > 0 && completedWrites > 0) {
    const totalTimeEstimate = (elapsedTime * totalFiles) / completedWrites;
    const remainingTime = Math.round(totalTimeEstimate - elapsedTime);
    etaString = ` (ETA: ~${remainingTime}s remaining)`;
  }
  
  // Build status message with multiple lines for better readability
  let statusMsg = `Saving model responses [${completedWrites}/${totalFiles}] ${percentComplete}% complete${etaString}\n`;
  
  // Stats with colored output
  statusMsg += `  ${styleSuccess(`${succeededWrites} succeeded`)}, ${styleError(`${failedWrites} failed`)}`;
  
  // Add current file if provided
  if (currentFile) {
    statusMsg += `\n  Current: ${styleInfo(currentFile)}`;
  }
  
  // Add elapsed time
  statusMsg += `\n  Elapsed: ${elapsedTime}s`;
  
  return statusMsg;
}

function formatModelProcessingStatus(
  models: ModelConfig[], 
  modelStatuses: Record<string, { 
    status: 'pending' | 'success' | 'error';
    message?: string;
    detailedError?: Error;
  }>,
  startTime: number,
  currentModel?: string,
  options?: RunOptions
): string {
  const pendingCount = Object.values(modelStatuses).filter(s => s.status === 'pending').length;
  const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
  const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
  const totalCount = models.length;
  const completedCount = successCount + errorCount;
  const percentComplete = Math.round((completedCount / totalCount) * 100);
  
  // Format elapsed time in seconds
  const elapsedTime = Math.round((getCurrentTime() - startTime) / 1000);
  
  // Calculate estimated time remaining (if we have at least 10% progress)
  let etaString = '';
  if (percentComplete >= 10 && completedCount > 0 && elapsedTime > 0) {
    const totalTimeEstimate = (elapsedTime * totalCount) / completedCount;
    const remainingTime = Math.round(totalTimeEstimate - elapsedTime);
    etaString = ` (ETA: ~${remainingTime}s remaining)`;
  }
  
  // Build mode-specific prefix
  let modePrefix = 'Processing models';
  if (options?.specificModel) {
    modePrefix = `Running with model ${options.specificModel}`;
  } else if (options?.groupName) {
    modePrefix = `Running with group "${options.groupName}"`;
  } else if (options?.groups && options.groups.length > 0) {
    modePrefix = `Running with groups [${options.groups.join(', ')}]`;
  }
  
  // Build status message
  let statusMsg = `${modePrefix} [${completedCount}/${totalCount}] ${percentComplete}% complete${etaString}\n`;
  
  // Stats line
  statusMsg += `  ${styleSuccess(`${successCount} succeeded`)}, ${styleError(`${errorCount} failed`)}, ${styleDim(`${pendingCount} pending`)}`;
  
  // Add current model if provided
  if (currentModel) {
    statusMsg += `\n  Current: ${styleInfo(currentModel)}`;
  }
  
  // Add elapsed time
  statusMsg += `\n  Elapsed: ${elapsedTime}s`;
  
  return statusMsg;
}

export async function runThinktank(options: RunOptions): Promise<string> {
  // Start timing the full execution
  const startTimeTotal = getCurrentTime();
  
  const spinner = ora('Starting thinktank...').start();
  
  // Track the output directory path for later use
  let outputDirectoryPath: string | undefined;
  
  // Define a type for model status
  type ModelStatus = {
    status: 'pending' | 'success' | 'error';
    message?: string;
    detailedError?: Error;
  };
  
  // For tracking model statuses
  const modelStatuses: Record<string, ModelStatus> = {};
  
  // Store timing metadata for different stages
  const timings = {
    configLoad: 0,
    inputRead: 0,
    directoryCreation: 0,
    modelPreparation: 0,
    apiCalls: 0,
    fileWrites: 0,
    total: 0
  };
  
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const startTimeConfig = getCurrentTime();
    const config = await loadConfig({ configPath: options.configPath });
    timings.configLoad = getElapsedTime(startTimeConfig);
    spinner.text = `Loading configuration... (${timings.configLoad}ms)`;
    
    // 2. Read input file
    spinner.text = 'Reading input file...';
    const startTimeInput = getCurrentTime();
    const prompt = await readFileContent(options.input);
    timings.inputRead = getElapsedTime(startTimeInput);
    spinner.text = `Reading input file... (${timings.inputRead}ms)`;
    
    // 2.5 Create output directory with simplified naming
    // Generate the output directory path based on CLI mode (specific model or group)
    let directoryIdentifier: string | undefined;
    
    if (options.specificModel) {
      // Use the specific model as the directory identifier
      directoryIdentifier = options.specificModel;
    } else if (options.groupName) {
      // Use the group name as the directory identifier
      directoryIdentifier = options.groupName;
    }
    
    // Generate output directory path with the identifier
    outputDirectoryPath = generateOutputDirectoryPath(options.output, directoryIdentifier);
    
    spinner.text = `Creating output directory: ${outputDirectoryPath}`;
    const startTimeDirectory = getCurrentTime();
    try {
      // Create the directory with recursive option to ensure parent directories exist
      await fs.mkdir(outputDirectoryPath, { recursive: true });
      timings.directoryCreation = getElapsedTime(startTimeDirectory);
      spinner.info(styleInfo(`Output directory created: ${outputDirectoryPath} (${timings.directoryCreation}ms)`));
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      spinner.fail(formatError(
        `Failed to create output directory: ${errorMessage}`, 
        errorCategories.FILESYSTEM, 
        'Check your write permissions and ensure the path is valid'
      ));
      throw new ThinktankError(
        `Failed to create output directory: ${errorMessage}`,
        error instanceof Error ? error : undefined
      );
    }
    
    // 3. Select models based on specificModel, groupName, groups and/or model filters
    spinner.text = 'Preparing models...';
    const startTimePreparation = getCurrentTime();
    let models: ModelConfig[];
    
    // Prioritize specificModel over all other selection options
    if (options.specificModel) {
      // Handle a single model specification in provider:modelId format
      spinner.text = `Using specific model: ${options.specificModel}...`;
      
      // Split the specificModel string to get provider and modelId
      const [provider, modelId] = options.specificModel.split(':');
      
      if (!provider || !modelId) {
        // Create a detailed error with suggestions
        const { createModelFormatError } = await import('../atoms/consoleUtils');
        const { getProviderIds } = await import('../organisms/llmRegistry');
        
        // Get available providers and models for better error messages
        const availableProviders = getProviderIds();
        const enabledModels = getEnabledModels(config);
        const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
        
        // Create detailed error
        const modelError = createModelFormatError(
          options.specificModel,
          availableProviders,
          availableModels
        );
        
        // Display error with spinner
        spinner.fail(formatError(
          modelError.message, 
          errorCategories.CONFIG, 
          'Check the model format and ensure it follows the provider:modelId convention'
        ));
        
        // Convert to ThinktankError
        const modelFormatError = new ThinktankError(modelError.message);
        modelFormatError.category = (modelError as any).category;
        modelFormatError.suggestions = (modelError as any).suggestions;
        modelFormatError.examples = (modelError as any).examples;
        
        throw modelFormatError;
      }
      
      // Find the model config for this specific model
      const specificModelConfig = config.models.find(model => 
        model.provider === provider && model.modelId === modelId
      );
      
      if (!specificModelConfig) {
        // Create a detailed error with suggestions about model not found
        const { createModelNotFoundError } = await import('../atoms/consoleUtils');
        
        // Get all available models for better suggestions
        const enabledModels = getEnabledModels(config);
        const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
        
        // Create detailed error
        const modelError = createModelNotFoundError(
          options.specificModel,
          availableModels
        );
        
        // Display error with spinner
        spinner.fail(formatError(
          modelError.message, 
          errorCategories.CONFIG, 
          'Check your configuration file and make sure the model is defined'
        ));
        
        // Convert to ThinktankError
        const modelNotFoundError = new ThinktankError(modelError.message);
        modelNotFoundError.category = (modelError as any).category;
        modelNotFoundError.suggestions = (modelError as any).suggestions;
        modelNotFoundError.examples = (modelError as any).examples;
        
        throw modelNotFoundError;
      }
      
      if (!specificModelConfig.enabled) {
        const message = `Model "${options.specificModel}" is disabled in configuration.`;
        spinner.warn(styleWarning(message));
      }
      
      models = [specificModelConfig];
    }
    // Next priority is groupName (single group)
    else if (options.groupName) {
      // Handle a single group name
      spinner.text = `Selecting models from group: ${options.groupName}...`;
      
      if (!config.groups || !config.groups[options.groupName]) {
        // Create more helpful group not found error
        const message = `Group "${options.groupName}" not found in configuration.`;
        
        // Get available groups for better suggestions
        const availableGroups = config.groups ? Object.keys(config.groups) : [];
        
        // Create ThinktankError with detailed suggestions
        const groupError = new ThinktankError(message);
        groupError.category = errorCategories.CONFIG;
        
        // Add helpful suggestions
        const suggestions = [
          'Check your configuration file and make sure the group is defined'
        ];
        
        // List available groups if any
        if (availableGroups.length > 0) {
          suggestions.push(`Available groups: ${availableGroups.join(', ')}`);
        } else {
          suggestions.push('No groups defined in the configuration');
        }
        
        // Add configuration hint
        suggestions.push(
          'Groups must be defined in your thinktank.config.json file',
          'Use "thinktank models" to list all available models and their groups'
        );
        
        groupError.suggestions = suggestions;
        
        // Add examples based on available groups or defaults
        // Extract filename from input path for examples
        const inputFilename = options.input.split('/').pop() || 'prompt.txt';
        
        if (availableGroups.length > 0) {
          groupError.examples = availableGroups.map(group => `thinktank ${inputFilename} ${group}`);
        } else {
          groupError.examples = [
            `thinktank ${inputFilename} default`,
            `thinktank ${inputFilename} coding`,
            `thinktank ${inputFilename} fast`
          ];
        }
        
        // Display error with spinner
        spinner.fail(formatError(message, errorCategories.CONFIG, 'Check your configuration file and make sure the group is defined'));
        
        throw groupError;
      }
      
      models = getEnabledModelsFromGroups(config, [options.groupName]);
      
      // Apply model filters if specified
      if (options.models && options.models.length > 0) {
        spinner.text = 'Applying model filters to group models...';
        
        // Get all models matching the filters
        const filteredModels = options.models.flatMap(modelFilter => 
          filterModels(config, modelFilter)
        );
        
        // Create a set of model keys for efficient filtering
        const filteredModelKeys = new Set(
          filteredModels.map(model => `${model.provider}:${model.modelId}`)
        );
        
        // Keep only models that are both in the group and match filters
        models = models.filter(model => {
          const key = `${model.provider}:${model.modelId}`;
          return filteredModelKeys.has(key);
        });
      }
    }
    // Then fall back to the array versions
    else {
      // Track which approach we're using for user feedback
      const useGroupsSelection = options.groups && options.groups.length > 0;
      const useModelSelection = options.models && options.models.length > 0;
      
      if (useGroupsSelection) {
        // Get models from specified groups
        spinner.text = `Selecting models from groups: ${options.groups!.join(', ')}...`;
        models = getEnabledModelsFromGroups(config, options.groups!);
        
        // If models filter is also specified, further filter the group models
        if (useModelSelection) {
          spinner.text = 'Applying model filters to group models...';
          
          // Get all models matching the filters
          const filteredModels = options.models!.flatMap(modelFilter => 
            filterModels(config, modelFilter)
          );
          
          // Create a set of model keys for efficient filtering
          const filteredModelKeys = new Set(
            filteredModels.map(model => `${model.provider}:${model.modelId}`)
          );
          
          // Keep only models that are both in groups and match filters
          models = models.filter(model => {
            const key = `${model.provider}:${model.modelId}`;
            return filteredModelKeys.has(key);
          });
        }
      } else if (useModelSelection) {
        // Filter models by CLI args
        spinner.text = 'Selecting models by filter...';
        models = options.models!.flatMap(modelFilter => 
          filterModels(config, modelFilter)
        );
        
        // Remove duplicates
        const modelKeys = new Set<string>();
        models = models.filter(model => {
          const key = getModelConfigKey(model);
          if (modelKeys.has(key)) {
            return false;
          }
          modelKeys.add(key);
          return true;
        });
        
        // Filter to only enabled models
        models = models.filter(model => model.enabled);
      } else {
        // Use all enabled models
        spinner.text = 'Using all enabled models...';
        models = getEnabledModels(config);
      }
    }
    
    // Check if we have any models after all filtering
    if (models.length === 0) {
      if (options.specificModel) {
        const message = `Specific model "${options.specificModel}" is not available or not enabled.`;
        spinner.warn(styleWarning(message));
        return message;
      } else if (options.groupName) {
        const message = `No enabled models found in the specified group: ${options.groupName}`;
        spinner.warn(styleWarning(message));
        return message;
      } else if (options.groups && options.groups.length > 0 && options.models && options.models.length > 0) {
        const message = 'No enabled models found matching both group and model filters.';
        spinner.warn(styleWarning(message));
        return message;
      } else if (options.groups && options.groups.length > 0) {
        const groupList = options.groups.join(', ');
        const message = `No enabled models found in the specified groups: ${groupList}`;
        spinner.warn(styleWarning(message));
        return message;
      } else if (options.models && options.models.length > 0) {
        const message = 'No enabled models matched the specified filters.';
        spinner.warn(styleWarning(message));
        return message;
      } else {
        const message = 'No enabled models found in configuration.';
        spinner.warn(styleWarning(message));
        return message;
      }
    }
    
    // 4. Validate API keys
    const { missingKeyModels } = validateModelApiKeys(config);
    
    // Log warnings for models with missing API keys
    if (missingKeyModels.length > 0) {
      // Create a detailed error with provider-specific instructions
      const apiKeyError = createMissingApiKeyError(missingKeyModels);
      
      // Get the raw model names for display
      const modelNames = missingKeyModels.map(getModelConfigKey).join(', ');
      
      // Show warning with basic information (will expand on this if all models are missing keys)
      spinner.warn(styleWarning(`Missing API keys for models: ${modelNames}`));
      
      // Filter out models with missing keys
      models = models.filter(model => 
        !missingKeyModels.some(m => 
          m.provider === model.provider && m.modelId === model.modelId
        )
      );
      
      if (models.length === 0) {
        // No models with valid API keys - show detailed error
        // Convert apiKeyError to ThinktankError for more details
        const thinktankError = new ThinktankError(apiKeyError.message);
        thinktankError.category = (apiKeyError as any).category;
        thinktankError.suggestions = (apiKeyError as any).suggestions;
        thinktankError.examples = (apiKeyError as any).examples;
        
        // Display comprehensive error with spinner
        spinner.fail(formatError(
          'No models with valid API keys available.', 
          errorCategories.API,
          'See below for instructions on obtaining and setting API keys'
        ));
        
        // Display the detailed suggestions
        if (thinktankError.suggestions && thinktankError.suggestions.length > 0) {
          // eslint-disable-next-line no-console
          console.log('\n' + styleHeader('Missing API Keys:'));
          
          thinktankError.suggestions.forEach(suggestion => {
            if (suggestion.trim().startsWith('•')) {
              // eslint-disable-next-line no-console
              console.log(styleInfo(`  ${suggestion}`));
            } else if (suggestion.trim() === '') {
              // eslint-disable-next-line no-console
              console.log('');
            } else {
              // eslint-disable-next-line no-console
              console.log(styleInfo(`  ${suggestion}`));
            }
          });
        }
        
        // Show examples if available
        if (thinktankError.examples && thinktankError.examples.length > 0) {
          // eslint-disable-next-line no-console
          console.log('\n' + styleHeader('Example commands:'));
          
          thinktankError.examples.forEach(example => {
            // eslint-disable-next-line no-console
            console.log(styleSuccess(`  $ ${example}`));
          });
        }
        
        return 'No models with valid API keys available.';
      }
    }
    
    // 5. Prepare API calls
    spinner.text = `Preparing to query ${models.length} model${models.length === 1 ? '' : 's'}...`;
    
    // Create mode-specific header for the models list
    let headerText: string;
    if (options.specificModel) {
      headerText = `Using model ${styleInfo(options.specificModel)}:`;
    } else if (options.groupName) {
      headerText = `Models in group ${styleInfo(options.groupName)} (${models.length}):`;
    } else if (options.groups && options.groups.length > 0) {
      headerText = `Models in selected groups (${models.length} total):`;
    } else {
      headerText = `Models to be queried (${models.length}):`;
    }
    
    // Display header
    spinner.info(styleInfo(headerText));
    
    // Group models by their group for better display
    const modelsByGroup = new Map<string, ModelConfig[]>();
    
    // Organize models by group
    models.forEach(model => {
      const groupInfo = findModelGroup(config, model);
      const groupName = groupInfo?.groupName || 'default';
      
      if (!modelsByGroup.has(groupName)) {
        modelsByGroup.set(groupName, []);
      }
      
      modelsByGroup.get(groupName)!.push(model);
    });
    
    // Calculate API key status for all models
    const missingApiKeyModels = new Set<string>();
    missingKeyModels.forEach(model => {
      missingApiKeyModels.add(`${model.provider}:${model.modelId}`);
    });
    
    // Format and display models based on CLI mode
    const allLines: string[] = [];
    
    // If we have multiple groups, group the models in the display
    if (modelsByGroup.size > 1 && !options.specificModel) {
      // Display models grouped by their group
      let modelIndex = 1;
      
      for (const [groupName, groupModels] of modelsByGroup.entries()) {
        // Add group header with description if available
        const group = config.groups?.[groupName];
        const groupIcon = groupName === 'default' ? '⚪' : '🔶';
        const groupDescription = group?.description ? ` - ${styleDim(group.description)}` : '';
        
        allLines.push(`\n  ${groupIcon} ${styleInfo(groupName)} group${groupDescription}:`);
        
        // Add models in this group
        groupModels.forEach(model => {
          const configKey = getModelConfigKey(model);
          const icon = missingApiKeyModels.has(configKey) ? '❌' : '✅';
          const statusIcon = model.enabled ? icon : '⚫';
          
          // Format model line with status indicators
          allLines.push(`    ${modelIndex}. ${statusIcon} ${configKey}${!model.enabled ? styleDim(' (disabled)') : ''}${missingApiKeyModels.has(configKey) ? styleError(' (missing API key)') : ''}`);
          modelIndex++;
        });
      }
    } else {
      // Flat list of models (single group or specificModel mode)
      models.forEach((model, index) => {
        const configKey = getModelConfigKey(model);
        const groupInfo = findModelGroup(config, model);
        const groupName = groupInfo?.groupName || 'default';
        const showGroup = !options.specificModel && !options.groupName && groupName !== 'default';
        
        // Determine status icon based on API key and enabled status
        const icon = missingApiKeyModels.has(configKey) ? '❌' : '✅';
        const statusIcon = model.enabled ? icon : '⚫';
        
        // Format with status icons and relevant information
        allLines.push(`  ${index + 1}. ${statusIcon} ${configKey}${showGroup ? ` (${styleInfo(groupName)} group)` : ''}${!model.enabled ? styleDim(' (disabled)') : ''}${missingApiKeyModels.has(configKey) ? styleError(' (missing API key)') : ''}`);
      });
    }
    
    const formattedModelList = allLines.join('\n');
    
    // Add a legend if needed
    if (models.some(m => !m.enabled) || missingApiKeyModels.size > 0) {
      const legendItems = [];
      if (models.some(m => !m.enabled)) {
        legendItems.push(`${styleDim('⚫ = disabled')}`);
      }
      if (missingApiKeyModels.size > 0) {
        legendItems.push(`${styleError('❌ = missing API key')}`);
      }
      const legend = `\n  Legend: ${legendItems.join(', ')}`;
      // eslint-disable-next-line no-console
      console.log(formattedModelList + legend);
    } else {
      // eslint-disable-next-line no-console
      console.log(formattedModelList);
    }
    
    // Initialize status tracking
    models.forEach(model => {
      const configKey = getModelConfigKey(model);
      modelStatuses[configKey] = { status: 'pending' };
    });
    
    // Group models by their group for processing
    const modelsByGroupForProcessing = new Map<string, { models: ModelConfig[], description?: string }>();
    
    // Use the existing groups mapping we created for display
    for (const [groupName, groupModels] of modelsByGroup.entries()) {
      // Get group description if available
      let description: string | undefined;
      if (groupName !== 'default' && config.groups && config.groups[groupName]) {
        description = config.groups[groupName].description;
      }
      
      // Initialize group for processing
      modelsByGroupForProcessing.set(groupName, { models: groupModels, description });
    }
    
    // Print headers and organize models by group
    const callPromises: Array<Promise<LLMResponse & { configKey: string }>> = [];
    
    // Process each group
    for (const [groupName, groupData] of modelsByGroupForProcessing.entries()) {
      // Display group header if it wasn't already shown in the model list display
      if ((modelsByGroupForProcessing.size > 1 || groupName !== 'default') && 
          !options.specificModel && 
          (options.groupName === undefined || options.groupName === groupName)) {
        // Update spinner with group info
        spinner.text = `Processing group: ${groupName} (${groupData.models.length} models)`;
      }
      
      // Process each model in this group
      // We need to use a regular for loop instead of forEach to support async operations
      for (const model of groupData.models) {
        const provider = getProvider(model.provider);
        const configKey = getModelConfigKey(model);
        
        if (!provider) {
          // Handle provider not found error - create a resolved promise with an error response
          const errorPromise = (async () => {
            // Create more helpful provider not found error
            const { getProviderIds } = await import('../organisms/llmRegistry');
            const availableProviders = getProviderIds();
            
            // Basic error message
            const errorMessage = `Provider '${model.provider}' not found for model ${configKey}`;
            
            // Create ThinktankError with detailed suggestions
            const providerError = new ThinktankError(errorMessage);
            providerError.category = errorCategories.CONFIG;
            
            // Add specific provider suggestions
            const suggestions = [
              `Provider "${model.provider}" is not registered in the system`,
              'Providers must be imported and registered before use'
            ];
            
            // List available providers (handle undefined/null for tests)
            if (availableProviders && availableProviders.length > 0) {
              suggestions.push(`Available providers: ${availableProviders.join(', ')}`);
            } else {
              suggestions.push('No providers are currently registered');
            }
            
            // Add technical suggestions
            suggestions.push(
              'Ensure the provider module is correctly imported in the application',
              'Check src/templates/runThinktank.ts for provider imports',
              'Provider modules should be in src/molecules/llmProviders/'
            );
            
            providerError.suggestions = suggestions;
            
            // Format the error for display
            const formattedError = formatError(
              errorMessage, 
              errorCategories.CONFIG, 
              'Check your configuration and ensure the provider module is correctly imported'
            );
            
            // Show as warning but track as error
            spinner.warn(formattedError);
            
            // Store error in model status with detailed info
            modelStatuses[configKey] = { 
              status: 'error', 
              message: formattedError,
              detailedError: providerError
            };
            
            // Return error response
            return {
              provider: model.provider,
              modelId: model.modelId,
              text: '',
              error: errorMessage,
              configKey,
            };
          })();
          
          // Add the promise to our collection
          callPromises.push(errorPromise);
          
          // Skip to the next iteration in the loop
          continue;
        }
        
        // Determine which system prompt to use, with the following precedence:
        // 1. CLI override (options.systemPrompt)
        // 2. Model-specific system prompt (model.systemPrompt)
        // 3. Explicitly requested group's system prompt (options.groupName)
        // 4. Group system prompt (from the group the model belongs to)
        // 5. Default system prompt
        let systemPrompt: SystemPrompt | undefined;
        let modelGroupName: string | undefined;
        
        if (options.systemPrompt) {
          // Use CLI override
          systemPrompt = {
            text: options.systemPrompt,
            metadata: { source: 'cli-override' }
          };
        } else if (model.systemPrompt) {
          // Use model-specific system prompt
          systemPrompt = model.systemPrompt;
        } else if (options.groupName && config.groups && config.groups[options.groupName]) {
          // If a specific group was requested and exists, use its system prompt
          modelGroupName = options.groupName;
          systemPrompt = config.groups[options.groupName].systemPrompt;
        } else if (options.specificModel) {
          // If a specific model was requested, find which group it belongs to
          const groupInfo = findModelGroup(config, model);
          if (groupInfo) {
            modelGroupName = groupInfo.groupName;
            systemPrompt = groupInfo.systemPrompt;
          }
        } else {
          // For regular (non-specificModel, non-groupName) cases, find the group
          const groupInfo = findModelGroup(config, model);
          if (groupInfo) {
            modelGroupName = groupInfo.groupName;
            systemPrompt = groupInfo.systemPrompt;
          }
        }
        
        // If no system prompt was found, use a default
        if (!systemPrompt) {
          systemPrompt = {
            text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.',
            metadata: { source: 'default-fallback' }
          };
        }
        
        // Create promise for this model
        const responsePromise = provider.generate(prompt, model.modelId, model.options, systemPrompt)
          .then(response => {
            // Update status
            modelStatuses[configKey] = { status: 'success' };
            
            // Update spinner text with detailed progress
            spinner.text = formatModelProcessingStatus(models, modelStatuses, startTimeApiCalls, configKey, options);
            
            // Add group information to the response if applicable
            const responseWithGroup: LLMResponse & { configKey: string } = {
              ...response,
              configKey,
            };
            
            if (modelGroupName && systemPrompt) {
              responseWithGroup.groupInfo = {
                name: modelGroupName,
                systemPrompt
              };
            }
            
            return responseWithGroup;
          })
          .catch(error => {
            // Get error message and categorize it
            const errorMessage = error instanceof Error ? error.message : String(error);
            const errorObj = error instanceof Error ? error : new Error(String(error));
            const category = categorizeError(errorObj);
            const tip = getTroubleshootingTip(errorObj, category);
            
            // Format error message with category and tip
            const formattedError = formatError(errorMessage, category, tip);
            
            // Update status with formatted message
            modelStatuses[configKey] = { 
              status: 'error', 
              message: formattedError
            };
            
            // Update spinner text with detailed progress
            spinner.text = formatModelProcessingStatus(models, modelStatuses, startTimeApiCalls, configKey, options);
            
            // Log the error with additional model context
            console.error(styleError(`Error in model ${configKey}: ${formattedError}`));
            
            return {
              provider: model.provider,
              modelId: model.modelId,
              text: '',
              error: errorMessage,
              errorCategory: category,
              errorTip: tip,
              configKey,
            };
          });
        
        callPromises.push(responsePromise);
      }
    }
    
    // Complete model preparation timing
    timings.modelPreparation = getElapsedTime(startTimePreparation);
    
    // Initial status message with preparation timing
    const uniqueGroups = new Set<string>();
    models.forEach(model => {
      const groupInfo = findModelGroup(config, model);
      uniqueGroups.add(groupInfo?.groupName || 'default');
    });
    
    // Create mode-specific message
    let modeMessage = '';
    if (options.specificModel) {
      modeMessage = `Running prompt through model ${styleInfo(options.specificModel)}`;
    } else if (options.groupName) {
      modeMessage = `Running prompt through group ${styleInfo(options.groupName)} (${models.length} model${models.length === 1 ? '' : 's'})`;
    } else if (options.groups && options.groups.length > 0) {
      modeMessage = `Running prompt through groups [${options.groups.map(g => styleInfo(g)).join(', ')}] (${models.length} model${models.length === 1 ? '' : 's'})`;
    } else {
      modeMessage = `Running prompt through ${models.length} model${models.length === 1 ? '' : 's'} in ${uniqueGroups.size} group${uniqueGroups.size === 1 ? '' : 's'}`;
    }
    
    spinner.text = `${modeMessage}\n  Preparation: ${timings.modelPreparation}ms`;
    
    // 6. Execute calls concurrently
    const startTimeApiCalls = getCurrentTime();
    const results = await Promise.all(callPromises);
    timings.apiCalls = getElapsedTime(startTimeApiCalls);
    
    // 7. Show model completion summary
    const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
    const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
    
    // Create mode-specific completion message
    let completionMessage = '';
    if (options.specificModel) {
      completionMessage = options.specificModel;
    } else if (options.groupName) {
      completionMessage = `${options.groupName} group (${successCount + errorCount} model${successCount + errorCount === 1 ? '' : 's'})`;
    } else {
      completionMessage = `${successCount + errorCount} model${successCount + errorCount === 1 ? '' : 's'}`;
    }
    
    if (errorCount > 0) {
      // Log models with errors
      const percentage = Math.round((successCount / (successCount + errorCount)) * 100);
      spinner.warn(styleWarning(
        `Processing complete for ${completionMessage} - ${successCount} of ${successCount + errorCount} models completed successfully (${percentage}%)`
      ));
      
      // Group errors by category
      const errorsByCategory: Record<string, Array<{ model: string, message: string }>> = {};
      
      Object.entries(modelStatuses)
        .filter(([_, status]) => status.status === 'error')
        .forEach(([model, status]) => {
          let category = errorCategories.UNKNOWN;
          const message = status.message || 'Unknown error';
          
          // Try to extract category from the error message
          Object.values(errorCategories).forEach(cat => {
            if (message.includes(cat)) {
              category = cat;
            }
          });
          
          if (!errorsByCategory[category]) {
            errorsByCategory[category] = [];
          }
          
          errorsByCategory[category].push({ 
            model, 
            message: status.message || 'Unknown error'
          });
        });
      
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('Models with errors:'));
      
      // Display errors by category
      Object.entries(errorsByCategory).forEach(([category, errors]) => {
        // eslint-disable-next-line no-console
        console.log(styleWarning(`\n${category} errors (${errors.length}):`));
        
        errors.forEach(({ model, message }) => {
          // eslint-disable-next-line no-console
          console.log(styleError(`  - ${model}: ${message}`));
        });
      });
    } else {
      // Show success message with completion time
      const completionTimeText = timings.apiCalls > 1000 
        ? `${(timings.apiCalls / 1000).toFixed(2)}s` 
        : `${timings.apiCalls}ms`;
      
      spinner.succeed(styleSuccess(
        `Successfully completed ${completionMessage} in ${completionTimeText}`
      ));
    }
    
    // 8. Write individual files (now always done)
    const successRate = Math.round((results.filter(r => !r.error).length / results.length) * 100);
    const apiCallsTime = timings.apiCalls > 1000 
      ? `${(timings.apiCalls / 1000).toFixed(2)}s` 
      : `${timings.apiCalls}ms`;
    
    spinner.text = `Saving ${results.length} model response${results.length === 1 ? '' : 's'} to files...\n` +
      `  API calls completed in ${apiCallsTime}\n` +
      `  Success rate: ${styleInfo(`${successRate}%`)}`;
    
    const startTimeFileWrites = getCurrentTime();
    
    // Track stats for reporting
    let succeededWrites = 0;
    let failedWrites = 0;
    const fileWritePromises: Promise<void>[] = [];
    type FileDetail = {
      model: string;
      filename: string;
      status: 'pending' | 'success' | 'error';
      error?: string;
    };
    
    const fileDetails: FileDetail[] = [];
    
    // Simplified directory structure: all files go in the main output directory
    // No need to create group subdirectories
    
    // Process each result
    results.forEach((result) => {
      // Create sanitized filename from provider and model
      const sanitizedProvider = sanitizeFilename(result.provider);
      const sanitizedModelId = sanitizeFilename(result.modelId);
      let filename: string;
      let filePath: string;
      
      // Format the filename to include group information when relevant
      if (result.groupInfo?.name && result.groupInfo.name !== 'default' && !options.specificModel) {
        // Include group in filename but don't create subdirectories
        const sanitizedGroupName = sanitizeFilename(result.groupInfo.name);
        filename = `${sanitizedGroupName}-${sanitizedProvider}-${sanitizedModelId}.md`;
      } else {
        filename = `${sanitizedProvider}-${sanitizedModelId}.md`;
      }
      
      // All files go directly in the output directory
      filePath = path.join(outputDirectoryPath!, filename);
      
      // Format the response as Markdown
      const markdownContent = formatResponseAsMarkdown(result, options.includeMetadata);
      
      // Add to tracking array
      const fileDetail: FileDetail = {
        model: result.configKey,
        filename: filename, // Just use the filename directly
        status: 'pending'
      };
      fileDetails.push(fileDetail);
      
      // Create file write promise with error handling
      const writePromise = fs.writeFile(filePath, markdownContent)
        .then(() => {
          succeededWrites++;
          fileDetail.status = 'success';
          // Update progress with detailed information
          spinner.text = formatFileWritingStatus(
            results.length, 
            succeededWrites, 
            failedWrites, 
            startTimeFileWrites, 
            fileDetail.filename
          );
        })
        .catch((error) => {
          failedWrites++;
          fileDetail.status = 'error';
          const errorMessage = error instanceof Error ? error.message : String(error);
          fileDetail.error = errorMessage;
          // Also log the error for immediate feedback
          console.error(formatError(
            `Failed to write file ${fileDetail.filename}: ${errorMessage}`,
            errorCategories.FILESYSTEM,
            'Check your write permissions and available disk space'
          ));
          // Update progress with detailed information
          spinner.text = formatFileWritingStatus(
            results.length, 
            succeededWrites, 
            failedWrites, 
            startTimeFileWrites, 
            fileDetail.filename
          );
        });
      
      fileWritePromises.push(writePromise);
    });
    
    // Wait for all file writes to complete
    await Promise.all(fileWritePromises);
    timings.fileWrites = getElapsedTime(startTimeFileWrites);
    
    // Get the file writing time in a readable format
    const fileWriteTime = timings.fileWrites > 1000 
      ? `${(timings.fileWrites / 1000).toFixed(2)}s` 
      : `${timings.fileWrites}ms`;
    
    // Report results based on CLI mode and success/failure status
    if (failedWrites === 0) {
      const directoryDisplay = styleInfo(outputDirectoryPath || '');
      
      if (options.specificModel) {
        spinner.succeed(styleSuccess(`Model response saved to ${directoryDisplay} (${fileWriteTime})`));
      } else if (options.groupName) {
        spinner.succeed(styleSuccess(`${succeededWrites} model responses from group "${options.groupName}" saved to ${directoryDisplay} (${fileWriteTime})`));
      } else {
        spinner.succeed(styleSuccess(`All ${succeededWrites} model responses saved to ${directoryDisplay} (${fileWriteTime})`));
      }
      
      // eslint-disable-next-line no-console
      console.log(`\n${styleInfo(`📁 Output directory: ${outputDirectoryPath}`)}`);
      
      // Show models summary
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('📊 Response summary:'));
      
      // Display model count
      // eslint-disable-next-line no-console
      console.log(styleInfo(`  📝 ${succeededWrites} model response${succeededWrites === 1 ? '' : 's'} saved`));
      
      // Show model list with simplified display
      results.forEach((result, index) => {
        const configKey = result.configKey;
        const groupName = result.groupInfo?.name || 'default';
        
        const modelIcon = '🤖';
        const groupText = groupName !== 'default' ? ` (${groupName} group)` : '';
        const hasError = result.error ? ' ❌' : ' ✅';
        
        // eslint-disable-next-line no-console
        console.log(`  ${index + 1}. ${modelIcon} ${configKey}${groupText}${hasError}`);
      });
    } else {
      // Some files failed to write
      spinner.warn(styleWarning(`Completed with issues: ${succeededWrites} successful, ${failedWrites} failed writes (${fileWriteTime})`));
      // eslint-disable-next-line no-console
      console.log(`\n${styleInfo(`📁 Output directory: ${outputDirectoryPath}`)}`);
      
      // Show files with errors
      const failedFiles = fileDetails.filter(file => file.status === 'error');
      
      // Group file errors by category
      const fileErrorsByCategory: Record<string, Array<{ filename: string, error: string }>> = {};
      
      failedFiles.forEach(file => {
        const errorMessage = file.error || 'Unknown error';
        const category = categorizeError(errorMessage);
        
        if (!fileErrorsByCategory[category]) {
          fileErrorsByCategory[category] = [];
        }
        
        fileErrorsByCategory[category].push({
          filename: file.filename,
          error: errorMessage
        });
      });
      
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('❌ Files with errors:'));
      
      // Display file errors by category
      Object.entries(fileErrorsByCategory).forEach(([category, errors]) => {
        // eslint-disable-next-line no-console
        console.log(styleWarning(`\n${category} errors (${errors.length}):`));
        
        errors.forEach(({ filename, error }) => {
          const tip = getTroubleshootingTip(error, category);
          // eslint-disable-next-line no-console
          console.log(styleError(`  - ${filename}: ${error}`));
          
          if (tip) {
            // eslint-disable-next-line no-console
            console.log(styleInfo(`    💡 Tip: ${tip}`));
          }
        });
      });
    }
    
    // Calculate total execution time
    timings.total = getElapsedTime(startTimeTotal);
    
    // Log timing summary if requested
    if (options.includeMetadata) {
      // eslint-disable-next-line no-console
      console.log('\n' + styleHeader('Execution timing:'));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Total:            ${timings.total}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Config loading:   ${timings.configLoad}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Input reading:    ${timings.inputRead}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Dir creation:     ${timings.directoryCreation}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  Model preparation:${timings.modelPreparation}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  API calls:        ${timings.apiCalls}ms`));
      // eslint-disable-next-line no-console
      console.log(styleDim(`  File writing:     ${timings.fileWrites}ms`));
    }
    
    // Add timing information to results metadata
    const resultsWithTimings = results.map(result => {
      if (!result.metadata) {
        result.metadata = {};
      }
      
      result.metadata.executionTimings = { ...timings };
      return result;
    });
    
    // Always return formatted results for potential console display by CLI
    const formattedResults = formatResults(resultsWithTimings, {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors,
      // Only use table format in real CLI usage, not in tests
      useTable: process.env.NODE_ENV !== 'test',
    });
    return formattedResults;
  } catch (error) {
    spinner.fail(formatErrorWithTip(error instanceof Error ? error : 'An unknown error occurred'));
    
    if (error instanceof Error) {
      throw new ThinktankError(`Error running thinktank: ${error.message}`, error);
    }
    
    throw new ThinktankError('Unknown error running thinktank');
  }
}