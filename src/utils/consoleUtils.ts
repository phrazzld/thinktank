/**
 * Console utility module for terminal styling and formatting
 * 
 * Centralizes all terminal styling logic to maintain consistency
 * and provide reusable formatting helpers across the application.
 */

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-return */
/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
import chalk from 'chalk';

// Re-export our configured chalk instance
export const colors = chalk;

/**
 * Extended Error interface with additional metadata for better error handling and display
 */
export interface DetailedError extends Error {
  category?: string;
  suggestions?: string[];
  examples?: string[];
}

// Import the error system from core/errors
import {
  ThinktankError,
  errorCategories as coreErrorCategories,
  createFileNotFoundError as coreCreateFileNotFoundError,
  createModelFormatError as coreCreateModelFormatError,
  createMissingApiKeyError as coreCreateMissingApiKeyError,
  createModelNotFoundError as coreCreateModelNotFoundError
} from '../core/errors';

// Re-export error categories (for backward compatibility)
export const errorCategories = coreErrorCategories;

// Define commonly used Unicode symbols (no emojis)
export const symbols = {
  tick: '+',
  cross: 'x',
  warning: '!',
  info: 'i',
  pointer: '>',
  line: '-',
  bullet: '*',
};

/**
 * Styles text as a success message
 * @param text - The text to style
 * @returns Styled text with a success indicator
 */
export function styleSuccess(text: string): string {
  return `${colors.green(symbols.tick)} ${text}`;
}

/**
 * Styles text as an error message
 * @param text - The text to style
 * @returns Styled text with an error indicator
 */
export function styleError(text: string): string {
  return `${colors.red(symbols.cross)} ${text}`;
}

/**
 * Styles text as a warning message
 * @param text - The text to style
 * @returns Styled text with a warning indicator
 */
export function styleWarning(text: string): string {
  return `${colors.yellow(symbols.warning)} ${text}`;
}

/**
 * Styles text as an informational message
 * @param text - The text to style
 * @returns Styled text with an info indicator
 */
export function styleInfo(text: string): string {
  return `${colors.blue(symbols.info)} ${text}`;
}

/**
 * Styles text as a header
 * @param text - The text to style
 * @returns Styled text as a header
 */
export function styleHeader(text: string): string {
  return colors.bold.blue(text);
}

/**
 * Styles text as dimmed/secondary content
 * @param text - The text to style
 * @returns Styled dimmed text
 */
export function styleDim(text: string): string {
  return colors.dim(text);
}

/**
 * A divider line for visual separation
 * @param length - The length of the divider line
 * @returns A styled divider line
 */
export function divider(length = 80): string {
  return styleDim(symbols.line.repeat(length));
}

/**
 * Categories of errors for consistent categorization
 * @deprecated Import from src/core/errors instead
 */
// Already imported and re-exported above

/**
 * Formats an error message with consistent styling
 * @param error - The error object or message
 * @param category - Optional error category
 * @param tip - Optional troubleshooting tip
 * @returns Formatted error message
 * @deprecated Use ThinktankError.format() directly for ThinktankError instances
 */
export function formatError(
  error: Error | string, 
  category: string = errorCategories.UNKNOWN,
  tip?: string
): string {
  // If it's a ThinktankError, use its built-in format method
  if (error instanceof ThinktankError) {
    // Return the formatted error with tip if provided and not already in suggestions
    const formatted = error.format();
    if (tip && (!error.suggestions || !error.suggestions.includes(tip))) {
      return `${formatted}\n  ${colors.cyan(symbols.info)} Tip: ${tip}`;
    }
    return formatted;
  }
  
  // Otherwise, continue with the existing formatting logic
  const errorMsg = error instanceof Error ? error.message : error;
  let output = `${colors.red.bold('Error')}${category ? ` (${colors.yellow(category)})` : ''}: ${errorMsg}`;
  
  if (tip) {
    output += `\n  ${colors.cyan(symbols.info)} Tip: ${tip}`;
  }
  
  return output;
}

/**
 * Tries to categorize an error based on its message or type
 * @param error - The error to categorize
 * @returns The error category
 * @deprecated Use direct access to ThinktankError.category for ThinktankError instances
 */
export function categorizeError(error: Error | string): string {
  // If it's a ThinktankError, use its category
  if (error instanceof ThinktankError) {
    return error.category;
  }
  
  // Otherwise, use the existing categorization logic
  const message = error instanceof Error ? error.message : error;
  const lowerMsg = message.toLowerCase();
  
  if (lowerMsg.includes('api key') || lowerMsg.includes('authentication') || 
      lowerMsg.includes('auth') || lowerMsg.includes('401') || lowerMsg.includes('403')) {
    return errorCategories.API;
  }
  
  if (lowerMsg.includes('econnrefused') || lowerMsg.includes('etimedout') || 
      lowerMsg.includes('enotfound') || lowerMsg.includes('network')) {
    return errorCategories.NETWORK;
  }
  
  if (lowerMsg.includes('config') || lowerMsg.includes('settings')) {
    return errorCategories.CONFIG;
  }
  
  if (lowerMsg.includes('enoent') || lowerMsg.includes('file not found') || 
      lowerMsg.includes('directory') || lowerMsg.includes('path')) {
    return errorCategories.FILESYSTEM;
  }
  
  if (lowerMsg.includes('permission') || lowerMsg.includes('access denied') ||
      lowerMsg.includes('eacces')) {
    return errorCategories.PERMISSION;
  }
  
  if (lowerMsg.includes('validation') || lowerMsg.includes('invalid') || 
      lowerMsg.includes('schema') || lowerMsg.includes('required')) {
    return errorCategories.VALIDATION;
  }
  
  return errorCategories.UNKNOWN;
}

