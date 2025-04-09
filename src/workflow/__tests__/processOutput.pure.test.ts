/**
 * Unit tests for processOutput pure function
 *
 * This test file verifies that the processOutput function in outputHandler.ts
 * correctly transforms LLM responses into structured data without performing I/O.
 */

import { LLMResponse } from '../../core/types';
import * as outputHandler from '../outputHandler';
import * as helpers from '../../utils/helpers';

// Get direct reference to the function we're testing
const { processOutput } = outputHandler;

// Mock the helpers module
jest.mock('../../utils/helpers', () => ({
  sanitizeFilename: jest.fn(str => str.replace(/[^a-z0-9-]/gi, '_')),
  generateOutputDirectoryPath: jest.fn(() => '/mock/output/directory'),
}));

// Mock the formatForConsole and other functions in outputHandler
jest.mock('../outputHandler', () => {
  const actual = jest.requireActual('../outputHandler');
  return {
    ...actual,
    formatForConsole: jest.fn(() => 'Mock console output'),
    formatResponseAsMarkdown: jest.fn(
      (response, includeMetadata) =>
        `Mock content for ${response.configKey}${includeMetadata ? ' with metadata' : ''}`
    ),
    generateFilename: jest.fn(
      response =>
        `${response.groupInfo?.name ? response.groupInfo.name + '-' : ''}${response.provider}-${response.modelId}.md`
    ),
  };
});

describe('processOutput (Pure)', () => {
  // Reset mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // Sample test responses
  const mockResponses: Array<LLMResponse & { configKey: string }> = [
    {
      provider: 'openai',
      modelId: 'gpt-4o',
      text: 'Response from GPT-4o',
      configKey: 'openai:gpt-4o',
      metadata: { responseTime: 1200 },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-opus',
      text: 'Response from Claude 3 Opus',
      configKey: 'anthropic:claude-3-opus',
      groupInfo: { name: 'coding' },
    },
    {
      provider: 'error',
      modelId: 'model',
      text: '',
      error: 'Error message',
      configKey: 'error:model',
    },
  ];

  it('should transform responses into structured data without I/O', () => {
    // Call the pure function
    const result = processOutput(mockResponses);

    // Verify structure of result
    expect(result).toHaveProperty('files');
    expect(result).toHaveProperty('directoryPath');
    expect(result).toHaveProperty('consoleOutput');

    // Verify files array structure
    expect(result.files).toHaveLength(3);
    expect(result.files[0]).toHaveProperty('filename');
    expect(result.files[0]).toHaveProperty('content');
    expect(result.files[0]).toHaveProperty('modelKey');

    // Verify directory path
    expect(result.directoryPath).toBe('/mock/output/directory');
    expect(helpers.generateOutputDirectoryPath).toHaveBeenCalled();
  });

  it('should pass options to formatters correctly', () => {
    // Call with specific options
    const options = {
      includeMetadata: true,
      useColors: false,
      includeThinking: true,
      useTable: true,
      outputDirectory: '/custom/path',
      friendlyRunName: 'test-run',
    };

    processOutput(mockResponses, options);

    // Verify directory path generation
    expect(helpers.generateOutputDirectoryPath).toHaveBeenCalledWith(
      '/custom/path',
      undefined, // no directoryIdentifier passed
      'test-run'
    );
  });

  it('should handle empty responses array', () => {
    // Force the mock to return a specific value for empty arrays
    const formatForConsoleMock = jest.spyOn(outputHandler, 'formatForConsole');
    formatForConsoleMock.mockReturnValue('No results to display.');

    const result = processOutput([]);

    expect(result.files).toHaveLength(0);
    expect(result.directoryPath).toBe('/mock/output/directory');
    expect(result.consoleOutput).toBe('No results to display.');
  });

  it('should include group in filename when available', () => {
    const result = processOutput([mockResponses[1]]);

    // Since we mocked generateFilename to include the group name if available,
    // the generated filename should include the group name
    expect(result.files[0].filename).toMatch(/coding/);
    expect(result.files[0].modelKey).toBe('anthropic:claude-3-opus');
  });
});
