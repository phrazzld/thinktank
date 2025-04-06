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
  (pathPattern: string | RegExp, content: string | Error): void;
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