/**
 * Unit tests for the formatCompletionSummary utility
 */
import { formatCompletionSummary } from '../formatCompletionSummary';
import { CompletionSummaryData } from '../../workflow/types';

// Mock these to prevent styles in tests
jest.mock('../consoleUtils', () => {
  const red = (text: string) => `RED(${text})`;
  red.bold = (text: string) => `RED_BOLD(${text})`;
  
  return {
    colors: {
      // Simple implementation of functions
      blue: (text: string) => `BLUE(${text})`,
      green: (text: string) => `GREEN(${text})`,
      yellow: (text: string) => `YELLOW(${text})`,
      red
    },
    styleSuccess: (text: string) => `SUCCESS(${text})`,
    styleError: (text: string) => `ERROR(${text})`,
    styleDim: (text: string) => `DIM(${text})`
  };
});

describe('formatCompletionSummary', () => {
  // Setup test data
  const successfulData: CompletionSummaryData = {
    totalModels: 2,
    successCount: 2,
    failureCount: 0,
    errors: [],
    runName: 'test-run',
    outputDirectoryPath: '/path/to/output',
    totalExecutionTimeMs: 1500
  };

  const partialSuccessData: CompletionSummaryData = {
    totalModels: 2,
    successCount: 1,
    failureCount: 1,
    errors: [
      {
        modelKey: 'openai:gpt-4o',
        message: 'API Error',
        category: 'API'
      }
    ],
    runName: 'test-run',
    outputDirectoryPath: '/path/to/output',
    totalExecutionTimeMs: 1500
  };

  const allFailuresData: CompletionSummaryData = {
    totalModels: 2,
    successCount: 0,
    failureCount: 2,
    errors: [
      {
        modelKey: 'openai:gpt-4o',
        message: 'API Error',
        category: 'API'
      },
      {
        modelKey: 'anthropic:claude-3',
        message: 'Network Error',
        category: 'NETWORK'
      }
    ],
    runName: 'test-run',
    outputDirectoryPath: '/path/to/output',
    totalExecutionTimeMs: 1500
  };

  const noModelsData: CompletionSummaryData = {
    totalModels: 0,
    successCount: 0,
    failureCount: 0,
    errors: [],
    runName: 'test-run',
    outputDirectoryPath: '/path/to/output',
    totalExecutionTimeMs: 0
  };

  it('should format successful completion summary with colors', () => {
    const result = formatCompletionSummary(successfulData, { useColors: true });
    
    expect(result.summaryText).toContain('SUCCESS');
    expect(result.summaryText).toContain('test-run');
    expect(result.summaryText).toContain('1.50s');
    expect(result.summaryText).toContain('/path/to/output');
    expect(result.errorDetails).toBeUndefined();
  });

  it('should format successful completion summary without colors', () => {
    const result = formatCompletionSummary(successfulData, { useColors: false });
    
    expect(result.summaryText).not.toContain('SUCCESS');
    expect(result.summaryText).toContain('Successfully completed');
    expect(result.summaryText).toContain('test-run');
    expect(result.summaryText).toContain('1.50s');
    expect(result.summaryText).toContain('/path/to/output');
    expect(result.errorDetails).toBeUndefined();
  });

  it('should format partial success summary', () => {
    const result = formatCompletionSummary(partialSuccessData);
    
    expect(result.summaryText).toContain('test-run');
    expect(result.summaryText).toContain('50%');
    expect(result.summaryText).toContain('/path/to/output');
    
    expect(result.errorDetails).toBeDefined();
    expect(result.errorDetails!.length).toBeGreaterThan(0);
    expect(result.errorDetails!.some(line => line.includes('API'))).toBeTruthy();
    expect(result.errorDetails!.some(line => line.includes('openai:gpt-4o'))).toBeTruthy();
  });

  it('should format all failures summary', () => {
    const result = formatCompletionSummary(allFailuresData);
    
    expect(result.summaryText).toContain('test-run');
    expect(result.summaryText).toContain('failed');
    expect(result.summaryText).toContain('/path/to/output');
    
    expect(result.errorDetails).toBeDefined();
    expect(result.errorDetails!.length).toBeGreaterThan(0);
    expect(result.errorDetails!.some(line => line.includes('API'))).toBeTruthy();
    expect(result.errorDetails!.some(line => line.includes('NETWORK'))).toBeTruthy();
    expect(result.errorDetails!.some(line => line.includes('openai:gpt-4o'))).toBeTruthy();
    expect(result.errorDetails!.some(line => line.includes('anthropic:claude-3'))).toBeTruthy();
  });

  it('should format no models queried summary', () => {
    const result = formatCompletionSummary(noModelsData);
    
    expect(result.summaryText).toContain('test-run');
    expect(result.summaryText).toContain('No models were queried');
    expect(result.summaryText).toContain('/path/to/output');
    expect(result.errorDetails).toBeUndefined();
  });

  it('should handle missing totalExecutionTimeMs', () => {
    const data = { ...successfulData, totalExecutionTimeMs: undefined };
    const result = formatCompletionSummary(data);
    
    expect(result.summaryText).toContain('test-run');
    // The "s" test isn't reliable since "s" appears in "Successfully", so we'll test differently
    expect(result.summaryText).not.toMatch(/\d+\.\d+s/); // No numbers followed by "s"
    expect(result.summaryText).not.toMatch(/\d+ms/); // No numbers followed by "ms"
  });

  it('should format milliseconds correctly for small durations', () => {
    const data = { ...successfulData, totalExecutionTimeMs: 750 };
    const result = formatCompletionSummary(data);
    
    expect(result.summaryText).toContain('0.75s');
  });
});