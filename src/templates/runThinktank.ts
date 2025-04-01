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
  divider, 
  formatError, 
  formatErrorWithTip,
  errorCategories,
  categorizeError,
  getTroubleshootingTip
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
 */
export class ThinktankError extends Error {
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
 * @returns Formatted status string
 */
/**
 * Formats a group header for display in the console
 * 
 * @param groupName - The name of the group
 * @param models - The models in the group
 * @param description - Optional group description
 * @returns Formatted header string for console output
 */
function formatGroupHeader(
  groupName: string,
  models: ModelConfig[],
  description?: string
): string {
  // Use styling functions imported at the top of the file
  
  const enabledModels = models.filter(m => m.enabled);
  
  // Create header lines
  const lines: string[] = [];
  
  // Add a blank line for separation
  lines.push('');
  
  // Format the main header with group name and model count
  const header = `Group: ${groupName} (${enabledModels.length} model${enabledModels.length === 1 ? '' : 's'})`;
  lines.push(styleHeader(header));
  
  // Add description if available
  if (description) {
    lines.push(styleDim(description));
  }
  
  // Add divider
  lines.push(divider(80));
  
  // Return the formatted header
  return lines.join('\n');
}

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
  if (percentComplete >= 10 && elapsedTime > 0) {
    const totalTimeEstimate = (elapsedTime * totalFiles) / completedWrites;
    const remainingTime = Math.round(totalTimeEstimate - elapsedTime);
    etaString = ` - ETA: ~${remainingTime}s remaining`;
  }
  
  // Build status message
  let statusMsg = `Writing files [${completedWrites}/${totalFiles}] ${percentComplete}% complete - `;
  statusMsg += `${succeededWrites} succeeded, ${failedWrites} failed`;
  
  // Add current file if provided
  if (currentFile) {
    statusMsg += ` - Current: ${currentFile}`;
  }
  
  // Add elapsed time and ETA
  statusMsg += ` - Elapsed: ${elapsedTime}s${etaString}`;
  
  return statusMsg;
}

function formatModelProcessingStatus(
  models: ModelConfig[], 
  modelStatuses: Record<string, { status: 'pending' | 'success' | 'error', message?: string }>,
  startTime: number,
  currentModel?: string
): string {
  const pendingCount = Object.values(modelStatuses).filter(s => s.status === 'pending').length;
  const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
  const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
  const totalCount = models.length;
  const completedCount = successCount + errorCount;
  const percentComplete = Math.round((completedCount / totalCount) * 100);
  
  // Format elapsed time in seconds
  const elapsedTime = Math.round((getCurrentTime() - startTime) / 1000);
  
  // Build status message
  let statusMsg = `Processing models [${completedCount}/${totalCount}] ${percentComplete}% complete - `;
  statusMsg += `${successCount} succeeded, ${errorCount} failed, ${pendingCount} pending`;
  
  // Add current model if provided
  if (currentModel) {
    statusMsg += ` - Current: ${currentModel}`;
  }
  
  // Add elapsed time
  statusMsg += ` - Elapsed: ${elapsedTime}s`;
  
  return statusMsg;
}

