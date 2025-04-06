/**
 * Mock utilities for the Node.js fs/promises module in tests.
 * 
 * This module provides a comprehensive set of utilities for mocking filesystem 
 * operations in Jest tests, allowing precise control over file access, reads, writes,
 * stats, directory listings, and more.
 * 
 * @module mockFsUtils
 * 
 * @example
 * ```typescript
 * import { 
 *   resetMockFs, 
 *   setupMockFs, 
 *   mockReadFile, 
 *   mockStat 
 * } from '../../../__tests__/utils/mockFsUtils';
 *
 * // Reset and setup mocks before each test
 * beforeEach(() => {
 *   resetMockFs();
 *   setupMockFs();
 *   
 *   // Configure specific mock behaviors
 *   mockReadFile('/path/to/file.txt', 'File content');
 *   mockStat('/path/to/file.txt', { isFile: () => true });
 * });
 * ```
 */
import fs from 'fs/promises';
import { Stats } from 'fs';
import { jest } from '@jest/globals';

// Mock fs/promises module at the top level
jest.mock('fs/promises');

/**
 * Re-export the mocked fs module for direct access if needed
 */
export const mockedFs = jest.mocked(fs);

/**
 * Default configuration values for the mock filesystem
 */
const DEFAULT_CONFIG: FsMockConfig = {
  defaultAccessBehavior: true, // Access allowed by default
  defaultFileContent: '', // Empty string by default
  defaultStats: {
    isFile: () => true,
    isDirectory: () => false,
    size: 0,
    birthtime: new Date(),
    mtime: new Date(),
    atime: new Date()
  },
  defaultAccessErrorCode: 'ENOENT',
  defaultReadErrorCode: 'ENOENT',
  defaultWriteErrorCode: 'EACCES',
  defaultStatErrorCode: 'ENOENT'
};

/**
 * Access rule used by the mockAccess function
 * Defines the behavior when a path matches a specific pattern
 */
interface AccessRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Whether access should be allowed */
  allowed: boolean;
  /** Optional error code to use if denied */
  errorCode?: string;
  /** Optional error message to use if denied */
  errorMessage?: string;
}

/** Registry of path-specific access rules */
const accessRules: AccessRule[] = [];

/**
 * Read rule used by the mockReadFile function
 * Defines the behavior when a path matches a specific pattern
 */
interface ReadFileRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Content to return or error to throw */
  content: string | Buffer | Error;
}

/** Registry of path-specific read file rules */
const readFileRules: ReadFileRule[] = [];

/**
 * Stat rule used by the mockStat function
 * Defines the behavior when a path matches a specific pattern
 */
interface StatRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Stats to return or error to throw */
  result: MockedStats | Error;
}

/** Registry of path-specific stat rules */
const statRules: StatRule[] = [];

/**
 * Readdir rule used by the mockReaddir function
 * Defines the behavior when a path matches a specific pattern
 */
interface ReaddirRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Directory entries to return or error to throw */
  result: string[] | Error;
}

/** Registry of path-specific readdir rules */
const readdirRules: ReaddirRule[] = [];

/**
 * Mkdir rule used by the mockMkdir function
 * Defines the behavior when a path matches a specific pattern
 */
interface MkdirRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Whether mkdir should succeed (true) or fail (false or Error) */
  result: boolean | Error;
}

/** Registry of path-specific mkdir rules */
const mkdirRules: MkdirRule[] = [];

/**
 * WriteFile rule used by the mockWriteFile function
 * Defines the behavior when a path matches a specific pattern
 */
interface WriteFileRule {
  /** Pattern to match against paths */
  pattern: string | RegExp;
  /** Whether writeFile should succeed (true) or fail (false or Error) */
  result: boolean | Error;
}

/** Registry of path-specific writeFile rules */
const writeFileRules: WriteFileRule[] = [];

