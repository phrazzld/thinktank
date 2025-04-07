/**
 * Unit tests for the _setupWorkflow helper function
 */
import { _setupWorkflow } from '../runThinktankHelpers';
import * as configManager from '../../core/configManager';
import * as nameGenerator from '../../utils/nameGenerator';
import * as outputHandler from '../outputHandler';
import { ConfigError, FileSystemError, PermissionError } from '../../core/errors';

// Mock dependencies
jest.mock('../../core/configManager');
jest.mock('../../utils/nameGenerator');
jest.mock('../outputHandler');

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_setupWorkflow Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
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
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function
    const result = await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      }
    });

    // Verify the result
    expect(result).toEqual({
      config: sampleConfig,
      friendlyRunName: 'clever-meadow',
      outputDirectoryPath: '/fake/output/dir/clever-meadow'
    });

    // Verify mocks were called correctly
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: undefined });
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
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with configPath
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        configPath: '/custom/config.json'
      }
    });

    // Verify configPath was passed
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: '/custom/config.json' });
  });

  it('should use provided output directory', async () => {
    // Setup mocks
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/custom/output/dir/clever-meadow');

    // Call the function with output
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        output: '/custom/output/dir'
      }
    });

    // Verify output directory was passed
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith({
      outputDirectory: '/custom/output/dir',
      directoryIdentifier: undefined,
      friendlyRunName: 'clever-meadow'
    });
  });

  it('should use model identifier for directory naming', async () => {
    // Setup mocks
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with specificModel
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        specificModel: 'mock:model'
      }
    });

    // Verify directoryIdentifier was passed
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith({
      outputDirectory: undefined,
      directoryIdentifier: 'mock:model',
      friendlyRunName: 'clever-meadow'
    });
  });

  it('should use group name for directory naming when provided', async () => {
    // Setup mocks
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/fake/output/dir/clever-meadow');

    // Call the function with groupName
    await _setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt',
        groupName: 'coding'
      }
    });

    // Verify directoryIdentifier was passed
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith({
      outputDirectory: undefined,
      directoryIdentifier: 'coding',
      friendlyRunName: 'clever-meadow'
    });
  });

  it('should handle config loading errors', async () => {
    // Setup mocks
    const configError = new ConfigError('Configuration loading failed', {
      suggestions: ['Check your config file']
    });
    (configManager.loadConfig as jest.Mock).mockRejectedValue(configError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(ConfigError);

    // Verify spinner had the right text
    expect(mockSpinner.text).toBe('Loading configuration...');
  });

  it('should handle directory creation errors', async () => {
    // Setup mocks
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    
    const fsError = new Error('Directory creation failed');
    (fsError as NodeJS.ErrnoException).code = 'EACCES';
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(fsError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(PermissionError);

    // Verify spinner had the right text
    expect(mockSpinner.text).toBe('Creating output directory...');
  });

  it('should handle ENOENT errors', async () => {
    // Setup mocks
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    
    const fsError = new Error('File or directory not found');
    (fsError as NodeJS.ErrnoException).code = 'ENOENT';
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(fsError);

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(FileSystemError);

    // Verify it produces the expected error message
    try {
      await _setupWorkflow({
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt'
        }
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
    (configManager.loadConfig as jest.Mock).mockResolvedValue(sampleConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('clever-meadow');
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(new Error('Unknown error'));

    // Call the function and expect it to throw
    await expect(_setupWorkflow({
      spinner: mockSpinner,
      options: {
        input: 'test-prompt.txt'
      }
    })).rejects.toThrow(FileSystemError);

    // Verify error is properly wrapped
    try {
      await _setupWorkflow({
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt'
        }
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
    }
  });
});
