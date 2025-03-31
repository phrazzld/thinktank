/**
 * Unit tests for output formatter
 */
import { 
  formatResponse, 
  formatResponses, 
  formatResults,
  FormatOptions,
} from '../outputFormatter';
import { LLMResponse } from '../../atoms/types';

// Mock chalk to prevent color output in tests
jest.mock('chalk', () => ({
  bold: {
    blue: jest.fn(text => `BOLD_BLUE(${text})`),
  },
  red: jest.fn(text => `RED(${text})`),
  gray: jest.fn(text => `GRAY(${text})`),
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
      
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-4)');
      expect(result).toContain('This is a test response.');
      expect(result).not.toContain('Metadata:');
    });
    
    it('should format a response with metadata when requested', () => {
      const options: FormatOptions = { includeMetadata: true };
      const result = formatResponse(sampleResponse, options);
      
      expect(result).toContain('GRAY(Metadata:)');
      expect(result).toContain('GRAY(  tokens: 10)');
      expect(result).toContain('GRAY(  latency: "123ms")');
    });
    
    it('should format errors when present', () => {
      const result = formatResponse(errorResponse);
      
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-4)');
      expect(result).toContain('RED(Error: API Error: Something went wrong)');
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
      
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-4)');
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-3.5)');
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
      
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-4)');
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-3.5)');
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
      
      expect(result).toContain('BOLD_BLUE(Model: openai:gpt-4)');
      expect(result).not.toContain('wrong');
    });
    
    it('should return a message when no results are provided', () => {
      const result = formatResults([]);
      
      expect(result).toBe('No results to display.');
    });
  });
});