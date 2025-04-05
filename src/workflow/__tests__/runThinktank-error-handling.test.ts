/**
 * Tests specifically for error handling in runThinktank
 */
import { runThinktank, RunOptions } from '../runThinktank';
import { 
  ConfigError, 
  ApiError,
  FileSystemError,
  PermissionError,
  errorCategories,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError
} from '../../core/errors';

// Mock imports instead of requires
import * as inputHandlerModule from '../inputHandler';
import * as configManagerModule from '../../core/configManager';
import * as outputHandlerModule from '../outputHandler';
import * as modelSelectorModule from '../modelSelector';
import * as queryExecutorModule from '../queryExecutor';

// Mock dependencies
jest.mock('../../utils/fileReader');
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');
jest.mock('fs/promises');
jest.mock('../inputHandler');
jest.mock('../modelSelector');
jest.mock('../queryExecutor');
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

describe('runThinktank Error Handling', () => {
  
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should throw FileSystemError when prompt file cannot be read', async () => {
    // Create a FileSystemError using the factory function
    const fileNotFoundError = createFileNotFoundError('nonexistent.txt');
    
    // Override with error case for this specific test - this is the key behavior
    // Make this reject immediately to simulate file not found error
    (inputHandlerModule.processInput as jest.Mock).mockRejectedValue(fileNotFoundError);
    
    // Call with valid options
    const options: RunOptions = {
      input: 'nonexistent.txt'
    };
    
    // Expect it to throw with a helpful error
    await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
    // Test for more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('FileSystemError');
      expect(error.category).toBe(errorCategories.FILESYSTEM);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.filePath).toBe('nonexistent.txt');
    }
  });
  
  it('should throw PermissionError when output directory creation fails', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup directory creation to fail with permission error
    const permissionError = new PermissionError('Permission denied: Failed to create output directory', {
      suggestions: [
        'Check that you have write permissions for the directory',
        'Try using a different output path',
        'Ensure the parent directory exists and is writable'
      ]
    });
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockRejectedValue(permissionError);
    
    // Call with valid options
    const options: RunOptions = {
      input: 'prompt.txt',
      output: '/invalid/dir'
    };
    
    // Expect it to throw with a helpful error
    await expect(runThinktank(options)).rejects.toThrow(PermissionError);
    
    // Test for more specific error properties
    try {
      await runThinktank(options);
    } catch (error: any) {
      expect(error.name).toBe('PermissionError');
      expect(error.category).toBe(errorCategories.PERMISSION);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/permission denied/i);
    }
  });
  
  it('should properly handle model format errors', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Use factory function to create a model format error
    const modelFormatError = createModelFormatError(
      'openai-gpt4',  // Invalid format (missing colon)
      ['openai', 'anthropic', 'google'],
      ['openai:gpt-4o', 'anthropic:claude-3-opus']
    );
    
    // Create a mock implementation that checks if the selection has the correct model format
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation((_config: any, opts: any) => {
      if (opts.specificModel && !opts.specificModel.includes(':')) {
        throw modelFormatError;
      }
      return { models: [], warnings: [], missingApiKeyModels: [], disabledModels: [] };
    });
    
    // Expect it to throw with a helpful error
    await expect(runThinktank({ input: 'prompt.txt', specificModel: 'openai-gpt4' }))
      .rejects.toThrow(ConfigError);
    
    // Use a try-catch to test error properties
    try {
      await runThinktank({ input: 'prompt.txt', specificModel: 'openai-gpt4' }); // Invalid format
      fail('Should have thrown an error');
    } catch (error: any) {
      // Verify the correct error type is thrown
      expect(error).toBeDefined();
      expect(error.name).toBe('ConfigError');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/model format|provider:modelId/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }
  });
  
  it('should throw ConfigError for model not found errors', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Create a ConfigError instance for model not found
    const modelNotFoundError = new ConfigError('Model "openai:nonexistent-model" not found in configuration.', {
      suggestions: [
        'Check that the model is correctly spelled and exists in your configuration',
        'Available models: openai:gpt-4o, anthropic:claude-3-opus'
      ],
      examples: [
        'thinktank run prompt.txt --models=openai:gpt-4o',
        'thinktank run prompt.txt --group=default'
      ]
    });
    
    // Mock selectModels to throw for non-existent models
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation((_config: any, opts: any) => {
      if (opts.specificModel && opts.specificModel.includes('nonexistent')) {
        throw modelNotFoundError;
      }
      return { models: [], warnings: [], missingApiKeyModels: [], disabledModels: [] };
    });
    
    // Expect it to throw with a helpful error
    await expect(runThinktank({ input: 'prompt.txt', specificModel: 'openai:nonexistent-model' }))
      .rejects.toThrow(ConfigError);
    
    // Call with a non-existent model
    try {
      await runThinktank({ input: 'prompt.txt', specificModel: 'openai:nonexistent-model' });
      fail('Should have thrown an error');
    } catch (error: any) {
      // Verify the correct error type is thrown
      expect(error).toBeDefined();
      expect(error.name).toBe('ConfigError');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/model.*not found/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }
  });
  
  it('should throw ApiError for missing API key errors', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Create an ApiError instance for missing API keys using the factory function
    const missingApiKeyError = createMissingApiKeyError([
      { provider: 'openai', modelId: 'gpt-4o' },
      { provider: 'anthropic', modelId: 'claude-3-opus' }
    ]);
    
    // Mock selectModels to throw for missing API keys
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation((_config: any, opts: any) => {
      // API key issues when using validateApiKeys=true
      if (opts.validateApiKeys) {
        throw missingApiKeyError;
      }
      return { models: [], warnings: [], missingApiKeyModels: [], disabledModels: [] };
    });
    
    // Expect it to throw with a helpful error
    await expect(runThinktank({ input: 'prompt.txt' }))
      .rejects.toThrow(ApiError);
    
    // Call with options that would require API keys
    try {
      await runThinktank({ input: 'prompt.txt' });
      fail('Should have thrown an error');
    } catch (error: any) {
      // Verify the correct error type is thrown
      expect(error).toBeDefined();
      expect(error.name).toBe('ApiError');
      expect(error.category).toBe(errorCategories.API);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toMatch(/missing api key/i);
      expect(error.examples).toBeDefined();
      expect(error.examples.length).toBeGreaterThan(0);
    }
  });
  
  it('should properly handle and report model API errors', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup model selection success
    (modelSelectorModule.selectModels as jest.Mock).mockReturnValue({
      models: [
        { provider: 'openai', modelId: 'gpt-4o', enabled: true },
        { provider: 'anthropic', modelId: 'claude-3-opus', enabled: true }
      ],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: []
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Set up an API error for the second model
    const apiRateLimitError = new ApiError('Rate limit exceeded', {
      providerId: 'anthropic',
      suggestions: [
        'Try again later or reduce the number of requests',
        'Consider using a different model or provider'
      ]
    });
    
    // Setup query execution with mixed success/error results
    (queryExecutorModule.executeQueries as jest.Mock).mockResolvedValue({
      responses: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          text: 'This is a test response',
          configKey: 'openai:gpt-4o'
        },
        {
          provider: 'anthropic',
          modelId: 'claude-3-opus',
          text: '',
          error: 'Rate limit exceeded',
          errorCategory: 'API Rate Limit',
          errorTip: 'Try again later or reduce the number of requests',
          configKey: 'anthropic:claude-3-opus'
        }
      ],
      statuses: {
        'openai:gpt-4o': {
          status: 'success',
          startTime: Date.now() - 1000,
          endTime: Date.now(),
          durationMs: 1000
        },
        'anthropic:claude-3-opus': {
          status: 'error',
          message: 'Rate limit exceeded',
          detailedError: apiRateLimitError,
          startTime: Date.now() - 1200,
          endTime: Date.now(),
          durationMs: 1200
        }
      },
      timing: {
        startTime: Date.now() - 1500,
        endTime: Date.now(),
        durationMs: 1500
      }
    });
    
    // Setup file writing error
    const fileWriteError = new FileSystemError('Failed to write file', {
      filePath: '/output/directory/path/anthropic-claude-3-opus.md',
      suggestions: [
        'Check file system permissions',
        'Ensure the directory exists and is writable',
        'Verify there is enough disk space'
      ]
    });
    
    // Setup file writing with mixed success/error results
    (outputHandlerModule.writeResponsesToFiles as jest.Mock).mockResolvedValue({
      succeededWrites: 1,
      failedWrites: 1,
      files: [
        { 
          filename: 'openai-gpt-4o.md', 
          status: 'success' 
        },
        { 
          filename: 'anthropic-claude-3-opus.md', 
          status: 'error',
          error: fileWriteError
        }
      ],
      timing: {
        startTime: Date.now() - 500,
        endTime: Date.now(),
        durationMs: 500
      }
    });
    
    // Setup console formatting
    (outputHandlerModule.formatForConsole as jest.Mock).mockReturnValue('Formatted console output');
    
    // Call with standard options
    const options: RunOptions = {
      input: 'prompt.txt'
    };
    
    // Execute and verify it handles mixed results gracefully
    const result = await runThinktank(options);
    
    // Verify we get the formatted output
    expect(result).toBe('Formatted console output');
    
    // Verify queryExecutor was called
    expect(queryExecutorModule.executeQueries).toHaveBeenCalled();
    
    // Verify file writing was called with all responses, including error ones
    expect(outputHandlerModule.writeResponsesToFiles).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({ 
          provider: 'openai', 
          modelId: 'gpt-4o',
          text: 'This is a test response'
        }),
        expect.objectContaining({ 
          provider: 'anthropic', 
          modelId: 'claude-3-opus',
          error: 'Rate limit exceeded'
        })
      ]),
      expect.any(String),
      expect.any(Object)
    );
  });
  
  it('should handle group not found errors', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Create a ConfigError for group not found
    const groupError = new ConfigError('Group "nonexistent-group" not found in configuration.', {
      suggestions: [
        'Check your configuration file and make sure the group is defined',
        'Available groups: default, fast, premium',
        'Groups must be defined in your thinktank.config.json file'
      ],
      examples: [
        'thinktank run prompt.txt --group=default',
        'thinktank config groups list'
      ]
    });
    
    // Mock selectModels to throw for non-existent groups
    (modelSelectorModule.selectModels as jest.Mock).mockImplementation((_config: any, opts: any) => {
      if (opts.groupName && opts.groupName === 'nonexistent-group') {
        throw groupError;
      }
      return { models: [], warnings: [], missingApiKeyModels: [], disabledModels: [] };
    });
    
    // Expect it to throw with helpful error
    await expect(runThinktank({ input: 'prompt.txt', groupName: 'nonexistent-group' }))
      .rejects.toThrow(ConfigError);
    
    // Call with a non-existent group
    try {
      await runThinktank({ input: 'prompt.txt', groupName: 'nonexistent-group' });
      fail('Should have thrown an error');
    } catch (error: any) {
      // Verify the correct error type is thrown
      expect(error).toBeDefined();
      expect(error.name).toBe('ConfigError');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions.length).toBeGreaterThan(0);
      expect(error.message).toContain('nonexistent-group');
      expect(error.message).toContain('not found');
    }
  });
  
  it('should handle file write errors gracefully', async () => {
    // Setup success path for input processing
    (inputHandlerModule.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt',
      sourceType: 'file',
      sourcePath: 'test-prompt.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    });
    
    // Setup config loading
    (configManagerModule.loadConfig as jest.Mock).mockResolvedValue({
      models: [],
      groups: {}
    });
    
    // Setup model selection success
    (modelSelectorModule.selectModels as jest.Mock).mockReturnValue({
      models: [
        { provider: 'openai', modelId: 'gpt-4o', enabled: true }
      ],
      missingApiKeyModels: [],
      disabledModels: [],
      warnings: []
    });
    
    // Setup output directory creation
    (outputHandlerModule.createOutputDirectory as jest.Mock).mockResolvedValue('/output/directory/path');
    
    // Setup query execution success
    (queryExecutorModule.executeQueries as jest.Mock).mockResolvedValue({
      responses: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          text: 'This is a test response',
          configKey: 'openai:gpt-4o'
        }
      ],
      statuses: {
        'openai:gpt-4o': {
          status: 'success',
          startTime: Date.now() - 1000,
          endTime: Date.now(),
          durationMs: 1000
        }
      },
      timing: {
        startTime: Date.now() - 1500,
        endTime: Date.now(),
        durationMs: 1500
      }
    });
    
    // Create permission error for file writing
    const permissionError = new PermissionError('Permission denied when writing to file', {
      suggestions: [
        'Check file permissions',
        'Ensure you have write access to the directory',
        'Try using a different output path'
      ]
    });
    
    // Setup file writing to throw an error
    (outputHandlerModule.writeResponsesToFiles as jest.Mock).mockResolvedValue({
      succeededWrites: 0,
      failedWrites: 1,
      files: [
        { 
          filename: 'openai-gpt-4o.md', 
          status: 'error',
          error: permissionError
        }
      ],
      timing: {
        startTime: Date.now() - 500,
        endTime: Date.now(),
        durationMs: 500
      }
    });
    
    // Setup console formatting
    (outputHandlerModule.formatForConsole as jest.Mock).mockReturnValue('Formatted console output');
    
    // Call with standard options
    const options: RunOptions = {
      input: 'prompt.txt',
      output: '/invalid/path'
    };
    
    // Execute and verify it handles write errors gracefully
    const result = await runThinktank(options);
    
    // Verify we get the formatted output despite file write errors
    expect(result).toBe('Formatted console output');
    
    // Verify writeResponsesToFiles was called
    expect(outputHandlerModule.writeResponsesToFiles).toHaveBeenCalled();
    
    // Since the file writing failed but the API calls succeeded, 
    // the function should still return the API responses
    expect(result).toBeTruthy();
  });
  
  it('should properly propagate error causes through the call chain', async () => {
    // Create a chain of errors with causes
    const rootCause = new Error('Network connection failed');
    const apiError = new ApiError('Failed to connect to OpenAI API', {
      providerId: 'openai',
      cause: rootCause,
      suggestions: [
        'Check your internet connection',
        'Verify the API endpoint is correct',
        'Try again later'
      ]
    });
    
    // Setup input processing to throw an error
    (inputHandlerModule.processInput as jest.Mock).mockRejectedValue(apiError);
    
    // Call with valid options
    const options: RunOptions = {
      input: 'prompt.txt'
    };
    
    // Expect it to throw with a helpful error that preserves the cause chain
    await expect(runThinktank(options)).rejects.toThrow(ApiError);
    
    try {
      await runThinktank(options);
    } catch (error: any) {
      // Verify the correct error type is thrown
      expect(error).toBeDefined();
      expect(error.name).toBe('ApiError');
      expect(error.category).toBe(errorCategories.API);
      expect(error.cause).toBeDefined();
      expect(error.cause?.message).toBe('Network connection failed');
      expect(error.providerId).toBe('openai');
    }
  });
});