/**
 * Configuration manager for loading and validating application config
 */
import { z } from 'zod';
import { fileExists, readFileContent } from '../molecules/fileReader';
import { AppConfig, ModelConfig } from '../atoms/types';
import { CONFIG_SEARCH_PATHS, DEFAULT_CONFIG } from '../atoms/constants';
import { getApiKey as getApiKeyHelper } from '../atoms/helpers';
import dotenv from 'dotenv';

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
  models: z.array(modelConfigSchema).min(1),
  description: z.string().optional(),
});

// Zod schema for application configuration
export const appConfigSchema = z.object({
  models: z.array(modelConfigSchema),
  groups: z.record(z.string(), modelGroupSchema).optional(),
});

// Type definition from zod schema
export type ValidatedAppConfig = z.infer<typeof appConfigSchema>;

/**
 * Loads configuration from file system or specified path
 * 
 * @param options - Configuration loading options
 * @returns The loaded configuration
 * @throws {ConfigError} If configuration cannot be loaded or is invalid
 */
export async function loadConfig(options: LoadConfigOptions = {}): Promise<AppConfig> {
  const { configPath, mergeWithDefaults = true } = options;
  
  try {
    let rawConfig: AppConfig;
    
    // If a specific config path is provided, use it
    if (configPath) {
      const exists = await fileExists(configPath);
      if (!exists) {
        throw new ConfigError(`Configuration file not found at specified path: ${configPath}`);
      }
      
      const configContent = await readFileContent(configPath);
      rawConfig = parseJsonSafely(configContent);
    } else {
      // Otherwise, try paths in order of preference
      rawConfig = await tryLoadConfigFromPaths(CONFIG_SEARCH_PATHS);
    }
    
    // Merge with defaults if requested
    const config = mergeWithDefaults 
      ? mergeConfigs(DEFAULT_CONFIG, rawConfig) 
      : rawConfig;
    
    // Validate configuration
    const validationResult = appConfigSchema.safeParse(config);
    if (!validationResult.success) {
      throw new ConfigError(`Invalid configuration: ${validationResult.error.message}`);
    }
    
    return validationResult.data;
  } catch (error) {
    if (error instanceof ConfigError) {
      throw error;
    }
    
    if (error instanceof Error) {
      throw new ConfigError('Failed to load configuration', error);
    }
    
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

/**
 * Attempts to load configuration from a list of paths
 * 
 * @param paths - Paths to try in order of preference
 * @returns The loaded configuration
 * @throws {ConfigError} If no configuration can be loaded
 */
async function tryLoadConfigFromPaths(paths: string[]): Promise<AppConfig> {
  for (const path of paths) {
    if (await fileExists(path)) {
      try {
        const configContent = await readFileContent(path);
        return parseJsonSafely(configContent);
      } catch (error) {
        // Log and try next path
        console.warn(`Failed to load config from ${path}: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
    }
  }
  
  // If no config is found, use default
  return structuredClone(DEFAULT_CONFIG);
}

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
 * Gets models from a specific group in the configuration
 * 
 * @param config - The application configuration
 * @param groupName - The name of the group to get models from
 * @returns Array of model configurations from the specified group
 */
export function getGroup(config: AppConfig, groupName: string): ModelConfig[] {
  // If looking for the default group or the config doesn't have groups,
  // return the top-level models array
  if (groupName === 'default' || !config.groups) {
    return config.models;
  }
  
  // If the group exists, return its models
  const group = config.groups[groupName];
  if (group) {
    return group.models;
  }
  
  // If the group doesn't exist, fall back to the default models array
  return config.models;
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