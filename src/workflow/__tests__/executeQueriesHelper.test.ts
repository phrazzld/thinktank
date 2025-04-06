/**
 * Unit tests for the _executeQueries helper function
 */
import { _executeQueries } from '../runThinktankHelpers';
import * as queryExecutor from '../queryExecutor';
import { ApiError } from '../../core/errors';
import { QueryExecutorError, ModelQueryStatus } from '../queryExecutor';
import { AppConfig, ModelConfig } from '../../core/types';

// Mock dependencies
jest.mock('../queryExecutor');

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_executeQueries Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  // Sample app config for tests
  const sampleConfig: AppConfig = {
    models: [],
    groups: {}
  };

  // Sample models for testing
  const testModels: ModelConfig[] = [
    {
      provider: 'mock',
      modelId: 'mock-model',
      enabled: true
    } as ModelConfig,
    {
      provider: 'openai',
      modelId: 'gpt-4o',
      enabled: true
    } as ModelConfig
  ];

  it('should successfully execute queries', async () => {
    // Setup mocks
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
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
          status: 'success',
          startTime: 1,
          endTime: 2,
          durationMs: 1
        },
        'openai:gpt-4o': {
          status: 'success',
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
    });

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
        systemPrompt: 'You are a helpful assistant'
      }
    });

    // Verify the result
    expect(result.queryResults.responses.length).toBe(2);
    expect(result.queryResults.responses[0].text).toBe('Mock response');
    expect(result.queryResults.responses[1].text).toBe('OpenAI response');

    // Verify mocks were called correctly
    expect(queryExecutor.executeQueries).toHaveBeenCalledWith(
      sampleConfig,
      testModels,
      expect.objectContaining({
        prompt: 'Test prompt',
        systemPrompt: 'You are a helpful assistant'
      })
    );

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Query execution complete');
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.start).toHaveBeenCalled();
  });

  it('should handle thinking capability for Claude models', async () => {
    // Setup mocks
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
      responses: [
        {
          provider: 'anthropic',
          modelId: 'claude-3-opus',
          text: 'Claude response with thinking',
          configKey: 'anthropic:claude-3-opus',
          metadata: {
            thinking: 'Deep thinking process...'
          }
        }
      ],
      statuses: {
        'anthropic:claude-3-opus': {
          status: 'success',
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
    });

    // Call the function with thinking enabled
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [
        {
          provider: 'anthropic',
          modelId: 'claude-3-opus',
          enabled: true
        } as ModelConfig
      ],
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
        enableThinking: true
      }
    });

    // Verify enableThinking was passed to queryExecutor
    expect(queryExecutor.executeQueries).toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      expect.objectContaining({
        enableThinking: true
      })
    );
  });

  it('should handle timeout override', async () => {
    // Setup mocks
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: 'Quick response',
          configKey: 'mock:mock-model',
          metadata: {}
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
    });

    // Call the function with custom timeout
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [testModels[0]],
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
        timeoutMs: 60000 // 1 minute timeout
      }
    });

    // Verify timeout was passed to queryExecutor
    expect(queryExecutor.executeQueries).toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      expect.objectContaining({
        timeoutMs: 60000
      })
    );
  });

  it('should handle status updates during execution', async () => {
    // Setup mocks with side effect to call onStatusUpdate
    (queryExecutor.executeQueries as jest.Mock).mockImplementation((_config, _models, options) => {
      // Call the status update callback to simulate model execution status changes
      if (options.onStatusUpdate) {
        // Running state
        options.onStatusUpdate('mock:mock-model', {
          status: 'running'
        } as ModelQueryStatus, {});
        
        // Success state
        options.onStatusUpdate('mock:mock-model', {
          status: 'success',
          durationMs: 1500
        } as ModelQueryStatus, {});
      }

      // Return mock query result
      return Promise.resolve({
        responses: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            text: 'Response after status updates',
            configKey: 'mock:mock-model',
            metadata: {}
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
      });
    });

    // Call the function
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [testModels[0]],
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    });

    // Verify spinner text updates from status updates
    // Should have been updated to reflect the running state
    expect(mockSpinner.text).toContain('Query execution complete');
  });

  it('should handle partial success with some failures', async () => {
    // Setup mocks with mixture of success and failure
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
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
          text: '',
          error: 'API call failed',
          configKey: 'openai:gpt-4o',
          metadata: {}
        }
      ],
      statuses: {
        'mock:mock-model': {
          status: 'success',
          startTime: 1,
          endTime: 2,
          durationMs: 1
        },
        'openai:gpt-4o': {
          status: 'error',
          message: 'API call failed',
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
    });

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    });

    // Verify the results include both success and error
    expect(result.queryResults.responses.length).toBe(2);
    expect(result.queryResults.responses[0].text).toBeTruthy();
    expect(result.queryResults.responses[1].error).toBeTruthy();

    // Verify spinner shows appropriate warnings
    expect(mockSpinner.warn).toHaveBeenCalled();
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.text).toContain('Query execution complete');
  });

  it('should handle all models failing', async () => {
    // Setup mocks with all failures
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: '',
          error: 'Timeout error',
          configKey: 'mock:mock-model',
          metadata: {}
        },
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          text: '',
          error: 'API key error',
          configKey: 'openai:gpt-4o',
          metadata: {}
        }
      ],
      statuses: {
        'mock:mock-model': {
          status: 'error',
          message: 'Timeout error',
          startTime: 1,
          endTime: 2,
          durationMs: 1
        },
        'openai:gpt-4o': {
          status: 'error',
          message: 'API key error',
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
    });

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    });

    // Verify all responses have errors
    expect(result.queryResults.responses.length).toBe(2);
    expect(result.queryResults.responses.every(r => r.error)).toBeTruthy();
    expect(Object.values(result.queryResults.statuses).every(s => s.status === 'error')).toBeTruthy();

    // Verify spinner shows appropriate warnings and error summary
    expect(mockSpinner.warn).toHaveBeenCalled();
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.text).toContain('Query execution complete');
  });

  it('should handle QueryExecutor error by wrapping it in ApiError', async () => {
    // Setup mocks
    const executorError = new QueryExecutorError('Failed to execute queries');
    (queryExecutor.executeQueries as jest.Mock).mockRejectedValue(executorError);

    // Call the function and expect it to throw
    await expect(_executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(ApiError);

    // Verify correct wrapping
    try {
      await _executeQueries({
        spinner: mockSpinner,
        config: sampleConfig,
        models: testModels,
        prompt: 'Test prompt',
        options: {
          input: 'test-prompt.txt'
        }
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      if (error instanceof ApiError) {
        expect(error.message).toBe('Failed to execute queries');
        expect(error.cause).toBe(executorError);
      }
    }
  });

  it('should handle existing ApiError by rethrowing', async () => {
    // Setup mocks
    const apiError = new ApiError('API connection error');
    (queryExecutor.executeQueries as jest.Mock).mockRejectedValue(apiError);

    // Call the function and expect it to throw the original ApiError
    await expect(_executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(apiError);
  });

  it('should handle unknown errors by wrapping them in ApiError', async () => {
    // Setup mocks
    const unknownError = new Error('Something unexpected happened');
    (queryExecutor.executeQueries as jest.Mock).mockRejectedValue(unknownError);

    // Call the function and expect it to throw ApiError
    await expect(_executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      prompt: 'Test prompt',
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(ApiError);

    // Verify proper error wrapping
    try {
      await _executeQueries({
        spinner: mockSpinner,
        config: sampleConfig,
        models: testModels,
        prompt: 'Test prompt',
        options: {
          input: 'test-prompt.txt'
        }
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      if (error instanceof ApiError) {
        expect(error.message).toContain('Failed to execute queries');
        expect(error.cause).toBe(unknownError);
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
      }
    }
  });
});