/**
 * Resets all fs mock functions to their initial state.
 * 
 * This function should be called in the `beforeEach` hook of your tests to prevent
 * test cross-contamination. It clears all mock implementations, mock calls history,
 * and path-specific rules that may have been configured.
 * 
 * @example
 * ```typescript
 * beforeEach(() => {
 *   resetMockFs();
 *   setupMockFs();
 * });
 * ```
 */
export function resetMockFs(): void {
  jest.clearAllMocks();
  
  // Reset specific behavior mocks
  mockedFs.access.mockReset();
  mockedFs.readFile.mockReset();
  mockedFs.writeFile.mockReset();
  mockedFs.stat.mockReset();
  mockedFs.readdir.mockReset();
  mockedFs.mkdir.mockReset();
  
  // Clear all path-specific configurations
  accessRules.length = 0;
  readFileRules.length = 0;
  statRules.length = 0;
  readdirRules.length = 0;
  mkdirRules.length = 0;
  writeFileRules.length = 0;
}

/**
 * Configures the mocked fs module with default behaviors.
 * 
 * This function sets up mock implementations for all fs/promises functions,
 * applying the provided configuration or default values. It should be called
 * after `resetMockFs()` to establish baseline behavior for all fs operations.
 * 
 * @param config - Optional configuration to customize the default behaviors
 * 
 * @example
 * ```typescript
 * // Setup with defaults (files are accessible, empty content)
 * setupMockFs();
 * 
 * // Setup with custom defaults
 * setupMockFs({
 *   defaultFileContent: 'Default file content',
 *   defaultAccessBehavior: true,
 *   defaultStats: {
 *     isFile: () => true,
 *     isDirectory: () => false,
 *     size: 1024
 *   }
 * });
 * ```
 */
