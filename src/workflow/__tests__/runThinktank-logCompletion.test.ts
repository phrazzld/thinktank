/**
 * Integration test for the runThinktank completion summary logging
 * 
 * Tests that the main runThinktank function correctly formats and logs
 * completion summaries after refactoring to remove _logCompletionSummary
 */
import { runThinktank, RunOptions } from '../runThinktank';
import * as formatUtils from '../../utils/formatCompletionSummary';
import * as logger from '../../utils/logger';

// Mock implementations
jest.mock('../../utils/logger', () => {
  const mockLogger = {
    plain: jest.fn(),
    error: jest.fn(),
    info: jest.fn(),
    debug: jest.fn(),
    warn: jest.fn(),
    success: jest.fn(),
    verbose: jest.fn(),
    setLevel: jest.fn(),
    getLevel: jest.fn(),
    configure: jest.fn()
  };
  return {
    logger: mockLogger,
    Logger: jest.fn().mockImplementation(() => mockLogger)
  };
});

// Mock formatCompletionSummary
jest.mock('../../utils/formatCompletionSummary', () => ({
  formatCompletionSummary: jest.fn().mockReturnValue({
    summaryText: 'Mock summary text',
    errorDetails: ['Mock error detail 1', 'Mock error detail 2']
  })
}));

// Mock other functions called by runThinktank
jest.mock('../runThinktankHelpers', () => ({
  _setupWorkflow: jest.fn().mockResolvedValue({
    config: {},
    friendlyRunName: 'mock-run',
    outputDirectoryPath: '/mock/output/dir'
  }),
  _processInput: jest.fn().mockResolvedValue({
    inputResult: { content: 'Mock input', sourceType: 'direct' },
    combinedContent: 'Mock input'
  }),
  _selectModels: jest.fn().mockReturnValue({
    models: [{
      provider: 'mock',
      modelId: 'mock-model',
      enabled: true
    }],
    warnings: []
  }),
  _executeQueries: jest.fn().mockResolvedValue({
    queryResults: {
      responses: [{
        provider: 'mock',
        modelId: 'mock-model',
        text: 'Mock response',
        configKey: 'mock:mock-model'
      }],
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
  }),
  _processOutput: jest.fn().mockResolvedValue({
    files: [{
      filename: 'mock-output.md',
      content: 'Mock content',
      modelKey: 'mock:mock-model'
    }],
    consoleOutput: 'Mock console output'
  }),
  _handleWorkflowError: jest.fn()
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
    text: ''
  };
  
  return {
    __esModule: true,
    default: jest.fn(() => mockSpinner),
    configureSpinnerFactory: jest.fn()
  };
});

// Mock classes
jest.mock('../../core/FileSystem');
jest.mock('../../core/ConcreteConfigManager');
jest.mock('../../core/LLMClient');

describe('runThinktank Completion Summary Integration', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should call formatCompletionSummary with correct data', async () => {
    // Sample options
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Run the function
    await runThinktank(options);

    // Verify formatCompletionSummary was called with expected data
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledTimes(1);
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledWith(
      expect.objectContaining({
        totalModels: 1,
        successCount: 1,
        failureCount: 0,
        outputDirectoryPath: '/mock/output/dir',
        runName: 'mock-run'
      }),
      expect.any(Object)
    );
  });

  it('should log formatted summary text and error details', async () => {
    // Sample options
    const options: RunOptions = {
      input: 'test-prompt.txt'
    };

    // Run the function
    await runThinktank(options);

    // Verify logger.logger.plain was called with summary text and error details
    expect(logger.logger.plain).toHaveBeenCalledWith('Mock summary text');
    expect(logger.logger.plain).toHaveBeenCalledWith('Mock error detail 1');
    expect(logger.logger.plain).toHaveBeenCalledWith('Mock error detail 2');
  });

  it('should pass useColors option from options', async () => {
    // Sample options with useColors: false
    const options: RunOptions = {
      input: 'test-prompt.txt',
      useColors: false
    };

    // Run the function
    await runThinktank(options);

    // Verify formatCompletionSummary was called with useColors: false
    expect(formatUtils.formatCompletionSummary).toHaveBeenCalledWith(
      expect.any(Object),
      expect.objectContaining({
        useColors: false
      })
    );
  });
});
