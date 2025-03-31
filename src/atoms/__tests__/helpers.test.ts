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
    
    it('should generate unique directory names for consecutive calls', () => {
      const dirName1 = generateRunDirectoryName();
      
      // Small delay - sleep for a millisecond to ensure different timestamp
      const wait = (ms: number): Promise<void> => new Promise(resolve => setTimeout(resolve, ms));
      return wait(1).then(() => {
        const dirName2 = generateRunDirectoryName();
        expect(dirName1).not.toBe(dirName2);
      });
    });
  });
  
  describe('resolveOutputDirectory', () => {
    it('should use the provided output directory when specified', () => {
      // Mock the generateRunDirectoryName function to return a fixed value
      const spy = jest.spyOn(global.Date.prototype, 'toISOString').mockReturnValue('2023-01-01T12:34:56.789Z');
      
      const outputDir = resolveOutputDirectory('/custom/path');
      expect(outputDir).toBe(path.join('/custom/path', 'thinktank_run_20230101_123456_789'));
      
      spy.mockRestore();
    });
    
    it('should use the default directory when no output is specified', () => {
      // Mock the generateRunDirectoryName function to return a fixed value
      const spy = jest.spyOn(global.Date.prototype, 'toISOString').mockReturnValue('2023-01-01T12:34:56.789Z');
      
      const outputDir = resolveOutputDirectory();
      expect(outputDir).toBe(path.join(process.cwd(), 'thinktank_outputs', 'thinktank_run_20230101_123456_789'));
      
      spy.mockRestore();
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
  });
});