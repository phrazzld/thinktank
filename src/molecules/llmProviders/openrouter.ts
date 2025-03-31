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
   * Translates generic model options to OpenRouter-specific parameters
   * 
   * @param options - Generic model options
   * @returns OpenRouter-specific parameters
   */
  private mapOptions(options?: ModelOptions): Record<string, unknown> {
    if (!options) {
      return {};
    }
    
    const params: Record<string, unknown> = {};
    
    // Map temperature (0-1 scale)
    if (options.temperature !== undefined) {
      params.temperature = options.temperature;
    }
    
    // Map maxTokens to max_tokens
    if (options.maxTokens !== undefined) {
      params.max_tokens = options.maxTokens;
    }
    
    // Add any other options directly
    // This allows passing OpenRouter-specific options from the config
    Object.entries(options).forEach(([key, value]) => {
      if (key !== 'temperature' && key !== 'maxTokens') {
        params[key] = value;
      }
    });
    
    return params;
  }

  /**
   * Generates text using the OpenRouter API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @returns The API response as an LLMResponse
   * @throws {OpenRouterProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const params = this.mapOptions(options);
      
      const response = await client.chat.completions.create({
        model: modelId,
        messages: [{ role: 'user', content: prompt }],
        ...params,
      });
      
      // Extract the response text
      const responseText = response.choices[0]?.message?.content || '';
      
      // Extract standard metadata
      const metadata: Record<string, unknown> = {
        usage: response.usage,
        model: response.model,
        id: response.id,
        created: response.created,
      };
      
      // Check for OpenRouter-specific fields, if any
      // Using type assertions with Object.prototype.hasOwnProperty to safely check for additional properties
      // Note: In a more complete implementation, we'd have definitive types from OpenRouter's API docs
      const responseObj = response as unknown;
      if (responseObj !== null && typeof responseObj === 'object') {
        // Check for the route property that OpenRouter might add
        if (Object.prototype.hasOwnProperty.call(responseObj, 'route')) {
          metadata.route = (responseObj as { route: unknown }).route;
        }
      }
      
      // Return formatted response
      return {
        provider: this.providerId,
        modelId,
        text: responseText,
        metadata,
      };
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        if (error instanceof OpenRouterProviderError) {
          throw error; // Re-throw our own errors
        }
        
        // Handle OpenAI/OpenRouter API errors
        throw new OpenRouterProviderError(`OpenRouter API error: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new OpenRouterProviderError('Unknown error occurred while generating text from OpenRouter');
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