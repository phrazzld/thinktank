/**
 * Centralized error handling system for thinktank
 * 
 * This module defines a consistent error hierarchy and error handling utilities
 * to improve error reporting and troubleshooting across the application.
 */

import { colors } from '../utils/consoleUtils';

/**
 * Error categories for consistent categorization across the application
 */
export const errorCategories = {
  API: 'API',
  CONFIG: 'Configuration',
  NETWORK: 'Network',
  FILESYSTEM: 'File System',
  PERMISSION: 'Permission',
  VALIDATION: 'Validation',
  INPUT: 'Input',
  UNKNOWN: 'Unknown',
};

/**
 * Base error class for all thinktank errors
 * Provides a consistent structure for error reporting
 */
export class ThinktankError extends Error {
  /**
   * Error category for grouping similar errors
   */
  category: string = errorCategories.UNKNOWN;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands or configurations
   */
  examples?: string[];

  /**
   * The original error that caused this error
   */
  cause?: Error;
  
  constructor(message: string, options?: {
    cause?: Error;
    category?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message);
    this.name = 'ThinktankError';
    
    if (options?.cause) {
      this.cause = options.cause;
    }
    
    if (options?.category) {
      this.category = options.category;
    }
    
    if (options?.suggestions) {
      this.suggestions = options.suggestions;
    }
    
    if (options?.examples) {
      this.examples = options.examples;
    }
  }
  
  /**
   * Formats the error for display
   */
  format(): string {
    let output = `${colors.red.bold('Error')} (${colors.yellow(this.category)}): ${this.message}`;
    
    if (this.suggestions?.length) {
      output += '\n\nSuggestions:';
      this.suggestions.forEach(suggestion => {
        output += `\n  • ${suggestion}`;
      });
    }
    
    if (this.examples?.length) {
      output += '\n\nExamples:';
      this.examples.forEach(example => {
        output += `\n  - ${example}`;
      });
    }
    
    return output;
  }
}

/**
 * Configuration-related errors
 */
export class ConfigError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.CONFIG,
    });
    this.name = 'ConfigError';
  }
}

/**
 * API interaction errors
 */
export class ApiError extends ThinktankError {
  providerId?: string;
  
  constructor(message: string, options?: {
    cause?: Error;
    providerId?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    const formattedMessage = options?.providerId 
      ? `[${options.providerId}] ${message}`
      : message;
    
    super(formattedMessage, {
      ...options,
      category: errorCategories.API,
    });
    
    this.name = 'ApiError';
    this.providerId = options?.providerId;
  }
}

/**
 * File system related errors
 */
export class FileSystemError extends ThinktankError {
  filePath?: string;
  
  constructor(message: string, options?: {
    cause?: Error;
    filePath?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.FILESYSTEM,
    });
    
    this.name = 'FileSystemError';
    this.filePath = options?.filePath;
  }
}

/**
 * Input validation errors
 */
export class ValidationError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.VALIDATION,
    });
    this.name = 'ValidationError';
  }
}

/**
 * Network-related errors
 */
export class NetworkError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.NETWORK,
    });
    this.name = 'NetworkError';
  }
}

/**
 * Permission-related errors
 */
export class PermissionError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.PERMISSION,
    });
    this.name = 'PermissionError';
  }
}

/**
 * Input processing errors
 */
export class InputError extends ThinktankError {
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.INPUT,
    });
    this.name = 'InputError';
  }
}