/**
 * Integration tests for runThinktank.ts
 * 
 * This file tests the runThinktank function using mocked interfaces instead of
 * directly mocking the helper functions.
 */
import { runThinktank, RunOptions } from '../runThinktank';
import { ConfigError, ApiError, FileSystemError } from '../../core/errors';
import * as helpers from '../runThinktankHelpers';
import * as nameGenerator from '../../utils/nameGenerator';
import { FileWriteStatus } from '../outputHandler';
import { FileSystem, ConfigManagerInterface, LLMClient } from '../../core/interfaces';
import { AppConfig, LLMResponse } from '../../core/types';
import { Stats } from 'fs';
import { InputSourceType } from '../inputHandler';
import { ExecuteQueriesResult, ProcessInputResult } from '../runThinktankTypes';

// Create mock implementations of the interfaces
const mockFileSystem: jest.Mocked<FileSystem> = {
  readFileContent: jest.fn().mockResolvedValue('Test prompt content'),
  writeFile: jest.fn().mockResolvedValue(undefined),
  fileExists: jest.fn().mockResolvedValue(true),
  mkdir: jest.fn().mockResolvedValue(undefined),
  readdir: jest.fn().mockResolvedValue(['file1.txt', 'file2.txt']),
  stat: jest.fn().mockResolvedValue({ 
    isFile: () => true,
    isDirectory: () => false 
  } as unknown as Stats),
  access: jest.fn().mockResolvedValue(undefined),
  getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
  getConfigFilePath: jest.fn().mockResolvedValue('/mock/config/file.json')
};

