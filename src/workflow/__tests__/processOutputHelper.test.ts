/**
 * Unit tests for the _processOutput helper function
 */
import { _processOutput } from '../runThinktankHelpers';
import * as outputHandler from '../outputHandler';
import { ThinktankError } from '../../core/errors';
import { QueryExecutionResult } from '../queryExecutor';
import { RunOptions } from '../runThinktank';
import { LLMResponse } from '../../core/types';

// Mock dependencies
jest.mock('../outputHandler', () => ({
  generateFilename: jest.fn((response, _opt) => 
    `${response.provider}-${response.modelId}.md`),
  formatResponseAsMarkdown: jest.fn((response, includeMetadata) => 
    `Mock content for ${response.configKey}${includeMetadata ? ' with metadata' : ''}`),
  formatForConsole: jest.fn((responses, options) => 
    `Mock console output with ${responses.length} responses${options.includeMetadata ? ' including metadata' : ''}`)
}));

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_processOutput Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  // Sample query results for testing
  const sampleQueryResults: QueryExecutionResult = {
    responses: [
      {
        provider: 'mock',
        modelId: 'model-a',
        text: 'Sample response A',
        error: undefined,
        metadata: { responseTime: 100 },
        groupInfo: { name: 'test-group' },
        configKey: 'mock:model-a'
      } as LLMResponse & { configKey: string },
      {
        provider: 'mock',
        modelId: 'model-b',
        text: 'Sample response B',
        error: undefined,
        metadata: { responseTime: 150 },
        groupInfo: { name: 'default' },
        configKey: 'mock:model-b'
      } as LLMResponse & { configKey: string }
    ],
    statuses: {
      'mock:model-a': {
        status: 'success',
        startTime: 1000,
        endTime: 1100,
        durationMs: 100
      },
      'mock:model-b': {
        status: 'success',
        startTime: 1000,
        endTime: 1150,
        durationMs: 150
      }
    },
    timing: {
      startTime: 1000,
      endTime: 1150,
      durationMs: 150
    },
    combinedContent: 'Test prompt'
  };

  // Sample options for testing
  const sampleOptions: RunOptions = {
    input: 'test-prompt.txt',
    includeMetadata: true
  };

  it('should generate file data for each response', async () => {
    // Act
    const result = await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      options: sampleOptions
    });

    // Assert
    expect(result).toBeDefined();
    expect(result.files).toHaveLength(2);
    
    // Verify first file data
    expect(result.files[0]).toEqual({
      filename: 'mock-model-a.md',
      content: 'Mock content for mock:model-a with metadata',
      modelKey: 'mock:model-a'
    });
    
    // Verify second file data
    expect(result.files[1]).toEqual({
      filename: 'mock-model-b.md',
      content: 'Mock content for mock:model-b with metadata',
      modelKey: 'mock:model-b'
    });
    
    // Verify console output
    expect(result.consoleOutput).toEqual('Mock console output with 2 responses including metadata');
    
    // Verify that the formatters were called correctly
    expect(outputHandler.generateFilename).toHaveBeenCalledTimes(2);
    expect(outputHandler.formatResponseAsMarkdown).toHaveBeenCalledTimes(2);
    expect(outputHandler.formatForConsole).toHaveBeenCalledTimes(1);
  });

  it('should handle options correctly', async () => {
    // Sample options with different settings
    const options: RunOptions = {
      input: 'test-prompt.txt',
      includeMetadata: false,
      useColors: false,
      specificModel: 'mock:model-a'
    };

    // Act
    const result = await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      options
    });

    // Assert
    expect(result).toBeDefined();
    
    // Verify that generateFilename was called with specific model option
    expect(outputHandler.generateFilename).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ includeGroup: false })
    );
    
    // Verify that formatResponseAsMarkdown was called with includeMetadata: false
    expect(outputHandler.formatResponseAsMarkdown).toHaveBeenCalledWith(
      expect.anything(),
      false
    );
    
    // Verify that formatForConsole was called with the correct options
    expect(outputHandler.formatForConsole).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        includeMetadata: false,
        useColors: false
      })
    );
  });

  it('should update the spinner with processing information', async () => {
    // Act
    await _processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      options: sampleOptions
    });

    // Assert
    // Spinner should have been updated with formatting status
    expect(mockSpinner.text).toContain('Formatted 2 results');
  });

  it('should handle errors and wrap them appropriately', async () => {
    // Configure mock to throw an error
    (outputHandler.formatResponseAsMarkdown as jest.Mock).mockImplementationOnce(() => {
      throw new Error('Formatting error');
    });

    // Act & Assert
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      options: sampleOptions
    })).rejects.toThrow(ThinktankError);
    
    // Reset the mock and make it throw again for the second test
    (outputHandler.formatResponseAsMarkdown as jest.Mock).mockImplementationOnce(() => {
      throw new Error('Formatting error');
    });
    
    // Verify the error message
    await expect(_processOutput({
      spinner: mockSpinner,
      queryResults: sampleQueryResults,
      options: sampleOptions
    })).rejects.toThrow('Error formatting output: Formatting error');
  });
});