/**
 * Gets a troubleshooting tip based on the error category
 * @param error - The error message or object
 * @param category - The error category
 * @returns A helpful tip or undefined if none available
 * @deprecated Use ThinktankError.suggestions for ThinktankError instances
 */
export function getTroubleshootingTip(error: Error | string, category: string): string | undefined {
  // If it's a ThinktankError with suggestions, use the first suggestion
  if (error instanceof ThinktankError && error.suggestions && error.suggestions.length > 0) {
    return error.suggestions[0];
  }
  
  // Otherwise, use the existing logic
  const message = error instanceof Error ? error.message : error;
  const lowerMsg = message.toLowerCase();
  
  switch(category) {
    case errorCategories.API:
      if (lowerMsg.includes('api key')) {
        return 'Check your API key in your environment variables or config file';
      }
      if (lowerMsg.includes('rate limit') || lowerMsg.includes('429')) {
        return 'You\'ve hit the rate limit. Wait a while before trying again';
      }
      return 'Verify your API credentials and permissions';
      
    case errorCategories.NETWORK:
      return 'Check your internet connection and try again';
      
    case errorCategories.CONFIG:
      return 'Verify your thinktank.config.json file for errors';
      
    case errorCategories.FILESYSTEM:
      if (lowerMsg.includes('permission') || lowerMsg.includes('eacces')) {
        return 'Check file permissions or run with elevated privileges';
      }
      if (lowerMsg.includes('file not found') || lowerMsg.includes('enoent') || lowerMsg.includes('no such file')) {
        return 'Check that the file exists at the specified path and you have permission to read it';
      }
      return 'Verify the file path and ensure it exists';
      
    default:
      return undefined;
  }
}

/**
 * Formats an error with automatically determined category and tip
 * @param error - The error to format
 * @returns A formatted error message with category and tip
 * @deprecated Use ThinktankError.format() directly for ThinktankError instances
 */
export function formatErrorWithTip(error: Error | string): string {
  // If it's a ThinktankError, use its built-in format method directly
  if (error instanceof ThinktankError) {
    return error.format();
  }
  
  // Otherwise, use the existing logic
  const category = categorizeError(error);
  const tip = getTroubleshootingTip(error, category);
  return formatError(error, category, tip);
}

/**
 * Creates a detailed file not found error with helpful suggestions
 * 
 * @param filePath - The path to the file that wasn't found
 * @param errorMessage - Optional custom error message
 * @returns A ThinktankError with suggestions
 * @deprecated Import from src/core/errors instead
 */
export function createFileNotFoundError(filePath: string, errorMessage?: string): Error {
  // Use the imported function from core/errors
  return coreCreateFileNotFoundError(filePath, errorMessage);
}

/**
 * Handles common types of model format errors
 * 
 * @param modelSpecification - The problematic model specification
 * @param availableProviders - Optional array of available provider IDs
 * @param availableModels - Optional array of available model specifications
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions
 * @deprecated Import from src/core/errors instead
 */
export function createModelFormatError(
  modelSpecification: string,
  availableProviders: string[] = [],
  availableModels: string[] = [],
  errorMessage?: string
): Error {
  // Use the imported function from core/errors
  return coreCreateModelFormatError(
    modelSpecification,
    availableProviders,
    availableModels,
    errorMessage
  );
}

/**
 * Handles errors when a specific model is not found or unavailable
 * 
 * @param modelSpecification - The model specification that wasn't found
 * @param availableModels - Optional array of available model specifications
 * @param groupName - Optional group name if relevant to the context
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions
 */
/**
 * Creates a helpful error message for missing API keys
 * 
 * @param missingModels - Array of models with missing API keys
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions for setting API keys
 * @deprecated Import from src/core/errors instead
 */
export function createMissingApiKeyError(
  missingModels: Array<{ provider: string; modelId: string }>,
  errorMessage?: string
): Error {
  // Use the imported function from core/errors
  return coreCreateMissingApiKeyError(missingModels, errorMessage);
}

/**
 * Creates an error for model not found in configuration
 * 
 * @param modelSpecification - The model specification that wasn't found
 * @param availableModels - Optional array of available model specifications
 * @param groupName - Optional group name if relevant to the context
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with helpful suggestions
 * @deprecated Import from src/core/errors instead
 */
export function createModelNotFoundError(
  modelSpecification: string,
  availableModels: string[] = [],
  groupName?: string,
  errorMessage?: string
): Error {
  // Use the imported function from core/errors
  return coreCreateModelNotFoundError(
    modelSpecification,
    availableModels,
    groupName,
    errorMessage
  );
}