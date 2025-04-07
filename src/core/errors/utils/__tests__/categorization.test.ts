/**
 * Tests for error categorization utilities
 */
import { 
  categorizeError, 
  isRateLimitError,
  isTokenLimitError,
  isContentPolicyError,
  isAuthError,
  isNetworkError
} from '../categorization';
import { errorCategories } from '../../base';

describe('Error Categorization Utilities', () => {
  describe('categorizeError', () => {
    it('should categorize network errors correctly', () => {
      const testCases = [
        'Network connection failed',
        'Connection timed out',
        'Socket hang up',
        'DNS resolution failed',
        'ECONNREFUSED: Connection refused'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.NETWORK);
      });
    });
    
    it('should categorize file system errors correctly', () => {
      const testCases = [
        'File not found: test.txt',
        'Directory does not exist',
        'ENOENT: no such file or directory',
        'Invalid path specified'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.FILESYSTEM);
      });
    });
    
    it('should categorize permission errors correctly', () => {
      const testCases = [
        'Permission denied for file',
        'Access denied to resource',
        'Forbidden operation',
        'Unauthorized: Invalid credentials'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.PERMISSION);
      });
    });
    
    it('should categorize configuration errors correctly', () => {
      const testCases = [
        'Invalid configuration value',
        'Missing required config setting',
        'Option not supported',
        'Settings file is corrupted'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.CONFIG);
      });
    });
    
    it('should categorize validation errors correctly', () => {
      const testCases = [
        'Validation failed for input',
        'Invalid parameter format',
        'Schema validation error',
        'Required field missing'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.VALIDATION);
      });
    });
    
    it('should categorize API errors correctly', () => {
      const testCases = [
        'API request failed',
        'Endpoint returned an error',
        'Service is unavailable',
        'Provider error: quota exceeded'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.API);
      });
    });
    
    it('should categorize input errors correctly', () => {
      const testCases = [
        'Invalid input format',
        'Prompt exceeds maximum length',
        'Query contains unsupported characters',
        'Parameter outside of allowed range'
      ];
      
      testCases.forEach(message => {
        expect(categorizeError(new Error(message))).toBe(errorCategories.INPUT);
      });
    });
    
    it('should default to unknown category for unrecognized errors', () => {
      const message = 'Some completely unexpected and unique error';
      expect(categorizeError(new Error(message))).toBe(errorCategories.UNKNOWN);
    });
  });
  
  describe('Provider error detection functions', () => {
    describe('isRateLimitError', () => {
      it('should detect rate limit errors correctly', () => {
        const validMessages = [
          'Rate limit exceeded',
          'Too many requests, status code: 429',
          'Quota exceeded for this model',
          'You have exceeded your current quota'
        ];
        
        validMessages.forEach(message => {
          expect(isRateLimitError(message)).toBe(true);
        });
        
        const invalidMessages = [
          'Authentication failed',
          'Token limit exceeded',
          'Network error',
          'Model not found'
        ];
        
        invalidMessages.forEach(message => {
          expect(isRateLimitError(message)).toBe(false);
        });
      });
    });
    
    describe('isTokenLimitError', () => {
      it('should detect token limit errors correctly', () => {
        const validMessages = [
          'Token limit exceeded',
          'Maximum context length exceeded',
          'Input exceeds maximum token count',
          'This exceeds the model\'s context window'
        ];
        
        validMessages.forEach(message => {
          expect(isTokenLimitError(message)).toBe(true);
        });
        
        const invalidMessages = [
          'Authentication failed',
          'Rate limit exceeded',
          'Network error',
          'Model not found'
        ];
        
        invalidMessages.forEach(message => {
          expect(isTokenLimitError(message)).toBe(false);
        });
      });
    });
    
    describe('isContentPolicyError', () => {
      it('should detect content policy errors correctly', () => {
        const validMessages = [
          'Content policy violation',
          'Content filter triggered',
          'Your request violates our usage policies',
          'The safety system has identified potentially harmful content'
        ];
        
        validMessages.forEach(message => {
          expect(isContentPolicyError(message)).toBe(true);
        });
        
        const invalidMessages = [
          'Authentication failed',
          'Rate limit exceeded',
          'Network error',
          'Model not found'
        ];
        
        invalidMessages.forEach(message => {
          expect(isContentPolicyError(message)).toBe(false);
        });
      });
    });
    
    describe('isAuthError', () => {
      it('should detect authentication errors correctly', () => {
        const validMessages = [
          'Authentication failed',
          'Invalid API key provided',
          'Unauthorized: 401',
          'Authorization header missing or malformed'
        ];
        
        validMessages.forEach(message => {
          expect(isAuthError(message)).toBe(true);
        });
        
        const invalidMessages = [
          'Rate limit exceeded',
          'Token limit exceeded',
          'Network error',
          'Model not found'
        ];
        
        invalidMessages.forEach(message => {
          expect(isAuthError(message)).toBe(false);
        });
      });
    });
    
    describe('isNetworkError', () => {
      it('should detect network errors correctly', () => {
        const validMessages = [
          'Network connection failed',
          'Connection timeout',
          'ECONNRESET: socket hang up',
          'DNS resolution failed'
        ];
        
        validMessages.forEach(message => {
          expect(isNetworkError(message)).toBe(true);
        });
        
        const invalidMessages = [
          'Authentication failed',
          'Rate limit exceeded',
          'Token limit exceeded',
          'Model not found'
        ];
        
        invalidMessages.forEach(message => {
          expect(isNetworkError(message)).toBe(false);
        });
      });
    });
  });
});
