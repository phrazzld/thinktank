/**
 * Unit tests for configuration manager
 */
import {
  mockFsModules,
  resetVirtualFs,
  getVirtualFs,
  createFsError,
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Mock fileReader module to proxy to our virtual filesystem
jest.mock('../../utils/fileReader', () => {
  const originalModule = jest.requireActual('../../utils/fileReader');

  return {
    ...originalModule,
    fileExists: jest.fn().mockImplementation(async (path: string) => {
      try {
        getVirtualFs().statSync(path);
        return true;
      } catch (error) {
        return false;
      }
    }),
    readFileContent: jest.fn().mockImplementation(async (path: string) => {
      try {
        return getVirtualFs().readFileSync(path, 'utf-8');
      } catch (error) {
        throw createFsError('ENOENT', 'File not found', 'open', path);
      }
    }),
    writeFile: jest.fn().mockImplementation(async (path: string, content: string) => {
      try {
        // Create parent directories if they don't exist
        const dirPath = path.substring(0, path.lastIndexOf('/'));
        if (dirPath) {
          getVirtualFs().mkdirSync(dirPath, { recursive: true });
        }
        getVirtualFs().writeFileSync(path, content);
      } catch (error) {
        throw createFsError('EACCES', 'Permission denied', 'open', path);
      }
    }),
    getConfigFilePath: jest.fn().mockResolvedValue('/test/xdg/config.json'),
  };
});

jest.mock('../../utils/helpers');

// Mock constants to override default behavior for tests
jest.mock('../constants', () => ({
  DEFAULT_CONFIG: {
    models: [],
  },
  DEFAULT_CONFIG_TEMPLATE_PATH: '/test/default/template.json',
}));

// Import modules after mocking
import {
  loadConfig,
  getEnabledModels,
  filterModels,
  validateModelApiKeys,
  getEnabledGroupModels,
  getEnabledModelsFromGroups,
  findModelGroup,
  ConfigError,
  getActiveConfigPath,
  getDefaultConfigPath,
  saveConfig,
} from '../configManager';
import { fileExists, readFileContent, getConfigFilePath, writeFile } from '../../utils/fileReader';
import { getApiKey } from '../../utils/helpers';
import { AppConfig } from '../types';

// Type the mocks
const mockedFileExists = jest.mocked(fileExists);
const mockedReadFileContent = jest.mocked(readFileContent);
const mockedGetConfigFilePath = jest.mocked(getConfigFilePath);
const mockedWriteFile = jest.mocked(writeFile);
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
    resetVirtualFs();

    // Setup the virtual filesystem with default structure
    const virtualFs = getVirtualFs();

    // Create default directories
    virtualFs.mkdirSync('/test/xdg', { recursive: true });
    virtualFs.mkdirSync('/test/default', { recursive: true });

    // Create default template file
    virtualFs.writeFileSync(
      '/test/default/template.json',
      JSON.stringify({
        models: [{ provider: 'template', modelId: 'model', enabled: true }],
      })
    );

    // Create config file
    virtualFs.writeFileSync('/test/xdg/config.json', validConfigContent);
  });

  describe('loadConfig', () => {
    it('should load configuration from specified path', async () => {
      // Create a config file at a specific path
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/path/to', { recursive: true });
      virtualFs.writeFileSync('/path/to/config.json', validConfigContent);

      const config = await loadConfig({
        configPath: '/path/to/config.json',
      });

      expect(mockedFileExists).toHaveBeenCalledWith('/path/to/config.json');
      expect(mockedReadFileContent).toHaveBeenCalledWith('/path/to/config.json');
      expect(config.models).toHaveLength(1);
      expect(config.models[0].provider).toBe('testprovider');
    });

    it('should throw error when specified config file does not exist', async () => {
      // Ensure the file doesn't exist in our virtual filesystem
      expect(await fileExists('/nonexistent.json')).toBe(false);

      await expect(loadConfig({ configPath: '/nonexistent.json' })).rejects.toThrow(ConfigError);
      await expect(loadConfig({ configPath: '/nonexistent.json' })).rejects.toThrow(
        'Configuration file not found at specified path: /nonexistent.json'
      );
    });

    it('should use XDG config path when no specific path is provided and file exists', async () => {
      // Config file already exists in our setup

      const config = await loadConfig();

      // Should have checked the XDG path
      expect(mockedGetConfigFilePath).toHaveBeenCalled();
      expect(mockedFileExists).toHaveBeenCalledWith('/test/xdg/config.json');
      // Should have read the XDG path content
      expect(mockedReadFileContent).toHaveBeenCalledWith('/test/xdg/config.json');
      // Should have the expected content
      expect(config.models).toHaveLength(1);
      expect(config.models[0].provider).toBe('testprovider');
    });

    it('should create a new config file when XDG config does not exist', async () => {
      // Remove the XDG config file
      getVirtualFs().unlinkSync('/test/xdg/config.json');

      const config = await loadConfig();

      // Should have checked the XDG path
      expect(mockedGetConfigFilePath).toHaveBeenCalled();
      expect(mockedFileExists).toHaveBeenCalledWith('/test/xdg/config.json');

      // Should have checked and read the template
      expect(mockedFileExists).toHaveBeenCalledWith('/test/default/template.json');
      expect(mockedReadFileContent).toHaveBeenCalledWith('/test/default/template.json');

      // Should have written the new config file
      expect(mockedWriteFile).toHaveBeenCalledWith('/test/xdg/config.json', expect.any(String));

      // Should return the template content
      expect(config.models).toHaveLength(1);
      expect(config.models[0].provider).toBe('template');

      // Verify the file was actually written to the virtual filesystem
      expect(getVirtualFs().existsSync('/test/xdg/config.json')).toBe(true);
    });

    it('should validate configuration structure', async () => {
      // Write invalid config
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('/path/to', { recursive: true });
      virtualFs.writeFileSync(
        '/path/to/config.json',
        JSON.stringify({
          models: [{ enabled: true }], // Missing provider and modelId
        })
      );

      await expect(loadConfig({ configPath: '/path/to/config.json' })).rejects.toThrow(ConfigError);
      await expect(loadConfig({ configPath: '/path/to/config.json' })).rejects.toThrow(
        /Invalid configuration/
      );
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
      mockedGetApiKey.mockImplementation(model => {
        return model.provider === 'p1' || model.provider === 'p3' ? 'api-key' : null;
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
            ],
          },
        },
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
            models: [{ provider: 'p4', modelId: 'm4', enabled: true }],
          },
          default: {
            name: 'default',
            systemPrompt: { text: 'You are a helpful assistant.' },
            models: [
              { provider: 'p1', modelId: 'm1', enabled: true },
              { provider: 'p3', modelId: 'm3', enabled: true },
            ],
          },
        },
      };

      const groupModels = getEnabledGroupModels(config, 'default');

      expect(groupModels).toHaveLength(2);
      expect(groupModels[0].provider).toBe('p1');
      expect(groupModels[1].provider).toBe('p3');
    });

    it('should return empty array when group does not exist', () => {
      const config: AppConfig = {
        models: [
          { provider: 'p1', modelId: 'm1', enabled: true },
          { provider: 'p2', modelId: 'm2', enabled: false },
        ],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [{ provider: 'p3', modelId: 'm3', enabled: true }],
          },
        },
      };

      const groupModels = getEnabledGroupModels(config, 'nonexistent');

      expect(groupModels).toHaveLength(0);
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
            ],
          },
          creative: {
            name: 'creative',
            systemPrompt: { text: 'You are a creative assistant.' },
            models: [
              { provider: 'p3', modelId: 'm3', enabled: true }, // Duplicate
              { provider: 'p5', modelId: 'm5', enabled: true },
            ],
          },
        },
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
            models: [{ provider: 'p3', modelId: 'm3', enabled: true }],
          },
        },
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
            models: [{ provider: 'p3', modelId: 'm3', enabled: true }],
          },
        },
      };

      const groupModels = getEnabledModelsFromGroups(config, ['nonexistent', 'coding']);

      expect(groupModels).toHaveLength(1);
      expect(groupModels.some(m => m.provider === 'p3')).toBe(true);
    });
  });

  describe('findModelGroup', () => {
    it('should find the group a model belongs to', () => {
      const config: AppConfig = {
        models: [{ provider: 'p1', modelId: 'm1', enabled: true }],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [{ provider: 'p2', modelId: 'm2', enabled: true }],
          },
          creative: {
            name: 'creative',
            systemPrompt: { text: 'You are a creative assistant.' },
            models: [{ provider: 'p3', modelId: 'm3', enabled: true }],
          },
        },
      };

      const model = { provider: 'p2', modelId: 'm2', enabled: true };
      const groupInfo = findModelGroup(config, model);

      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('coding');
      expect(groupInfo?.systemPrompt.text).toBe('You are a coding assistant.');
    });

    it('should return default group for models in the models array', () => {
      const config: AppConfig = {
        models: [{ provider: 'p1', modelId: 'm1', enabled: true }],
        groups: {
          default: {
            name: 'default',
            systemPrompt: { text: 'You are a helpful assistant.' },
            models: [{ provider: 'p2', modelId: 'm2', enabled: true }],
          },
        },
      };

      const model = { provider: 'p1', modelId: 'm1', enabled: true };
      const groupInfo = findModelGroup(config, model);

      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('default');
    });

    it('should return undefined for models not in any group or the default array', () => {
      const config: AppConfig = {
        models: [{ provider: 'p1', modelId: 'm1', enabled: true }],
        groups: {
          coding: {
            name: 'coding',
            systemPrompt: { text: 'You are a coding assistant.' },
            models: [{ provider: 'p2', modelId: 'm2', enabled: true }],
          },
        },
      };

      const model = { provider: 'unknown', modelId: 'unknown', enabled: true };
      const groupInfo = findModelGroup(config, model);

      expect(groupInfo).toBeUndefined();
    });

    it('should handle configs without groups', () => {
      const config: AppConfig = {
        models: [{ provider: 'p1', modelId: 'm1', enabled: true }],
      };

      const model = { provider: 'p1', modelId: 'm1', enabled: true };
      const groupInfo = findModelGroup(config, model);

      expect(groupInfo).toBeDefined();
      expect(groupInfo?.groupName).toBe('default');
      expect(groupInfo?.systemPrompt.text).toBe('You are a helpful assistant.');
    });
  });

  describe('getDefaultConfigPath', () => {
    it('should return the path in the current working directory', () => {
      // Save original process.cwd
      const originalCwd = process.cwd;

      // Mock process.cwd
      process.cwd = jest.fn().mockReturnValue('/test/current/directory');

      const result = getDefaultConfigPath();

      expect(result).toBe('/test/current/directory/thinktank.config.json');

      // Restore original process.cwd
      process.cwd = originalCwd;
    });
  });

  describe('getActiveConfigPath', () => {
    it('should return the XDG config path', async () => {
      // getConfigFilePath is already mocked to return '/test/xdg/config.json'
      const result = await getActiveConfigPath();

      expect(result).toBe('/test/xdg/config.json');
      expect(mockedGetConfigFilePath).toHaveBeenCalled();
    });
  });

  describe('saveConfig', () => {
    it('should save configuration to specified path', async () => {
      const config: AppConfig = {
        models: [{ provider: 'test', modelId: 'model', enabled: true }],
      };

      await saveConfig(config, '/custom/path/config.json');

      expect(mockedWriteFile).toHaveBeenCalledWith('/custom/path/config.json', expect.any(String));

      // Verify the JSON was actually written to the virtual filesystem
      expect(getVirtualFs().existsSync('/custom/path/config.json')).toBe(true);

      // Verify the JSON contains our test model
      const savedContent = getVirtualFs().readFileSync('/custom/path/config.json', 'utf-8');
      expect(savedContent).toContain('test');
      expect(savedContent).toContain('model');
    });

    it('should save to XDG path when no path specified', async () => {
      const config: AppConfig = {
        models: [{ provider: 'test', modelId: 'model', enabled: true }],
      };

      await saveConfig(config);

      expect(mockedGetConfigFilePath).toHaveBeenCalled();
      expect(mockedWriteFile).toHaveBeenCalledWith('/test/xdg/config.json', expect.any(String));

      // Verify the file was actually written to the virtual filesystem
      expect(getVirtualFs().existsSync('/test/xdg/config.json')).toBe(true);
    });

    it('should throw ConfigError for invalid configuration', async () => {
      // Invalid config (missing required fields)
      const invalidConfig = {
        models: [{ enabled: true }], // Missing provider and modelId
      };

      // Cast to fool TypeScript but test the runtime validation
      await expect(saveConfig(invalidConfig as AppConfig)).rejects.toThrow(ConfigError);

      // Verify no write attempt was made
      expect(mockedWriteFile).not.toHaveBeenCalled();
    });
  });
});
