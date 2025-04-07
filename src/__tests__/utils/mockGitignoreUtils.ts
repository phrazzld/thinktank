/**
 * Mock utilities for gitignore-based file filtering in tests.
 * 
 * This module provides utilities for mocking gitignore pattern matching and
 * directory filtering functionality, allowing tests to simulate .gitignore
 * behavior without relying on actual filesystem operations. It also provides
 * integration with the virtual filesystem for a more realistic testing approach.
 * 
 * @module mockGitignoreUtils
 * 
 * @example
 * Basic usage with direct mock configuration:
 * ```typescript
 * import { 
 *   resetMockGitignore, 
 *   setupMockGitignore, 
 *   mockShouldIgnorePath,
 *   mockCreateIgnoreFilter 
 * } from '../../../__tests__/utils/mockGitignoreUtils';
 *
 * // Reset and setup mocks before each test
 * beforeEach(() => {
 *   resetMockGitignore();
 *   setupMockGitignore();
 *   
 *   // Configure specific mock behaviors
 *   mockShouldIgnorePath(/\.log$/, true); // Ignore all log files
 *   mockCreateIgnoreFilter('/project', ['node_modules', '*.tmp']);
 * });
 * ```
 * 
 * Integration with virtual filesystem:
 * ```typescript
 * import { 
 *   resetVirtualFs, 
 *   createVirtualFs 
 * } from '../../../__tests__/utils/virtualFsUtils';
 * import { 
 *   resetMockGitignore, 
 *   setupMockGitignore,
 *   configureMockGitignoreFromVirtualFs,
 *   addVirtualGitignoreFile
 * } from '../../../__tests__/utils/mockGitignoreUtils';
 * 
 * beforeEach(() => {
 *   // Setup virtual filesystem
 *   resetVirtualFs();
 *   createVirtualFs({
 *     'project/file.txt': 'content',
 *     'project/logs/app.log': 'log content'
 *   });
 *   
 *   // Reset and setup gitignore mocks
 *   resetMockGitignore();
 *   setupMockGitignore();
 *   
 *   // Add .gitignore files to virtual filesystem and configure mocks
 *   addVirtualGitignoreFile('project/.gitignore', '*.log\ntmp/');
 *   
 *   // Or configure mocks from existing virtual .gitignore files
 *   configureMockGitignoreFromVirtualFs();
 * });
 * ```
 */
import { jest } from '@jest/globals';
import * as gitignoreUtils from '../../utils/gitignoreUtils';
import { Ignore } from 'ignore';
import { getVirtualFs } from './virtualFsUtils';
import path from 'path';

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
 * Resets all gitignore mock functions to their initial state.
 * 
 * This function should be called in the `beforeEach` hook of your tests to prevent
 * test cross-contamination. It clears all mock implementations, mock calls history,
 * and path-specific rules that may have been configured.
 * 
 * @example
 * ```typescript
 * beforeEach(() => {
 *   resetMockGitignore();
 *   setupMockGitignore();
 * });
 * ```
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
 * Configures the mocked gitignoreUtils module with default behaviors.
 * 
 * This function sets up mock implementations for gitignore-related functions,
 * applying the provided configuration or default values. It should be called
 * after `resetMockGitignore()` to establish baseline behavior for gitignore operations.
 * 
 * @param config - Optional configuration to customize the default behaviors
 * 
 * @example
 * ```typescript
 * // Setup with defaults (don't ignore files by default)
 * setupMockGitignore();
 * 
 * // Setup with custom defaults
 * setupMockGitignore({
 *   defaultIgnoreBehavior: false,
 *   defaultIgnorePatterns: ['node_modules', '.git', 'dist', '*.log'],
 *   defaultIncludePatterns: ['important.log']
 * });
 * ```
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
 * Helper function to normalize file paths for virtual filesystem.
 * Removes leading slash if present, since memfs expects paths without leading slashes.
 * 
 * @param filePath - The file path to normalize
 * @returns Normalized file path
 * 
 * @example
 * ```typescript
 * normalizeVirtualFsPath('/path/to/file.txt') // Returns 'path/to/file.txt'
 * normalizeVirtualFsPath('path/to/file.txt')  // Returns 'path/to/file.txt' (unchanged)
 * ```
 */
