/**
 * Jest setup file for test environment configuration
 * 
 * This file is executed after the test framework is installed in the environment
 * but before each test file is executed. It's ideal for adding global test utilities
 * and setting up global before/after hooks.
 */

// Note: We don't need to import Jest globals since they're already available in the global scope
// In the Jest test environment, afterEach, beforeEach, and jest are already defined

// Import common test utilities for convenience
const { resetVirtualFs } = require('../src/__tests__/utils/virtualFsUtils');

// Reset mocks after each test
afterEach(() => {
  // Clear all mock implementations and history
  jest.clearAllMocks();
});

// Log setup information (helpful for debugging test environment issues)
console.log('Jest environment setup complete');