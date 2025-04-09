/**
 * File system setup utilities for tests
 * 
 * This module provides specialized setup helpers for file system related tests,
 * building on the virtual filesystem (memfs) approach.
 */
import path from 'path';
import fs from 'fs/promises';
import { Stats } from 'fs';
import { 
  resetVirtualFs, 
  createVirtualFs, 
  getVirtualFs,
  createFsError as createFsErrorUtil,
  normalizePathForMemfs
} from '../../src/__tests__/utils/virtualFsUtils';
import { normalizePathGeneral } from '../../src/utils/pathUtils';
import { FileSystem } from '../../src/core/interfaces';

/**
 * Sets up the virtual filesystem with a given file structure.
 * Resets the filesystem before applying the structure by default.
 * 
 * @param structure - Object mapping file paths to their content. Paths will be normalized.
 *                   Use paths ending in '/' or empty string content ('') to explicitly create directories.
 * @param options - Optional settings. `reset: true` (default) clears the FS first.
 * 
 * @example
 * setupWithFiles({
 *   '/path/to/file.txt': 'File content',
 *   '/path/to/dir/file.js': 'console.log("Hello");',
 *   '/empty-dir/': '' // Creates an empty directory
 * });
 */
export function setupWithFiles(
  structure: Record<string, string>,
  options: { reset?: boolean } = { reset: true }
): void {
  if (options.reset !== false) {
    resetVirtualFs();
  }
  // createVirtualFs handles normalization and directory creation
  createVirtualFs(structure, { reset: false });
}

// Keep setupBasicFs as an alias for backward compatibility
export const setupBasicFs = setupWithFiles;

/**
 * Sets up the virtual filesystem with a single file.
 * Resets the filesystem before creating the file.
 * 
 * @param filePath - Path to the file (will be normalized).
 * @param content - Content of the file.
 * 
 * @example
 * setupWithSingleFile('/config/settings.json', '{ "port": 3000 }');
 */
export function setupWithSingleFile(filePath: string, content: string): void {
  resetVirtualFs();
  // createVirtualFs handles normalization and directory creation
  createVirtualFs({ [filePath]: content }, { reset: false });
}

/**
 * Sets up the virtual filesystem with a nested directory structure.
 * This is a wrapper around setupWithFiles that emphasizes its ability to handle nested paths.
 * Resets the filesystem before applying the structure.
 * 
 * @param structure - Object mapping file/directory paths to string content.
 *                   Use paths ending in '/' or empty string content ('') to explicitly create directories.
 * 
 * @example
 * setupWithNestedDirectories({
 *   '/deep/nested/dir/file.txt': 'content',
 *   '/deep/nested/empty/': '', // Creates empty directory
 *   '/another/path/with/file.js': 'console.log("nested");'
 * });
 */
export function setupWithNestedDirectories(
  structure: Record<string, string>
): void {
  setupWithFiles(structure, { reset: true });
}

/**
 * Sets up a standard project structure with customizable files
 * 
 * @param basePath - The root path where the project will be created
 * @param files - An object mapping relative file paths to their content
 * 
 * Usage:
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
  setupBasicFs(structure);
}

/**
 * Gets direct access to the virtual filesystem for advanced operations
 * 
 * @returns The virtual filesystem instance
 * 
 * Usage:
 * ```typescript
 * const vfs = getFs();
 * vfs.writeFileSync('/path/to/file.txt', 'content');
 * ```
 */
export function getFs(): ReturnType<typeof getVirtualFs> {
  return getVirtualFs();
}

/**
 * Creates a standardized filesystem error for testing error handling
 * 
 * @param code - Error code (e.g., 'ENOENT', 'EACCES')
 * @param message - Error message
 * @param syscall - System call that failed
 * @param filepath - Path that caused the error
 * @returns A properly formatted filesystem error
 * 
 * Usage:
 * ```typescript
 * const error = createFsError('ENOENT', 'File not found', 'open', '/missing.txt');
 * ```
 */
export function createFsError(
  code: string,
  message: string,
  syscall: string,
  filepath: string
): NodeJS.ErrnoException {
  return createFsErrorUtil(code, message, syscall, filepath);
}

/**
 * Sets up a directory structure for testing file reading operations
 * 
 * @param baseDir - Base directory path
 * @param files - Optional specific files to create
 * @returns Object with paths to created files
 * 
 * Usage:
 * ```typescript
 * const { testFile, configFile } = setupFileReaderTest('/project');
 * ```
 */
export function setupFileReaderTest(
  baseDir: string = '/test',
  files?: Record<string, string>
): { testFile: string; configFile: string; nestedFile: string } {
  const testFile = path.join(baseDir, 'test.txt');
  const configFile = path.join(baseDir, 'config.json');
  const nestedFile = path.join(baseDir, 'nested', 'file.txt');
  
  const defaultFiles = {
    [testFile]: 'This is a test file',
    [configFile]: JSON.stringify({ test: true }),
    [nestedFile]: 'Nested file content'
  };
  
  setupBasicFs({
    ...defaultFiles,
    ...(files || {})
  });
  
  return {
    testFile,
    configFile,
    nestedFile
  };
}

