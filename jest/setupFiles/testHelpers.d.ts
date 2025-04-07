/**
 * Type definitions for common test helpers
 */

/**
 * Ensures a value is wrapped in a promise
 * @param value - The value to promisify
 * @returns Promise that resolves to the value
 */
export function promisify<T>(value: T): Promise<T>;

/**
 * Waits for the specified number of milliseconds
 * @param ms - Milliseconds to wait
 * @returns Promise that resolves after the specified time
 */
export function wait(ms: number): Promise<void>;

/**
 * Creates a mock object with spies for all methods
 * @param implementation - Implementation of the mock
 * @returns Mock object with all methods spied on
 */
export function createMockObject<T extends Record<string, unknown>>(implementation?: T): Record<keyof T, jest.Mock>;

/**
 * Mock spinner interface
 */
export interface MockSpinner {
  start: jest.Mock;
  stop: jest.Mock;
  succeed: jest.Mock;
  fail: jest.Mock;
  warn: jest.Mock;
  info: jest.Mock;
  text: string;
}

/**
 * Creates a mock spinner for testing CLI output
 * @returns Mock spinner object with common methods
 */
export function createMockSpinner(): MockSpinner;
