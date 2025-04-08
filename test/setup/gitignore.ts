/**
 * Gitignore setup utilities for tests
 * 
 * This module provides specialized setup helpers for gitignore-related tests,
 * building on the virtual filesystem (memfs) approach.
 */
import path from 'path';
import { addVirtualGitignoreFile } from '../../src/__tests__/utils/virtualFsUtils';
import { clearIgnoreCache, shouldIgnorePath } from '../../src/utils/gitignoreUtils';
import { normalizePathGeneral } from '../../src/utils/pathUtils';
import { setupBasicFs } from './fs';

/**
 * Creates a virtual .gitignore file
 * 
 * @param gitignorePath - Path where the .gitignore file should be created
 * @param content - Content of the .gitignore file
 * @returns Promise that resolves when the file has been created
 * 
 * Usage:
 * ```typescript
 * await addGitignoreFile('/project/.gitignore', '*.log\n/node_modules/');
 * ```
 */
export async function addGitignoreFile(gitignorePath: string, content: string): Promise<void> {
  const normalizedPath = normalizePathGeneral(gitignorePath, true);
  await addVirtualGitignoreFile(normalizedPath, content);
}

/**
 * Sets up a basic gitignore environment in the virtual filesystem
 * 
 * @param projectPath - Base path for the project
 * @param patterns - Default gitignore patterns
 * @returns Promise that resolves when setup is complete
 * 
 * Usage:
 * ```typescript
 * await setupBasicGitignore();
 * // or
 * await setupBasicGitignore('/custom/path', '*.txt\n/build/');
 * ```
 */
export async function setupBasicGitignore(
  projectPath: string = '/project', 
  patterns: string = 'node_modules/\n*.log\n.DS_Store\n/dist/'
): Promise<void> {
  const normalizedPath = normalizePathGeneral(projectPath, true);
  await addGitignoreFile(path.join(normalizedPath, '.gitignore'), patterns);
}

/**
 * Sets up a directory with files and a .gitignore file
 * 
 * @param dirPath - The base directory path
 * @param gitignoreContent - Content for the .gitignore file
 * @param files - Files to create within the directory
 * @param options - Additional options (reset: boolean)
 * @returns Promise that resolves when setup is complete
 * 
 * Usage:
 * ```typescript
 * await setupWithGitignore('/project', '*.log\nnode_modules/', {
 *   'src/index.ts': 'console.log("Hello");',
 *   'debug.log': 'This should be ignored'
 * });
 * 
 * // Add more files without resetting existing ones
 * await setupWithGitignore('/project/subdir', '*.tmp', {
 *   'temp.tmp': 'Temp file'
 * }, { reset: false });
 * ```
 */
export async function setupWithGitignore(
  dirPath: string,
  gitignoreContent: string,
  files: Record<string, string>,
  options: { reset: boolean } = { reset: true }
): Promise<void> {
  // Normalize the directory path
  const normalizedPath = normalizePathGeneral(dirPath, true);
  
  // Create files structure with full paths
  const structure: Record<string, string> = {};
  
  // Add files with full paths
  Object.entries(files).forEach(([relativePath, content]) => {
    const fullPath = normalizePathGeneral(path.join(normalizedPath, relativePath), true);
    structure[fullPath] = content;
  });
  
  // Set up the filesystem with the specified reset option
  setupBasicFs(structure, options);
  
  // Add .gitignore file
  await addGitignoreFile(path.join(normalizedPath, '.gitignore'), gitignoreContent);
}

/**
 * Creates a test structure with multiple .gitignore files in different directories
 * 
 * @param baseDir - The base directory for the test project
 * @param gitignoreFiles - Object mapping relative paths to gitignore content
 * @param projectFiles - Object mapping relative paths to file content
 * @returns Promise that resolves when setup is complete
 * 
 * Usage:
 * ```typescript
 * await setupMultiGitignore('/project', {
 *   '.gitignore': '*.log',
 *   'src/.gitignore': '*.tmp'
 * }, {
 *   'src/index.js': 'console.log("Hello");',
 *   'src/temp.tmp': 'Will be ignored by src/.gitignore',
 *   'log.log': 'Will be ignored by .gitignore at root'
 * });
 * ```
 */
export async function setupMultiGitignore(
  baseDir: string,
  gitignoreFiles: Record<string, string>,
  projectFiles: Record<string, string>
): Promise<void> {
  // Normalize the base directory
  const normalizedBaseDir = normalizePathGeneral(baseDir, true);
  
  // Set up project files
  setupBasicFs(
    Object.entries(projectFiles).reduce((acc, [relativePath, content]) => {
      const fullPath = normalizePathGeneral(path.join(normalizedBaseDir, relativePath), true);
      acc[fullPath] = content;
      return acc;
    }, {} as Record<string, string>)
  );
  
  // Add all gitignore files
  for (const [relativePath, content] of Object.entries(gitignoreFiles)) {
    const fullPath = normalizePathGeneral(path.join(normalizedBaseDir, relativePath), true);
    await addGitignoreFile(fullPath, content);
  }
  
  // Clear the cache to ensure clean state
  clearIgnoreCache();
}

/**
 * Creates a test helper to check if paths should be ignored
 * 
 * @param baseDir - The base directory to check paths against
 * @returns Function that checks if a relative path should be ignored
 * 
 * Usage:
 * ```typescript
 * const shouldIgnore = createIgnoreChecker('/project');
 * const isIgnored = await shouldIgnore('node_modules/file.js');
 * expect(isIgnored).toBe(true);
 * ```
 */
export function createIgnoreChecker(baseDir: string): (relativePath: string) => Promise<boolean> {
  const normalizedBaseDir = normalizePathGeneral(baseDir, true);
  
  return async (relativePath: string): Promise<boolean> => {
    const fullPath = normalizePathGeneral(path.join(normalizedBaseDir, relativePath), true);
    return shouldIgnorePath(normalizedBaseDir, fullPath);
  };
}
