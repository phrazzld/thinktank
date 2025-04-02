/**
 * Configuration manager for loading and validating application config
 */
import { z } from 'zod';
import { fileExists, readFileContent, writeFile, getConfigFilePath } from '../utils/fileReader';
import { AppConfig, ModelConfig, ModelGroup, ModelOptions, SystemPrompt } from './types';
import { DEFAULT_CONFIG, DEFAULT_CONFIG_TEMPLATE_PATH } from './constants';
import { getApiKey as getApiKeyHelper } from '../utils/helpers';
import dotenv from 'dotenv';
import { logger } from '../utils/logger';
import path from 'path';

// Re-export getApiKey for use in other modules
export const getApiKey = getApiKeyHelper;

// Load environment variables early
dotenv.config();

/**
 * Configuration loading options
 */
export interface LoadConfigOptions {
  configPath?: string;
  mergeWithDefaults?: boolean;
}

/**
 * Configuration error class
 */
export class ConfigError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'ConfigError';
  }
}

// Zod schema for model options
const modelOptionsSchema = z.object({
  temperature: z.number().min(0).max(1).optional(),
  maxTokens: z.number().positive().int().optional(),
}).passthrough();

// Zod schema for system prompts
export const systemPromptSchema = z.object({
  text: z.string().min(1),
  metadata: z.record(z.unknown()).optional(),
});

// Zod schema for model configuration
const modelConfigSchema = z.object({
  provider: z.string().min(1),
  modelId: z.string().min(1),
  enabled: z.boolean(),
  apiKeyEnvVar: z.string().optional(),
  options: modelOptionsSchema.optional(),
  systemPrompt: systemPromptSchema.optional(),
});

// Zod schema for model groups
export const modelGroupSchema = z.object({
  name: z.string().min(1),
  systemPrompt: systemPromptSchema,
  models: z.array(modelConfigSchema).default([]), // No minimum required - empty array is valid
  description: z.string().optional(),
});

// Zod schema for application configuration
export const appConfigSchema = z.object({
  models: z.array(modelConfigSchema).default([]), // Default to empty array if not provided
  groups: z.record(z.string(), modelGroupSchema).optional(),
});

// Type definition from zod schema
export type ValidatedAppConfig = z.infer<typeof appConfigSchema>;

/**
 * Loads configuration from file system or specified path
 * 
 * By default, loads from the XDG standard location. If a file doesn't exist there,
 * a default configuration file will be created.
 * 
 * @param options - Configuration loading options
 * @returns The loaded configuration
 * @throws {ConfigError} If configuration cannot be loaded or is invalid
 */