export function setupMockFs(config?: FsMockConfig): void {
  // Merge provided config with defaults
  const mergedConfig = { ...DEFAULT_CONFIG, ...config };
  
  // Configure fs.access with path-specific behavior support
  mockedFs.access.mockImplementation((path, _mode) => {
    // Convert path to string for comparison (it could be URL or Buffer too)
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of accessRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        if (rule.allowed) {
          return Promise.resolve(undefined);
        } else {
          // Format message according to specific error code
          const errorMessage = rule.errorMessage || 'File not found or access denied';
          
          // Create appropriate error based on code
          const error = createFsError(
            rule.errorCode || mergedConfig.defaultAccessErrorCode || 'ENOENT',
            errorMessage,
            'access',
            pathStr
          );
          return Promise.reject(error);
        }
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched
    if (mergedConfig.defaultAccessBehavior) {
      // Access allowed - resolve successfully
      return Promise.resolve(undefined);
    } else {
      // Access denied - reject with error
      const error = createFsError(
        mergedConfig.defaultAccessErrorCode || 'ENOENT',
        'File not found or access denied',
        'access',
        pathStr
      );
      return Promise.reject(error);
    }
  });
  
  // Configure fs.readFile with path-specific behavior support
  mockedFs.readFile.mockImplementation((path, options?) => {
    // Convert path to string for comparison (it could be URL, Buffer, or FileHandle)
    // If path is a FileHandle, we won't try to match it with our rules
    if (typeof path !== 'string' && !(path instanceof Buffer) && !('href' in path)) {
      return Promise.resolve(mergedConfig.defaultFileContent as any);
    }
    
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of readFileRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        // If the rule specifies an error, reject with it
        if (rule.content instanceof Error) {
          return Promise.reject(rule.content);
        }
        
        const content = rule.content;
        
        // Handle encoding option for string content
        if (typeof content === 'string') {
          // Special handling for raw/normalized content
          // Convert line endings for Windows compatibility in tests
          const normalizedContent = typeof options === 'string' && options === 'utf-8' 
            ? content.replace(/\r\n/g, '\n')
            : content;
            
          // If no encoding is specified, return a Buffer
          if (!options || (typeof options === 'object' && !options.encoding)) {
            return Promise.resolve(Buffer.from(normalizedContent));
          }
          // Otherwise return the string directly
          return Promise.resolve(normalizedContent);
        }
        
        // For Buffer content, return it directly
        return Promise.resolve(content);
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched
    const defaultContent = mergedConfig.defaultFileContent || '';
    
    // Handle encoding option for default content
    if (typeof defaultContent === 'string') {
      if (!options || (typeof options === 'object' && !options.encoding)) {
        return Promise.resolve(Buffer.from(defaultContent));
      }
    }
    
    return Promise.resolve(defaultContent as any);
  });
  
  // Configure fs.writeFile with path-specific behavior support
  mockedFs.writeFile.mockImplementation((path, _data, _options) => {
    // Convert path to string for comparison (it could be URL, Buffer, or FileHandle)
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of writeFileRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        // If the rule specifies an error, reject with it
        if (rule.result instanceof Error) {
          return Promise.reject(rule.result);
        }
        
        // If the rule is boolean, it determines success (true) or failure (false)
        if (rule.result === true) {
          return Promise.resolve(undefined);
        } else {
          // Default error for failure case
          const error = createFsError(
            'EACCES', 
            'Permission denied',
            'writeFile',
            pathStr
          );
          return Promise.reject(error);
        }
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched (success)
    return Promise.resolve(undefined);
  });
  
  // Configure fs.stat with path-specific behavior support
  mockedFs.stat.mockImplementation((path) => {
    // Convert path to string for comparison (it could be URL, Buffer, or FileHandle)
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of statRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        // If the rule specifies an error, reject with it
        if (rule.result instanceof Error) {
          return Promise.reject(rule.result);
        }
        
        // Otherwise, create and return stats from the provided MockedStats
        const stats = createStats(rule.result);
        // Cast to any to avoid TypeScript issues with BigIntStats vs Stats
        return Promise.resolve(stats as any);
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched
    const stats = createStats(mergedConfig.defaultStats || {});
    // Cast to any to avoid TypeScript issues with BigIntStats vs Stats
    return Promise.resolve(stats as any);
  });
  
  // Configure fs.readdir with path-specific behavior support
  mockedFs.readdir.mockImplementation((path) => {
    // Convert path to string for comparison (it could be URL, Buffer, or FileHandle)
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of readdirRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        // If the rule specifies an error, reject with it
        if (rule.result instanceof Error) {
          return Promise.reject(rule.result);
        }
        
        // Otherwise, return the list of directory entries
        // Cast to any to avoid TypeScript issues with different return types
        return Promise.resolve(rule.result as any);
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched (empty directory)
    // Cast to any to avoid TypeScript issues with different return types
    return Promise.resolve([] as any);
  });
  
  // Configure fs.mkdir with path-specific behavior support
  mockedFs.mkdir.mockImplementation((path) => {
    // Convert path to string for comparison (it could be URL, Buffer, or FileHandle)
    const pathStr = String(path);
    
    // Check if we have any path-specific rules
    for (const rule of mkdirRules) {
      const matches = 
        (typeof rule.pattern === 'string' && rule.pattern === pathStr) || 
        (rule.pattern instanceof RegExp && rule.pattern.test(pathStr));
      
      if (matches) {
        // If the rule specifies an error, reject with it
        if (rule.result instanceof Error) {
          return Promise.reject(rule.result);
        }
        
        // If the rule is boolean, it determines success (true) or failure (false)
        if (rule.result === true) {
          // Cast to any to handle different return types based on recursive option
          return Promise.resolve(undefined as any);
        } else {
          // Default error for failure case
          const error = createFsError(
            'EACCES', 
            'Permission denied',
            'mkdir',
            pathStr
          );
          return Promise.reject(error);
        }
      }
    }
    
    // Fall back to default behavior if no path-specific rule matched (success)
    // Cast to any to handle different return types based on recursive option
    return Promise.resolve(undefined as any);
  });
}

/**
 * Interface for a mocked filesystem error
 * 
 * NOTE: This interface is deprecated and will be removed.
 * Use NodeJS.ErrnoException directly instead, which properly
 * represents the structure of filesystem errors.
 * 
 * @deprecated Use NodeJS.ErrnoException instead
 */
