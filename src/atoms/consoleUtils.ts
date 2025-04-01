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

// Define commonly used Unicode symbols
export const symbols = {
  tick: '✓',
  cross: '✖',
  warning: '⚠',
  info: 'ℹ',
  pointer: '❯',
  line: '─',
  bullet: '•',
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
  return styleDim('─'.repeat(length));
}

/**
 * Categories of errors for consistent categorization
 */
export const errorCategories = {
  API: 'API',
  CONFIG: 'Configuration',
  NETWORK: 'Network',
  FILESYSTEM: 'File System',
  PERMISSION: 'Permission',
  VALIDATION: 'Validation',
  UNKNOWN: 'Unknown',
};

/**
 * Formats an error message with consistent styling
 * @param error - The error object or message
 * @param category - Optional error category
 * @param tip - Optional troubleshooting tip
 * @returns Formatted error message
 */
export function formatError(
  error: Error | string, 
  category: string = errorCategories.UNKNOWN,
  tip?: string
): string {
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
 */
export function categorizeError(error: Error | string): string {
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
 */
export function getTroubleshootingTip(error: Error | string, category: string): string | undefined {
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
 */
export function formatErrorWithTip(error: Error | string): string {
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
 */
export function createFileNotFoundError(filePath: string, errorMessage?: string): Error {
  // Import path if available, otherwise use a simpler approach
  let path: any;
  try {
    path = require('path');
  } catch {
    // No path module, use simpler approach
    path = {
      isAbsolute: (p: string) => p.startsWith('/'),
      dirname: (p: string) => {
        const parts = p.split('/');
        return parts.slice(0, -1).join('/') || '.';
      },
      basename: (p: string) => {
        const parts = p.split('/');
        return parts[parts.length - 1];
      },
      join: (...parts: string[]) => parts.join('/')
    };
  }

  // Get current working directory
  const currentDir = process.cwd();

  // Create the error with a default message if none provided
  const message = errorMessage || `File not found: ${filePath}`;
  
  // This requires the ThinktankError class which we can't import here
  // to avoid circular dependencies, so we'll create a regular Error
  // and add the properties the consumer can convert it if needed
  const error = new Error(message);
  
  // Add metadata
  (error as any).category = errorCategories.FILESYSTEM;
  
  // Extract path components
  const isAbsolutePath = path.isAbsolute(filePath);
  const dirname = path.dirname(filePath);
  const basename = path.basename(filePath);
  
  // Build suggestions
  const suggestions = [
    `Check that the file exists at the specified path: ${isAbsolutePath ? filePath : path.join(currentDir, filePath)}`,
    `Current working directory: ${currentDir}`
  ];
  
  // Add path-specific suggestions
  if (!isAbsolutePath && dirname !== '.') {
    suggestions.push(`Ensure the directory exists: ${path.join(currentDir, dirname)}`);
  }
  
  // Add common filename pattern suggestions
  if (!basename.includes('.')) {
    suggestions.push(`The file may need an extension (e.g., ${basename}.txt, ${basename}.md)`);
  }
  
  // Add general suggestions
  suggestions.push(
    `Use a relative path (./path/to/file.txt) or absolute path (/full/path/to/file.txt)`,
    `Make sure the file has read permissions`
  );
  
  (error as any).suggestions = suggestions;
  
  // Add examples
  (error as any).examples = [
    `path/to/${basename}.txt`,
    `./path/to/${basename}.txt`,
    `${path.join(currentDir, basename)}.txt`
  ];
  
  return error;
}