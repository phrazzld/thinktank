/**
 * Factory functions for creating file system errors
 * 
 * This module provides factory functions for creating common file system errors
 * with helpful guidance and context-specific suggestions.
 */

import { FileSystemError } from '../types/filesystem';

/**
 * Creates a file not found error with helpful suggestions and examples.
 * 
 * This factory function generates a FileSystemError with context-aware suggestions
 * based on the provided file path. It includes guidance on checking file existence,
 * path formatting, file extensions, and permissions.
 * 
 * @param filePath - The path to the file that wasn't found
 * @param errorMessage - Optional custom error message (defaults to "File not found: {filePath}")
 * @returns A FileSystemError with relevant suggestions and examples
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw createFileNotFoundError('/path/to/config.json');
 * 
 * // With custom error message
 * throw createFileNotFoundError('./prompt.txt', 'Unable to locate prompt file');
 * ```
 */
export function createFileNotFoundError(filePath: string, errorMessage?: string): FileSystemError {
  const message = errorMessage || `File not found: ${filePath}`;
  const currentDir = process.cwd();
  
  // Determine if path is absolute
  const isAbsolutePath = filePath.startsWith('/');
  
  // Get dirname and basename
  const lastSlashIndex = filePath.lastIndexOf('/');
  const dirname = lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '.';
  const basename = lastSlashIndex > 0 ? filePath.substring(lastSlashIndex + 1) : filePath;
  
  // Build suggestions
  const suggestions = [
    `Check that the file exists at the specified path: ${isAbsolutePath ? filePath : `${currentDir}/${filePath}`}`,
    `Current working directory: ${currentDir}`
  ];
  
  // Add path-specific suggestions
  if (!isAbsolutePath && dirname !== '.') {
    suggestions.push(`Ensure the directory exists: ${currentDir}/${dirname}`);
  }
  
  // Add common filename pattern suggestions
  if (!basename.includes('.')) {
    suggestions.push(`The file may need an extension (e.g., ${basename}.txt, ${basename}.md)`);
  }
  
  // Add general suggestions
  suggestions.push(
    `Use a relative path (./path/to/file.txt) or absolute path (/full/path/to/file.txt)`,
    `Make sure the file has read permissions`
  );
  
  // Add examples
  const examples = [
    `path/to/${basename}.txt`,
    `./path/to/${basename}.txt`,
    `${currentDir}/${basename}.txt`
  ];
  
  return new FileSystemError(message, {
    filePath,
    suggestions,
    examples
  });
}