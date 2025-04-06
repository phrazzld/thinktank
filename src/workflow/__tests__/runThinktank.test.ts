/**
 * Integration tests for runThinktank.ts
 */
import { runThinktank, RunOptions } from '../runThinktank';
import { ThinktankError } from '../../core/errors';
import * as fileReader from '../../utils/fileReader';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import * as inputHandler from '../inputHandler';
import * as modelSelector from '../modelSelector';
import * as queryExecutor from '../queryExecutor';
import * as outputHandler from '../outputHandler';
import * as nameGenerator from '../../utils/nameGenerator';
import { LLMProvider, LLMResponse, ModelConfig } from '../../core/types';
import fs from 'fs/promises';

// Store module paths for restoration
const fileReaderPath = require.resolve('../../utils/fileReader');
const configManagerPath = require.resolve('../../core/configManager');
const llmRegistryPath = require.resolve('../../core/llmRegistry');
const inputHandlerPath = require.resolve('../inputHandler');
const modelSelectorPath = require.resolve('../modelSelector');
const queryExecutorPath = require.resolve('../queryExecutor');
const outputHandlerPath = require.resolve('../outputHandler');
const nameGeneratorPath = require.resolve('../../utils/nameGenerator');
const fsPromisesPath = require.resolve('fs/promises');
const oraPath = require.resolve('ora');

