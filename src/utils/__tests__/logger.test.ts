/**
 * Tests for the logger utility
 */
import { Logger, LogLevel, configureLogger, logger as defaultLogger } from '../logger';

// Mock console methods
const originalConsoleLog = console.log;
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;

describe('Logger', () => {
  // Setup mocks before each test
  let consoleLogMock: jest.SpyInstance;
  let consoleErrorMock: jest.SpyInstance;
  let consoleWarnMock: jest.SpyInstance;
  
  beforeEach(() => {
    // Mock console methods
    consoleLogMock = jest.spyOn(console, 'log').mockImplementation(() => {});
    consoleErrorMock = jest.spyOn(console, 'error').mockImplementation(() => {});
    consoleWarnMock = jest.spyOn(console, 'warn').mockImplementation(() => {});
  });
  
  afterEach(() => {
    // Restore console methods
    consoleLogMock.mockRestore();
    consoleErrorMock.mockRestore();
    consoleWarnMock.mockRestore();
  });
  
  afterAll(() => {
    // Ensure console is fully restored after all tests
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
    console.warn = originalConsoleWarn;
  });
  
  describe('Logger instance', () => {
    it('should create a logger with default options', () => {
      const logger = new Logger();
      expect(logger.getLevel()).toBe(LogLevel.INFO);
    });
    
    it('should create a logger with custom options', () => {
      const logger = new Logger({ level: LogLevel.DEBUG });
      expect(logger.getLevel()).toBe(LogLevel.DEBUG);
    });
    
    it('should allow changing log level after creation', () => {
      const logger = new Logger();
      expect(logger.getLevel()).toBe(LogLevel.INFO);
      
      logger.setLevel(LogLevel.VERBOSE);
      expect(logger.getLevel()).toBe(LogLevel.VERBOSE);
    });
    
    it('should configure multiple options at once', () => {
      const logger = new Logger();
      logger.configure({ level: LogLevel.ERROR, useColors: false });
      
      expect(logger.getLevel()).toBe(LogLevel.ERROR);
    });
  });
  
  describe('Logging methods', () => {
    let logger: Logger;
    
    beforeEach(() => {
      logger = new Logger(); // Default level: INFO
    });
    
    it('should log errors at any log level', () => {
      logger.setLevel(LogLevel.ERROR);
      logger.error('Test error');
      expect(consoleErrorMock).toHaveBeenCalled();
      
      consoleErrorMock.mockClear();
      logger.error('Test error with object', new Error('Details'));
      expect(consoleErrorMock).toHaveBeenCalledTimes(2); // Once for message, once for error details
    });
    
    it('should log warnings at WARN level and above', () => {
      logger.setLevel(LogLevel.ERROR);
      logger.warn('Test warning');
      expect(consoleWarnMock).not.toHaveBeenCalled();
      
      logger.setLevel(LogLevel.WARN);
      logger.warn('Test warning');
      expect(consoleWarnMock).toHaveBeenCalled();
    });
    
    it('should log info at INFO level and above', () => {
      logger.setLevel(LogLevel.WARN);
      logger.info('Test info');
      expect(consoleLogMock).not.toHaveBeenCalled();
      
      logger.setLevel(LogLevel.INFO);
      logger.info('Test info');
      expect(consoleLogMock).toHaveBeenCalled();
    });
    
    it('should log success at INFO level and above', () => {
      logger.setLevel(LogLevel.WARN);
      logger.success('Test success');
      expect(consoleLogMock).not.toHaveBeenCalled();
      
      logger.setLevel(LogLevel.INFO);
      logger.success('Test success');
      expect(consoleLogMock).toHaveBeenCalled();
    });
    
    it('should log debug at DEBUG level and above', () => {
      logger.setLevel(LogLevel.INFO);
      logger.debug('Test debug');
      expect(consoleLogMock).not.toHaveBeenCalled();
      
      logger.setLevel(LogLevel.DEBUG);
      logger.debug('Test debug');
      expect(consoleLogMock).toHaveBeenCalled();
    });
    
    it('should log verbose at VERBOSE level only', () => {
      logger.setLevel(LogLevel.DEBUG);
      logger.verbose('Test verbose');
      expect(consoleLogMock).not.toHaveBeenCalled();
      
      logger.setLevel(LogLevel.VERBOSE);
      logger.verbose('Test verbose');
      expect(consoleLogMock).toHaveBeenCalled();
    });
    
    it('should always log plain messages regardless of level', () => {
      logger.setLevel(LogLevel.ERROR);
      logger.plain('Test plain');
      expect(consoleLogMock).toHaveBeenCalled();
    });
  });
  
  describe('Default logger and configuration', () => {
    it('should expose a default logger instance', () => {
      expect(defaultLogger).toBeInstanceOf(Logger);
    });
    
    it('should configure the default logger based on options', () => {
      // Start with default (INFO)
      expect(defaultLogger.getLevel()).toBe(LogLevel.INFO);
      
      // Configure for quiet mode
      configureLogger({ quiet: true });
      expect(defaultLogger.getLevel()).toBe(LogLevel.ERROR);
      
      // Configure for verbose mode
      configureLogger({ verbose: true });
      expect(defaultLogger.getLevel()).toBe(LogLevel.VERBOSE);
      
      // Configure for debug mode
      configureLogger({ debug: true });
      expect(defaultLogger.getLevel()).toBe(LogLevel.DEBUG);
      
      // Reset to default
      configureLogger({});
      expect(defaultLogger.getLevel()).toBe(LogLevel.INFO);
    });
    
    it('should handle priority between conflicting options', () => {
      // quiet takes precedence over verbose and debug
      configureLogger({ quiet: true, verbose: true, debug: true });
      expect(defaultLogger.getLevel()).toBe(LogLevel.ERROR);
      
      // debug takes precedence over verbose
      configureLogger({ debug: true, verbose: true });
      expect(defaultLogger.getLevel()).toBe(LogLevel.DEBUG);
    });
  });
});