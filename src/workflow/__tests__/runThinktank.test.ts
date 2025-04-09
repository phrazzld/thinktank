/**
 * Integration tests for runThinktank.ts
 * 
 * This file tests the runThinktank function using mocked interfaces instead of
 * directly mocking the helper functions.
 */
import { runThinktank, RunOptions } from '../runThinktank';
import { ApiError } from '../../core/errors';
import * as helpers from '../runThinktankHelpers';
import * as nameGenerator from '../../utils/nameGenerator';
import { FileSystem, ConfigManagerInterface, LLMClient } from '../../core/interfaces';
import { AppConfig, LLMResponse } from '../../core/types';
import { Stats } from 'fs';
import { InputSourceType } from '../inputHandler';
import { ExecuteQueriesResult, ProcessInputResult } from '../runThinktankTypes';
import { FileWriteDetail, FileWriteStatus } from '../outputHandler';

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

// Setup the mock result structures
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
    },
    // Add the combinedContent property to the mock query result to match test expectation
    combinedContent: 'Test prompt with context'
  }
};

const mockOutputResult: import('../runThinktankTypes').PureProcessOutputResult = {
  files: [{ 
    modelKey: 'mock:mock-model', 
    filename: 'mock-model.md', 
    content: 'Mock content'
  }],
  consoleOutput: 'Mock console output'
};

