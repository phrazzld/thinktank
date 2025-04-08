/**
 * Mock factories to generate consistent mocks for testing
 * 
 * @deprecated This approach is deprecated. Please use the centralized mock setup from
 * jest/setupFiles/ for all new tests and when refactoring existing tests. See jest/README.md
 * for details on the preferred approach.
 * 
 * This module provides factory functions for creating consistent mock objects
 * that can be used across test files. It helps ensure that mock implementations
 * are consistent and properly typed.
 * 
 * Migration guide:
 * 
 * Before (using mockFactories - deprecated):
 * ```typescript
 * import { createFileReaderMocks } from '../mockFactories';
 * const mocks = createFileReaderMocks();
 * jest.spyOn(fileReader, 'readContextFile').mockImplementation(mocks.readContextFile);
 * ```
 * 
 * After (using centralized approach - preferred):
 * ```typescript
 * import { setupBasicFs } from '../../../jest/setupFiles/fs';
 * beforeEach(() => {
 *   setupBasicFs({
 *     '/test.txt': 'mocked content'
 *   });
 * });
 * ```
 */

import { ContextFileResult } from '../../utils/fileReaderTypes';

/**
 * Creates a consistent set of mock functions for the fileReader module
 * 
 * @deprecated Use setupBasicFs and other helpers from jest/setupFiles/fs instead
 * @returns Object containing mock functions matching the fileReader module interface
 */
export const createFileReaderMocks = (): { [key: string]: jest.Mock | number } => {
  const mockFileReadError = jest.fn().mockImplementation((message: string, cause?: Error) => {
    const error = new Error(message);
    Object.defineProperty(error, 'name', { value: 'FileReadError' });
    Object.defineProperty(error, 'cause', { value: cause });
    return error;
  });

  return {
    fileExists: jest.fn().mockResolvedValue(true),
    
    readContextFile: jest.fn().mockImplementation((filePath: string): Promise<ContextFileResult> => 
      Promise.resolve({
        path: filePath,
        content: 'Mocked content',
        error: null
      })
    ),
    
    readDirectoryContents: jest.fn().mockResolvedValue([]),
    
    readFileContent: jest.fn().mockResolvedValue('Mocked file content'),
    
    readContextPaths: jest.fn().mockResolvedValue([]),
    
    formatCombinedInput: jest.fn().mockReturnValue('Mocked combined input'),
    
    getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
    
    getConfigFilePath: jest.fn().mockResolvedValue('/mock/config.json'),
    
    writeFile: jest.fn().mockResolvedValue(undefined),
    
    isBinaryFile: jest.fn().mockReturnValue(false),
    
    FileReadError: mockFileReadError,
    
    MAX_FILE_SIZE: 10 * 1024 * 1024 // 10MB
  };
};
