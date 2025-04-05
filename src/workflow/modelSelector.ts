/**
 * Model selector module for determining which models to query
 * 
 * Encapsulates all the logic related to model selection based on CLI flags,
 * configuration settings, and group specifications.
 */
import { AppConfig, ModelConfig } from '../core/types';
import { 
  getEnabledModels, 
  getEnabledModelsFromGroups, 
  findModel
} from '../core/configManager';
import { getProviderIds } from '../core/llmRegistry';
import { ThinktankError, errorCategories } from '../core/errors';

/**
 * Error class for model selection errors
 */
export class ModelSelectionError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.CONFIG
    });
    this.name = 'ModelSelectionError';
  }
}

/**
 * Options for model selection
 */
export interface ModelSelectionOptions {
  /**
   * Array of specific model identifiers to use in provider:modelId format
   * Highest priority in the selection hierarchy
   */
  models?: string[];
  
  /**
   * Single specific model to use in provider:modelId format
   * Second priority in the selection hierarchy
   */
  specificModel?: string;
  
  /**
   * Single group name to use
   * Third priority in the selection hierarchy
   */
  groupName?: string;
  
  /**
   * Array of group names to use
   * Fourth priority in the selection hierarchy
   */
  groups?: string[];
  
  /**
   * Whether to include disabled models when explicitly requested
   * Default: true
   */
  includeDisabled?: boolean;
  
  /**
   * Whether to validate API keys for the selected models
   * Default: true
   */
  validateApiKeys?: boolean;
  
  /**
   * Whether to throw errors for invalid models/groups
   * If false, invalid models will be silently filtered out
   * Default: true
   */
  throwOnError?: boolean;
}

/**
 * Result of the model selection process
 */
export interface ModelSelectionResult {
  /**
   * The selected models
   */
  models: ModelConfig[];
  
  /**
   * Models that were selected but are missing API keys
   */
  missingApiKeyModels: ModelConfig[];
  
  /**
   * Models that were selected but are disabled
   */
  disabledModels: ModelConfig[];
  
  /**
   * Array of warnings encountered during selection
   */
  warnings: string[];
}

/**
 * Validates and processes multiple model specifiers
 * 
 * @param config - The application configuration
 * @param modelIdentifiers - Array of model identifiers in provider:modelId format
 * @param includeDisabled - Whether to include disabled models
 * @returns Object containing selected models and errors
 */
function processModelIdentifiers(
  config: AppConfig,
  modelIdentifiers: string[],
  includeDisabled = true
): {
  models: ModelConfig[];
  errors: string[];
  disabledModels: ModelConfig[];
} {
  const selectedModels: ModelConfig[] = [];
  const errors: string[] = [];
  const disabledModels: ModelConfig[] = [];
  
  // Process each model identifier
  for (const modelIdentifier of modelIdentifiers) {
    const [provider, modelId] = modelIdentifier.split(':');
    
    // Validate provider and modelId format
    if (!provider || !modelId) {
      errors.push(`Invalid model format: "${modelIdentifier}". Must use provider:modelId format (e.g., openai:gpt-4o).`);
      continue;
    }
    
    // Find the model in the configuration
    const modelConfig = findModel(config, provider, modelId);
    
    if (!modelConfig) {
      errors.push(`Model "${modelIdentifier}" not found in configuration.`);
      continue;
    }
    
    // Check if the model is disabled
    if (!modelConfig.enabled) {
      disabledModels.push(modelConfig);
      // Only add disabled models if includeDisabled is true
      if (includeDisabled) {
        selectedModels.push(modelConfig);
      }
      continue;
    }
    
    // Add the found model to our collection
    selectedModels.push(modelConfig);
  }
  
  return { models: selectedModels, errors, disabledModels };
}

/**
 * Validates and processes a single model specifier
 * 
 * @param config - The application configuration
 * @param modelIdentifier - Model identifier in provider:modelId format
 * @param includeDisabled - Whether to include disabled models
 * @returns The model configuration
 * @throws {ModelSelectionError} If the model is invalid or not found
 */
