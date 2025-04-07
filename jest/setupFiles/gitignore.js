/**
 * Gitignore utility mock setup for Jest
 * 
 * This file sets up the basic mocking for gitignore utility functions
 * using the mockGitignoreUtils implementation.
 * 
 * Tests should import the mock utilities from src/__tests__/utils/mockGitignoreUtils
 * to configure specific gitignore behaviors.
 */

// Import the mock module to ensure it's available
const mockGitignoreUtils = require('../../src/__tests__/utils/mockGitignoreUtils');

// Mock the gitignore utilities
jest.mock('../../src/utils/gitignoreUtils');

// Export helper functions for convenience (available by importing from './jest/setupFiles/gitignore')
module.exports = {
  /**
   * Sets up basic gitignore mocking with standard ignored patterns
   * @param {Object} config - Configuration options for mock setup
   */
  setupBasicGitignore: function(config = {}) {
    const { resetMockGitignore, setupMockGitignore } = require('../../src/__tests__/utils/mockGitignoreUtils');
    
    resetMockGitignore();
    setupMockGitignore(config);
  },

  /**
   * Configures gitignore mocking based on virtual filesystem .gitignore files
   * @param {string} rootPath - Root path to start scanning from
   */
  setupGitignoreFromVirtualFs: function(rootPath = '/') {
    const {
      resetMockGitignore,
      setupMockGitignore,
      configureMockGitignoreFromVirtualFs
    } = require('../../src/__tests__/utils/mockGitignoreUtils');
    
    resetMockGitignore();
    setupMockGitignore();
    configureMockGitignoreFromVirtualFs(rootPath);
  },

  /**
   * Creates a virtual .gitignore file and configures mocks based on its content
   * @param {string} gitignorePath - Path where the .gitignore file should be created
   * @param {string} content - Content of the .gitignore file
   */
  addGitignoreFile: function(gitignorePath, content) {
    const { addVirtualGitignoreFile } = require('../../src/__tests__/utils/mockGitignoreUtils');
    addVirtualGitignoreFile(gitignorePath, content);
  }
};