export async function loadConfig(options: LoadConfigOptions = {}): Promise<AppConfig> {
  const { configPath, mergeWithDefaults = false } = options;
  
  try {
    let rawConfig: AppConfig;
    let configSource: string;
    
    if (configPath) {
      // Use specific provided path if given
      configSource = configPath;
      
      // Verify the file exists
      if (!await fileExists(configPath)) {
        throw new ConfigError(`Configuration file not found at specified path: ${configPath}`);
      }
      
      // Read and parse the configuration
      const configContent = await readFileContent(configPath);
      rawConfig = parseJsonSafely(configContent);
      
      logger.debug(`Loaded configuration from specified path: ${configPath}`);
    } else {
      // Use the XDG config path
      const xdgConfigPath = await getConfigFilePath();
      configSource = xdgConfigPath;
      
      // Check if the file exists
      if (await fileExists(xdgConfigPath)) {
        // Load the existing config
        const configContent = await readFileContent(xdgConfigPath);
        rawConfig = parseJsonSafely(configContent);
        
        logger.debug(`Loaded configuration from XDG path: ${xdgConfigPath}`);
      } else {
        // Create a default configuration if none exists
        logger.info(`Configuration file not found at ${xdgConfigPath}. Creating a default one.`);
        
        // Use the template if it exists, otherwise use the built-in default
        let defaultContent: string;
        
        if (await fileExists(DEFAULT_CONFIG_TEMPLATE_PATH)) {
          defaultContent = await readFileContent(DEFAULT_CONFIG_TEMPLATE_PATH);
          logger.debug(`Using template from: ${DEFAULT_CONFIG_TEMPLATE_PATH}`);
        } else {
          defaultContent = JSON.stringify(DEFAULT_CONFIG, null, 2);
          logger.debug('Using built-in default configuration');
        }
        
        // Save the default configuration
        await writeFile(xdgConfigPath, defaultContent);
        logger.debug(`Created default configuration at: ${xdgConfigPath}`);
        
        // Parse the default content
        rawConfig = parseJsonSafely(defaultContent);
      }
    }
    
    // Merge with defaults if requested
    // Note: This option is maintained for backward compatibility but is generally no longer needed
    // since we create a complete default config if none exists
    let config = rawConfig;
    if (mergeWithDefaults) {
      logger.debug('Merging with default configuration');
      config = mergeConfigs(DEFAULT_CONFIG, rawConfig);
    }
    
    // Validate configuration using Zod schema
    const validationResult = appConfigSchema.safeParse(config);
    if (!validationResult.success) {
      // Extract detailed validation errors
      const errorDetails = validationResult.error.errors
        .map(err => `${err.path.join('.')}: ${err.message}`)
        .join('; ');
        
      throw new ConfigError(`Invalid configuration in ${configSource}: ${errorDetails}`);
    }
    
    // Normalize the configuration to ensure at least a default group exists
    const normalizedConfig = normalizeConfig(validationResult.data);
    return normalizedConfig;
  } catch (error) {
    // Re-throw ConfigError instances
    if (error instanceof ConfigError) {
      throw error;
    }
    
    // Wrap other errors with context
    if (error instanceof Error) {
      throw new ConfigError(`Failed to load configuration: ${error.message}`, error);
    }
    
    // Handle unexpected non-Error exceptions
    throw new ConfigError('Unknown error loading configuration');
  }
}

/**
 * Parses JSON content safely with proper type checking
 * 
 * @param content - JSON string to parse
 * @returns Parsed AppConfig
 * @throws {ConfigError} If JSON is invalid
 */
function parseJsonSafely(content: string): AppConfig {
  try {
    // Parse the JSON content with unknown type
    const parsed = JSON.parse(content) as unknown;
    
    // Type guard for basic object validation
    if (typeof parsed !== 'object' || parsed === null) {
      throw new ConfigError('Configuration must be a JSON object');
    }
    
    // Type guard for models property
    if (!('models' in parsed) || !Array.isArray((parsed as Record<string, unknown>).models)) {
      throw new ConfigError('Configuration must contain a "models" array');
    }
    
    // At this point we know it has a models array, so it's safe to cast
    // The full validation will happen with zod later
    return parsed as AppConfig;
  } catch (error) {
    if (error instanceof ConfigError) {
      throw error;
    }
    
    if (error instanceof Error) {
      throw new ConfigError(`Failed to parse configuration JSON: ${error.message}`);
    }
    
    throw new ConfigError('Unknown error parsing configuration JSON');
  }
}

// Note: tryLoadConfigFromPaths was removed as it's no longer needed
// The loadConfig function now handles all the logic for finding and creating config files

/**
 * Merges user configuration with default configuration
 * 
 * @param defaultConfig - Default configuration
 * @param userConfig - User-provided configuration
 * @returns Merged configuration
 */