function processSingleModelIdentifier(
  config: AppConfig,
  modelIdentifier: string,
  includeDisabled = true
): ModelConfig {
  const [provider, modelId] = modelIdentifier.split(':');
  
  // Validate provider and modelId format
  if (!provider || !modelId) {
    const availableProviders = getProviderIds();
    const enabledModels = getEnabledModels(config);
    const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
    
    const error = new ModelSelectionError(
      `Invalid model format: "${modelIdentifier}". Models must be in provider:modelId format (e.g., openai:gpt-4o).`,
      {
        suggestions: [
          'Use the provider:modelId format (e.g., openai:gpt-4o)',
          `Available providers: ${availableProviders.join(', ')}`,
          availableModels.length > 0 
            ? `Available models: ${availableModels.join(', ')}` 
            : 'No enabled models found in configuration'
        ]
      }
    );
    
    throw error;
  }
  
  // Find the model in the configuration
  const modelConfig = findModel(config, provider, modelId);
  
  if (!modelConfig) {
    // Get all available models for better suggestions
    const enabledModels = getEnabledModels(config);
    const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
    
    const error = new ModelSelectionError(
      `Model "${modelIdentifier}" not found in configuration.`,
      {
        suggestions: [
          'Check that the model is correctly spelled and exists in your configuration',
          availableModels.length > 0 
            ? `Available models: ${availableModels.join(', ')}` 
            : 'No enabled models found in configuration',
          'Use "thinktank models" to list all available models'
        ]
      }
    );
    
    throw error;
  }
  
  // Return the model configuration if it's enabled or if we include disabled models
  if (modelConfig.enabled || includeDisabled) {
    return modelConfig;
  }
  
  // Throw an error if the model is disabled and we don't include disabled models
  const error = new ModelSelectionError(
    `Model "${modelIdentifier}" is disabled in configuration.`,
    {
      suggestions: [
        'Enable the model in your configuration with:',
        `  thinktank config models update ${provider} ${modelId} --enable`,
        'Or use an enabled model instead'
      ]
    }
  );
  
  throw error;
}

/**
 * Validates and processes a group name
 * 
 * @param config - The application configuration
 * @param groupName - The name of the group to use
 * @returns Array of model configurations from the group
 * @throws {ModelSelectionError} If the group is invalid or not found
 */
function processGroupName(
  config: AppConfig,
  groupName: string
): ModelConfig[] {
  // Check if the group exists
  if (!config.groups || !config.groups[groupName]) {
    // Get available groups for better suggestions
    const availableGroups = config.groups ? Object.keys(config.groups) : [];
    
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
    
    const error = new ModelSelectionError(
      `Group "${groupName}" not found in configuration.`,
      { suggestions }
    );
    
    throw error;
  }
  
  // Get all enabled models from the group
  return getEnabledModelsFromGroups(config, [groupName]);
}

/**
 * Validates API keys for the selected models
 * 
 * @param config - The application configuration
 * @param models - The models to validate
 * @returns Array of models that are missing API keys
 */
function validateApiKeys(
  _config: AppConfig,
  models: ModelConfig[]
): ModelConfig[] {
  // Extract models that need API keys and check if they're present
  const missingKeyModels: ModelConfig[] = [];
  
  for (const model of models) {
    // Get the API key from environment variables
    const apiKeyEnvVar = model.apiKeyEnvVar || `${model.provider.toUpperCase()}_API_KEY`;
    const apiKey = process.env[apiKeyEnvVar];
    
    // If the API key is missing, add the model to the missing keys list
    if (!apiKey) {
      missingKeyModels.push(model);
    }
  }
  
  return missingKeyModels;
}

/**
 * Selects which models to query based on the provided options and configuration
 * 
 * @param config - The application configuration
 * @param options - Options for model selection
 * @returns Result of the model selection process
 * @throws {ModelSelectionError} If model selection fails and throwOnError is true
 */
