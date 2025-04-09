/**
 * Tests to verify the interface definitions can be implemented and used correctly
 */
import { ConsoleLogger, UISpinner } from '../interfaces';

describe('Interface Definitions', () => {
  describe('ConsoleLogger', () => {
    it('can be implemented as a mock', () => {
      // Create a simple mock implementation of the ConsoleLogger interface
      const mockLogger: ConsoleLogger = {
        error: jest.fn(),
        warn: jest.fn(),
        info: jest.fn(),
        success: jest.fn(),
        debug: jest.fn(),
        plain: jest.fn(),
      };

      // Use the mock implementation
      mockLogger.info('Test info message');
      mockLogger.error('Test error message', new Error('Test error'));
      mockLogger.warn('Test warning message');
      mockLogger.success('Test success message');
      mockLogger.debug('Test debug message');
      mockLogger.plain('Test plain message');

      // Verify the methods were called with correct arguments
      expect(mockLogger.info).toHaveBeenCalledWith('Test info message');
      expect(mockLogger.error).toHaveBeenCalledWith('Test error message', expect.any(Error));
      expect(mockLogger.warn).toHaveBeenCalledWith('Test warning message');
      expect(mockLogger.success).toHaveBeenCalledWith('Test success message');
      expect(mockLogger.debug).toHaveBeenCalledWith('Test debug message');
      expect(mockLogger.plain).toHaveBeenCalledWith('Test plain message');
    });

    it('can be used for dependency injection', () => {
      // Create a function that uses the ConsoleLogger interface
      function logMessage(logger: ConsoleLogger, level: string, message: string): void {
        switch (level) {
          case 'info':
            logger.info(message);
            break;
          case 'error':
            logger.error(message);
            break;
          case 'warn':
            logger.warn(message);
            break;
          case 'success':
            logger.success(message);
            break;
          case 'debug':
            logger.debug(message);
            break;
          default:
            logger.plain(message);
        }
      }

      // Create a mock logger
      const mockLogger: ConsoleLogger = {
        error: jest.fn(),
        warn: jest.fn(),
        info: jest.fn(),
        success: jest.fn(),
        debug: jest.fn(),
        plain: jest.fn(),
      };

      // Use the function with the mock logger
      logMessage(mockLogger, 'info', 'Test info message');

      // Verify the mock was called correctly
      expect(mockLogger.info).toHaveBeenCalledWith('Test info message');
    });
  });

  describe('UISpinner', () => {
    it('can be implemented as a mock with chaining methods', () => {
      // Create a mock implementation of the UISpinner interface
      const mockSpinner: UISpinner = {
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        text: '',
        isSpinning: false,
      };

      // Use the mock implementation with chaining
      mockSpinner
        .start('Starting process')
        .setText('Processing item 1')
        .setText('Processing item 2')
        .succeed('Process completed');

      // Verify the methods were called with correct arguments
      expect(mockSpinner.start).toHaveBeenCalledWith('Starting process');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item 1');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item 2');
      expect(mockSpinner.succeed).toHaveBeenCalledWith('Process completed');
    });

    it('can be used for dependency injection', () => {
      // Create a function that uses the UISpinner interface
      function runProcessWithProgress(spinner: UISpinner, items: string[]): void {
        spinner.start('Starting process');

        items.forEach(item => {
          spinner.setText(`Processing ${item}`);
          // Process the item...
        });

        spinner.succeed('Process completed');
      }

      // Create a mock spinner
      const mockSpinner: UISpinner = {
        start: jest.fn().mockReturnThis(),
        stop: jest.fn().mockReturnThis(),
        succeed: jest.fn().mockReturnThis(),
        fail: jest.fn().mockReturnThis(),
        warn: jest.fn().mockReturnThis(),
        info: jest.fn().mockReturnThis(),
        setText: jest.fn().mockReturnThis(),
        text: '',
        isSpinning: false,
      };

      // Use the function with the mock spinner
      runProcessWithProgress(mockSpinner, ['item1', 'item2']);

      // Verify the mock was called correctly
      expect(mockSpinner.start).toHaveBeenCalledWith('Starting process');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item1');
      expect(mockSpinner.setText).toHaveBeenCalledWith('Processing item2');
      expect(mockSpinner.succeed).toHaveBeenCalledWith('Process completed');
    });
  });
});
