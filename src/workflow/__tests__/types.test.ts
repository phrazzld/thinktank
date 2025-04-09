/**
 * Tests for the data interfaces defined in src/workflow/types.ts
 */
import {
  FileData,
  PureProcessOutputResult,
  CompletionSummaryData,
  PureCompletionSummaryResult,
} from '../types';

describe('Data Interfaces', () => {
  test('FileData allows creation of valid file data object', () => {
    const fileData: FileData = {
      filename: 'test-file.md',
      content: 'Test content',
      modelKey: 'openai:gpt-4o',
    };

    expect(fileData.filename).toBe('test-file.md');
    expect(fileData.content).toBe('Test content');
    expect(fileData.modelKey).toBe('openai:gpt-4o');
  });

  test('PureProcessOutputResult allows creation of valid result object', () => {
    const result: PureProcessOutputResult = {
      files: [
        {
          filename: 'model1.md',
          content: 'Content for model 1',
          modelKey: 'provider1:model1',
        },
        {
          filename: 'model2.md',
          content: 'Content for model 2',
          modelKey: 'provider2:model2',
        },
      ],
      consoleOutput: 'Formatted console output',
    };

    expect(result.files).toHaveLength(2);
    expect(result.files[0].filename).toBe('model1.md');
    expect(result.consoleOutput).toBe('Formatted console output');
  });

  test('CompletionSummaryData allows creation of valid data object', () => {
    const data: CompletionSummaryData = {
      totalModels: 5,
      successCount: 3,
      failureCount: 2,
      errors: [
        {
          modelKey: 'provider1:model1',
          message: 'Error message 1',
          category: 'API',
        },
        {
          modelKey: 'provider2:model2',
          message: 'Error message 2',
        },
      ],
      runName: 'test-run',
      outputDirectoryPath: '/path/to/output',
      totalExecutionTimeMs: 2500,
    };

    expect(data.totalModels).toBe(5);
    expect(data.successCount).toBe(3);
    expect(data.failureCount).toBe(2);
    expect(data.errors).toHaveLength(2);
    expect(data.errors[0].category).toBe('API');
    expect(data.errors[1].category).toBeUndefined();
  });

  test('PureCompletionSummaryResult allows creation of valid result object', () => {
    const result: PureCompletionSummaryResult = {
      summaryText: 'Summary of the run',
      errorDetails: ['Error 1 details', 'Error 2 details'],
    };

    expect(result.summaryText).toBe('Summary of the run');
    expect(result.errorDetails).toHaveLength(2);
  });

  test('PureCompletionSummaryResult allows omitting errorDetails when no errors', () => {
    const result: PureCompletionSummaryResult = {
      summaryText: 'Summary of the successful run',
    };

    expect(result.summaryText).toBe('Summary of the successful run');
    expect(result.errorDetails).toBeUndefined();
  });
});
