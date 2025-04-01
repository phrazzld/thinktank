/**
 * Unit tests for configuration manager
 */
import {
  loadConfig,
  mergeConfigs,
  getEnabledModels,
  filterModels,
  validateModelApiKeys,
  getEnabledGroupModels,
  getEnabledModelsFromGroups,
  findModelGroup,
  ConfigError,
} from '../configManager';
import { fileExists, readFileContent } from '../../utils/fileReader';
import { getApiKey } from '../../utils/helpers';
import { AppConfig } from '../types';

// Mock the file reader and helpers
jest.mock('../../utils/fileReader');
jest.mock('../../utils/helpers');

// Mock constants to override default behavior for tests
jest.mock('../constants', () => ({
  CONFIG_SEARCH_PATHS: ['/test/path1', '/test/path2', '/test/path3'],
  DEFAULT_CONFIG: {
    models: [],
  },
}));

const mockedFileExists = jest.mocked(fileExists);
const mockedReadFileContent = jest.mocked(readFileContent);
const mockedGetApiKey = jest.mocked(getApiKey);

describe('Config Manager', () => {
  const validConfigContent = JSON.stringify({
    models: [
      {
        provider: 'testprovider',
        modelId: 'testmodel',
        enabled: true,
      },
    ],
  });

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Default mock behavior
    mockedFileExists.mockResolvedValue(true);
    mockedReadFileContent.mockResolvedValue(validConfigContent);
  });

  describe('loadConfig', () => {
    it('should load configuration from specified path', async () => {
      // Turn off merging with defaults for this test
      const config = await loadConfig({ 
        configPath: '/path/to/config.json',
        mergeWithDefaults: false 
      });
      
      expect(mockedFileExists).toHaveBeenCalledWith('/path/to/config.json');
      expect(mockedReadFileContent).toHaveBeenCalledWith('/path/to/config.json');
      expect(config.models).toHaveLength(1);
      expect(config.models[0].provider).toBe('testprovider');
    });
    
    it('should throw error when specified config file does not exist', async () => {
      mockedFileExists.mockResolvedValue(false);
      
      await expect(loadConfig({ configPath: '/nonexistent.json' }))
        .rejects.toThrow(ConfigError);
      await expect(loadConfig({ configPath: '/nonexistent.json' }))
        .rejects.toThrow('Configuration file not found at specified path: /nonexistent.json');
    });
    
    it('should search multiple paths when no specific path is provided', async () => {
      // Make only the second path exist
      mockedFileExists.mockImplementation(async (path) => {
        return path === '/test/path2';
      });
      
      await loadConfig({ mergeWithDefaults: false });
      
      // Should have checked multiple paths
      expect(mockedFileExists.mock.calls.length).toBeGreaterThan(1);
      expect(mockedFileExists).toHaveBeenCalledWith('/test/path1');
      expect(mockedFileExists).toHaveBeenCalledWith('/test/path2');
      // Should have only read the content of the second path
      expect(mockedReadFileContent).toHaveBeenCalledWith('/test/path2');
    });
    
    it('should validate configuration structure', async () => {
      // Invalid config (missing required fields)
      mockedReadFileContent.mockResolvedValue(JSON.stringify({
        models: [{ enabled: true }], // Missing provider and modelId
      }));
      
      await expect(loadConfig({ configPath: '/path/to/config.json' }))
        .rejects.toThrow(ConfigError);
      await expect(loadConfig({ configPath: '/path/to/config.json' }))
        .rejects.toThrow(/Invalid configuration/);
    });
    
    it('should merge with default config when mergeWithDefaults is true', async () => {
      // Custom default config for this test
      const defaultConfig: AppConfig = {
        models: [
          {
            provider: 'defaultprovider',
            modelId: 'defaultmodel',
            enabled: false,
          },
        ],
      };
      
      // Mock structuredClone to return our custom default config
      const originalStructuredClone = global.structuredClone;
      global.structuredClone = jest.fn().mockImplementation(() => defaultConfig);
      
      mockedReadFileContent.mockResolvedValue(JSON.stringify({
        models: [
          {
            provider: 'newprovider',
            modelId: 'newmodel',
            enabled: true,
          },
        ],
      }));
      
      const config = await loadConfig({ 
        configPath: '/path/to/config.json',
        mergeWithDefaults: true,
      });
      
      // Restore original structuredClone
      global.structuredClone = originalStructuredClone;
      
      // Should contain both default model and the new model
      expect(config.models).toHaveLength(2);
      expect(config.models.some(m => m.provider === 'newprovider')).toBe(true);
      expect(config.models.some(m => m.provider === 'defaultprovider')).toBe(true);
    });
    
    it('should not merge with default config when mergeWithDefaults is false', async () => {
      mockedReadFileContent.mockResolvedValue(JSON.stringify({
        models: [
          {
            provider: 'newprovider',
            modelId: 'newmodel',
            enabled: true,
          },
        ],
      }));
      
      const config = await loadConfig({ 
        configPath: '/path/to/config.json',
        mergeWithDefaults: false,
      });
      
      // Should only contain the specified model
      expect(config.models).toHaveLength(1);
      expect(config.models[0].provider).toBe('newprovider');
    });
  });
  
  describe('mergeConfigs', () => {
    it('should merge user config with default config', () => {
      const defaultConfig: AppConfig = {
        models: [
          {
            provider: 'provider1',
            modelId: 'model1',
            enabled: false,
            options: { temperature: 0.5 },
          },
        ],
      };
      
      const userConfig: Partial<AppConfig> = {
        models: [
          {
            provider: 'provider1',
            modelId: 'model1',
            enabled: true, // Changed
          },
          {
            provider: 'provider2',
            modelId: 'model2',
            enabled: true, // New model
          },
        ],
      };
      
      const merged = mergeConfigs(defaultConfig, userConfig);
      
      expect(merged.models).toHaveLength(2);
      expect(merged.models[0].enabled).toBe(true); // Should be updated
      expect(merged.models[0].options?.temperature).toBe(0.5); // Should be preserved
      expect(merged.models[1].provider).toBe('provider2'); // New model should be added
    });
    
    it('should merge model options correctly', () => {
      const defaultConfig: AppConfig = {
        models: [
          {
            provider: 'provider1',
            modelId: 'model1',
            enabled: false,
            options: { temperature: 0.5, maxTokens: 1000 },
          },
        ],
      };
      
      const userConfig: Partial<AppConfig> = {
        models: [
          {
            provider: 'provider1',
            modelId: 'model1',
            enabled: false, // Added required field
            options: { temperature: 0.7 }, // Only override temperature
          },
        ],
      };
      
      const merged = mergeConfigs(defaultConfig, userConfig);
      
      expect(merged.models[0].options?.temperature).toBe(0.7); // Should be updated
      expect(merged.models[0].options?.maxTokens).toBe(1000); // Should be preserved
    });
  });
  
  describe('getEnabledModels', () => {
    it('should return only enabled models', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
          { provider: 'p3', modelId: 'm3', enabled: true },
        ],
      };
      
      const enabled = getEnabledModels(config);
      
      expect(enabled).toHaveLength(2);
      expect(enabled[0].provider).toBe('p1');
      expect(enabled[1].provider).toBe('p3');
    });
  });
  
  describe('filterModels', () => {
    it('should filter models by provider', () => {
      const config: AppConfig = {
        models: [
          { provider: 'openai', modelId: 'm1', enabled: true },
          { provider: 'anthropic', modelId: 'm2', enabled: true },
          { provider: 'openai', modelId: 'm3', enabled: false },
        ],
      };
      
      const filtered = filterModels(config, 'openai');
      
      expect(filtered).toHaveLength(2);
      expect(filtered.every(m => m.provider === 'openai')).toBe(true);
    });
    
    it('should filter models by modelId', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'gpt-4', enabled: true },
          { provider: 'p2', modelId: 'claude', enabled: true },
          { provider: 'p3', modelId: 'gpt-4', enabled: false },
        ],
      };
      
      const filtered = filterModels(config, 'gpt-4');
      
      expect(filtered).toHaveLength(2);
      expect(filtered.every(m => m.modelId === 'gpt-4')).toBe(true);
    });
    
    it('should filter models by combined key', () => {
      const config: AppConfig = {
        models: [
          { provider: 'openai', modelId: 'gpt-4', enabled: true },
          { provider: 'anthropic', modelId: 'claude', enabled: true },
          { provider: 'openai', modelId: 'gpt-3', enabled: false },
        ],
      };
      
      const filtered = filterModels(config, 'openai:gpt-4');
      
      expect(filtered).toHaveLength(1);
      expect(filtered[0].provider).toBe('openai');
      expect(filtered[0].modelId).toBe('gpt-4');
    });
  });
  
  describe('validateModelApiKeys', () => {
    it('should separate models with and without API keys', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: true },
          { provider: 'p3', modelId: 'm3', enabled: true },
          { provider: 'p4', modelId: 'm4', enabled: false }, // Disabled, should be ignored
        ],
      };
      
      // Mock API key retrieval
      mockedGetApiKey.mockImplementation((model) => {
        return model.provider === 'p1' || model.provider === 'p3' 
          ? 'api-key' 
          : null;
      });
      
      const { validModels, missingKeyModels } = validateModelApiKeys(config);
      
      expect(validModels).toHaveLength(2);
      expect(validModels[0].provider).toBe('p1');
      expect(validModels[1].provider).toBe('p3');
      
      expect(missingKeyModels).toHaveLength(1);
      expect(missingKeyModels[0].provider).toBe('p2');
    });
  });

  describe('getEnabledGroupModels', () => {
    it('should return enabled models from a specific group', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
              { provider: 'p4', modelId: 'm4', enabled: false },
              { provider: 'p5', modelId: 'm5', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledGroupModels(config, 'coding');
      
      expect(groupModels).toHaveLength(2);
      expect(groupModels[0].provider).toBe('p3');
      expect(groupModels[1].provider).toBe('p5');
    });
    
    it('should return enabled models from default group when group name is default', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
          { provider: 'p3', modelId: 'm3', enabled: true },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p4', modelId: 'm4', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledGroupModels(config, 'default');
      
      expect(groupModels).toHaveLength(2);
      expect(groupModels[0].provider).toBe('p1');
      expect(groupModels[1].provider).toBe('p3');
    });
    
    it('should return models from default group when group does not exist', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledGroupModels(config, 'nonexistent');
      
      expect(groupModels).toHaveLength(1);
      expect(groupModels[0].provider).toBe('p1');
    });
  });

  describe('getEnabledModelsFromGroups', () => {
    it('should return models from multiple groups without duplicates', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
              { provider: 'p4', modelId: 'm4', enabled: false },
            ]
          },
          creative: {
            name: 'creative',
            systemPrompt: { text: 'You are a creative assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true }, // Duplicate
              { provider: 'p5', modelId: 'm5', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledModelsFromGroups(config, ['coding', 'creative']);
      
      expect(groupModels).toHaveLength(2); // No duplicates
      expect(groupModels.some(m => m.provider === 'p3' && m.modelId === 'm3')).toBe(true);
      expect(groupModels.some(m => m.provider === 'p5' && m.modelId === 'm5')).toBe(true);
    });
    
    it('should return all enabled models when no groups are specified', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledModelsFromGroups(config, []);
      
      expect(groupModels).toHaveLength(1);
      expect(groupModels[0].provider).toBe('p1');
    });
    
    it('should handle nonexistent groups gracefully', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
            ]
          }
        }
      };
      
      const groupModels = getEnabledModelsFromGroups(config, ['nonexistent', 'coding']);
      
      expect(groupModels).toHaveLength(2);
      expect(groupModels.some(m => m.provider === 'p1')).toBe(true);
      expect(groupModels.some(m => m.provider === 'p3')).toBe(true);
    });
  });

  describe('findModelGroup', () => {
    it('should find the group a model belongs to', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p2', modelId: 'm2', enabled: true },
            ]
          },
          creative: {
            name: 'creative',
            systemPrompt: { text: 'You are a creative assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true },
            ]
          }
        }
      };
      
      const model = { provider: 'p2', modelId: 'm2', enabled: true };
      const groupInfo = findModelGroup(config, model);
      
      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('coding');
      expect(groupInfo?.systemPrompt.text).toBe('You are a coding assistant.');
    });
    
    it('should return default group for models in the models array', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
        ],
        groups: {
          default: {
            name: 'default',
            systemPrompt: { text: 'You are a helpful assistant.' },
            models: [
              { provider: 'p2', modelId: 'm2', enabled: true },
            ]
          }
        }
      };
      
      const model = { provider: 'p1', modelId: 'm1', enabled: true };
      const groupInfo = findModelGroup(config, model);
      
      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('default');
    });
    
    it('should return undefined for models not in any group or the default array', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [
              { provider: 'p2', modelId: 'm2', enabled: true },
            ]
          }
        }
      };
      
      const model = { provider: 'unknown', modelId: 'unknown', enabled: true };
      const groupInfo = findModelGroup(config, model);
      
      expect(groupInfo).toBeUndefined();
    });
    
    it('should handle configs without groups', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
        ]
      };
      
      const model = { provider: 'p1', modelId: 'm1', enabled: true };
      const groupInfo = findModelGroup(config, model);
      
      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('default');
      expect(groupInfo?.systemPrompt.text).toBe('You are a helpful assistant.');
    });
  });
});