export function mergeConfigs(defaultConfig: AppConfig, userConfig: Partial<AppConfig>): AppConfig {
  // Start with a deep copy of the default config
  const mergedConfig: AppConfig = structuredClone(defaultConfig);
  
  // If user config has models, merge them
  if (userConfig.models) {
    // Create a map of existing models for faster lookup
    const modelMap = new Map<string, number>();
    mergedConfig.models.forEach((model, index) => {
      const key = `${model.provider}:${model.modelId}`;
      modelMap.set(key, index);
    });
    
    // Update existing models or add new ones
    userConfig.models.forEach(userModel => {
      const key = `${userModel.provider}:${userModel.modelId}`;
      const existingIndex = modelMap.get(key);
      
      if (existingIndex !== undefined) {
        // Update existing model
        mergedConfig.models[existingIndex] = {
          ...mergedConfig.models[existingIndex],
          ...userModel,
          options: userModel.options 
            ? { ...mergedConfig.models[existingIndex].options, ...userModel.options }
            : mergedConfig.models[existingIndex].options,
        };
      } else {
        // Add new model
        mergedConfig.models.push(userModel);
      }
    });
  }
  
  // If user config has groups, merge them
  if (userConfig.groups) {
    // Initialize groups in merged config if it doesn't exist
    if (!mergedConfig.groups) {
      mergedConfig.groups = {};
    }
    
    // Merge each group from user config
    Object.entries(userConfig.groups).forEach(([groupName, userGroup]) => {
      const existingGroup = mergedConfig.groups?.[groupName];
      
      if (existingGroup) {
        // Merge existing group
        mergedConfig.groups![groupName] = {
          ...existingGroup,
          ...userGroup,
          // Merge system prompt if both exist
          systemPrompt: userGroup.systemPrompt 
            ? { 
                ...existingGroup.systemPrompt,
                ...userGroup.systemPrompt,
              }
            : existingGroup.systemPrompt,
          // Merge models array
          models: mergeModelArrays(existingGroup.models, userGroup.models),
        };
      } else {
        // Add new group
        mergedConfig.groups![groupName] = structuredClone(userGroup);
      }
    });
  }
  
  return mergedConfig;
}

/**
 * Merges two arrays of model configurations
 * 
 * @param baseModels - Base array of model configurations
 * @param overrideModels - Override array of model configurations
 * @returns Merged array of model configurations
 */
function mergeModelArrays(baseModels: ModelConfig[], overrideModels: ModelConfig[]): ModelConfig[] {
  const result = structuredClone(baseModels);
  const modelMap = new Map<string, number>();
  
  // Create a map of existing models for faster lookup
  result.forEach((model, index) => {
    const key = `${model.provider}:${model.modelId}`;
    modelMap.set(key, index);
  });
  
  // Update existing models or add new ones
  overrideModels.forEach(overrideModel => {
    const key = `${overrideModel.provider}:${overrideModel.modelId}`;
    const existingIndex = modelMap.get(key);
    
    if (existingIndex !== undefined) {
      // Update existing model
      result[existingIndex] = {
        ...result[existingIndex],
        ...overrideModel,
        options: overrideModel.options 
          ? { ...result[existingIndex].options, ...overrideModel.options }
          : result[existingIndex].options,
      };
    } else {
      // Add new model
      result.push(structuredClone(overrideModel));
    }
  });
  
  return result;
}

/**
 * Gets all enabled models from the configuration
 * 
 * @param config - The application configuration
 * @returns Array of enabled model configurations
 */
export function getEnabledModels(config: AppConfig): ModelConfig[] {
  return config.models.filter(model => model.enabled);
}

/**
 * Filters models by provider, model ID, or combined key
 * 
 * @param config - The application configuration
 * @param filter - Provider, model ID, or combined key filter (e.g., "openai", "gpt-4o", "openai:gpt-4o")
 * @returns Filtered model configurations
 */
export function filterModels(config: AppConfig, filter: string): ModelConfig[] {
  return config.models.filter(model => {
    const combined = `${model.provider}:${model.modelId}`;
    return (
      model.provider === filter || 
      model.modelId === filter || 
      combined === filter
    );
  });
}

/**
 * Checks if all enabled models have API keys available
 * 
 * @param config - The application configuration
 * @returns Object containing valid models and missing key models
 */
export function validateModelApiKeys(config: AppConfig): {
  validModels: ModelConfig[];
  missingKeyModels: ModelConfig[];
} {
  const enabledModels = getEnabledModels(config);
  
  const validModels: ModelConfig[] = [];
  const missingKeyModels: ModelConfig[] = [];
  
  for (const model of enabledModels) {
    const apiKey = getApiKeyHelper(model);
    
    if (apiKey) {
      validModels.push(model);
    } else {
      missingKeyModels.push(model);
    }
  }
  
  return { validModels, missingKeyModels };
}

