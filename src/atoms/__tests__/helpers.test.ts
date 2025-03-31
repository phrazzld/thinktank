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
      // Reset process.env for each test
      process.env = { ...originalEnv };
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
    
    it('should return undefined if no API key is found', () => {
      const config: ModelConfig = {
        provider: 'unknown',
        modelId: 'model',
        enabled: true,
      };
      
      expect(getApiKey(config)).toBeUndefined();
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
    
    it('should use the default directory when no output is specified', () => {
      const outputDir = resolveOutputDirectory();
      expect(outputDir).toBe(path.resolve(process.cwd(), 'thinktank_outputs'));
    });
    
    it('should resolve relative paths to absolute paths', () => {
      const outputDir = resolveOutputDirectory('./relative/path');
      expect(outputDir).toBe(path.resolve('./relative/path'));
    });
    
    it('should handle empty string as output option', () => {
      const outputDir = resolveOutputDirectory('');
      expect(outputDir).toBe(path.resolve(process.cwd(), 'thinktank_outputs'));
    });
    
    it('should allow customizing the default directory name', () => {
      const outputDir = resolveOutputDirectory(undefined, 'custom_default');
      expect(outputDir).toBe(path.resolve(process.cwd(), 'custom_default'));
    });
  });
  
  describe('generateOutputDirectoryPath', () => {
    // Instead of trying to mock the internal function, we'll directly test with the actual result pattern
    
    it('should append a timestamped subdirectory to the output path', () => {
      const outputPath = generateOutputDirectoryPath('/custom/path');
      // Match the expected pattern rather than an exact string
      expect(outputPath).toMatch(/^\/custom\/path\/thinktank_run_\d{8}_\d{6}_\d{3}$/);
    });
    
    it('should use the default directory with a timestamped subdirectory when no output is specified', () => {
      const outputPath = generateOutputDirectoryPath();
      // Match the expected pattern with the default directory
      const escapedCwd = process.cwd().replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const pattern = new RegExp(`^${escapedCwd}\\/thinktank_outputs\\/thinktank_run_\\d{8}_\\d{6}_\\d{3}$`);
      expect(outputPath).toMatch(pattern);
    });
    
    it('should include a timestamped directory in the path', () => {
      // Capture the timestamp part from the run directory
      const path1 = generateOutputDirectoryPath('/test/option');
      const timestampMatch = path1.match(/thinktank_run_(\d{8}_\d{6}_\d{3})$/);
      
      expect(timestampMatch).not.toBeNull();
      if (timestampMatch) {
        const timestamp = timestampMatch[1];
        expect(timestamp).toMatch(/^\d{8}_\d{6}_\d{3}$/);
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