export interface MockedFsError {
  /**
   * Error code (e.g., 'ENOENT', 'EACCES', etc.)
   */
  code: string;

  /**
   * Human-readable error message
   */
  message: string;

  /**
   * Whether this is a system error (true for fs errors)
   */
  syscall?: string;

  /**
   * The path that caused the error
   */
  path?: string;
}

/**
 * Interface representing partial file stats
 * Mirrors the important properties of Node.js fs.Stats class
 * but allows for easier creation in tests
 */
export interface MockedStats {
  /**
   * Whether the path is a file
   */
  isFile?: () => boolean;

  /**
   * Whether the path is a directory
   */
  isDirectory?: () => boolean;

  /**
   * Whether the path is a symbolic link
   */
  isSymbolicLink?: () => boolean;

  /**
   * File size in bytes
   */
  size?: number;

  /**
   * File creation time
   */
  birthtime?: Date;

  /**
   * Last modification time
   */
  mtime?: Date;

  /**
   * Last access time
   */
  atime?: Date;
}

/**
 * Configuration options for the mock filesystem
 */
export interface FsMockConfig {
  /**
   * Default behavior for fs.access (true = allow access, false = deny access)
   */
  defaultAccessBehavior?: boolean;

  /**
   * Default content returned by fs.readFile if no specific mock is set
   */
  defaultFileContent?: string;

  /**
   * Default stats returned by fs.stat if no specific mock is set
   */
  defaultStats?: MockedStats;

  /**
   * Default error code for access errors
   */
  defaultAccessErrorCode?: string;

  /**
   * Default error code for read errors
   */
  defaultReadErrorCode?: string;

  /**
   * Default error code for write errors
   */
  defaultWriteErrorCode?: string;

  /**
   * Default error code for stat errors
   */
  defaultStatErrorCode?: string;
}

/**
 * Interface for the mockAccess function
 */
export interface MockAccessFunction {
  /**
   * Configures fs.access to resolve or reject for specific paths
   * @param pathPattern - Path or regex pattern to match
   * @param allowed - Whether access should be allowed (true) or denied (false)
   * @param options - Optional error details if denied
   */
  (
    pathPattern: string | RegExp,
    allowed: boolean,
    options?: {
      errorCode?: string;
      errorMessage?: string;
    }
  ): void;
}

/**
 * Interface for the mockReadFile function
 */
export interface MockReadFileFunction {
  /**
   * Configures fs.readFile to return content or throw an error for specific paths
   * @param pathPattern - Path or regex pattern to match
   * @param content - Content to return or Error to throw
   */
  (pathPattern: string | RegExp, content: string | Error | Buffer): void;
}

/**
 * Interface for the mockStat function
 */
export interface MockStatFunction {
  /**
   * Configures fs.stat to return stats or throw an error for specific paths
   * @param pathPattern - Path or regex pattern to match
   * @param stats - Stats object to return or Error to throw
   */
  (pathPattern: string | RegExp, stats: MockedStats | Error): void;
}

/**
 * Interface for the mockReaddir function
 */
export interface MockReaddirFunction {
  /**
   * Configures fs.readdir to return entries or throw an error for specific directories
   * @param pathPattern - Path or regex pattern to match
   * @param entries - Directory entries to return or Error to throw
   */
  (pathPattern: string | RegExp, entries: string[] | Error): void;
}

/**
 * Interface for the mockMkdir function
 */
export interface MockMkdirFunction {
  /**
   * Configures fs.mkdir to succeed or fail for specific paths
   * @param pathPattern - Path or regex pattern to match
   * @param success - Whether mkdir should succeed (true) or fail (false or Error)
   */
  (pathPattern: string | RegExp, success: boolean | Error): void;
}

/**
 * Interface for the mockWriteFile function
 */
export interface MockWriteFileFunction {
  /**
   * Configures fs.writeFile to succeed or fail for specific paths
   * @param pathPattern - Path or regex pattern to match
   * @param success - Whether writeFile should succeed (true) or fail (false or Error)
   */
  (pathPattern: string | RegExp, success: boolean | Error): void;
}