/**
 * Normalizes a configuration to include a default group if not present
 * and ensures the default group contains all enabled models
 * 
 * This ensures configurations have a consistent structure for the rest of the
 * application to work with, while maintaining backward compatibility.
 * 
 * @param config - The validated configuration to normalize
 * @returns Normalized configuration with a default group
 */
function normalizeConfig(config: AppConfig): AppConfig {
  // Create a deep copy to avoid modifying the original
  const normalizedConfig = structuredClone(config);
  
  // Initialize groups object if it doesn't exist
  normalizedConfig.groups = normalizedConfig.groups || {};
  
  // If the default group doesn't exist, create it with standard system prompt
  // This ensures there's always at least a minimal default group
  if (!normalizedConfig.groups.default) {
    normalizedConfig.groups.default = {
      name: 'default',
      systemPrompt: {
        text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information. If you are unsure about something, admit it rather than making up an answer.',
        metadata: {
          source: 'default-config-normalization'
        }
      },
      models: [],
      description: 'Default model group',
    };
  }
  
  // Note: We no longer automatically add models to the default group
  // This ensures we respect the user's explicit configuration choices
  
  return normalizedConfig;
}

/**
 * Gets models from a specific group in the configuration
 * 
 * @param config - The application configuration
 * @param groupName - The name of the group to get models from
 * @returns Array of model configurations from the specified group
 */
export function getGroup(config: AppConfig, groupName: string): ModelConfig[] {
  // If the group exists, return its models
  if (config.groups && config.groups[groupName]) {
    return config.groups[groupName].models;
  }
  
  // If the group doesn't exist, return an empty array to indicate the group wasn't found
  // This preserves the user's explicit configuration choices
  return [];
}

/**
 * Gets enabled models from a specific group in the configuration
 * 
 * @param config - The application configuration
 * @param groupName - The name of the group to get enabled models from
 * @returns Array of enabled model configurations from the specified group
 */
export function getEnabledGroupModels(config: AppConfig, groupName: string): ModelConfig[] {
  const groupModels = getGroup(config, groupName);
  return groupModels.filter(model => model.enabled);
}

/**
 * Gets enabled models from multiple groups
 * 
 * @param config - The application configuration
 * @param groupNames - Array of group names to get models from
 * @returns Array of unique enabled model configurations from all specified groups
 */
export function getEnabledModelsFromGroups(config: AppConfig, groupNames: string[]): ModelConfig[] {
  // If no groups specified, return all enabled models
  if (!groupNames || groupNames.length === 0) {
    return getEnabledModels(config);
  }

  // Use a Map to ensure unique models by provider:modelId key
  const modelMap = new Map<string, ModelConfig>();
  
  // Process each group and add its enabled models to the map
  groupNames.forEach(groupName => {
    const groupModels = getEnabledGroupModels(config, groupName);
    
    groupModels.forEach(model => {
      const key = `${model.provider}:${model.modelId}`;
      
      // Only add if not already in the map
      if (!modelMap.has(key)) {
        modelMap.set(key, { ...model });
      }
    });
  });
  
  // Return all unique models as an array
  return Array.from(modelMap.values());
}

/**
 * Finds the group a model belongs to
 * 
 * @param config - The application configuration
 * @param model - The model configuration to find
 * @returns Object containing the group name and system prompt, or undefined if not found
 */
export function findModelGroup(
  config: AppConfig, 
  model: ModelConfig
): { groupName: string; systemPrompt: SystemPrompt } | undefined {
  if (!config.groups) {
    return {
      groupName: 'default',
      systemPrompt: { text: 'You are a helpful assistant.' }
    };
  }
  
  // Check each group
  for (const [groupName, group] of Object.entries(config.groups)) {
    const isInGroup = group.models.some(
      groupModel => 
        groupModel.provider === model.provider && 
        groupModel.modelId === model.modelId
    );
    
    if (isInGroup) {
      return {
        groupName,
        systemPrompt: group.systemPrompt
      };
    }
  }
  
  // If the model is not in any group but is in the top-level models array,
  // consider it part of the default group
  const isInDefaultModels = config.models.some(
    defaultModel => 
      defaultModel.provider === model.provider && 
      defaultModel.modelId === model.modelId
  );
  
  if (isInDefaultModels) {
    return {
      groupName: 'default',
      systemPrompt: config.groups.default?.systemPrompt || { text: 'You are a helpful assistant.' }
    };
  }
  
  // Model not found in any group
  return undefined;
}

