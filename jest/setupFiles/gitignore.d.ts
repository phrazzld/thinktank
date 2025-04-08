/**
 * Type definitions for gitignore mock setup
 */

/**
 * Creates a virtual .gitignore file
 * 
 * @param gitignorePath - Path where the .gitignore file should be created
 * @param content - Content of the .gitignore file
 * @returns Promise that resolves when the file has been created
 */
export function addGitignoreFile(gitignorePath: string, content: string): Promise<void>;

/**
 * Sets up a basic gitignore environment in the virtual filesystem
 * 
 * @param projectPath - Base path for the project
 * @param patterns - Default gitignore patterns
 * @returns Promise that resolves when setup is complete
 */
export function setupBasicGitignore(
  projectPath?: string, 
  patterns?: string
): Promise<void>;

/**
 * Clears the gitignore cache to ensure test isolation
 * 
 * This should be called in beforeEach to prevent test interdependencies
 */
export function clearGitignoreCache(): void;

/**
 * Creates a mock gitignore result for a path
 * 
 * @param shouldIgnore - Whether the path should be ignored
 * @returns A mock function that returns the specified result
 */
export function createGitignoreMock(shouldIgnore?: boolean): jest.Mock;
