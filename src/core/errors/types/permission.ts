/**
 * Permission-related error classes
 * 
 * This module provides errors related to permissions,
 * such as insufficient file system permissions, API access restrictions,
 * or unauthorized operations.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * Permission-related error class.
 * 
 * This error class is used for issues related to permissions,
 * such as insufficient file system permissions, API access restrictions,
 * or unauthorized operations.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new PermissionError('Permission denied');
 * 
 * // With suggestions
 * throw new PermissionError('Permission denied when writing to output directory', {
 *   suggestions: [
 *     'Check file system permissions for the output directory',
 *     'Run the command with appropriate privileges',
 *     'Select a different output directory that you have write access to'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class PermissionError extends ThinktankError {
  /**
   * Creates a new PermissionError instance.
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
      category: errorCategories.PERMISSION,
    });
  }
}
