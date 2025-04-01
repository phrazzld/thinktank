/**
 * Tests specifically for error handling in runThinktank
 */
import { runThinktank, ThinktankError, RunOptions } from '../runThinktank';
import * as fileReader from '../../utils/fileReader';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import { LLMProvider, LLMResponse, ModelConfig } from '../../core/types';
import fs from 'fs/promises';
import { errorCategories } from '../../utils/consoleUtils';

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

class MockProvider implements LLMProvider {
  providerId = 'mock';
  
  constructor(private throwError = false) {}
  
  async generate(prompt: string, modelId: string): Promise<LLMResponse> {
    if (this.throwError) {
      throw new Error('API Error: Failed to generate response');
    }
    
    return {
      provider: this.providerId,
      modelId,
      text: `Mock response for: ${prompt}`,
      metadata: { tokens: 10 }
    };
  }
}

describe('runThinktank Error Handling', () => {
  // Default mock config
  const mockConfig = {
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true
      }
    ],
    defaultGroup: 'default',
    groups: {
      default: {
        name: 'default',
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            enabled: true
          }
        ]
      }
    }
  };
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Default mock implementations
    (fileReader.readFileContent as jest.Mock).mockResolvedValue('Test prompt');
    (configManager.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (configManager.getEnabledModels as jest.Mock).mockImplementation((config) => {
      return config.models.filter((model: ModelConfig) => model.enabled);
    });
    (configManager.getEnabledModelsFromGroups as jest.Mock).mockImplementation((config) => {
      return config.models.filter((model: ModelConfig) => model.enabled);
    });
    (configManager.findModelGroup as jest.Mock).mockReturnValue({
      groupName: 'default',
      systemPrompt: mockConfig.groups.default.systemPrompt
    });
    (configManager.validateModelApiKeys as jest.Mock).mockReturnValue({
      missingKeyModels: []
    });
    (llmRegistry.getProvider as jest.Mock).mockReturnValue(new MockProvider());
    (fs.mkdir as jest.Mock).mockResolvedValue(undefined);
    (fs.writeFile as jest.Mock).mockResolvedValue(undefined);
  });

  it('should throw ThinktankError when prompt file cannot be read', async () => {
    // Setup file reading to fail
    (fileReader.readFileContent as jest.Mock).mockRejectedValueOnce(
      new Error('ENOENT: File not found')
    );
    
    // Call with valid options
    const options: RunOptions = {
      input: 'nonexistent.txt'
    };
    
    // Expect it to throw with a helpful error
    await expect(runThinktank(options)).rejects.toThrow(ThinktankError);
    await expect(runThinktank(options)).rejects.toThrow(/file/i);
  });
  
  it('should handle output directory creation error', async () => {
    // Setup directory creation to fail
    (fs.mkdir as jest.Mock).mockRejectedValueOnce(
      new Error('EACCES: Permission denied')
    );
    
    // Call with valid options
    const options: RunOptions = {
      input: 'prompt.txt',
      output: '/invalid/dir'
    };
    
    // Expect it to throw with a helpful error
    await expect(runThinktank(options)).rejects.toThrow(ThinktankError);
    await expect(runThinktank(options)).rejects.toThrow(/directory/i);
  });
  
  it('should handle invalid model format errors', async () => {
    // Call with invalid model format
    const options: RunOptions = {
      input: 'prompt.txt',
      specificModel: 'invalid-format'
    };
    
    // Expect it to throw with model format error
    await expect(runThinktank(options)).rejects.toThrow(/model format/i);
    
    const error = await runThinktank(options).catch(e => e);
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.category).toBe(errorCategories.CONFIG);
    expect(error.suggestions?.length).toBeGreaterThan(0);
  });
  
  it('should handle model not found errors', async () => {
    // Mock getProvider to return undefined (provider not found)
    (llmRegistry.getProvider as jest.Mock).mockReturnValueOnce(undefined);
    
    // Call with unknown provider
    const options: RunOptions = {
      input: 'prompt.txt',
      specificModel: 'unknown:model'
    };
    
    // Should continue execution but include error in results
    const result = await runThinktank(options);
    
    // Result should indicate the provider wasn't found
    expect(result).toContain('Provider');
    expect(result).toContain('not found');
  });
  
  it('should handle missing API key errors', async () => {
    // Setup API key validation to return missing keys
    (configManager.validateModelApiKeys as jest.Mock).mockReturnValueOnce({
      missingKeyModels: [
        { provider: 'mock', modelId: 'mock-model' }
      ]
    });
    
    // Call with valid options but missing API key
    const options: RunOptions = {
      input: 'prompt.txt'
    };
    
    // Should return a message about missing API keys
    const result = await runThinktank(options);
    expect(result).toContain('No models with valid API keys available');
  });
  
  it('should properly handle and report model API errors', async () => {
    // Mock provider to throw error
    (llmRegistry.getProvider as jest.Mock).mockReturnValueOnce(new MockProvider(true));
    
    // Call with valid options
    const options: RunOptions = {
      input: 'prompt.txt'
    };
    
    // Should complete but include error in results
    const result = await runThinktank(options);
    
    // Result should contain the API error
    expect(result).toContain('API Error');
  });
  
  it('should handle group not found errors', async () => {
    // Call with non-existent group
    const options: RunOptions = {
      input: 'prompt.txt',
      groupName: 'nonexistent-group'
    };
    
    // Mock findModelGroup to return undefined
    (configManager.getEnabledModelsFromGroups as jest.Mock).mockReturnValueOnce([]);
    
    // Should return a message about the group not found
    const result = await runThinktank(options);
    expect(result).toContain('No enabled models found in the specified group');
  });
  
  it('should handle file write errors gracefully', async () => {
    // Setup file writing to fail
    (fs.writeFile as jest.Mock).mockRejectedValueOnce(
      new Error('ENOSPC: No space left on device')
    );
    
    // Call with valid options
    const options: RunOptions = {
      input: 'prompt.txt'
    };
    
    // Should complete but include error in results
    const result = await runThinktank(options);
    
    // Result should indicate issue with writing files
    expect(result).toContain('error');
    expect(result).toContain('failed writes');
  });
});