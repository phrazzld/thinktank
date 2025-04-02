/**
 * OpenRouter provider implementation for thinktank
 */
import OpenAI from 'openai';
import axios from 'axios';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';

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
      dangerouslyAllowBrowser: false,
      maxRetries: 0, // Disable automatic retries to avoid hanging
      defaultHeaders: {
        'HTTP-Referer': 'https://github.com/phrazzld/thinktank',
        'X-Title': 'thinktank CLI',
      },
    });
    
    return this.client;
  }
  
  /**
   * Translates standard options format to OpenRouter-specific parameters
   * 
   * @param options - Model options from the cascading configuration system
   * @returns OpenRouter-specific parameters
   */
  private mapOptions(options: ModelOptions): Record<string, unknown> {
    const params: Record<string, unknown> = {};
    
    // Map standard parameters to OpenRouter-specific format
    // Note: The options should already have appropriate defaults from resolveModelOptions
    
    // Map temperature directly (same scale)
    if (options.temperature !== undefined) {
      params.temperature = options.temperature;
    }
    
    // Map maxTokens to max_tokens
    if (options.maxTokens !== undefined) {
      params.max_tokens = options.maxTokens;
    }
    
    // Add all other options directly, excluding ones we've already processed
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
   * @param systemPrompt - Optional system prompt to control model behavior
   * @returns The API response as an LLMResponse
   * @throws {OpenRouterProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options: ModelOptions = {}, // Default to empty object if not provided
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const params = this.mapOptions(options);
      
      // Prepare messages array, including system prompt if provided
      const messages = [];
      if (systemPrompt) {
        messages.push({ role: 'system' as const, content: systemPrompt.text });
      }
      messages.push({ role: 'user' as const, content: prompt });
      
      const response = await client.chat.completions.create({
        model: modelId,
        messages,
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
   * 
   * @param apiKey - The API key to use for authentication
   * @returns Promise resolving to array of available models
   * @throws {OpenRouterProviderError} If the API call fails
   */
  public async listModels(apiKey: string): Promise<LLMAvailableModel[]> {
    // Define interface for OpenRouter models response
    interface OpenRouterModelData {
      id: string;
      name?: string;
      context_length?: number;
      pricing?: {
        prompt?: number;
        completion?: number;
      };
    }
    
    interface OpenRouterModelsResponse {
      data: OpenRouterModelData[];
    }
  
    try {
      // OpenRouter's models endpoint URL
      const url = 'https://openrouter.ai/api/v1/models';
      
      // Make the request with axios
      const response = await axios.get<OpenRouterModelsResponse>(url, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'HTTP-Referer': 'https://github.com/phrazzld/thinktank',
          'X-Title': 'thinktank CLI',
        },
      });
      
      // Validate response structure
      if (!response.data || !response.data.data || !Array.isArray(response.data.data)) {
        throw new OpenRouterProviderError('Invalid response format received from OpenRouter list models API');
      }
      
      // Map the response to LLMAvailableModel format
      return response.data.data.map((model) => {
        const id = model.id || 'unknown';
        
        // Build description from available fields
        let description = '';
        
        if (model.name) {
          description += model.name;
        }
        
        if (model.context_length) {
          description += description ? ` (${model.context_length} tokens)` : `Context: ${model.context_length} tokens`;
        }
        
        if (model.pricing) {
          const prompt = model.pricing.prompt;
          const completion = model.pricing.completion;
          
          if (prompt && completion) {
            description += description ? 
              ` • $${prompt}/1M prompt, $${completion}/1M completion` : 
              `Pricing: $${prompt}/1M prompt, $${completion}/1M completion`;
          }
        }
        
        return {
          id,
          description: description || `Model ID: ${id}`,
        };
      });
    } catch (error) {
      // Handle axios errors
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        // Use a simpler approach to get error message that's type-safe
        let message = error.message;
        
        // Extract API error message if available
        try {
          if (error.response?.data && 
              typeof error.response.data === 'object' && 
              error.response.data !== null) {
            
            // Try to access common error message patterns
            const data = error.response.data as { error?: { message?: string } };
            if (data.error?.message) {
              message = data.error.message;
            }
          }
        } catch (extractError) {
          // If we can't extract the message, just use the original error message
        }
        
        throw new OpenRouterProviderError(`OpenRouter API error listing models (Status: ${status}): ${message}`, error);
      }
      
      // Handle other errors
      if (error instanceof Error) {
        if (error instanceof OpenRouterProviderError) {
          throw error; // Re-throw our own errors
        }
        throw new OpenRouterProviderError(`Error listing OpenRouter models: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new OpenRouterProviderError('Unknown error occurred while listing models from OpenRouter');
    }
  }
}

// Export a default instance
export const openrouterProvider = new OpenRouterProvider();