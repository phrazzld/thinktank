/**
 * Virtual filesystem utilities for testing
 * 
 * This module provides a simple interface to the memfs in-memory filesystem library,
 * allowing tests to create and interact with a virtual filesystem without complex mocking.
 * 
 * @module virtualFsUtils
 * 
 * @example
 * ```typescript
 * import { createVirtualFs, resetVirtualFs, mockFsModules } from '../../__tests__/utils/virtualFsUtils';
 * 
 * // Setup mocks (must be before importing fs modules)
 * jest.mock('fs', () => mockFsModules().fs);
 * jest.mock('fs/promises', () => mockFsModules().fsPromises);
 * 
 * // Import fs modules after mocking
 * import fs from 'fs';
 * import fsPromises from 'fs/promises';
 * 
 * // Setup before tests
 * beforeEach(() => {
 *   resetVirtualFs();
 *   
 *   // Create virtual filesystem structure
 *   createVirtualFs({
 *     '/path/to/file.txt': 'File content',
 *     '/path/to/dir/nested.txt': 'Nested file content'
 *   });
 * });
 * 
 * it('should read file content', async () => {
 *   const content = await fsPromises.readFile('/path/to/file.txt', 'utf-8');
 *   expect(content).toBe('File content');
 * });
 * ```
 */
import { Volume, createFsFromVolume } from 'memfs';
import type { IFs } from 'memfs';

// Create a volume that will be shared across all imports
const vol = Volume.fromJSON({});

// Create a filesystem based on this volume
const fs = createFsFromVolume(vol);

/**
 * Creates a virtual filesystem with the specified structure.
 * 
 * The virtual filesystem is completely reset before creating the new structure,
 * so any existing files or directories will be removed.
 * 
 * @param structure - Object mapping file paths to their content
 * @param options - Optional configuration options
 * 
 * @example
 * ```typescript
 * // Create a simple file structure
 * createVirtualFs({
 *   '/path/to/file.txt': 'File content',
 *   '/path/to/empty-file.txt': '',
 *   '/path/to/nested/file.js': 'console.log("Hello");'
 * });
 * ```
 */
export function createVirtualFs(
  structure: Record<string, string>,
  options: { reset?: boolean } = { reset: true }
): void {
  if (options.reset) {
    resetVirtualFs();
  }
  
  // Load the structure into the volume
  vol.fromJSON(structure);
}

/**
 * Resets the virtual filesystem, removing all files and directories.
 * 
 * This function should typically be called in the beforeEach hook to ensure
 * a clean filesystem state before each test.
 * 
 * @example
 * ```typescript
 * beforeEach(() => {
 *   resetVirtualFs();
 * });
 * ```
 */
export function resetVirtualFs(): void {
  vol.reset();
}

/**
 * Gets the virtual filesystem instance for direct manipulation.
 * 
 * This function returns the current Volume instance, allowing direct access
 * to all fs methods for more advanced use cases.
 * 
 * @returns The current memfs Volume instance
 * 
 * @example
 * ```typescript
 * // Get direct access to the virtual filesystem
 * const virtualFs = getVirtualFs();
 * 
 * // Directly use fs methods
 * virtualFs.writeFileSync('/path/to/file.txt', 'New content');
 * const content = virtualFs.readFileSync('/path/to/file.txt', 'utf8');
 * ```
 */
export function getVirtualFs(): IFs {
  return vol as unknown as IFs;
}

/**
 * Sets up mocking of the fs and fs/promises modules to use the virtual filesystem.
 * 
 * This function returns an object with references to both the 'fs' and 'fs/promises'
 * implementations from memfs. It's meant to be used with jest.mock.
 * 
 * @returns Object containing references to fs and fs/promises implementations
 * 
 * @example
 * ```typescript
 * // In your test file:
 * import { mockFsModules } from '../../__tests__/utils/virtualFsUtils';
 * 
 * // Then in a separate file or section (not inside any tests):
 * jest.mock('fs', () => mockFsModules().fs);
 * jest.mock('fs/promises', () => mockFsModules().fsPromises);
 * ```
 */
export function mockFsModules(): { fs: any; fsPromises: any } {
  // Return the fs and promises implementations based on our volume
  return {
    fs,
    fsPromises: fs.promises
  };
}

/**
 * Creates a standardized Node.js filesystem error object.
 * 
 * This helper function generates error objects that match the structure of
 * real Node.js filesystem errors, with appropriate code, message, path, and errno properties.
 * 
 * @param code - Error code (e.g., 'ENOENT', 'EACCES')
 * @param message - Error message
 * @param syscall - System call that failed
 * @param filepath - Path that caused the error
 * @returns A properly formatted filesystem error
 * 
 * @example
 * ```typescript
 * // Create a "file not found" error
 * const error = createFsError('ENOENT', 'File not found', 'open', '/path/to/missing.txt');
 * 
 * // Create a permission denied error
 * const error = createFsError('EACCES', 'Permission denied', 'access', '/path/to/protected.txt');
 * ```
 */
export function createFsError(
  code: string,
  message: string,
  syscall: string,
  filepath: string
): NodeJS.ErrnoException {
  // Create base error
  const error = new Error(message) as NodeJS.ErrnoException;
  
  // Set standard properties
  error.code = code;
  error.syscall = syscall;
  error.path = filepath;
  
  // Map common error codes to errno numbers for compatibility
  if (code === 'ENOENT') error.errno = -2;
  else if (code === 'EACCES') error.errno = -13;
  else if (code === 'EPERM') error.errno = -1;
  else if (code === 'EROFS') error.errno = -30;
  else if (code === 'EBUSY') error.errno = -16;
  else if (code === 'EMFILE') error.errno = -24;
  
  return error;
}