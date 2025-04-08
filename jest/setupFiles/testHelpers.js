/**
 * Common test helper functions for Jest tests
 * 
 * This file provides utility functions that are commonly needed across
 * multiple test files, such as waiting for promises, mocking timers,
 * or creating test fixtures.
 * 
 * PREFERRED APPROACH: Import these helpers directly rather than creating
 * similar utilities in individual test files to ensure consistent behavior.
 */

// Export helper functions for convenience (available by importing from './jest/setupFiles/testHelpers')
module.exports = {
  /**
   * Ensures a value is wrapped in a promise
   * 
   * @param {any} value - The value to promisify
   * @returns {Promise<any>} Promise that resolves to the value
   * 
   * @example
   * const result = await promisify('test');
   */
  promisify: function(value) {
    return Promise.resolve(value);
  },

  /**
   * Waits for the specified number of milliseconds
   * 
   * @param {number} ms - Milliseconds to wait
   * @returns {Promise<void>} Promise that resolves after the specified time
   * 
   * @example
   * await wait(100); // waits 100ms
   */
  wait: function(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  },

  /**
   * Creates a mock object with spies for all methods
   * 
   * @param {Object} implementation - Implementation of the mock
   * @returns {Object} Mock object with all methods spied on
   * 
   * @example
   * const mockApi = createMockObject({
   *   getData: () => ({ result: 'test' }),
   *   processData: (data) => data.toUpperCase()
   * });
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
   * 
   * @returns {Object} Mock spinner object with common methods
   * 
   * @example
   * const spinner = createMockSpinner();
   * // Later in test
   * expect(spinner.succeed).toHaveBeenCalledWith('Operation completed');
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
  },
  
  /**
   * Creates a standard mock for HTTP/network requests
   * 
   * @param {Object} response - The response object to return
   * @param {number} [statusCode=200] - HTTP status code
   * @returns {Function} A mock function for network requests
   * 
   * @example
   * jest.spyOn(global, 'fetch').mockImplementation(createNetworkMock({ data: 'test' }));
   */
  createNetworkMock: function(response, statusCode = 200) {
    return jest.fn().mockResolvedValue({
      ok: statusCode >= 200 && statusCode < 300,
      status: statusCode,
      json: jest.fn().mockResolvedValue(response),
      text: jest.fn().mockResolvedValue(JSON.stringify(response))
    });
  },
  
  /**
   * Creates a network error mock for testing error handling
   * 
   * @param {string} message - Error message
   * @param {string} [name='NetworkError'] - Error name/type
   * @returns {Function} A mock function that rejects with the specified error
   * 
   * @example
   * jest.spyOn(global, 'fetch').mockImplementation(createNetworkErrorMock('Connection refused'));
   */
  createNetworkErrorMock: function(message, name = 'NetworkError') {
    return jest.fn().mockImplementation(() => {
      const error = new Error(message);
      error.name = name;
      return Promise.reject(error);
    });
  },
  
  /**
   * Creates a mock for LLM API providers responses
   * 
   * @param {Object} responseData - The LLM response data to return
   * @param {boolean} [success=true] - Whether the response should be successful
   * @returns {Object} A mock LLM response object
   * 
   * @example
   * const mockResponse = createLlmResponseMock({ text: 'Generated text' });
   * jest.spyOn(provider, 'generate').mockResolvedValue(mockResponse);
   */
  createLlmResponseMock: function(responseData, success = true) {
    return {
      modelId: 'mock-model',
      providerId: 'mock-provider',
      response: responseData.text || 'Mock response',
      rawResponse: responseData,
      error: success ? null : new Error('LLM API Error'),
      usage: {
        promptTokens: 10,
        completionTokens: 20,
        totalTokens: 30
      },
      timestamp: new Date().toISOString()
    };
  }
};
