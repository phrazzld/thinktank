/**
 * Logger utility module for consistent logging throughout the application
 *
 * Provides a configurable logger with support for different verbosity levels
 * and consistent formatting of messages.
 */
import chalk from 'chalk';
import { symbols } from './consoleUtils';

/**
 * Verbosity levels for the logger
 */
export enum LogLevel {
  ERROR = 0, // Critical errors only
  WARN = 1, // Warnings and errors
  INFO = 2, // Normal operational information
  DEBUG = 3, // More detailed information
  VERBOSE = 4, // Everything including detailed debugging information
}

/**
 * Options for configuring the logger
 */
export interface LoggerOptions {
  level?: LogLevel;
  useColors?: boolean;
  timestamp?: boolean;
  prefix?: string;
}

/**
 * Default options for the logger
 */
const DEFAULT_OPTIONS: LoggerOptions = {
  level: LogLevel.INFO,
  useColors: true,
  timestamp: false,
  prefix: '',
};

/**
 * Logger class for managing application logging
 */
export class Logger {
  private options: Required<LoggerOptions>;

  /**
   * Create a new logger instance
   * @param options - Configuration options for the logger
   */
  constructor(options: LoggerOptions = {}) {
    this.options = {
      ...DEFAULT_OPTIONS,
      ...options,
    } as Required<LoggerOptions>;
  }

  /**
   * Set the verbosity level of the logger
   * @param level - The log level to set
   */
  public setLevel(level: LogLevel): void {
    this.options.level = level;
  }

  /**
   * Get the current verbosity level
   * @returns The current log level
   */
  public getLevel(): LogLevel {
    return this.options.level;
  }

  /**
   * Enable or disable colored output
   * @param useColors - Whether to use colors in log output
   */
  public setColors(useColors: boolean): void {
    this.options.useColors = useColors;
  }

  /**
   * Set a prefix for all log messages
   * @param prefix - The prefix to add to all messages
   */
  public setPrefix(prefix: string): void {
    this.options.prefix = prefix;
  }

  /**
   * Enable or disable timestamps in log messages
   * @param enable - Whether to include timestamps
   */
  public setTimestamp(enable: boolean): void {
    this.options.timestamp = enable;
  }

  /**
   * Update logger options
   * @param options - New options to apply
   */
  public configure(options: LoggerOptions): void {
    this.options = {
      ...this.options,
      ...options,
    };
  }

  /**
   * Log an error message (highest severity)
   * @param message - The message to log
   * @param error - Optional error object to include details from
   */
  public error(message: string, error?: Error): void {
    if (this.options.level >= LogLevel.ERROR) {
      const formattedMessage = this.format(message, 'error');

      // eslint-disable-next-line no-console
      console.error(formattedMessage);

      // If an error object is provided, log its details
      if (error) {
        const details = error.stack || error.message;
        // eslint-disable-next-line no-console
        console.error(this.options.useColors ? chalk.red(details) : details);
      }
    }
  }

  /**
   * Log a warning message
   * @param message - The message to log
   */
  public warn(message: string): void {
    if (this.options.level >= LogLevel.WARN) {
      const formattedMessage = this.format(message, 'warn');
      // eslint-disable-next-line no-console
      console.warn(formattedMessage);
    }
  }

  /**
   * Log an informational message (normal priority)
   * @param message - The message to log
   */
  public info(message: string): void {
    if (this.options.level >= LogLevel.INFO) {
      const formattedMessage = this.format(message, 'info');
      // eslint-disable-next-line no-console
      console.log(formattedMessage);
    }
  }

  /**
   * Log a success message (info level but styled as success)
   * @param message - The message to log
   */
  public success(message: string): void {
    if (this.options.level >= LogLevel.INFO) {
      const formattedMessage = this.format(message, 'success');
      // eslint-disable-next-line no-console
      console.log(formattedMessage);
    }
  }

