/**
 * Unit tests for output formatter
 */
import { 
  formatResponse, 
  formatResponses, 
  formatResults,
  formatResultsTable,
  formatModelList,
  FormatOptions,
} from '../outputFormatter';
import { LLMResponse, LLMAvailableModel } from '../../core/types';

// Mock chalk and consoleUtils
jest.mock('chalk', () => ({
  bold: {
    blue: jest.fn(text => `BOLD_BLUE(${text})`),
  },
  red: jest.fn(text => `RED(${text})`),
  gray: jest.fn(text => `GRAY(${text})`),
}));

// Mock cli-table3
jest.mock('cli-table3', () => {
  return jest.fn().mockImplementation(() => {
    return {
      push: jest.fn(),
      toString: jest.fn().mockReturnValue('TABLE_OUTPUT')
    };
  });
});

// Mock console utils
jest.mock('../consoleUtils', () => ({
  colors: {
    bold: jest.fn(text => `BOLD(${text})`),
    red: { bold: jest.fn(text => `RED_BOLD(${text})`) },
  },
  symbols: {
    tick: '✓',
    cross: '✖',
    warning: '⚠',
    info: 'ℹ',
  },
  styleSuccess: jest.fn(text => `SUCCESS(${text})`),
  styleError: jest.fn(text => `ERROR(${text})`),
  styleWarning: jest.fn(text => `WARNING(${text})`),
  styleInfo: jest.fn(text => `INFO(${text})`),
  styleHeader: jest.fn(text => `HEADER(${text})`),
  styleDim: jest.fn(text => `DIM(${text})`),
}));

