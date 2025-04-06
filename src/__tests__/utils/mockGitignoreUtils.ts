/**
 * Mock utilities for gitignore-based file filtering in tests
 * Provides a consistent interface for mocking gitignore utilities
 */
import { jest } from '@jest/globals';
import * as gitignoreUtils from '../../utils/gitignoreUtils';
import { Ignore } from 'ignore';

/**
 * Interface for test results returned by ignore.test() and ignore.checkIgnore()
 * Based on the TestResult interface from the ignore package
 */
interface TestResult {
  ignored: boolean;
  unignored: boolean;
  rule?: {
    pattern: string;
    negative?: boolean;
  };
}

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
 * Registry of path-specific rules for shouldIgnorePath function
 */
const shouldIgnorePathRules: IgnorePathRule[] = [];

/**
 * Registry of directory-specific ignore patterns for createIgnoreFilter function
 */
interface IgnoreFilterRule {
  directoryPath: string;
  ignorePatterns: string[] | ((path: string) => boolean);
}

const createIgnoreFilterRules: IgnoreFilterRule[] = [];

/**
 * Default configuration values for gitignore mocks
 */
const DEFAULT_CONFIG: GitignoreMockConfig = {
  defaultIgnoreBehavior: false, // By default, don't ignore paths
  defaultIgnorePatterns: [
    'node_modules',
    '.git',
    'dist',
    'build',
    'coverage',
    '.cache'
  ],
  defaultIncludePatterns: []
};

/**
 * Reset all gitignore mock functions to their initial state
 * This should be called before each test to prevent test pollution
 */
export function resetMockGitignore(): void {
  jest.clearAllMocks();
  
  // Reset specific behavior mocks
  mockedGitignoreUtils.shouldIgnorePath.mockReset();
  mockedGitignoreUtils.createIgnoreFilter.mockReset();
  mockedGitignoreUtils.clearIgnoreCache.mockReset();
  
  // Clear all path-specific configurations
  shouldIgnorePathRules.length = 0;
  createIgnoreFilterRules.length = 0;
}

/**
 * Utility function to create a mock ignore filter
 * @param ignorePatterns - Array of patterns to ignore
 * @param includePatterns - Array of patterns to always include
 * @returns A mock ignore filter object
 */
function createMockIgnoreFilter(ignorePatterns: string[] = [], includePatterns: string[] = []): Ignore {
  // Define the core ignores function
  const ignoresFunction = (path: string): boolean => {
    // First check include patterns (override ignore patterns)
    for (const pattern of includePatterns) {
      // Very simple pattern matching for testing purposes
      if (path === pattern || path.startsWith(pattern + '/')) {
        return false;
      }
    }
    
    // Then check ignore patterns
    for (const pattern of ignorePatterns) {
      // Very simple pattern matching for testing purposes
      if (path === pattern || 
          path.startsWith(pattern + '/') || 
          (pattern.startsWith('*.') && path.endsWith(pattern.substring(1)))) {
        return true;
      }
    }
    
    // Default to not ignoring
    return false;
  };

  // Create mock test result objects
  const createTestResult = (path: string): TestResult => {
    const ignored = ignoresFunction(path);
    return {
      ignored,
      unignored: !ignored
    };
  };

  // Create a proper mock Ignore implementation that matches the interface
  const mockIgnore = {
    // Create a self-returning mock function for the add method
    add: jest.fn().mockImplementation(function(this: any) { 
      return this;
    }),
    ignores: ignoresFunction,
    filter: jest.fn((paths: readonly string[]) => {
      return paths.filter(path => !ignoresFunction(path));
    }),
    createFilter: jest.fn(() => {
      return (path: string) => !ignoresFunction(path);
    }),
    test: jest.fn((path: string) => createTestResult(path)),
    checkIgnore: jest.fn((path: string) => createTestResult(path))
  } as Ignore;
  
  return mockIgnore;
}

