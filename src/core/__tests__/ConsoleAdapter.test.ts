/**
 * Tests for ConsoleAdapter implementation
 * 
 * These tests verify that ConsoleAdapter properly delegates all ConsoleLogger interface
 * method calls to the provided logger instance. The tests focus on testing behavior
 * (delegation) rather than implementation details, following the project's testing
 * philosophy.
 */
import { ConsoleAdapter } from '../ConsoleAdapter';
import { Logger } from '../../utils/logger';
import { ConsoleLogger } from '../interfaces';

/**
 * Mock implementation of Logger for testing
 * Implements all methods from Logger interface with Jest mock functions
 */
class MockLogger implements Partial<Logger> {
  error = jest.fn();
  warn = jest.fn();
  info = jest.fn();
  success = jest.fn();
  debug = jest.fn();
  plain = jest.fn();
  verbose = jest.fn();
  setLevel = jest.fn();
  getLevel = jest.fn();
  setColors = jest.fn();
  setPrefix = jest.fn();
  setTimestamp = jest.fn();
  configure = jest.fn();
}

describe('ConsoleAdapter', () => {
  let mockLogger: MockLogger;
  let consoleAdapter: ConsoleAdapter;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Create a new MockLogger for each test
    mockLogger = new MockLogger();

    // Create the adapter with our mock logger
    consoleAdapter = new ConsoleAdapter(mockLogger as unknown as Logger);
  });

  describe('error', () => {
    it('should delegate to logger.error with message and error object', () => {
      // Arrange
      const message = 'Test error message';
      const error = new Error('Test error');

      // Act
      consoleAdapter.error(message, error);

      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(message, error);
      expect(mockLogger.error).toHaveBeenCalledTimes(1);
    });

    it('should correctly handle being called without an error object', () => {
      // Arrange
      const message = 'Test error message';

      // Act
      consoleAdapter.error(message);

      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(message, undefined);
      expect(mockLogger.error).toHaveBeenCalledTimes(1);
    });

    it('should correctly handle empty message', () => {
      // Arrange
      const message = '';

      // Act
      consoleAdapter.error(message);

      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(message, undefined);
      expect(mockLogger.error).toHaveBeenCalledTimes(1);
    });
  });

  describe('warn', () => {
    it('should delegate to logger.warn with the provided message', () => {
      // Arrange
      const message = 'Test warning message';

      // Act
      consoleAdapter.warn(message);

      // Assert
      expect(mockLogger.warn).toHaveBeenCalledWith(message);
      expect(mockLogger.warn).toHaveBeenCalledTimes(1);
    });

    it('should handle empty message', () => {
      // Act
      consoleAdapter.warn('');

      // Assert
      expect(mockLogger.warn).toHaveBeenCalledWith('');
      expect(mockLogger.warn).toHaveBeenCalledTimes(1);
    });
  });

  describe('info', () => {
    it('should delegate to logger.info with the provided message', () => {
      // Arrange
      const message = 'Test info message';

      // Act
      consoleAdapter.info(message);

      // Assert
      expect(mockLogger.info).toHaveBeenCalledWith(message);
      expect(mockLogger.info).toHaveBeenCalledTimes(1);
    });
  });

  describe('success', () => {
    it('should delegate to logger.success with the provided message', () => {
      // Arrange
      const message = 'Test success message';

      // Act
      consoleAdapter.success(message);

      // Assert
      expect(mockLogger.success).toHaveBeenCalledWith(message);
      expect(mockLogger.success).toHaveBeenCalledTimes(1);
    });
  });

  describe('debug', () => {
    it('should delegate to logger.debug with the provided message', () => {
      // Arrange
      const message = 'Test debug message';

      // Act
      consoleAdapter.debug(message);

      // Assert
      expect(mockLogger.debug).toHaveBeenCalledWith(message);
      expect(mockLogger.debug).toHaveBeenCalledTimes(1);
    });
  });

  describe('plain', () => {
    it('should delegate to logger.plain with the provided message', () => {
      // Arrange
      const message = 'Test plain message';

      // Act
      consoleAdapter.plain(message);

      // Assert
      expect(mockLogger.plain).toHaveBeenCalledWith(message);
      expect(mockLogger.plain).toHaveBeenCalledTimes(1);
    });
  });

  describe('constructor dependency injection', () => {
    it('should accept a custom logger instance', () => {
      // Create a minimal mock with only the methods we'll use
      const customLogger = {
        info: jest.fn(),
        error: jest.fn(),
      };

      // Create adapter with our custom logger
      const adapterWithCustomLogger = new ConsoleAdapter(customLogger as unknown as Logger);

      // Use multiple methods to verify proper delegation
      adapterWithCustomLogger.info('Custom logger info test');
      adapterWithCustomLogger.error('Custom logger error', new Error('Test'));

      // Verify custom logger was used for both calls
      expect(customLogger.info).toHaveBeenCalledWith('Custom logger info test');
      expect(customLogger.error).toHaveBeenCalledWith('Custom logger error', new Error('Test'));
    });

    // Note: Testing the default singleton logger behavior would require
    // complex mocking of the module system and isn't necessary for
    // validating the adapter's behavior, as we've already tested
    // that it correctly forwards all method calls.
  });

  describe('interface compliance', () => {
    it('should implement all ConsoleLogger methods', () => {
      // This test verifies that the adapter implements all interface methods
      // by calling each one and ensuring it doesn't throw
      
      // Define the methods that should be implemented
      const methods: Array<keyof ConsoleLogger> = [
        'error', 'warn', 'info', 'success', 'debug', 'plain'
      ];
      
      // Check that each method exists and is callable
      methods.forEach(method => {
        expect(typeof consoleAdapter[method]).toBe('function');
        
        // Call the method (with appropriate arguments)
        if (method === 'error') {
          consoleAdapter[method]('Test message', new Error('Test error'));
        } else {
          (consoleAdapter[method] as (msg: string) => void)('Test message');
        }
        
        // Verify the corresponding mock method was called
        expect(mockLogger[method]).toHaveBeenCalled();
      });
    });
  });
});
