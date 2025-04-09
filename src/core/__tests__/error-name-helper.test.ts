/**
 * Tests for the error name helper method functionality
 *
 * These tests verify that the error name property is correctly set
 * for all error classes in the hierarchy.
 */
import { ThinktankError, ConfigError } from '../errors';

describe('Error Name Helper Method', () => {
  // Test current behavior (for baseline)
  test('base ThinktankError currently sets name manually', () => {
    const error = new ThinktankError('Test error');
    expect(error.name).toBe('ThinktankError');
  });

  test('subclass ConfigError currently sets name manually', () => {
    const error = new ConfigError('Test error');
    expect(error.name).toBe('ConfigError');
  });

  // Test inheritance of name property through dynamic subclassing
  test('custom subclass should automatically get correct name', () => {
    // Create a custom error subclass
    class CustomTestError extends ThinktankError {
      constructor(message: string) {
        super(message);
        // Note: Not setting this.name explicitly
      }
    }

    const error = new CustomTestError('Test error');
    expect(error.name).toBe('CustomTestError');
  });

  // Test deeper inheritance chain
  test('error name is correctly set in deep inheritance chain', () => {
    // Create a deeper inheritance chain
    class Level1Error extends ThinktankError {
      constructor(message: string) {
        super(message);
        // No explicit name set here
      }
    }

    class Level2Error extends Level1Error {
      constructor(message: string) {
        super(message);
        // No explicit name set here
      }
    }

    class Level3Error extends Level2Error {
      constructor(message: string) {
        super(message);
        // No explicit name set here
      }
    }

    const error = new Level3Error('Test error');
    expect(error.name).toBe('Level3Error');
  });

  // Test instanceof behavior
  test('instanceof check works with custom error classes', () => {
    class CustomTestError extends ThinktankError {
      constructor(message: string) {
        super(message);
        // No explicit name set here
      }
    }

    const error = new CustomTestError('Test error');
    expect(error instanceof CustomTestError).toBe(true);
    expect(error instanceof ThinktankError).toBe(true);
    expect(error instanceof Error).toBe(true);
  });
});
