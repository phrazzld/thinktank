/**
 * Gitignore utility mock setup for Jest
 * 
 * This file sets up the basic mocking for gitignore utility functions.
 * We previously used mockGitignoreUtils, but we've now migrated to using
 * the actual gitignoreUtils implementation with a virtual filesystem.
 */

// Import the virtualFsUtils to ensure it's available
const { addVirtualGitignoreFile } = require('../../src/__tests__/utils/virtualFsUtils');

// We no longer need to mock gitignoreUtils since tests use the actual implementation
// with a virtual filesystem

// Export helper functions for convenience (available by importing from './jest/setupFiles/gitignore')
module.exports = {
  /**
   * Creates a virtual .gitignore file
   * @param {string} gitignorePath - Path where the .gitignore file should be created
   * @param {string} content - Content of the .gitignore file
   */
  addGitignoreFile: function(gitignorePath, content) {
    addVirtualGitignoreFile(gitignorePath, content);
  }
};