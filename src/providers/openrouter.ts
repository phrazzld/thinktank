/**
 * OpenRouter provider implementation for thinktank
 */
import OpenAI from 'openai';
import axios from 'axios';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';
import { ApiError } from '../core/errors';

/**
 * OpenRouter provider error class
 * Extends ApiError for better error handling while maintaining backward compatibility
 */
export class OpenRouterProviderError extends ApiError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      providerId: 'openrouter' // Always set the provider ID to 'openrouter'
    });
    
    this.name = 'OpenRouterProviderError';
    
    // Ensure the name property is correctly set and non-enumerable
    // This ensures instanceof checks work correctly in tests
    Object.defineProperty(this, 'name', {
      value: 'OpenRouterProviderError',
      enumerable: false,
      configurable: true
    });
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
      throw new OpenRouterProviderError('OpenRouter API key is missing', {
        suggestions: [
          'Set the OPENROUTER_API_KEY environment variable with your OpenRouter API key',
          'Provide the API key when creating the OpenRouterProvider instance: new OpenRouterProvider(apiKey)',
          'Get your API key from OpenRouter at: https://openrouter.ai/keys',
          'Make sure the API key has access to the models you are trying to use'
        ],
        examples: [
          'export OPENROUTER_API_KEY=your_api_key_here',
          'const openrouterProvider = new OpenRouterProvider("your_api_key_here")',
          'OPENROUTER_API_KEY=your_api_key_here npm start'
        ]
      });
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
      if (error instanceof OpenRouterProviderError) {
        throw error; // Re-throw our own errors
      }
      
      if (error instanceof Error) {
        const errorMessage = error.message.toLowerCase();
        
        // Handle authentication errors
        if (errorMessage.includes('auth') || errorMessage.includes('api key') || 
            errorMessage.includes('unauthorized') || errorMessage.includes('invalid key')) {
          throw new OpenRouterProviderError(`Authentication failed`, {
            cause: error,
            suggestions: [
              'Verify your OpenRouter API key is correct and not expired',
              'Check that your API key has access to the requested model',
              'Ensure you have sufficient credits in your OpenRouter account',
              'Generate a new API key if the current one is not working'
            ],
            examples: [
              'export OPENROUTER_API_KEY=your_new_api_key_here',
              'const provider = new OpenRouterProvider("your_new_api_key_here")'
            ]
          });
        }
        
        // Handle rate limiting errors
        if (errorMessage.includes('rate') || errorMessage.includes('limit') || 
            errorMessage.includes('quota') || errorMessage.includes('too many') ||
            errorMessage.includes('requests')) {
          throw new OpenRouterProviderError(`Rate limit or quota exceeded`, {
            cause: error,
            suggestions: [
              'Wait a few minutes before trying again',
              'Reduce the frequency of your API requests',
              'Add more credits to your OpenRouter account',
              'Switch to a model with higher rate limits or lower costs'
            ]
          });
        }
        
        // Handle model availability errors
        if (errorMessage.includes('model') && 
            (errorMessage.includes('not found') || errorMessage.includes('unavailable') || 
             errorMessage.includes('invalid'))) {
          throw new OpenRouterProviderError(`Model unavailable or not found`, {
            cause: error,
            suggestions: [
              'Verify the model ID is correct (format: provider/model)',
              'Check if the model is available through OpenRouter',
              'The model might be temporarily unavailable or deprecated',
              'Try with a different model from the same provider'
            ],
            examples: [
              'Use "openai/gpt-4o" or "anthropic/claude-3-opus-20240229"',
              'List available models with provider.listModels(apiKey)'
            ]
          });
        }
        
        // Handle content filtering/safety errors
        if (errorMessage.includes('safety') || errorMessage.includes('harmful') || 
            errorMessage.includes('content') || errorMessage.includes('policy') || 
            errorMessage.includes('blocked')) {
          throw new OpenRouterProviderError(`Content blocked by safety filters`, {
            cause: error,
            suggestions: [
              'Your prompt may have triggered content safety filters',
              'Modify your prompt to avoid potentially sensitive topics',
              'Check for harmful or dangerous content in your prompt',
              'Different models have different safety settings and capabilities'
            ]
          });
        }
        
        // Handle token/context errors
        if (errorMessage.includes('token') || errorMessage.includes('context') || 
            errorMessage.includes('exceed') || errorMessage.includes('length') || 
            errorMessage.includes('too long')) {
          throw new OpenRouterProviderError(`Token limit or context length exceeded`, {
            cause: error,
            suggestions: [
              'Reduce the length of your prompt',
              'Use a model with a larger context window',
              'Break up your request into smaller chunks',
              'Set a lower maxTokens value to limit response length'
            ],
            examples: [
              'Try "anthropic/claude-3-opus-20240229" for longer context',
              'provider.generate(prompt, modelId, { maxTokens: 1000 })'
            ]
          });
        }
        
        // Generic error for unrecognized cases
        throw new OpenRouterProviderError(`OpenRouter API error: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check your network connection',
            'Verify your prompt follows the model\'s guidelines',
            'Ensure the model ID is correctly formatted (provider/model)',
            'Try with a different model or lower parameter settings'
          ]
        });
      }
      
      // Handle unknown errors
      throw new OpenRouterProviderError('Unknown error occurred while generating text from OpenRouter', {
        suggestions: [
          'Check your network connection',
          'Verify your environment setup',
          'Try with a simpler prompt or different model',
          'Check OpenRouter status for service disruptions'
        ]
      });
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
      if (error instanceof OpenRouterProviderError) {
        throw error; // Re-throw our own errors
      }
      
      // Handle axios errors
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        
        // Extract API error message if available
        let message = error.message;
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
        
        // Handle different HTTP status codes specifically
        switch (status) {
          case 401:
            throw new OpenRouterProviderError(`Authentication failed (Status: 401)`, {
              cause: error,
              suggestions: [
                'Verify your OpenRouter API key is correct',
                'Ensure your API key has not expired',
                'Generate a new API key if necessary',
                'Check that your API key has permission to list models'
              ],
              examples: [
                'export OPENROUTER_API_KEY=your_new_api_key_here',
                'Get a new key at https://openrouter.ai/keys'
              ]
            });
            
          case 403:
            throw new OpenRouterProviderError(`Permission denied (Status: 403)`, {
              cause: error,
              suggestions: [
                'Your API key does not have permission to access this resource',
                'Check if your OpenRouter account has the required access level',
                'Verify if your account has any restrictions or limitations',
                'Contact OpenRouter support if you believe this is an error'
              ]
            });
            
          case 429:
            throw new OpenRouterProviderError(`Rate limit exceeded (Status: 429)`, {
              cause: error,
              suggestions: [
                'You have exceeded your API rate limit or quota',
                'Wait a few minutes before trying again',
                'Add more credits to your OpenRouter account',
                'Reduce the frequency of your API requests'
              ]
            });
            
          case 404:
            throw new OpenRouterProviderError(`Resource not found (Status: 404)`, {
              cause: error,
              suggestions: [
                'The endpoint URL might have changed',
                'Check if the API version in the URL is correct',
                'Verify that OpenRouter is available in your region'
              ]
            });
            
          case 500:
          case 502:
          case 503:
          case 504:
            throw new OpenRouterProviderError(`OpenRouter server error (Status: ${status})`, {
              cause: error,
              suggestions: [
                'This is likely a temporary issue with OpenRouter\'s servers',
                'Wait a few minutes and try again',
                'Check OpenRouter status page for any reported outages',
                'Try using a different model or endpoint if available'
              ]
            });
            
          default:
            // Generic error for other status codes
            throw new OpenRouterProviderError(`OpenRouter API error listing models (Status: ${status}): ${message}`, {
              cause: error,
              suggestions: [
                'Check your network connection',
                'Verify your API key is correctly formatted',
                'Ensure your OpenRouter account is in good standing',
                'Try again after a few minutes'
              ]
            });
        }
      }
      
      // Handle other errors
      if (error instanceof Error) {
        // Check for network or connectivity issues
        const errorMessage = error.message.toLowerCase();
        
        if (errorMessage.includes('network') || errorMessage.includes('connect') || 
            errorMessage.includes('timeout') || errorMessage.includes('enotfound')) {
          throw new OpenRouterProviderError(`Network error connecting to OpenRouter API`, {
            cause: error,
            suggestions: [
              'Check your internet connection',
              'Verify that DNS resolution is working correctly',
              'Ensure your firewall or security software allows connections to OpenRouter\'s APIs',
              'If you\'re using a proxy or VPN, try disabling it temporarily'
            ]
          });
        }
        
        // Generic error for other Error types
        throw new OpenRouterProviderError(`Error listing OpenRouter models: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check your environment setup',
            'Verify that your API key is correctly formatted',
            'Ensure you have the correct permissions in your OpenRouter account'
          ]
        });
      }
      
      // Handle unknown errors
      throw new OpenRouterProviderError('Unknown error occurred while listing models from OpenRouter', {
        suggestions: [
          'Check your network connection',
          'Verify your environment setup',
          'Check OpenRouter status for service disruptions',
          'Try with a different API key or account'
        ]
      });
    }
  }
}

// Export a default instance
export const openrouterProvider = new OpenRouterProvider();