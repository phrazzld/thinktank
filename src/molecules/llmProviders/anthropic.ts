/**
 * Anthropic provider implementation for Thinktank
 */
import Anthropic from '@anthropic-ai/sdk';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel } from '../../atoms/types';
import { registerProvider } from '../../organisms/llmRegistry';

/**
 * Anthropic provider error class
 */
export class AnthropicProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'AnthropicProviderError';
  }
}

/**
 * Implements the LLMProvider interface for Anthropic
 */
export class AnthropicProvider implements LLMProvider {
  public readonly providerId = 'anthropic';
  private client: Anthropic | null = null;
  
  /**
   * Creates an instance of the Anthropic provider
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
   * Gets or initializes the Anthropic client
   * 
   * @returns The Anthropic client instance
   * @throws {AnthropicProviderError} If the API key is missing
   */
  private getClient(): Anthropic {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.ANTHROPIC_API_KEY;
    
    if (!apiKey) {
      throw new AnthropicProviderError('Anthropic API key is missing. Set ANTHROPIC_API_KEY environment variable or provide it when creating the provider.');
    }
    
    this.client = new Anthropic({ apiKey });
    return this.client;
  }
  
  /**
   * Translates generic model options to Anthropic-specific parameters
   * 
   * @param options - Generic model options
   * @returns Anthropic-specific parameters
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
    // This allows passing Anthropic-specific options from the config
    Object.entries(options).forEach(([key, value]) => {
      if (key !== 'temperature' && key !== 'maxTokens') {
        params[key] = value;
      }
    });
    
    return params;
  }
  
  /**
   * Generates text from the Anthropic API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @returns The API response as an LLMResponse
   * @throws {AnthropicProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const params = this.mapOptions(options);
      
      // max_tokens is required in Anthropic API
      // Note: The role must be explicitly typed as "user" or "assistant"
      const requestParams = {
        model: modelId,
        messages: [{ role: 'user' as const, content: prompt }],
        max_tokens: 1024, // Default value if not provided
        ...params,
      };
      
      const response = await client.messages.create(requestParams);
      
      // Extract the response text - handle the ContentBlock type
      let responseText = '';
      if (response.content.length > 0) {
        const firstContent = response.content[0];
        if ('text' in firstContent) {
          responseText = firstContent.text;
        }
      }
      
      // Return formatted response
      return {
        provider: this.providerId,
        modelId,
        text: responseText,
        metadata: {
          usage: response.usage,
          model: response.model,
          id: response.id,
          type: response.type,
        },
      };
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        if (error instanceof AnthropicProviderError) {
          throw error; // Re-throw our own errors
        }
        
        // Handle Anthropic API errors
        throw new AnthropicProviderError(`Anthropic API error: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new AnthropicProviderError('Unknown error occurred while generating text from Anthropic');
    }
  }

  /**
   * Lists available models from the Anthropic API
   * 
   * @param apiKey - The API key to use for authentication
   * @returns Promise resolving to array of available models
   * @throws {AnthropicProviderError} If the API call fails
   */
  public async listModels(apiKey: string): Promise<LLMAvailableModel[]> {
    try {
      // Use the provided API key directly instead of the one from the constructor
      // This allows fetching models with a different key
      const client = new Anthropic({ apiKey });
      
      // Fetch models using the SDK
      const response = await client.models.list();
      
      // Map the response to LLMAvailableModel format
      return response.data.map(model => ({
        id: model.id,
        description: model.display_name
      }));
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        throw new AnthropicProviderError(`Anthropic API error when listing models: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new AnthropicProviderError('Unknown error occurred while listing models from Anthropic');
    }
  }
}

// Export a default instance
export const anthropicProvider = new AnthropicProvider();