  /**
   * Log a debug message (lower priority)
   * @param message - The message to log
   */
  public debug(message: string): void {
    if (this.options.level >= LogLevel.DEBUG) {
      const formattedMessage = this.format(message, 'debug');
      // eslint-disable-next-line no-console
      console.log(formattedMessage);
    }
  }

  /**
   * Log a verbose message (lowest priority)
   * @param message - The message to log
   */
  public verbose(message: string): void {
    if (this.options.level >= LogLevel.VERBOSE) {
      const formattedMessage = this.format(message, 'verbose');
      // eslint-disable-next-line no-console
      console.log(formattedMessage);
    }
  }

  /**
   * Log a plain message without any formatting
   * @param message - The message to log
   */
  public plain(message: string): void {
    // Plain messages are always shown regardless of log level
    // eslint-disable-next-line no-console
    console.log(message);
  }

  /**
   * Format a log message based on its type and options
   * @param message - The message to format
   * @param type - The type of message (error, warn, info, etc.)
   * @returns Formatted message with appropriate styling
   */
  private format(
    message: string,
    type: 'error' | 'warn' | 'info' | 'success' | 'debug' | 'verbose'
  ): string {
    const { useColors, timestamp, prefix } = this.options;

    // Build message components
    const components: string[] = [];

    // Add timestamp if enabled
    if (timestamp) {
      const time = new Date().toISOString().split('T')[1].split('.')[0];
      components.push(useColors ? chalk.gray(`[${time}]`) : `[${time}]`);
    }

    // Add prefix if specified
    if (prefix) {
      components.push(useColors ? chalk.gray(prefix) : prefix);
    }

    // Add level indicator
    let indicator = '';
    switch (type) {
      case 'error':
        indicator = useColors ? chalk.red(symbols.cross) : symbols.cross;
        break;
      case 'warn':
        indicator = useColors ? chalk.yellow(symbols.warning) : symbols.warning;
        break;
      case 'info':
        indicator = useColors ? chalk.blue(symbols.info) : symbols.info;
        break;
      case 'success':
        indicator = useColors ? chalk.green(symbols.tick) : symbols.tick;
        break;
      case 'debug':
        indicator = useColors ? chalk.cyan('•') : '•';
        break;
      case 'verbose':
        indicator = useColors ? chalk.gray('›') : '>';
        break;
    }
    components.push(indicator);

    // Add the message with appropriate styling
    let styledMessage = message;
    if (useColors) {
      switch (type) {
        case 'error':
          styledMessage = chalk.red(message);
          break;
        case 'warn':
          styledMessage = chalk.yellow(message);
          break;
        case 'info':
          styledMessage = message; // No color for info
          break;
        case 'success':
          styledMessage = chalk.green(message);
          break;
        case 'debug':
          styledMessage = chalk.cyan(message);
          break;
        case 'verbose':
          styledMessage = chalk.gray(message);
          break;
      }
    }
    components.push(styledMessage);

    return components.join(' ');
  }
}

/**
 * Create a default logger instance for the application
 */
export const logger = new Logger();

/**
 * Configure the global logger based on command-line options
 *
 * @param options - Options to control logger configuration
 * @param options.verbose - Whether verbose mode is enabled
 * @param options.quiet - Whether quiet mode is enabled
 * @param options.debug - Whether debug mode is enabled
 * @param options.noColor - Whether to disable colors
 */
export function configureLogger(options: {
  verbose?: boolean;
  quiet?: boolean;
  debug?: boolean;
  noColor?: boolean;
}): void {
  // Determine the log level based on flags
  let level = LogLevel.INFO; // Default level

  if (options.quiet) {
    level = LogLevel.ERROR; // Only show errors
  } else if (options.debug) {
    level = LogLevel.DEBUG; // Show debug information
  } else if (options.verbose) {
    level = LogLevel.VERBOSE; // Show everything
  }

  // Update the logger configuration
  logger.configure({
    level,
    useColors: !options.noColor,
  });
}

// Export module for easy access
export default logger;