/**
 * Creates a Node.js-like filesystem error.
 * 
 * This utility function creates an Error object with the properties expected
 * from Node.js filesystem errors, making it suitable for use in mocking
 * filesystem operation failures.
 * 
 * @param code - Error code (e.g., 'ENOENT', 'EACCES', 'EPERM', 'EROFS')
 * @param message - Error message to include in the error
 * @param syscall - System call that failed (e.g., 'access', 'readFile', 'writeFile')
 * @param filepath - Path that caused the error
 * @returns Error with proper fs error properties
 * 
 * @example
 * ```typescript
 * // Create a file not found error
 * const notFoundError = createFsError(
 *   'ENOENT',
 *   'File not found',
 *   'readFile',
 *   '/path/to/missing.txt'
 * );
 * 
 * // Create a permission denied error
 * const permissionError = createFsError(
 *   'EACCES',
 *   'Permission denied',
 *   'access',
 *   '/path/to/protected.txt'
 * );
 * 
 * // Use the error in a mock
 * mockReadFile('/path/to/error.txt', notFoundError);
 * ```
 */
export function createFsError(
  code: string,
  message: string,
  syscall: string,
  filepath: string
): NodeJS.ErrnoException {
  // First create a basic error object
  const error = new Error() as NodeJS.ErrnoException;
  
  // Set standard properties
  error.code = code;
  error.syscall = syscall;
  error.path = filepath;
  error.message = message; // Use the message directly as provided
  
  // Map common error codes to errno numbers for better compatibility
  if (code === 'ENOENT') error.errno = -2;
  else if (code === 'EACCES') error.errno = -13;
  else if (code === 'EPERM') error.errno = -1;
  else if (code === 'EROFS') error.errno = -30;
  else if (code === 'EBUSY') error.errno = -16;
  else if (code === 'EMFILE') error.errno = -24;
  
  return error;
}

/**
 * Configures fs.access to resolve or reject for specific paths.
 * 
 * This function allows you to mock the behavior of fs.access for specific paths
 * or path patterns, controlling whether access is allowed or denied, and what
 * error should be returned when denied.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param allowed - Whether access should be allowed (true) or denied (false)
 * @param options - Optional error details if denied
 * 
 * @example
 * ```typescript
 * // Allow access to a specific file
 * mockAccess('/path/to/file.txt', true);
 * 
 * // Deny access to a specific file (file not found)
 * mockAccess('/path/to/missing.txt', false, {
 *   errorCode: 'ENOENT',
 *   errorMessage: 'File not found'
 * });
 * 
 * // Deny access to a specific file (permission denied)
 * mockAccess('/path/to/protected.txt', false, {
 *   errorCode: 'EACCES',
 *   errorMessage: 'Permission denied'
 * });
 * 
 * // Use regex pattern to match multiple paths
 * mockAccess(/\.log$/, false); // Deny access to all .log files
 * ```
 */
export const mockAccess: MockAccessFunction = (
  pathPattern: string | RegExp,
  allowed: boolean,
  options?: {
    errorCode?: string;
    errorMessage?: string;
  }
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = accessRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    accessRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  accessRules.unshift({
    pattern: pathPattern,
    allowed,
    errorCode: options?.errorCode,
    errorMessage: options?.errorMessage
  });
};

/**
 * Configures fs.readFile to return content or throw an error for specific paths.
 * 
 * This function allows you to mock the behavior of fs.readFile for specific paths
 * or path patterns, controlling what content should be returned or what error
 * should be thrown.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param content - Content to return (string or Buffer) or Error to throw
 * 
 * @example
 * ```typescript
 * // Return text content for a specific file
 * mockReadFile('/path/to/file.txt', 'File content');
 * 
 * // Return binary content for a specific file
 * mockReadFile('/path/to/binary.bin', Buffer.from([0x00, 0xFF, 0x42]));
 * 
 * // Simulate a file read error
 * mockReadFile('/path/to/error.txt', createFsError(
 *   'ENOENT',
 *   'File not found',
 *   'readFile',
 *   '/path/to/error.txt'
 * ));
 * 
 * // Use regex pattern to match multiple files
 * mockReadFile(/\.json$/, '{"key": "value"}');
 * ```
 */