/**
 * Save configuration to a file
 * 
 * By default, saves to the XDG standard location. Ensures the configuration
 * is valid before saving to prevent corrupted configuration files.
 * 
 * @param config - The configuration to save
 * @param configPath - Optional path to the configuration file. If not provided, the XDG config path will be used.
 * @throws {ConfigError} If the configuration is invalid or cannot be saved
 */
export async function saveConfig(config: AppConfig, configPath?: string): Promise<void> {
  try {
    // Start with a deep validation of the configuration before attempting to save
    // This helps prevent corrupted configuration files
    const validationResult = appConfigSchema.safeParse(config);
    if (!validationResult.success) {
      // Extract detailed validation errors for better debugging
      const errorDetails = validationResult.error.errors
        .map(err => `${err.path.join('.')}: ${err.message}`)
        .join('; ');
        
      throw new ConfigError(`Cannot save invalid configuration: ${errorDetails}`);
    }
    
    // Use provided path or get the XDG config path
    const targetPath = configPath || await getConfigFilePath();
    logger.debug(`Preparing to save configuration to ${targetPath}`);
    
    // Normalize the configuration before saving to ensure consistency
    // This is important for maintaining a reliable file format
    const normalizedConfig = normalizeConfig(validationResult.data);
    
    // Convert configuration to pretty-printed JSON with consistent formatting
    const configJson = JSON.stringify(normalizedConfig, null, 2);
    
    // Verify config can be parsed back (sanity check)
    try {
      JSON.parse(configJson);
    } catch (parseError) {
      throw new ConfigError(`Generated configuration JSON is invalid: ${parseError instanceof Error ? parseError.message : 'Unknown error'}`);
    }
    
    // Write the configuration to the file
    // writeFile already handles directory creation via fs.mkdir
    await writeFile(targetPath, configJson);
    
    logger.info(`Configuration saved successfully to ${targetPath}`);
  } catch (error) {
    // Re-throw ConfigError instances
    if (error instanceof ConfigError) {
      throw error;
    }
    
    // Handle file system errors with specific messages
    if (error instanceof Error) {
      if ((error as NodeJS.ErrnoException).code === 'EACCES') {
        throw new ConfigError(`Permission denied when saving configuration to ${configPath || 'default location'}. Check file permissions.`, error);
      }
      
      if ((error as NodeJS.ErrnoException).code === 'ENOSPC') {
        throw new ConfigError(`Not enough disk space to save configuration to ${configPath || 'default location'}.`, error);
      }
      
      // Generic error with message
      throw new ConfigError(`Failed to save configuration: ${error.message}`, error);
    }
    
    // Fallback for non-Error exceptions
    throw new ConfigError('Unknown error occurred while saving configuration');
  }
}

/**
 * Find a model in the configuration
 * 
 * @param config - The application configuration
 * @param provider - The provider ID
 * @param modelId - The model ID
 * @returns The model configuration if found, undefined otherwise
 */
export function findModel(
  config: AppConfig,
  provider: string,
  modelId: string
): ModelConfig | undefined {
  return config.models.find(model => 
    model.provider === provider && model.modelId === modelId
  );
}

/**
 * Add or update a model in the configuration
 * 
 * @param config - The application configuration to modify
 * @param model - The model configuration to add or update
 * @returns The updated configuration
 * @throws {ConfigError} If the model configuration is invalid
 */
