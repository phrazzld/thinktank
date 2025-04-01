/**
 * Integration tests for runThinktank.ts
 */
import { runThinktank, ThinktankError, RunOptions } from '../runThinktank';
import * as fileReader from '../../utils/fileReader';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import { LLMProvider, LLMResponse, ModelConfig, SystemPrompt } from '../../core/types';
import fs from 'fs/promises';

// Mock dependencies
jest.mock('../../utils/fileReader');
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');
jest.mock('fs/promises');
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

// Mock provider for testing
class MockProvider implements LLMProvider {
  providerId = 'mock';
  
  async generate(prompt: string, modelId: string): Promise<LLMResponse> {
    return {
      provider: this.providerId,
      modelId,
      text: `Mock response for prompt: ${prompt}`,
      metadata: {
        usage: { total_tokens: 10 },
        model: modelId,
        id: 'mock-response-id',
      }
    };
  }
}

describe('runThinktank', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Default mock implementations
    (fileReader.readFileContent as jest.Mock).mockResolvedValue('Test prompt');
    (configManager.loadConfig as jest.Mock).mockResolvedValue({
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
        },
        coding: {
          name: 'coding',
          systemPrompt: { text: 'You are a coding assistant.' },
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
    });
    (configManager.getEnabledModels as jest.Mock).mockImplementation((config) => {
      return config.models.filter((model: ModelConfig) => model.enabled);
    });
    (configManager.filterModels as jest.Mock).mockImplementation((config, modelFilter) => {
      return config.models.filter((model: ModelConfig) => 
        `${model.provider}:${model.modelId}`.includes(modelFilter)
      );
    });
    (configManager.validateModelApiKeys as jest.Mock).mockReturnValue({
      missingKeyModels: []
    });
    
    // Mock llmRegistry.getProvider
    const mockProvider = new MockProvider();
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId) => {
      if (providerId === 'mock') {
        return mockProvider;
      }
      return null;
    });

    // Mock fs
    (fs.writeFile as jest.Mock).mockResolvedValue(undefined);
    (fs.mkdir as jest.Mock).mockResolvedValue(undefined);
  });

  it('should run successfully with valid inputs', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    // Verify outputs
    expect(fileReader.readFileContent).toHaveBeenCalledWith('test-prompt.txt');
    expect(configManager.loadConfig).toHaveBeenCalled();
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('mock');
    expect(result).toContain('Mock response for prompt: Test prompt');
  });

  it('should handle specified models correctly', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      models: ['mock:mock-model'],
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    expect(configManager.filterModels).toHaveBeenCalledWith(
      expect.anything(),
      'mock:mock-model'
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
    
    // Verify we're handling specificModel correctly
    expect(configManager.loadConfig).toHaveBeenCalled();
  });
  
  it('should handle specific group name parameter correctly', async () => {
    // Mock group-related functions
    (configManager.getEnabledModelsFromGroups as jest.Mock).mockImplementation(
      (_config, groupNames) => {
        if (groupNames.includes('coding')) {
          return [{
            provider: 'mock',
            modelId: 'mock-model',
            enabled: true
          }];
        }
        return [];
      }
    );
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      groupName: 'coding',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify getEnabledModelsFromGroups was called with the group name
    expect(configManager.getEnabledModelsFromGroups).toHaveBeenCalledWith(
      expect.anything(),
      ['coding']
    );
  });
  
  it('should select system prompt from CLI override when provided', async () => {
    // Save the original implementation
    const originalGenerate = (llmRegistry.getProvider('mock') as LLMProvider).generate;
    
    // Replace with a mock that captures the system prompt
    let capturedSystemPrompt: SystemPrompt | undefined;
    (llmRegistry.getProvider as jest.Mock).mockImplementation(() => ({
      providerId: 'mock',
      generate: (_prompt: string, modelId: string, _options: unknown, systemPrompt?: SystemPrompt) => {
        capturedSystemPrompt = systemPrompt;
        return Promise.resolve({
          provider: 'mock',
          modelId,
          text: 'Mock response',
          metadata: {}
        });
      }
    }));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      systemPrompt: 'Custom CLI system prompt override',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify system prompt from CLI was used
    expect(capturedSystemPrompt).toBeDefined();
    expect(capturedSystemPrompt?.text).toBe('Custom CLI system prompt override');
    expect(capturedSystemPrompt?.metadata?.source).toBe('cli-override');
    
    // Restore the original implementation if needed
    if (originalGenerate) {
      (llmRegistry.getProvider('mock') as LLMProvider).generate = originalGenerate;
    }
  });
  
  it('should select system prompt from model when available', async () => {
    // Set up config with a model that has a system prompt
    (configManager.loadConfig as jest.Mock).mockResolvedValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
          systemPrompt: {
            text: 'Model-specific system prompt',
            metadata: { source: 'model-config' }
          }
        }
      ]
    });
    
    // Save the original implementation
    const originalGenerate = (llmRegistry.getProvider('mock') as LLMProvider)?.generate;
    
    // Replace with a mock that captures the system prompt
    let capturedSystemPrompt: SystemPrompt | undefined;
    (llmRegistry.getProvider as jest.Mock).mockImplementation(() => ({
      providerId: 'mock',
      generate: (_prompt: string, modelId: string, _options: unknown, systemPrompt?: SystemPrompt) => {
        capturedSystemPrompt = systemPrompt;
        return Promise.resolve({
          provider: 'mock',
          modelId,
          text: 'Mock response',
          metadata: {}
        });
      }
    }));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify model-specific system prompt was used
    expect(capturedSystemPrompt).toBeDefined();
    expect(capturedSystemPrompt?.text).toBe('Model-specific system prompt');
    
    // Restore the original implementation if needed
    if (originalGenerate) {
      (llmRegistry.getProvider('mock') as LLMProvider).generate = originalGenerate;
    }
  });
  
  it('should select system prompt from specified group', async () => {
    // Set up config with groups
    (configManager.loadConfig as jest.Mock).mockResolvedValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true
        }
      ],
      groups: {
        coding: {
          name: 'coding',
          systemPrompt: {
            text: 'You are a coding assistant.',
            metadata: { source: 'group-config' }
          },
          models: [
            {
              provider: 'mock',
              modelId: 'mock-model',
              enabled: true
            }
          ]
        }
      }
    });
    
    // Mock group-related functions
    (configManager.getEnabledModelsFromGroups as jest.Mock).mockImplementation(
      (_config, groupNames) => {
        if (groupNames.includes('coding')) {
          return [{
            provider: 'mock',
            modelId: 'mock-model',
            enabled: true
          }];
        }
        return [];
      }
    );
    
    // Save the original implementation
    const originalGenerate = (llmRegistry.getProvider('mock') as LLMProvider)?.generate;
    
    // Replace with a mock that captures the system prompt
    let capturedSystemPrompt: SystemPrompt | undefined;
    (llmRegistry.getProvider as jest.Mock).mockImplementation(() => ({
      providerId: 'mock',
      generate: (_prompt: string, modelId: string, _options: unknown, systemPrompt?: SystemPrompt) => {
        capturedSystemPrompt = systemPrompt;
        return Promise.resolve({
          provider: 'mock',
          modelId,
          text: 'Mock response',
          metadata: {}
        });
      }
    }));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      groupName: 'coding',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify group-specific system prompt was used
    expect(capturedSystemPrompt).toBeDefined();
    expect(capturedSystemPrompt?.text).toBe('You are a coding assistant.');
    
    // Restore the original implementation if needed
    if (originalGenerate) {
      (llmRegistry.getProvider('mock') as LLMProvider).generate = originalGenerate;
    }
  });
  
  it('should use default system prompt when no others are available', async () => {
    // Set up config with no system prompts
    (configManager.loadConfig as jest.Mock).mockResolvedValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true
          // No systemPrompt property
        }
      ],
      groups: {
        // No groups defined
      }
    });
    
    // Mock findModelGroup to return undefined (no group found)
    (configManager.findModelGroup as jest.Mock).mockReturnValue(undefined);
    
    // Save the original implementation
    const originalGenerate = (llmRegistry.getProvider('mock') as LLMProvider)?.generate;
    
    // Replace with a mock that captures the system prompt
    let capturedSystemPrompt: SystemPrompt | undefined;
    (llmRegistry.getProvider as jest.Mock).mockImplementation(() => ({
      providerId: 'mock',
      generate: (_prompt: string, modelId: string, _options: unknown, systemPrompt?: SystemPrompt) => {
        capturedSystemPrompt = systemPrompt;
        return Promise.resolve({
          provider: 'mock',
          modelId,
          text: 'Mock response',
          metadata: {}
        });
      }
    }));
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify default system prompt was used
    expect(capturedSystemPrompt).toBeDefined();
    expect(capturedSystemPrompt?.text).toContain('You are a helpful, accurate');
    expect(capturedSystemPrompt?.metadata?.source).toBe('default-fallback');
    
    // Restore the original implementation if needed
    if (originalGenerate) {
      (llmRegistry.getProvider('mock') as LLMProvider).generate = originalGenerate;
    }
  });

  it('should create output directory and write files when output path is provided', async () => {
    // Reset mocks to track calls
    (fs.mkdir as jest.Mock).mockClear();
    (fs.writeFile as jest.Mock).mockClear();
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      output: 'output-dir',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify directory creation was attempted
    expect(fs.mkdir).toHaveBeenCalled();
    
    // Verify file write was attempted for the model
    expect(fs.writeFile).toHaveBeenCalled();
    
    // Verify content format and path
    const writeFileCalls = (fs.writeFile as jest.Mock).mock.calls;
    expect(writeFileCalls.length).toBeGreaterThan(0);
    
    // Check that the content has Markdown format
    const [_, content] = writeFileCalls[0];
    expect(typeof content).toBe('string');
    expect(content).toContain('# mock:mock-model');
    expect(content).toContain('## Response');
  });

  it('should handle missing API keys gracefully', async () => {
    (configManager.validateModelApiKeys as jest.Mock).mockReturnValue({
      missingKeyModels: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        }
      ]
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    expect(result).toContain('No models with valid API keys available');
  });

  it('should handle no enabled models gracefully', async () => {
    (configManager.getEnabledModels as jest.Mock).mockReturnValue([]);

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    expect(result).toContain('No enabled models found in configuration');
  });
  
  it('should throw error for invalid group name', async () => {
    // Set up a custom mock for this test to simulate group not found
    (configManager.loadConfig as jest.Mock).mockResolvedValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true
        }
      ],
      groups: {
        // Only has default group, not the nonexistent-group we'll request
        default: {
          name: 'default',
          systemPrompt: { text: 'You are a helpful assistant.' },
          models: []
        }
      }
    });
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      groupName: 'nonexistent-group',
      includeMetadata: false,
      useColors: false,
    };

    // Should throw an error
    await expect(runThinktank(options)).rejects.toThrow(
      'Group "nonexistent-group" not found in configuration'
    );
  });
  
  it('should throw error for invalid specific model', async () => {
    // Set up a custom mock for this test to simulate model not found
    (configManager.loadConfig as jest.Mock).mockResolvedValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true
        }
      ]
    });
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      specificModel: 'invalid:model',
      includeMetadata: false,
      useColors: false,
    };

    // Should throw an error
    await expect(runThinktank(options)).rejects.toThrow(
      'Model "invalid:model" not found in configuration'
    );
  });

  it('should handle provider not found gracefully', async () => {
    // Mock getProvider to return null for missing provider
    (llmRegistry.getProvider as jest.Mock).mockReturnValue(null);
    
    // Mock getProviderIds for the dynamic import in error handling
    jest.mock('../../organisms/llmRegistry', () => ({
      getProvider: jest.fn().mockReturnValue(null),
      getProviderIds: jest.fn().mockReturnValue(['openai', 'anthropic'])
    }), { virtual: true });
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Test passes if no exception is thrown
  });

  it('should handle LLM errors gracefully', async () => {
    // Mock provider that throws an error
    (llmRegistry.getProvider as jest.Mock).mockImplementation(() => {
      return {
        providerId: 'error',
        generate: () => Promise.reject(new Error('API error'))
      };
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    expect(result).toContain('API error');
  });

  it('should throw ThinktankError for file read errors', async () => {
    (fileReader.readFileContent as jest.Mock).mockRejectedValue(
      new Error('File not found')
    );

    const options: RunOptions = {
      input: 'nonexistent.txt',
      includeMetadata: false,
      useColors: false,
    };

    await expect(runThinktank(options)).rejects.toThrow(ThinktankError);
  });
});