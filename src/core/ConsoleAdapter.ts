/**
 * Adapter implementation of the ConsoleLogger interface
 * Wraps existing logging functionality from logger.ts
 */
import { ConsoleLogger } from './interfaces';
import { Logger, logger as singletonLogger } from '../utils/logger';

/**
 * ConsoleAdapter implementation of the ConsoleLogger interface that wraps
 * existing logging functionality from logger.ts. This adapter allows
 * higher-level components to depend on an interface rather than concrete
 * implementation, facilitating dependency injection and testability.
 */
export class ConsoleAdapter implements ConsoleLogger {
  private loggerInstance: Logger;

  /**
   * Creates a new ConsoleAdapter.
   * @param loggerToUse Optional Logger instance to use. Defaults to the global singleton.
   */
  constructor(loggerToUse: Logger = singletonLogger) {
    this.loggerInstance = loggerToUse;
  }

  /**
   * Logs an error message with the highest severity
   * @param message - The error message to log
   * @param error - Optional error object for additional context
   */
  error(message: string, error?: Error): void {
    this.loggerInstance.error(message, error);
  }

  /**
   * Logs a warning message with high severity
   * @param message - The warning message to log
   */
  warn(message: string): void {
    this.loggerInstance.warn(message);
  }

  /**
   * Logs an informational message with normal priority
   * @param message - The informational message to log
   */
  info(message: string): void {
    this.loggerInstance.info(message);
  }

  /**
   * Logs a success message with a success indicator
   * @param message - The success message to log
   */
  success(message: string): void {
    this.loggerInstance.success(message);
  }

  /**
   * Logs a debug message with lower priority
   * @param message - The debug message to log
   */
  debug(message: string): void {
    this.loggerInstance.debug(message);
  }

  /**
   * Logs a plain message without any specific styling
   * @param message - The plain message to log
   */
  plain(message: string): void {
    this.loggerInstance.plain(message);
  }
}
