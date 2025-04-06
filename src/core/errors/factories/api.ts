/**
 * Factory functions for creating API-related errors
 * 
 * This module provides factory functions for creating common API errors
 * with helpful guidance and context-specific suggestions.
 */

import { ApiError } from '../types/api';

/**
 * Creates an error for missing API keys with provider-specific guidance.
 * 
 * This factory function generates an ApiError for missing API keys, with
 * customized suggestions based on the specific providers that need API keys.
 * It includes provider-specific links to obtain API keys and instructions
 * for setting environment variables.
 * 
 * @param missingModels - Array of models with missing API keys (each with provider and modelId)
 * @param errorMessage - Optional custom error message
 * @returns An ApiError with provider-specific guidance on obtaining and setting API keys
 * 
 * @example
 * ```typescript
 * // Single missing API key
 * throw createMissingApiKeyError([{ provider: 'openai', modelId: 'gpt-4o' }]);
 * 
 * // Multiple missing API keys
 * throw createMissingApiKeyError([
 *   { provider: 'openai', modelId: 'gpt-4o' },
 *   { provider: 'anthropic', modelId: 'claude-3-opus' }
 * ]);
 * ```
 */
export function createMissingApiKeyError(
  missingModels: Array<{ provider: string; modelId: string }>,
  errorMessage?: string
): ApiError {
  // Create the error with a default message if none provided
  const message = errorMessage || 
    `Missing API key${missingModels.length > 1 ? 's' : ''} for ${missingModels.length} model${missingModels.length > 1 ? 's' : ''}`;
  
  // Group models by provider for better suggestions
  const providerModels: Record<string, string[]> = {};
  missingModels.forEach(model => {
    if (!providerModels[model.provider]) {
      providerModels[model.provider] = [];
    }
    providerModels[model.provider].push(`${model.provider}:${model.modelId}`);
  });
  
  // Build suggestions for each provider
  const suggestions: string[] = [];
  
  // Add suggestions for each provider
  Object.entries(providerModels).forEach(([provider, models]) => {
    const modelsText = models.join(', ');
    suggestions.push(`Missing API key for ${provider} model${models.length > 1 ? 's' : ''}: ${modelsText}`);
    
    const envVarName = `${provider.toUpperCase()}_API_KEY`;
    
    // Provider-specific instructions
    switch (provider.toLowerCase()) {
      case 'openai':
        suggestions.push(
          `To use OpenAI models, get your API key from: https://platform.openai.com/api-keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'anthropic':
        suggestions.push(
          `To use Anthropic Claude models, get your API key from: https://console.anthropic.com/keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'google':
        suggestions.push(
          `To use Google AI models, get your API key from: https://aistudio.google.com/app/apikey`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'openrouter':
        suggestions.push(
          `To use OpenRouter models, get your API key from: https://openrouter.ai/keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      default:
        suggestions.push(
          `Get an API key for ${provider} from their developer portal`,
          `Set the ${envVarName} environment variable with your key`
        );
    }
  });
  
  // Add general environment variable setup instructions
  suggestions.push(
    `\nTo set environment variables:`,
    '• For Bash/Zsh: Add `export PROVIDER_API_KEY=your_key_here` to your ~/.bashrc or ~/.zshrc',
    '• For Windows Command Prompt: Use `set PROVIDER_API_KEY=your_key_here`',
    '• For PowerShell: Use `$env:PROVIDER_API_KEY = "your_key_here"`',
    '• For a local project: Create a .env file with `PROVIDER_API_KEY=your_key_here`'
  );
  
  // Add example commands
  const examples = Object.keys(providerModels).map(provider => {
    const envVarName = `${provider.toUpperCase()}_API_KEY`;
    return `export ${envVarName}=your_${provider}_key_here`;
  });
  
  return new ApiError(message, { suggestions, examples });
}