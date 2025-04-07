/**
 * Type definitions for gitignore mock setup
 */

/**
 * Configuration options for gitignore mocks
 */
export interface GitignoreMockConfig {
  /**
   * Default behavior for shouldIgnorePath when no specific rule matches
   * If true, paths are ignored by default
   * If false, paths are included by default
   */
  defaultIgnoreBehavior?: boolean;

  /**
   * Default patterns to always ignore, in addition to path-specific rules
   */
  defaultIgnorePatterns?: string[];

  /**
   * Default patterns to always include, overriding other rules
   */
  defaultIncludePatterns?: string[];
}

/**
 * Sets up basic gitignore mocking with standard ignored patterns
 * @param config - Configuration options for mock setup
 */
export function setupBasicGitignore(config?: GitignoreMockConfig): void;

/**
 * Configures gitignore mocking based on virtual filesystem .gitignore files
 * @param rootPath - Root path to start scanning from
 */
export function setupGitignoreFromVirtualFs(rootPath?: string): void;

/**
 * Creates a virtual .gitignore file and configures mocks based on its content
 * @param gitignorePath - Path where the .gitignore file should be created
 * @param content - Content of the .gitignore file
 */
export function addGitignoreFile(gitignorePath: string, content: string): void;