export async function runThinktank(options: RunOptions): Promise<string> {
  // Start timing the full execution
  const startTimeTotal = getCurrentTime();
  
  const spinner = ora('Starting thinktank...').start();
  
  // Track the output directory path for later use
  let outputDirectoryPath: string | undefined;
  
  // For tracking model statuses
  const modelStatuses: Record<string, { status: 'pending' | 'success' | 'error', message?: string }> = {};
  
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
    
    // 2.5 Create output directory - this is now always done
    // Generate the output directory path with timestamped subdirectory
    outputDirectoryPath = generateOutputDirectoryPath(options.output);
    
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
        const message = `Invalid model format: "${options.specificModel}". Use "provider:modelId" format (e.g., "openai:gpt-4o").`;
        spinner.fail(formatError(message, errorCategories.CONFIG, 'Check the model format and try again'));
        throw new ThinktankError(message);
      }
      
      // Find the model config for this specific model
      const specificModelConfig = config.models.find(model => 
        model.provider === provider && model.modelId === modelId
      );
      
      if (!specificModelConfig) {
        const message = `Model "${options.specificModel}" not found in configuration.`;
        spinner.fail(formatError(message, errorCategories.CONFIG, 'Check your configuration file and make sure the model is defined'));
        throw new ThinktankError(message);
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
        const message = `Group "${options.groupName}" not found in configuration.`;
        spinner.fail(formatError(message, errorCategories.CONFIG, 'Check your configuration file and make sure the group is defined'));
        throw new ThinktankError(message);
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
      const modelNames = missingKeyModels.map(getModelConfigKey).join(', ');
      spinner.warn(styleWarning(`Missing API keys for models: ${modelNames}`));
      
      // Filter out models with missing keys
      models = models.filter(model => 
        !missingKeyModels.some(m => 
          m.provider === model.provider && m.modelId === model.modelId
        )
      );
      
      if (models.length === 0) {
        spinner.fail(formatError(
          'No models with valid API keys available.', 
          errorCategories.API,
          'Check your environment variables or config file for API keys'
        ));
        return 'No models with valid API keys available.';
      }
    }
    
    // 5. Prepare API calls
    spinner.text = `Preparing to query ${models.length} model${models.length === 1 ? '' : 's'}...`;
    
    // List models being used
    const modelsHeader = styleInfo(`Models to be queried (${models.length}):`);
    spinner.info(modelsHeader);
    
    // Display each model on its own line
    const formattedModelList = models.map((model, index) => {
      const configKey = getModelConfigKey(model);
      const groupInfo = findModelGroup(config, model);
      const groupName = groupInfo?.groupName || 'default';
      
      // Format with bullet points and group information
      return `  ${index + 1}. ${configKey}${groupName !== 'default' ? ` (${groupName} group)` : ''}`;
    }).join('\n');
    
    // eslint-disable-next-line no-console
    console.log(formattedModelList);
    
    // Initialize status tracking
    models.forEach(model => {
      const configKey = getModelConfigKey(model);
      modelStatuses[configKey] = { status: 'pending' };
    });
    
    // Group models by their group for display and processing
    const modelsByGroup = new Map<string, { models: ModelConfig[], description?: string }>();
    
    // Function to add a model to its group
    const addModelToGroup = (model: ModelConfig): void => {
      // Find which group this model belongs to
      const groupInfo = findModelGroup(config, model);
      const groupName = groupInfo?.groupName || 'default';
      
      // Get group description if available
      let description: string | undefined;
      if (groupName !== 'default' && config.groups && config.groups[groupName]) {
        description = config.groups[groupName].description;
      }
      
      // Initialize group if not already in the map
      if (!modelsByGroup.has(groupName)) {
        modelsByGroup.set(groupName, { models: [], description });
      }
      
      // Add model to its group
      modelsByGroup.get(groupName)!.models.push(model);
    };
    
    // Organize models by group
    models.forEach(addModelToGroup);
    
    // Print headers and organize models by group
    const callPromises: Array<Promise<LLMResponse & { configKey: string }>> = [];
    
    // Process each group
    for (const [groupName, groupData] of modelsByGroup.entries()) {
      // Display group header
      if (modelsByGroup.size > 1 || groupName !== 'default') {
        // Log the group header to the console
        // eslint-disable-next-line no-console
        console.log(formatGroupHeader(groupName, groupData.models, groupData.description));
        
        // Update spinner with group info
        spinner.text = `Processing group: ${groupName} (${groupData.models.length} models)`;
      }
      
      // Process each model in this group
      groupData.models.forEach(model => {
        const provider = getProvider(model.provider);
        const configKey = getModelConfigKey(model);
        
        if (!provider) {
          const errorMessage = `Provider '${model.provider}' not found for model ${configKey}`;
          const formattedError = formatError(
            errorMessage, 
            errorCategories.CONFIG, 
            'Check your configuration and ensure the provider module is correctly imported'
          );
          
          // Show as warning but track as error
          spinner.warn(formattedError);
          
          modelStatuses[configKey] = { 
            status: 'error', 
            message: formattedError
          };
          return;
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
            spinner.text = formatModelProcessingStatus(models, modelStatuses, startTimeApiCalls, configKey);
            
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
            spinner.text = formatModelProcessingStatus(models, modelStatuses, startTimeApiCalls, configKey);
            
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
      });
    }
    
    // Complete model preparation timing
    timings.modelPreparation = getElapsedTime(startTimePreparation);
    
    // Initial status message with preparation timing
    const uniqueGroups = new Set<string>();
    models.forEach(model => {
      const groupInfo = findModelGroup(config, model);
      uniqueGroups.add(groupInfo?.groupName || 'default');
    });
    
    spinner.text = `Sending prompt to ${models.length} model${models.length === 1 ? '' : 's'}... ` +
      `(Preparation: ${timings.modelPreparation}ms, Groups: ${uniqueGroups.size})`;
    
    // 6. Execute calls concurrently
    const startTimeApiCalls = getCurrentTime();
    const results = await Promise.all(callPromises);
    timings.apiCalls = getElapsedTime(startTimeApiCalls);
    
    // 7. Show model completion summary
    const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
    const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
    
    if (errorCount > 0) {
      // Log models with errors
      spinner.warn(styleWarning(`${successCount} of ${successCount + errorCount} models completed successfully`));
      
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
      spinner.succeed(styleSuccess(`All ${successCount} models completed successfully`));
    }
    
    // 8. Write individual files (now always done)
    spinner.text = `Writing model responses to individual files... (API calls completed in ${timings.apiCalls}ms, Success rate: ${Math.round((results.filter(r => !r.error).length / results.length) * 100)}%)`;
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
    
    // First, create all needed group directories
    const groupDirs = new Set<string>();
    
    results.forEach(result => {
      if (result.groupInfo?.name && result.groupInfo.name !== 'default') {
        groupDirs.add(sanitizeFilename(result.groupInfo.name));
      }
    });
    
    // Create all group directories concurrently
    if (groupDirs.size > 0) {
      spinner.text = `Creating group directories (${groupDirs.size}): ${Array.from(groupDirs).join(', ')}...`;
      const createDirPromises = Array.from(groupDirs).map(async (groupName) => {
        const groupDir = path.join(outputDirectoryPath!, groupName);
        try {
          await fs.mkdir(groupDir, { recursive: true });
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : String(error);
          const errorObj = error instanceof Error ? error : new Error(String(error));
          const category = categorizeError(errorObj);
          const tip = getTroubleshootingTip(errorObj, category) || 
            'Check permissions and ensure parent directories exist';
          
          // eslint-disable-next-line no-console
          console.warn(formatError(
            `Could not create group directory ${groupDir}: ${errorMessage}`, 
            category, 
            tip
          ));
        }
      });
      
      await Promise.all(createDirPromises);
    }
    
    // Process each result
    results.forEach((result) => {
      // Create sanitized filename from provider and model
      const sanitizedProvider = sanitizeFilename(result.provider);
      const sanitizedModelId = sanitizeFilename(result.modelId);
      let filename: string;
      let filePath: string;
      
      // Format the filename based on whether it belongs to a group
      if (result.groupInfo?.name && result.groupInfo.name !== 'default') {
        const sanitizedGroupName = sanitizeFilename(result.groupInfo.name);
        // When in group subdirectory, no need to prefix filename with group
        filename = `${sanitizedProvider}-${sanitizedModelId}.md`;
        filePath = path.join(outputDirectoryPath!, sanitizedGroupName, filename);
      } else {
        filename = `${sanitizedProvider}-${sanitizedModelId}.md`;
        filePath = path.join(outputDirectoryPath!, filename);
      }
      
      // Format the response as Markdown
      const markdownContent = formatResponseAsMarkdown(result, options.includeMetadata);
      
      // Add to tracking array
      const fileDetail: FileDetail = {
        model: result.configKey,
        filename: result.groupInfo?.name && result.groupInfo.name !== 'default' 
          ? `${result.groupInfo.name}/${filename}` // Include group in display path
          : filename,
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
    
    // Group results by group for reporting
    const groupedResults = new Map<string, number>();
    results.forEach(result => {
      const groupName = result.groupInfo?.name || 'default';
      groupedResults.set(groupName, (groupedResults.get(groupName) || 0) + 1);
    });
    
    // Report results
    if (failedWrites === 0) {
      spinner.succeed(styleSuccess(`All ${succeededWrites} model responses written to ${outputDirectoryPath}`));
      // eslint-disable-next-line no-console
      console.log(`\n${styleInfo(`Output directory: ${outputDirectoryPath}`)}`);
      
      // If using groups, show group summary
      if (groupedResults.size > 1 || !groupedResults.has('default')) {
        // eslint-disable-next-line no-console
        console.log('\n' + styleHeader('Group summary:'));
        
        for (const [group, count] of groupedResults.entries()) {
          if (group === 'default') {
            // eslint-disable-next-line no-console
            console.log(styleDim(`  - Default group: ${count} models`));
          } else {
            // eslint-disable-next-line no-console
            console.log(styleInfo(`  - ${group} group: ${count} models`));
          }
        }
      }
    } else {
      spinner.warn(styleWarning(`Completed with issues: ${succeededWrites} successful, ${failedWrites} failed writes`));
      // eslint-disable-next-line no-console
      console.log(`\n${styleInfo(`Output directory: ${outputDirectoryPath}`)}`);
      
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
      console.log('\n' + styleHeader('Files with errors:'));
      
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
            console.log(styleInfo(`    Tip: ${tip}`));
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