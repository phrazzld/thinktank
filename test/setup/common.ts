/**
 * Common test setup utilities
 * 
 * This module provides standard setup and teardown hooks for tests,
 * ensuring consistent test environment and proper isolation between tests.
 */
import { resetVirtualFs } from '../../src/__tests__/utils/virtualFsUtils';
import { clearIgnoreCache } from '../../src/utils/gitignoreUtils';

/**
 * Sets up standard beforeEach and afterEach hooks for tests
 * 
 * This function configures:
 * - beforeEach: Resets virtual filesystem, clears gitignore cache, resets Jest mocks
 * - afterEach: Restores Jest mocks
 * 
 * Usage:
 * ```typescript
 * import { setupTestHooks } from '../../../test/setup/common';
 * 
 * describe('My test suite', () => {
 *   setupTestHooks(); // Sets up hooks for all tests in this suite
 *   
 *   it('should do something', () => {
 *     // Test runs with clean environment
 *   });
 * });
 * ```
 */
export function setupTestHooks(): void {
  beforeEach(() => {
    // Reset virtual filesystem
    resetVirtualFs();
    
    // Clear gitignore cache to prevent test interdependencies
    clearIgnoreCache();
    
    // Reset all Jest mocks
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    // Restore all mocked functions
    jest.restoreAllMocks();
  });
}

/**
 * Mocks the process.env object with specific environment variables
 * 
 * @param envVars - Object containing environment variables to mock
 * @returns Function to restore the original environment
 * 
 * Usage:
 * ```typescript
 * const restoreEnv = mockEnv({ NODE_ENV: 'test', API_KEY: 'test-key' });
 * // Test with mocked environment
 * restoreEnv(); // Restore original environment
 * ```
 */
export function mockEnv(envVars: Record<string, string>): () => void {
  const originalEnv = { ...process.env };
  
  // Set mock environment variables
  Object.entries(envVars).forEach(([key, value]) => {
    process.env[key] = value;
  });
  
  // Return function to restore original environment
  return () => {
    // Restore original environment variables
    Object.keys(envVars).forEach((key) => {
      if (key in originalEnv) {
        process.env[key] = originalEnv[key];
      } else {
        delete process.env[key];
      }
    });
  };
}

/**
 * Creates a test-friendly random string for use in test fixtures
 * 
 * @param prefix - Optional prefix for the random string
 * @returns A random string prefixed with 'test-' or the provided prefix
 * 
 * Usage:
 * ```typescript
 * const testId = createTestId(); // e.g., 'test-a1b2c3'
 * const namedTestId = createTestId('user'); // e.g., 'user-a1b2c3'
 * ```
 */
export function createTestId(prefix = 'test'): string {
  return `${prefix}-${Math.random().toString(36).substring(2, 8)}`;
}

/**
 * Waits for a specified duration in milliseconds
 * Useful for testing asynchronous behavior
 * 
 * @param ms - Milliseconds to wait
 * @returns Promise that resolves after the specified time
 * 
 * Usage:
 * ```typescript
 * await wait(100); // Waits for 100ms
 * ```
 */
export function wait(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}
