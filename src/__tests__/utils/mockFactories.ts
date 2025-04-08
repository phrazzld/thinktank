/**
 * Mock factories to generate consistent mocks for testing
 * 
 * This module provides factory functions for creating consistent mock objects
 * that can be used across test files. It helps ensure that mock implementations
 * are consistent and properly typed.
 */

import { ContextFileResult } from '../../utils/fileReaderTypes';

/**
 * Creates a consistent set of mock functions for the fileReader module
 * 
 * @returns Object containing mock functions matching the fileReader module interface
 */
export const createFileReaderMocks = (): ReturnType<typeof jest.fn> & { [key: string]: any } => {
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
