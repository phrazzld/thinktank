/**
 * OpenAI provider implementation for Thinktank
 */
import OpenAI from 'openai';
import { LLMProvider, LLMResponse, ModelOptions } from '../../atoms/types';
import { registerProvider } from '../../organisms/llmRegistry';

/**
 * OpenAI provider error class
 */
export class OpenAIProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'OpenAIProviderError';
  }
}

/**
 * Implements the LLMProvider interface for OpenAI
 */
export class OpenAIProvider implements LLMProvider {
  public readonly providerId = 'openai';
  private client: OpenAI | null = null;
  
  /**
   * Creates an instance of the OpenAI provider
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
   * Gets or initializes the OpenAI client
   * 
   * @returns The OpenAI client instance
   * @throws {OpenAIProviderError} If the API key is missing
   */
  private getClient(): OpenAI {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.OPENAI_API_KEY;
    
    if (!apiKey) {
      throw new OpenAIProviderError('OpenAI API key is missing. Set OPENAI_API_KEY environment variable or provide it when creating the provider.');
    }
    
    this.client = new OpenAI({ apiKey });
    return this.client;
  }
  
  /**
   * Translates generic model options to OpenAI-specific parameters
   * 
   * @param options - Generic model options
   * @returns OpenAI-specific parameters
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
    // This allows passing OpenAI-specific options from the config
    Object.entries(options).forEach(([key, value]) => {
      if (key !== 'temperature' && key !== 'maxTokens') {
        params[key] = value;
      }
    });
    
    return params;
  }
  
  /**
   * Generates text from the OpenAI API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @returns The API response as an LLMResponse
   * @throws {OpenAIProviderError} If the API call fails
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
      
      // Return formatted response
      return {
        provider: this.providerId,
        modelId,
        text: responseText,
        metadata: {
          usage: response.usage,
          model: response.model,
          id: response.id,
          created: response.created,
        },
      };
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        if (error instanceof OpenAIProviderError) {
          throw error; // Re-throw our own errors
        }
        
        // Handle OpenAI API errors
        throw new OpenAIProviderError(`OpenAI API error: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new OpenAIProviderError('Unknown error occurred while generating text from OpenAI');
    }
  }
}

// Export a default instance
export const openaiProvider = new OpenAIProvider();