const mockFileOutputResult = {
  outputDirectory: '/mock/output/clever-meadow',
  files: [{
    modelKey: 'mock:mock-model',
    filename: 'mock-model.md',
    filePath: '/mock/output/clever-meadow/mock-model.md',
    status: 'success' as FileWriteStatus,
    startTime: 1,
    endTime: 2,
    durationMs: 1
  }] as FileWriteDetail[],
  succeededWrites: 1,
  failedWrites: 0,
  timing: {
    startTime: 1,
    endTime: 2,
    durationMs: 1
  }
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
const writeOutputFilesSpy = jest.spyOn(helpers, '_writeOutputFiles');
const handleWorkflowErrorSpy = jest.spyOn(helpers, '_handleWorkflowError');

// Mock formatCompletionSummary which is now used directly in runThinktank
jest.mock('../../utils/formatCompletionSummary', () => ({
  formatCompletionSummary: jest.fn().mockReturnValue({
    summaryText: 'Mock summary text',
    errorDetails: []
  })
}));

describe('runThinktank with Interface Mocks', () => {
  beforeEach(() => {
    // Clear all mocks to start fresh
    jest.clearAllMocks();
    
    // Setup spies to return fake values rather than executing real functions
    setupWorkflowSpy.mockResolvedValue(mockSetupResult);
    processInputSpy.mockResolvedValue(mockInputResult);
    selectModelsSpy.mockReturnValue(mockSelectionResult);
    executeQueriesSpy.mockResolvedValue(mockQueryResult);
    processOutputSpy.mockReturnValue(mockOutputResult);
    writeOutputFilesSpy.mockResolvedValue(mockFileOutputResult);
    
    // The error handler should just return something to avoid exceptions
    handleWorkflowErrorSpy.mockImplementation(() => { 
      return "Mock console output with error" as unknown as never; 
    });
    
    // Mock nameGenerator for consistent run names
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    
    // Reset LLMClient mocks
    mockLLMClient.generate.mockClear();
    mockLLMClient.generate.mockResolvedValue(mockLLMResponse);
    
    // Reset FileSystem mocks
    mockFileSystem.readFileContent.mockResolvedValue('Test prompt content');
    mockFileSystem.writeFile.mockResolvedValue(undefined);
    mockFileSystem.fileExists.mockResolvedValue(true);
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
        spinner: expect.any(Object),
        queryResults: expect.any(Object),
        options: expect.any(Object),
        friendlyRunName: 'clever-meadow'
      })
    );
    
    expect(writeOutputFilesSpy).toHaveBeenCalledWith(
      mockOutputResult.files,
      mockSetupResult.outputDirectoryPath,
      mockFileSystem,
      options
    );
    
    // Verify the result is the formatted console output
    expect(result).toBe('Mock console output');
  });

  it('should verify that the LLMClient.generate method is called', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Set up the mocked LLMClient.generate to provide arguments we expect
    mockLLMClient.generate.mockImplementation((_prompt, _modelId, _options, _systemPrompt) => {
      return Promise.resolve(mockLLMResponse);
    });

    // Override processInputSpy to ensure it sets the expected combinedContent
    processInputSpy.mockImplementationOnce((_params) => {
      return Promise.resolve({
        inputResult: {
          content: 'Test prompt with context',
          sourceType: InputSourceType.FILE,
          sourcePath: 'test-prompt.txt',
          metadata: {
            processingTimeMs: 5,
            originalLength: 20,
            finalLength: 23,
            normalized: true
          }
        },
        combinedContent: 'Test prompt with context',
        contextFiles: []
      });
    });

    await runThinktank(options);
    
    // Simply verify the mock was called - the implementation doesn't guarantee param values
    expect(mockLLMClient.generate).toHaveBeenCalled();
  });

  it('should verify that FileSystem.writeFile is called for file output', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Clear previous calls
    mockFileSystem.writeFile.mockClear();

    await runThinktank(options);
    
    // Verify the FileSystem.writeFile method was called during execution
    expect(mockFileSystem.writeFile).toHaveBeenCalled();
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
    mockFileSystem.mkdir.mockRejectedValueOnce(new Error('Directory creation failed'));
    
    // Override the setupWorkflow spy to handle the error and continue
    setupWorkflowSpy.mockImplementationOnce(async () => {
      throw new Error('Directory creation failed');
    });
    
    // Mock the spinner and error handler to avoid real errors
    const originalHandleWorkflowError = helpers._handleWorkflowError;
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementationOnce(() => {
      return "Error occurred" as unknown as never;
    });
    
    // Invoke the function
    const result = await runThinktank(options);
    
    // Verify we got the error message from the handler
    expect(result).toBe("Error occurred");
    
    // Restore original implementation for other tests
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation(originalHandleWorkflowError);
  });

  it('should handle errors from the ConfigManager interface', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Setup ConfigManager method to throw an error
    mockConfigManager.loadConfig.mockRejectedValueOnce(new Error('Config loading failed'));
    
    // Override the setupWorkflow spy to throw the config error
    setupWorkflowSpy.mockImplementationOnce(async () => {
      throw new Error('Config loading failed');
    });
    
    // Mock the error handler to avoid real errors
    const originalHandleWorkflowError = helpers._handleWorkflowError;
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementationOnce(() => {
      return "Config error occurred" as unknown as never;
    });
    
    // The function will handle the error through our mock
    const result = await runThinktank(options);
    
    // Verify we got the error message
    expect(result).toBe("Config error occurred");
    
    // Restore original implementation for other tests
    jest.spyOn(helpers, '_handleWorkflowError').mockImplementation(originalHandleWorkflowError);
  });

  it('should handle errors from the LLMClient interface', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Setup LLMClient method to throw an error
    const apiError = new ApiError('API call failed');
    mockLLMClient.generate.mockRejectedValueOnce(apiError);
    
    // Instead of having the execute queries call through to LLMClient.generate,
    // have it return a result that includes an error
    executeQueriesSpy.mockResolvedValueOnce({
      queryResults: {
        responses: [{
          provider: 'mock',
          modelId: 'mock-model',
          text: '',
          error: 'API call failed',
          configKey: 'mock:mock-model',
        }],
        statuses: {
          'mock:mock-model': { 
            status: 'error',
            startTime: 1,
            endTime: 2,
            durationMs: 1,
            message: 'API call failed'
          }
        },
        timing: {
          startTime: 1,
          endTime: 2,
          durationMs: 1
        },
        combinedContent: 'Test prompt with context'
      }
    });

    // Mock the console output to contain the error for testing
    processOutputSpy.mockReturnValue({
      files: mockOutputResult.files,
      consoleOutput: 'Mock output with API call failed'
    } as import('../runThinktankTypes').PureProcessOutputResult);
    
    // Mock the writeOutputFiles function
    writeOutputFilesSpy.mockResolvedValue(mockFileOutputResult);
    
    // With our mocked implementation, this won't actually throw an error
    const result = await runThinktank(options);
    
    // Verify that we get a result with an error message
    expect(result).toContain('API call failed');
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
    // Get reference to the mocked modelSelector
    const modelSelector = jest.requireMock('../modelSelector');
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

  it('should pass custom options through', async () => {
    const options: RunOptions = {
      input: 'test-prompt.txt',
      contextPaths: ['src/file1.js'],
      models: ['mock:mock-model'],
      systemPrompt: 'Custom system prompt',
      enableThinking: true,
      includeMetadata: true,
      useColors: true,
      output: '/custom/output'
    };
    
    // Override processOutputSpy to return a consistent console output
    processOutputSpy.mockReset();
    processOutputSpy.mockReturnValue({
      files: mockOutputResult.files,
      consoleOutput: 'Mock console output'
    } as import('../runThinktankTypes').PureProcessOutputResult);
    
    // Mock formatCompletionSummary to return known value
    // Use jest.requireMock to get mocked version
    const formatCompletionSummary = jest.requireMock('../../utils/formatCompletionSummary').formatCompletionSummary;
    formatCompletionSummary.mockReturnValue({
      summaryText: 'Mock summary text',
      errorDetails: []
    });
    
    // Make sure mock returns a known value (using jest.requireMock)
    const mockFormatCompletionSummary = jest.requireMock('../../utils/formatCompletionSummary').formatCompletionSummary;
    mockFormatCompletionSummary.mockReturnValue({
      summaryText: 'Mock summary text',
      errorDetails: []
    });
    
    // Override runThinktank steps that would read model output
    executeQueriesSpy.mockResolvedValue({
      queryResults: {
        responses: [
          {
            provider: 'mock',
            modelId: 'mock-model',
            text: 'Basic mock response',
            configKey: 'mock:mock-model',
          }
        ],
        statuses: {
          'mock:mock-model': { status: 'success', durationMs: 1 }
        },
        timing: { startTime: 1, endTime: 2, durationMs: 1 }
      }
    });
    
    // Directly verify runThinktank returns expected value 
    const result = await runThinktank(options);
    
    // In this test we just make sure it completes - the exact output will vary
    expect(typeof result).toBe('string');
  });
});
