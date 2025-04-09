/**
 * Integration test for the runThinktank completion summary logging
 *
 * Tests that the main runThinktank function correctly formats and logs
 * completion summaries after refactoring to remove _logCompletionSummary
 */
import { runThinktank, RunOptions } from '../runThinktank';
import * as formatUtils from '../../utils/formatCompletionSummary';
import { ConsoleLogger } from '../../core/interfaces';
import { FileSystem } from '../../core/interfaces';
import { ConfigManagerInterface } from '../../core/interfaces';
import { LLMClient } from '../../core/interfaces';

// Mock the formatCompletionSummary function
jest.mock('../../utils/formatCompletionSummary', () => {
  return {
    formatCompletionSummary: jest.fn().mockReturnValue({
      summaryText: 'Mock summary text',
      errorDetails: ['Mock error detail 1', 'Mock error detail 2'],
    }),
  };
});

// Create mock implementations for dependencies
const mockFileSystem: jest.Mocked<FileSystem> = {
  readFileContent: jest.fn().mockResolvedValue(''),
  writeFile: jest.fn().mockResolvedValue(undefined),
  fileExists: jest.fn().mockResolvedValue(true),
  mkdir: jest.fn().mockResolvedValue(undefined),
  readdir: jest.fn().mockResolvedValue([]),
  stat: jest.fn().mockResolvedValue({} as any),
  access: jest.fn().mockResolvedValue(undefined),
  getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
  getConfigFilePath: jest.fn().mockResolvedValue('/mock/config/file.json'),
};

const mockConfigManager: jest.Mocked<ConfigManagerInterface> = {
  loadConfig: jest.fn().mockResolvedValue({}),
  saveConfig: jest.fn().mockResolvedValue(undefined),
  getActiveConfigPath: jest.fn().mockResolvedValue('/mock/config/path'),
  getDefaultConfigPath: jest.fn().mockReturnValue('/mock/default/config.json'),
  addOrUpdateModel: jest.fn().mockReturnValue({}),
  removeModel: jest.fn().mockReturnValue({}),
  addOrUpdateGroup: jest.fn().mockReturnValue({}),
  removeGroup: jest.fn().mockReturnValue({}),
  addModelToGroup: jest.fn().mockReturnValue({}),
  removeModelFromGroup: jest.fn().mockReturnValue({}),
};

const mockLlmClient: jest.Mocked<LLMClient> = {
  generate: jest.fn().mockResolvedValue({
    model: 'mock:model',
    response: 'Mock response',
    error: null,
  }),
};

const mockConsoleLogger: jest.Mocked<ConsoleLogger> = {
  plain: jest.fn(),
  error: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  success: jest.fn(),
  debug: jest.fn(),
};

// Mock other functions called by runThinktank
jest.mock('../runThinktankHelpers', () => ({
  _setupWorkflow: jest.fn().mockResolvedValue({
    config: {},
    friendlyRunName: 'mock-run',
    outputDirectoryPath: '/mock/output/dir',
  }),
  _processInput: jest.fn().mockResolvedValue({
    inputResult: { content: 'Mock input', sourceType: 'direct' },
    combinedContent: 'Mock input',
  }),
  _selectModels: jest.fn().mockReturnValue({
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true,
      },
    ],
    warnings: [],
  }),
  _executeQueries: jest.fn().mockResolvedValue({
    queryResults: {
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: 'Mock response',
          configKey: 'mock:mock-model',
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
  }),
  _processOutput: jest.fn().mockResolvedValue({
    files: [
      {
        filename: 'mock-output.md',
        content: 'Mock content',
        modelKey: 'mock:mock-model',
      },
    ],
    consoleOutput: 'Mock console output',
  }),
  _handleWorkflowError: jest.fn(),
}));

// Mock spinner
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
    configureSpinnerFactory: jest.fn(),
  };
});

// Mock write files functions
jest.mock('../io', () => ({
  writeFiles: jest.fn().mockResolvedValue({
    successCount: 1,
    errorCount: 0,
    timing: { startTime: 1, endTime: 2, durationMs: 1 },
  }),
  updateSpinnerWithFileOutput: jest.fn(),
}));

describe('runThinktank Completion Summary Integration', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should call formatCompletionSummary with correct data', async () => {
    // Sample options
    const options: RunOptions = {
      input: 'test-prompt.txt',
    };

    // Run the function
    await runThinktank(
      options,
      mockFileSystem,
      mockConfigManager,
      mockLlmClient,
      mockConsoleLogger
    );

    // Verify formatCompletionSummary was called with expected data
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledTimes(1);
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledWith(
      expect.objectContaining({
        totalModels: 1,
        successCount: 1,
        failureCount: 0,
        outputDirectoryPath: '/mock/output/dir',
        runName: 'mock-run',
      }),
      expect.any(Object)
    );
  });

  it('should log formatted summary text and error details', async () => {
    // Sample options
    const options: RunOptions = {
      input: 'test-prompt.txt',
    };

    // Run the function
    await runThinktank(
      options,
      mockFileSystem,
      mockConfigManager,
      mockLlmClient,
      mockConsoleLogger
    );

    // Verify the mock ConsoleLogger was called with summary text and error details
    expect(mockConsoleLogger.plain).toHaveBeenCalledWith('Mock summary text');
    expect(mockConsoleLogger.plain).toHaveBeenCalledWith('Mock error detail 1');
    expect(mockConsoleLogger.plain).toHaveBeenCalledWith('Mock error detail 2');
  });

  it('should pass useColors option from options', async () => {
    // Sample options with useColors: false
    const options: RunOptions = {
      input: 'test-prompt.txt',
      useColors: false,
    };

    // Run the function
    await runThinktank(
      options,
      mockFileSystem,
      mockConfigManager,
      mockLlmClient,
      mockConsoleLogger
    );

    // Verify formatCompletionSummary was called with useColors: false
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledWith(
      expect.any(Object),
      expect.objectContaining({
        useColors: false,
      })
    );
  });
});
