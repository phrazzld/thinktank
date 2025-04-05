/**
 * OpenAI provider implementation for thinktank
 */
import OpenAI from 'openai';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';
import { ApiError } from '../core/errors';

/**
 * OpenAI provider error class
 * Extends ApiError for better error handling while maintaining backward compatibility
 */
export class OpenAIProviderError extends ApiError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      providerId: 'openai' // Always set the provider ID to 'openai'
    });
    
    this.name = 'OpenAIProviderError';
    
    // Ensure the name property is correctly set and non-enumerable
    // This ensures instanceof checks work correctly in tests
    Object.defineProperty(this, 'name', {
      value: 'OpenAIProviderError',
      enumerable: false,
      configurable: true
    });
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
      throw new OpenAIProviderError('OpenAI API key is missing. Set OPENAI_API_KEY environment variable or provide it when creating the provider.', {
        suggestions: [
          'Set the OPENAI_API_KEY environment variable in your shell or .env file',
          'Get an API key from the OpenAI platform: https://platform.openai.com/api-keys',
          'Provide the API key directly when creating the provider instance'
        ],
        examples: [
          'export OPENAI_API_KEY=your_api_key',
          'const provider = new OpenAIProvider("your_api_key")'
        ]
      });
    }
    
    this.client = new OpenAI({ apiKey });
    return this.client;
  }
  
  /**
   * Translates standard options format to OpenAI-specific parameters
   * 
   * @param options - Model options from the cascading configuration system
   * @param modelId - The model ID being used
   * @returns OpenAI-specific parameters
   */
  private mapOptions(options: ModelOptions, modelId: string): Record<string, unknown> {
    const params: Record<string, unknown> = {};
    
    // Map standard parameters to OpenAI-specific format
    // Note: The options should already have appropriate defaults from resolveModelOptions
    
    // Handle temperature for all models except o3-mini (which doesn't accept it)
    if (options.temperature !== undefined && modelId !== 'o3-mini') {
      params.temperature = options.temperature;
    }
    
    // Map maxTokens based on the model-specific parameter name
    if (options.maxTokens !== undefined) {
      // Handle o3-mini model specially - it requires max_completion_tokens instead of max_tokens
      if (modelId === 'o3-mini') {
        params.max_completion_tokens = options.maxTokens;
      } else {
        params.max_tokens = options.maxTokens;
      }
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
   * Generates text from the OpenAI API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @param systemPrompt - Optional system prompt to control model behavior
   * @returns The API response as an LLMResponse
   * @throws {OpenAIProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options: ModelOptions = {},  // Default to empty object if not provided
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const params = this.mapOptions(options, modelId);
      
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
        // Re-throw our own errors
        if (error instanceof OpenAIProviderError) {
          throw error;
        }
        
        // Check for specific error types based on message content
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (errorMessage.includes('rate limit') || 
            errorMessage.includes('429') || 
            errorMessage.includes('too many requests')) {
          throw new OpenAIProviderError(`Rate limit exceeded: ${error.message}`, {
            cause: error,
            suggestions: [
              'Wait before sending additional requests',
              'Implement exponential backoff in your code',
              'Reduce the frequency of requests to the OpenAI API',
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
        if (errorMessage.includes('key') && 
            (errorMessage.includes('invalid') || 
             errorMessage.includes('expired') || 
             errorMessage.includes('incorrect'))) {
          throw new OpenAIProviderError(`API key error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that your OpenAI API key is valid and not expired',
              'Ensure the API key has the correct permissions',
              'Generate a new API key in the OpenAI platform if needed'
            ],
            examples: [
              'export OPENAI_API_KEY=your_new_api_key'
            ]
          });
        }
        
        // Handle model-specific errors
        if (errorMessage.includes('model')) {
          throw new OpenAIProviderError(`Model error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that the specified model ID is correct',
              'Verify that the model is available in your OpenAI account',
              'Some models may require special access or fine-tuning'
            ],
            examples: [
              'Use a generally available model: gpt-3.5-turbo, gpt-4o',
              'Check available models with: await provider.listModels(apiKey)'
            ]
          });
        }
        
        // Handle token/context length errors
        if (errorMessage.includes('token') || 
            errorMessage.includes('context') || 
            errorMessage.includes('length')) {
          throw new OpenAIProviderError(`Token limit exceeded: ${error.message}`, {
            cause: error,
            suggestions: [
              'Reduce the length of your input prompt',
              'Reduce the max_tokens parameter in your request',
              'Use a model with a larger context window',
              'Split your content into smaller chunks'
            ]
          });
        }
        
        // Generic API error for other cases
        throw new OpenAIProviderError(`${error.message}`, {
          cause: error,
          suggestions: [
            'Check the OpenAI API documentation for more information',
            'Review the OpenAI status page for any ongoing issues',
            'Ensure your request parameters are valid'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw new OpenAIProviderError('Unknown error occurred while generating text from OpenAI', {
        suggestions: [
          'Check your network connection',
          'Verify your request parameters',
          'Try again later or contact OpenAI support if the issue persists'
        ]
      });
    }
  }

  /**
   * Lists available models from the OpenAI API
   * 
   * @param apiKey - The API key to use for authentication
   * @returns Promise resolving to array of available models
   * @throws {OpenAIProviderError} If the API call fails
   */
  public async listModels(apiKey: string): Promise<LLMAvailableModel[]> {
    try {
      // Use the provided API key directly instead of the one from the constructor
      // This allows fetching models with a different key
      const client = new OpenAI({ apiKey });
      
      // Fetch models using the SDK - this returns an AsyncIterable
      const modelsList = client.models.list();
      
      // Convert AsyncIterable to array of LLMAvailableModel
      const models: LLMAvailableModel[] = [];
      
      // Iterate through the AsyncIterable
      for await (const model of modelsList) {
        models.push({
          id: model.id,
          description: `Owned by: ${model.owned_by}`
        });
      }
      
      return models;
    } catch (error) {
      // Handle specific error cases
      if (error instanceof Error) {
        // Check for specific error types based on message content
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (errorMessage.includes('rate limit') || 
            errorMessage.includes('429') || 
            errorMessage.includes('too many requests')) {
          throw new OpenAIProviderError(`Rate limit exceeded when listing models: ${error.message}`, {
            cause: error,
            suggestions: [
              'Wait before sending additional requests',
              'Try again after a short delay',
              'Reduce the frequency of requests to the OpenAI API'
            ]
          });
        }
        
        // Handle API authentication errors
        if (errorMessage.includes('key') && 
            (errorMessage.includes('invalid') || 
             errorMessage.includes('expired') || 
             errorMessage.includes('incorrect'))) {
          throw new OpenAIProviderError(`API key error when listing models: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check that your OpenAI API key is valid and not expired',
              'Ensure the API key has the correct permissions',
              'Generate a new API key in the OpenAI platform if needed'
            ],
            examples: [
              'export OPENAI_API_KEY=your_new_api_key'
            ]
          });
        }
        
        // Generic API error for other cases
        throw new OpenAIProviderError(`Error listing OpenAI models: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check the OpenAI API documentation for more information',
            'Ensure your account has access to the OpenAI API',
            'Verify your network connection'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw new OpenAIProviderError('Unknown error occurred while listing models from OpenAI', {
        suggestions: [
          'Check your network connection',
          'Verify your API key is correct',
          'Try again later or contact OpenAI support if the issue persists'
        ]
      });
    }
  }
}

// Export a default instance
export const openaiProvider = new OpenAIProvider();