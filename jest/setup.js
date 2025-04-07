/**
 * Global Jest setup file
 * 
 * This file is executed once before all tests to set up global configuration.
 * It runs before the test framework is installed into the environment.
 */

// Set a longer timeout for tests that may take longer to complete
jest.setTimeout(10000);

// Common mock setup imports
require('./setupFiles/fs');
require('./setupFiles/gitignore');

// Handle global error formatting (optional)
// Example: Configure how Jest formats error messages
// const errorMessageFormatter = {};
// expect.addSnapshotSerializer(errorMessageFormatter);