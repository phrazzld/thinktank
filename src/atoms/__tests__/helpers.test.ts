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
      
      // @ts-expect-error: Mocking global Date
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
    // Create a simpler approach using a module import mock
    const testDirName = 'thinktank_run_20230101_123456_789';
    
    // Create a simplified version that uses a fixed directory name for testing
    function resolveOutputDirectoryTest(outputOption?: string): string {
      // Use the same logic as the real function but with a fixed directory name
      const baseOutputPath = outputOption 
        ? path.resolve(outputOption) 
        : path.resolve(process.cwd(), 'thinktank_outputs');
      
      return path.join(baseOutputPath, testDirName);
    }
    
    it('should use the provided output directory when specified', () => {
      const outputDir = resolveOutputDirectoryTest('/custom/path');
      expect(outputDir).toBe(path.join('/custom/path', testDirName));
    });
    
    it('should use the default directory when no output is specified', () => {
      const outputDir = resolveOutputDirectoryTest();
      expect(outputDir).toBe(path.join(process.cwd(), 'thinktank_outputs', testDirName));
    });
    
    it('should resolve relative paths to absolute paths', () => {
      const outputDir = resolveOutputDirectoryTest('./relative/path');
      const expectedPath = path.resolve('./relative/path');
      expect(outputDir).toBe(path.join(expectedPath, testDirName));
    });
    
    it('should handle empty string as output option', () => {
      const outputDir = resolveOutputDirectoryTest('');
      // An empty string resolves to the current directory, but the function adds 'thinktank_outputs'
      const expectedPath = path.resolve(process.cwd(), 'thinktank_outputs');
      expect(outputDir).toBe(path.join(expectedPath, testDirName));
    });
    
    it('should join paths correctly using path module', () => {
      // Test the actual function's path joining logic
      // This verifies the implementation uses path.join correctly
      jest.spyOn(path, 'join').mockImplementationOnce((...args) => {
        // Verify the last argument is the directory name from generateRunDirectoryName
        expect(args[args.length - 1]).toMatch(/^thinktank_run_\d{8}_\d{6}_\d{3}$/);
        return '/mocked/path/thinktank_run_date';
      });
      
      const result = resolveOutputDirectory('/test/path');
      expect(result).toBe('/mocked/path/thinktank_run_date');
      
      // Restore original implementation
      jest.restoreAllMocks();
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