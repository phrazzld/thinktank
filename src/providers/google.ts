/**
 * Google Gemini provider implementation for thinktank
 */
import { GoogleGenerativeAI, GenerationConfig, HarmCategory, HarmBlockThreshold } from "@google/generative-ai";
import axios from 'axios';
import { LLMProvider, LLMResponse, ModelOptions, LLMAvailableModel, SystemPrompt } from '../core/types';
import { registerProvider } from '../core/llmRegistry';

/**
 * Google provider error class
 */
export class GoogleProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'GoogleProviderError';
  }
}

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
   * @throws {GoogleProviderError} If the API key is missing
   */
  private getClient(): GoogleGenerativeAI {
    if (this.client) {
      return this.client;
    }
    
    // Use the provided API key or fall back to the environment variable
    const apiKey = this.apiKey || process.env.GEMINI_API_KEY;
    
    if (!apiKey) {
      throw new GoogleProviderError('Google API key is missing. Set GEMINI_API_KEY environment variable or provide it when creating the provider.');
    }
    
    this.client = new GoogleGenerativeAI(apiKey);
    return this.client;
  }
  
  /**
   * Translates generic model options to Google Gemini-specific parameters
   * 
   * @param options - Generic model options
   * @returns Gemini-specific generation config
   */
  private mapOptions(options?: ModelOptions): GenerationConfig {
    const generationConfig: GenerationConfig = {};
    
    if (!options) {
      return generationConfig;
    }
    
    // Map temperature (0-1 scale)
    if (options.temperature !== undefined) {
      generationConfig.temperature = options.temperature;
    }
    
    // Map maxTokens to maxOutputTokens
    if (options.maxTokens !== undefined) {
      generationConfig.maxOutputTokens = options.maxTokens;
    }
    
    // Map other known Gemini-specific parameters if provided
    if (options.topP !== undefined) {
      generationConfig.topP = options.topP as number;
    }
    
    if (options.topK !== undefined) {
      generationConfig.topK = options.topK as number;
    }
    
    // Add any other options directly
    // This allows passing Gemini-specific options from the config
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
    options?: ModelOptions,
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
      if (error instanceof GoogleProviderError) {
        throw error; // Re-throw our own errors
      }
      
      if (error instanceof Error) {
        throw new GoogleProviderError(`Google API error: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new GoogleProviderError('Unknown error occurred while generating text from Google Gemini');
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
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        const responseData = error.response?.data as { error?: { message?: string } } | undefined;
        const message = responseData?.error?.message || error.message;
        throw new GoogleProviderError(`Google API error listing models (Status: ${status}): ${message}`, error);
      }
      
      if (error instanceof Error) {
        throw new GoogleProviderError(`Error listing Google models: ${error.message}`, error);
      }
      
      // Handle unknown errors
      throw new GoogleProviderError('Unknown error occurred while listing models from Google');
    }
  }
}

// Export a default instance
export const googleProvider = new GoogleProvider();