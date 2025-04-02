/**
 * Anthropic provider implementation for thinktank
 */
import Anthropic from '@anthropic-ai/sdk';
import { 
  LLMProvider, 
  LLMResponse, 
  ModelOptions, 
  LLMAvailableModel, 
  SystemPrompt,
  ThinkingOptions 
} from '../core/types';
import { registerProvider } from '../core/llmRegistry';

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
    
    this.client = new Anthropic({ 
      apiKey,
      maxRetries: 0 // Disable automatic retries to avoid hanging connections 
    });
    return this.client;
  }
  
  /**
   * Translates standard options format to Anthropic-specific parameters
   * 
   * @param options - Model options from the cascading configuration system
   * @returns Anthropic-specific parameters
   */
  private mapOptions(options: ModelOptions): Record<string, unknown> {
    const params: Record<string, unknown> = {};
    
    // Map standard parameters to Anthropic-specific format
    // Note: The options should already have appropriate defaults from resolveModelOptions
    
    // Map temperature directly (same scale for Anthropic)
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
   * Generates text from the Anthropic API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @param systemPrompt - Optional system prompt to control model behavior
   * @returns The API response as an LLMResponse
   * @throws {AnthropicProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options: ModelOptions = {},  // Default to empty object if not provided
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const params = this.mapOptions(options);
      
      // Prepare base parameters
      const baseParams = {
        model: modelId,
        messages: [{ role: 'user' as const, content: prompt }],
        max_tokens: options.maxTokens || 1000, // Always require max_tokens for Anthropic API
        ...(systemPrompt && { system: systemPrompt.text }),
        ...params, // For other parameters
      };
      
      // Get the response based on whether thinking is enabled
      let response;
      if (options.thinking) {
        const thinkingOpt = options.thinking as unknown as ThinkingOptions;
        
        // Force temperature to 1 when thinking is enabled - Anthropic API requirement
        const thinkingParams = {
          ...baseParams,
          temperature: 1, // Override any other temperature value
          max_tokens: options.maxTokens || 1000, // Explicitly include max_tokens even though it's in baseParams
          thinking: {
            type: 'enabled' as const,
            budget_tokens: thinkingOpt.budget_tokens || 16000
          }
        };
        
        response = await client.messages.create(thinkingParams);
      } else {
        response = await client.messages.create(baseParams);
      }
      
      // Extract the response text - handle the ContentBlock type
      let responseText = '';
      if (response.content.length > 0) {
        const firstContent = response.content[0];
        if ('text' in firstContent) {
          responseText = firstContent.text;
        }
      }
      
      // Extract thinking output if available
      // Use unknown and type assertion for safer type handling
      const responseObj = response as unknown as { thinking?: unknown };
      
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
          thinking: responseObj.thinking,
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