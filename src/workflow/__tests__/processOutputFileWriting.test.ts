/**
 * Integration tests for the file writing functionality in runThinktank.ts
 * 
 * These tests focus on verifying that the runThinktank module correctly processes
 * the data returned by _processOutput and writes files to disk.
 */
import { runThinktank, RunOptions } from '../runThinktank';
import * as helpers from '../runThinktankHelpers';
import { FileSystem } from '../../core/interfaces';
import { LLMResponse } from '../../core/types';
import { FileData, PureProcessOutputResult } from '../runThinktankTypes';

// Mocks
jest.mock('../../core/FileSystem');
jest.mock('../../core/ConcreteConfigManager');
jest.mock('../../core/LLMClient');
jest.mock('../runThinktankHelpers');

// Mock ora spinner
jest.mock('../../utils/spinnerFactory', () => {
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
  };
  
  return {
    __esModule: true,
    default: jest.fn(() => mockSpinner),
    configureSpinnerFactory: jest.fn()
  };
});

describe('runThinktank File Writing Functionality', () => {
  // Create real FileSystem implementation for verification
  const mockFileSystem: jest.Mocked<FileSystem> = {
    readFileContent: jest.fn().mockResolvedValue('Test content'),
    writeFile: jest.fn().mockResolvedValue(undefined),
    fileExists: jest.fn().mockResolvedValue(true),
    mkdir: jest.fn().mockResolvedValue(undefined),
    readdir: jest.fn().mockResolvedValue(['file1.txt']),
    stat: jest.fn().mockResolvedValue({ isFile: () => true, isDirectory: () => false }),
    access: jest.fn().mockResolvedValue(undefined),
    getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
    getConfigFilePath: jest.fn().mockResolvedValue('/mock/config/file.json')
  };

  // Mock LLM Responses
  const mockResponses: Array<LLMResponse & { configKey: string }> = [
    {
      provider: 'mock',
      modelId: 'model-a',
      text: 'Response from model A',
      configKey: 'mock:model-a',
      metadata: { tokens: 100 }
    },
    {
      provider: 'mock',
      modelId: 'model-b',
      text: 'Response from model B',
      configKey: 'mock:model-b',
      metadata: { tokens: 150 }
    }
  ];

  // Mock file data to be returned by _processOutput
  const mockFileData: FileData[] = [
    {
      filename: 'mock-model-a.md',
      content: 'Content for model A',
      modelKey: 'mock:model-a'
    },
    {
      filename: 'mock-model-b.md',
      content: 'Content for model B',
      modelKey: 'mock:model-b'
    }
  ];

  beforeEach(() => {
    jest.clearAllMocks();

    // Setup the mocked FileSystem adapter to verify file writes
    (jest.requireMock('../../core/FileSystemAdapter').FileSystemAdapter as jest.Mock).mockImplementation(() => mockFileSystem);

    // Setup the individual helper functions
    (helpers._setupWorkflow as jest.Mock).mockResolvedValue({
      config: { models: [] },
      friendlyRunName: 'test-run',
      outputDirectoryPath: '/test/output/test-run'
    });

    (helpers._processInput as jest.Mock).mockResolvedValue({
      inputResult: {
        content: 'Test input',
        sourceType: 'direct',
        metadata: { processingTimeMs: 5, originalLength: 10, finalLength: 10, normalized: false }
      },
      combinedContent: 'Test input',
      contextFiles: []
    });

    (helpers._selectModels as jest.Mock).mockReturnValue({
      models: [{ provider: 'mock', modelId: 'model-a' }, { provider: 'mock', modelId: 'model-b' }],
      modeDescription: 'All enabled models',
      warnings: []
    });

    (helpers._executeQueries as jest.Mock).mockResolvedValue({
      queryResults: {
        responses: mockResponses,
        statuses: {
          'mock:model-a': { status: 'success', durationMs: 100 },
          'mock:model-b': { status: 'success', durationMs: 150 }
        },
        timing: { startTime: 1000, endTime: 1150, durationMs: 150 },
        combinedContent: 'Test input'
      }
    });

    // Mock the refactored _processOutput to return structured data
    const mockPureProcessOutputResult: PureProcessOutputResult = {
      files: mockFileData,
      consoleOutput: 'Formatted output for console'
    };

    (helpers._processOutput as jest.Mock).mockResolvedValue(mockPureProcessOutputResult);
    
    // formatCompletionSummary is now used directly in runThinktank.ts
    jest.mock('../../utils/formatCompletionSummary', () => ({
      formatCompletionSummary: jest.fn().mockReturnValue({
        summaryText: 'Mock summary text',
        errorDetails: []
      })
    }));
    
    (helpers._handleWorkflowError as unknown as jest.Mock).mockImplementation(() => {
      return "Error occurred" as unknown as never;
    });
  });

  it('should write files based on data from _processOutput', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    await runThinktank(options);

    // Verify files were written
    expect(mockFileSystem.mkdir).toHaveBeenCalledWith('/test/output/test-run', expect.any(Object));
    expect(mockFileSystem.writeFile).toHaveBeenCalledTimes(2);
    
    // Check that the content and filenames match what was returned by _processOutput
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/test/output/test-run/mock-model-a.md',
      'Content for model A'
    );
    
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/test/output/test-run/mock-model-b.md',
      'Content for model B'
    );
  });

  it('should handle file writing errors', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Make the second file write fail
    mockFileSystem.writeFile
      .mockResolvedValueOnce(undefined) // First write succeeds
      .mockRejectedValueOnce(new Error('File write error')); // Second write fails

    await runThinktank(options);

    // Verify first file was written successfully
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/test/output/test-run/mock-model-a.md',
      'Content for model A'
    );

    // Verify second file attempted but failed
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/test/output/test-run/mock-model-b.md',
      'Content for model B'
    );

    // In our test setup, our mock doesn't call formatCompletionSummary directly,
    // so we don't need to verify it was called.
  });

  it('should create parent directories if needed', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Mock mkdir to check if it's creating parent directories
    mockFileSystem.mkdir.mockClear();

    await runThinktank(options);

    // Verify mkdir was called for the output directory
    expect(mockFileSystem.mkdir).toHaveBeenCalledWith('/test/output/test-run', expect.any(Object));
    
    // Verify that parent directory was created recursively
    expect(mockFileSystem.mkdir).toHaveBeenCalledWith(
      expect.stringContaining('/test/output/test-run'),
      expect.objectContaining({ recursive: true })
    );
  });

  it('should return console output from _processOutput', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    const result = await runThinktank(options);

    // Verify that it returns the console output from _processOutput
    expect(result).toBe('Formatted output for console');
  });
});
