/**
 * Factory functions for creating configuration errors
 * 
 * This module provides factory functions for creating common configuration errors
 * with helpful guidance and context-specific suggestions.
 */

import { ConfigError } from '../types/config';

/**
 * Creates an error for invalid model format with context-aware suggestions.
 * 
 * This factory function generates a ConfigError for issues with model specification
 * format. It analyzes the given model string and generates appropriate suggestions
 * based on the specific issue detected (missing colon, missing provider, etc.).
 * 
 * When provided with available providers and models, it enhances the suggestions
 * with specific examples from the available options.
 * 
 * @param modelSpecification - The invalid model specification (e.g., "openai-gpt4" instead of "openai:gpt-4")
 * @param availableProviders - Optional array of available provider IDs (e.g., ["openai", "anthropic"])
 * @param availableModels - Optional array of available model specifications (e.g., ["openai:gpt-4o", "anthropic:claude-3"])
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with helpful suggestions based on the specific format issue
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw createModelFormatError('openai-gpt4');
 * 
 * // With available providers and models
 * throw createModelFormatError(
 *   'openai-gpt4',
 *   ['openai', 'anthropic', 'google'],
 *   ['openai:gpt-4o', 'openai:gpt-3.5-turbo', 'anthropic:claude-3-opus']
 * );
 * ```
 */
export function createModelFormatError(
  modelSpecification: string,
  availableProviders: string[] = [],
  availableModels: string[] = [],
  errorMessage?: string
): ConfigError {
  // Create the error with a default message if none provided
  let message = errorMessage;
  
  if (!message) {
    if (!modelSpecification.includes(':')) {
      message = `Invalid model format: "${modelSpecification}". Model must be specified as "provider:modelId".`;
    } else if (modelSpecification.endsWith(':')) {
      message = `Invalid model format: "${modelSpecification}". Missing model ID after provider.`;
    } else if (modelSpecification.startsWith(':')) {
      message = `Invalid model format: "${modelSpecification}". Missing provider name before model ID.`;
    } else {
      message = `Model not found: "${modelSpecification}". Use "provider:modelId" format.`;
    }
  }
  
  // Parse the model specification
  const [provider] = modelSpecification.split(':');
  
  // Build suggestions
  const suggestions: string[] = [
    `Model specifications must use the format "provider:modelId" (e.g., "openai:gpt-4o")`
  ];
  
  // Specific error cases
  if (!modelSpecification.includes(':')) {
    suggestions.push(`Add a colon between provider and model ID: "${modelSpecification}" → "provider:${modelSpecification}"`);
  } else if (modelSpecification.endsWith(':')) {
    suggestions.push(`Specify a model ID after the provider: "${modelSpecification}modelId"`);
    
    // If we have models from this provider, suggest some
    if (availableModels.length > 0) {
      const matchingModels = availableModels.filter(m => m.startsWith(`${provider}:`));
      if (matchingModels.length > 0) {
        const models = matchingModels.slice(0, 3).join(', ') + 
          (matchingModels.length > 3 ? ', ...' : '');
        suggestions.push(`Available models for ${provider}: ${models}`);
      }
    }
  } else if (modelSpecification.startsWith(':')) {
    suggestions.push(`Specify a provider before the model ID: "provider${modelSpecification}"`);
    
    // If we have providers, suggest some
    if (availableProviders.length > 0) {
      const providersList = availableProviders.slice(0, 5).join(', ') + 
        (availableProviders.length > 5 ? ', ...' : '');
      suggestions.push(`Available providers: ${providersList}`);
    }
  }
  
  // Add general provider/model suggestions
  if (availableProviders.length > 0) {
    suggestions.push(`Available providers: ${availableProviders.join(', ')}`);
  }
  
  if (availableModels.length > 0) {
    // Limit to a reasonable number of examples
    const modelExamples = availableModels.slice(0, 5);
    const exampleList = modelExamples.join(', ') + 
      (availableModels.length > 5 ? ', ...' : '');
    suggestions.push(`Example models: ${exampleList}`);
  }
  
  // Add examples
  const examples = [
    'openai:gpt-4o',
    'anthropic:claude-3-7-sonnet-20250219',
    'google:gemini-pro'
  ];
  
  return new ConfigError(message, { suggestions, examples });
}

