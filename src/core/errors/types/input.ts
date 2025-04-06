/**
 * Input-related error classes
 * 
 * This module provides errors related to user input processing and validation,
 * such as invalid formats, missing required fields, or incorrect values.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * Input validation error class.
 * 
 * This error class is used for issues related to validation of user inputs,
 * such as invalid formats, missing required fields, or incorrect values.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new ValidationError('Invalid parameter format');
 * 
 * // With suggestions
 * throw new ValidationError('Prompt exceeds maximum allowed length', {
 *   suggestions: [
 *     'Limit your prompt to 4000 characters',
 *     'Break up long prompts into multiple requests'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class ValidationError extends ThinktankError {
  /**
   * Creates a new ValidationError instance.
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
      category: errorCategories.VALIDATION,
    });
  }
}

/**
 * Input processing error class.
 * 
 * This error class is used for issues related to processing user inputs,
 * such as invalid prompt formats, unsupported file types, or content issues.
 * Unlike ValidationError which focuses on format/schema validation, this error
 * is related to semantic issues or processing failures.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new InputError('Failed to process input file');
 * 
 * // With suggestions
 * throw new InputError('Input file contains unsupported markdown syntax', {
 *   suggestions: [
 *     'Use only basic markdown syntax',
 *     'Remove complex tables or diagrams',
 *     'Check for syntax errors in your markdown'
 *   ],
 *   examples: [
 *     '# Heading\n\nSimple paragraph with **bold** and *italic* text'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class InputError extends ThinktankError {
  /**
   * Creates a new InputError instance.
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
      category: errorCategories.INPUT,
    });
  }
}