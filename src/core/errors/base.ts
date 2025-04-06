/**
 * Base error classes and constants for the thinktank error system
 * 
 * This module defines the core error categories and base error class
 * that all specialized errors inherit from.
 */

import { colors } from '../../utils/consoleUtils';

/**
 * Error categories for consistent categorization across the application.
 * 
 * These categories are used to classify errors in a standardized way,
 * allowing for consistent error handling, display, and filtering.
 * 
 * @property {string} API - Errors related to external API interactions (e.g., OpenAI, Anthropic)
 * @property {string} CONFIG - Configuration-related errors (e.g., invalid settings, missing config)
 * @property {string} NETWORK - Network connectivity issues (e.g., timeouts, connection failures)
 * @property {string} FILESYSTEM - File system operation errors (e.g., file not found, permission denied)
 * @property {string} PERMISSION - Permission-related errors (e.g., insufficient permissions)
 * @property {string} VALIDATION - Input validation errors (e.g., invalid format, schema violations)
 * @property {string} INPUT - User input processing errors (e.g., invalid prompts)
 * @property {string} UNKNOWN - Uncategorized or internal errors
 */
export const errorCategories = {
  API: 'API',
  CONFIG: 'Configuration',
  NETWORK: 'Network',
  FILESYSTEM: 'File System',
  PERMISSION: 'Permission',
  VALIDATION: 'Validation',
  INPUT: 'Input',
  UNKNOWN: 'Unknown',
};

/**
 * Base error class for all thinktank errors.
 * 
 * This class extends the native JavaScript Error class and provides additional
 * properties and methods for enhanced error reporting, troubleshooting assistance,
 * and consistent formatting across the application.
 * 
 * All specialized error types in the thinktank application should extend this class
 * rather than the native Error class to ensure consistent behavior and properties.
 * 
 * @example
 * ```typescript
 * // Creating a basic ThinktankError
 * const error = new ThinktankError('Something went wrong');
 * 
 * // Creating a more detailed error
 * const detailedError = new ThinktankError('Configuration file is invalid', {
 *   category: errorCategories.CONFIG,
 *   suggestions: [
 *     'Check the JSON syntax in your configuration file',
 *     'Ensure all required fields are present'
 *   ],
 *   examples: [
 *     '{ "models": [], "groups": {} }'
 *   ]
 * });
 * 
 * // Error with a cause
 * try {
 *   JSON.parse(invalidJson);
 * } catch (parseError) {
 *   throw new ThinktankError('Failed to parse configuration', {
 *     cause: parseError,
 *     category: errorCategories.CONFIG
 *   });
 * }
 * ```
 * 
 * @extends {Error}
 */
export class ThinktankError extends Error {
  /**
   * The error category for grouping similar errors.
   * 
   * This property allows for categorization of errors to help with filtering,
   * handling, and displaying errors in a more structured way. Default is 'Unknown'.
   * 
   * @type {string}
   * @default errorCategories.UNKNOWN
   */
  category: string = errorCategories.UNKNOWN;
  
  /**
   * List of suggestions to help resolve the error.
   * 
   * These suggestions are displayed to the user to provide actionable
   * guidance on how to fix the issue.
   * 
   * @type {string[] | undefined}
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands, configurations, or usage patterns.
   * 
   * These examples help users understand the correct way to use the
   * functionality that caused the error.
   * 
   * @type {string[] | undefined}
   */
  examples?: string[];

  /**
   * The original error that caused this error.
   * 
   * This property facilitates error chaining, allowing for more detailed
   * error diagnostics by preserving the underlying cause.
   * 
   * @type {Error | undefined}
   */
  cause?: Error;
  
  /**
   * Sets the error name property based on the constructor's name
   * 
   * This protected method is called by the constructor to ensure that
   * the error name is properly set for all subclasses in the hierarchy.
   * This allows new error types to inherit the correct naming behavior
   * without having to manually set the name property.
   * 
   * @protected
   */
  protected setErrorName(): void {
    // Get the constructor's name (which is the class name)
    const constructorName = this.constructor.name;
    // Set the name property to the class name
    this.name = constructorName;
  }

  /**
   * Creates a new ThinktankError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.category - The error category from errorCategories
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    category?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message);
    // Set the error name based on the class name
    this.setErrorName();
    
    if (options?.cause) {
      this.cause = options.cause;
    }
    
    if (options?.category) {
      this.category = options.category;
    }
    
    if (options?.suggestions) {
      this.suggestions = options.suggestions;
    }
    
    if (options?.examples) {
      this.examples = options.examples;
    }
  }
  
  /**
   * Formats the error for display in CLI and logging contexts.
   * 
   * This method generates a user-friendly, formatted representation of the error
   * that includes the error message, category, suggestions, and examples.
   * The output uses ANSI colors for better readability in terminal environments.
   * 
   * @returns A formatted string representation of the error
   * 
   * @example
   * ```typescript
   * const error = new ConfigError('Invalid model format');
   * console.log(error.format());
   * // Output: 
   * // Error (Configuration): Invalid model format
   * ```
   */
  format(): string {
    let output = `${colors.red.bold('Error')} (${colors.yellow(this.category)}): ${this.message}`;
    
    if (this.suggestions?.length) {
      output += '\n\nSuggestions:';
      this.suggestions.forEach(suggestion => {
        output += `\n  • ${suggestion}`;
      });
    }
    
    if (this.examples?.length) {
      output += '\n\nExamples:';
      this.examples.forEach(example => {
        output += `\n  - ${example}`;
      });
    }
    
    return output;
  }
}