/**
 * Path normalization utilities for use in both tests and application code.
 *
 * These utilities handle cross-platform path normalization, ensuring consistent
 * formatting of paths regardless of operating system. They are particularly useful
 * for test assertions, path comparisons, and interacting with external libraries
 * that have specific path format requirements.
 */
import path from 'path';

/**
 * Normalizes a path to use forward slashes consistently and handles leading slash
 * according to the specified option.
 *
 * @param inputPath - The path to normalize
 * @param keepLeadingSlash - Whether to preserve/add a leading slash (default: false)
 * @returns The normalized path
 *
 * @example
 * // Basic usage (no leading slash)
 * normalizePathGeneral('path\\to\\file') // returns 'path/to/file'
 *
 * @example
 * // With leading slash
 * normalizePathGeneral('path/to/file', true) // returns '/path/to/file'
 */
export function normalizePathGeneral(inputPath: string, keepLeadingSlash = false): string {
  if (!inputPath) return '.';

  // Replace any backslashes with forward slashes
  let normalized = inputPath.replace(/\\/g, '/');

  // Normalize path using path.posix to handle . and .. segments
  normalized = path.posix.normalize(normalized);

  // Handle specific case for ./
  if (normalized === './') {
    normalized = '.';
  }

  // Handle leading slash based on parameter
  if (!keepLeadingSlash && normalized.startsWith('/')) {
    normalized = normalized.substring(1);
  } else if (keepLeadingSlash && !normalized.startsWith('/')) {
    normalized = '/' + normalized;
  }

  return normalized;
}

/**
 * Normalizes two paths to a common format for accurate comparison.
 * Both paths will be returned in the same format (both with or without leading slashes).
 *
 * @param path1 - First path to compare
 * @param path2 - Second path to compare
 * @returns Tuple of [normalizedPath1, normalizedPath2]
 *
 * @example
 * const [nPath1, nPath2] = normalizePathsForComparison('/path\\to/file.txt', 'path/to\\file.txt');
 * expect(nPath1).toBe(nPath2); // would pass because both are normalized to 'path/to/file.txt'
 */
export function normalizePathsForComparison(path1: string, path2: string): [string, string] {
  // Normalize both paths without leading slashes for consistent comparison
  return [normalizePathGeneral(path1, false), normalizePathGeneral(path2, false)];
}

/**
 * Creates a properly formatted path for gitignore pattern matching.
 * Makes the path relative to the given base path and ensures forward slashes.
 *
 * The ignore library requires paths to be:
 * 1. Relative (not absolute)
 * 2. Using forward slashes
 * 3. Without a leading './' prefix
 *
 * @param inputPath - The path to normalize
 * @param basePath - The base path that contains the .gitignore file
 * @returns The normalized path, relative to basePath
 *
 * @example
 * normalizePathForGitignore('/project/src/file.js', '/project')
 * // returns 'src/file.js'
 */
export function normalizePathForGitignore(inputPath: string, basePath: string): string {
  // Convert to consistent forward slashes first
  const forwardSlashInput = inputPath.replace(/\\/g, '/');
  const forwardSlashBase = basePath.replace(/\\/g, '/');

  // Use path.posix to ensure consistent forward slash behavior
  let relativePath = path.posix.relative(forwardSlashBase, forwardSlashInput);

  // Handle empty relative path (input is the base path)
  if (!relativePath) {
    return '.';
  }

  // Remove leading ./ if present
  if (relativePath.startsWith('./')) {
    relativePath = relativePath.substring(2);
  }

  return relativePath;
}
