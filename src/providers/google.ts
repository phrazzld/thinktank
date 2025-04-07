/**
 * Google Gemini provider implementation for thinktank
 */
import { GoogleGenerativeAI, GenerationConfig, HarmCategory, HarmBlockThreshold } from "@google/generative-ai";
import axios from 'axios';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';
import { 
  ApiError,
  createProviderApiKeyMissingError,
  createProviderRateLimitError,
  createProviderTokenLimitError,
  createProviderContentPolicyError,
  createProviderUnknownError,
  createProviderNetworkError,
  isProviderRateLimitError,
  isProviderNetworkError,
  isProviderContentPolicyError
} from '../core/errors';

/**
 * Google provider error class - maintained for backward compatibility
 * This type alias ensures existing code that checks for GoogleProviderError
 * will continue to work
 */
export type GoogleProviderError = ApiError;
// Create a constructor alias that returns an ApiError
export const GoogleProviderError = ApiError;

/**
 * Implements the LLMProvider interface for Google Gemini
 */
export class GoogleProvider implements LLMProvider {
  public readonly providerId = 'google';
  private client: GoogleGenerativeAI | null = null;
  
  /**
   * Creates an instance of the Google Gemini provider
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
   * Gets or initializes the Google Generative AI client
   * 
   * @returns The Google Generative AI client instance
   * @throws {ApiError} If the API key is missing
   */
  private getClient(): GoogleGenerativeAI {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.GEMINI_API_KEY;
    
    if (!apiKey) {
      throw createProviderApiKeyMissingError(
        'google',
        'Google',
        'https://aistudio.google.com/app/apikey'
      );
    }
    
    this.client = new GoogleGenerativeAI(apiKey);
    return this.client;
  }
  
  /**
   * Translates standard options format to Google Gemini-specific parameters
   * 
   * @param options - Model options from the cascading configuration system
   * @returns Gemini-specific generation config
   */
  private mapOptions(options: ModelOptions): GenerationConfig {
    const generationConfig: GenerationConfig = {};
    
    // Map standard parameters to Gemini-specific format
    // Note: The options should already have appropriate defaults from resolveModelOptions
    
    // Map temperature directly (same scale for Gemini)
    if (options.temperature !== undefined) {
      generationConfig.temperature = options.temperature;
    }
    
    // Map maxTokens to maxOutputTokens
    if (options.maxTokens !== undefined) {
      generationConfig.maxOutputTokens = options.maxTokens;
    }
    
    // Map Gemini-specific parameters
    if (options.topP !== undefined) {
      generationConfig.topP = options.topP as number;
    }
    
    if (options.topK !== undefined) {
      generationConfig.topK = options.topK as number;
    }
    
    // Add all other options directly, excluding ones we've already processed
    Object.entries(options).forEach(([key, value]) => {
      if (!['temperature', 'maxTokens', 'topP', 'topK'].includes(key)) {
        // Type assertion for additional config properties
        (generationConfig as Record<string, unknown>)[key] = value;
      }
    });
    
    return generationConfig;
  }
  