// Sample AppConfig
const mockConfig: AppConfig = {
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

const mockConfigManager: jest.Mocked<ConfigManagerInterface> = {
  loadConfig: jest.fn().mockResolvedValue(mockConfig),
  saveConfig: jest.fn().mockResolvedValue(undefined),
  getActiveConfigPath: jest.fn().mockResolvedValue('/mock/config/file.json'),
  getDefaultConfigPath: jest.fn().mockReturnValue('/mock/default/config.json')
};

// Create a response that includes the required configKey property
const mockLLMResponse: LLMResponse & { configKey: string } = {
  provider: 'mock',
  modelId: 'mock-model',
  text: 'Mock response for prompt: Test prompt',
  configKey: 'mock:mock-model',
  metadata: {
    usage: { total_tokens: 10 },
    model: 'mock-model',
    id: 'mock-response-id',
  }
};

const mockLLMClient: jest.Mocked<LLMClient> = {
  generate: jest.fn().mockResolvedValue(mockLLMResponse)
};

// Mock the concrete class constructors
jest.mock('../../core/FileSystem', () => ({
  ConcreteFileSystem: jest.fn(() => mockFileSystem)
}));

jest.mock('../../core/ConcreteConfigManager', () => ({
  ConcreteConfigManager: jest.fn(() => mockConfigManager)
}));

jest.mock('../../core/LLMClient', () => ({
  ConcreteLLMClient: jest.fn(() => mockLLMClient)
}));

// Mock the model selector module to bypass API key validation and return default mocked models
jest.mock('../modelSelector', () => {
  return {
    selectModels: jest.fn().mockImplementation(() => {
      // Return a valid selection result with our test models
      return {
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
      };
    })
  };
});

// Store original module path for restoration
const helpersPath = require.resolve('../runThinktankHelpers');
const nameGeneratorPath = require.resolve('../../utils/nameGenerator');
const oraPath = require.resolve('ora');

// Spy on helper functions instead of completely mocking them
// This allows us to verify that the helper functions are called with the correct arguments
// including our mocked interface instances
const setupWorkflowSpy = jest.spyOn(helpers, '_setupWorkflow');
const processInputSpy = jest.spyOn(helpers, '_processInput');
const selectModelsSpy = jest.spyOn(helpers, '_selectModels');
const executeQueriesSpy = jest.spyOn(helpers, '_executeQueries');
const processOutputSpy = jest.spyOn(helpers, '_processOutput');
const logCompletionSummarySpy = jest.spyOn(helpers, '_logCompletionSummary');
const handleWorkflowErrorSpy = jest.spyOn(helpers, '_handleWorkflowError');

// Setup mock implementations for these spies so they return values without executing real logic
// But we'll still be able to verify that they were called with the mocked interfaces
const mockSetupResult = {
  config: mockConfig,
  friendlyRunName: 'clever-meadow',
  outputDirectoryPath: '/mock/output/clever-meadow'
};

const mockInputResult: ProcessInputResult = {
  inputResult: {
    content: 'Test prompt content',
    sourceType: 'file' as InputSourceType,
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
  combinedContent: 'Test prompt with context',
  contextFiles: []
};

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

const mockQueryResult: ExecuteQueriesResult = {
  queryResults: {
    responses: [mockLLMResponse],
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
};

const mockOutputResult = {
  fileOutputResult: {
    outputDirectory: '/mock/output/clever-meadow',
    files: [{ 
      modelKey: 'mock:mock-model', 
      filename: 'mock-model.md', 
      status: 'success' as FileWriteStatus,
      filePath: '/mock/output/clever-meadow/mock-model.md'
    }],
    succeededWrites: 1,
    failedWrites: 0,
    timing: { startTime: 1, endTime: 2, durationMs: 1 }
  },
  consoleOutput: 'Mock console output'
};

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

jest.mock('../../utils/nameGenerator');

describe('runThinktank with Interface Mocks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup spies to return fake values rather than executing real functions
    setupWorkflowSpy.mockResolvedValue(mockSetupResult);
    processInputSpy.mockResolvedValue(mockInputResult);
    selectModelsSpy.mockReturnValue(mockSelectionResult);
    executeQueriesSpy.mockResolvedValue(mockQueryResult);
    processOutputSpy.mockResolvedValue(mockOutputResult);
    logCompletionSummarySpy.mockReturnValue({});
    
    // Make error handler re-throw for error testing
    handleWorkflowErrorSpy.mockImplementation(({ error }) => { throw error; });
    
    // Mock nameGenerator for consistent run names
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
  });
  
  afterEach(() => {
    jest.restoreAllMocks();
  });
  
  // Restore all mocked modules after tests
  afterAll(() => {    
    // Clear module cache to ensure fresh imports
    delete require.cache[helpersPath];
    delete require.cache[nameGeneratorPath];
    delete require.cache[oraPath];
  });

  it('should instantiate and pass interface implementations to helper functions', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
    };

    const result = await runThinktank(options);
    
    // Verify helpers were called with the correct mocked interfaces
    expect(setupWorkflowSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        configManager: mockConfigManager,
        fileSystem: mockFileSystem,
        spinner: expect.any(Object),
        options
      })
    );
    
    expect(processInputSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        fileSystem: mockFileSystem,
        spinner: expect.any(Object),
        input: 'test-prompt.txt'
      })
    );
    
    expect(executeQueriesSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        llmClient: mockLLMClient,
        spinner: expect.any(Object),
        models: expect.any(Array),
        combinedContent: expect.any(String)
      })
    );
    
    expect(processOutputSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        fileSystem: mockFileSystem,
        spinner: expect.any(Object),
        queryResults: expect.any(Object),
        outputDirectoryPath: '/mock/output/clever-meadow'
      })
    );
    
    // Verify the result is the formatted console output
    expect(result).toBe('Mock console output');
  });

  it('should verify that the LLMClient.generate method is called', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Override the executeQueries spy to call the real LLMClient.generate method
    // We need to ensure the returned mock has the correct type
    executeQueriesSpy.mockImplementationOnce(async ({ llmClient, combinedContent }) => {
      // Actually call the mock LLMClient to verify it's properly passed
      await llmClient.generate(combinedContent, 'mock:mock-model', {});
      return mockQueryResult;
    });

    await runThinktank(options);
    
    // Verify that the LLMClient.generate method was called
    expect(mockLLMClient.generate).toHaveBeenCalledTimes(1);
    expect(mockLLMClient.generate).toHaveBeenCalledWith(
      'Test prompt with context',  // combinedContent from mockInputResult
      'mock:mock-model',
      expect.any(Object),
      undefined
    );
  });

  it('should verify that FileSystem.writeFile is called by processOutput', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Override the processOutput spy to call the real FileSystem.writeFile method
    processOutputSpy.mockImplementationOnce(async ({ fileSystem }) => {
      // Actually call the mock FileSystem to verify it's properly passed
      await fileSystem.writeFile('/mock/output/clever-meadow/mock-model.md', 'File content');
      return mockOutputResult;
    });

    await runThinktank(options);
    
    // Verify that the FileSystem.writeFile method was called
    expect(mockFileSystem.writeFile).toHaveBeenCalledTimes(1);
    expect(mockFileSystem.writeFile).toHaveBeenCalledWith(
      '/mock/output/clever-meadow/mock-model.md',
      'File content'
    );
  });

  it('should verify that ConfigManager.loadConfig is called by setupWorkflow', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      configPath: '/custom/config.json'
    };

    // Override the setupWorkflow spy to call the real ConfigManager.loadConfig method
    setupWorkflowSpy.mockImplementationOnce(async ({ configManager, options }) => {
      // Actually call the mock ConfigManager to verify it's properly passed
      await configManager.loadConfig({ configPath: options.configPath });
      return mockSetupResult;
    });

    await runThinktank(options);
    
    // Verify that the ConfigManager.loadConfig method was called
    expect(mockConfigManager.loadConfig).toHaveBeenCalledTimes(1);
    expect(mockConfigManager.loadConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        configPath: '/custom/config.json'
      })
    );
  });

  it('should handle errors from the FileSystem interface', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Setup FileSystem method to throw an error
    const fileError = new FileSystemError('File not found');
    mockFileSystem.readFileContent.mockRejectedValueOnce(fileError);
    
    // Make processInput use the real FileSystem that will throw
    processInputSpy.mockImplementationOnce(async ({ fileSystem }) => {
      if (fileSystem) {
        await fileSystem.readFileContent('test-prompt.txt');
      }
      return mockInputResult;
    });

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    
    // Verify the error handler was called
    expect(handleWorkflowErrorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        error: fileError
      })
    );
  });

  it('should handle errors from the ConfigManager interface', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Setup ConfigManager method to throw an error
    const configError = new ConfigError('Config loading failed');
    mockConfigManager.loadConfig.mockRejectedValueOnce(configError);
    
    // Make setupWorkflow use the real ConfigManager that will throw
    setupWorkflowSpy.mockImplementationOnce(async ({ configManager }) => {
      await configManager.loadConfig();
      return mockSetupResult;
    });

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(ConfigError);
    
    // Verify the error handler was called
    expect(handleWorkflowErrorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        error: configError
      })
    );
  });

  it('should handle errors from the LLMClient interface', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Setup LLMClient method to throw an error
    const apiError = new ApiError('API call failed');
    mockLLMClient.generate.mockRejectedValueOnce(apiError);
    
    // Make executeQueries use the real LLMClient that will throw
    executeQueriesSpy.mockImplementationOnce(async ({ llmClient, combinedContent }) => {
      await llmClient.generate(combinedContent, 'mock:mock-model', {});
      return mockQueryResult;
    });

    // Should propagate the error
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    
    // Verify the error handler was called
    expect(handleWorkflowErrorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        error: apiError
      })
    );
  });

  it('should return early when no models are selected', async () => {
    // Override the standard _selectModels implementation for this test only
    // To return an empty model set
    const emptyModelSelectionResult: any = {
      models: [],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: [],
      modeDescription: 'All enabled models'
    };
    // Add self-reference for backward compatibility
    emptyModelSelectionResult.modelSelectionResult = emptyModelSelectionResult;
    
    // Import the actual modelSelector module to override just its method
    const modelSelector = require('../modelSelector');
    const originalSelectModels = modelSelector.selectModels;
    
    // Replace with our empty model selection implementation
    modelSelector.selectModels = jest.fn().mockReturnValue(emptyModelSelectionResult);
    
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    const result = await runThinktank(options);
    
    // Verify that the execute queries helper was not called
    expect(executeQueriesSpy).not.toHaveBeenCalled();
    
    // Verify that a warning message was returned
    expect(result).toContain('No models available');
    
    // Restore the original mock implementation
    modelSelector.selectModels = originalSelectModels;
  });

  it('should pass custom options through to the appropriate helpers', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      contextPaths: ['src/file1.js'],
      // Using mock:mock-model for model which exists in our mockConfig
      models: ['mock:mock-model'],
      systemPrompt: 'Custom system prompt',
      enableThinking: true,
      includeMetadata: true,
      useColors: true,
      output: '/custom/output'
    };

    await runThinktank(options);
    
    // Verify options are passed to setupWorkflow
    expect(setupWorkflowSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          output: '/custom/output'
        })
      })
    );
    
    // Verify contextPaths are passed to processInput
    expect(processInputSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        contextPaths: ['src/file1.js']
      })
    );
    
    // Verify models are passed to selectModels
    expect(selectModelsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          models: ['mock:mock-model']
        })
      })
    );
    
    // Verify system prompt and enableThinking are passed to executeQueries
    expect(executeQueriesSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          systemPrompt: 'Custom system prompt',
          enableThinking: true
        })
      })
    );
    
    // Verify includeMetadata is passed to processOutput
    expect(processOutputSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        options: expect.objectContaining({
          includeMetadata: true,
          useColors: true
        })
      })
    );
  });
});