/**
 * Unit tests for LLM provider registry
 */
import { 
  registerProvider, 
  getProvider, 
  hasProvider, 
  getProviderIds, 
  getAllProviders, 
  clearRegistry,
  callProvider,
  ProviderRegistryError,
} from '../llmRegistry';
import { LLMProvider, LLMResponse, ModelConfig, ModelOptions, SystemPrompt } from '../types';
import * as configManager from '../configManager';

describe('LLM Registry', () => {
  // Sample provider implementations for testing
  class TestProvider1 implements LLMProvider {
    providerId = 'test1';
    
    async generate(
      _prompt: string, 
      modelId: string, 
      _options?: ModelOptions
    ): Promise<LLMResponse> {
      return {
        provider: this.providerId,
        modelId,
        text: `Test response from ${this.providerId} model ${modelId}`,
      };
    }
  }
  
  class TestProvider2 implements LLMProvider {
    providerId = 'test2';
    
    async generate(
      _prompt: string, 
      modelId: string, 
      _options?: ModelOptions
    ): Promise<LLMResponse> {
      return {
        provider: this.providerId,
        modelId,
        text: `Test response from ${this.providerId} model ${modelId}`,
      };
    }
  }
  
  // Clear registry before each test
  beforeEach(() => {
    clearRegistry();
  });
  
  describe('registerProvider', () => {
    it('should register a valid provider', () => {
      const provider = new TestProvider1();
      registerProvider(provider);
      
      expect(hasProvider('test1')).toBe(true);
      expect(getProvider('test1')).toBe(provider);
    });
    
    it('should throw an error when registering an undefined provider', () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect(() => registerProvider(undefined as any)).toThrow(ProviderRegistryError);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect(() => registerProvider(undefined as any))
        .toThrow('Cannot register undefined or null provider');
    });
    
    it('should throw an error when registering a provider without providerId', () => {
      // Create an object that's missing the providerId property
      const invalidProvider = {
        generate: async (): Promise<LLMResponse> => ({
          provider: 'invalid',
          modelId: 'model',
          text: 'response',
        }),
      };
      
      // Force-cast to LLMProvider to simulate a runtime error
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect(() => registerProvider(invalidProvider as any)).toThrow(ProviderRegistryError);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect(() => registerProvider(invalidProvider as any)).toThrow('Provider must have a providerId');
    });
    
    it('should throw an error when registering a duplicate provider', () => {
      const provider1 = new TestProvider1();
      const provider2 = new TestProvider1(); // Same providerId
      
      registerProvider(provider1);
      
      expect(() => registerProvider(provider2)).toThrow(ProviderRegistryError);
      expect(() => registerProvider(provider2))
        .toThrow('Provider with ID \'test1\' is already registered');
    });
  });
  
  describe('getProvider', () => {
    it('should return the registered provider', () => {
      const provider = new TestProvider1();
      registerProvider(provider);
      
      expect(getProvider('test1')).toBe(provider);
    });
    
    it('should return undefined for non-existent provider', () => {
      expect(getProvider('nonexistent')).toBeUndefined();
    });
  });
  
  describe('hasProvider', () => {
    it('should return true for registered provider', () => {
      registerProvider(new TestProvider1());
      
      expect(hasProvider('test1')).toBe(true);
    });
    
    it('should return false for non-existent provider', () => {
      expect(hasProvider('nonexistent')).toBe(false);
    });
  });
  
  describe('getProviderIds', () => {
    it('should return empty array when no providers are registered', () => {
      expect(getProviderIds()).toEqual([]);
    });
    
    it('should return all registered provider IDs', () => {
      registerProvider(new TestProvider1());
      registerProvider(new TestProvider2());
      
      const ids = getProviderIds();
      expect(ids).toHaveLength(2);
      expect(ids).toContain('test1');
      expect(ids).toContain('test2');
    });
  });
  
  describe('getAllProviders', () => {
    it('should return empty array when no providers are registered', () => {
      expect(getAllProviders()).toEqual([]);
    });
    
    it('should return all registered providers', () => {
      const provider1 = new TestProvider1();
      const provider2 = new TestProvider2();
      
      registerProvider(provider1);
      registerProvider(provider2);
      
      const providers = getAllProviders();
      expect(providers).toHaveLength(2);
      expect(providers).toContain(provider1);
      expect(providers).toContain(provider2);
    });
  });
  
  describe('clearRegistry', () => {
    it('should remove all registered providers', () => {
      registerProvider(new TestProvider1());
      registerProvider(new TestProvider2());
      
      expect(getProviderIds()).toHaveLength(2);
      
      clearRegistry();
      
      expect(getProviderIds()).toHaveLength(0);
      expect(hasProvider('test1')).toBe(false);
      expect(hasProvider('test2')).toBe(false);
    });
  });
  
  describe('callProvider', () => {
    // Mock the resolveModelOptions function
    beforeEach(() => {
      jest.spyOn(configManager, 'resolveModelOptions').mockImplementation(
        (_provider, _modelId, userOpts, groupOpts, cliOpts) => {
          return { mockResolved: true, ...userOpts, ...groupOpts, ...cliOpts };
        }
      );
    });
    
    afterEach(() => {
      jest.restoreAllMocks();
    });
    
    it('should call the provider with resolved options', async () => {
      const provider = new TestProvider1();
      const generateSpy = jest.spyOn(provider, 'generate');
      
      registerProvider(provider);
      
      const modelConfig: ModelConfig = {
        provider: 'test1',
        modelId: 'model1',
        enabled: true,
        options: { temperature: 0.5 }
      };
      
      const groupOptions: ModelOptions = { maxTokens: 2000 };
      const cliOptions: ModelOptions = { temperature: 0.8 };
      const systemPrompt: SystemPrompt = { text: 'Test system prompt' };
      
      await callProvider(
        'test1',
        'model1',
        'Test prompt',
        modelConfig,
        groupOptions,
        cliOptions,
        systemPrompt
      );
      
      // Check that resolveModelOptions was called with the right arguments
      expect(configManager.resolveModelOptions).toHaveBeenCalledWith(
        'test1',
        'model1',
        modelConfig.options,
        groupOptions,
        cliOptions
      );
      
      // Check that generate was called with the right arguments
      expect(generateSpy).toHaveBeenCalledWith(
        'Test prompt',
        'model1',
        expect.objectContaining({
          mockResolved: true,
          temperature: 0.8, // CLI option should override model config
          maxTokens: 2000    // From group options
        }),
        systemPrompt
      );
    });
    
    it('should throw an error when the provider is not found', async () => {
      await expect(callProvider('nonexistent', 'model1', 'Test prompt'))
        .rejects.toThrow(ProviderRegistryError);
      
      await expect(callProvider('nonexistent', 'model1', 'Test prompt'))
        .rejects.toThrow("Provider 'nonexistent' not found for model nonexistent:model1");
    });
  });
});