export function selectModels(
  config: AppConfig, 
  options: ModelSelectionOptions = {}
): ModelSelectionResult {
  const { 
    models: modelIds, 
    specificModel, 
    groupName, 
    groups,
    includeDisabled = true,
    validateApiKeys: shouldValidateApiKeys = true,
    throwOnError = true
  } = options;
  
  let selectedModels: ModelConfig[] = [];
  const warnings: string[] = [];
  const disabledModels: ModelConfig[] = [];
  
  try {
    // 1. Check if we have multiple specifically requested models (highest priority)
    if (modelIds && modelIds.length > 0) {
      // Process the model identifiers
      const { models, errors, disabledModels: disabled } = processModelIdentifiers(config, modelIds, includeDisabled);
      
      // Add any disabled models to the list
      disabledModels.push(...disabled);
      
      // Process any warnings for specific models
      if (errors.length > 0 && models.length > 0) {
        // If we have some valid models, just add warnings
        warnings.push(...errors);
      } else if (errors.length > 0 && models.length === 0) {
        // If all models had errors, throw
        if (throwOnError) {
          // Get available providers and models for better suggestions
          const availableProviders = getProviderIds();
          const enabledModels = getEnabledModels(config);
          const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
          
          const error = new ModelSelectionError(
            `None of the specified models could be used: ${errors.join(', ')}`,
            {
              suggestions: [
                'Check that you have specified valid models in provider:modelId format',
                'Make sure the models exist in your configuration and are enabled',
                `Available providers: ${availableProviders.join(', ')}`,
                availableModels.length > 0 
                  ? `Available models: ${availableModels.join(', ')}` 
                  : 'No enabled models found in configuration'
              ]
            }
          );
          
          throw error;
        } else {
          // Just add warnings if we're not throwing
          warnings.push(...errors);
        }
      }
      
      // Use the found models
      selectedModels = models;
      
      // If we're also using a group, apply it as a filter
      if (groupName && selectedModels.length > 0) {
        try {
          // Get the models in the group
          const groupModels = getEnabledModelsFromGroups(config, [groupName]);
          
          // Create a set of model keys for efficient filtering
          const groupModelKeys = new Set(
            groupModels.map(model => `${model.provider}:${model.modelId}`)
          );
          
          // Filter to only models that are both explicitly requested and in the group
          const filteredModels = selectedModels.filter(model => {
            const key = `${model.provider}:${model.modelId}`;
            const isInGroup = groupModelKeys.has(key);
            
            if (!isInGroup) {
              warnings.push(`Model "${key}" is not in group "${groupName}" and will be skipped.`);
            }
            
            return isInGroup;
          });
          
          // If we have models left after filtering, use them
          if (filteredModels.length > 0) {
            selectedModels = filteredModels;
          } else {
            const warning = `None of the specified models are in group "${groupName}". Using specified models without group filtering.`;
            warnings.push(warning);
            // Leave the selectedModels as is
          }
        } catch (error) {
          // If the group doesn't exist, add a warning but continue with the selected models
          if (error instanceof ModelSelectionError) {
            warnings.push(`Group "${groupName}" not found in configuration. Using models without group filtering.`);
          } else {
            throw error;
          }
        }
      }
    }
    // 2. Check if we have a single specific model (second priority)
    else if (specificModel) {
      try {
        // Process the single model
        const model = processSingleModelIdentifier(config, specificModel, includeDisabled);
        
        // Check if the model is disabled
        if (!model.enabled) {
          disabledModels.push(model);
          warnings.push(`Model "${specificModel}" is disabled in configuration.`);
        }
        
        // Use just this model
        selectedModels = [model];
      } catch (error) {
        if (throwOnError) {
          throw error;
        } else {
          // Add warning and select no models
          if (error instanceof Error) {
            warnings.push(error.message);
          }
        }
      }
    }
    // 3. Check if we have a group name (third priority)
    else if (groupName) {
      try {
        // Process the group name
        selectedModels = processGroupName(config, groupName);
      } catch (error) {
        if (throwOnError) {
          throw error;
        } else {
          // Add warning and use empty selection
          if (error instanceof Error) {
            warnings.push(error.message);
          }
        }
      }
    }
    // 4. Check if we have multiple groups (fourth priority)
    else if (groups && groups.length > 0) {
      // Get enabled models from all specified groups
      selectedModels = getEnabledModelsFromGroups(config, groups);
    }
    // 5. Default: use all enabled models
    else {
      // Use all enabled models
      selectedModels = getEnabledModels(config);
    }
    
    // Check if we have any models after all filtering
    if (selectedModels.length === 0) {
      const message = determineNoModelsMessage(options);
      warnings.push(message);
      
      if (throwOnError) {
        const error = new ModelSelectionError(message, {
          suggestions: [
            'Check your configuration to ensure there are enabled models',
            'You can enable models with: thinktank config models update <provider> <modelId> --enable',
            'Use "thinktank models" to list all available models'
          ]
        });
        
        throw error;
      }
    }
    
    // Validate API keys if requested
    let missingApiKeyModels: ModelConfig[] = [];
    if (shouldValidateApiKeys) {
      missingApiKeyModels = validateApiKeys(config, selectedModels);
      
      // Filter out models with missing API keys
      if (missingApiKeyModels.length > 0) {
        const modelNames = missingApiKeyModels.map(model => `${model.provider}:${model.modelId}`).join(', ');
        warnings.push(`Missing API keys for models: ${modelNames}`);
        
        // Filter out models with missing keys only if throwOnError is false
        // or if we would still have models left after filtering
        const wouldHaveModelsLeft = selectedModels.length - missingApiKeyModels.length > 0;
        
        if (!throwOnError || wouldHaveModelsLeft) {
          selectedModels = selectedModels.filter(model => 
            !missingApiKeyModels.some(m => 
              m.provider === model.provider && m.modelId === model.modelId
            )
          );
        } else if (throwOnError && !wouldHaveModelsLeft) {
          // If we wouldn't have any models left and throwOnError is true, throw
          const error = new ModelSelectionError('No models with valid API keys available.', {
            suggestions: [
              'Check that you have set the correct environment variables for your API keys',
              'You can set them in your .env file or in your environment',
              `Missing API keys for: ${modelNames}`
            ]
          });
          
          throw error;
        }
      }
    }
    
    // Return the selected models, warnings, and model lists
    return {
      models: selectedModels,
      missingApiKeyModels,
      disabledModels,
      warnings
    };
  } catch (error) {
    // If we're not throwing on errors, return an empty result with the error as a warning
    if (!throwOnError) {
      return {
        models: [],
        missingApiKeyModels: [],
        disabledModels,
        warnings: [...warnings, error instanceof Error ? error.message : String(error)]
      };
    }
    
    // Rethrow any non-ModelSelectionError errors
    if (!(error instanceof ModelSelectionError)) {
      if (error instanceof Error) {
        throw new ModelSelectionError(`Error selecting models: ${error.message}`, { cause: error });
      } else {
        throw new ModelSelectionError(`Unknown error selecting models: ${String(error)}`);
      }
    }
    
    // Rethrow ModelSelectionError as is
    throw error;
  }
}

/**
 * Determines the appropriate message for when no models are found
 * 
 * @param options - The model selection options
 * @returns Appropriate message for the specific context
 */
function determineNoModelsMessage(options: ModelSelectionOptions): string {
  const { specificModel, groupName, groups, models } = options;
  
  if (specificModel) {
    return `Specific model "${specificModel}" is not available or not enabled.`;
  } else if (groupName) {
    return `No enabled models found in the specified group: ${groupName}`;
  } else if (groups && groups.length > 0 && models && models.length > 0) {
    return 'No enabled models found matching both group and model filters.';
  } else if (groups && groups.length > 0) {
    const groupList = groups.join(', ');
    return `No enabled models found in the specified groups: ${groupList}`;
  } else if (models && models.length > 0) {
    return 'No enabled models matched the specified filters.';
  } else {
    return 'No enabled models found in configuration.';
  }
}