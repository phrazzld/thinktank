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

// The DetailedError interface has been removed
// Use ThinktankError from '../core/errors' instead

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
 * Styles text as an info message
 * @param text - The text to style
 * @returns Styled text with an info indicator
 */
export function styleInfo(text: string): string {
  return `${colors.blue(symbols.info)} ${text}`;
}

/**
 * Creates a styled section header
 * @param title - The section title
 * @returns Styled section header
 */
export function styleSectionHeader(title: string): string {
  return colors.bold.underline(`\n${title}`);
}

/**
 * Creates a styled header
 * @param text - The text to style
 * @returns Styled text as a header
 */
export function styleHeader(text: string): string {
  return colors.bold.blue(text);
}

/**
 * Dims text for less prominence
 * @param text - The text to style
 * @returns Dimmed text
 */
export function styleDim(text: string): string {
  return colors.dim(text);
}

/**
 * Creates a divider line
 * @param length - The length of the divider (defaults to 80)
 * @returns A horizontal line of specified length
 */
export function divider(length = 80): string {
  return symbols.line.repeat(length);
}

/*
 * Error-related functions have been moved to the core error system:
 * - For error categories: use 'errorCategories' from '../core/errors'
 * - For error formatting: use the 'format()' method on ThinktankError instances
 * - For error categorization: use 'categorizeError' from '../core/errors/utils/categorization'
 * - For file errors: use 'createFileNotFoundError' from '../core/errors'
 * - For model format errors: use 'createModelFormatError' from '../core/errors'
 * - For model not found errors: use 'createModelNotFoundError' from '../core/errors'
 * - For missing API key errors: use 'createMissingApiKeyError' from '../core/errors'
 */