/**
 * Configure the mocked gitignoreUtils module with default behaviors
 * @param config - Optional configuration to customize the default behaviors
 */
export function setupMockGitignore(config?: GitignoreMockConfig): void {
  // Merge provided config with defaults
  const mergedConfig = { ...DEFAULT_CONFIG, ...config };
  
  // Configure shouldIgnorePath
  mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (basePath, filePath) => {
    // Check if we have any path-specific rules
    for (const rule of shouldIgnorePathRules) {
      const stringPath = `${basePath}/${filePath}`.replace(/\/\//g, '/');
      const matches = 
        (typeof rule.pattern === 'string' && stringPath === rule.pattern) || 
        (typeof rule.pattern === 'string' && filePath === rule.pattern) ||
        (rule.pattern instanceof RegExp && rule.pattern.test(stringPath));
      
      if (matches) {
        return rule.ignored;
      }
    }
    
    // Fall back to default behavior
    return mergedConfig.defaultIgnoreBehavior || false;
  });
  
  // Configure createIgnoreFilter
  mockedGitignoreUtils.createIgnoreFilter.mockImplementation(async (directoryPath) => {
    // Check if we have any directory-specific rules
    for (const rule of createIgnoreFilterRules) {
      if (rule.directoryPath === directoryPath) {
        if (Array.isArray(rule.ignorePatterns)) {
          return createMockIgnoreFilter(
            rule.ignorePatterns, 
            mergedConfig.defaultIncludePatterns || []
          );
        } else {
          // For function-based rules, create a custom ignores function
          const ignorePatternFn = rule.ignorePatterns as (path: string) => boolean;
          
          // Create mock test result objects for function-based rules
          const createTestResult = (path: string): TestResult => {
            const ignored = ignorePatternFn(path);
            return {
              ignored,
              unignored: !ignored
            };
          };
          
          return {
            ignores: ignorePatternFn,
            add: jest.fn().mockImplementation(function(this: any) { 
              return this;
            }),
            filter: jest.fn((paths: readonly string[]) => {
              return paths.filter(path => !ignorePatternFn(path));
            }),
            createFilter: jest.fn(() => {
              return (path: string) => !ignorePatternFn(path);
            }),
            test: jest.fn((path: string) => createTestResult(path)),
            checkIgnore: jest.fn((path: string) => createTestResult(path))
          } as Ignore;
        }
      }
    }
    
    // If no specific rule matched, use the default configuration
    return createMockIgnoreFilter(
      mergedConfig.defaultIgnorePatterns || [], 
      mergedConfig.defaultIncludePatterns || []
    );
  });
  
  // Configure clearIgnoreCache (simple mock)
  mockedGitignoreUtils.clearIgnoreCache.mockImplementation(() => {
    // No actual implementation needed for most tests
  });
}

/**
 * Configure shouldIgnorePath to return specific results for given path patterns
 * @param pathPattern - Path or regex pattern to match
 * @param ignored - Whether matching paths should be ignored
 */
export const mockShouldIgnorePath: MockShouldIgnorePathFunction = 
  (pathPattern: string | RegExp, ignored: boolean): void => {
    // Find and remove any existing rule with the same pattern
    const existingIndex = shouldIgnorePathRules.findIndex(rule => 
      (typeof rule.pattern === 'string' && rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
        pathPattern instanceof RegExp && 
        rule.pattern.toString() === pathPattern.toString())
    );
    
    if (existingIndex !== -1) {
      shouldIgnorePathRules.splice(existingIndex, 1);
    }
    
    // Add new rule at the beginning for higher precedence
    shouldIgnorePathRules.unshift({
      pattern: pathPattern,
      ignored
    });
  };

/**
 * Configure createIgnoreFilter to use specific ignore rules
 * Implementation will be added in a future task
 */
export const mockCreateIgnoreFilter: MockCreateIgnoreFilterFunction =
  (_directoryPath: string, _ignorePatterns: string[] | ((path: string) => boolean)): void => {
    // Implementation will be added in a future task
  };