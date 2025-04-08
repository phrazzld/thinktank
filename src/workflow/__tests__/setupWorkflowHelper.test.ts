/**
 * Unit tests for the _setupWorkflow helper function
 */
import { _setupWorkflow } from '../runThinktankHelpers';
import * as nameGenerator from '../../utils/nameGenerator';
import * as outputHandler from '../outputHandler';
import { ConfigError, FileSystemError, PermissionError } from '../../core/errors';
import { ConfigManagerInterface, FileSystem } from '../../core/interfaces';
import { Stats } from 'fs';

// Mock dependencies
jest.mock('../../core/configManager');
jest.mock('../../utils/nameGenerator');
jest.mock('../outputHandler');

// Create a mock ConfigManagerInterface for testing
class MockConfigManager implements ConfigManagerInterface {
  loadConfig = jest.fn();
  saveConfig = jest.fn();
  getActiveConfigPath = jest.fn();
  getDefaultConfigPath = jest.fn();
}

// Create a mock FileSystem for testing
const mockFileSystem: jest.Mocked<FileSystem> = {
  readFileContent: jest.fn().mockResolvedValue('Test file content'),
  writeFile: jest.fn().mockResolvedValue(undefined),
  fileExists: jest.fn().mockResolvedValue(true),
  mkdir: jest.fn().mockResolvedValue(undefined),
  readdir: jest.fn().mockResolvedValue(['file1.txt', 'file2.txt']),
  stat: jest.fn().mockResolvedValue({ 
    isFile: () => true,
    isDirectory: () => false 
  } as unknown as Stats),
  access: jest.fn().mockResolvedValue(undefined),
  getConfigDir: jest.fn().mockResolvedValue('/mock/config/dir'),
  getConfigFilePath: jest.fn().mockResolvedValue('/mock/config/file.json')
};

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_setupWorkflow Helper', () => {
  // Reset all mocks before each test
  // Create mockConfigManager before each test
  let mockConfigManager: MockConfigManager;
  
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
    // Create a fresh mockConfigManager for each test
    mockConfigManager = new MockConfigManager();
    // Reset mockFileSystem methods
    Object.values(mockFileSystem).forEach(method => {
      if (jest.isMockFunction(method)) {
        method.mockClear();
      }
    });
  });

  // Sample app config for tests
  const sampleConfig = {
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true
      }
    ],
    groups: {
      default: {
        name: 'default',
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: []
      }
    }
  };

  it('should successfully set up the workflow', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function
    const result = await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    });

    // Verify the result
    expect(result).toEqual({
      config: sampleConfig,
      friendlyRunName: 'clever-meadow',
      outputDirectoryPath: '/fake/output/dir/clever-meadow'
    });

    // Verify mocks were called correctly
    expect(mockConfigManager.loadConfig).toHaveBeenCalledWith({ configPath: undefined });
    expect(nameGenerator.generateFunName).toHaveBeenCalled();
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith({
      outputDirectory: undefined,
      directoryIdentifier: undefined,
      friendlyRunName: 'clever-meadow'
    });

    // Verify spinner interactions
    expect(mockSpinner.info).toHaveBeenCalledTimes(2);
    expect(mockSpinner.start).toHaveBeenCalledTimes(2); // After each info
    expect(mockSpinner.text).toBe('Setup completed successfully');
  });

  it('should use provided config path', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with configPath
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        configPath: '/custom/config.json'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    });

    // Verify configPath was passed
    expect(mockConfigManager.loadConfig).toHaveBeenCalledWith({ configPath: '/custom/config.json' });
  });

  it('should use provided output directory', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/custom/output/dir/clever-meadow');

    // Call the function with output
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        output: '/custom/output/dir'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    });

    // Verify output directory was passed, with the fileSystem parameter
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(
      {
        outputDirectory: '/custom/output/dir',
        directoryIdentifier: undefined,
        friendlyRunName: 'clever-meadow'
      },
      mockFileSystem
    );
  });

  it('should use model identifier for directory naming', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with specificModel
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        specificModel: 'mock:model'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    });

    // Verify directoryIdentifier was passed
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(
      {
        outputDirectory: undefined,
        directoryIdentifier: 'mock:model',
        friendlyRunName: 'clever-meadow'
      },
      mockFileSystem
    );
  });

  it('should use group name for directory naming when provided', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with groupName
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        groupName: 'coding'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    });

    // Verify directoryIdentifier was passed
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(
      {
        outputDirectory: undefined,
        directoryIdentifier: 'coding',
        friendlyRunName: 'clever-meadow'
      },
      mockFileSystem
    );
  });

  it('should handle config loading errors', async () => {
    // Setup mocks
    const configError = new ConfigError('Configuration loading failed', {
      suggestions: ['Check your config file']
    });
    mockConfigManager.loadConfig.mockRejectedValue(configError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    })).rejects.toThrow(ConfigError);

    // Verify spinner had the right text
    expect(mockSpinner.text).toBe('Loading configuration...');
  });

  it('should handle directory creation errors', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    
    const fsError = new Error('Directory creation failed');
    (fsError as NodeJS.ErrnoException).code = 'EACCES';
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(fsError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    })).rejects.toThrow(PermissionError);

    // Verify spinner had the right text
    expect(mockSpinner.text).toBe('Creating output directory...');
  });

  it('should handle ENOENT errors', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    
    const fsError = new Error('File or directory not found');
    (fsError as NodeJS.ErrnoException).code = 'ENOENT';
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(fsError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    })).rejects.toThrow(FileSystemError);

    // Verify it produces the expected error message
    try {
      await _setupWorkflow({
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt'
        },
        configManager: mockConfigManager,
        fileSystem: mockFileSystem
      });
    } catch (error) {
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('File or directory not found');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
      }
    }
  });

  it('should handle generic errors with appropriate wrapping', async () => {
    // Setup mocks
    mockConfigManager.loadConfig.mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(new Error('Unknown error'));

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      },
      configManager: mockConfigManager,
      fileSystem: mockFileSystem
    })).rejects.toThrow(FileSystemError);

    // Verify error is properly wrapped
    try {
      await _setupWorkflow({
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt'
        },
        configManager: mockConfigManager,
        fileSystem: mockFileSystem
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
    }
  });
});
