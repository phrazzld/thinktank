/**
 * Integration tests for runThinktank.ts
 */
import { runThinktank, RunOptions } from '../runThinktank';
import { ConfigError, ApiError, FileSystemError } from '../../core/errors';
import * as helpers from '../runThinktankHelpers';
import * as nameGenerator from '../../utils/nameGenerator';
import { FileWriteStatus } from '../outputHandler';

// Store original module path for restoration
const helpersPath = require.resolve('../runThinktankHelpers');
const nameGeneratorPath = require.resolve('../../utils/nameGenerator');
const oraPath = require.resolve('ora');

// Mock dependencies
jest.mock('../runThinktankHelpers');
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

describe('runThinktank', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  // Restore all mocked modules after tests
  afterAll(() => {
    jest.unmock('../runThinktankHelpers');
    jest.unmock('../../utils/nameGenerator');
    jest.unmock('ora');
    
    // Clear module cache to ensure fresh imports
    delete require.cache[helpersPath];
    delete require.cache[nameGeneratorPath];
    delete require.cache[oraPath];
  });
  
  beforeEach(() => {
    // Default helper mock implementations
    const mockConfig = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
          options: { temperature: 0.7 }
        }
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
              options: { temperature: 0.7 }
            }
          ]
        }
      }
    };
    
    // Setup workflow helper mock
    (helpers._setupWorkflow as jest.Mock).mockResolvedValue({
      config: mockConfig,
      friendlyRunName: 'clever-meadow',
      outputDirectoryPath: '/fake/output/dir'
    });
    
    // Process input helper mock
    (helpers._processInput as jest.Mock).mockResolvedValue({
      inputResult: {
        content: 'Test prompt with context',
        sourceType: 'file',
        sourcePath: 'test-prompt.txt',
        metadata: {
          processingTimeMs: 5,
          originalLength: 11,
          finalLength: 25,
          normalized: true,
          hasContextFiles: true,
          contextFilesCount: 1
        }
      },
      // Add the combinedContent property to match the updated interface
      combinedContent: 'Test prompt with context'
    });
    
    // Select models helper mock with flattened structure
    const mockSelectionResult: any = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
          options: { temperature: 0.7 }
        }
      ],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: [],
      modeDescription: 'All enabled models'
    };
    // Add self-reference for backward compatibility
    mockSelectionResult.modelSelectionResult = mockSelectionResult;
    
    (helpers._selectModels as jest.Mock).mockReturnValue(mockSelectionResult);
    
    // Execute queries helper mock
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
            }
          }
        ],
        statuses: {
          'mock:mock-model': { 
            status: 'success',
            startTime: 1,
            endTime: 2,
            durationMs: 1
          }
        },
        timing: {
          startTime: 1,
          endTime: 2,
          durationMs: 1
        }
      }
    });
    
    // Process output helper mock
    (helpers._processOutput as jest.Mock).mockResolvedValue({
      fileOutputResult: {
        outputDirectory: '/fake/output/dir',
        files: [{ 
          modelKey: 'mock:mock-model', 
          filename: 'mock-model.md', 
          status: 'success' as FileWriteStatus,
          filePath: '/fake/output/dir/mock-model.md'
        }],
        succeededWrites: 1,
        failedWrites: 0,
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      },
      consoleOutput: 'Mock console output'
    });
    
    // Log completion summary helper mock
    (helpers._logCompletionSummary as jest.Mock).mockReturnValue({});
    
    // Error handling helper mock - should never actually run in success tests
    // Convert the error handler to a simple function that just rethrows the error
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation((params: any) => {
      throw params.error;
    });
    
    // Mock nameGenerator
    (nameGenerator.generateFunName as jest.Mock).mockResolvedValue('clever-meadow');
  });

  it('should run successfully with valid inputs', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    // Verify our helper functions were called in the correct sequence
    expect(helpers._setupWorkflow).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        options
      })
    );
    
    expect(helpers._processInput).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        input: 'test-prompt.txt'
      })
    );
    
    expect(helpers._selectModels).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        config: expect.any(Object),
        options
      })
    );
    
    expect(helpers._executeQueries).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        config: expect.any(Object),
        models: expect.any(Array),
        combinedContent: expect.any(String), // Now expects combinedContent instead of prompt
        options
      })
    );
    
    expect(helpers._processOutput).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        queryResults: expect.any(Object),
        outputDirectoryPath: '/fake/output/dir',
        options,
        friendlyRunName: 'clever-meadow'
      })
    );
    
    expect(helpers._logCompletionSummary).toHaveBeenCalledWith(
      expect.objectContaining({
        queryResults: expect.any(Object),
        fileOutputResult: expect.any(Object),
        options: expect.objectContaining({
          input: 'test-prompt.txt',
          friendlyRunName: 'clever-meadow'
        }),
        outputDirectoryPath: '/fake/output/dir'
      })
    );
    
    // Verify the result is the formatted console output
    expect(result).toBe('Mock console output');
  });

  it('should handle specified models correctly', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      models: ['mock:mock-model'],
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify we pass the models list to the select models helper
    expect(helpers._selectModels).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          models: ['mock:mock-model']
        })
      })
    );
  });
  
  it('should pass contextPaths from options to _processInput', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      contextPaths: ['src/file1.js', 'src/dir1/'],
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify contextPaths is passed to _processInput
    expect(helpers._processInput).toHaveBeenCalledWith(
      expect.objectContaining({
        spinner: expect.any(Object),
        input: 'test-prompt.txt',
        contextPaths: ['src/file1.js', 'src/dir1/']
      })
    );
  });
  
  it('should handle specific model parameter correctly', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      specificModel: 'mock:mock-model',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify we pass the specificModel to the select models helper
    expect(helpers._selectModels).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          specificModel: 'mock:mock-model'
        })
      })
    );
  });
  
  it('should handle specific group name parameter correctly', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      groupName: 'coding',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify we pass the groupName to the select models helper
    expect(helpers._selectModels).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          groupName: 'coding'
        })
      })
    );
  });
  
  it('should select system prompt from CLI override when provided', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      systemPrompt: 'Custom CLI system prompt override',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify system prompt from CLI was passed to query execution helper
    expect(helpers._executeQueries).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          systemPrompt: 'Custom CLI system prompt override'
        })
      })
    );
  });
  
  it('should enable thinking for Claude models when requested', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      enableThinking: true,
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify enableThinking was passed to query execution helper
    expect(helpers._executeQueries).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          enableThinking: true
        })
      })
    );
  });
  
  it('should include metadata when specified', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: true,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify includeMetadata was passed to process output helper
    expect(helpers._processOutput).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          includeMetadata: true
        })
      })
    );
  });
  
  it('should create output directory with custom path when provided', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: '/custom/output/path',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify output path was passed to setup workflow helper
    expect(helpers._setupWorkflow).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          output: '/custom/output/path'
        })
      })
    );
  });

  it('should return early when no models are selected', async () => {
    // Mock _selectModels to return empty models array with flattened structure
    const mockModelSelectionResult: any = {
      models: [],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: [],
      modeDescription: 'All enabled models'
    };
    // Add self-reference for backward compatibility
    mockModelSelectionResult.modelSelectionResult = mockModelSelectionResult;
    
    (helpers._selectModels as jest.Mock).mockReturnValueOnce(mockModelSelectionResult);
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    // Verify that the execute queries helper was not called
    expect(helpers._executeQueries).not.toHaveBeenCalled();
    
    // Verify that a warning message was returned
    expect(result).toContain('No models available');
  });

  it('should throw error from setup workflow helper', async () => {
    // Mock _setupWorkflow to throw an error
    const setupError = new ConfigError('Config loading failed');
    (helpers._setupWorkflow as jest.Mock).mockRejectedValueOnce(setupError);

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
    
    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: setupError,
        options,
        workflowState: expect.any(Object)
      })
    );
  });
  
  it('should throw error from process input helper', async () => {
    // Mock _processInput to throw an error
    const inputError = new FileSystemError('File not found');
    (helpers._processInput as jest.Mock).mockRejectedValueOnce(inputError);

    const options: RunOptions = {
      input: 'nonexistent.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    
    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: inputError,
        options,
        workflowState: expect.any(Object)
      })
    );
  });
  
  it('should throw error from select models helper', async () => {
    // Mock _selectModels to throw an error
    const modelError = new ConfigError('Invalid model format');
    (helpers._selectModels as jest.Mock).mockImplementationOnce(() => {
      throw modelError;
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      specificModel: 'invalid-format',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
    
    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: modelError,
        options,
        workflowState: expect.any(Object)
      })
    );
  });
  
  it('should throw error from execute queries helper', async () => {
    // Mock _executeQueries to throw an error
    const apiError = new ApiError('API call failed');
    (helpers._executeQueries as jest.Mock).mockRejectedValueOnce(apiError);

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    
    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: apiError,
        options,
        workflowState: expect.any(Object)
      })
    );
  });
  
  it('should throw error from process output helper', async () => {
    // Mock _processOutput to throw an error
    const fsError = new FileSystemError('Write error');
    (helpers._processOutput as jest.Mock).mockRejectedValueOnce(fsError);

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    
    // Verify the error handler was called with the right parameters
    expect(helpers._handleWorkflowError).toHaveBeenCalledWith(
      expect.objectContaining({
        error: fsError,
        options,
        workflowState: expect.any(Object)
      })
    );
  });
  
  it('should display extra metadata when includeMetadata is true', async () => {
    // Mock console.log to verify it's called with metadata
    const consoleLogSpy = jest.spyOn(console, 'log').mockImplementation();
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: true,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify that the completion summary was called and console.log was used
    expect(helpers._logCompletionSummary).toHaveBeenCalled();
    expect(consoleLogSpy).toHaveBeenCalled();
    
    // Restore console.log
    consoleLogSpy.mockRestore();
  });
});