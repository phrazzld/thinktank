/**
 * Core types for the thinktank application
 */

/**
 * System prompt for LLM models
 */
export interface SystemPrompt {
  text: string;
  metadata?: Record<string, unknown>;
}

/**
 * Group of models with a common system prompt
 */
export interface ModelGroup {
  name: string;
  systemPrompt: SystemPrompt;
  models: ModelConfig[];
  description?: string;
}

/**
 * Thinking capability configuration for Claude models
 */
export interface ThinkingOptions {
  type: 'enabled'; 
  budget_tokens: number;
}

/**
 * Standard options for LLM parameters
 */
export interface ModelOptions {
  temperature?: number;
  maxTokens?: number;
  thinking?: ThinkingOptions;
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
  systemPrompt?: SystemPrompt;
}

/**
 * Overall application configuration
 * 
 * Supports both legacy models array and groups object:
 * - The `models` array is treated as the default group when groups are used
 * - The `groups` object allows for defining named sets of models with specific system prompts
 */
export interface AppConfig {
  models: ModelConfig[];
  groups?: Record<string, ModelGroup>;
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
  /**
   * Optional group information if the model is part of a group
   */
  groupInfo?: {
    name: string;
    systemPrompt?: SystemPrompt;
  };
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
  
  /**
   * Generates a response from the LLM
   * 
   * @param prompt The user prompt to send to the LLM
   * @param modelId The ID of the model to use
   * @param options Optional model parameters
   * @param systemPrompt Optional system prompt to control model behavior
   * @returns Promise resolving to LLMResponse
   */
  generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions,
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse>;
  
  /**
   * Optional method to list available models from the provider
   * 
   * @param apiKey The API key to use for authentication
   * @returns Promise resolving to array of available models
   */
  listModels?(apiKey: string): Promise<LLMAvailableModel[]>;
}