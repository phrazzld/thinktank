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
import { ApiError } from '../core/errors';

/**
 * Anthropic provider error class
 * Extends ApiError for better error handling while maintaining backward compatibility
 */
export class AnthropicProviderError extends ApiError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      providerId: 'anthropic' // Always set the provider ID to 'anthropic'
    });
    
    this.name = 'AnthropicProviderError';
    
    // Ensure the name property is correctly set and non-enumerable
    // This ensures instanceof checks work correctly in tests
    Object.defineProperty(this, 'name', {
      value: 'AnthropicProviderError',
      enumerable: false,
      configurable: true
    });
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
      throw new AnthropicProviderError('Anthropic API key is missing. Set ANTHROPIC_API_KEY environment variable or provide it when creating the provider.', {
        suggestions: [
          'Set the ANTHROPIC_API_KEY environment variable in your shell or .env file',
          'Get an API key from the Anthropic console: https://console.anthropic.com/keys',
          'Provide the API key directly when creating the provider instance'
        ],
        examples: [
          'export ANTHROPIC_API_KEY=your_api_key',
          'const provider = new AnthropicProvider("your_api_key")'
        ]
      });
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
        // Re-throw our own errors
        if (error instanceof AnthropicProviderError) {
          throw error;
        }
        
        // Check for specific error types based on message content
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (errorMessage.includes('rate limit') || errorMessage.includes('429') || errorMessage.includes('too many requests')) {
          throw new AnthropicProviderError(`Rate limit exceeded: ${error.message}`, {
            cause: error,
            suggestions: [
              'Wait before sending additional requests',
              'Implement exponential backoff in your code',
              'Reduce the frequency of requests to the Anthropic API',
              'Consider using a different model with higher rate limits'
            ],
            examples: [
              '// Example exponential backoff implementation',
              'const backoff = (retries) => Math.pow(2, retries) * 1000;',
              'for (let i = 0; i < maxRetries; i++) {',
              '  try {',
              '    return await provider.generate(prompt, modelId);',
              '  } catch (e) {',
              '    if (!isRateLimitError(e) || i === maxRetries - 1) throw e;',
              '    await new Promise(r => setTimeout(r, backoff(i)));',
              '  }',
              '}'
            ]
          });
        }
        
        // Handle API authentication errors
        if (errorMessage.includes('key') && (errorMessage.includes('invalid') || errorMessage.includes('expired'))) {
          throw new AnthropicProviderError(`API key error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that your Anthropic API key is valid and not expired',
              'Ensure the API key has the correct permissions',
              'Generate a new API key in the Anthropic console if needed'
            ],
            examples: [
              'export ANTHROPIC_API_KEY=your_new_api_key'
            ]
          });
        }
        
        // Handle model-specific errors
        if (errorMessage.includes('model')) {
          throw new AnthropicProviderError(`Model error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that the specified model ID is correct',
              'Verify that the model is available in your Anthropic account',
              'Some models may require special access or be in limited preview'
            ]
          });
        }
        
        // Generic API error for other cases
        throw new AnthropicProviderError(`${error.message}`, {
          cause: error,
          suggestions: [
            'Check the Anthropic API documentation for more information',
            'Review the Claude API status page for any ongoing issues',
            'Ensure your request parameters are valid'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw new AnthropicProviderError('Unknown error occurred while generating text from Anthropic', {
        suggestions: [
          'Check your network connection',
          'Verify your request parameters',
          'Try again later or contact Anthropic support if the issue persists'
        ]
      });
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
        // Check for specific error types based on message content
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (errorMessage.includes('rate limit') || errorMessage.includes('429') || errorMessage.includes('too many requests')) {
          throw new AnthropicProviderError(`Rate limit exceeded when listing models: ${error.message}`, {
            cause: error,
            suggestions: [
              'Wait before sending additional requests',
              'Try again after a short delay',
              'Reduce the frequency of requests to the Anthropic API'
            ]
          });
        }
        
        // Handle API authentication errors
        if (errorMessage.includes('key') && (errorMessage.includes('invalid') || errorMessage.includes('expired'))) {
          throw new AnthropicProviderError(`API key error when listing models: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that your Anthropic API key is valid and not expired',
              'Ensure the API key has the correct permissions',
              'Generate a new API key in the Anthropic console if needed'
            ],
            examples: [
              'export ANTHROPIC_API_KEY=your_new_api_key'
            ]
          });
        }
        
        // Generic API error for other cases
        throw new AnthropicProviderError(`Error listing Anthropic models: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check the Anthropic API documentation for more information',
            'Ensure your account has access to Anthropic models',
            'Verify your network connection'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw new AnthropicProviderError('Unknown error occurred while listing models from Anthropic', {
        suggestions: [
          'Check your network connection',
          'Verify your API key is correct',
          'Try again later or contact Anthropic support if the issue persists'
        ]
      });
    }
  }
}

// Export a default instance
export const anthropicProvider = new AnthropicProvider();