export function normalizeVirtualFsPath(filePath: string): string {
  return filePath.startsWith('/') ? filePath.substring(1) : filePath;
}

/**
 * Creates a .gitignore file in the virtual filesystem and configures mocks based on its content.
 * 
 * This function creates a .gitignore file at the specified path in the virtual
 * filesystem and then configures the mock gitignore utilities to respect these
 * patterns when performing ignore checks.
 * 
 * @param gitignorePath - Path where the .gitignore file should be created
 * @param content - Content of the .gitignore file (ignore patterns, one per line)
 * 
 * @example
 * ```typescript
 * // Add a .gitignore file that ignores log files and the tmp directory
 * addVirtualGitignoreFile('project/.gitignore', '*.log\ntmp/');
 * 
 * // Test that the patterns are respected
 * const shouldIgnore = await gitignoreUtils.shouldIgnorePath('project', 'app.log');
 * expect(shouldIgnore).toBe(true);
 * ```
 */
export function addVirtualGitignoreFile(gitignorePath: string, content: string): void {
  // Normalize path for the virtual filesystem
  const normalizedPath = normalizeVirtualFsPath(gitignorePath);
  
  // Get virtual filesystem reference
  const virtualFs = getVirtualFs();
  
  // Create the directory if it doesn't exist
  const dirPath = path.dirname(normalizedPath);
  virtualFs.mkdirSync(dirPath, { recursive: true });
  
  // Write the .gitignore file
  virtualFs.writeFileSync(normalizedPath, content);
  
  // Parse the patterns
  const patterns = content.split('\n')
    .map(line => line.trim())
    .filter(line => line && !line.startsWith('#'));
  
  // Configure mock for the containing directory
  const baseDirPath = '/' + dirPath; // Add leading slash for mock config
  
  // Configure the mock to always return true for matched patterns
  mockCreateIgnoreFilter(baseDirPath, patterns);
  
  // Explicitly configure shouldIgnorePath for each pattern
  patterns.forEach(pattern => {
    if (pattern.startsWith('*.')) {
      // For file extension patterns
      const extension = pattern.substring(1); // e.g., "*.log" -> ".log"
      mockShouldIgnorePath(new RegExp(`\\${extension}$`), true);
    } else if (pattern.endsWith('/')) {
      // For directory patterns
      const dirName = pattern.substring(0, pattern.length - 1);
      mockShouldIgnorePath(new RegExp(`${dirName}/`), true);
    } else {
      // For other patterns
      mockShouldIgnorePath(pattern, true);
    }
  });
}

/**
 * Scans the virtual filesystem for .gitignore files and configures mocks based on their content.
 * 
 * This function finds all .gitignore files in the virtual filesystem and configures
 * the mock gitignore utilities to respect the patterns they contain. This allows
 * tests to use a more realistic approach to gitignore filtering by creating actual
 * .gitignore files in the virtual filesystem.
 * 
 * @param rootPath - Optional root path to start scanning from (defaults to '/')
 * 
 * @example
 * ```typescript
 * // Setup virtual filesystem with .gitignore files
 * createVirtualFs({
 *   'project/.gitignore': '*.log\ntmp/',
 *   'project/api/.gitignore': '*.cache',
 * });
 * 
 * // Configure mocks based on these .gitignore files
 * configureMockGitignoreFromVirtualFs();
 * 
 * // Test that the patterns are respected
 * const shouldIgnore1 = await gitignoreUtils.shouldIgnorePath('project', 'app.log');
 * expect(shouldIgnore1).toBe(true);
 * 
 * const shouldIgnore2 = await gitignoreUtils.shouldIgnorePath('project/api', 'data.cache');
 * expect(shouldIgnore2).toBe(true);
 * ```
 */
export function configureMockGitignoreFromVirtualFs(rootPath: string = '/'): void {
  // Normalize root path for the virtual filesystem
  const normalizedRootPath = normalizeVirtualFsPath(rootPath);
  
  // Get virtual filesystem reference
  const virtualFs = getVirtualFs();
  
  // Find all .gitignore files recursively
  findGitignoreFiles(normalizedRootPath, virtualFs);
}

/**
 * Recursively finds .gitignore files in the virtual filesystem and configures mocks.
 * 
 * @param dirPath - Directory path to search in
 * @param virtualFs - Reference to the virtual filesystem
 */
