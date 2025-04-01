/**
 * Tests for the console utilities module
 */
import * as consoleUtils from '../consoleUtils';

// Mock chalk and figures to test our styling without actually modifying strings
jest.mock('chalk', () => {
  const mockRed: any = jest.fn((text) => `red(${text})`);
  mockRed.bold = jest.fn((text) => `red.bold(${text})`);
  
  return {
    green: jest.fn((text) => `green(${text})`),
    yellow: jest.fn((text) => `yellow(${text})`),
    blue: jest.fn((text) => `blue(${text})`),
    cyan: jest.fn((text) => `cyan(${text})`),
    dim: jest.fn((text) => `dim(${text})`),
    bold: {
      blue: jest.fn((text) => `bold.blue(${text})`),
    },
    red: mockRed
  };
});

jest.mock('figures', () => {
  return {
    tick: '✓',
    cross: '✖',
    warning: '⚠',
    info: 'ℹ',
    pointer: '❯',
    line: '─',
    bullet: '●',
  };
});

describe('consoleUtils', () => {
  describe('styling functions', () => {
    test('styleSuccess should format text with green tick', () => {
      const result = consoleUtils.styleSuccess('Success message');
      expect(result).toBe('green(✓) Success message');
    });

    test('styleError should format text with red cross', () => {
      const result = consoleUtils.styleError('Error message');
      expect(result).toBe('red(✖) Error message');
    });

    test('styleWarning should format text with yellow warning', () => {
      const result = consoleUtils.styleWarning('Warning message');
      expect(result).toBe('yellow(⚠) Warning message');
    });

    test('styleInfo should format text with blue info', () => {
      const result = consoleUtils.styleInfo('Info message');
      expect(result).toBe('blue(ℹ) Info message');
    });

    test('styleHeader should format text as bold blue', () => {
      const result = consoleUtils.styleHeader('Header');
      expect(result).toBe('bold.blue(Header)');
    });

    test('styleDim should format text as dimmed', () => {
      const result = consoleUtils.styleDim('Dimmed text');
      expect(result).toBe('dim(Dimmed text)');
    });

    test('divider should create a styled horizontal line', () => {
      const result = consoleUtils.divider(5);
      expect(result).toBe('dim(─────)');
    });
  });

  describe('error formatting', () => {
    test('formatError should format error with category and tip', () => {
      const result = consoleUtils.formatError(
        'Something went wrong', 
        consoleUtils.errorCategories.API, 
        'Check your API key'
      );
      expect(result).toContain('red.bold(Error)');
      expect(result).toContain('yellow(API)');
      expect(result).toContain('Something went wrong');
      expect(result).toContain('cyan(ℹ)');
      expect(result).toContain('Tip: Check your API key');
    });

    test('formatError should handle Error objects', () => {
      const error = new Error('Failed to connect');
      const result = consoleUtils.formatError(error);
      expect(result).toContain('Failed to connect');
    });

    test('categorizeError should detect API errors', () => {
      const error = new Error('Invalid API key provided');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.API);
    });

    test('categorizeError should detect network errors', () => {
      const error = new Error('ETIMEDOUT: Connection timed out');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.NETWORK);
    });

    test('categorizeError should return UNKNOWN for unrecognized errors', () => {
      const error = new Error('Some completely random error');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.UNKNOWN);
    });

    test('getTroubleshootingTip should return appropriate tips', () => {
      const apiError = new Error('Invalid API key');
      const tip = consoleUtils.getTroubleshootingTip(
        apiError, 
        consoleUtils.errorCategories.API
      );
      expect(tip).toContain('Check your API key');
    });

    test('formatErrorWithTip should automatically categorize and add tip', () => {
      const error = new Error('API key is invalid');
      const result = consoleUtils.formatErrorWithTip(error);
      expect(result).toContain('API');
      expect(result).toContain('Check your API key');
    });
  });
});