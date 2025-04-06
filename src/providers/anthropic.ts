/**
 * Anthropic provider implementation for thinktank
 */
import Anthropic from '@anthropic-ai/sdk';
// Import MessageParam type from Anthropic SDK
import type { MessageParam } from '@anthropic-ai/sdk/resources/messages/messages';
import { 
  LLMProvider, 
  LLMResponse, 
  ModelOptions, 
  LLMAvailableModel
} from '../core/types';
import { registerProvider } from '../core/llmRegistry';
import { ApiError } from '../core/errors';
import {
  createProviderApiKeyMissingError,
  createProviderRateLimitError,
  createProviderModelNotFoundError,
  createProviderTokenLimitError,
  createProviderContentPolicyError,
  createProviderNetworkError,
  createProviderUnknownError,
  isProviderRateLimitError,
  isProviderTokenLimitError,
  isProviderContentPolicyError,
  isProviderNetworkError
} from '../core/errors/factories/provider';

// NOTE: We've removed the custom AnthropicProviderError class
// and now use the standardized ApiError created by the provider error factories

/**
 * Implements the LLMProvider interface for Anthropic
 */
export class AnthropicProvider implements LLMProvider {
  public readonly providerId = 'anthropic';
  private client?: Anthropic;
  private apiKey?: string;
  
