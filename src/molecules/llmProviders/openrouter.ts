/**
 * OpenRouter provider implementation for thinktank
 */
import OpenAI from 'openai';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel } from '../../atoms/types';
import { registerProvider } from '../../organisms/llmRegistry';

/**
 * OpenRouter provider error class
 */
export class OpenRouterProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'OpenRouterProviderError';
  }
}

/**
 * Implements the LLMProvider interface for OpenRouter
 * OpenRouter provides an API compatible with OpenAI's API but gives access to various models
 */
export class OpenRouterProvider implements LLMProvider {
  public readonly providerId = 'openrouter';
  private client: OpenAI | null = null;
  
  /**
   * Creates an instance of the OpenRouter provider
   * 
   * @param apiKey - Optional API key to use instead of environment variable
   */
  constructor(private readonly apiKey?: string) {
    // Auto-register this provider instance
    try {
      registerProvider(this);
    } catch (error) {
      // Ignore if already registered
      if (!(error instanceof Error && error.message.includes('already registered'))) {
        throw error;
      }
    }
  }
  
  /**
   * Gets or initializes the OpenAI client configured for OpenRouter
   * 
   * @returns The OpenAI client instance configured for OpenRouter
   * @throws {OpenRouterProviderError} If the API key is missing
   */
  private getClient(): OpenAI {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.OPENROUTER_API_KEY;
    
    if (!apiKey) {
      throw new OpenRouterProviderError('OpenRouter API key is missing. Set OPENROUTER_API_KEY environment variable or provide it when creating the provider.');
    }
    
    // Create a new OpenAI client with OpenRouter configuration
    this.client = new OpenAI({
      baseURL: 'https://openrouter.ai/api/v1',
      apiKey,
      defaultHeaders: {
        'HTTP-Referer': 'https://github.com/phrazzld/thinktank',
        'X-Title': 'thinktank CLI',
      },
    });
    
    return this.client;
  }

  /**
   * Generates text using the OpenRouter API
   * This will be implemented in the next task
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @returns The API response as an LLMResponse
   * @throws {OpenRouterProviderError} If the API call fails
   */
  public generate(
    _prompt: string,
    _modelId: string,
    _options?: ModelOptions
  ): Promise<LLMResponse> {
    try {
      // Call getClient() to ensure it's not marked as unused
      // This will be replaced with actual implementation in the next task
      this.getClient();
      
      // Placeholder implementation to satisfy the interface
      // Will be properly implemented in a future task
      return Promise.reject(new OpenRouterProviderError('OpenRouter generate method not implemented yet'));
    } catch (error) {
      // If getClient() throws an error (e.g., missing API key), propagate it
      if (error instanceof OpenRouterProviderError) {
        return Promise.reject(error);
      }
      return Promise.reject(new OpenRouterProviderError(`Unexpected error: ${error instanceof Error ? error.message : String(error)}`));
    }
  }

  /**
   * Lists available models from the OpenRouter API
   * This will be implemented in a future task
   * 
   * @param apiKey - The API key to use for authentication
   * @returns Promise resolving to array of available models
   * @throws {OpenRouterProviderError} If the API call fails
   */
  public listModels(_apiKey: string): Promise<LLMAvailableModel[]> {
    // Placeholder implementation to satisfy the interface
    // Will be properly implemented in a future task
    return Promise.reject(new OpenRouterProviderError('OpenRouter listModels method not implemented yet'));
  }
}

// Export a default instance
export const openrouterProvider = new OpenRouterProvider();