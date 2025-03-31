/**
 * Unit tests for the list models workflow
 */
import { listAvailableModels } from '../listModelsWorkflow';
import * as configManager from '../../organisms/configManager';
import * as llmRegistry from '../../organisms/llmRegistry';
import * as outputFormatter from '../../molecules/outputFormatter';
import { AppConfig, LLMProvider } from '../../atoms/types';

// Mock dependencies
jest.mock('../../organisms/configManager');
jest.mock('../../organisms/llmRegistry');
jest.mock('../../molecules/outputFormatter');

describe('List Models Workflow', () => {
  // Setup common test data
  const mockConfig: AppConfig = {
    models: [
      { provider: 'openai', modelId: 'gpt-4o', enabled: true },
      { provider: 'openai', modelId: 'gpt-3.5-turbo', enabled: false },
      { provider: 'anthropic', modelId: 'claude-3-opus-20240229', enabled: true },
      { provider: 'nonexistent', modelId: 'model', enabled: true },
    ]
  };
  
  // Define mock providers
  const mockOpenAIProvider: LLMProvider = {
    providerId: 'openai',
    generate: jest.fn(),
    listModels: jest.fn().mockResolvedValue([
      { id: 'gpt-4o', description: 'Latest GPT-4' },
      { id: 'gpt-3.5-turbo', description: 'GPT-3.5 Turbo' }
    ])
  };
  
  const mockAnthropicProvider: LLMProvider = {
    providerId: 'anthropic',
    generate: jest.fn(),
    listModels: jest.fn().mockResolvedValue([
      { id: 'claude-3-opus-20240229', description: 'Claude 3 Opus' },
      { id: 'claude-3-sonnet-20240229', description: 'Claude 3 Sonnet' }
    ])
  };
  
  // Provider without listModels method
  const mockLegacyProvider: LLMProvider = {
    providerId: 'legacy',
    generate: jest.fn()
    // No listModels method
  };
  
  const mockProviders: Record<string, LLMProvider> = {
    'openai': mockOpenAIProvider,
    'anthropic': mockAnthropicProvider,
    'legacy': mockLegacyProvider
  };
  
  beforeEach(() => {
    // Reset mocks
    jest.resetAllMocks();
    
    // Set up mock implementations
    (configManager.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId: string) => mockProviders[providerId]);
    (configManager.getApiKey as jest.Mock).mockImplementation((config) => {
      if (config.provider === 'openai') return 'openai-api-key';
      if (config.provider === 'anthropic') return 'anthropic-api-key';
      if (config.provider === 'legacy') return 'legacy-api-key';
      return undefined;
    });
    (outputFormatter.formatModelList as jest.Mock).mockImplementation(() => 'Formatted model list');
  });
  
  it('should fetch models from all configured providers by default', async () => {
    const result = await listAvailableModels({});
    
    // Should load config
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: undefined });
    
    // Should get the providers from the registry
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('openai');
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('anthropic');
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('nonexistent');
    
    // Should call listModels on providers that have it
    expect(mockOpenAIProvider.listModels).toHaveBeenCalledWith('openai-api-key');
    expect(mockAnthropicProvider.listModels).toHaveBeenCalledWith('anthropic-api-key');
    
    // Should format the results
    expect(outputFormatter.formatModelList).toHaveBeenCalled();
    expect(result).toBe('Formatted model list');
  });
  
  it('should filter providers by the specified provider ID', async () => {
    await listAvailableModels({ provider: 'anthropic' });
    
    // Should only get anthropic provider
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('anthropic');
    expect(llmRegistry.getProvider).not.toHaveBeenCalledWith('openai');
    
    // Should only call listModels on anthropic
    expect(mockAnthropicProvider.listModels).toHaveBeenCalledWith('anthropic-api-key');
    expect(mockOpenAIProvider.listModels).not.toHaveBeenCalled();
  });
  
  it('should use a custom config file when specified', async () => {
    await listAvailableModels({ config: 'custom-config.json' });
    
    // Should load the custom config
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: 'custom-config.json' });
  });
  
  it('should handle providers without listModels method', async () => {
    // Modify mock config to include legacy provider
    const configWithLegacy = {
      ...mockConfig,
      models: [
        ...mockConfig.models,
        { provider: 'legacy', modelId: 'legacy-model', enabled: true }
      ]
    };
    (configManager.loadConfig as jest.Mock).mockResolvedValue(configWithLegacy);
    
    await listAvailableModels({});
    
    // Should get all providers
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('legacy');
    
    // Should call formatModelList with appropriate errors for providers without listModels
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('legacy');
    expect(formatCallArg.legacy).toHaveProperty('error');
    expect(formatCallArg.legacy.error).toBe("Provider 'legacy' does not support listing models");
  });
  
  it('should handle missing API keys gracefully', async () => {
    // Mock getApiKey to return undefined for anthropic
    (configManager.getApiKey as jest.Mock).mockImplementation((config) => {
      if (config.provider === 'openai') return 'openai-api-key';
      return undefined;
    });
    
    await listAvailableModels({});
    
    // Should not try to call listModels on anthropic
    expect(mockAnthropicProvider.listModels).not.toHaveBeenCalled();
    
    // Should include error in formatted results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg.anthropic).toHaveProperty('error');
    expect(formatCallArg.anthropic.error).toContain('API key');
  });
  
  it('should handle API errors from providers', async () => {
    // Make anthropic.listModels throw an error
    (mockAnthropicProvider.listModels as jest.Mock).mockRejectedValue(new Error('API error'));
    
    await listAvailableModels({});
    
    // Should include error in formatted results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg.anthropic).toHaveProperty('error');
    expect(formatCallArg.anthropic.error).toContain('API error');
  });
});