  /**
   * Generates text from the Google Gemini API
   * 
   * @param prompt - The prompt to send to the API
   * @param modelId - The ID of the model to use
   * @param options - Optional parameters for the request
   * @param systemPrompt - Optional system prompt to control model behavior
   * @returns The API response as an LLMResponse
   * @throws {GoogleProviderError} If the API call fails
   */
  public async generate(
    prompt: string,
    modelId: string,
    options: ModelOptions = {}, // Default to empty object if not provided
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse> {
    try {
      const client = this.getClient();
      const generationConfig = this.mapOptions(options);
      
      const model = client.getGenerativeModel({
        model: modelId,
        // Default safety settings
        safetySettings: [
          { category: HarmCategory.HARM_CATEGORY_HARASSMENT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
          { category: HarmCategory.HARM_CATEGORY_HATE_SPEECH, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
          { category: HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
          { category: HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
        ],
      });
      
      // Prepare contents array, including system prompt if compatible
      const contents = [];
      
      // Handle system prompts based on model compatibility
      if (systemPrompt) {
        // For gemini-2.0-flash, system prompts are not supported at all
        if (modelId === "gemini-2.0-flash") {
          // Add system prompt text as a prefix to user prompt for backward compatibility
          const combinedPrompt = `${systemPrompt.text}\n\n${prompt}`;
          contents.push({ role: "user", parts: [{ text: combinedPrompt }] });
        } 
        // For gemini-2.5 models, use "model" role instead of "system"
        else if (modelId.startsWith("gemini-2.5")) {
          contents.push({ role: "model", parts: [{ text: systemPrompt.text }] });
          contents.push({ role: "user", parts: [{ text: prompt }] });
        }
        // For other models, attempt standard system prompt
        else {
          try {
            contents.push({ role: "system", parts: [{ text: systemPrompt.text }] });
            contents.push({ role: "user", parts: [{ text: prompt }] });
          } catch (error) {
            // Fallback: if system role fails, use model role
            contents.length = 0; // Clear array
            contents.push({ role: "model", parts: [{ text: systemPrompt.text }] });
            contents.push({ role: "user", parts: [{ text: prompt }] });
          }
        }
      } else {
        // No system prompt, just add the user prompt
        contents.push({ role: "user", parts: [{ text: prompt }] });
      }
      
      const result = await model.generateContent({
        contents,
        generationConfig,
      });
      
      const response = result.response;
      const responseText = response.text();
      
      // Extract metadata
      const metadata: Record<string, unknown> = {
        finishReason: response.candidates?.[0]?.finishReason,
        safetyRatings: response.candidates?.[0]?.safetyRatings,
      };
      
      return {
        provider: this.providerId,
        modelId,
        text: responseText,
        metadata,
      };
    } catch (error) {
      if (error instanceof ApiError) {
        throw error; // Re-throw our own errors
      }
      
      // Detect specific error types and provide customized suggestions
      if (error instanceof Error) {
        const errorMessage = error.message.toLowerCase();
        
        // Handle authentication errors
        if (errorMessage.includes('auth') || errorMessage.includes('unauthorized') || 
            errorMessage.includes('unauthenticated') || errorMessage.includes('invalid key')) {
          throw new ApiError(`Authentication failed`, {
            providerId: 'google',
            cause: error,
            suggestions: [
              'Verify your Google API key is correct and not expired',
              'Check that your API key has access to the requested model',
              'Ensure you have quota/credits remaining in your Google AI Studio account',
              'Generate a new API key if the current one is not working'
            ],
            examples: [
              'export GEMINI_API_KEY=your_new_api_key_here',
              'const provider = new GoogleProvider("your_new_api_key_here")'
            ]
          });
        }
        
        // Handle rate limiting errors
        if (isProviderRateLimitError(errorMessage)) {
          throw createProviderRateLimitError('google', 'Google', error);
        }
        
        // Handle model availability errors
        if (errorMessage.includes('model') && 
            (errorMessage.includes('not found') || errorMessage.includes('unavailable'))) {
          throw new ApiError(`Model unavailable or not found`, {
            providerId: 'google',
            cause: error,
            suggestions: [
              'Verify the model ID is correct and available in your region',
              'Check if the model requires special access (e.g., gemini-pro-vision)',
              'Try with a different version of the model',
              'Check the Google AI Studio status page for outages'
            ],
            examples: [
              'Use "gemini-1.5-pro" instead of "gemini-pro"',
              'Use "gemini-1.0-pro" for older stable version'
            ]
          });
        }
        
        // Handle content filtering/safety errors
        if (isProviderContentPolicyError(errorMessage)) {
          throw createProviderContentPolicyError('google', 'Google', error);
        }
        
        // Handle token limit errors
        if (errorMessage.includes('token') || errorMessage.includes('length')) {
          throw createProviderTokenLimitError('google', 'Google', error);
        }
        
        // Handle network errors
        if (isProviderNetworkError(errorMessage)) {
          throw createProviderNetworkError('google', 'Google', error);
        }
        
        // Generic error for unrecognized cases
        throw new ApiError(`Google API error: ${error.message}`, {
          providerId: 'google',
          cause: error,
          suggestions: [
            'Check your network connection',
            'Verify your prompt does not exceed token limits',
            'Ensure the model ID is correct',
            'Try with a different model or lower parameter settings'
          ]
        });
      }
      
      // Handle unknown errors
      throw createProviderUnknownError('google', 'Google');
    }
  }
  
  /**
   * Lists available models from the Google Gemini API
   * 
   * @param apiKey - The API key to use for authentication
   * @returns Promise resolving to array of available models
   * @throws {GoogleProviderError} If the API call fails
   */
  public async listModels(apiKey: string): Promise<LLMAvailableModel[]> {
    try {
      const url = `https://generativelanguage.googleapis.com/v1beta/models?key=${apiKey}`;
      
      interface GoogleModel {
        name: string;
        displayName?: string;
        description?: string;
        inputTokenLimit?: number;
        outputTokenLimit?: number;
      }
      
      const response = await axios.get<{ models: GoogleModel[] }>(url);
      
      if (!response.data || !Array.isArray(response.data.models)) {
        throw new GoogleProviderError('Invalid response format received from Google list models API');
      }
      
      // Map the response to LLMAvailableModel format
      return response.data.models.map(model => ({
        id: model.name.startsWith('models/') ? model.name.substring(7) : model.name,
        description: model.displayName || model.description || 
                     `Input token limit: ${model.inputTokenLimit}, Output token limit: ${model.outputTokenLimit}`,
      }));
    } catch (error) {
      if (error instanceof GoogleProviderError) {
        throw error; // Re-throw our own errors
      }
      
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        const responseData = error.response?.data as { error?: { message?: string } } | undefined;
        const errorMessage = responseData?.error?.message || error.message;
        
        // Handle different HTTP status codes specifically
        switch (status) {
          case 401:
            throw new ApiError(`Invalid API key`, {
              providerId: 'google',
              cause: error,
              suggestions: [
                'Verify your Google API key is correct',
                'Ensure your API key has not expired',
                'Generate a new API key if necessary',
                'Check that your API key has permission to list models'
              ],
              examples: [
                'export GEMINI_API_KEY=your_new_api_key_here',
                'Get a new key at https://aistudio.google.com/app/apikey'
              ]
            });
            
          case 403:
            throw new GoogleProviderError(`Permission denied (Status: 403)`, {
              cause: error,
              suggestions: [
                'Your API key does not have permission to access this resource',
                'Check if your Google account has access to the requested models',
                'Verify you have completed any required verification steps',
                'Check if your account has any restrictions or limitations'
              ]
            });
            
          case 429:
            throw new GoogleProviderError(`Rate limit exceeded (Status: 429)`, {
              cause: error,
              suggestions: [
                'You have exceeded your quota or rate limit',
                'Wait a few minutes before trying again',
                'Request a quota increase from Google AI Studio',
                'Reduce the frequency of your API requests'
              ]
            });
            
          case 404:
            throw new GoogleProviderError(`Resource not found (Status: 404)`, {
              cause: error,
              suggestions: [
                'The requested endpoint may have changed',
                'Check if the API version in the URL is correct',
                'Verify that the Google Generative AI API is available in your region'
              ]
            });
            
          case 500:
          case 502:
          case 503:
          case 504:
            throw new GoogleProviderError(`Google API server error (Status: ${status})`, {
              cause: error,
              suggestions: [
                'This is likely a temporary issue with Google\'s servers',
                'Wait a few minutes and try again',
                'Check Google AI Studio status page for any reported outages',
                'Try using a different model or endpoint if available'
              ]
            });
            
          default:
            // Generic error for other status codes
            throw new GoogleProviderError(`Google API error listing models (Status: ${status}): ${errorMessage}`, {
              cause: error,
              suggestions: [
                'Check your network connection',
                'Verify your API key is correctly formatted',
                'Ensure your Google account is in good standing',
                'Try again after a few minutes'
              ]
            });
        }
      }
      
      if (error instanceof Error) {
        // Check for network or connectivity issues
        const errorMessage = error.message.toLowerCase();
        
        if (errorMessage.includes('network') || errorMessage.includes('connect') || 
            errorMessage.includes('timeout') || errorMessage.includes('enotfound')) {
          throw new GoogleProviderError(`Network error connecting to Google API`, {
            cause: error,
            suggestions: [
              'Check your internet connection',
              'Verify that DNS resolution is working correctly',
              'Ensure your firewall or security software allows connections to Google\'s APIs',
              'If you\'re using a proxy or VPN, try disabling it temporarily'
            ]
          });
        }
        
        // Handle rate limiting errors
        if (errorMessage.includes('rate') || errorMessage.includes('limit') || 
            errorMessage.includes('quota') || errorMessage.includes('capacity')) {
          throw new GoogleProviderError(`Error listing Google models: ${error.message}`, {
            cause: error,
            suggestions: [
              'You may have exceeded your API rate limit or quota',
              'Wait a few minutes before trying again',
              'Request a quota increase from Google AI Studio',
              'Reduce the frequency of your API requests'
            ]
          });
        }
        
        // Generic error for other Error types
        throw new GoogleProviderError(`Error listing Google models: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check your environment setup',
            'Verify that your API key is correctly formatted',
            'Ensure you have the correct permissions in your Google account'
          ]
        });
      }
      
      // Handle unknown errors
      throw new GoogleProviderError('Unknown error occurred while listing models from Google', {
        suggestions: [
          'Check your network connection',
          'Verify your environment setup',
          'Check Google AI Studio status for service disruptions',
          'Try with a different API key or account'
        ]
      });
    }
  }
}

// Export a default instance
export const googleProvider = new GoogleProvider();