  /**
   * Creates a new AnthropicProvider instance
   * 
   * @param apiKey - Optional API key (falls back to environment variable)
   */
  constructor(apiKey?: string) {
    this.apiKey = apiKey;
    
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
   * @throws {ApiError} If the API key is missing
   */
  private getClient(): Anthropic {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.ANTHROPIC_API_KEY;
    
    if (!apiKey) {
      throw createProviderApiKeyMissingError(
        'anthropic',
        'Anthropic',
        'https://console.anthropic.com/keys'
      );
    }
    
    this.client = new Anthropic({ 
      apiKey,
      maxRetries: 0 // Disable automatic retries to avoid hanging connections 
    });
    return this.client;
  }

  /**
   * Generate a response from an Anthropic model
   * 
   * @param prompt - The input prompt to send to the model
   * @param modelId - The model ID (e.g., "claude-3-opus-20240229")
   * @param options - Additional options for the generation
   * @returns A properly formatted LLMResponse object
   * @throws {ApiError} When API request fails
   */
  async generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      
      // Create a message structure
      const messages: MessageParam[] = [];
      
      // Add user message (the prompt)
      messages.push({
        role: 'user' as const,
        content: prompt
      });
      
      // Set options for the API request
      const requestOptions: Anthropic.MessageCreateParamsNonStreaming = {
        model: modelId,
        max_tokens: options?.maxTokens || 1000,
        temperature: options?.temperature ?? 0.7,
        messages,
      };
      
      // Copy additional options (like topP) to requestOptions
      if (options) {
        Object.entries(options).forEach(([key, value]) => {
          if (key !== 'maxTokens' && key !== 'temperature' && key !== 'systemPrompt') {
            // Convert camelCase to snake_case for Anthropic API
            const snakeCaseKey = key.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
            (requestOptions as any)[snakeCaseKey] = value;
          }
        });
      }
      
      // Add system prompt if provided
      if (options?.systemPrompt && typeof options.systemPrompt === 'object' && 'text' in options.systemPrompt) {
        requestOptions.system = options.systemPrompt.text as string;
      }
      
      // Handle thinking feature if specified
      if (options?.thinking && typeof options.thinking === 'object' && options.thinking.type === 'enabled') {
        // Set temperature to exactly 1 as required by Anthropic API for thinking
        requestOptions.temperature = 1;
        // Include the thinking parameter directly
        (requestOptions as any).thinking = options.thinking;
      }
      
      // Request the completion
      const response = await client.messages.create(requestOptions);
      
      // Extract the response text from the first content block
      let responseText = '';
      if (response.content && response.content.length > 0) {
        responseText = response.content
          .filter(c => c.type === 'text')
          .map(c => c.text)
          .join('\n');
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
        // Re-throw our own errors
        if (error instanceof ApiError && error.providerId === 'anthropic') {
          throw error;
        }
        
        // Check for specific error types based on message content
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (isProviderRateLimitError(errorMessage)) {
          throw createProviderRateLimitError('anthropic', 'Anthropic', error);
        }
        
        // Handle API authentication errors
        if (errorMessage.includes('key') && (errorMessage.includes('invalid') || errorMessage.includes('expired'))) {
          throw new ApiError(
            `API key error: ${error.message}`,
            {
              providerId: 'anthropic',
              cause: error,
              suggestions: [
                'Check that your Anthropic API key is valid and not expired',
                'Ensure the API key has the correct permissions',
                'Generate a new API key in the Anthropic console if needed'
              ],
              examples: [
                'export ANTHROPIC_API_KEY=your_new_api_key'
              ]
            }
          );
        }
        
        // Handle model-specific errors
        if (errorMessage.includes('model')) {
          throw createProviderModelNotFoundError('anthropic', 'Anthropic', modelId);
        }
        
        // Handle token limit errors
        if (isProviderTokenLimitError(errorMessage)) {
          throw createProviderTokenLimitError('anthropic', 'Anthropic', error);
        }
        
        // Handle content policy violations
        if (isProviderContentPolicyError(errorMessage)) {
          throw createProviderContentPolicyError('anthropic', 'Anthropic', error);
        }
        
        // Handle network errors
        if (isProviderNetworkError(errorMessage)) {
          throw createProviderNetworkError('anthropic', 'Anthropic', error);
        }
        
        // Generic API error for other cases
        throw new ApiError(
          `${error.message}`,
          {
            providerId: 'anthropic',
            cause: error,
            suggestions: [
              'Check the Anthropic API documentation for more information',
              'Review the Claude API status page for any ongoing issues',
              'Ensure your request parameters are valid'
            ]
          }
        );
      }
      
      // Handle unknown errors (non-Error objects)
      throw createProviderUnknownError('anthropic', 'Anthropic');
    }
  }

  /**
   * Lists available models from the Anthropic API
   * 
   * @returns Array of available model objects
   * @throws {ApiError} When API request fails
   */
  async listModels(apiKey?: string): Promise<LLMAvailableModel[]> {
    try {
      // Store the original API key
      const originalApiKey = this.apiKey;
      
      // Use the provided API key if available
      if (apiKey) {
        this.apiKey = apiKey;
      }
      
      try {
        // Ensure we have a valid API key/client, but we don't use it directly
        this.getClient();
      } finally {
        // Restore the original API key
        this.apiKey = originalApiKey;
      }
      
      // Anthropic doesn't have a dedicated models endpoint, so we list known models
      // This could be updated when they add a proper models endpoint
      const knownModels = [
        {
          id: 'claude-3-opus-20240229',
          name: 'Claude 3 Opus',
          description: 'Anthropic\'s most powerful model for highly complex tasks'
        },
        {
          id: 'claude-3-sonnet-20240229',
          name: 'Claude 3 Sonnet',
          description: 'Anthropic\'s balanced model for enterprise workloads'
        },
        {
          id: 'claude-3-haiku-20240307',
          name: 'Claude 3 Haiku',
          description: 'Anthropic\'s fastest and most compact model for near-instant responsiveness'
        },
        {
          id: 'claude-2.0',
          name: 'Claude 2',
          description: 'Anthropic\'s previous generation flagship model'
        },
        {
          id: 'claude-instant-1.2',
          name: 'Claude Instant 1.2',
          description: 'Anthropic\'s previous generation lightweight model'
        }
      ];
      
      // Format the models into LLMAvailableModel objects
      return knownModels.map(model => ({
        id: model.id,
        name: model.name,
        provider: this.providerId,
        description: model.description,
        capabilities: {
          streamingSupport: true,
          promptTemplate: 'Default',
          systemMessageSupport: true,
          thinkingStepsSupport: false,
        },
        availability: 'available', // Assuming all listed models are available
        pricing: {
          inputPerMillionTokens: model.id.includes('3-opus') ? 15 : 
                                 model.id.includes('3-sonnet') ? 7.5 : 
                                 model.id.includes('3-haiku') ? 0.25 : 
                                 model.id.includes('claude-2') ? 8 : 1.63,
          outputPerMillionTokens: model.id.includes('3-opus') ? 75 : 
                                  model.id.includes('3-sonnet') ? 24 : 
                                  model.id.includes('3-haiku') ? 1.25 : 
                                  model.id.includes('claude-2') ? 24 : 5.51,
          currency: 'USD'
        },
        contextWindow: model.id.includes('3-opus') ? 200000 : 
                       model.id.includes('3-sonnet') ? 200000 : 
                       model.id.includes('3-haiku') ? 200000 : 
                       model.id.includes('claude-2') ? 100000 : 100000,
      }));
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        // Re-throw our own errors
        if (error instanceof ApiError && error.providerId === 'anthropic') {
          throw error;
        }
        
        // Check for specific error types
        const errorMessage = error.message.toLowerCase();
        
        // Handle API key errors
        if (errorMessage.includes('key') || errorMessage.includes('auth')) {
          throw new ApiError(`Anthropic API key error: ${error.message}`, {
            providerId: 'anthropic',
            cause: error,
            suggestions: [
              'Check that your Anthropic API key is valid',
              'Ensure the ANTHROPIC_API_KEY environment variable is correctly set',
              'Generate a new API key from the Anthropic console'
            ]
          });
        }
        
        // Handle other API errors
        throw new ApiError(`Error listing Anthropic models: ${error.message}`, {
          providerId: 'anthropic',
          cause: error,
          suggestions: [
            'Check your API key and permissions',
            'Verify your network connection',
            'The Anthropic API may be experiencing issues'
          ]
        });
      }
      
      // Handle unknown errors
      throw new ApiError('Unknown error occurred while listing Anthropic models', {
        providerId: 'anthropic',
        suggestions: [
          'Check your network connection',
          'Verify your environment setup',
          'Try again later'
        ]
      });
    }
  }
}

// Export a default instance
export const anthropicProvider = new AnthropicProvider();

// Auto-register the provider instance
try {
  registerProvider(anthropicProvider);
} catch (error) {
  // Ignore if already registered
  if (!(error instanceof Error && error.message.includes('already registered'))) {
    throw error;
  }
}