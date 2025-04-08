/**
 * Tests for the ConcreteConfigManager implementation
 */
import { ConcreteConfigManager } from '../ConcreteConfigManager';
import * as configFns from '../configManager';
import { ConfigError } from '../errors';
import { AppConfig, ModelConfig, ModelGroup } from '../types';

// Mock configManager functions
jest.mock('../configManager', () => ({
  loadConfig: jest.fn(),
  saveConfig: jest.fn(),
  getActiveConfigPath: jest.fn(),
  getDefaultConfigPath: jest.fn(),
  addOrUpdateModel: jest.fn(),
  removeModel: jest.fn(),
  addOrUpdateGroup: jest.fn(),
  removeGroup: jest.fn(),
  addModelToGroup: jest.fn(),
  removeModelFromGroup: jest.fn(),
}));

describe('ConcreteConfigManager', () => {
  // Sample configuration for testing
  const mockConfig: AppConfig = {
    models: [
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
        options: { temperature: 0.7 }
      }
    ]
  };

  // Sample model for testing
  const mockModel: ModelConfig = {
    provider: 'anthropic',
    modelId: 'claude-3-sonnet-20240229',
    enabled: true
  };

  // Sample group details for testing
  const mockGroupDetails: Partial<Omit<ModelGroup, 'name'>> = {
    systemPrompt: { text: 'Test system prompt' },
    models: [],
    description: 'Test group'
  };

  beforeEach(() => {
    // Reset all mocks before each test
    jest.clearAllMocks();
    
    // Set up default mock implementations
    (configFns.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (configFns.getActiveConfigPath as jest.Mock).mockResolvedValue('/path/to/config.json');
    (configFns.getDefaultConfigPath as jest.Mock).mockReturnValue('/path/to/default/config.json');
    (configFns.addOrUpdateModel as jest.Mock).mockImplementation((config, model) => ({ 
      ...config, 
      models: [...config.models, model] 
    }));
    (configFns.removeModel as jest.Mock).mockImplementation((config) => config);
    (configFns.addOrUpdateGroup as jest.Mock).mockImplementation((config) => config);
    (configFns.removeGroup as jest.Mock).mockImplementation((config) => config);
    (configFns.addModelToGroup as jest.Mock).mockImplementation((config) => config);
    (configFns.removeModelFromGroup as jest.Mock).mockImplementation((config) => config);
  });

  describe('loadConfig', () => {
    it('should delegate to configManager.loadConfig', async () => {
      const configManager = new ConcreteConfigManager();
      const options = { configPath: '/custom/path/config.json' };
      
      const result = await configManager.loadConfig(options);
      
      expect(configFns.loadConfig).toHaveBeenCalledWith(options);
      expect(result).toEqual(mockConfig);
    });

    it('should re-throw ConfigError from configManager.loadConfig', async () => {
      const configManager = new ConcreteConfigManager();
      const configError = new ConfigError('Config loading error');
      (configFns.loadConfig as jest.Mock).mockRejectedValue(configError);
      
      await expect(configManager.loadConfig()).rejects.toThrow(configError);
      expect(configFns.loadConfig).toHaveBeenCalled();
    });

    it('should wrap non-ThinktankError errors in ConfigError', async () => {
      const configManager = new ConcreteConfigManager();
      const genericError = new Error('Generic error');
      (configFns.loadConfig as jest.Mock).mockRejectedValue(genericError);
      
      try {
        await configManager.loadConfig();
        fail('Expected an error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ConfigError);
        const configError = error as ConfigError;
        expect(configError.message).toContain('Failed to load configuration');
        expect(configError.message).toContain('Generic error');
        expect(configError.cause).toBe(genericError);
      }
    });
  });

  describe('saveConfig', () => {
    it('should delegate to configManager.saveConfig', async () => {
      const configManager = new ConcreteConfigManager();
      const configPath = '/custom/path/config.json';
      
      await configManager.saveConfig(mockConfig, configPath);
      
      expect(configFns.saveConfig).toHaveBeenCalledWith(mockConfig, configPath);
    });

    it('should re-throw ConfigError from configManager.saveConfig', async () => {
      const configManager = new ConcreteConfigManager();
      const configError = new ConfigError('Config saving error');
      (configFns.saveConfig as jest.Mock).mockRejectedValue(configError);
      
      await expect(configManager.saveConfig(mockConfig)).rejects.toThrow(configError);
      expect(configFns.saveConfig).toHaveBeenCalled();
    });

    it('should wrap non-ThinktankError errors in ConfigError', async () => {
      const configManager = new ConcreteConfigManager();
      const genericError = new Error('Generic error');
      (configFns.saveConfig as jest.Mock).mockRejectedValue(genericError);
      
      try {
        await configManager.saveConfig(mockConfig);
        fail('Expected an error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ConfigError);
        const configError = error as ConfigError;
        expect(configError.message).toContain('Failed to save configuration');
        expect(configError.message).toContain('Generic error');
        expect(configError.cause).toBe(genericError);
      }
    });
  });

  describe('getActiveConfigPath', () => {
    it('should delegate to configManager.getActiveConfigPath', async () => {
      const configManager = new ConcreteConfigManager();
      const expectedPath = '/path/to/config.json';
      
      const result = await configManager.getActiveConfigPath();
      
      expect(configFns.getActiveConfigPath).toHaveBeenCalled();
      expect(result).toEqual(expectedPath);
    });

    it('should re-throw ConfigError from configManager.getActiveConfigPath', async () => {
      const configManager = new ConcreteConfigManager();
      const configError = new ConfigError('Config path error');
      (configFns.getActiveConfigPath as jest.Mock).mockRejectedValue(configError);
      
      await expect(configManager.getActiveConfigPath()).rejects.toThrow(configError);
      expect(configFns.getActiveConfigPath).toHaveBeenCalled();
    });

    it('should wrap non-ThinktankError errors in ConfigError', async () => {
      const configManager = new ConcreteConfigManager();
      const genericError = new Error('Generic error');
      (configFns.getActiveConfigPath as jest.Mock).mockRejectedValue(genericError);
      
      try {
        await configManager.getActiveConfigPath();
        fail('Expected an error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(ConfigError);
        const configError = error as ConfigError;
        expect(configError.message).toContain('Failed to get active config path');
        expect(configError.message).toContain('Generic error');
        expect(configError.cause).toBe(genericError);
      }
    });
  });

  describe('getDefaultConfigPath', () => {
    it('should delegate to configManager.getDefaultConfigPath', () => {
      const configManager = new ConcreteConfigManager();
      const expectedPath = '/path/to/default/config.json';
      
      const result = configManager.getDefaultConfigPath();
      
      expect(configFns.getDefaultConfigPath).toHaveBeenCalled();
      expect(result).toEqual(expectedPath);
    });
  });

  describe('CRUD operations', () => {
    it('should delegate addOrUpdateModel to configManager.addOrUpdateModel', () => {
      const configManager = new ConcreteConfigManager();
      
      const result = configManager.addOrUpdateModel(mockConfig, mockModel);
      
      expect(configFns.addOrUpdateModel).toHaveBeenCalledWith(mockConfig, mockModel);
      expect(result.models).toContain(mockModel);
    });

    it('should delegate removeModel to configManager.removeModel', () => {
      const configManager = new ConcreteConfigManager();
      const provider = 'openai';
      const modelId = 'gpt-4o';
      
      configManager.removeModel(mockConfig, provider, modelId);
      
      expect(configFns.removeModel).toHaveBeenCalledWith(mockConfig, provider, modelId);
    });

    it('should delegate addOrUpdateGroup to configManager.addOrUpdateGroup', () => {
      const configManager = new ConcreteConfigManager();
      const groupName = 'testGroup';
      
      configManager.addOrUpdateGroup(mockConfig, groupName, mockGroupDetails);
      
      expect(configFns.addOrUpdateGroup).toHaveBeenCalledWith(mockConfig, groupName, mockGroupDetails);
    });

    it('should delegate removeGroup to configManager.removeGroup', () => {
      const configManager = new ConcreteConfigManager();
      const groupName = 'testGroup';
      
      configManager.removeGroup(mockConfig, groupName);
      
      expect(configFns.removeGroup).toHaveBeenCalledWith(mockConfig, groupName);
    });

    it('should delegate addModelToGroup to configManager.addModelToGroup', () => {
      const configManager = new ConcreteConfigManager();
      const groupName = 'testGroup';
      const provider = 'anthropic';
      const modelId = 'claude-3-sonnet-20240229';
      
      configManager.addModelToGroup(mockConfig, groupName, provider, modelId);
      
      expect(configFns.addModelToGroup).toHaveBeenCalledWith(mockConfig, groupName, provider, modelId);
    });

    it('should delegate removeModelFromGroup to configManager.removeModelFromGroup', () => {
      const configManager = new ConcreteConfigManager();
      const groupName = 'testGroup';
      const provider = 'anthropic';
      const modelId = 'claude-3-sonnet-20240229';
      
      configManager.removeModelFromGroup(mockConfig, groupName, provider, modelId);
      
      expect(configFns.removeModelFromGroup).toHaveBeenCalledWith(mockConfig, groupName, provider, modelId);
    });
  });
});
