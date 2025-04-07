/**
 * Test helpers for Jest tests
 * 
 * This module provides utility functions that are useful for testing, including
 * convenience functions for working with promises and timing.
 */

// Re-export the createFsError function from virtualFsUtils for convenience
import { createFsError } from './virtualFsUtils';
export { createFsError };

/**
 * Ensure a value is wrapped in a promise
 * 
 * Useful for tests where you want to ensure a value is a promise
 * or when mocking async functions
 */
export function promisify<T>(value: T): Promise<T> {
  return Promise.resolve(value);
}

/**
 * Wait for a specified number of milliseconds
 * 
 * Useful for adding brief pauses in tests without using setTimeout
 */
export async function wait(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}