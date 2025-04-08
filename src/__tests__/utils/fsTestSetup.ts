/**
 * Reusable utilities for filesystem mocking in tests
 * 
 * This module provides helper functions for common filesystem mock setups,
 * reducing duplication across test files and standardizing the approach.
 */
import path from 'path';
import fs from 'fs/promises';
import { 
  resetVirtualFs, 
  createVirtualFs, 
  addVirtualGitignoreFile
} from './virtualFsUtils';
import { normalizePathGeneral } from '../../utils/pathUtils';

/**
 * Sets up a basic filesystem with the specified file structure
 * 
 * This function resets the virtual filesystem and then creates
 * the specified files with their content.
 * 
 * @param structure - An object mapping file paths to their content
 * @example
 * ```typescript
 * setupBasicFiles({
 *   '/path/to/file.txt': 'Content of file',
 *   '/path/to/another.js': 'console.log("Hello");'
 * });
 * ```
 * 
 * @note This function resets the virtual filesystem before creating files.
 * If you need to set up multiple structures in a single test without
 * resetting between them, use createVirtualFs with reset: false instead.
 */
export function setupBasicFiles(structure: Record<string, string>): void {
  resetVirtualFs();
  createVirtualFs(structure);
}

/**
 * Sets up a standard project structure with customizable files
 * 
 * This function creates a standard project structure at the specified
 * base path, and adds the provided files to it.
 * 
 * @param basePath - The root path where the project will be created
 * @param files - An object mapping relative file paths to their content
 * @example
 * ```typescript
 * setupProjectStructure('/project', {
 *   'src/index.ts': 'console.log("Hello");',
 *   'README.md': '# Project'
 * });
 * ```
 */
export function setupProjectStructure(
  basePath: string, 
  files: Record<string, string>
): void {
  // Normalize the base path
  const normalizedBasePath = normalizePathGeneral(basePath, true);
  
  // Create structure object with full paths
  const structure: Record<string, string> = {};
  
  // Add files with full paths
  Object.entries(files).forEach(([relativePath, content]) => {
    const fullPath = normalizePathGeneral(path.join(normalizedBasePath, relativePath), true);
    structure[fullPath] = content;
  });
  
  // Set up the filesystem
  setupBasicFiles(structure);
}

/**
 * Sets up a directory with files and a .gitignore file
 * 
 * This function creates the specified directory structure with files
 * and also adds a .gitignore file with the given content.
 * 
 * @param dirPath - The base directory path
 * @param gitignoreContent - Content for the .gitignore file
 * @param files - Files to create within the directory
 * @example
 * ```typescript
 * await setupWithGitignore('/project', '*.log\nnode_modules/', {
 *   'src/index.ts': 'console.log("Hello");',
 *   'debug.log': 'This should be ignored'
 * });
 * ```
 * 
 * @note This function resets the virtual filesystem before creating the project.
 * If you need to set up multiple directories with gitignore files in a single test,
 * use setupProjectStructure and addVirtualGitignoreFile separately to avoid resets.
 */
export async function setupWithGitignore(
  dirPath: string,
  gitignoreContent: string,
  files: Record<string, string>
): Promise<void> {
  // Set up the basic directory structure with files
  setupProjectStructure(dirPath, files);
  
  // Add .gitignore file
  const gitignorePath = normalizePathGeneral(path.join(dirPath, '.gitignore'), true);
  await addVirtualGitignoreFile(gitignorePath, gitignoreContent);
}

/**
 * Mocks the fileExists function to work with the virtual filesystem
 * 
 * This function takes a mock function and configures it to check
 * for file existence using the virtual filesystem's fs.access method.
 * 
 * @param mockFn - The jest mock function for fileExists
 * @example
 * ```typescript
 * // In beforeEach:
 * const mockedFileExists = jest.mocked(fileReader.fileExists);
 * mockFileExists(mockedFileExists);
 * ```
 */
export function mockFileExists(mockFn: jest.Mock): void {
  mockFn.mockImplementation(async (filePath: string) => {
    try {
      // Use fs.access which is already mocked by memfs
      await fs.access(filePath);
      return true;
    } catch (error) {
      return false;
    }
  });
}

/**
 * Common pattern for mocking gitignore functionality
 * 
 * This function sets up the gitignore cache clearing and fileExists
 * mock that is commonly used in gitignore-related tests.
 * 
 * @param gitignoreUtils - The gitignoreUtils module to clear cache
 * @param mockedFileExists - The mocked fileExists function
 * @example
 * ```typescript
 * // In beforeEach:
 * setupGitignoreMocking(gitignoreUtils, mockedFileExists);
 * ```
 */
export function setupGitignoreMocking(
  gitignoreUtils: { clearIgnoreCache: () => void },
  mockedFileExists: jest.Mock
): void {
  // Clear gitignore cache
  gitignoreUtils.clearIgnoreCache();
  
  // Mock fileExists to use virtual filesystem
  mockFileExists(mockedFileExists);
}

/**
 * Sets up a reusable beforeEach hook for cache clearing in tests
 * 
 * This helper creates a consistent pattern for clearing caches in tests
 * that use gitignore utilities. It returns a function that can be used
 * in beforeEach hooks across test files.
 * 
 * @param gitignoreUtils - The gitignoreUtils module to clear cache
 * @param mockedFileExists - The mocked fileExists function
 * @returns A function that can be used in beforeEach hooks
 * @example
 * ```typescript
 * // At the top of your test file:
 * const clearCachesBeforeEach = setupCacheClearing(gitignoreUtils, mockedFileExists);
 * 
 * // In your describe blocks:
 * describe('your tests', () => {
 *   beforeEach(clearCachesBeforeEach);
 *   
 *   // Tests that rely on a clean cache
 * });
 * ```
 */
export function setupCacheClearing(
  gitignoreUtils: { clearIgnoreCache: () => void },
  mockedFileExists: jest.Mock
): () => void {
  return () => {
    // Clear all mocks
    jest.clearAllMocks();
    
    // Reset virtual filesystem
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Set up fileExists mock
    mockFileExists(mockedFileExists);
  };
}
