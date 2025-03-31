/**
 * Unit tests for helper functions
 */
import { getModelConfigKey, getDefaultApiKeyEnvVar, normalizeText, getApiKey } from '../helpers';
import { ModelConfig } from '../types';

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
});