export function addOrUpdateModel(config: AppConfig, model: ModelConfig): AppConfig {
  try {
    // Validate the model configuration
    const modelValidation = modelConfigSchema.safeParse(model);
    if (!modelValidation.success) {
      throw new ConfigError(`Invalid model configuration: ${modelValidation.error.message}`);
    }
    
    // Make a deep copy of the config
    const updatedConfig = structuredClone(config);
    
    // Check if the model already exists
    const existingIndex = updatedConfig.models.findIndex(
      existing => existing.provider === model.provider && existing.modelId === model.modelId
    );
    
    if (existingIndex !== -1) {
      // Update existing model
      updatedConfig.models[existingIndex] = {
        ...updatedConfig.models[existingIndex],
        ...model,
        options: model.options 
          ? { ...updatedConfig.models[existingIndex].options, ...model.options }
          : updatedConfig.models[existingIndex].options,
      };
    } else {
      // Add new model
      updatedConfig.models.push(structuredClone(model));
    }
    
    return updatedConfig;
  } catch (error) {
    if (error instanceof ConfigError) {
      throw error;
    }
    
    if (error instanceof Error) {
      throw new ConfigError(`Failed to add or update model: ${error.message}`, error);
    }
    
    throw new ConfigError('Unknown error adding or updating model');
  }
}

/**
 * Remove a model from the configuration
 * 
 * @param config - The application configuration to modify
 * @param provider - The provider ID
 * @param modelId - The model ID
 * @returns The updated configuration
 * @throws {ConfigError} If the model is not found
 */
export function removeModel(
  config: AppConfig,
  provider: string,
  modelId: string
): AppConfig {
  // Make a deep copy of the config
  const updatedConfig = structuredClone(config);
  
  // Find the model
  const index = updatedConfig.models.findIndex(
    model => model.provider === provider && model.modelId === modelId
  );
  
  if (index === -1) {
    throw new ConfigError(`Model ${provider}:${modelId} not found in configuration`);
  }
  
  // Remove the model from the top-level models array
  updatedConfig.models.splice(index, 1);
  
  // Also remove the model from any groups it's part of
  if (updatedConfig.groups) {
    for (const group of Object.values(updatedConfig.groups)) {
      const groupIndex = group.models.findIndex(
        model => model.provider === provider && model.modelId === modelId
      );
      
      if (groupIndex !== -1) {
        group.models.splice(groupIndex, 1);
      }
    }
  }
  
  return updatedConfig;
}

/**
 * Get all groups in the configuration
 * 
 * @param config - The application configuration
 * @returns Array of group names
 */
export function getGroupNames(config: AppConfig): string[] {
  if (!config.groups) {
    return ['default'];
  }
  
  return Object.keys(config.groups);
}

/**
 * Add or update a group in the configuration
 * 
 * @param config - The application configuration to modify
 * @param groupName - The name of the group
 * @param groupDetails - The group details (system prompt, models, etc.)
 * @returns The updated configuration
 * @throws {ConfigError} If the group configuration is invalid
 */
export function addOrUpdateGroup(
  config: AppConfig,
  groupName: string,
  groupDetails: Partial<Omit<ModelGroup, 'name'>>
): AppConfig {
  try {
    // Validate the inputs
    if (!groupName || groupName.trim() === '') {
      throw new ConfigError('Group name cannot be empty');
    }
    
    // Make a deep copy of the config
    const updatedConfig = structuredClone(config);
    
    // Initialize groups object if it doesn't exist
    if (!updatedConfig.groups) {
      updatedConfig.groups = {};
    }
    
    // Create a complete group object by merging with existing data or defaults
    const existingGroup = updatedConfig.groups[groupName];
    
    // Default system prompt if none is provided
    const defaultSystemPrompt: SystemPrompt = {
      text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.'
    };
    
    // Use provided system prompt, or existing, or default (in that order)
    const systemPrompt: SystemPrompt = groupDetails.systemPrompt || 
      existingGroup?.systemPrompt || 
      defaultSystemPrompt;
    
    // Use provided models, or existing, or empty array (in that order)
    const models: ModelConfig[] = groupDetails.models || 
      existingGroup?.models || 
      [];
    
    // Use provided description, or existing, or default (in that order)
    const description: string = groupDetails.description || 
      existingGroup?.description || 
      `Model group: ${groupName}`;
    
    const group: ModelGroup = {
      name: groupName,
      systemPrompt,
      models,
      description
    };
    
    // Validate the group with the schema
    const groupValidation = modelGroupSchema.safeParse(group);
    if (!groupValidation.success) {
      throw new ConfigError(`Invalid group configuration: ${groupValidation.error.message}`);
    }
    
    // Add or update the group
    updatedConfig.groups[groupName] = group;
    
    return updatedConfig;
  } catch (error) {
    if (error instanceof ConfigError) {
      throw error;
    }
    
    if (error instanceof Error) {
      throw new ConfigError(`Failed to add or update group: ${error.message}`, error);
    }
    
    throw new ConfigError('Unknown error adding or updating group');
  }
}

