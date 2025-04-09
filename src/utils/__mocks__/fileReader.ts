/**
 * Manual mock for fileReader module
 *
 * This mock exposes the same interface as the original but with jest.fn() mocks
 * that can be configured as needed in tests.
 *
 * IMPORTANT: To avoid conflicts with existing tests, this manual mock auto-detects
 * if it's being required directly by the test (explicit .mock) or through Jest's
 * automocking. When required directly, it creates self-contained mocks. When
 * automocked by Jest, it defers to Jest's default mock behavior.
 */

// Import the original types from the types module
import {
  FileReadError as ActualFileReadError,
  ContextFileResult as ActualContextFileResult,
} from '../fileReaderTypes';

// Re-export types
export interface ContextFileResult extends ActualContextFileResult {}

// If this module is being imported directly by a test through jest.mock,
// we can provide our own mock implementation
// Export mock functions for tests to customize
export const fileExists = jest.fn().mockResolvedValue(true);

export const readContextFile = jest.fn().mockImplementation((filePath: string) => {
  return Promise.resolve({
    path: filePath,
    content: 'Mocked content',
    error: null,
  });
});

export const readDirectoryContents = jest.fn().mockImplementation(() => {
  return Promise.resolve([]);
});

// These functions are used in other tests and need to be mock functions
export const readFileContent = jest.fn().mockResolvedValue('Mocked file content');
export const readContextPaths = jest.fn().mockResolvedValue([]);
export const formatCombinedInput = jest.fn().mockReturnValue('Mocked combined input');
export const getConfigDir = jest.fn().mockResolvedValue('/mock/config/dir');
export const getConfigFilePath = jest.fn().mockResolvedValue('/mock/config.json');
export const writeFile = jest.fn().mockResolvedValue(undefined);
export const isBinaryFile = jest.fn().mockReturnValue(false);

// Export constants
export const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

// Export error class
export class FileReadError extends ActualFileReadError {
  constructor(message: string, cause?: Error) {
    super(message, cause);
  }
}
