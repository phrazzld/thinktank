/**
 * Tests for the output formatter functions
 */
import {
  formatResponseForMarkdownFile,
  formatResponse,
  formatResultsForConsole,
  formatCompletionSummary
} from '../outputFormatter';
import { LLMResponse } from '../../core/types';
import {
  CompletionSummaryData
} from '../../workflow/types';

// Mock data
const mockResponse: LLMResponse = {
  provider: 'openai',
  modelId: 'gpt-4o',
  text: 'Test response content',
  error: undefined,
  metadata: {
    responseTime: 1234,
    usage: {
      total_tokens: 42
    }
  },
  groupInfo: {
    name: 'test-group'
  }
};

const mockResponseWithError: LLMResponse = {
  provider: 'openai',
  modelId: 'gpt-4o',
  text: '', // Use empty string instead of undefined
  error: 'Test error message',
  metadata: {
    responseTime: 0,
  },
  groupInfo: {
    name: 'test-group'
  }
};

const mockCompletionSummaryData: CompletionSummaryData = {
  totalModels: 3,
  successCount: 2,
  failureCount: 1,
  errors: [
    {
      modelKey: 'anthropic:claude-3-opus',
      message: 'Request timed out',
      category: 'Network'
    }
  ],
  runName: 'test-run',
  outputDirectoryPath: '/path/to/output',
  totalExecutionTimeMs: 5000
};

describe('outputFormatter', () => {
  describe('formatResponseForMarkdownFile', () => {
    it('formats a successful response correctly', () => {
      const result = formatResponseForMarkdownFile(mockResponse, { includeMetadata: true });
      
      expect(result).toContain('# Model: openai:gpt-4o');
      expect(result).toContain('Group: test-group');
      expect(result).toContain('Test response content');
      expect(result).toContain('```json');
      expect(result).not.toContain('Error:');
    });
    
    it('formats an error response correctly', () => {
      const result = formatResponseForMarkdownFile(mockResponseWithError);
      
      expect(result).toContain('# Model: openai:gpt-4o');
      expect(result).toContain('## Error');
      expect(result).toContain('Test error message');
    });
    
    it('respects format options', () => {
      const result = formatResponseForMarkdownFile(mockResponse, { includeMetadata: false });
      
      expect(result).not.toContain('## Metadata');
      expect(result).not.toContain('```json');
    });
  });
  
  describe('formatResponse', () => {
    it('calls formatResponseForMarkdownFile', () => {
      // Verify that formatResponse is just an alias
      const mockOptions = { includeMetadata: true, includeText: false };
      const result = formatResponse(mockResponse, mockOptions);
      const expected = formatResponseForMarkdownFile(mockResponse, mockOptions);
      
      expect(result).toBe(expected);
    });
  });
  
  describe('formatResultsForConsole', () => {
    it('formats console output with default options', () => {
      const mockResponseWithConfigKey = { ...mockResponse, configKey: 'openai:gpt-4o' };
      const result = formatResultsForConsole([mockResponseWithConfigKey]);
      
      expect(result).toContain('Model: openai:gpt-4o');
      expect(result).toContain('Test response content');
    });
    
    it('ensures useColors is set by default', () => {
      // We can't easily test the colors directly, but we can check that
      // passing no options doesn't break the function
      const mockResponseWithConfigKey = { ...mockResponse, configKey: 'openai:gpt-4o' };
      const result = formatResultsForConsole([mockResponseWithConfigKey]);
      
      expect(result).toBeTruthy();
    });
  });
  
  describe('formatCompletionSummary', () => {
    it('formats successful summary correctly', () => {
      const successData: CompletionSummaryData = {
        ...mockCompletionSummaryData,
        successCount: 3,
        failureCount: 0,
        errors: []
      };
      
      const result = formatCompletionSummary(successData, { useColors: false });
      
      expect(result.summaryText).toContain('Successfully completed');
      expect(result.summaryText).toContain('test-run');
      expect(result.summaryText).toContain('in 5.00s');
      expect(result.summaryText).toContain('/path/to/output');
      expect(result.errorDetails).toBeUndefined();
    });
    
    it('formats partial success summary correctly', () => {
      const result = formatCompletionSummary(mockCompletionSummaryData, { useColors: false });
      
      expect(result.summaryText).toContain('Partially completed');
      expect(result.summaryText).toContain('67% success (2/3)');
      expect(result.errorDetails).toBeDefined();
      expect(result.errorDetails?.length).toBeGreaterThan(0);
      expect(result.errorDetails![0]).toContain('Failed Models:');
      expect(result.errorDetails![1]).toContain('anthropic:claude-3-opus');
    });
    
    it('formats full failure summary correctly', () => {
      const failureData: CompletionSummaryData = {
        ...mockCompletionSummaryData,
        successCount: 0,
        failureCount: 3,
        errors: [
          { modelKey: 'model1', message: 'Error 1' },
          { modelKey: 'model2', message: 'Error 2' },
          { modelKey: 'model3', message: 'Error 3' }
        ]
      };
      
      const result = formatCompletionSummary(failureData, { useColors: false });
      
      expect(result.summaryText).toContain('All models failed for');
      expect(result.errorDetails).toBeDefined();
      expect(result.errorDetails?.length).toBeGreaterThan(3); // Header + 3 errors
    });
    
    it('groups errors by category', () => {
      const categorizedErrors: CompletionSummaryData = {
        ...mockCompletionSummaryData,
        successCount: 0,
        failureCount: 3,
        errors: [
          { modelKey: 'model1', message: 'Error 1', category: 'Network' },
          { modelKey: 'model2', message: 'Error 2', category: 'API' },
          { modelKey: 'model3', message: 'Error 3', category: 'Network' }
        ]
      };
      
      const result = formatCompletionSummary(categorizedErrors, { useColors: false });
      
      expect(result.errorDetails).toBeDefined();
      const errorText = result.errorDetails!.join('\n');
      
      expect(errorText).toContain('Network errors:');
      expect(errorText).toContain('API errors:');
    });
  });
});
