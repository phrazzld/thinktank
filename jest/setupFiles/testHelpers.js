/**
 * Common test helper functions for Jest tests
 * 
 * This file provides utility functions that are commonly needed across
 * multiple test files, such as waiting for promises, mocking timers,
 * or creating test fixtures.
 */

// Export helper functions for convenience (available by importing from './jest/setupFiles/testHelpers')
module.exports = {
  /**
   * Ensures a value is wrapped in a promise
   * @param {any} value - The value to promisify
   * @returns {Promise<any>} Promise that resolves to the value
   */
  promisify: function(value) {
    return Promise.resolve(value);
  },

  /**
   * Waits for the specified number of milliseconds
   * @param {number} ms - Milliseconds to wait
   * @returns {Promise<void>} Promise that resolves after the specified time
   */
  wait: function(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  },

  /**
   * Creates a mock object with spies for all methods
   * @param {Object} implementation - Implementation of the mock
   * @returns {Object} Mock object with all methods spied on
   */
  createMockObject: function(implementation = {}) {
    return Object.entries(implementation).reduce((mockObj, [key, value]) => {
      if (typeof value === 'function') {
        mockObj[key] = jest.fn(value);
      } else {
        mockObj[key] = value;
      }
      return mockObj;
    }, {});
  },

  /**
   * Creates a mock spinner for testing CLI output
   * @returns {Object} Mock spinner object with common methods
   */
  createMockSpinner: function() {
    return {
      start: jest.fn(),
      stop: jest.fn(),
      succeed: jest.fn(),
      fail: jest.fn(),
      warn: jest.fn(),
      info: jest.fn(),
      text: ''
    };
  }
};