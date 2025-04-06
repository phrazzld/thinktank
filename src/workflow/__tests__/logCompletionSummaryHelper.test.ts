/**
 * Unit tests for the _logCompletionSummary helper function
 */
import { _logCompletionSummary } from '../runThinktankHelpers';
// Import needed to mock implementation
import '../../utils/consoleUtils';
import { ModelQueryStatus } from '../queryExecutor';
import { FileOutputResult, FileWriteStatus } from '../outputHandler';

// Mock the console.log function
const originalConsoleLog = console.log;
let consoleOutput: string[] = [];

describe('_logCompletionSummary Helper', () => {
  // Set up console.log mock before each test
  beforeEach(() => {
    consoleOutput = [];
    console.log = jest.fn((...args) => {
      // Capture console output for assertions
      consoleOutput.push(args.join(' '));
    });
  });

  // Restore original console.log after each test
  afterEach(() => {
    console.log = originalConsoleLog;
  });

  // Sample query results for testing
  const successfulQueryResults: any = {
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
        status: 'success' as ModelQueryStatus['status'],
        startTime: 100,
        endTime: 200,
        durationMs: 100
      },
      'openai:gpt-4o': {
        status: 'success' as ModelQueryStatus['status'],
        startTime: 100,
        endTime: 300,
        durationMs: 200
      }
    },
    timing: {
      startTime: 100,
      endTime: 300,
      durationMs: 200
    }
  };

  // Sample file output results for testing
  const successfulFileOutput: FileOutputResult = {
    outputDirectory: '/path/to/output/dir',
    files: [
      { modelKey: 'mock:mock-model', filename: 'mock-model.md', status: 'success' as FileWriteStatus, filePath: '/path/to/output/dir/mock-model.md' },
      { modelKey: 'openai:gpt-4o', filename: 'openai-gpt4o.md', status: 'success' as FileWriteStatus, filePath: '/path/to/output/dir/openai-gpt4o.md' }
    ],
    succeededWrites: 2,
    failedWrites: 0,
    timing: { startTime: 300, endTime: 350, durationMs: 50 }
  };

  it('should log successful completion summary', () => {
    // Call the function
    const result = _logCompletionSummary({
      queryResults: successfulQueryResults,
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify empty result (function primarily writes to console)
    expect(result).toEqual({});

    // Verify console output
    expect(consoleOutput.length).toBeGreaterThan(0);
    expect(consoleOutput.some(line => line.includes('Output saved to'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Successfully completed'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('clever-meadow'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Results Summary'))).toBeTruthy();
    
    // Verify it shows success for all models
    expect(consoleOutput.some(line => line.includes('mock:mock-model'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('openai:gpt-4o'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Success'))).toBeTruthy();
  });

  it('should use specific model name in summary when provided', () => {
    // Call the function with specificModel
    _logCompletionSummary({
      queryResults: {
        ...successfulQueryResults,
        responses: [successfulQueryResults.responses[0]],
        statuses: { 'mock:mock-model': successfulQueryResults.statuses['mock:mock-model'] }
      },
      fileOutputResult: {
        ...successfulFileOutput,
        files: [successfulFileOutput.files[0]],
        succeededWrites: 1,
        failedWrites: 0
      },
      options: {
        input: 'test-prompt.txt',
        specificModel: 'mock:mock-model',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output mentions the specific model
    expect(consoleOutput.some(line => line.includes('mock:mock-model'))).toBeTruthy();
  });

  it('should use group name in summary when provided', () => {
    // Call the function with groupName
    _logCompletionSummary({
      queryResults: successfulQueryResults,
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        groupName: 'coding',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output mentions the group name
    expect(consoleOutput.some(line => line.includes('coding group'))).toBeTruthy();
  });

  it('should correctly format completion time', () => {
    // For milliseconds
    _logCompletionSummary({
      queryResults: {
        ...successfulQueryResults,
        timing: { startTime: 100, endTime: 800, durationMs: 700 }
      },
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify milliseconds format
    expect(consoleOutput.some(line => line.includes('700ms'))).toBeTruthy();

    // Reset console output
    consoleOutput = [];

    // For seconds
    _logCompletionSummary({
      queryResults: {
        ...successfulQueryResults,
        timing: { startTime: 100, endTime: 3100, durationMs: 3000 }
      },
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify seconds format
    expect(consoleOutput.some(line => line.includes('3.00s'))).toBeTruthy();
  });

  it('should handle partial success with some failures', () => {
    // Setup mixed success/failure results
    const mixedQueryResults: any = {
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
          error: 'API error',
          configKey: 'openai:gpt-4o',
          metadata: {}
        }
      ],
      statuses: {
        'mock:mock-model': {
          status: 'success' as ModelQueryStatus['status'],
          startTime: 100,
          endTime: 200,
          durationMs: 100
        },
        'openai:gpt-4o': {
          status: 'error' as ModelQueryStatus['status'],
          message: 'API error',
          startTime: 100,
          endTime: 300,
          durationMs: 200
        }
      },
      timing: {
        startTime: 100,
        endTime: 300,
        durationMs: 200
      }
    };

    // Call the function
    _logCompletionSummary({
      queryResults: mixedQueryResults,
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output contains warning
    expect(consoleOutput.some(line => line.includes('models completed successfully'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Failed Models'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Successful Models'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('API error'))).toBeTruthy();
    
    // Should show correct percentages
    expect(consoleOutput.some(line => line.includes('50%'))).toBeTruthy();
  });

  it('should handle failed file writes', () => {
    // Setup file output with failures
    const failedFileOutput: FileOutputResult = {
      outputDirectory: '/path/to/output/dir',
      files: [
        { modelKey: 'mock:mock-model', filename: 'mock-model.md', status: 'success' as FileWriteStatus, filePath: '/path/to/output/dir/mock-model.md' },
        { 
          modelKey: 'openai:gpt-4o', 
          filename: 'openai-gpt4o.md', 
          status: 'error' as FileWriteStatus,
          error: 'Permission denied',
          filePath: '/path/to/output/dir/openai-gpt4o.md'
        }
      ],
      succeededWrites: 1,
      failedWrites: 1,
      timing: { startTime: 300, endTime: 350, durationMs: 50 }
    };

    // Call the function
    _logCompletionSummary({
      queryResults: successfulQueryResults,
      fileOutputResult: failedFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output shows file write warning
    expect(consoleOutput.some(line => line.includes('1 of 2 output files were written'))).toBeTruthy();
  });

  it('should handle all API calls failing', () => {
    // Setup all failures in query results
    const allFailuresQueryResults: any = {
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: '',
          error: 'Timeout',
          configKey: 'mock:mock-model',
          metadata: {}
        },
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          text: '',
          error: 'API key invalid',
          configKey: 'openai:gpt-4o',
          metadata: {}
        }
      ],
      statuses: {
        'mock:mock-model': {
          status: 'error' as ModelQueryStatus['status'],
          message: 'Timeout',
          startTime: 100,
          endTime: 200,
          durationMs: 100
        },
        'openai:gpt-4o': {
          status: 'error' as ModelQueryStatus['status'],
          message: 'API key invalid',
          startTime: 100,
          endTime: 300,
          durationMs: 200
        }
      },
      timing: {
        startTime: 100,
        endTime: 300,
        durationMs: 200
      }
    };

    // Setup file output with all failures
    const allFailedFileOutput: FileOutputResult = {
      outputDirectory: '/path/to/output/dir',
      files: [
        { 
          modelKey: 'mock:mock-model', 
          filename: 'mock-model.md', 
          status: 'error' as FileWriteStatus,
          error: 'Permission denied',
          filePath: '/path/to/output/dir/mock-model.md'
        },
        { 
          modelKey: 'openai:gpt-4o', 
          filename: 'openai-gpt4o.md', 
          status: 'error' as FileWriteStatus,
          error: 'Permission denied',
          filePath: '/path/to/output/dir/openai-gpt4o.md'
        }
      ],
      succeededWrites: 0,
      failedWrites: 2,
      timing: { startTime: 300, endTime: 350, durationMs: 50 }
    };

    // Call the function
    _logCompletionSummary({
      queryResults: allFailuresQueryResults,
      fileOutputResult: allFailedFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output shows appropriate messages
    expect(consoleOutput.some(line => line.includes('No output files were written'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('0% of models completed successfully'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('Failed Models (2)'))).toBeTruthy();
    
    // Should correctly categorize errors (may be unknown category in this case)
    expect(consoleOutput.some(line => line.includes('errors'))).toBeTruthy();
  });

  it('should handle errors with categories', () => {
    // Setup query results with categorized errors
    const categorizedErrorsQueryResults: any = {
      responses: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          text: '',
          error: 'Network error',
          errorCategory: 'NETWORK',
          configKey: 'mock:mock-model',
          metadata: {}
        },
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          text: '',
          error: 'API key invalid',
          errorCategory: 'API',
          configKey: 'openai:gpt-4o',
          metadata: {}
        }
      ],
      statuses: {
        'mock:mock-model': {
          status: 'error' as ModelQueryStatus['status'],
          message: 'Network error',
          detailedError: { category: 'NETWORK' },
          startTime: 100,
          endTime: 200,
          durationMs: 100
        },
        'openai:gpt-4o': {
          status: 'error' as ModelQueryStatus['status'],
          message: 'API key invalid',
          detailedError: { category: 'API' },
          startTime: 100,
          endTime: 300,
          durationMs: 200
        }
      },
      timing: {
        startTime: 100,
        endTime: 300,
        durationMs: 200
      }
    };

    // Call the function
    _logCompletionSummary({
      queryResults: categorizedErrorsQueryResults,
      fileOutputResult: successfulFileOutput,
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output shows error categories
    expect(consoleOutput.some(line => line.includes('NETWORK'))).toBeTruthy();
    expect(consoleOutput.some(line => line.includes('API'))).toBeTruthy();
  });

  it('should format correctly with no models queried', () => {
    // Setup empty query results
    const emptyQueryResults: any = {
      responses: [],
      statuses: {},
      timing: {
        startTime: 100,
        endTime: 100,
        durationMs: 0
      }
    };

    // Call the function
    _logCompletionSummary({
      queryResults: emptyQueryResults,
      fileOutputResult: {
        outputDirectory: '/path/to/output/dir',
        files: [],
        succeededWrites: 0,
        failedWrites: 0,
        timing: { startTime: 100, endTime: 100, durationMs: 0 }
      },
      options: {
        input: 'test-prompt.txt',
        friendlyRunName: 'clever-meadow'
      },
      outputDirectoryPath: '/path/to/output/dir'
    });

    // Verify console output shows appropriate message
    expect(consoleOutput.some(line => line.includes('No models were queried'))).toBeTruthy();
  });
});