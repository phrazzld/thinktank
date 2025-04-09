/**
 * Unit tests for the _executeQueries helper function
 */
import { _executeQueries } from '../runThinktankHelpers';
import { ApiError, ThinktankError, errorCategories } from '../../core/errors';
import { AppConfig, ModelConfig, LLMResponse } from '../../core/types';
import * as configManager from '../../core/configManager';
import { LLMClient } from '../../core/interfaces';

// Mock dependencies
jest.mock('../../core/configManager', () => ({
  findModelGroup: jest.fn(),
}));

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_executeQueries Helper', () => {
  // Create a mock LLMClient
  const mockLLMClient: jest.Mocked<LLMClient> = {
    generate: jest.fn(),
  };

  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
    // Reset mockLLMClient
    mockLLMClient.generate.mockReset();
    // Reset configManager.findModelGroup
    (configManager.findModelGroup as jest.Mock).mockReset();
  });

  // Sample app config for tests
  const sampleConfig: AppConfig = {
    models: [],
    groups: {},
  };

  // Sample models for testing
  const testModels: ModelConfig[] = [
    {
      provider: 'mock',
      modelId: 'mock-model',
      enabled: true,
    } as ModelConfig,
    {
      provider: 'openai',
      modelId: 'gpt-4o',
      enabled: true,
    } as ModelConfig,
  ];

  it('should successfully execute queries', async () => {
    // Setup mocks
    const mockResponses = [
      {
        provider: 'mock',
        modelId: 'mock-model',
        text: 'Mock response',
        metadata: {},
      },
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        text: 'OpenAI response',
        metadata: {},
      },
    ];

    // Mock successful responses from the LLMClient
    mockLLMClient.generate
      .mockResolvedValueOnce(mockResponses[0])
      .mockResolvedValueOnce(mockResponses[1]);

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      combinedContent: 'Test prompt with context',
      options: {
        input: 'test-prompt.txt',
        systemPrompt: 'You are a helpful assistant',
      },
      llmClient: mockLLMClient,
    });

    // Verify the result contains both responses
    expect(result.queryResults.responses.length).toBe(2);
    expect(result.queryResults.responses[0].text).toBe('Mock response');
    expect(result.queryResults.responses[1].text).toBe('OpenAI response');

    // Verify llmClient.generate was called with correct arguments
    expect(mockLLMClient.generate).toHaveBeenCalledTimes(2);
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt with context',
      'mock:mock-model',
      expect.any(Object),
      expect.objectContaining({
        text: 'You are a helpful assistant',
      })
    );
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt with context',
      'openai:gpt-4o',
      expect.any(Object),
      expect.objectContaining({
        text: 'You are a helpful assistant',
      })
    );

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Query execution complete');
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.start).toHaveBeenCalled();
  });

  it('should handle thinking capability for Claude models', async () => {
    // Setup model with Claude model
    const claudeModel: ModelConfig = {
      provider: 'anthropic',
      modelId: 'claude-3-opus',
      enabled: true,
    } as ModelConfig;

    // Setup mock response
    const mockResponse: LLMResponse = {
      provider: 'anthropic',
      modelId: 'claude-3-opus',
      text: 'Claude response with thinking',
      metadata: {
        thinking: 'Deep thinking process...',
      },
    };

    // Mock successful response from LLMClient
    mockLLMClient.generate.mockResolvedValueOnce(mockResponse);

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function with thinking enabled
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [claudeModel],
      combinedContent: 'Test prompt with context for thinking',
      options: {
        input: 'test-prompt.txt',
        enableThinking: true,
      },
      llmClient: mockLLMClient,
    });

    // Verify the LLMClient was called with thinking option
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt with context for thinking',
      'anthropic:claude-3-opus',
      expect.objectContaining({
        thinking: expect.objectContaining({
          type: 'enabled',
        }),
      }),
      undefined
    );
  });

  it('should handle timeout override', async () => {
    // Setup mock response
    const mockResponse: LLMResponse = {
      provider: 'mock',
      modelId: 'mock-model',
      text: 'Quick response',
      metadata: {},
    };

    // Mock successful response from LLMClient
    mockLLMClient.generate.mockResolvedValueOnce(mockResponse);

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function with custom timeout
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [testModels[0]],
      combinedContent: 'Test prompt with timeout',
      options: {
        input: 'test-prompt.txt',
        timeoutMs: 60000, // 1 minute timeout
      },
      llmClient: mockLLMClient,
    });

    // Verify the LLMClient was called with timeout option
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt with timeout',
      'mock:mock-model',
      expect.objectContaining({
        timeout: 60000,
      }),
      undefined
    );
  });

  it('should handle partial success with some failures', async () => {
    // Setup mock responses - one success, one error
    mockLLMClient.generate
      .mockResolvedValueOnce({
        provider: 'mock',
        modelId: 'mock-model',
        text: 'Mock response',
        metadata: {},
      })
      .mockRejectedValueOnce(new Error('API call failed'));

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      combinedContent: 'Test prompt with partial failures',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
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
    // Mock rejected responses from LLMClient
    mockLLMClient.generate
      .mockRejectedValueOnce(new Error('Timeout error'))
      .mockRejectedValueOnce(new Error('API key error'));

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: testModels,
      combinedContent: 'Test prompt with all failures',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify all responses have errors
    expect(result.queryResults.responses.length).toBe(2);
    expect(result.queryResults.responses.every(r => r.error)).toBeTruthy();
    expect(
      Object.values(result.queryResults.statuses).every(s => s.status === 'error')
    ).toBeTruthy();

    // Verify spinner shows appropriate warnings and error summary
    expect(mockSpinner.warn).toHaveBeenCalled();
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.text).toContain('Query execution complete');
  });

  it('should use system prompts from model if available', async () => {
    // Setup model with systemPrompt
    const modelWithSystemPrompt: ModelConfig = {
      provider: 'model',
      modelId: 'with-system-prompt',
      enabled: true,
      systemPrompt: {
        text: 'Model-specific system prompt',
        metadata: { source: 'model-config' },
      },
    } as ModelConfig;

    // Mock successful response
    mockLLMClient.generate.mockResolvedValueOnce({
      provider: 'model',
      modelId: 'with-system-prompt',
      text: 'Response with model-specific system prompt',
      metadata: {},
    });

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [modelWithSystemPrompt],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify the LLMClient was called with model's system prompt
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt',
      'model:with-system-prompt',
      expect.any(Object),
      expect.objectContaining({
        text: 'Model-specific system prompt',
        metadata: { source: 'model-config' },
      })
    );
  });

  it('should use system prompts from group if available and model has none', async () => {
    // Setup model without systemPrompt
    const model: ModelConfig = {
      provider: 'group',
      modelId: 'model',
      enabled: true,
    } as ModelConfig;

    // Mock group with system prompt
    const mockGroup = {
      groupName: 'test-group',
      systemPrompt: {
        text: 'Group system prompt',
        metadata: { source: 'group-config' },
      },
    };

    // Mock successful response
    mockLLMClient.generate.mockResolvedValueOnce({
      provider: 'group',
      modelId: 'model',
      text: 'Response with group system prompt',
      metadata: {},
    });

    // Mock findModelGroup to return the group
    (configManager.findModelGroup as jest.Mock).mockReturnValue(mockGroup);

    // Call the function
    await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [model],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify the LLMClient was called with group's system prompt
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt',
      'group:model',
      expect.any(Object),
      expect.objectContaining({
        text: 'Group system prompt',
        metadata: { source: 'group-config' },
      })
    );
  });

  it('should add group info to response if available', async () => {
    // Setup model
    const model: ModelConfig = {
      provider: 'test',
      modelId: 'model',
      enabled: true,
    } as ModelConfig;

    // Mock group
    const mockGroup = {
      groupName: 'test-group',
      systemPrompt: {
        text: 'Group system prompt',
        metadata: { source: 'group-config' },
      },
    };

    // Mock successful response without groupInfo
    mockLLMClient.generate.mockResolvedValueOnce({
      provider: 'test',
      modelId: 'model',
      text: 'Response',
      metadata: {},
    });

    // Mock findModelGroup to return the group
    (configManager.findModelGroup as jest.Mock).mockReturnValue(mockGroup);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [model],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify groupInfo was added to the response
    expect(result.queryResults.responses[0].groupInfo).toEqual({
      name: 'test-group',
      systemPrompt: mockGroup.systemPrompt,
    });
  });

  it('should not add group info if response already has it', async () => {
    // Setup model
    const model: ModelConfig = {
      provider: 'test',
      modelId: 'model',
      enabled: true,
    } as ModelConfig;

    // Mock group
    const mockGroup = {
      groupName: 'test-group',
      systemPrompt: {
        text: 'Group system prompt',
        metadata: { source: 'group-config' },
      },
    };

    // Create response that already has groupInfo
    const responseWithGroupInfo = {
      provider: 'test',
      modelId: 'model',
      text: 'Response with existing group info',
      metadata: {},
      groupInfo: {
        name: 'client-added-group',
        systemPrompt: { text: 'Client-added system prompt' },
      },
    };

    // Mock successful response with existing groupInfo
    mockLLMClient.generate.mockResolvedValueOnce(responseWithGroupInfo);

    // Mock findModelGroup to return the group
    (configManager.findModelGroup as jest.Mock).mockReturnValue(mockGroup);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [model],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify existing groupInfo was preserved
    expect(result.queryResults.responses[0].groupInfo).toEqual({
      name: 'client-added-group',
      systemPrompt: { text: 'Client-added system prompt' },
    });
  });

  it('should capture ThinktankError in response', async () => {
    // Create a ThinktankError (non-ApiError)
    const thinktankError = new ThinktankError('Custom Thinktank error', {
      category: errorCategories.CONFIG,
    });

    // Mock LLMClient to throw the ThinktankError
    mockLLMClient.generate.mockRejectedValueOnce(thinktankError);

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [testModels[0]],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify the error is captured in the response
    expect(result.queryResults.responses[0].error).toBe('Custom Thinktank error');
    expect((result.queryResults.responses[0] as any).errorCategory).toBe(errorCategories.CONFIG);
    expect(result.queryResults.statuses['mock:mock-model'].status).toBe('error');
  });

  it('should capture ApiError in response', async () => {
    // Create an ApiError
    const apiError = new ApiError('API communication error');

    // Mock LLMClient to throw the ApiError
    mockLLMClient.generate.mockRejectedValueOnce(apiError);

    // Mock findModelGroup to return null (no group)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(null);

    // Call the function
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [testModels[0]],
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify the error is captured in the response
    expect(result.queryResults.responses[0].error).toBe('API communication error');
    expect((result.queryResults.responses[0] as any).errorCategory).toBe(errorCategories.API);
    expect(result.queryResults.statuses['mock:mock-model'].status).toBe('error');
  });

  it('should receive empty responses array for empty models array', async () => {
    // Call the function with an empty models array
    const result = await _executeQueries({
      spinner: mockSpinner,
      config: sampleConfig,
      models: [], // Empty models array
      combinedContent: 'Test prompt',
      options: {
        input: 'test-prompt.txt',
      },
      llmClient: mockLLMClient,
    });

    // Verify the response contains an empty responses array
    expect(result.queryResults.responses).toEqual([]);
    expect(Object.keys(result.queryResults.statuses).length).toBe(0);
  });
});
