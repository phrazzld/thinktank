/**
 * Unit tests for the _selectModels helper function
 */
import { _selectModels } from '../runThinktankHelpers';
import * as modelSelector from '../modelSelector';
import { ConfigError } from '../../core/errors';
import { ModelSelectionError } from '../modelSelector';
import { AppConfig, ModelConfig } from '../../core/types';

// Mock dependencies
jest.mock('../modelSelector');

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_selectModels Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  // Sample app config for tests
  const sampleConfig: AppConfig = {
    models: [
      {
        provider: 'mock',
        modelId: 'mock-model',
        enabled: true,
      } as ModelConfig,
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
      } as ModelConfig,
    ],
    groups: {
      default: {
        name: 'default',
        systemPrompt: { text: 'You are a helpful assistant.' },
        models: [],
      },
    },
  };

  it('should successfully select models', () => {
    // Setup mocks
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        },
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true,
        },
      ],
      warnings: [],
      disabledModels: [],
      missingApiKeyModels: [],
    });

    // Call the function
    const result = _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
      },
    });

    // Verify the result
    expect(result.modelSelectionResult.models.length).toBe(2);
    expect(result.modeDescription).toBe('All enabled models');

    // Verify mocks were called correctly
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      sampleConfig,
      expect.objectContaining({
        includeDisabled: true,
        validateApiKeys: true,
      })
    );

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Model selection complete');
    expect(mockSpinner.info).toHaveBeenCalled();
    expect(mockSpinner.start).toHaveBeenCalled();
  });

  it('should return a flattened structure with ModelSelectionResult & modeDescription properties', () => {
    // Setup mocks
    const mockSelectionResult = {
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        },
      ],
      warnings: [],
      disabledModels: [],
      missingApiKeyModels: [],
    };

    (modelSelector.selectModels as jest.Mock).mockReturnValue(mockSelectionResult);

    // Call the function
    const result = _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
      },
    });

    // Verify we have a properly flattened structure
    expect(result).toHaveProperty('models'); // Direct property from ModelSelectionResult
    expect(result).toHaveProperty('warnings'); // Direct property from ModelSelectionResult
    expect(result).toHaveProperty('disabledModels'); // Direct property from ModelSelectionResult
    expect(result).toHaveProperty('missingApiKeyModels'); // Direct property from ModelSelectionResult
    expect(result).toHaveProperty('modeDescription'); // Our added property

    // Should have the same models array from the original selection result
    expect(result.models).toBe(mockSelectionResult.models);

    // Backward compatibility check - modelSelectionResult should contain the same models array
    expect(result.modelSelectionResult.models).toBe(result.models);
  });

  it('should select specific model when provided', () => {
    // Setup mocks
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
      models: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true,
        },
      ],
      warnings: [],
      disabledModels: [],
      missingApiKeyModels: [],
    });

    // Call the function with specificModel
    const result = _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
        specificModel: 'openai:gpt-4o',
      },
    });

    // Verify the result
    expect(result.modelSelectionResult.models.length).toBe(1);
    expect(result.modelSelectionResult.models[0].modelId).toBe('gpt-4o');
    expect(result.modeDescription).toBe('Specific model: openai:gpt-4o');

    // Verify specific model was passed correctly
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      sampleConfig,
      expect.objectContaining({
        specificModel: 'openai:gpt-4o',
      })
    );
  });

  it('should select models by group when specified', () => {
    // Setup mocks
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        },
      ],
      warnings: [],
      disabledModels: [],
      missingApiKeyModels: [],
    });

    // Call the function with groupName
    const result = _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
        groupName: 'coding',
      },
    });

    // Verify the result
    expect(result.modeDescription).toBe('Group: coding');

    // Verify group name was passed correctly
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      sampleConfig,
      expect.objectContaining({
        groupName: 'coding',
      })
    );
  });

  it('should select models by array when provided', () => {
    // Setup mocks
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        },
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true,
        },
      ],
      warnings: [],
      disabledModels: [],
      missingApiKeyModels: [],
    });

    // Call the function with models array
    const result = _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
        models: ['mock:mock-model', 'openai:gpt-4o'],
      },
    });

    // Verify the result
    expect(result.modeDescription).toBe('Selected models: mock:mock-model, openai:gpt-4o');

    // Verify models array was passed correctly
    expect(modelSelector.selectModels).toHaveBeenCalledWith(
      sampleConfig,
      expect.objectContaining({
        models: ['mock:mock-model', 'openai:gpt-4o'],
      })
    );
  });

  it('should display warnings from model selection', () => {
    // Setup mocks
    (modelSelector.selectModels as jest.Mock).mockReturnValue({
      models: [
        {
          provider: 'mock',
          modelId: 'mock-model',
          enabled: true,
        },
      ],
      warnings: ['Warning: Some models were disabled', 'API key missing for provider'],
      disabledModels: [],
      missingApiKeyModels: [],
    });

    // Call the function
    _selectModels({
      spinner: mockSpinner,
      config: sampleConfig,
      options: {
        input: 'test-prompt.txt',
      },
    });

    // Verify warnings were displayed
    expect(mockSpinner.warn).toHaveBeenCalledTimes(2);
    expect(mockSpinner.start).toHaveBeenCalledTimes(2); // After warnings + after info
  });

  it('should handle ModelSelectionError', () => {
    // Setup mocks
    const selectionError = new ModelSelectionError('No matching models found', {
      suggestions: ['Try a different model or group'],
    });

    // Manually set properties for jest mock
    selectionError.message = 'No matching models found';
    selectionError.suggestions = ['Try a different model or group'];

    (modelSelector.selectModels as jest.Mock).mockImplementation(() => {
      throw selectionError;
    });

    // Call the function and expect it to throw
    expect(() =>
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      })
    ).toThrow(ConfigError);

    // Verify proper error wrapping
    try {
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ConfigError);
      if (error instanceof ConfigError) {
        expect(error.message).toBe('No matching models found');
        expect(error.cause).toBe(selectionError);
        expect(error.suggestions).toEqual(['Try a different model or group']);
      }
    }
  });

  it('should handle ConfigError by rethrowing it', () => {
    // Setup mocks
    const configError = new ConfigError('Bad configuration');
    (modelSelector.selectModels as jest.Mock).mockImplementation(() => {
      throw configError;
    });

    // Call the function and expect it to throw the original ConfigError
    expect(() =>
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      })
    ).toThrow(configError);
  });

  it('should handle unknown errors by wrapping them in ConfigError', () => {
    // Setup mocks
    const unknownError = new Error('Something went wrong');
    (modelSelector.selectModels as jest.Mock).mockImplementation(() => {
      throw unknownError;
    });

    // Call the function and expect it to throw a ConfigError
    expect(() =>
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      })
    ).toThrow(ConfigError);

    // Verify proper error wrapping
    try {
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ConfigError);
      if (error instanceof ConfigError) {
        expect(error.message).toContain('Error selecting models');
        expect(error.cause).toBe(unknownError);
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
      }
    }
  });

  it('should handle mock ModelSelectionError objects', () => {
    // Setup mocks to return a plain object instead of an Error instance
    // This mimics how some test mocks might behave
    const mockErrorObj = {
      name: 'ModelSelectionError',
      message: 'Mock model selection error',
      suggestions: ['Try another model'],
    };
    (modelSelector.selectModels as jest.Mock).mockImplementation(() => {
      throw mockErrorObj;
    });

    // Call the function and expect it to throw ConfigError
    expect(() =>
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      })
    ).toThrow(ConfigError);

    // Verify error details
    try {
      _selectModels({
        spinner: mockSpinner,
        config: sampleConfig,
        options: {
          input: 'test-prompt.txt',
        },
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ConfigError);
      if (error instanceof ConfigError) {
        expect(error.message).toBe('Mock model selection error');
        expect(error.suggestions).toEqual(['Try another model']);
      }
    }
  });
});
