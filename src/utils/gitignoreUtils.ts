/**
 * Utilities for handling .gitignore-based file filtering
 *
 * This module provides functionality for loading, parsing, and applying
 * .gitignore rules to filter files and directories during context gathering.
 */
import fs from 'fs/promises';
import path from 'path';
import ignore from 'ignore';
import { fileExists } from './fileReader';
import { normalizePathForGitignore } from './pathUtils';

// Cache of parsed gitignore patterns by directory path
// This avoids re-reading and re-parsing gitignore files for the same directories
const ignoreCache = new Map<string, ignore.Ignore>();

// Default patterns to ignore even if no .gitignore file exists
const DEFAULT_IGNORE_PATTERNS = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'coverage',
  '.cache',
  '.next',
  '.nuxt',
  '.output',
  '.vscode',
  '.idea',
];

/**
 * Creates an ignore filter that respects .gitignore rules for a given directory
 *
 * @param directoryPath - The directory path to load .gitignore rules from
 * @returns A Promise resolving to an ignore.Ignore instance that can be used to filter paths
 */
export async function createIgnoreFilter(directoryPath: string): Promise<ignore.Ignore> {
  // Check if we already have a cached ignore filter for this directory
  if (ignoreCache.has(directoryPath)) {
    return ignoreCache.get(directoryPath)!;
  }

  // Create a new ignore filter with default patterns
  const ignoreFilter = ignore();

  // Add default patterns
  ignoreFilter.add(DEFAULT_IGNORE_PATTERNS);

  // Try to load .gitignore file if it exists
  const gitignorePath = path.join(directoryPath, '.gitignore');

  // Check if file exists - this is important to ensure the mock is called in tests
  const exists = await fileExists(gitignorePath);

  if (exists) {
    try {
      // Read and parse the .gitignore file
      const gitignoreContent = await fs.readFile(gitignorePath, 'utf-8');

      // Add patterns from .gitignore file
      ignoreFilter.add(gitignoreContent);
    } catch (error) {
      // If there's an error reading the .gitignore file, just continue with default patterns
      console.warn(
        `Warning: Could not read .gitignore file at ${gitignorePath}. Using default ignore patterns.`
      );
    }
  }

  // Cache the ignore filter for future use
  ignoreCache.set(directoryPath, ignoreFilter);

  return ignoreFilter;
}

/**
 * Checks if a path should be ignored based on .gitignore rules
 *
 * @param basePath - The base directory path containing the .gitignore file
 * @param filePath - The path to check (relative to basePath)
 * @returns A Promise resolving to true if the path should be ignored, false otherwise
 */
export async function shouldIgnorePath(basePath: string, filePath: string): Promise<boolean> {
  const ignoreFilter = await createIgnoreFilter(basePath);

  // Use dedicated utility to normalize path for gitignore pattern matching
  const normalizedPath = normalizePathForGitignore(filePath, basePath);

  // The ignore library returns boolean for .ignores(), but we'll double-check
  // to ensure our tests and implementation work consistently
  return Boolean(ignoreFilter.ignores(normalizedPath));
}

/**
 * Clears the ignore filter cache
 *
 * This is mainly useful for testing or when .gitignore files might change
 * during the application's lifecycle.
 */
export function clearIgnoreCache(): void {
  ignoreCache.clear();
}
