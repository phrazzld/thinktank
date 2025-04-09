/**
 * API-related error classes
 *
 * This module provides errors related to API interactions,
 * such as authentication failures, rate limits, or server errors.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * API interaction error class.
 *
 * This error class is used for issues related to external API interactions,
 * such as authentication failures, rate limits, or server errors from LLM
 * providers like OpenAI, Anthropic, or Google.
 *
 * The `providerId` property allows for provider-specific error handling and
 * guidance.
 *
 * @example
 * ```typescript
 * // Basic usage
 * throw new ApiError('Failed to generate response', {
 *   providerId: 'openai'
 * });
 *
 * // With detailed options
 * throw new ApiError('API rate limit exceeded', {
 *   providerId: 'anthropic',
 *   suggestions: [
 *     'Wait and try again later',
 *     'Reduce the frequency of requests',
 *     'Consider upgrading your API tier'
 *   ],
 *   cause: originalError
 * });
 * ```
 *
 * @extends {ThinktankError}
 */
export class ApiError extends ThinktankError {
  /**
   * The identifier of the provider that generated the error.
   *
   * This allows for provider-specific error handling and guidance
   * (e.g., 'openai', 'anthropic', 'google').
   *
   * @type {string | undefined}
   */
  providerId?: string;

  /**
   * Creates a new ApiError instance.
   *
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.providerId - Identifier of the provider that generated the error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(
    message: string,
    options?: {
      cause?: Error;
      providerId?: string;
      suggestions?: string[];
      examples?: string[];
    }
  ) {
    // Prefix the message with the provider ID if available
    const formattedMessage = options?.providerId ? `[${options.providerId}] ${message}` : message;

    super(formattedMessage, {
      ...options,
      category: errorCategories.API,
    });

    this.providerId = options?.providerId;
  }
}
