/**
 * Unit tests for helper functions
 */
import { 
  getModelConfigKey, 
  getDefaultApiKeyEnvVar, 
  normalizeText, 
  getApiKey,
  generateRunDirectoryName,
  resolveOutputDirectory,
  generateOutputDirectoryPath,
  sanitizeFilename
} from '../helpers';
import { ModelConfig } from '../types';
import path from 'path';

describe('Helper Functions', () => {
  describe('getModelConfigKey', () => {
    it('should generate the correct key from provider and modelId', () => {
      const config: ModelConfig = {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
      };
      
      expect(getModelConfigKey(config)).toBe('openai:gpt-4o');
    });
    
    it('should handle special characters in provider and modelId', () => {
      const config: ModelConfig = {
        provider: 'custom-provider',
        modelId: 'special_model.v1',
        enabled: true,
      };
      
      expect(getModelConfigKey(config)).toBe('custom-provider:special_model.v1');
    });
  });
  
  describe('getDefaultApiKeyEnvVar', () => {
    it('should generate uppercase environment variable names with _API_KEY suffix', () => {
      expect(getDefaultApiKeyEnvVar('openai')).toBe('OPENAI_API_KEY');
      expect(getDefaultApiKeyEnvVar('anthropic')).toBe('ANTHROPIC_API_KEY');
    });
    
    it('should handle special characters and mixed case in provider names', () => {
      expect(getDefaultApiKeyEnvVar('CustomProvider')).toBe('CUSTOMPROVIDER_API_KEY');
      expect(getDefaultApiKeyEnvVar('custom-provider')).toBe('CUSTOM-PROVIDER_API_KEY');
    });
  });
  
  describe('getApiKey', () => {
    const originalEnv = process.env;
    
    beforeEach(() => {
      // Reset process.env for each test - use a completely clean object to avoid real env vars
      process.env = {};
    });
    
    afterAll(() => {
      // Restore original environment after all tests
      process.env = originalEnv;
    });
    
    it('should retrieve API key from custom environment variable if specified', () => {
      process.env.CUSTOM_API_KEY = 'custom-key';
      
      const config: ModelConfig = {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
        apiKeyEnvVar: 'CUSTOM_API_KEY',
      };
      
      expect(getApiKey(config)).toBe('custom-key');
    });
    
    it('should fall back to default environment variable if custom is not specified', () => {
      process.env.OPENAI_API_KEY = 'default-key';
      
      const config: ModelConfig = {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
      };
      
      expect(getApiKey(config)).toBe('default-key');
    });
    
    it('should return null if no API key is found', () => {
      const config: ModelConfig = {
        provider: 'unknown',
        modelId: 'model',
        enabled: true,
      };
      
      expect(getApiKey(config)).toBeNull();
    });
    
    it('should handle case-insensitive provider matching', () => {
      process.env.GOOGLE_API_KEY = 'google-key';
      
      const config: ModelConfig = {
        provider: 'Google', // Uppercase
        modelId: 'gemini-pro',
        enabled: true,
      };
      
      expect(getApiKey(config)).toBe('google-key');
    });
    
    it('should try multiple possible environment variable names', () => {
      process.env.GEMINI_API_KEY = 'gemini-key';
      
      const config: ModelConfig = {
        provider: 'google',
        modelId: 'gemini-pro',
        enabled: true,
      };
      
      expect(getApiKey(config)).toBe('gemini-key');
    });
  });
  
  describe('normalizeText', () => {
    it('should trim leading and trailing whitespace', () => {
      expect(normalizeText('  hello world  ')).toBe('hello world');
    });
    
    it('should normalize internal whitespace to single spaces', () => {
      expect(normalizeText('hello  world\t\ntest')).toBe('hello world test');
    });
    
    it('should handle empty or whitespace-only strings', () => {
      expect(normalizeText('')).toBe('');
      expect(normalizeText('   ')).toBe('');
    });
  });
  
  describe('generateRunDirectoryName', () => {
    it('should generate a directory name with the correct format', () => {
      const dirName = generateRunDirectoryName();
      expect(dirName).toMatch(/^thinktank_run_\d{8}_\d{6}_\d{3}$/);
    });
    
    it('should include the correct prefix', () => {
      const dirName = generateRunDirectoryName();
      expect(dirName.startsWith('thinktank_run_')).toBe(true);
    });
    
    it('should generate a name with the correct timestamp format', () => {
      // Mock Date to return a specific timestamp
      const mockDate = new Date('2023-04-15T12:34:56.789Z');
      const originalDate = global.Date;
      
      // Mocking global Date constructor
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      global.Date = jest.fn(() => mockDate) as any;
      global.Date.UTC = originalDate.UTC;
      global.Date.parse = originalDate.parse;
      global.Date.now = originalDate.now;
      
      // Make sure toISOString returns what we expect
      mockDate.toISOString = jest.fn().mockReturnValue('2023-04-15T12:34:56.789Z');
      
      try {
        const dirName = generateRunDirectoryName();
        expect(dirName).toBe('thinktank_run_20230415_123456_789');
      } finally {
        // Restore original Date
        global.Date = originalDate;
      }
    });
  });
  
  describe('resolveOutputDirectory', () => {
    it('should use the provided output directory when specified', () => {
      const outputDir = resolveOutputDirectory('/custom/path');
      expect(outputDir).toBe('/custom/path');
    });
    
    it('should resolve relative paths to absolute paths', () => {
      const outputDir = resolveOutputDirectory('./relative/path');
      expect(outputDir).toBe(path.resolve('./relative/path'));
    });
  });
  
  describe('generateOutputDirectoryPath', () => {
    it('should append a timestamped run directory to the output path', () => {
      const outputPath = generateOutputDirectoryPath('/custom/path');
      // Match the expected pattern with new format: run-YYYYMMDD-HHmmss
      expect(outputPath).toMatch(/^\/custom\/path\/run-\d{8}-\d{6}$/);
    });
    
    it('should include specified identifier in the directory name when provided', () => {
      const outputPath = generateOutputDirectoryPath('/custom/path', 'test-group');
      // Check that identifier is included in path
      expect(outputPath).toContain('test-group-');
      expect(outputPath).toMatch(/^\/custom\/path\/test-group-\d{8}-\d{6}$/);
    });
    
    it('should replace colons with hyphens for provider:model format', () => {
      const outputPath = generateOutputDirectoryPath('/custom/path', 'openai:gpt-4o');
      // Check that colon is replaced with hyphen
      expect(outputPath).toContain('openai-gpt-4o-');
      expect(outputPath).not.toContain(':');
    });
    
    it('should use thinktank-output as default directory when no output is specified', () => {
      // Mock path.resolve to avoid system-specific paths
      const originalResolve = path.resolve;
      path.resolve = jest.fn((...args) => args.join('/'));
      
      try {
        const outputPath = generateOutputDirectoryPath();
        expect(outputPath).toContain('thinktank-output/run-');
      } finally {
        path.resolve = originalResolve;
      }
    });
  });
  
  describe('sanitizeFilename', () => {
    it('should replace invalid characters with underscores', () => {
      expect(sanitizeFilename('file/with:invalid*chars?')).toBe('file_with_invalid_chars_');
    });
    
    it('should handle empty or undefined input', () => {
      expect(sanitizeFilename('')).toBe('unnamed');
      expect(sanitizeFilename(undefined as unknown as string)).toBe('unnamed');
    });
    
    it('should replace whitespace with underscores', () => {
      expect(sanitizeFilename('file with spaces')).toBe('file_with_spaces');
    });
    
    it('should remove leading and trailing dots and hyphens', () => {
      expect(sanitizeFilename('..file-name-.')).toBe('file-name');
    });
    
    it('should collapse multiple underscores to a single one', () => {
      expect(sanitizeFilename('file__with___underscores')).toBe('file_with_underscores');
    });
    
    it('should truncate long filenames', () => {
      const longString = 'a'.repeat(300);
      expect(sanitizeFilename(longString).length).toBe(255);
    });
    
    it('should handle model IDs with special characters', () => {
      expect(sanitizeFilename('openai:gpt-4-turbo-2024-04-09')).toBe('openai_gpt-4-turbo-2024-04-09');
    });
    
    it('should handle non-ASCII characters', () => {
      expect(sanitizeFilename('résumé.pdf')).toBe('résumé.pdf');
      expect(sanitizeFilename('日本語')).toBe('日本語');
      expect(sanitizeFilename('Россия')).toBe('Россия');
    });
    
    it('should handle special characters that are valid in filenames', () => {
      expect(sanitizeFilename('file-name_with.special+characters!')).toBe('file-name_with.special+characters!');
    });
    
    it('should remove control characters', () => {
      // eslint-disable-next-line no-control-regex
      const stringWithControlChars = 'test\u0000\u0001\u0002file';
      expect(sanitizeFilename(stringWithControlChars)).toBe('testfile');
    });
  });
});