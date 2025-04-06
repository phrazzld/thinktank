/**
 * Mock utilities for the Node.js fs/promises module in tests
 * Provides a consistent interface for mocking filesystem operations
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
 * Resets all fs mock functions to their initial state
 * This should be called before each test to prevent test pollution
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
}

/**
 * Configures the mocked fs module with default behaviors
 * @param config - Optional configuration to customize the default behaviors
 */
export function setupMockFs(config?: FsMockConfig): void {
  // Merge provided config with defaults
  const mergedConfig = { ...DEFAULT_CONFIG, ...config };
  
  // Configure fs.access with path-specific behavior support
  mockedFs.access.mockImplementation((path) => {
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
          const error = createFsError(
            rule.errorCode || mergedConfig.defaultAccessErrorCode || 'ENOENT',
            rule.errorMessage || 'File not found or access denied',
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
          // If no encoding is specified, return a Buffer
          if (!options || (typeof options === 'object' && !options.encoding)) {
            return Promise.resolve(Buffer.from(content));
          }
          // Otherwise return the string directly
          return Promise.resolve(content);
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
  
  // Configure fs.writeFile
  mockedFs.writeFile.mockResolvedValue(undefined);
  
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
  
  // Configure fs.mkdir
  mockedFs.mkdir.mockResolvedValue(undefined);
}

/**
 * Interface for a mocked filesystem error
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
 * Creates a Node.js-like filesystem error
 * @param code - Error code (e.g., 'ENOENT', 'EACCES')
 * @param message - Error message
 * @param syscall - System call that failed
 * @param filepath - Path that caused the error
 * @returns Error with proper fs error properties
 */
export function createFsError(
  code: string,
  message: string,
  syscall: string,
  filepath: string
): Error & MockedFsError {
  const error = new Error(message) as Error & MockedFsError;
  error.code = code;
  error.syscall = syscall;
  error.path = filepath;
  return error;
}

/**
 * Configures fs.access to resolve or reject for specific paths
 * @param pathPattern - Path or regex pattern to match
 * @param allowed - Whether access should be allowed (true) or denied (false)
 * @param options - Optional error details if denied
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
 * Configures fs.readFile to return content or throw an error for specific paths
 * @param pathPattern - Path or regex pattern to match
 * @param content - Content to return or Error to throw
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
 * Configures fs.stat to return stats or throw an error for specific paths
 * @param pathPattern - Path or regex pattern to match
 * @param statsOrError - Stats object to return or Error to throw
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
 * Configures fs.readdir to return entries or throw an error for specific directories
 * @param pathPattern - Path or regex pattern to match
 * @param entries - Directory entries to return or Error to throw
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
 * Creates a Stats-like object from partial stats
 * @param stats - Partial stats object with properties to include
 * @returns A full Stats-like object with all required methods
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