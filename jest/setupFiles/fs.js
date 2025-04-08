/**
 * Filesystem mock setup for Jest
 * 
 * This file sets up the basic mocking for filesystem modules (fs, fs/promises)
 * using the virtualFsUtils implementation for a standardized approach to testing
 * filesystem operations.
 * 
 * PREFERRED APPROACH: Import these helpers directly rather than using manual mocks or
 * mockFactories.ts to ensure consistent test behavior.
 * 
 * PATH HANDLING: Always use normalizePath() for all paths used with the virtual filesystem.
 * This ensures paths are consistently normalized across operating systems. When setting up
 * test files or asserting paths, always normalize both the expected and actual paths.
 */

// Import the mock utilities
const { 
  mockFsModules, 
  resetVirtualFs, 
  createVirtualFs, 
  createFsError: createFsErrorUtil,
  getVirtualFs,
  normalizePathForMemfs,
  createMockStats,
  createMockDirent
} = require('../../src/__tests__/utils/virtualFsUtils');

// Mock the fs module
jest.mock('fs', () => mockFsModules().fs);

// Mock the fs/promises module
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Export helper functions for convenience (available by importing from './jest/setupFiles/fs')
module.exports = {
  /**
   * Sets up a basic filesystem structure for tests
   * 
   * @param {Object} structure - Object mapping file paths to content
   * @param {Object} options - Additional options (reset: boolean)
   * 
   * @example
   * setupBasicFs({
   *   '/path/to/file.txt': 'File content',
   *   '/path/to/dir/file.js': 'console.log("Hello");'
   * });
   * 
   * NOTE: All paths in the structure will be automatically normalized to ensure consistency 
   * across operating systems. It's still recommended to use the normalizePath() helper 
   * for path keys in the structure to make the normalization explicit.
   */
  setupBasicFs: function(structure = {}, options = { reset: true }) {
    if (options.reset) {
      resetVirtualFs();
    }
    createVirtualFs(structure, { reset: false });
  },

  /**
   * Resets the virtual filesystem, removing all files and directories.
   * 
   * @example
   * resetFs();
   */
  resetFs: function() {
    resetVirtualFs();
  },

  /**
   * Creates a standardized filesystem error
   * 
   * @param {string} code - Error code (e.g., 'ENOENT', 'EACCES')
   * @param {string} message - Error message
   * @param {string} syscall - System call that failed
   * @param {string} filepath - Path that caused the error
   * @returns {NodeJS.ErrnoException} A properly formatted filesystem error
   * 
   * @example
   * const error = createFsError('ENOENT', 'File not found', 'open', '/missing.txt');
   */
  createFsError: function(code, message, syscall, filepath) {
    return createFsErrorUtil(code, message, syscall, filepath);
  },

  /**
   * Gets direct access to the virtual filesystem for advanced operations
   * 
   * @returns {Object} The virtual filesystem instance
   * 
   * @example
   * const vfs = getFs();
   * vfs.writeFileSync('/path/to/file.txt', 'content');
   */
  getFs: function() {
    return getVirtualFs();
  },

  /**
   * Creates a mock fs.Stats object for testing
   * 
   * @param {boolean} isFile - Whether this represents a file (true) or directory (false)
   * @param {number} size - Size in bytes
   * @returns {Object} A mock Stats object
   * 
   * @example
   * const fileStats = createStats(true, 1024);
   * const dirStats = createStats(false);
   */
  createStats: function(isFile, size) {
    return createMockStats(isFile, size);
  },

  /**
   * Creates a mock fs.Dirent object for testing directory entries
   * 
   * @param {string} name - Name of the file or directory
   * @param {boolean} isFile - Whether this represents a file (true) or directory (false)
   * @returns {Object} A mock Dirent object
   * 
   * @example
   * const fileDirent = createDirent('file.txt', true);
   * const dirDirent = createDirent('folder', false);
   */
  createDirent: function(name, isFile) {
    return createMockDirent(name, isFile);
  },

  /**
   * Normalizes a path for use with the virtual filesystem
   * 
   * @param {string} path - The path to normalize
   * @returns {string} The normalized path
   * 
   * @example
   * const path = normalizePath('C:\path\to\file.txt'); // returns '/C:/path/to/file.txt'
   */
  normalizePath: function(path) {
    return normalizePathForMemfs(path);
  },

  /**
   * Normalizes a relative path from the virtual root
   * 
   * @param {string} relativePath - The relative path from virtual root
   * @returns {string} The normalized absolute path
   * 
   * @example
   * // Returns '/project/src/file.txt' (normalized for memfs)
   * const path = normalizeTestPath('project/src/file.txt');
   */
  normalizeTestPath: function(relativePath) {
    const pathModule = require('path');
    return normalizePathForMemfs(pathModule.join('/', relativePath));
  },

  /**
   * Inspects the current state of the virtual filesystem for debugging
   * 
   * @param {string} inspectPath - The path to inspect (defaults to root)
   * @returns {Object} JSON representation of the virtual filesystem
   * 
   * @example
   * // Log the entire virtual filesystem
   * console.log(inspectVirtualFs());
   * 
   * // Log a specific directory
   * console.log(inspectVirtualFs('/project'));
   */
  inspectVirtualFs: function(inspectPath = '/') {
    const fs = getVirtualFs();
    const fsState = fs.toJSON(inspectPath);
    return fsState;
  },

  /**
   * Creates a platform-specific test environment
   * 
   * @param {string} platform - The platform to simulate ('win32', 'darwin', 'linux')
   * @param {Object} envVars - Environment variables to set
   * @returns {Function} Function to restore the original environment
   * 
   * @example
   * const restore = setupPlatformEnv('win32', { APPDATA: 'C:\\Users\\Test\\AppData\\Roaming' });
   * // Run your tests...
   * restore(); // Restore original environment
   */
  setupPlatformEnv: function(platform, envVars = {}) {
    const originalPlatform = process.platform;
    const originalEnvVars = {};

    // Save original environment variables
    Object.keys(envVars).forEach(key => {
      originalEnvVars[key] = process.env[key];
    });

    // Set platform
    Object.defineProperty(process, 'platform', { 
      value: platform, 
      configurable: true 
    });

    // Set environment variables
    Object.keys(envVars).forEach(key => {
      process.env[key] = envVars[key];
    });

    // Return function to restore original values
    return function restoreEnv() {
      Object.defineProperty(process, 'platform', { 
        value: originalPlatform,
        configurable: true 
      });
      Object.keys(originalEnvVars).forEach(key => {
        if (originalEnvVars[key] === undefined) {
          delete process.env[key];
        } else {
          process.env[key] = originalEnvVars[key];
        }
      });
    };
  }
};
