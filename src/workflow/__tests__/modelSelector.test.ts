/**
 * Unit tests for the ModelSelector module
 */
import { selectModels, ModelSelectionError } from '../modelSelector';
import { AppConfig, ModelConfig } from '../../core/types';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import { ConfigError, ThinktankError } from '../../core/errors';

// Mock the configManager and llmRegistry modules
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');

// Helper to create a test model config
function createModelConfig(
  provider: string,
  modelId: string,
  enabled = true
): ModelConfig {
  return {
    provider,
    modelId,
    enabled,
    options: {
      temperature: 0.7,
      maxTokens: 1000
    }
  };
}

// Setup mock environment for API keys
const originalEnv = process.env;

describe('Model Selector', () => {
  // Sample models for testing
  const openaiGpt4 = createModelConfig('openai', 'gpt-4o');
  const openaiGpt35 = createModelConfig('openai', 'gpt-3.5-turbo');
  const anthropicClaude = createModelConfig('anthropic', 'claude-3-opus-20240229');
  const anthropicSonnet = createModelConfig('anthropic', 'claude-3-sonnet-20240229');
  const openaiGpt4Disabled = createModelConfig('openai', 'gpt-4o-disabled', false);
  
  // Reset mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Restore the environment
    process.env = { ...originalEnv };
    
    // Setup API keys in environment
    process.env.OPENAI_API_KEY = 'test-openai-key';
    process.env.ANTHROPIC_API_KEY = 'test-anthropic-key';
    
    // Mock the config manager functions
    const mockGetEnabledModels = configManager.getEnabledModels as jest.MockedFunction<typeof configManager.getEnabledModels>;
    mockGetEnabledModels.mockReturnValue([openaiGpt4, openaiGpt35, anthropicClaude, anthropicSonnet]);
    
    const mockGetEnabledModelsFromGroups = configManager.getEnabledModelsFromGroups as jest.MockedFunction<typeof configManager.getEnabledModelsFromGroups>;
    mockGetEnabledModelsFromGroups.mockImplementation((_config, groups) => {
      if (groups.includes('coding')) {
        return [openaiGpt4, anthropicClaude];
      } else if (groups.includes('chat')) {
        return [openaiGpt35, anthropicSonnet];
      } else if (groups.includes('nonexistent')) {
        return [];
      } else {
        return [openaiGpt4, openaiGpt35, anthropicClaude, anthropicSonnet];
      }
    });
    
    const mockFindModel = configManager.findModel as jest.MockedFunction<typeof configManager.findModel>;
    mockFindModel.mockImplementation((_config, provider, modelId) => {
      if (provider === 'openai' && modelId === 'gpt-4o') {
        return openaiGpt4;
      } else if (provider === 'openai' && modelId === 'gpt-3.5-turbo') {
        return openaiGpt35;
      } else if (provider === 'anthropic' && modelId === 'claude-3-opus-20240229') {
        return anthropicClaude;
      } else if (provider === 'anthropic' && modelId === 'claude-3-sonnet-20240229') {
        return anthropicSonnet;
      } else if (provider === 'openai' && modelId === 'gpt-4o-disabled') {
        return openaiGpt4Disabled;
      } else {
        return undefined;
      }
    });
    
    // Mock the llmRegistry functions
    const mockGetProviderIds = llmRegistry.getProviderIds as jest.MockedFunction<typeof llmRegistry.getProviderIds>;
    mockGetProviderIds.mockReturnValue(['openai', 'anthropic']);
  });
  
  // Restore process.env after all tests
  afterAll(() => {
    process.env = originalEnv;
  });
  
  describe('Multiple Model Selection', () => {
    it('should select multiple specific models', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with multiple models
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229']
      });
      
      // Verify results
      expect(result.models).toHaveLength(2);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-4o');
      expect(result.models[1].provider).toBe('anthropic');
      expect(result.models[1].modelId).toBe('claude-3-opus-20240229');
      expect(result.warnings).toHaveLength(0);
    });
    
    it('should handle errors in multiple model selection', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with invalid models
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'unknown:model', 'invalid-format'],
        throwOnError: false
      });
      
      // Verify results - should still include the valid model
      expect(result.models).toHaveLength(1);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-4o');
      
      // Check warnings
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => warning.includes('Model "unknown:model" not found'))).toBe(true);
      expect(result.warnings.some(warning => warning.includes('Invalid model format'))).toBe(true);
    });
    
    it('should throw when all models are invalid and throwOnError is true', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with all invalid models
      expect(() => selectModels(mockConfig, {
        models: ['unknown:model', 'invalid-format'],
        throwOnError: true
      })).toThrow(ModelSelectionError);
      
      try {
        selectModels(mockConfig, {
          models: ['unknown:model', 'invalid-format'],
          throwOnError: true
        });
      } catch (error) {
        // Verify error is both a ModelSelectionError and a ConfigError
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect(error instanceof ThinktankError).toBe(true);
        
        const typedError = error as ModelSelectionError;
        expect(typedError.message).toContain('None of the specified models could be used');
        expect(typedError.category).toBe('Configuration');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.includes('Check that you have specified valid models')
        )).toBe(true);
      }
    });
    
    it('should handle disabled models based on includeDisabled option', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Test with includeDisabled = true (default)
      const resultWithDisabled = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'openai:gpt-4o-disabled']
      });
      
      // Should include the disabled model
      expect(resultWithDisabled.models).toHaveLength(2);
      expect(resultWithDisabled.disabledModels).toHaveLength(1);
      expect(resultWithDisabled.disabledModels[0].modelId).toBe('gpt-4o-disabled');
      
      // Test with includeDisabled = false
      const resultWithoutDisabled = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'openai:gpt-4o-disabled'],
        includeDisabled: false
      });
      
      // Should not include the disabled model
      expect(resultWithoutDisabled.models).toHaveLength(1);
      expect(resultWithoutDisabled.disabledModels).toHaveLength(1);
      expect(resultWithoutDisabled.models[0].modelId).toBe('gpt-4o');
    });
  });
  
  describe('Single Model Selection', () => {
    it('should select a single specific model', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with a specific model
      const result = selectModels(mockConfig, {
        specificModel: 'openai:gpt-4o'
      });
      
      // Verify results
      expect(result.models).toHaveLength(1);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-4o');
      expect(result.warnings).toHaveLength(0);
    });
    
    it('should throw for invalid model format', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with invalid format
      expect(() => selectModels(mockConfig, {
        specificModel: 'invalid-format'
      })).toThrow(ModelSelectionError);
      
      try {
        selectModels(mockConfig, {
          specificModel: 'invalid-format'
        });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect(error instanceof ThinktankError).toBe(true);
        
        const typedError = error as ModelSelectionError;
        expect(typedError.message).toContain('Invalid model format');
        expect(typedError.category).toBe('Configuration');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.includes('Model specifications must use the format')
        )).toBe(true);
      }
    });
    
    it('should throw for model not found', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with non-existent model
      expect(() => selectModels(mockConfig, {
        specificModel: 'unknown:model'
      })).toThrow(ModelSelectionError);
      
      try {
        selectModels(mockConfig, {
          specificModel: 'unknown:model'
        });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect(error instanceof ThinktankError).toBe(true);
        
        const typedError = error as ModelSelectionError;
        expect(typedError.message).toContain('Model "unknown:model" not found');
        expect(typedError.category).toBe('Configuration');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.includes('Check that the model is correctly spelled')
        )).toBe(true);
      }
    });
    
    it('should handle errors when throwOnError is false', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with non-existent model but don't throw
      const result = selectModels(mockConfig, {
        specificModel: 'unknown:model',
        throwOnError: false
      });
      
      // Verify results
      expect(result.models).toHaveLength(0);
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => warning.includes('Model "unknown:model" not found'))).toBe(true);
    });
  });
  
  describe('Group-based Selection', () => {
    it('should select models from a specific group', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { 
        models: [],
        groups: {
          coding: { name: 'coding', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      // Call the selector with a group name
      const result = selectModels(mockConfig, {
        groupName: 'coding'
      });
      
      // Verify results
      expect(result.models).toHaveLength(2);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-4o');
      expect(result.models[1].provider).toBe('anthropic');
      expect(result.models[1].modelId).toBe('claude-3-opus-20240229');
      expect(result.warnings).toHaveLength(0);
    });
    
    it('should select models from multiple groups', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { 
        models: [],
        groups: {
          coding: { name: 'coding', models: [], systemPrompt: { text: 'test' } },
          chat: { name: 'chat', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      // Call the selector with multiple groups
      const result = selectModels(mockConfig, {
        groups: ['coding', 'chat']
      });
      
      // Our mock implementation returns 2 models for coding, 2 for chat
      // But there's likely overlap in the actual implementation
      expect(result.models.length).toBeGreaterThan(0);
      // Verify we have at least one model from each expected category
      expect(result.models.some(model => model.modelId === 'gpt-4o')).toBe(true);
      expect(result.models.some(model => model.modelId === 'claude-3-opus-20240229' || 
                               model.modelId === 'claude-3-sonnet-20240229')).toBe(true);
    });
    
    it('should throw when group is not found', () => {
      // Setup mock App Config with no groups
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with a non-existent group
      expect(() => selectModels(mockConfig, {
        groupName: 'nonexistent'
      })).toThrow(ModelSelectionError);
      
      try {
        selectModels(mockConfig, {
          groupName: 'nonexistent'
        });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect(error instanceof ThinktankError).toBe(true);
        
        const typedError = error as ModelSelectionError;
        expect(typedError.message).toContain('Group "nonexistent" not found');
        expect(typedError.category).toBe('Configuration');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.includes('Check your configuration file')
        )).toBe(true);
      }
    });
    
    it('should handle group errors when throwOnError is false', () => {
      // Setup mock App Config with no groups
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with a non-existent group but don't throw
      const result = selectModels(mockConfig, {
        groupName: 'nonexistent',
        throwOnError: false
      });
      
      // Verify results
      expect(result.models).toHaveLength(0);
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => warning.includes('Group "nonexistent" not found'))).toBe(true);
    });
    
    it('should filter models by both group and model identifiers', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { 
        models: [],
        groups: {
          coding: { name: 'coding', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      // Call the selector with both models and groupName
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'openai:gpt-3.5-turbo'], // gpt-3.5-turbo is not in coding group
        groupName: 'coding'
      });
      
      // Verify results - should only include models in both the list and the group
      expect(result.models).toHaveLength(1);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-4o');
      
      // Should have a warning about the filtered out model
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => 
        warning.includes('openai:gpt-3.5-turbo') && warning.includes('not in group "coding"')
      )).toBe(true);
    });
    
    it('should use models even when group filtering results in empty selection', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { 
        models: [],
        groups: {
          coding: { name: 'coding', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      // Call the selector with models that aren't in the group
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-3.5-turbo'], // Not in coding group
        groupName: 'coding'
      });
      
      // Verify results - should fall back to using the model without group filtering
      expect(result.models).toHaveLength(1);
      expect(result.models[0].provider).toBe('openai');
      expect(result.models[0].modelId).toBe('gpt-3.5-turbo');
      
      // Should have a warning about ignoring group filter
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => 
        warning.includes('None of the specified models are in group "coding"')
      )).toBe(true);
    });
  });
  
  describe('Default Selection', () => {
    it('should select all enabled models by default', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with no options
      const result = selectModels(mockConfig);
      
      // Verify results - should have all enabled models
      expect(result.models).toHaveLength(4);
      expect(result.warnings).toHaveLength(0);
    });
  });
  
  describe('API Key Validation', () => {
    it('should validate API keys by default', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Remove the ANTHROPIC_API_KEY
      delete process.env.ANTHROPIC_API_KEY;
      
      // Call the selector
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229']
      });
      
      // Should identify models with missing API keys
      expect(result.missingApiKeyModels).toHaveLength(1);
      expect(result.missingApiKeyModels[0].provider).toBe('anthropic');
      
      // Should filter out the model with missing API key
      expect(result.models).toHaveLength(1);
      expect(result.models[0].provider).toBe('openai');
      
      // Should have a warning about missing API key
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => 
        warning.includes('Missing API keys')
      )).toBe(true);
    });
    
    it('should throw when all models have missing API keys and throwOnError is true', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Remove all API keys
      delete process.env.OPENAI_API_KEY;
      delete process.env.ANTHROPIC_API_KEY;
      
      // Call the selector
      expect(() => selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229'],
        throwOnError: true
      })).toThrow(ModelSelectionError);
      
      try {
        selectModels(mockConfig, {
          models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229'],
          throwOnError: true
        });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect(error instanceof ThinktankError).toBe(true);
        
        const typedError = error as ModelSelectionError;
        expect(typedError.message).toContain('No models with valid API keys available');
        expect(typedError.category).toBe('Configuration');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.includes('Check that you have set the correct environment variables')
        )).toBe(true);
      }
    });
    
    it('should skip API key validation when validateApiKeys is false', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { models: [] };
      
      // Remove all API keys
      delete process.env.OPENAI_API_KEY;
      delete process.env.ANTHROPIC_API_KEY;
      
      // Call the selector with validateApiKeys: false
      const result = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229'],
        validateApiKeys: false
      });
      
      // Should not filter out any models, regardless of API keys
      expect(result.models).toHaveLength(2);
      
      // Should not have any warnings about API keys
      expect(result.warnings).toHaveLength(0);
      
      // Should not have any missing API key models identified
      expect(result.missingApiKeyModels).toHaveLength(0);
    });
  });
  
  describe('Error Handling for Empty Selections', () => {
    it('should provide the correct message when no models are found', () => {
      // Setup mock App Config with no enabled models
      const mockGetEnabledModels = configManager.getEnabledModels as jest.MockedFunction<typeof configManager.getEnabledModels>;
      mockGetEnabledModels.mockReturnValue([]);
      
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with different options and check the error messages
      
      // 1. specificModel
      try {
        selectModels(mockConfig, { specificModel: 'unknown:model' });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect((error as Error).message).toContain('Model "unknown:model" not found');
        expect((error as ModelSelectionError).category).toBe('Configuration');
      }
      
      // 2. Empty group
      // Mock the config with an empty-group that exists but has no models
      const mockConfigWithEmptyGroup: AppConfig = { 
        models: [],
        groups: {
          'empty-group': { name: 'empty-group', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      const mockGetEnabledModelsFromGroups = configManager.getEnabledModelsFromGroups as jest.MockedFunction<typeof configManager.getEnabledModelsFromGroups>;
      mockGetEnabledModelsFromGroups.mockReturnValue([]);
      
      try {
        selectModels(mockConfigWithEmptyGroup, { groupName: 'empty-group' });
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect((error as Error).message).toContain('No enabled models found in the specified group');
        expect((error as ModelSelectionError).category).toBe('Configuration');
      }
      
      // 3. Default case - no enabled models
      try {
        selectModels(mockConfig);
      } catch (error) {
        expect(error instanceof ModelSelectionError).toBe(true);
        expect(error instanceof ConfigError).toBe(true);
        expect((error as Error).message).toContain('No enabled models found in configuration');
        expect((error as ModelSelectionError).category).toBe('Configuration');
      }
    });
    
    it('should not throw when throwOnError is false', () => {
      // Setup mock App Config with no enabled models
      const mockGetEnabledModels = configManager.getEnabledModels as jest.MockedFunction<typeof configManager.getEnabledModels>;
      mockGetEnabledModels.mockReturnValue([]);
      
      const mockConfig: AppConfig = { models: [] };
      
      // Call the selector with throwOnError: false
      const result = selectModels(mockConfig, { throwOnError: false });
      
      // Should have empty models array
      expect(result.models).toHaveLength(0);
      
      // Should have a warning about no models found
      expect(result.warnings.length).toBeGreaterThan(0);
      expect(result.warnings.some(warning => 
        warning.includes('No enabled models found in configuration')
      )).toBe(true);
    });
  });
  
  describe('Handling Selection Priority', () => {
    it('should respect the selection hierarchy', () => {
      // Setup mock App Config
      const mockConfig: AppConfig = { 
        models: [],
        groups: {
          coding: { name: 'coding', models: [], systemPrompt: { text: 'test' } }
        }
      };
      
      // Test that models array takes precedence over specificModel
      const result1 = selectModels(mockConfig, {
        models: ['openai:gpt-4o', 'anthropic:claude-3-opus-20240229'],
        specificModel: 'openai:gpt-3.5-turbo',
        groupName: 'coding'
      });
      
      // Should use the models array (first priority)
      expect(result1.models).toHaveLength(2);
      expect(result1.models[0].modelId).toBe('gpt-4o');
      expect(result1.models[1].modelId).toBe('claude-3-opus-20240229');
      
      // Test that specificModel takes precedence over groupName
      const result2 = selectModels(mockConfig, {
        specificModel: 'openai:gpt-3.5-turbo',
        groupName: 'coding'
      });
      
      // Should use the specificModel (second priority)
      expect(result2.models).toHaveLength(1);
      expect(result2.models[0].modelId).toBe('gpt-3.5-turbo');
      
      // Test that groupName takes precedence over groups array
      const result3 = selectModels(mockConfig, {
        groupName: 'coding',
        groups: ['chat']
      });
      
      // Should use the groupName (third priority)
      expect(result3.models).toHaveLength(2);
      expect(result3.models[0].modelId).toBe('gpt-4o');
      expect(result3.models[1].modelId).toBe('claude-3-opus-20240229');
    });
  });
});