/**
 * Configuration-related error classes
 * 
 * This module provides errors related to configuration issues,
 * such as invalid configuration files or model formats.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * Configuration-related error class.
 * 
 * This error class is used for issues related to the application configuration,
 * such as invalid configuration files, missing required properties, or invalid
 * configuration values.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new ConfigError('Invalid model configuration');
 * 
 * // With suggestions and examples
 * throw new ConfigError('Required configuration file not found', {
 *   suggestions: [
 *     'Create a thinktank.config.json file in your project root',
 *     'Copy the template from the examples directory'
 *   ],
 *   examples: [
 *     'cp templates/thinktank.config.default.json ./thinktank.config.json'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class ConfigError extends ThinktankError {
  /**
   * Creates a new ConfigError instance.
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
      category: errorCategories.CONFIG,
    });
  }
}
