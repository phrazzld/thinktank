/**
 * Filesystem mock setup for Jest
 * 
 * This file sets up the basic mocking for filesystem modules (fs, fs/promises)
 * using the virtualFsUtils implementation.
 * 
 * Tests should import the mock utilities from src/__tests__/utils/virtualFsUtils
 * to create specific filesystem configurations.
 */

// Import the mock utilities
const { mockFsModules } = require('../../src/__tests__/utils/virtualFsUtils');

// Mock the fs module
jest.mock('fs', () => mockFsModules().fs);

// Mock the fs/promises module
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Export helper functions for convenience (available by importing from './jest/setupFiles/fs')
module.exports = {
  /**
   * Sets up a basic filesystem structure for tests
   * @param {Object} structure - Object mapping file paths to content
   */
  setupBasicFs: function(structure = {}) {
    const { resetVirtualFs, createVirtualFs } = require('../../src/__tests__/utils/virtualFsUtils');
    
    resetVirtualFs();
    createVirtualFs(structure);
  },

  /**
   * Creates a standardized filesystem error
   * @param {string} code - Error code (e.g., 'ENOENT', 'EACCES')
   * @param {string} message - Error message
   * @param {string} syscall - System call that failed
   * @param {string} filepath - Path that caused the error
   * @returns {NodeJS.ErrnoException} A properly formatted filesystem error
   */
  createFsError: function(code, message, syscall, filepath) {
    const { createFsError } = require('../../src/__tests__/utils/virtualFsUtils');
    return createFsError(code, message, syscall, filepath);
  }
};