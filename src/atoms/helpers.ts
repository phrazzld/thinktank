/**
 * Helper functions for the Thinktank application
 */
import { ModelConfig } from './types';

/**
 * Generates a unique key for a model configuration in the format "provider:modelId"
 * Used for model identification throughout the application
 * 
 * @param config - The model configuration object
 * @returns The unique model key
 */
export function getModelConfigKey(config: ModelConfig): string {
  return `${config.provider}:${config.modelId}`;
}

/**
 * Determines the conventional environment variable name for a provider's API key
 * Example: "openai" -> "OPENAI_API_KEY"
 * 
 * @param provider - The provider name (e.g., "openai", "anthropic")
 * @returns The standard environment variable name for the provider's API key
 */
export function getDefaultApiKeyEnvVar(provider: string): string {
  // Convert to uppercase and append _API_KEY
  return `${provider.toUpperCase()}_API_KEY`;
}

/**
 * Safely extracts a model's API key from environment variables
 * First checks the custom apiKeyEnvVar if specified, then falls back to default naming
 * 
 * @param config - The model configuration
 * @returns The API key if found, undefined otherwise
 */
export function getApiKey(config: ModelConfig): string | undefined {
  // First try the custom environment variable if specified
  if (config.apiKeyEnvVar && process.env[config.apiKeyEnvVar]) {
    return process.env[config.apiKeyEnvVar];
  }
  
  // Fall back to the default environment variable pattern
  const defaultEnvVar = getDefaultApiKeyEnvVar(config.provider);
  return process.env[defaultEnvVar];
}

/**
 * Trims and normalizes a string to remove extra whitespace
 * Useful for handling user-provided prompts
 * 
 * @param text - The input text to normalize
 * @returns The normalized text
 */
export function normalizeText(text: string): string {
  // Remove leading/trailing whitespace and normalize internal whitespace
  return text.trim().replace(/\s+/g, ' ');
}