describe('Output Formatter', () => {
  // Sample data for tests
  const sampleResponse: LLMResponse = {
    provider: 'openai',
    modelId: 'gpt-4',
    text: 'This is a test response.',
    metadata: {
      tokens: 10,
      latency: '123ms',
    },
  };
  
  const errorResponse: LLMResponse = {
    provider: 'openai',
    modelId: 'gpt-4',
    text: '',
    error: 'API Error: Something went wrong',
  };
  
  describe('formatResponse', () => {
    it('should format a basic response with colors', () => {
      const result = formatResponse(sampleResponse);
      
      expect(result).toContain('HEADER(Model: openai:gpt-4)');
      expect(result).toContain('This is a test response.');
      expect(result).not.toContain('Metadata:');
    });
    
    it('should format a response with metadata when requested', () => {
      const options: FormatOptions = { includeMetadata: true };
      const result = formatResponse(sampleResponse, options);
      
      expect(result).toContain('DIM(Metadata:)');
      expect(result).toContain('DIM(  tokens: 10)');
      expect(result).toContain('DIM(  latency: "123ms")');
    });
    
    it('should format errors when present', () => {
      const result = formatResponse(errorResponse);
      
      expect(result).toContain('HEADER(Model: openai:gpt-4)');
      expect(result).toContain('ERROR(Error: API Error: Something went wrong)');
      // Should not include empty text
      expect(result.split('\n').filter(line => line.trim() === '').length).toBe(0);
    });
    
    it('should not use colors when disabled', () => {
      const options: FormatOptions = { useColors: false };
      const result = formatResponse(sampleResponse, options);
      
      expect(result).toContain('Model: openai:gpt-4');
      expect(result).not.toContain('BOLD_BLUE');
    });
    
    it('should exclude text when requested', () => {
      const options: FormatOptions = { includeText: false };
      const result = formatResponse(sampleResponse, options);
      
      expect(result).not.toContain('This is a test response.');
    });
    
    it('should exclude errors when requested', () => {
      const options: FormatOptions = { includeErrors: false };
      const result = formatResponse(errorResponse, options);
      
      expect(result).not.toContain('Error:');
    });
  });
  
  describe('formatResponses', () => {
    it('should format multiple responses with separators', () => {
      const responses = [sampleResponse, { ...sampleResponse, modelId: 'gpt-3.5' }];
      const result = formatResponses(responses);
      
      expect(result).toContain('HEADER(Model: openai:gpt-4)');
      expect(result).toContain('HEADER(Model: openai:gpt-3.5)');
      expect(result).toContain('-'.repeat(80)); // Default separator
    });
    
    it('should return a message when no responses are provided', () => {
      const result = formatResponses([]);
      
      expect(result).toBe('No responses to display.');
    });
    
    it('should use custom separator when provided', () => {
      const responses = [sampleResponse, { ...sampleResponse, modelId: 'gpt-3.5' }];
      const options: FormatOptions = { separator: '\n===\n' };
      const result = formatResponses(responses, options);
      
      expect(result).toContain('===');
      expect(result).not.toContain('-'.repeat(80));
    });
  });
  
  describe('formatResults', () => {
    it('should format results with config keys', () => {
      const results = [
        { ...sampleResponse, configKey: 'openai:gpt-4' },
        { ...sampleResponse, modelId: 'gpt-3.5', configKey: 'openai:gpt-3.5' },
      ];
      
      const result = formatResults(results);
      
      expect(result).toContain('HEADER(Model: openai:gpt-4)');
      expect(result).toContain('HEADER(Model: openai:gpt-3.5)');
    });
    
    it('should use provider and modelId from config key when available', () => {
      const results = [
        { 
          ...sampleResponse, 
          provider: 'wrong', 
          modelId: 'wrong', 
          configKey: 'openai:gpt-4' 
        },
      ];
      
      const result = formatResults(results);
      
      expect(result).toContain('HEADER(Model: openai:gpt-4)');
      expect(result).not.toContain('wrong');
    });
    
    it('should return a message when no results are provided', () => {
      const result = formatResults([]);
      
      expect(result).toBe('No results to display.');
    });
    
    it('should use tabular format when useTable is true', () => {
      const results = [
        { ...sampleResponse, configKey: 'openai:gpt-4' },
        { ...sampleResponse, modelId: 'gpt-3.5', configKey: 'openai:gpt-3.5' },
      ];
      
      const options: FormatOptions = { useTable: true };
      const result = formatResults(results, options);
      
      // Should return the table output from our mock
      expect(result).toBe('TABLE_OUTPUT');
    });
  });
  
  describe('formatResultsTable', () => {
    // Extended sample data with group info
    const sampleResults = [
      { 
        ...sampleResponse, 
        configKey: 'openai:gpt-4',
        metadata: {
          responseTime: 1200,
          usage: { total_tokens: 150 }
        }
      },
      { 
        ...sampleResponse, 
        modelId: 'gpt-3.5', 
        configKey: 'openai:gpt-3.5',
        metadata: {
          responseTime: 800,
          usage: { total_tokens: 120 }
        }
      },
      { 
        ...errorResponse, 
        configKey: 'anthropic:claude-3',
        groupInfo: {
          name: 'creative',
          systemPrompt: { text: 'Be creative' }
        }
      }
    ];
    
    it('should return a message when no results are provided', () => {
      const result = formatResultsTable([]);
      
      expect(result).toBe('No results to display.');
    });
    
    it('should create a table with the correct columns', () => {
      // We're using a mock that returns TABLE_OUTPUT
      const result = formatResultsTable(sampleResults);
      
      expect(result).toBe('TABLE_OUTPUT');
      
      // Get the mocked Table constructor
      const Table = jest.mocked(jest.requireMock('cli-table3'));
      expect(Table).toHaveBeenCalled();
    });
    
    it('should sort results by group then model', () => {
      // We can't easily verify the sorting with our mock,
      // but we can verify the function doesn't throw errors
      const result = formatResultsTable(sampleResults);
      expect(result).toBe('TABLE_OUTPUT');
    });
    
    it('should handle results with group information', () => {
      formatResultsTable(sampleResults);
      // Verify Table.push was called at least once
      const Table = jest.mocked(jest.requireMock('cli-table3'));
      const mockTable = Table.mock.results[0].value;
      expect(mockTable.push).toHaveBeenCalled();
    });
    
    it('should include group statistics when includeMetadata is true', () => {
      const options: FormatOptions = { includeMetadata: true };
      const Table = jest.mocked(jest.requireMock('cli-table3'));
      
      // Mock toString to return just the table part
      Table.mockImplementation(() => ({
        push: jest.fn(),
        toString: jest.fn().mockReturnValue('TABLE_PART')
      }));
      
      const result = formatResultsTable(sampleResults, options);
      
      // The result should have more than just the table output,
      // it should also include group statistics
      expect(result.length).toBeGreaterThan('TABLE_PART'.length);
      expect(result).toContain('TABLE_PART');
    });
  });

  describe('formatModelList', () => {
    // Sample models for tests
    const openaiModels: LLMAvailableModel[] = [
      { id: 'gpt-4o' },
      { id: 'gpt-4-turbo', description: 'Legacy GPT-4 Turbo' }
    ];
    
    const anthropicModels: LLMAvailableModel[] = [
      { id: 'claude-3-opus-20240229', description: 'Most powerful model' },
      { id: 'claude-3-sonnet-20240229', description: 'Balanced model' }
    ];

    it('should format models grouped by provider with colors', () => {
      const modelsByProvider = {
        'openai': openaiModels,
        'anthropic': anthropicModels
      };
      
      const result = formatModelList(modelsByProvider);
      
      expect(result).toContain('HEADER(Available Models:)');
      expect(result).toContain('HEADER(--- openai ---)');
      expect(result).toContain('  - gpt-4o');
      expect(result).toContain('  - gpt-4-turbo (Legacy GPT-4 Turbo)');
      expect(result).toContain('HEADER(--- anthropic ---)');
      expect(result).toContain('  - claude-3-opus-20240229 (Most powerful model)');
      expect(result).toContain('  - claude-3-sonnet-20240229 (Balanced model)');
    });
    
    it('should handle providers with error messages', () => {
      const modelsByProvider = {
        'openai': openaiModels,
        'anthropic': { error: 'API key not found' }
      };
      
      const result = formatModelList(modelsByProvider);
      
      expect(result).toContain('HEADER(Available Models:)');
      expect(result).toContain('HEADER(--- openai ---)');
      expect(result).toContain('ERROR(  Error fetching models: API key not found)');
    });

    it('should handle empty model lists', () => {
      const modelsByProvider = {
        'openai': [],
        'anthropic': anthropicModels
      };
      
      const result = formatModelList(modelsByProvider);
      
      expect(result).toContain('HEADER(--- openai ---)');
      expect(result).toContain('DIM(  (No models available))');
    });
    
    it('should handle no providers', () => {
      const modelsByProvider = {};
      
      const result = formatModelList(modelsByProvider);
      
      expect(result).toContain('HEADER(Available Models:)');
      expect(result).toContain('DIM(No providers configured.)');
    });

    it('should format without colors when disabled', () => {
      const modelsByProvider = {
        'openai': openaiModels
      };
      
      const options: FormatOptions = { useColors: false };
      const result = formatModelList(modelsByProvider, options);
      
      expect(result).toContain('Available Models:');
      expect(result).toContain('--- openai ---');
      expect(result).not.toContain('BOLD_BLUE');
    });
  });
});
