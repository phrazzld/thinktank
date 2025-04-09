/**
 * Template for listing available models from providers
 */
import { loadConfig, getApiKey } from '../core/configManager';
import { getProvider } from '../core/llmRegistry';
import { formatModelList } from '../utils/outputFormatter';
import { LLMAvailableModel } from '../core/types';

// Import provider modules to ensure they're registered
import '../providers/openai';
import '../providers/anthropic';
import '../providers/google';
import '../providers/openrouter';

/**
 * Options for listing models
 */
interface ListModelsOptions {
  /**
   * Path to a custom config file
   */
  config?: string;

  /**
   * Filter models by provider ID
   */
  provider?: string;
}

/**
 * Retrieves a list of available models from configured providers
 *
 * @param options - Options for listing models
 * @returns Formatted string showing available models
 */
export async function listAvailableModels(options: ListModelsOptions): Promise<string> {
  // Load configuration
  const config = await loadConfig({ configPath: options.config });

  // Get unique provider IDs from configuration (optionally filtered)
  let providerIds: string[] = [];

  if (options.provider) {
    // Filter models by the specified provider
    const matchingModels = config.models.filter(model => model.provider === options.provider);
    if (matchingModels.length > 0) {
      providerIds = [options.provider];
    }
  } else {
    // Get all unique provider IDs
    providerIds = Array.from(new Set(config.models.map(model => model.provider)));
  }

  // Prepare calls to providers
  const modelPromises: {
    [providerId: string]: Promise<LLMAvailableModel[] | { error: string }>;
  } = {};

  for (const providerId of providerIds) {
    try {
      // Get provider from registry
      const provider = getProvider(providerId);

      if (!provider) {
        modelPromises[providerId] = Promise.resolve({
          error: `Provider '${providerId}' not found in registry`,
        });
        continue;
      }

      if (!provider.listModels) {
        modelPromises[providerId] = Promise.resolve({
          error: `Provider '${providerId}' does not support listing models`,
        });
        continue;
      }

      // Find any model config for this provider to get the API key
      const modelConfig = config.models.find(model => model.provider === providerId);
      if (!modelConfig) {
        modelPromises[providerId] = Promise.resolve({
          error: `No models configured for provider '${providerId}'`,
        });
        continue;
      }

      // Get API key
      const apiKey = getApiKey(modelConfig);
      if (!apiKey) {
        modelPromises[providerId] = Promise.resolve({
          error: `Missing API key for provider '${providerId}'`,
        });
        continue;
      }

      // Create promise to fetch models
      modelPromises[providerId] = provider
        .listModels(apiKey)
        .then(models => models)
        .catch((error: Error) => {
          return { error: `Error fetching models: ${error.message}` };
        });
    } catch (error) {
      // Handle any unexpected errors
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      modelPromises[providerId] = Promise.resolve({
        error: `Unexpected error: ${errorMessage}`,
      });
    }
  }

  // Execute all promises concurrently
  const results = await Promise.all(
    Object.entries(modelPromises).map(async ([providerId, promise]) => {
      try {
        const result = await promise;
        return { providerId, result };
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        return {
          providerId,
          result: { error: `Unexpected error: ${errorMessage}` },
        };
      }
    })
  );

  // Organize results by provider
  const modelsByProvider: { [providerId: string]: LLMAvailableModel[] | { error: string } } = {};

  for (const { providerId, result } of results) {
    modelsByProvider[providerId] = result;
  }

  // Format and return the results
  return formatModelList(modelsByProvider);
}
