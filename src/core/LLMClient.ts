/**
 * Concrete LLMClient implementation.
 * 
 * This module provides a concrete implementation of the LLMClient interface
 * that wraps the existing provider logic through the llmRegistry.
 */

import { LLMClient, ConfigManagerInterface } from './interfaces';
import { LLMResponse, ModelOptions, SystemPrompt, AppConfig } from './types';
import { callProvider } from './llmRegistry';
import { findModelGroup, findModel } from './configManager';
import { 
  ApiError, 
  ConfigError, 
  ThinktankError
} from './errors';
import { createModelNotFoundError } from './errors/factories/config';

/**
 * Concrete implementation of the LLMClient interface that wraps the
 * existing provider logic through the llmRegistry's callProvider function.
 */
export class ConcreteLLMClient implements LLMClient {
  /**
   * Creates a new ConcreteLLMClient instance
   * 
   * @param configManager - Configuration manager for loading app config
   */
  constructor(private configManager: ConfigManagerInterface) {
    if (!configManager) {
      throw new Error("ConfigManagerInterface instance is required");
    }
  }

  /**
   * Generates a response from a language model.
   * 
   * @param prompt - The main user prompt
   * @param providerModelId - Model identifier in "provider:modelId" format (e.g., "openai:gpt-4o")
   * @param options - Optional parameters for generation
   * @param systemPrompt - Optional system prompt to guide model behavior
   * @returns A promise resolving to the standardized LLMResponse
   * @throws {ConfigError} If the providerModelId format is invalid or model not found
   * @throws {ApiError} If the API call fails
   */
  async generate(
    prompt: string,
    providerModelId: string,
    options?: ModelOptions,
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse> {
    let providerId: string | undefined;
    let modelId: string | undefined;

    try {
      // 1. Parse provider:modelId
      const parts = providerModelId.split(':');
      if (parts.length !== 2 || !parts[0] || !parts[1]) {
        throw new ConfigError(`Invalid model identifier format: "${providerModelId}". Expected "provider:modelId" format.`, {
          suggestions: [
            'Use format "provider:modelId", for example "openai:gpt-4o" or "anthropic:claude-3-opus-20240229"'
          ],
          examples: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229', 'google:gemini-1.5-pro']
        });
      }

      providerId = parts[0];
      modelId = parts[1];

      // 2. Load configuration
      const config: AppConfig = await this.configManager.loadConfig();

      // 3. Find model configuration
      const modelConfig = findModel(config, providerId, modelId);
      if (!modelConfig) {
        // Get available models for better error message
        const availableModels = config.models.map(m => `${m.provider}:${m.modelId}`);
        throw createModelNotFoundError(providerModelId, availableModels);
      }

      // 4. Find group information to determine system prompt
      const groupInfo = findModelGroup(config, modelConfig);
      const groupSystemPrompt = groupInfo?.systemPrompt;

      // 5. Determine final system prompt (override > model > group)
      const finalSystemPrompt = systemPrompt ?? modelConfig.systemPrompt ?? groupSystemPrompt;

      // 6. Delegate to callProvider
      const response = await callProvider(
        providerId,
        modelId,
        prompt,
        modelConfig,
        undefined, // Group options handled within callProvider
        options,   // Direct/CLI options have highest priority
        finalSystemPrompt
      );

      // 7. Add group information to response if applicable
      if (groupInfo && !response.groupInfo) {
        response.groupInfo = {
          name: groupInfo.groupName,
          systemPrompt: groupInfo.systemPrompt
        };
      }

      return response;

    } catch (error) {
      // Re-throw known errors
      if (error instanceof ConfigError || error instanceof ApiError || error instanceof ThinktankError) {
        throw error;
      }

      // Wrap unknown errors
      const errorMessage = error instanceof Error ? error.message : String(error);
      throw new ApiError(`LLM request failed for ${providerModelId}: ${errorMessage}`, {
        cause: error instanceof Error ? error : undefined,
        providerId: providerId,
        suggestions: [
          'Check API key and network connectivity',
          'Verify model availability and access permissions',
          'Review provider status for any ongoing issues'
        ]
      });
    }
  }
}