export const mockReadFile: MockReadFileFunction = (
  pathPattern: string | RegExp,
  content: string | Buffer | Error
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = readFileRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    readFileRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  readFileRules.unshift({
    pattern: pathPattern,
    content
  });
};

/**
 * Configures fs.stat to return stats or throw an error for specific paths.
 * 
 * This function allows you to mock the behavior of fs.stat for specific paths
 * or path patterns, controlling what stats should be returned or what error
 * should be thrown.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param statsOrError - Stats object to return or Error to throw
 * 
 * @example
 * ```typescript
 * // Configure a regular file
 * mockStat('/path/to/file.txt', {
 *   isFile: () => true,
 *   isDirectory: () => false,
 *   size: 1024
 * });
 * 
 * // Configure a directory
 * mockStat('/path/to/dir', {
 *   isFile: () => false,
 *   isDirectory: () => true,
 *   size: 4096
 * });
 * 
 * // Configure a stat error (file not found)
 * mockStat('/path/to/missing.txt', createFsError(
 *   'ENOENT',
 *   'No such file or directory',
 *   'stat',
 *   '/path/to/missing.txt'
 * ));
 * 
 * // Use regex to match multiple paths
 * mockStat(/\.jpg$/, {
 *   isFile: () => true,
 *   size: 1024 * 1024 // 1MB
 * });
 * ```
 */
export const mockStat: MockStatFunction = (
  pathPattern: string | RegExp,
  statsOrError: MockedStats | Error
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = statRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    statRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  statRules.unshift({
    pattern: pathPattern,
    result: statsOrError
  });
};

/**
 * Configures fs.readdir to return entries or throw an error for specific directories.
 * 
 * This function allows you to mock the behavior of fs.readdir for specific paths
 * or path patterns, controlling what directory entries should be returned or what
 * error should be thrown.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param entries - Directory entries to return or Error to throw
 * 
 * @example
 * ```typescript
 * // Return a list of files for a specific directory
 * mockReaddir('/path/to/dir', ['file1.txt', 'file2.js', 'subdir']);
 * 
 * // Return an empty directory
 * mockReaddir('/path/to/empty', []);
 * 
 * // Simulate a directory read error
 * mockReaddir('/path/to/error', createFsError(
 *   'ENOENT',
 *   'Directory not found',
 *   'readdir',
 *   '/path/to/error'
 * ));
 * 
 * // Use regex pattern to match multiple directories
 * mockReaddir(/\/config\/.*$/, ['settings.json', 'environment.json']);
 * ```
 */
export const mockReaddir: MockReaddirFunction = (
  pathPattern: string | RegExp,
  entries: string[] | Error
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = readdirRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    readdirRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  readdirRules.unshift({
    pattern: pathPattern,
    result: entries
  });
};

/**
 * Configures fs.mkdir to succeed or fail for specific paths.
 * 
 * This function allows you to mock the behavior of fs.mkdir for specific paths
 * or path patterns, controlling whether directory creation should succeed or
 * fail, and what error should be thrown in case of failure.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param success - Whether mkdir should succeed (true) or fail (false or Error)
 * 
 * @example
 * ```typescript
 * // Allow directory creation for a specific path
 * mockMkdir('/path/to/new/dir', true);
 * 
 * // Deny directory creation with a default error
 * mockMkdir('/path/to/fail/dir', false);
 * 
 * // Deny directory creation with a specific error
 * mockMkdir('/path/to/protected/dir', createFsError(
 *   'EACCES',
 *   'Permission denied',
 *   'mkdir',
 *   '/path/to/protected/dir'
 * ));
 * 
 * // Use regex pattern to match multiple paths
 * mockMkdir(/\/readonly\/.*$/, createFsError(
 *   'EROFS',
 *   'Read-only file system',
 *   'mkdir',
 *   '/readonly/path'
 * ));
 * ```
 */
