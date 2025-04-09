/**
 * Tests specifically for error handling in runThinktank
 */
import { runThinktank, RunOptions } from '../runThinktank';
import {
  ConfigError,
  ApiError,
  FileSystemError,
  PermissionError,
  ThinktankError,
  errorCategories,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError,
} from '../../core/errors';
import * as helpers from '../runThinktankHelpers';
import * as nameGenerator from '../../utils/nameGenerator';
import * as ioModule from '../io';
import { FileWriteStatus } from '../outputHandler';

// Store module paths for restoration
const helpersPath = require.resolve('../runThinktankHelpers');
const nameGeneratorPath = require.resolve('../../utils/nameGenerator');
const oraPath = require.resolve('ora');

// Mock dependencies
jest.mock('../runThinktankHelpers');
jest.mock('../io');
jest.mock('../../utils/nameGenerator');
jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => {
    return {
      start: jest.fn().mockReturnThis(),
      stop: jest.fn().mockReturnThis(),
      succeed: jest.fn().mockReturnThis(),
      fail: jest.fn().mockReturnThis(),
      warn: jest.fn().mockReturnThis(),
      info: jest.fn().mockReturnThis(),
      text: '',
    };
  });
});

describe('runThinktank Error Handling', () => {
  beforeEach(() => {
    jest.clearAllMocks();

    // Default helper mock implementations
    const mockConfig = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
          options: { temperature: 0.7 },
        },
      ],
      groups: {
        default: {
          name: 'default',
          systemPrompt: { text: 'You are a helpful assistant.' },
          models: [
            {
              provider: 'mock',
              modelId: 'mock-model',
              enabled: true,
              options: { temperature: 0.7 },
            },
          ],
        },
      },
    };

    // Setup workflow helper mock - success by default
    (helpers._setupWorkflow as jest.Mock).mockResolvedValue({
      config: mockConfig,
      friendlyRunName: 'clever-meadow',
      outputDirectoryPath: '/fake/output/dir',
    });

    // Process input helper mock - success by default
    (helpers._processInput as jest.Mock).mockResolvedValue({
      inputResult: {
        content: 'Test prompt',
        sourceType: 'file',
        sourcePath: 'test-prompt.txt',
        metadata: {
          processingTimeMs: 5,
          originalLength: 11,
          finalLength: 11,
          normalized: true,
        },
      },
    });

    // Select models helper mock - success by default with flattened structure
    const mockSelectionResult: any = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
          options: { temperature: 0.7 },
        },
      ],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: [],
      modeDescription: 'All enabled models',
    };
    // Add self-reference for backward compatibility
    mockSelectionResult.modelSelectionResult = mockSelectionResult;

    (helpers._selectModels as jest.Mock).mockReturnValue(mockSelectionResult);

    // Execute queries helper mock - success by default
    (helpers._executeQueries as jest.Mock).mockResolvedValue({
      queryResults: {
        responses: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            text: 'Mock response for prompt: Test prompt',
            configKey: 'mock:mock-model',
            metadata: {
              usage: { total_tokens: 10 },
              model: 'mock-model',
              id: 'mock-response-id',
            },
          },
        ],
        statuses: {
          'mock:mock-model': {
            status: 'success',
            startTime: 1,
            endTime: 2,
            durationMs: 1,
          },
        },
        timing: {
          startTime: 1,
          endTime: 2,
          durationMs: 1,
        },
      },
    });

    // Process output helper mock - success by default
    (helpers._processOutput as jest.Mock).mockReturnValue({
      files: [
        {
          modelKey: 'mock:mock-model',
          filename: 'mock-model.md',
          content: 'Mock content',
        },
      ],
      consoleOutput: 'Mock console output',
    });

    // Write files helper mock - success by default
    (ioModule.writeFiles as jest.Mock).mockResolvedValue({
      outputDirectory: '/fake/output/dir',
      files: [
        {
          modelKey: 'mock:mock-model',
          filename: 'mock-model.md',
          status: 'success' as FileWriteStatus,
          filePath: '/fake/output/dir/mock-model.md',
        },
      ],
      succeededWrites: 1,
      failedWrites: 0,
      timing: { startTime: 1, endTime: 2, durationMs: 1 },
    });

    // formatCompletionSummary is now used directly in runThinktank.ts
    jest.mock('../../utils/formatCompletionSummary', () => ({
      formatCompletionSummary: jest.fn().mockReturnValue({
        summaryText: 'Mock summary text',
        errorDetails: [],
      }),
    }));

    // Convert the error handler to a simple function that rethrows the error
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation((params: any) => {
      throw params.error;
    });

    // No reusable error handler needed since the default mock already rethrows the error

    // Mock nameGenerator
    (nameGenerator.generateFunName as jest.Mock).mockResolvedValue('clever-meadow');
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  // Restore all mocked modules after tests
  afterAll(() => {
    jest.unmock('../runThinktankHelpers');
    jest.unmock('../io');
    jest.unmock('../../utils/nameGenerator');
    jest.unmock('ora');

    // Clear module cache to ensure fresh imports
    delete require.cache[helpersPath];
    delete require.cache[nameGeneratorPath];
    delete require.cache[oraPath];
  });

  it('should throw FileSystemError when prompt file cannot be read', async () => {
    // Create a FileSystemError for file not found
    const fileNotFoundError = createFileNotFoundError('nonexistent.txt');

    // Mock _processInput to throw the error
    (helpers._processInput as jest.Mock).mockRejectedValueOnce(fileNotFoundError);

    // Call with test options
    const options: RunOptions = {
      input: 'nonexistent.txt',
    };

    // Expect it to throw the FileSystemError
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('FileSystemError');
      expect(error.category).toBe(errorCategories.FILESYSTEM);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.filePath).toBe('nonexistent.txt');
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: fileNotFoundError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw PermissionError when output directory creation fails', async () => {
    // Create a PermissionError for directory creation
    const permissionError = new PermissionError(
      'Permission denied: Failed to create output directory',
      {
        suggestions: [
          'Check that you have write permissions for the directory',
          'Try using a different output path',
          'Ensure the parent directory exists and is writable',
        ],
      }
    );

    // Mock _setupWorkflow to throw the error
    (helpers._setupWorkflow as jest.Mock).mockRejectedValueOnce(permissionError);

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
      output: '/invalid/dir',
    };

    // Expect it to throw the PermissionError
    await expect(runThinktank(options)).rejects.toThrow(PermissionError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('PermissionError');
      expect(error.category).toBe(errorCategories.PERMISSION);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/permission denied/i);
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: permissionError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw ConfigError when model format is invalid', async () => {
    // Create a ConfigError for invalid model format
    const modelFormatError = createModelFormatError(
      'openai-gpt4', // Invalid format (missing colon)
      ['openai', 'anthropic', 'google'],
      ['openai:gpt-4o', 'anthropic:claude-3-opus']
    );

    // Mock _selectModels to throw the error
    (helpers._selectModels as jest.Mock).mockImplementationOnce(() => {
      throw modelFormatError;
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
      specificModel: 'openai-gpt4',
    };

    // Expect it to throw the ConfigError
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('ConfigError');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/model format|provider:modelId/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: modelFormatError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw ConfigError when model is not found', async () => {
    // Create a ConfigError for model not found
    const modelNotFoundError = new ConfigError(
      'Model "openai:nonexistent-model" not found in configuration.',
      {
        suggestions: [
          'Check that the model is correctly spelled and exists in your configuration',
          'Available models: openai:gpt-4o, anthropic:claude-3-opus',
        ],
        examples: [
          'thinktank run prompt.txt --models=openai:gpt-4o',
          'thinktank run prompt.txt --group=default',
        ],
      }
    );

    // Mock _selectModels to throw the error
    (helpers._selectModels as jest.Mock).mockImplementationOnce(() => {
      throw modelNotFoundError;
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
      specificModel: 'openai:nonexistent-model',
    };

    // Expect it to throw the ConfigError
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('ConfigError');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/model.*not found/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: modelNotFoundError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw ApiError when API keys are missing', async () => {
    // Create an ApiError for missing API keys
    const missingApiKeyError = createMissingApiKeyError([
      { provider: 'openai', modelId: 'gpt-4o' },
      { provider: 'anthropic', modelId: 'claude-3-opus' },
    ]);

    // Mock _selectModels to throw the error
    (helpers._selectModels as jest.Mock).mockImplementationOnce(() => {
      throw missingApiKeyError;
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw the ApiError
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('ApiError');
      expect(error.category).toBe(errorCategories.API);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/missing api key/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: missingApiKeyError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw ApiError when query execution fails', async () => {
    // Create an ApiError for API call failure
    const apiError = new ApiError('Failed to generate response from API', {
      providerId: 'openai',
      cause: new Error('Network error'),
      suggestions: [
        'Check your internet connection',
        'Verify the API endpoint is correct',
        'Try again later',
      ],
    });

    // Mock _executeQueries to throw the error
    (helpers._executeQueries as jest.Mock).mockRejectedValueOnce(apiError);

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw the ApiError
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('ApiError');
      expect(error.category).toBe(errorCategories.API);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.cause).toBeDefined();
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: apiError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should throw FileSystemError when file writing fails', async () => {
    // Create a FileSystemError for file writing
    const fsError = new FileSystemError('Failed to write results to file', {
      filePath: '/output/directory/mock-model.md',
      suggestions: [
        'Check file system permissions',
        'Ensure the directory exists and is writable',
        'Verify there is enough disk space',
      ],
    });

    // Mock writeFiles to throw the error
    (ioModule.writeFiles as jest.Mock).mockRejectedValueOnce(fsError);

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw the FileSystemError
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    // Verify more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('FileSystemError');
      expect(error.category).toBe(errorCategories.FILESYSTEM);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.filePath).toBeDefined();
    }

    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: fsError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should handle multiple model errors gracefully', async () => {
    // Setup specific formatCompletionSummary mock for this test
    const formatCompletionSummaryMock = jest.requireMock(
      '../../utils/formatCompletionSummary'
    ).formatCompletionSummary;
    formatCompletionSummaryMock.mockClear(); // Clear previous calls
    formatCompletionSummaryMock.mockReturnValue({
      summaryText: 'Mock summary with errors',
      errorDetails: ['- Model error details'],
    });

    // Create query results with mixed success/error results
    const mixedQueryResults = {
      queryResults: {
        responses: [
          {
            provider: 'openai',
            modelId: 'gpt-4o',
            text: 'This is a test response',
            configKey: 'openai:gpt-4o',
            metadata: {
              usage: { total_tokens: 10 },
              model: 'gpt-4o',
              id: 'mock-response-id',
            },
          },
          {
            provider: 'anthropic',
            modelId: 'claude-3-opus',
            text: '',
            error: 'Rate limit exceeded',
            errorCategory: 'API Rate Limit',
            errorTip: 'Try again later or reduce the number of requests',
            configKey: 'anthropic:claude-3-opus',
          },
        ],
        statuses: {
          'openai:gpt-4o': {
            status: 'success',
            startTime: Date.now() - 1000,
            endTime: Date.now(),
            durationMs: 1000,
          },
          'anthropic:claude-3-opus': {
            status: 'error',
            message: 'Rate limit exceeded',
            detailedError: new ApiError('Rate limit exceeded', {
              providerId: 'anthropic',
              suggestions: [
                'Try again later or reduce the number of requests',
                'Consider using a different model or provider',
              ],
            }),
            startTime: Date.now() - 1200,
            endTime: Date.now(),
            durationMs: 1200,
          },
        },
        timing: {
          startTime: Date.now() - 1500,
          endTime: Date.now(),
          durationMs: 1500,
        },
      },
    };

    // Setup pure data output for files
    const pureProcessOutput = {
      files: [
        {
          modelKey: 'openai:gpt-4o',
          filename: 'openai-gpt-4o.md',
          content: 'Success content',
        },
        {
          modelKey: 'anthropic:claude-3-opus',
          filename: 'anthropic-claude-3-opus.md',
          content: 'Error content',
        },
      ],
      consoleOutput: 'Mixed results console output',
    };

    // Setup file output result with mixed success/failure
    const mixedFileOutputResult = {
      outputDirectory: '/fake/output/dir',
      files: [
        {
          modelKey: 'openai:gpt-4o',
          filename: 'openai-gpt-4o.md',
          status: 'success' as FileWriteStatus,
          filePath: '/fake/output/dir/openai-gpt-4o.md',
          startTime: 1,
          endTime: 2,
          durationMs: 1,
        },
        {
          modelKey: 'anthropic:claude-3-opus',
          filename: 'anthropic-claude-3-opus.md',
          status: 'error' as FileWriteStatus,
          error: 'Failed to write file',
          filePath: '/fake/output/dir/anthropic-claude-3-opus.md',
          startTime: 1,
          endTime: 2,
          durationMs: 1,
        },
      ],
      succeededWrites: 1,
      failedWrites: 1,
      timing: { startTime: 1, endTime: 2, durationMs: 1 },
    };

    // Mock _executeQueries to return mixed results
    (helpers._executeQueries as jest.Mock).mockResolvedValueOnce(mixedQueryResults);

    // Mock _processOutput to return pure data (no I/O)
    (helpers._processOutput as jest.Mock).mockReturnValueOnce(pureProcessOutput);

    // Mock writeFiles to return mixed success/failure result
    (ioModule.writeFiles as jest.Mock).mockResolvedValueOnce(mixedFileOutputResult);

    // Mock _selectModels to return multiple models with flattened structure
    const multiModelMock: any = {
      models: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true,
        },
        {
          provider: 'anthropic',
          modelId: 'claude-3-opus',
          enabled: true,
        },
      ],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: [],
      modeDescription: 'All enabled models',
    };
    // Add self-reference for backward compatibility
    multiModelMock.modelSelectionResult = multiModelMock;

    (helpers._selectModels as jest.Mock).mockReturnValueOnce(multiModelMock);

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Execute and expect it to complete without throwing
    const result = await runThinktank(options);

    // Verify the formatting result was returned
    expect(result).toBe('Mixed results console output');

    // Verify both executeQueries and processOutput were called
    expect(helpers._executeQueries).toHaveBeenCalled();
    expect(helpers._processOutput).toHaveBeenCalled();

    // Verify the _processOutput was called with query results from both models
    expect(helpers._processOutput).toHaveBeenCalledWith(
      expect.objectContaining({
        queryResults: expect.objectContaining({
          responses: expect.arrayContaining([
            expect.objectContaining({ provider: 'openai', modelId: 'gpt-4o' }),
            expect.objectContaining({ provider: 'anthropic', modelId: 'claude-3-opus' }),
          ]),
        }),
      })
    );

    // Verify our process output and write files functions were called
    expect(helpers._processOutput).toHaveBeenCalled();
    expect(ioModule.writeFiles).toHaveBeenCalled();
  });

  it('should properly propagate error causes through the call chain', async () => {
    // Create a chain of errors with causes
    const rootCause = new Error('Network connection failed');
    const apiError = new ApiError('[openai] Failed to connect to OpenAI API', {
      providerId: 'openai',
      cause: rootCause,
      suggestions: [
        'Check your internet connection',
        'Verify the API endpoint is correct',
        'Try again later',
      ],
    });

    // Mock _processInput to throw an ApiError with a cause
    (helpers._processInput as jest.Mock).mockRejectedValueOnce(apiError);

    // Mock _handleWorkflowError to preserve cause chain
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementationOnce((params: any) => {
      // Preserve the error and its cause chain
      throw params.error;
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw the ApiError with preserved cause chain
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    // Verify error propagation of cause chain
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('ApiError');
      expect(error.category).toBe(errorCategories.API);
      expect(error.cause).toBeDefined();
      expect(error.cause.message).toBe('Network connection failed');
      expect(error.providerId).toBe('openai');
    }

    // Verify the error handler was called with the correct parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: apiError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should categorize unknown errors using the error handler', async () => {
    // Create an unknown error (not a ThinktankError)
    const unknownError = new Error('Something went wrong');

    // Mock _setupWorkflow to throw an unknown error
    (helpers._setupWorkflow as jest.Mock).mockRejectedValueOnce(unknownError);

    // Spy on the _handleWorkflowError implementation to categorize the error
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementationOnce((params: any) => {
      // Simulate error categorization by returning a ThinktankError
      throw new ThinktankError(`Unknown error during workflow: ${params.error.message}`, {
        cause: params.error,
        category: errorCategories.UNKNOWN,
      });
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw a ThinktankError after categorization
    await expect(runThinktank(options)).rejects.toThrow(ThinktankError);
    // Verify error categorization
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error).toBeInstanceOf(ThinktankError);
      expect(error.category).toBe(errorCategories.UNKNOWN);
      expect(error.cause).toBe(unknownError);
    }

    // Verify the error handler was called with the correct parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: unknownError,
        options,
        workflowState: expect.any(Object),
      })
    );
  });

  it('should add workflow context to error suggestions', async () => {
    // Create a FileSystemError
    const fsError = new FileSystemError('Failed to write to file', {
      filePath: '/path/to/file.txt',
      suggestions: ['Check file permissions'],
    });

    // Mock writeFiles to throw the error
    (ioModule.writeFiles as jest.Mock).mockRejectedValueOnce(fsError);

    // Mock _handleWorkflowError to add workflow context to suggestions
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementationOnce((params: any) => {
      // Add workflow context to suggestions
      const error = params.error;
      const suggestions = error.suggestions || [];
      error.suggestions = [
        ...suggestions,
        `This error occurred during run: ${params.workflowState.friendlyRunName}`,
        `Output directory: ${params.workflowState.outputDirectoryPath}`,
      ];
      throw error;
    });

    // Call with test options
    const options: RunOptions = {
      input: 'prompt.txt',
    };

    // Expect it to throw the FileSystemError with enhanced suggestions
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    // Verify enhanced suggestions
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.suggestions).toContain('Check file permissions');
      expect(error.suggestions).toContain('This error occurred during run: clever-meadow');
      expect(error.suggestions).toContain('Output directory: /fake/output/dir');
    }

    // Verify the error handler was called with the correct parameters and workflow state
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: fsError,
        options,
        workflowState: expect.objectContaining({
          friendlyRunName: 'clever-meadow',
          outputDirectoryPath: '/fake/output/dir',
        }),
      })
    );
  });
});
