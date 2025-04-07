/**
 * Unit tests for the _processOutput helper function
 */
import { _processOutput } from '../runThinktankHelpers';
import * as outputHandler from '../outputHandler';
import { FileSystemError, PermissionError, ThinktankError } from '../../core/errors';
import { OutputHandlerError } from '../outputHandler';
import { ModelQueryStatus } from '../queryExecutor';

// Mock dependencies
jest.mock('../outputHandler');

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_processOutput Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  // Sample query results for testing
  const sampleQueryResults: any = {
    responses: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        text: 'Mock response',
        configKey: 'mock:mock-model',
        metadata: {}
      },
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        text: 'OpenAI response',
        configKey: 'openai:gpt-4o',
        metadata: {}
      }
    ],
    statuses: {
      'mock:mock-model': {
        status: 'success' as ModelQueryStatus['status'],
        startTime: 1,
        endTime: 2,
        durationMs: 1
      },
      'openai:gpt-4o': {
        status: 'success' as ModelQueryStatus['status'],
        startTime: 1,
        endTime: 3,
        durationMs: 2
      }
    },
    timing: {
      startTime: 1,
      endTime: 3,
      durationMs: 2
    }
  };

  it('should successfully process output', async () => {
    // Setup mocks
    (outputHandler.processOutput as jest.Mock).mockResolvedValue({
      fileOutput: {
        outputDirectory: '/path/to/output/dir',
        files: [
          { modelKey: 'mock:mock-model', filename: 'mock-model.md', status: 'success' },
          { modelKey: 'openai:gpt-4o', filename: 'openai-gpt4o.md', status: 'success' }
        ],
        succeededWrites: 2,
        failedWrites: 0,
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      },
      consoleOutput: 'Formatted console output'
    });

    // Call the function
    const result = await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt',
        includeMetadata: true
      },
      friendlyRunName: 'clever-meadow'
    });

    // Verify the result
    expect(result.fileOutputResult.succeededWrites).toBe(2);
    expect(result.fileOutputResult.failedWrites).toBe(0);
    expect(result.consoleOutput).toBe('Formatted console output');

    // Verify mocks were called correctly
    expect(outputHandler.processOutput).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          configKey: 'mock:mock-model'
        }),
        expect.objectContaining({
          configKey: 'openai:gpt-4o'
        })
      ]),
      expect.objectContaining({
        outputDirectory: '/path/to/output/dir',
        friendlyRunName: 'clever-meadow',
        includeMetadata: true
      })
    );

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Output processing complete');
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.start).toHaveBeenCalled();
  });

  it('should handle status updates during file writing', async () => {
    // Setup mock with side effects to call onStatusUpdate
    (outputHandler.processOutput as jest.Mock).mockImplementation((_responses, options) => {
      if (options.onStatusUpdate) {
        // Pending state
        options.onStatusUpdate({
          modelKey: 'mock:mock-model',
          filename: 'mock-model.md',
          status: 'pending'
        }, []);
        
        // Success state
        options.onStatusUpdate({
          modelKey: 'mock:mock-model',
          filename: 'mock-model.md',
          status: 'success'
        }, []);
      }

      return Promise.resolve({
        fileOutput: {
          outputDirectory: '/path/to/output/dir',
          files: [
            { modelKey: 'mock:mock-model', filename: 'mock-model.md', status: 'success' }
          ],
          succeededWrites: 1,
          failedWrites: 0,
          timing: { startTime: 1, endTime: 2, durationMs: 1 }
        },
        consoleOutput: 'Console output after status updates'
      });
    });

    // Call the function
    await _processOutput({
      spinner: mockSpinner,
      queryResults: {
        ...sampleQueryResults,
        responses: [sampleQueryResults.responses[0]]
      },
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    });

    // Verify spinner text updates from status updates
    expect(mockSpinner.text).toContain('Output processing complete');
  });

  it('should handle partial output success with some failures', async () => {
    // Setup mocks with mixture of success and failure
    (outputHandler.processOutput as jest.Mock).mockResolvedValue({
      fileOutput: {
        outputDirectory: '/path/to/output/dir',
        files: [
          { modelKey: 'mock:mock-model', filename: 'mock-model.md', status: 'success' },
          { 
            modelKey: 'openai:gpt-4o', 
            filename: 'openai-gpt4o.md', 
            status: 'error',
            error: 'Write failed: permission denied'
          }
        ],
        succeededWrites: 1,
        failedWrites: 1,
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      },
      consoleOutput: 'Formatted console output with warnings'
    });

    // Call the function
    const result = await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    });

    // Verify the result reflects partial success
    expect(result.fileOutputResult.succeededWrites).toBe(1);
    expect(result.fileOutputResult.failedWrites).toBe(1);

    // Verify spinner shows appropriate warnings
    expect(mockSpinner.warn).toHaveBeenCalled();
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.text).toContain('Output processing complete');
  });

  it('should handle all files failing', async () => {
    // Setup mocks with all failures
    (outputHandler.processOutput as jest.Mock).mockResolvedValue({
      fileOutput: {
        outputDirectory: '/path/to/output/dir',
        files: [
          { 
            modelKey: 'mock:mock-model', 
            filename: 'mock-model.md', 
            status: 'error',
            error: 'Disk full'
          },
          { 
            modelKey: 'openai:gpt-4o', 
            filename: 'openai-gpt4o.md', 
            status: 'error',
            error: 'Permission denied'
          }
        ],
        succeededWrites: 0,
        failedWrites: 2,
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      },
      consoleOutput: 'Console output with errors'
    });

    // Call the function
    const result = await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    });

    // Verify the result shows all failures
    expect(result.fileOutputResult.succeededWrites).toBe(0);
    expect(result.fileOutputResult.failedWrites).toBe(2);

    // Verify spinner shows appropriate warnings and error summary
    expect(mockSpinner.warn).toHaveBeenCalled();
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.text).toContain('Output processing complete');
  });

  it('should handle OutputHandlerError by wrapping in FileSystemError', async () => {
    // Setup mocks
    const outputError = new OutputHandlerError('Failed to write output files');
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(outputError);

    // Call the function and expect it to throw
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(FileSystemError);

    // Verify proper error wrapping
    try {
      await _processOutput({
        spinner: mockSpinner,
        queryResults: sampleQueryResults,
        outputDirectoryPath: '/path/to/output/dir',
        options: {
          input: 'test-prompt.txt'
        },
        friendlyRunName: 'clever-meadow'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toBe('Failed to write output files');
        expect(error.cause).toBe(outputError);
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
      }
    }
  });

  it('should handle existing FileSystemError by rethrowing', async () => {
    // Setup mocks
    const fsError = new FileSystemError('Directory not writable');
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(fsError);

    // Call the function and expect it to throw the original error
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(fsError);
  });

  it('should handle existing PermissionError by rethrowing', async () => {
    // Setup mocks
    const permError = new PermissionError('Permission denied');
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(permError);

    // Call the function and expect it to throw the original error
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(permError);
  });

  it('should handle EACCES NodeJS errors', async () => {
    // Setup mocks
    const nodeError = new Error('Permission denied') as NodeJS.ErrnoException;
    nodeError.code = 'EACCES';
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(nodeError);

    // Call the function and expect it to throw PermissionError
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(PermissionError);

    // Verify proper error transformation
    try {
      await _processOutput({
        spinner: mockSpinner,
        queryResults: sampleQueryResults,
        outputDirectoryPath: '/path/to/output/dir',
        options: {
          input: 'test-prompt.txt'
        },
        friendlyRunName: 'clever-meadow'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(PermissionError);
      if (error instanceof PermissionError) {
        expect(error.message).toContain('Permission denied when writing output');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.some(s => s.includes('permissions'))).toBeTruthy();
      }
    }
  });

  it('should handle ENOENT NodeJS errors', async () => {
    // Setup mocks
    const nodeError = new Error('No such file or directory') as NodeJS.ErrnoException;
    nodeError.code = 'ENOENT';
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(nodeError);

    // Call the function and expect it to throw FileSystemError
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/nonexistent/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(FileSystemError);

    // Verify proper error transformation
    try {
      await _processOutput({
        spinner: mockSpinner,
        queryResults: sampleQueryResults,
        outputDirectoryPath: '/nonexistent/dir',
        options: {
          input: 'test-prompt.txt'
        },
        friendlyRunName: 'clever-meadow'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('Directory not found');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.some(s => s.includes('output directory'))).toBeTruthy();
      }
    }
  });

  it('should handle ENOSPC NodeJS errors', async () => {
    // Setup mocks
    const nodeError = new Error('No space left on device') as NodeJS.ErrnoException;
    nodeError.code = 'ENOSPC';
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(nodeError);

    // Call the function and expect it to throw FileSystemError
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow(FileSystemError);

    // Verify error has appropriate message and suggestions
    try {
      await _processOutput({
        spinner: mockSpinner,
        queryResults: sampleQueryResults,
        outputDirectoryPath: '/path/to/output/dir',
        options: {
          input: 'test-prompt.txt'
        },
        friendlyRunName: 'clever-meadow'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('No space left on device');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.some(s => s.includes('disk space'))).toBeTruthy();
      }
    }
  });

  it('should handle unknown errors by wrapping them in ThinktankError', async () => {
    // Setup mocks
    const unknownError = new Error('Something unexpected happened');
    (outputHandler.processOutput as jest.Mock).mockRejectedValue(unknownError);

    // Call the function and expect it to throw
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      outputDirectoryPath: '/path/to/output/dir',
      options: {
        input: 'test-prompt.txt'
      },
      friendlyRunName: 'clever-meadow'
    })).rejects.toThrow();

    // Verify proper error wrapping
    try {
      await _processOutput({
        spinner: mockSpinner,
        queryResults: sampleQueryResults,
        outputDirectoryPath: '/path/to/output/dir',
        options: {
          input: 'test-prompt.txt'
        },
        friendlyRunName: 'clever-meadow'
      });
    } catch (error) {
      expect((error as ThinktankError).message).toContain('Error processing output');
      expect((error as ThinktankError).cause).toBe(unknownError);
      expect((error as ThinktankError).suggestions).toBeDefined();
      expect((error as ThinktankError).suggestions!.length).toBeGreaterThan(0);
    }
  });
});
