/**
 * Type definitions for filesystem mock setup
 */

/**
 * Sets up a basic filesystem structure for tests
 * @param structure - Object mapping file paths to content
 * @param options - Additional options (reset: boolean)
 */
export function setupBasicFs(structure?: Record<string, string>, options?: { reset: boolean }): void;

/**
 * Resets the virtual filesystem, removing all files and directories.
 */
export function resetFs(): void;

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

/**
 * Gets direct access to the virtual filesystem for advanced operations
 * @returns The virtual filesystem instance
 */
export function getFs(): any;

/**
 * Creates a mock fs.Stats object for testing
 * @param isFile - Whether this represents a file (true) or directory (false)
 * @param size - Size in bytes
 * @returns A mock Stats object
 */
export function createStats(isFile: boolean, size?: number): any;

/**
 * Creates a mock fs.Dirent object for testing directory entries
 * @param name - Name of the file or directory
 * @param isFile - Whether this represents a file (true) or directory (false)
 * @returns A mock Dirent object
 */
export function createDirent(name: string, isFile: boolean): any;

/**
 * Normalizes a path for use with the virtual filesystem
 * @param path - The path to normalize
 * @returns The normalized path
 */
export function normalizePath(path: string): string;
