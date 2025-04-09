/**
 * Type definitions for fileReader module.
 *
 * Extracted to a separate file to avoid circular dependencies in tests.
 */

/**
 * Result of reading a context file
 */
export interface ContextFileResult {
  /**
   * Path to the file that was read (original path provided)
   */
  path: string;

  /**
   * Content of the file, or null if there was an error
   */
  content: string | null;

  /**
   * Error information if reading failed, or null if successful
   */
  error: {
    /**
     * Error code (e.g., 'ENOENT', 'EACCES', 'NOT_FILE', etc.)
     */
    code: string;

    /**
     * Human-readable error message
     */
    message: string;
  } | null;
}

/**
 * Options for reading file content
 */
export interface ReadFileOptions {
  normalize?: boolean;
}

/**
 * Custom error for file reading operations
 */
export class FileReadError extends Error {
  constructor(
    message: string,
    public readonly cause?: Error
  ) {
    super(message);
    this.name = 'FileReadError';
  }
}
