/**
 * CLI testing setup utilities
 * 
 * This module provides specialized setup helpers for CLI-related tests.
 */
import path from 'path';
import { setupBasicFs } from './fs';
import { createVirtualConfigFile, createMinimalConfig } from './config';
import { normalizePathGeneral } from '../../src/utils/pathUtils';
import type { AppConfig } from '../../src/core/types';

/**
 * Sets up a mock CLI environment with standard files
 * 
 * @param baseDir - Base directory for test files
 * @param config - Optional custom configuration
 * @returns Object containing paths to created test files
 * 
 * Usage:
 * ```typescript
 * const { promptFile, configPath } = setupCliTest('/test');
 * ```
 */
export function setupCliTest(
  baseDir: string = '/test',
  config: AppConfig = createMinimalConfig()
): { 
  promptFile: string; 
  contextFile: string; 
  configPath: string 
} {
  const normalizedBaseDir = normalizePathGeneral(baseDir, true);
  
  // Create standard file paths
  const promptFile = path.join(normalizedBaseDir, 'prompt.txt');
  const contextFile = path.join(normalizedBaseDir, 'context.txt');
  const configPath = path.join(normalizedBaseDir, 'thinktank.config.json');
  
  // Create basic test files
  setupBasicFs({
    [promptFile]: 'This is a test prompt',
    [contextFile]: 'Test context data'
  });
  
  // Create config file
  createVirtualConfigFile(configPath, config);
  
  return { promptFile, contextFile, configPath };
}

/**
 * Creates a mock for process.argv
 * 
 * @param command - The CLI command to mock (e.g., 'run', 'config', 'models')
 * @param args - Additional command-specific arguments
 * @returns Function to restore the original process.argv
 * 
 * Usage:
 * ```typescript
 * const restore = mockCliArguments('run', ['prompt.txt', '--model', 'openai:gpt-4o']);
 * // Test with mocked CLI arguments
 * restore(); // Restore original process.argv
 * ```
 */
export function mockCliArguments(command: string, args: string[] = []): () => void {
  const originalArgv = process.argv;
  
  // Mock process.argv with [node, script, command, ...args]
  process.argv = ['node', 'thinktank', command, ...args];
  
  // Return function to restore original argv
  return () => {
    process.argv = originalArgv;
  };
}

/**
 * Sets up mocks for console output to capture CLI output
 * 
 * @returns Object containing mock functions and a restore function
 * 
 * Usage:
 * ```typescript
 * const { mockLog, mockError, restore } = mockConsoleOutput();
 * // Run CLI command
 * expect(mockLog).toHaveBeenCalledWith('Success message');
 * restore();
 * ```
 */
export function mockConsoleOutput(): {
  mockLog: jest.Mock;
  mockError: jest.Mock;
  mockInfo: jest.Mock;
  mockWarn: jest.Mock;
  restore: () => void;
} {
  // eslint-disable-next-line no-console
  const originalLog = console.log;
  // eslint-disable-next-line no-console
  const originalError = console.error;
  // eslint-disable-next-line no-console
  const originalInfo = console.info;
  // eslint-disable-next-line no-console
  const originalWarn = console.warn;
  
  const mockLog = jest.fn();
  const mockError = jest.fn();
  const mockInfo = jest.fn();
  const mockWarn = jest.fn();
  
  // eslint-disable-next-line no-console
  console.log = mockLog;
  // eslint-disable-next-line no-console
  console.error = mockError;
  // eslint-disable-next-line no-console
  console.info = mockInfo;
  // eslint-disable-next-line no-console
  console.warn = mockWarn;
  
  return {
    mockLog,
    mockError,
    mockInfo,
    mockWarn,
    restore: () => {
      // eslint-disable-next-line no-console
      console.log = originalLog;
      // eslint-disable-next-line no-console
      console.error = originalError;
      // eslint-disable-next-line no-console
      console.info = originalInfo;
      // eslint-disable-next-line no-console
      console.warn = originalWarn;
    }
  };
}

/**
 * Creates a mock for the CLI spinner
 * 
 * @returns Mock spinner object with jest.fn() methods
 * 
 * Usage:
 * ```typescript
 * const spinner = createMockSpinner();
 * // Mock implementation
 * jest.mock('ora', () => () => spinner);
 * // Use in tests
 * expect(spinner.start).toHaveBeenCalled();
 * ```
 */
export function createMockSpinner(): {
  start: jest.Mock;
  stop: jest.Mock;
  succeed: jest.Mock;
  fail: jest.Mock;
  warn: jest.Mock;
  info: jest.Mock;
  text: string;
} {
  return {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: ''
  };
}