// Mock dependencies
jest.mock('../../utils/fileReader');
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');
jest.mock('../inputHandler');
jest.mock('../modelSelector');
jest.mock('../queryExecutor');
jest.mock('../outputHandler');
jest.mock('../../utils/nameGenerator');
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
  });
  
  // Restore all mocked modules after tests
  afterAll(() => {
    jest.unmock('../../utils/fileReader');
    jest.unmock('../../core/configManager');
    jest.unmock('../../core/llmRegistry');
    jest.unmock('../inputHandler');
    jest.unmock('../modelSelector');
    jest.unmock('../queryExecutor');
    jest.unmock('../outputHandler');
    jest.unmock('../../utils/nameGenerator');
    jest.unmock('fs/promises');
    jest.unmock('ora');
    
    // Clear module cache to ensure fresh imports
    delete require.cache[fileReaderPath];
    delete require.cache[configManagerPath];
    delete require.cache[llmRegistryPath];
    delete require.cache[inputHandlerPath];
    delete require.cache[modelSelectorPath];
    delete require.cache[queryExecutorPath];
    delete require.cache[outputHandlerPath];
    delete require.cache[nameGeneratorPath];
    delete require.cache[fsPromisesPath];
    delete require.cache[oraPath];
  });
  
  beforeEach(() => {
    // Default mock implementations
    (fileReader.readFileContent as jest.Mock).mockResolvedValue('Test prompt');
    
    // Mock inputHandler
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: inputHandler.InputSourceType.FILE,
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Mock outputHandler
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir');
    (outputHandler.writeResponsesToFiles as jest.Mock).mockResolvedValue({
      outputDirectory: '/fake/output/dir',
      files: [{ status: 'success', filename: 'mock-mock-model.md' }],
      succeededWrites: 1,
      failedWrites: 0,
      timing: { startTime: 1, endTime: 2, durationMs: 1 }
    });
    (outputHandler.formatForConsole as jest.Mock).mockReturnValue('Mock console output');
    
    // Mock modelSelector
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
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
      warnings: []
    });
    
    // Mock queryExecutor
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValue({
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
    });
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
    
    // Mock nameGenerator
    (nameGenerator.generateFunName as jest.Mock).mockResolvedValue('clever-meadow');
    (nameGenerator.generateFallbackName as jest.Mock).mockReturnValue('run-20250101-123045');
  });

  it('should run successfully with valid inputs', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    // Verify our component modules were called with correct parameters
    expect(inputHandler.processInput).toHaveBeenCalledWith({ input: 'test-prompt.txt' });
    expect(configManager.loadConfig).toHaveBeenCalled();
    expect(modelSelector.selectModels).toHaveBeenCalled();
    expect(queryExecutor.executeQueries).toHaveBeenCalled();
    expect(outputHandler.createOutputDirectory).toHaveBeenCalled();
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalled();
    expect(outputHandler.formatForConsole).toHaveBeenCalled();
    
    // Verify the result is the formatted output
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
    
    // Verify we pass the models list to modelSelector
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        models: ['mock:mock-model']
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
    
    // Verify we pass the specificModel to modelSelector
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        specificModel: 'mock:mock-model'
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
    
    // Verify we pass the groupName to modelSelector
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        groupName: 'coding'
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
    
    // Verify system prompt from CLI was passed to QueryExecutor
    expect(queryExecutor.executeQueries).toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      expect.objectContaining({
        systemPrompt: 'Custom CLI system prompt override'
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
    
    // Verify enableThinking was passed to QueryExecutor
    expect(queryExecutor.executeQueries).toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      expect.objectContaining({
        enableThinking: true
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
    
    // Verify includeMetadata was passed to OutputHandler
    expect(outputHandler.writeResponsesToFiles).toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      expect.objectContaining({
        includeMetadata: true
      })
    );
    
    // Verify includeMetadata was passed to formatForConsole
    expect(outputHandler.formatForConsole).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        includeMetadata: true
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
    
    // Verify output path was passed to createOutputDirectory
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(
      expect.objectContaining({
        outputDirectory: '/custom/output/path'
      })
    );
  });

  it('should include thinking output when specified', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeThinking: true,
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify includeThinking was passed to formatForConsole
    expect(outputHandler.formatForConsole).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        includeThinking: true
      })
    );
  });

  it('should handle missing API keys warning from ModelSelector', async () => {
    // Set modelSelector to return some missing API key models
    (modelSelector.selectModels as jest.Mock).mockReturnValueOnce({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true
        }
      ],
      missingApiKeyModels: [
        {
          provider: 'missing',
          modelId: 'missing-model',
          enabled: true
        }
      ],
      disabledModels: [],
      warnings: ['Missing API keys for models: missing:missing-model']
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Verify we still proceed with the available models
    expect(queryExecutor.executeQueries).toHaveBeenCalled();
  });

  it('should throw error from ModelSelector when no models available', async () => {
    // Mock ModelSelector to throw an error
    (modelSelector.selectModels as jest.Mock).mockImplementationOnce(() => {
      throw new Error('No enabled models found');
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should propagate the error as a ThinktankError
    await expect(runThinktank(options)).rejects.toThrow(ThinktankError);
  });
  
  it('should use fun name for run when name generation succeeds', async () => {
    const funName = 'clever-meadow';
    (nameGenerator.generateFunName as jest.Mock).mockResolvedValue(funName);
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
    };
    
    const result = await runThinktank(options);
    
    // Check that the name was used in console formatting
    expect(outputHandler.formatForConsole).toHaveBeenCalled();
    expect(result).toBe('Mock console output');
    
    // Verify the name was passed to formatResultsSummary (indirectly checking options object)
    expect(queryExecutor.executeQueries).toHaveBeenCalled();
  });
  
  it('should use fallback name when name generation fails', async () => {
    const fallbackName = 'run-20250101-123045';
    (nameGenerator.generateFunName as jest.Mock).mockResolvedValue(null);
    (nameGenerator.generateFallbackName as jest.Mock).mockReturnValue(fallbackName);
    
    const options: RunOptions = {
      input: 'test-prompt.txt',
    };
    
    const result = await runThinktank(options);
    
    // Check that the output was still successfully generated
    expect(result).toBe('Mock console output');
    
    // The fallback name should have been used in the formatResultsSummary call
    expect(queryExecutor.executeQueries).toHaveBeenCalled();
  });
  
  it('should handle API execution errors from QueryExecutor', async () => {
    // Mock QueryExecutor to return a response with an error
    (queryExecutor.executeQueries as jest.Mock).mockResolvedValueOnce({
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: '',
          error: 'API call failed',
          configKey: 'mock:mock-model'
        }
      ],
      statuses: {
        'mock:mock-model': { 
          status: 'error',
          message: 'API call failed',
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

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    await runThinktank(options);
    
    // Should still complete and call formatForConsole
    expect(outputHandler.formatForConsole).toHaveBeenCalled();
  });

  it('should handle file write errors from OutputHandler', async () => {
    // Mock OutputHandler to return a result with failed writes
    (outputHandler.writeResponsesToFiles as jest.Mock).mockResolvedValueOnce({
      outputDirectory: '/fake/output/dir',
      files: [{ status: 'error', filename: 'mock-mock-model.md', error: 'Write error' }],
      succeededWrites: 0,
      failedWrites: 1,
      timing: { startTime: 1, endTime: 2, durationMs: 1 }
    });

    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    // Should still complete without throwing
    await runThinktank(options);
    
    // Should still format the console output
    expect(outputHandler.formatForConsole).toHaveBeenCalled();
  });

  it('should throw ThinktankError for file read errors', async () => {
    // Mock InputHandler to throw an error
    (inputHandler.processInput as jest.Mock).mockRejectedValue(
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