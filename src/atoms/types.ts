/**
 * Core types for the thinktank application
 */

/**
 * Standard options for LLM parameters
 */
export interface ModelOptions {
  temperature?: number;
  maxTokens?: number;
  // Allow additional properties with unknown type
  [key: string]: unknown;
}

/**
 * Configuration for a specific LLM model
 */
export interface ModelConfig {
  provider: string;
  modelId: string;
  enabled: boolean;
  apiKeyEnvVar?: string;
  options?: ModelOptions;
}

/**
 * Overall application configuration
 */
export interface AppConfig {
  models: ModelConfig[];
}

/**
 * Standardized response format from LLMs
 */
export interface LLMResponse {
  provider: string;
  modelId: string;
  text: string;
  error?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Information about an available model from a provider
 */
export interface LLMAvailableModel {
  id: string; // The model ID (e.g., "claude-3-sonnet-20240229")
  description?: string; // Optional description of the model
}

/**
 * Interface contract for LLM providers
 */
export interface LLMProvider {
  providerId: string;
  generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse>;
  
  /**
   * Optional method to list available models from the provider
   * @param apiKey The API key to use for authentication
   * @returns Promise resolving to array of available models
   */
  listModels?(apiKey: string): Promise<LLMAvailableModel[]>;
}