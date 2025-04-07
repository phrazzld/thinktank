/**
 * Type definitions for filesystem mock setup
 */

/**
 * Sets up a basic filesystem structure for tests
 * @param structure - Object mapping file paths to content
 */
export function setupBasicFs(structure?: Record<string, string>): void;

/**
 * Creates a standardized filesystem error
 * @param code - Error code (e.g., 'ENOENT', 'EACCES')
 * @param message - Error message
 * @param syscall - System call that failed
 * @param filepath - Path that caused the error
 * @returns A properly formatted filesystem error
 */
export function createFsError(
  code: string,
  message: string,
  syscall: string,
  filepath: string
): NodeJS.ErrnoException;