/**
 * Remove a group from the configuration
 * 
 * @param config - The application configuration to modify
 * @param groupName - The name of the group to remove
 * @returns The updated configuration
 * @throws {ConfigError} If the group is not found
 */
export function removeGroup(config: AppConfig, groupName: string): AppConfig {
  // Make a deep copy of the config
  const updatedConfig = structuredClone(config);
  
  // Validate that groups exists
  if (!updatedConfig.groups) {
    throw new ConfigError(`No groups defined in configuration`);
  }
  
  // Check if the group exists
  if (!updatedConfig.groups[groupName]) {
    throw new ConfigError(`Group ${groupName} not found in configuration`);
  }
  
  // Cannot remove default group if it's the only one
  if (groupName === 'default' && Object.keys(updatedConfig.groups).length === 1) {
    throw new ConfigError('Cannot remove the default group as it is required');
  }
  
  // Remove the group
  delete updatedConfig.groups[groupName];
  
  return updatedConfig;
}

/**
 * Add a model to a group
 * 
 * @param config - The application configuration to modify
 * @param groupName - The name of the group
 * @param provider - The provider ID
 * @param modelId - The model ID
 * @returns The updated configuration
 * @throws {ConfigError} If the group or model is not found
 */
export function addModelToGroup(
  config: AppConfig,
  groupName: string,
  provider: string,
  modelId: string
): AppConfig {
  // Make a deep copy of the config
  const updatedConfig = structuredClone(config);
  
  // Ensure groups object exists
  if (!updatedConfig.groups) {
    updatedConfig.groups = {
      default: {
        name: 'default',
        systemPrompt: {
          text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.'
        },
        models: [],
        description: 'Default model group'
      }
    };
  }
  
  // Check if the group exists, create if it's the default group
  if (!updatedConfig.groups[groupName]) {
    if (groupName === 'default') {
      updatedConfig.groups.default = {
        name: 'default',
        systemPrompt: {
          text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.'
        },
        models: [],
        description: 'Default model group'
      };
    } else {
      throw new ConfigError(`Group ${groupName} not found in configuration`);
    }
  }
  
  // Find the model in the top-level models array
  const model = findModel(updatedConfig, provider, modelId);
  if (!model) {
    throw new ConfigError(`Model ${provider}:${modelId} not found in configuration`);
  }
  
  // Check if the model already exists in the group
  const group = updatedConfig.groups[groupName];
  const modelExists = group.models.some(
    m => m.provider === provider && m.modelId === modelId
  );
  
  if (!modelExists) {
    // Add the model to the group
    group.models.push(structuredClone(model));
  }
  
  return updatedConfig;
}

/**
 * Remove a model from a group
 * 
 * @param config - The application configuration to modify
 * @param groupName - The name of the group
 * @param provider - The provider ID
 * @param modelId - The model ID
 * @returns The updated configuration
 * @throws {ConfigError} If the group or model is not found
 */
