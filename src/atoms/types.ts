/**
 * Core types for the Thinktank application
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
 * Interface contract for LLM providers
 */
export interface LLMProvider {
  providerId: string;
  generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse>;
}