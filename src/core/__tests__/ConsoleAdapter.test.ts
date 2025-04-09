/**
 * Tests for ConsoleAdapter implementation
 */
import { ConsoleAdapter } from '../ConsoleAdapter';
import { Logger } from '../../utils/logger';

// Create a mock Logger class
class MockLogger {
  error = jest.fn();
  warn = jest.fn();
  info = jest.fn();
  success = jest.fn();
  debug = jest.fn();
  plain = jest.fn();
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
    it('should delegate to logger.error', () => {
      // Arrange
      const message = 'Test error message';
      const error = new Error('Test error');

      // Act
      consoleAdapter.error(message, error);

      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(message, error);
    });

    it('should handle being called without an error object', () => {
      // Arrange
      const message = 'Test error message';

      // Act
      consoleAdapter.error(message);

      // Assert
      expect(mockLogger.error).toHaveBeenCalledWith(message, undefined);
    });
  });

  describe('warn', () => {
    it('should delegate to logger.warn', () => {
      // Arrange
      const message = 'Test warning message';

      // Act
      consoleAdapter.warn(message);

      // Assert
      expect(mockLogger.warn).toHaveBeenCalledWith(message);
    });
  });

  describe('info', () => {
    it('should delegate to logger.info', () => {
      // Arrange
      const message = 'Test info message';

      // Act
      consoleAdapter.info(message);

      // Assert
      expect(mockLogger.info).toHaveBeenCalledWith(message);
    });
  });

  describe('success', () => {
    it('should delegate to logger.success', () => {
      // Arrange
      const message = 'Test success message';

      // Act
      consoleAdapter.success(message);

      // Assert
      expect(mockLogger.success).toHaveBeenCalledWith(message);
    });
  });

  describe('debug', () => {
    it('should delegate to logger.debug', () => {
      // Arrange
      const message = 'Test debug message';

      // Act
      consoleAdapter.debug(message);

      // Assert
      expect(mockLogger.debug).toHaveBeenCalledWith(message);
    });
  });

  describe('plain', () => {
    it('should delegate to logger.plain', () => {
      // Arrange
      const message = 'Test plain message';

      // Act
      consoleAdapter.plain(message);

      // Assert
      expect(mockLogger.plain).toHaveBeenCalledWith(message);
    });
  });

  // This test verifies the constructor behavior
  describe('constructor', () => {
    it('should accept a custom logger instance', () => {
      // This is already tested by all the other tests that use mockLogger

      // Create a minimal mock just for this test
      const customLogger = {
        info: jest.fn(),
      };

      // Create adapter with our custom logger
      const adapterWithCustomLogger = new ConsoleAdapter(customLogger as unknown as Logger);

      // Use the adapter
      adapterWithCustomLogger.info('Custom logger test');

      // Verify custom logger was used
      expect(customLogger.info).toHaveBeenCalledWith('Custom logger test');
    });

    // Note: Testing the default singleton logger behavior would require
    // complex mocking of the module system and isn't necessary for
    // validating the adapter's behavior, as we've already tested
    // that it correctly forwards all method calls.
  });
});
