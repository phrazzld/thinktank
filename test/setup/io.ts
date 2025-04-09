/**
 * I/O utilities setup for tests
 * 
 * This module provides specialized setup helpers for I/O-related interfaces
 * like ConsoleLogger and UISpinner, enabling consistent and simple mocking
 * of these dependencies in tests.
 */
import { ConsoleLogger, UISpinner } from '../../src/core/interfaces';

/**
 * Creates a mock implementation of the ConsoleLogger interface for testing.
 * Each method is a Jest mock function.
 *
 * @returns A Jest-mocked ConsoleLogger instance.
 * 
 * Usage:
 * ```typescript
 * const mockLogger = createMockConsoleLogger();
 * // Configure specific behaviors if needed:
 * mockLogger.info.mockImplementation((msg) => console.log(`[TEST] ${msg}`));
 * // Test that the logger was called with the expected arguments:
 * expect(mockLogger.info).toHaveBeenCalledWith('Processing files...');
 * ```
 */
export function createMockConsoleLogger(): jest.Mocked<ConsoleLogger> {
  return {
    error: jest.fn(),
    warn: jest.fn(),
    info: jest.fn(),
    success: jest.fn(),
    debug: jest.fn(),
    plain: jest.fn(),
  } as unknown as jest.Mocked<ConsoleLogger>;
}

/**
 * Creates a mock implementation of the UISpinner interface for testing.
 * Methods like start, stop, succeed, fail return `this` for proper chaining behavior.
 *
 * @returns A Jest-mocked UISpinner instance.
 * 
 * Usage:
 * ```typescript
 * const mockSpinner = createMockUISpinner();
 * // The mock spinner can be used in place of a real spinner:
 * mockSpinner.start('Processing files...').succeed('Done!');
 * // Test that the spinner methods were called with the expected arguments:
 * expect(mockSpinner.start).toHaveBeenCalledWith('Processing files...');
 * expect(mockSpinner.succeed).toHaveBeenCalledWith('Done!');
 * ```
 */
export function createMockUISpinner(): jest.Mocked<UISpinner> {
  // Create a basic object to form the spinner mock
  const spinner: Partial<UISpinner> = {
    text: '',
    isSpinning: false
  };

  // Add chainable methods that return the spinner instance
  spinner.start = jest.fn().mockImplementation(function(this: UISpinner, text?: string) {
    if (text) this.text = text;
    this.isSpinning = true;
    return this;
  });

  spinner.stop = jest.fn().mockImplementation(function(this: UISpinner) {
    this.isSpinning = false;
    return this;
  });

  spinner.succeed = jest.fn().mockImplementation(function(this: UISpinner, text?: string) {
    if (text) this.text = text;
    this.isSpinning = false;
    return this;
  });

  spinner.fail = jest.fn().mockImplementation(function(this: UISpinner, text?: string) {
    if (text) this.text = text;
    this.isSpinning = false;
    return this;
  });

  spinner.warn = jest.fn().mockImplementation(function(this: UISpinner, text?: string) {
    if (text) this.text = text;
    this.isSpinning = false;
    return this;
  });

  spinner.info = jest.fn().mockImplementation(function(this: UISpinner, text?: string) {
    if (text) this.text = text;
    this.isSpinning = false;
    return this;
  });

  spinner.setText = jest.fn().mockImplementation(function(this: UISpinner, text: string) {
    this.text = text;
    return this;
  });

  // Bind the methods to the spinner object
  // Each method is already bound in its mockImplementation using 'this'

  return spinner as unknown as jest.Mocked<UISpinner>;
}