/**
 * Creates an error for model not found in configuration with specific suggestions.
 * 
 * This factory function generates a ConfigError when a requested model cannot
 * be found in the configuration. It provides context-specific suggestions based
 * on whether the issue is related to a model group or general configuration.
 * 
 * When provided with available models, it enhances the suggestions with similar
 * models or available providers to help users correct their model selection.
 * 
 * @param modelSpecification - The model specification that wasn't found (e.g., "openai:gpt-5")
 * @param availableModels - Optional array of available model specifications to suggest alternatives
 * @param groupName - Optional group name if relevant to the context (e.g., when model is missing from a specific group)
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with context-aware suggestions for resolving the issue
 * 
 * @example
 * ```typescript
 * // Basic usage - model not found
 * throw createModelNotFoundError('openai:nonexistent-model');
 * 
 * // When a model is not found in a specific group
 * throw createModelNotFoundError(
 *   'openai:gpt-4o', 
 *   ['openai:gpt-3.5-turbo', 'anthropic:claude-3-haiku'],
 *   'fast-models'
 * );
 * ```
 */
export function createModelNotFoundError(
  modelSpecification: string,
  availableModels: string[] = [],
  groupName?: string,
  errorMessage?: string
): ConfigError {
  const [provider, modelId] = modelSpecification.split(':');
  
  // Create the error with a default message if none provided
  let message = errorMessage;
  
  if (!message) {
    if (groupName) {
      message = `Model "${modelSpecification}" not found in group "${groupName}".`;
    } else {
      message = `Model "${modelSpecification}" not found in configuration.`;
    }
  }
  
  // Build suggestions
  const suggestions: string[] = [];
  
  // Suggest similar models by partial matching
  if (availableModels.length > 0) {
    // Find models with the same provider
    const sameProviderModels = availableModels.filter(m => m.startsWith(`${provider}:`));
    
    if (sameProviderModels.length > 0) {
      const providerModelList = sameProviderModels.slice(0, 5).join(', ') + 
        (sameProviderModels.length > 5 ? ', ...' : '');
      suggestions.push(`Available models from ${provider}: ${providerModelList}`);
    } else {
      // Provider not found
      suggestions.push(`Provider "${provider}" not found.`);
      
      // Find all available providers
      const availableProviders = new Set<string>();
      availableModels.forEach(m => {
        const parts = m.split(':');
        if (parts.length === 2) {
          availableProviders.add(parts[0]);
        }
      });
      
      if (availableProviders.size > 0) {
        suggestions.push(`Available providers: ${Array.from(availableProviders).join(', ')}`);
      }
    }
    
    // For specific model ID matching
    if (modelId) {
      // Find models with similar model IDs
      const similarModelIds = availableModels.filter(m => {
        const parts = m.split(':');
        return parts.length === 2 && parts[1].includes(modelId);
      });
      
      if (similarModelIds.length > 0) {
        const similarList = similarModelIds.slice(0, 3).join(', ') + 
          (similarModelIds.length > 3 ? ', ...' : '');
        suggestions.push(`Models with similar IDs: ${similarList}`);
      }
    }
    
    // Add a list of example models regardless
    const exampleList = availableModels.slice(0, 5).join(', ') + 
      (availableModels.length > 5 ? ', ...' : '');
    suggestions.push(`Available models: ${exampleList}`);
  }
  
  // Add configuration suggestions
  suggestions.push(
    'Check your configuration file to ensure the model is properly defined',
    'Models must be enabled in the configuration to be usable',
    'Use "thinktank config path" to locate your configuration file'
  );
  
  if (groupName) {
    suggestions.push(`Ensure the model is included in the "${groupName}" group configuration`);
  }
  
  // Add examples
  const examples = availableModels.length > 0 
    ? availableModels.slice(0, 3) 
    : [
        'openai:gpt-4o',
        'anthropic:claude-3-7-sonnet-20250219',
        'google:gemini-pro'
      ];
  
  return new ConfigError(message, { suggestions, examples });
}