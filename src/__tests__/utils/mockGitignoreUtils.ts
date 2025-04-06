/**
 * Mock utilities for gitignore-based file filtering in tests
 * Provides a consistent interface for mocking gitignore utilities
 */
import { jest } from '@jest/globals';
import * as gitignoreUtils from '../../utils/gitignoreUtils';

// Mock gitignoreUtils module
jest.mock('../../utils/gitignoreUtils');

/**
 * Re-export the mocked gitignoreUtils module for direct access if needed
 */
export const mockedGitignoreUtils = jest.mocked(gitignoreUtils);

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
 * Rule for path-specific ignore behavior
 */
export interface IgnorePathRule {
  /**
   * Path pattern to match (string or regular expression)
   */
  pattern: string | RegExp;

  /**
   * Whether matching paths should be ignored
   */
  ignored: boolean;
}

/**
 * Interface for the mockShouldIgnorePath function
 */
export interface MockShouldIgnorePathFunction {
  /**
   * Configures gitignoreUtils.shouldIgnorePath to return specific results for given path patterns
   * @param pathPattern - Path or regex pattern to match
   * @param ignored - Whether matching paths should be ignored
   */
  (pathPattern: string | RegExp, ignored: boolean): void;
}

/**
 * Interface for the mockCreateIgnoreFilter function
 */
export interface MockCreateIgnoreFilterFunction {
  /**
   * Configures gitignoreUtils.createIgnoreFilter to use specific ignore rules
   * @param directoryPath - The directory path to configure ignore rules for
   * @param ignorePatterns - Array of patterns to ignore, or a map function that determines if a path should be ignored
   */
  (
    directoryPath: string, 
    ignorePatterns: string[] | ((path: string) => boolean)
  ): void;
}

/**
 * Reset all gitignore mock functions to their initial state
 * This should be called before each test to prevent test pollution
 */
export function resetMockGitignore(): void {
  // Implementation will be added in a future task
}

/**
 * Configure the mocked gitignoreUtils module with default behaviors
 * @param config - Optional configuration to customize the default behaviors
 */
export function setupMockGitignore(_config?: GitignoreMockConfig): void {
  // Implementation will be added in a future task
}

/**
 * Configure shouldIgnorePath to return specific results for given path patterns
 * Implementation will be added in a future task
 */
export const mockShouldIgnorePath: MockShouldIgnorePathFunction = 
  (_pathPattern: string | RegExp, _ignored: boolean): void => {
    // Implementation will be added in a future task
  };

/**
 * Configure createIgnoreFilter to use specific ignore rules
 * Implementation will be added in a future task
 */
export const mockCreateIgnoreFilter: MockCreateIgnoreFilterFunction =
  (_directoryPath: string, _ignorePatterns: string[] | ((path: string) => boolean)): void => {
    // Implementation will be added in a future task
  };