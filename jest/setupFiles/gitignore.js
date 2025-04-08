/**
 * Gitignore utility mock setup for Jest
 * 
 * This file sets up the standard mocking for gitignore utility functions
 * to ensure consistent testing of gitignore-related functionality.
 * 
 * PREFERRED APPROACH: Import these helpers directly rather than using manual mocks or
 * setting up your own gitignore mocking to ensure consistent test behavior.
 */

// Import the virtualFsUtils and gitignoreUtils
const { addVirtualGitignoreFile } = require('../../src/__tests__/utils/virtualFsUtils');
const gitignoreUtils = require('../../src/utils/gitignoreUtils');

// Export helper functions for convenience (available by importing from './jest/setupFiles/gitignore')
module.exports = {
  /**
   * Creates a virtual .gitignore file
   * 
   * @param {string} gitignorePath - Path where the .gitignore file should be created
   * @param {string} content - Content of the .gitignore file
   * 
   * @example
   * addGitignoreFile('/project/.gitignore', '*.log\n/node_modules/');
   */
  addGitignoreFile: function(gitignorePath, content) {
    return addVirtualGitignoreFile(gitignorePath, content);
  },
  
  /**
   * Sets up a basic gitignore environment in the virtual filesystem
   * 
   * This function creates a default project structure with a .gitignore file
   * containing common patterns. It's a convenience wrapper to quickly set up
   * a standard gitignore testing environment.
   * 
   * @param {string} [projectPath='/project'] - Base path for the project
   * @param {string} [patterns='node_modules/\n*.log\n.DS_Store\n/dist/'] - Default gitignore patterns
   * 
   * @example
   * setupBasicGitignore();
   * // or
   * setupBasicGitignore('/custom/path', '*.txt\n/build/');
   */
  setupBasicGitignore: function(projectPath = '/project', patterns = 'node_modules/\n*.log\n.DS_Store\n/dist/') {
    return addVirtualGitignoreFile(`${projectPath}/.gitignore`, patterns);
  },
  
  /**
   * Clears the gitignore cache to ensure test isolation
   * 
   * This should be called in beforeEach to prevent test interdependencies
   * 
   * @example
   * beforeEach(() => {
   *   clearGitignoreCache();
   * });
   */
  clearGitignoreCache: function() {
    if (gitignoreUtils.clearIgnoreCache) {
      gitignoreUtils.clearIgnoreCache();
    }
  },
  
  /**
   * Creates a mock gitignore result for a path
   * 
   * @param {boolean} shouldIgnore - Whether the path should be ignored
   * @returns {Function} A mock function that returns the specified result
   * 
   * @example
   * const mockIgnore = createGitignoreMock(true);
   * jest.spyOn(gitignoreUtils, 'shouldIgnorePath').mockImplementation(mockIgnore);
   */
  createGitignoreMock: function(shouldIgnore = true) {
    return jest.fn().mockResolvedValue(shouldIgnore);
  }
};