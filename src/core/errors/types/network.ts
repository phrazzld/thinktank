/**
 * Network-related error classes
 * 
 * This module provides errors related to network connectivity,
 * such as timeouts, connection failures, or DNS resolution problems.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * Network-related error class.
 * 
 * This error class is used for issues related to network connectivity,
 * such as timeouts, connection failures, or DNS resolution problems.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new NetworkError('Connection timeout');
 * 
 * // With cause and suggestions
 * try {
 *   // Network operation
 * } catch (error) {
 *   throw new NetworkError('Failed to connect to API endpoint', {
 *     cause: error,
 *     suggestions: [
 *       'Check your internet connection',
 *       'Verify the API endpoint is correct and accessible',
 *       'Try again later as the service might be temporarily down'
 *     ]
 *   });
 * }
 * ```
 * 
 * @extends {ThinktankError}
 */
export class NetworkError extends ThinktankError {
  /**
   * Creates a new NetworkError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.NETWORK,
    });
  }
}