export function removeModelFromGroup(
  config: AppConfig,
  groupName: string,
  provider: string,
  modelId: string
): AppConfig {
  // Make a deep copy of the config
  const updatedConfig = structuredClone(config);
  
  // Check if the group exists
  if (!updatedConfig.groups || !updatedConfig.groups[groupName]) {
    throw new ConfigError(`Group ${groupName} not found in configuration`);
  }
  
  // Find the model in the group
  const group = updatedConfig.groups[groupName];
  const modelIndex = group.models.findIndex(
    model => model.provider === provider && model.modelId === modelId
  );
  
  if (modelIndex === -1) {
    throw new ConfigError(`Model ${provider}:${modelId} not found in group ${groupName}`);
  }
  
  // Remove the model from the group
  group.models.splice(modelIndex, 1);
  
  return updatedConfig;
}

/**
 * Get the default configuration file path
 * 
 * @returns The path to the default configuration file
 */
export function getDefaultConfigPath(): string {
  // Return the path in the project directory for backward compatibility
  return path.resolve(process.cwd(), 'thinktank.config.json');
}

/**
 * Get the currently used configuration file path
 * 
 * This returns the XDG config path. If the file doesn't exist yet, it will 
 * still return the path where the config will be created.
 * 
 * @returns The path to the active configuration file
 */
export async function getActiveConfigPath(): Promise<string> {
  // Return the XDG config path
  return getConfigFilePath();
}

// Base default options that apply to all models
const baseDefaultOptions: ModelOptions = {
  temperature: 0.7,
  maxTokens: 1000,
};

// Provider-specific default options
const providerDefaultOptions: Record<string, ModelOptions> = {
  anthropic: {
    thinking: {
      type: 'enabled',
      budget_tokens: 10000,
    },
  },
  openai: {
    temperature: 0.7,
  },
  google: {
    temperature: 0.7,
  },
  openrouter: {
    temperature: 0.7,
  },
};

// Model-specific default options
const modelDefaultOptions: Record<string, ModelOptions> = {
  'anthropic:claude-3-opus-20240229': {
    temperature: 0.7,
    maxTokens: 4000,
  },
  'anthropic:claude-3-sonnet-20240229': {
    temperature: 0.7,
    maxTokens: 4000,
  },
  'anthropic:claude-3-haiku-20240307': {
    temperature: 0.8,
    maxTokens: 2000,
  },
  'openai:gpt-4o': {
    temperature: 0.7,
    maxTokens: 4000,
  },
  'openai:gpt-3.5-turbo': {
    temperature: 0.8,
    maxTokens: 2000,
  },
};

/**
 * Resolves model options using the cascading configuration system
 * 
 * Merges options from multiple sources in the following order (lowest to highest priority):
 * 1. Base defaults
 * 2. Provider defaults
 * 3. Model-specific defaults
 * 4. User config defaults
 * 5. Group-specific overrides
 * 6. CLI invocation overrides
 * 
 * @param provider - Provider ID
 * @param modelId - Model ID
 * @param userConfigOptions - Options defined in user config
 * @param groupOptions - Options defined in a group config
 * @param cliOptions - Options provided via CLI
 * @returns Merged options with the correct priority
 */
export function resolveModelOptions(
  provider: string,
  modelId: string,
  userConfigOptions?: ModelOptions,
  groupOptions?: ModelOptions,
  cliOptions?: ModelOptions
): ModelOptions {
  // Start with base defaults (lowest priority)
  const resolvedOptions: ModelOptions = { ...baseDefaultOptions };
  
  // Apply provider defaults if available
  if (providerDefaultOptions[provider]) {
    Object.assign(resolvedOptions, providerDefaultOptions[provider]);
  }
  
  // Apply model-specific defaults if available
  const modelKey = `${provider}:${modelId}`;
  if (modelDefaultOptions[modelKey]) {
    Object.assign(resolvedOptions, modelDefaultOptions[modelKey]);
  }
  
  // Apply user configuration options if available
  if (userConfigOptions) {
    Object.assign(resolvedOptions, userConfigOptions);
  }
  
  // Apply group-specific overrides if available
  if (groupOptions) {
    Object.assign(resolvedOptions, groupOptions);
  }
  
  // Apply CLI options (highest priority)
  if (cliOptions) {
    Object.assign(resolvedOptions, cliOptions);
  }
  
  return resolvedOptions;
}