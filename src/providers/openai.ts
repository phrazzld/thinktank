/**
 * OpenAI provider implementation for thinktank
 */
import OpenAI from 'openai';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';
import { 
  ApiError,
  createProviderApiKeyMissingError,
  createProviderRateLimitError,
  createProviderTokenLimitError,
  createProviderContentPolicyError,
  createProviderUnknownError,
  isProviderRateLimitError,
  isProviderTokenLimitError,
  isProviderContentPolicyError,
  isProviderAuthError
} from '../core/errors';

/**
 * OpenAI provider error class - maintained for backward compatibility
 * This type alias ensures existing code that checks for OpenAIProviderError
 * will continue to work
 */
export type OpenAIProviderError = ApiError;
// Create a constructor alias that returns an ApiError
export const OpenAIProviderError = ApiError;

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
   * @throws {ApiError} If the API key is missing
   */
  private getClient(): OpenAI {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.OPENAI_API_KEY;
    
    if (!apiKey) {
      throw createProviderApiKeyMissingError(
        'openai',
        'OpenAI',
        'https://platform.openai.com/api-keys'
      );
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
        if (error instanceof ApiError) {
          throw error;
        }
        
        const errorMessage = error.message.toLowerCase();
        
        // Use factory functions and pattern detection utilities for more consistent error handling
        
        // Handle rate limit errors
        if (isProviderRateLimitError(errorMessage)) {
          throw createProviderRateLimitError('openai', 'OpenAI', error);
        }
        
        // Handle API authentication errors
        if (isProviderAuthError(errorMessage)) {
          throw new ApiError(`API key error: ${error.message}`, {
            providerId: 'openai',
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
          throw new ApiError(`Model error: ${error.message}`, {
            providerId: 'openai',
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
        if (isProviderTokenLimitError(errorMessage)) {
          throw createProviderTokenLimitError('openai', 'OpenAI', error);
        }
        
        // Handle content policy violations
        if (isProviderContentPolicyError(errorMessage)) {
          throw createProviderContentPolicyError('openai', 'OpenAI', error);
        }
        
        // Generic API error for other cases
        throw new ApiError(`${error.message}`, {
          providerId: 'openai',
          cause: error,
          suggestions: [
            'Check the OpenAI API documentation for more information',
            'Review the OpenAI status page for any ongoing issues',
            'Ensure your request parameters are valid'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw createProviderUnknownError('openai', 'OpenAI');
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
        // Re-throw our own errors
        if (error instanceof ApiError) {
          throw error;
        }
        
        const errorMessage = error.message.toLowerCase();
        
        // Handle rate limit errors
        if (isProviderRateLimitError(errorMessage)) {
          throw createProviderRateLimitError('openai', 'OpenAI', error);
        }
        
        // Handle API authentication errors
        if (isProviderAuthError(errorMessage)) {
          throw new ApiError(`API key error when listing models: ${error.message}`, {
            providerId: 'openai',
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
        throw new ApiError(`Error listing OpenAI models: ${error.message}`, {
          providerId: 'openai',
          cause: error,
          suggestions: [
            'Check the OpenAI API documentation for more information',
            'Ensure your account has access to the OpenAI API',
            'Verify your network connection'
          ]
        });
      }
      
      // Handle unknown errors (non-Error objects)
      throw createProviderUnknownError('openai', 'OpenAI');
    }
  }
}

// Export a default instance
export const openaiProvider = new OpenAIProvider();
