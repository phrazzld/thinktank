import { ApiError } from '../types/api';
import { 
  isRateLimitError, 
  isTokenLimitError, 
  isContentPolicyError, 
  isAuthError, 
  isNetworkError 
} from '../utils/categorization';

/**
 * Creates an API key missing error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param consoleUrl - URL to the provider's console/dashboard for obtaining API keys
 * @returns An ApiError with standardized message, suggestions, and examples
 */
export function createProviderApiKeyMissingError(
  providerId: string,
  providerName: string,
  consoleUrl: string
): ApiError {
  const envVarName = `${providerName.toUpperCase()}_API_KEY`;
  
  return new ApiError(`${providerName} API key is missing. Set ${envVarName} environment variable or provide it when creating the provider.`, {
    providerId,
    suggestions: [
      `Set the ${envVarName} environment variable in your shell or .env file`,
      `Get an API key from the ${providerName} console: ${consoleUrl}`,
      `Provide the API key directly when creating the provider instance`
    ],
    examples: [
      `export ${envVarName}=your_api_key`,
      `const provider = new ${providerName}Provider("your_api_key")`
    ]
  });
}

/**
 * Creates a rate limit error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param originalError - The original error that was thrown
 * @returns An ApiError with standardized message, suggestions, and examples
 */
export function createProviderRateLimitError(
  providerId: string,
  providerName: string,
  originalError: Error
): ApiError {
  return new ApiError(`Rate limit exceeded: ${originalError.message}`, {
    providerId,
    cause: originalError,
    suggestions: [
      'Wait before sending additional requests',
      'Implement exponential backoff in your code',
      `Reduce the frequency of requests to the ${providerName} API`,
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

/**
 * Creates a model not found error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param modelId - The model ID that wasn't found
 * @returns An ApiError with standardized message and suggestions
 */
export function createProviderModelNotFoundError(
  providerId: string,
  providerName: string,
  modelId: string
): ApiError {
  return new ApiError(`Model '${modelId}' not found for ${providerName} provider`, {
    providerId,
    suggestions: [
      `Check available models using the listModels() method`,
      `Verify the model ID is correct and supported by ${providerName}`,
      `Try using a different model from ${providerName}`
    ]
  });
}

/**
 * Creates a token limit error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param originalError - The original error that was thrown
 * @returns An ApiError with standardized message, suggestions, and examples
 */
export function createProviderTokenLimitError(
  providerId: string,
  providerName: string,
  originalError: Error
): ApiError {
  return new ApiError(`Token limit exceeded: ${originalError.message}`, {
    providerId,
    cause: originalError,
    suggestions: [
      'Use a shorter prompt',
      'Break your request into smaller chunks',
      `Try a different ${providerName} model with higher token limits`,
      'Reduce the temperature parameter to avoid unnecessary token usage'
    ]
  });
}

/**
 * Creates a content policy violation error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param originalError - The original error that was thrown
 * @returns An ApiError with standardized message and suggestions
 */
export function createProviderContentPolicyError(
  providerId: string,
  providerName: string,
  originalError: Error
): ApiError {
  return new ApiError(`Content policy violation: ${originalError.message}`, {
    providerId,
    cause: originalError,
    suggestions: [
      'Review and modify content that may violate provider policies',
      `Check ${providerName}'s content policy guidelines`,
      'Rephrase or remove sensitive content from your prompt'
    ]
  });
}

/**
 * Creates an unknown provider error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param originalError - Optional original error that was thrown
 * @returns An ApiError with standardized message and suggestions
 */
export function createProviderUnknownError(
  providerId: string,
  providerName: string,
  originalError?: Error
): ApiError {
  return new ApiError(`Unknown error occurred while generating text from ${providerName}`, {
    providerId,
    cause: originalError,
    suggestions: [
      'Check your network connection',
      'Verify your environment setup',
      'Try with a simpler prompt or different model',
      `Check ${providerName} status for service disruptions`
    ]
  });
}

/**
 * Creates a network error with standardized message and suggestions
 * 
 * @param providerId - The provider ID (e.g., 'openai', 'anthropic')
 * @param providerName - The human-readable provider name (e.g., 'OpenAI', 'Anthropic')
 * @param originalError - The original error that was thrown
 * @returns An ApiError with standardized message and suggestions
 */
export function createProviderNetworkError(
  providerId: string,
  providerName: string,
  originalError: Error
): ApiError {
  return new ApiError(`Network error connecting to ${providerName} API: ${originalError.message}`, {
    providerId,
    cause: originalError,
    suggestions: [
      'Check your internet connection',
      'Verify you can access the API endpoint',
      'Check if the service is experiencing downtime',
      'Try again later or use a different provider'
    ]
  });
}

/**
 * Utility function to detect rate limit errors
 * 
 * @param errorMessage - The error message to check
 * @returns Whether the error appears to be a rate limit error
 */
export function isProviderRateLimitError(errorMessage: string): boolean {
  return isRateLimitError(errorMessage);
}

/**
 * Utility function to detect token limit errors
 * 
 * @param errorMessage - The error message to check
 * @returns Whether the error appears to be a token limit error
 */
export function isProviderTokenLimitError(errorMessage: string): boolean {
  return isTokenLimitError(errorMessage);
}

/**
 * Utility function to detect content policy violations
 * 
 * @param errorMessage - The error message to check
 * @returns Whether the error appears to be a content policy violation
 */
export function isProviderContentPolicyError(errorMessage: string): boolean {
  return isContentPolicyError(errorMessage);
}

/**
 * Utility function to detect authentication errors
 * 
 * @param errorMessage - The error message to check
 * @returns Whether the error appears to be an authentication error
 */
export function isProviderAuthError(errorMessage: string): boolean {
  return isAuthError(errorMessage);
}

/**
 * Utility function to detect network errors
 * 
 * @param errorMessage - The error message to check
 * @returns Whether the error appears to be a network error
 */
export function isProviderNetworkError(errorMessage: string): boolean {
  return isNetworkError(errorMessage);
}