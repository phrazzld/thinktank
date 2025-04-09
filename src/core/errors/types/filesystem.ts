/**
 * File system related error classes
 *
 * This module provides errors related to file system operations,
 * such as file not found, permission denied, or directory creation failures.
 */

import { ThinktankError, errorCategories } from '../base';

/**
 * File system related error class.
 *
 * This error class is used for issues related to file system operations,
 * such as file not found, permission denied, or directory creation failures.
 *
 * The `filePath` property contains the path to the file or directory that
 * caused the error, which can be used for error reporting and troubleshooting.
 *
 * @example
 * ```typescript
 * // Basic usage
 * throw new FileSystemError('Failed to read file', {
 *   filePath: '/path/to/file.txt'
 * });
 *
 * // With suggestions
 * throw new FileSystemError('Permission denied while writing to file', {
 *   filePath: '/path/to/file.txt',
 *   suggestions: [
 *     'Check file permissions',
 *     'Ensure you have write access to the directory'
 *   ]
 * });
 * ```
 *
 * @extends {ThinktankError}
 */
export class FileSystemError extends ThinktankError {
  /**
   * The path to the file or directory that caused the error.
   *
   * This property is used for error reporting and troubleshooting file system issues.
   *
   * @type {string | undefined}
   */
  filePath?: string;

  /**
   * Creates a new FileSystemError instance.
   *
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.filePath - Path to the file or directory that caused the error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(
    message: string,
    options?: {
      cause?: Error;
      filePath?: string;
      suggestions?: string[];
      examples?: string[];
    }
  ) {
    super(message, {
      ...options,
      category: errorCategories.FILESYSTEM,
    });

    this.filePath = options?.filePath;
  }
}