/**
 * Mocks the fileExists function to work with the virtual filesystem
 * 
 * @param mockFn - The jest mock function for fileExists
 * 
 * Usage:
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
 * Creates a mock FileSystem implementation backed by memfs.
 * This provides a Jest-mocked object that implements the FileSystem interface
 * while delegating actual operations to the virtual filesystem.
 * 
 * @returns A Jest-mocked FileSystem object
 * 
 * Usage:
 * ```typescript
 * const mockFileSystem = createMockFileSystem();
 * // Configure specific behaviors if needed:
 * mockFileSystem.readFileContent.mockResolvedValueOnce('Custom content');
 * ```
 */
export function createMockFileSystem(): jest.Mocked<FileSystem> {
  const vfs = getVirtualFs(); // Get the memfs instance

  // Create mock methods that delegate to vfs
  const mockMethods = {
    readFileContent: jest.fn().mockImplementation((filePath: string, _options?: { normalize?: boolean }) => {
      try {
        const normPath = normalizePathForMemfs(filePath);
        return Promise.resolve(vfs.readFileSync(normPath, 'utf8') as string);
      } catch (e) {
        return Promise.reject(createFsErrorUtil('ENOENT', `File not found: ${filePath}`, 'readFile', filePath));
      }
    }),

    writeFile: jest.fn().mockImplementation((filePath: string, content: string) => {
      try {
        const normPath = normalizePathForMemfs(filePath);
        const dir = path.dirname(normPath);
        if (dir) {
          vfs.mkdirSync(dir, { recursive: true });
        }
        vfs.writeFileSync(normPath, content);
        return Promise.resolve();
      } catch (e) {
        const nodeError = e as NodeJS.ErrnoException;
        if (nodeError.code === 'EACCES') {
          return Promise.reject(createFsErrorUtil('EACCES', `Permission denied: ${filePath}`, 'writeFile', filePath));
        }
        return Promise.reject(createFsErrorUtil('UNKNOWN', `Failed to write file: ${filePath}`, 'writeFile', filePath));
      }
    }),

    fileExists: jest.fn().mockImplementation((filePath: string) => {
      const normPath = normalizePathForMemfs(filePath);
      return Promise.resolve(vfs.existsSync(normPath));
    }),

    mkdir: jest.fn().mockImplementation((dirPath: string, options?: { recursive?: boolean }) => {
      try {
        const normPath = normalizePathForMemfs(dirPath);
        vfs.mkdirSync(normPath, options);
        return Promise.resolve();
      } catch (e) {
        const nodeError = e as NodeJS.ErrnoException;
        if (nodeError.code === 'EACCES') {
          return Promise.reject(createFsErrorUtil('EACCES', `Permission denied: ${dirPath}`, 'mkdir', dirPath));
        }
        return Promise.reject(createFsErrorUtil('UNKNOWN', `Failed to create directory: ${dirPath}`, 'mkdir', dirPath));
      }
    }),

    readdir: jest.fn().mockImplementation((dirPath: string) => {
      try {
        const normPath = normalizePathForMemfs(dirPath);
        return Promise.resolve(vfs.readdirSync(normPath) as string[]);
      } catch (e) {
        return Promise.reject(createFsErrorUtil('ENOENT', `Directory not found: ${dirPath}`, 'readdir', dirPath));
      }
    }),

    stat: jest.fn().mockImplementation((filePath: string) => {
      try {
        const normPath = normalizePathForMemfs(filePath);
        return Promise.resolve(vfs.statSync(normPath) as Stats);
      } catch (e) {
        return Promise.reject(createFsErrorUtil('ENOENT', `Path not found: ${filePath}`, 'stat', filePath));
      }
    }),

    access: jest.fn().mockImplementation((filePath: string, _mode?: number) => {
      try {
        const normPath = normalizePathForMemfs(filePath);
        // memfs doesn't fully implement access, use statSync as a proxy
        vfs.statSync(normPath);
        return Promise.resolve();
      } catch (e) {
        const nodeError = e as NodeJS.ErrnoException;
        if (nodeError.code === 'ENOENT') {
          return Promise.reject(createFsErrorUtil('ENOENT', `Path not found: ${filePath}`, 'access', filePath));
        }
        return Promise.reject(createFsErrorUtil('EACCES', `Permission denied: ${filePath}`, 'access', filePath));
      }
    }),

    // Mock config path methods
    getConfigDir: jest.fn().mockResolvedValue('/mock/.config/thinktank'),
    getConfigFilePath: jest.fn().mockResolvedValue('/mock/.config/thinktank/config.json'),
  };

  // Cast to FileSystem with jest.Mocked type
  return mockMethods as unknown as jest.Mocked<FileSystem>;
}