function findGitignoreFiles(dirPath: string, virtualFs: any): void {
  try {
    // Check for a .gitignore file in this directory
    const gitignorePath = path.join(dirPath, '.gitignore');
    
    if (virtualFs.existsSync(gitignorePath)) {
      try {
        // Read the file content
        const content = virtualFs.readFileSync(gitignorePath, 'utf-8');
        
        // Parse the patterns
        const patterns = content.split('\n')
          .map((line: string) => line.trim())
          .filter((line: string) => line && !line.startsWith('#'));
        
        // Configure mock for this directory
        mockCreateIgnoreFilter(dirPath, patterns);
      } catch (error) {
        // If there's an error reading the file, just continue
        console.warn(`Warning: Could not read virtual .gitignore file at ${gitignorePath}.`);
      }
    }
    
    // Recursively check subdirectories
    try {
      const entries = virtualFs.readdirSync(dirPath, { withFileTypes: true });
      
      for (const entry of entries) {
        // Skip non-directories and common directories to ignore
        if (!entry.isDirectory() || 
            entry.name === 'node_modules' || 
            entry.name === '.git') {
          continue;
        }
        
        const subdirPath = path.join(dirPath, entry.name);
        findGitignoreFiles(subdirPath, virtualFs);
      }
    } catch (error) {
      // If there's an error reading the directory, just continue
      console.warn(`Warning: Could not read virtual directory at ${dirPath}.`);
    }
  } catch (error) {
    // If any unexpected error occurs, just continue
    console.warn(`Warning: Error processing virtual directory ${dirPath}: ${error}`);
  }
}

/**
 * Configures shouldIgnorePath to return specific results for given path patterns.
 * 
 * This function allows you to specify which paths should be ignored by the
 * gitignore filtering functionality, based on exact strings or regex patterns.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param ignored - Whether matching paths should be ignored (true) or included (false)
 * 
 * @example
 * ```typescript
 * // Ignore all log files
 * mockShouldIgnorePath(/\.log$/, true);
 * 
 * // Ignore a specific file
 * mockShouldIgnorePath('build/output.js', true);
 * 
 * // Never ignore a specific important file, even if it would match
 * // other ignore patterns
 * mockShouldIgnorePath('important.log', false);
 * 
 * // Ignore all files in a specific directory
 * mockShouldIgnorePath(/node_modules\//, true);
 * ```
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
 * Configures createIgnoreFilter to use specific ignore rules for a directory.
 * 
 * This function allows you to customize the behavior of gitignore pattern
 * filtering for specific directories, either using an array of patterns or
 * a custom function for more complex logic.
 * 
 * @param directoryPath - The directory path to configure ignore rules for
 * @param ignorePatterns - Array of patterns to ignore, or a function that determines if a path should be ignored
 * 
 * @example
 * ```typescript
 * // Using an array of patterns
 * mockCreateIgnoreFilter('/home/project', [
 *   'node_modules',
 *   'dist',
 *   '*.log',
 *   'tmp'
 * ]);
 * 
 * // Using a custom function for more complex logic
 * mockCreateIgnoreFilter('/home/project', (path) => {
 *   return path.includes('node_modules') || 
 *          path.endsWith('.log') ||
 *          path.includes('.git');
 * });
 * 
 * // Later usage in tests
 * const filter = await gitignoreUtils.createIgnoreFilter('/home/project');
 * expect(filter.ignores('node_modules/package.json')).toBe(true);
 * expect(filter.ignores('src/index.js')).toBe(false);
 * ```
 */
export const mockCreateIgnoreFilter: MockCreateIgnoreFilterFunction =
  (directoryPath: string, ignorePatterns: string[] | ((path: string) => boolean)): void => {
    // Find and remove any existing rule for the same directory
    const existingIndex = createIgnoreFilterRules.findIndex(rule => 
      rule.directoryPath === directoryPath
    );
    
    if (existingIndex !== -1) {
      createIgnoreFilterRules.splice(existingIndex, 1);
    }
    
    // Add new rule at the beginning for higher precedence
    createIgnoreFilterRules.unshift({
      directoryPath,
      ignorePatterns
    });
  };