export const mockMkdir: MockMkdirFunction = (
  pathPattern: string | RegExp,
  success: boolean | Error
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = mkdirRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    mkdirRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  mkdirRules.unshift({
    pattern: pathPattern,
    result: success
  });
};

/**
 * Configures fs.writeFile to succeed or fail for specific paths.
 * 
 * This function allows you to mock the behavior of fs.writeFile for specific paths
 * or path patterns, controlling whether file writing should succeed or fail, and
 * what error should be thrown in case of failure.
 * 
 * @param pathPattern - Path or regex pattern to match
 * @param success - Whether writeFile should succeed (true) or fail (false or Error)
 * 
 * @example
 * ```typescript
 * // Allow writing to a specific file
 * mockWriteFile('/path/to/writeable.txt', true);
 * 
 * // Deny writing with a default error (permission denied)
 * mockWriteFile('/path/to/readonly.txt', false);
 * 
 * // Deny writing with a specific error
 * mockWriteFile('/path/to/protected.txt', createFsError(
 *   'EACCES',
 *   'Permission denied',
 *   'writeFile',
 *   '/path/to/protected.txt'
 * ));
 * 
 * // Use regex pattern to deny writing to multiple files
 * mockWriteFile(/\.log$/, createFsError(
 *   'EPERM',
 *   'Operation not permitted',
 *   'writeFile',
 *   'log file'
 * ));
 * ```
 */
export const mockWriteFile: MockWriteFileFunction = (
  pathPattern: string | RegExp,
  success: boolean | Error
): void => {
  // Find and remove any existing rule with the same pattern
  const existingIndex = writeFileRules.findIndex(
    rule => 
      (typeof rule.pattern === 'string' && 
       typeof pathPattern === 'string' && 
       rule.pattern === pathPattern) ||
      (rule.pattern instanceof RegExp && 
       pathPattern instanceof RegExp && 
       rule.pattern.source === pathPattern.source)
  );
  
  if (existingIndex !== -1) {
    writeFileRules.splice(existingIndex, 1);
  }
  
  // Add new rule at the beginning for higher precedence
  writeFileRules.unshift({
    pattern: pathPattern,
    result: success
  });
};

/**
 * Creates a Stats-like object from partial stats.
 * 
 * This utility function creates a fully fleshed-out fs.Stats object from a
 * partial MockedStats object, filling in default values for any missing properties.
 * This is used internally by the mockStat function but can also be used directly
 * if needed.
 * 
 * @param stats - Partial stats object with properties to include
 * @returns A full Stats-like object with all required methods
 * 
 * @example
 * ```typescript
 * // Create a full Stats object for a file
 * const fileStats = createStats({
 *   isFile: () => true,
 *   isDirectory: () => false,
 *   size: 1024,
 *   mtime: new Date('2023-01-01')
 * });
 * 
 * // Create a full Stats object for a directory
 * const dirStats = createStats({
 *   isFile: () => false,
 *   isDirectory: () => true,
 *   size: 4096
 * });
 * ```
 */
export function createStats(stats: MockedStats): Stats {
  // Create a base Stats object with default values
  const baseStats: Stats = {
    isFile: () => false,
    isDirectory: () => false,
    isBlockDevice: () => false,
    isCharacterDevice: () => false,
    isSymbolicLink: () => false,
    isFIFO: () => false,
    isSocket: () => false,
    dev: 0,
    ino: 0,
    mode: 0,
    nlink: 1,
    uid: 0,
    gid: 0,
    rdev: 0,
    size: 0,
    blksize: 4096,
    blocks: 0,
    atimeMs: 0,
    mtimeMs: 0,
    ctimeMs: 0,
    birthtimeMs: 0,
    atime: new Date(),
    mtime: new Date(),
    ctime: new Date(),
    birthtime: new Date(),
  } as unknown as Stats;
  
  // Override with provided values
  return Object.assign(baseStats, stats);
}