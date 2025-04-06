/**
 * Tests for the _setupWorkflow helper function
 */
import { _setupWorkflow } from '../runThinktankHelpers';
import * as configManager from '../../core/configManager';
import * as nameGenerator from '../../utils/nameGenerator';
import * as outputHandler from '../outputHandler';
import { ConfigError, FileSystemError, PermissionError } from '../../core/errors';
import { SetupWorkflowParams } from '../runThinktankTypes';
import ora from 'ora';

// Store module paths for restoration
const configManagerPath = require.resolve('../../core/configManager');
const nameGeneratorPath = require.resolve('../../utils/nameGenerator');
const outputHandlerPath = require.resolve('../outputHandler');

// Mock dependencies
jest.mock('../../core/configManager');
jest.mock('../../utils/nameGenerator');
jest.mock('../outputHandler');
jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => {
    return {
      start: jest.fn().mockReturnThis(),
      stop: jest.fn().mockReturnThis(),
      succeed: jest.fn().mockReturnThis(),
      fail: jest.fn().mockReturnThis(),
      warn: jest.fn().mockReturnThis(),
      info: jest.fn().mockReturnThis(),
      text: '',
    };
  });
});

describe('_setupWorkflow', () => {
  // Setup mock spinner
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
  };

  // Common params for tests
  const defaultParams: SetupWorkflowParams = {
    spinner: mockSpinner as unknown as ora.Ora,
    options: {
      input: 'test-input.txt'
    }
  };

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup default mock implementations
    (configManager.loadConfig as jest.Mock).mockResolvedValue({
      models: [{ provider: 'test', modelId: 'model', enabled: true }],
      groups: { default: { name: 'default', systemPrompt: { text: 'Test prompt' }, models: [] } }
    });
    
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('test-name');
    
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/test/output/path');
  });

  afterAll(() => {
    jest.unmock('../../core/configManager');
    jest.unmock('../../utils/nameGenerator');
    jest.unmock('../outputHandler');
    jest.unmock('ora');
    
    // Clear module cache
    delete require.cache[configManagerPath];
    delete require.cache[nameGeneratorPath];
    delete require.cache[outputHandlerPath];
  });

  it('should successfully load config, generate name, and create directory', async () => {
    // Arrange
    const mockConfig = { models: [{ provider: 'mock', modelId: 'model', enabled: true }] };
    (configManager.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (nameGenerator.generateFunName as jest.Mock).mockReturnValue('mock-name');
    (outputHandler.createOutputDirectory as jest.Mock).mockResolvedValue('/mock/path');
    
    // Act
    const result = await _setupWorkflow(defaultParams);
    
    // Assert
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: undefined });
    expect(nameGenerator.generateFunName).toHaveBeenCalled();
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(expect.objectContaining({
      friendlyRunName: 'mock-name'
    }));
    
    expect(result).toEqual({
      config: mockConfig,
      friendlyRunName: 'mock-name',
      outputDirectoryPath: '/mock/path'
    });
    
    // Check spinner updates
    expect(mockSpinner.text).toBeDefined();
    expect(mockSpinner.info).toHaveBeenCalled();
  });

  it('should use custom config path if provided', async () => {
    // Arrange
    const params: SetupWorkflowParams = {
      ...defaultParams,
      options: { 
        ...defaultParams.options,
        configPath: '/custom/config.json'
      }
    };
    
    // Act
    await _setupWorkflow(params);
    
    // Assert
    expect(configManager.loadConfig).toHaveBeenCalledWith({ 
      configPath: '/custom/config.json'
    });
  });

  it('should pass directoryIdentifier based on options', async () => {
    // Arrange
    const params: SetupWorkflowParams = {
      ...defaultParams,
      options: { 
        ...defaultParams.options,
        specificModel: 'openai:gpt-4'
      }
    };
    
    // Act
    await _setupWorkflow(params);
    
    // Assert
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(expect.objectContaining({
      directoryIdentifier: 'openai:gpt-4'
    }));
  });

  it('should use groupName as directoryIdentifier if specificModel not provided', async () => {
    // Arrange
    const params: SetupWorkflowParams = {
      ...defaultParams,
      options: { 
        ...defaultParams.options,
        groupName: 'test-group'
      }
    };
    
    // Act
    await _setupWorkflow(params);
    
    // Assert
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(expect.objectContaining({
      directoryIdentifier: 'test-group'
    }));
  });

  it('should pass custom output directory if provided', async () => {
    // Arrange
    const params: SetupWorkflowParams = {
      ...defaultParams,
      options: { 
        ...defaultParams.options,
        output: '/custom/output'
      }
    };
    
    // Act
    await _setupWorkflow(params);
    
    // Assert
    expect(outputHandler.createOutputDirectory).toHaveBeenCalledWith(expect.objectContaining({
      outputDirectory: '/custom/output'
    }));
  });

  it('should throw ConfigError when config loading fails', async () => {
    // Arrange
    const configError = new ConfigError('Config error');
    (configManager.loadConfig as jest.Mock).mockRejectedValue(configError);
    
    // Act & Assert
    await expect(_setupWorkflow(defaultParams)).rejects.toThrow(ConfigError);
    expect(mockSpinner.text).toContain('Loading configuration');
  });

  it('should wrap non-ThinktankError errors during config loading', async () => {
    // Arrange
    const genericError = new Error('Generic error');
    (configManager.loadConfig as jest.Mock).mockRejectedValue(genericError);
    
    // Act & Assert
    await expect(_setupWorkflow(defaultParams)).rejects.toThrow(ConfigError);
    await expect(_setupWorkflow(defaultParams)).rejects.toMatchObject({
      message: expect.stringContaining('Generic error'),
      cause: genericError
    });
  });

  it('should throw FileSystemError when directory creation fails', async () => {
    // Arrange
    const fsError = new FileSystemError('Directory creation failed');
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(fsError);
    
    // Act & Assert
    await expect(_setupWorkflow(defaultParams)).rejects.toThrow(FileSystemError);
    expect(mockSpinner.text).toContain('Creating output directory');
  });

  it('should throw PermissionError when directory creation fails with permissions', async () => {
    // Arrange
    const permError = new Error('EACCES: permission denied');
    Object.defineProperty(permError, 'code', { value: 'EACCES' });
    (outputHandler.createOutputDirectory as jest.Mock).mockRejectedValue(permError);
    
    // Act & Assert
    await expect(_setupWorkflow(defaultParams)).rejects.toThrow(PermissionError);
    expect(mockSpinner.text).toContain('Creating output directory');
  });

  it('should update spinner with user-friendly messages', async () => {
    // Act
    await _setupWorkflow(defaultParams);
    
    // Assert - check that spinner text was updated at each step
    expect(mockSpinner.text).toBeDefined();
    // Spinner info should have been called for displaying the run name
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.start).toHaveBeenCalled();
  });
});