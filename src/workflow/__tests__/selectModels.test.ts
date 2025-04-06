/**
 * Tests for the _selectModels helper function
 * 
 * This file covers tests for the model selection helper function, which handles
 * selecting models with warnings display and error handling.
 */
import { jest } from '@jest/globals';
import ora from 'ora';
import { AppConfig, ModelConfig } from '../../core/types';
import { ConfigError } from '../../core/errors';
import * as modelSelectorModule from '../modelSelector';
import { _selectModels } from '../runThinktankHelpers';
import { SelectModelsParams } from '../runThinktankTypes';

// Mock the entire modelSelector module
jest.mock('../modelSelector');

describe('_selectModels helper function', () => {
  // Mock spinner
  const spinner = {
    text: '',
    info: jest.fn(),
    warn: jest.fn(),
    start: jest.fn()
  } as unknown as ora.Ora;

  // Mock config
  const mockConfig: AppConfig = {
    models: [] // Empty array of models
  };

  // Mock options
  const mockOptions = {
    specificModel: 'openai:gpt-4o',
    groupName: undefined,
    input: 'input.txt' // Correct field name
  };

  // Mock model selection result
  const mockModelSelectionResult: modelSelectorModule.ModelSelectionResult = {
    models: [
      { provider: 'openai', modelId: 'gpt-4o', enabled: true }
    ] as ModelConfig[],
    missingApiKeyModels: [],
    disabledModels: [],
    warnings: []
  };

  // Reset mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    spinner.text = '';
    (spinner.info as jest.Mock).mockClear();
    (spinner.warn as jest.Mock).mockClear();
    (spinner.start as jest.Mock).mockClear();
    
    // Set up mock implementation for selectModels
    (modelSelectorModule.selectModels as jest.Mock).mockReturnValue(mockModelSelectionResult);
  });

  test('successfully selects models with no warnings', async () => {
    // Setup
    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: mockOptions
    };

    // Execute
    const result = _selectModels(params);

    // Verify
    expect(modelSelectorModule.selectModels).toHaveBeenCalledWith(mockConfig, {
      specificModel: mockOptions.specificModel,
      groupName: mockOptions.groupName,
      includeDisabled: true,
      validateApiKeys: true
    });
    
    expect(result).toEqual({
      modelSelectionResult: mockModelSelectionResult,
      modeDescription: 'Specific model: openai:gpt-4o'
    });
    
    expect(spinner.text).toContain('Models selected');
    expect(spinner.warn).not.toHaveBeenCalled();
    expect(spinner.info).toHaveBeenCalled();
    expect(spinner.start).toHaveBeenCalled();
  });

  test('selects models with warnings and displays them', async () => {
    // Setup
    const warningModelResult = {
      ...mockModelSelectionResult,
      warnings: ['Warning: Some model is missing API key']
    };
    (modelSelectorModule.selectModels as jest.Mock).mockReturnValue(warningModelResult);

    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: mockOptions
    };

    // Execute
    const result = _selectModels(params);

    // Verify
    expect(result).toEqual({
      modelSelectionResult: warningModelResult,
      modeDescription: 'Specific model: openai:gpt-4o'
    });
    
    expect(spinner.warn).toHaveBeenCalledWith(expect.stringContaining('Warning'));
    expect(spinner.info).toHaveBeenCalled();
    expect(spinner.start).toHaveBeenCalled();
  });

  test('handles group selection mode', async () => {
    // Setup
    const groupOptions = {
      ...mockOptions,
      specificModel: undefined,
      groupName: 'fast-models'
    };
    
    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: groupOptions
    };

    // Execute
    const result = _selectModels(params);

    // Verify
    expect(modelSelectorModule.selectModels).toHaveBeenCalledWith(mockConfig, {
      specificModel: undefined,
      groupName: 'fast-models',
      includeDisabled: true,
      validateApiKeys: true
    });
    
    expect(result.modeDescription).toBe('Group: fast-models');
  });

  test('handles default selection mode when no specific mode is provided', async () => {
    // Setup
    const defaultOptions = {
      input: 'input.txt'
    };
    
    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: defaultOptions
    };

    // Execute
    const result = _selectModels(params);

    // Verify
    expect(modelSelectorModule.selectModels).toHaveBeenCalledWith(mockConfig, {
      specificModel: undefined,
      groupName: undefined,
      includeDisabled: true,
      validateApiKeys: true
    });
    
    expect(result.modeDescription).toBe('All enabled models');
  });

  test('properly wraps and rethrows ModelSelectionError as ConfigError', () => {
    // Setup
    // Create a ModelSelectionError with explicit properties
    const selectionError = {
      name: 'ModelSelectionError',
      message: 'Model not found',
      suggestions: ['Try another model'],
      examples: ['openai:gpt-4']
    };
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation(() => {
      throw selectionError;
    });

    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: mockOptions
    };

    // Execute & Verify
    expect(() => _selectModels(params)).toThrow(ConfigError);
    expect(() => _selectModels(params)).toThrow('Model not found');
    
    expect(spinner.text).toContain('Selecting models');
  });

  test('properly wraps and rethrows generic errors as ConfigError', () => {
    // Setup
    const genericError = new Error('Unexpected error');
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation(() => {
      throw genericError;
    });

    const params: SelectModelsParams = {
      spinner,
      config: mockConfig,
      options: mockOptions
    };

    // Execute & Verify
    expect(() => _selectModels(params)).toThrow(ConfigError);
    expect(() => _selectModels(params)).toThrow(/Unexpected error/);
    
    expect(spinner.text).toContain('Selecting models');
  });
});