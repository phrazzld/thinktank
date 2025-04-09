/**
 * Centralized error handling system for thinktank
 *
 * This module defines a consistent error hierarchy and error handling utilities
 * to improve error reporting and troubleshooting across the application.
 *
 * The error system is built around a hierarchy of error classes extending from the base
 * {@link ThinktankError} class, with specialized subclasses for different error categories.
 *
 * Key features:
 * - Consistent error categorization with {@link errorCategories}
 * - Rich error objects with suggestions and examples
 * - Support for error chaining with the `cause` property
 * - Standardized formatting through the {@link ThinktankError.format} method
 * - Factory functions for common error types
 *
 * @example
 * ```typescript
 * // Using error class directly
 * throw new ConfigError('Invalid configuration', {
 *   suggestions: ['Check your configuration file']
 * });
 *
 * // Using a factory function
 * throw createFileNotFoundError('/path/to/missing-file.txt');
 *
 * // Using a provider-specific factory function
 * throw createProviderRateLimitError('openai', 'OpenAI', error);
 * ```
 *
 * @module errors
 */

// Re-export base error class and categories
export { ThinktankError, errorCategories } from './base';

// Re-export all error classes
export {
  ConfigError,
  ApiError,
  FileSystemError,
  ValidationError,
  InputError,
  NetworkError,
  PermissionError,
} from './types';

// Re-export basic factory functions
export {
  createFileNotFoundError,
  createModelFormatError,
  createModelNotFoundError,
  createMissingApiKeyError,
} from './factories';

// Re-export provider factory functions
export {
  createProviderApiKeyMissingError,
  createProviderRateLimitError,
  createProviderModelNotFoundError,
  createProviderTokenLimitError,
  createProviderContentPolicyError,
  createProviderUnknownError,
  createProviderNetworkError,

  // Provider error detection utilities
  isProviderRateLimitError,
  isProviderTokenLimitError,
  isProviderContentPolicyError,
  isProviderAuthError,
  isProviderNetworkError,
} from './factories/provider';
