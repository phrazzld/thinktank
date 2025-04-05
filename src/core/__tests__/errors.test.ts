/**
 * Comprehensive tests for the error handling system
 * Verifies the core error classes and factory functions
 */
import { 
  ThinktankError, 
  ConfigError, 
  ApiError, 
  FileSystemError,
  ValidationError,
  NetworkError,
  PermissionError,
  InputError,
  errorCategories,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError,
  createModelNotFoundError
} from '../errors';

describe('Error categories', () => {
  test('error categories are defined', () => {
    expect(errorCategories).toBeDefined();
    expect(errorCategories.API).toBe('API');
    expect(errorCategories.CONFIG).toBe('Configuration');
    expect(errorCategories.NETWORK).toBe('Network');
    expect(errorCategories.FILESYSTEM).toBe('File System');
    expect(errorCategories.PERMISSION).toBe('Permission');
    expect(errorCategories.VALIDATION).toBe('Validation');
    expect(errorCategories.INPUT).toBe('Input');
    expect(errorCategories.UNKNOWN).toBe('Unknown');
  });
});

describe('ThinktankError', () => {
  test('creates error with correct defaults', () => {
    const error = new ThinktankError('Test error');
    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe('ThinktankError');
    expect(error.message).toBe('Test error');
    expect(error.category).toBe(errorCategories.UNKNOWN);
    expect(error.suggestions).toBeUndefined();
    expect(error.examples).toBeUndefined();
  });

  test('accepts cause in options', () => {
    const cause = new Error('Original error');
    const error = new ThinktankError('Test error', { cause });
    expect(error.cause).toBe(cause);
  });

  test('accepts category in options', () => {
    const error = new ThinktankError('Test error', { category: errorCategories.API });
    expect(error.category).toBe(errorCategories.API);
  });

  test('accepts suggestions in options', () => {
    const suggestions = ['Try this', 'Try that'];
    const error = new ThinktankError('Test error', { suggestions });
    expect(error.suggestions).toEqual(suggestions);
  });

  test('accepts examples in options', () => {
    const examples = ['Example 1', 'Example 2'];
    const error = new ThinktankError('Test error', { examples });
    expect(error.examples).toEqual(examples);
  });

  test('format method formats error correctly', () => {
    const error = new ThinktankError('Test error', {
      category: errorCategories.API,
      suggestions: ['Try this', 'Try that'],
      examples: ['Example 1', 'Example 2']
    });
    
    const formatted = error.format();
    expect(formatted).toContain('Error (API): Test error');
    expect(formatted).toContain('Try this');
    expect(formatted).toContain('Try that');
    expect(formatted).toContain('Example 1');
    expect(formatted).toContain('Example 2');
  });

  test('format method works with only suggestions', () => {
    const error = new ThinktankError('Test error', {
      suggestions: ['Try this', 'Try that']
    });
    
    const formatted = error.format();
    expect(formatted).toContain('Error (Unknown): Test error');
    expect(formatted).toContain('Suggestions:');
    expect(formatted).toContain('Try this');
    expect(formatted).toContain('Try that');
    expect(formatted).not.toContain('Examples:');
  });

  test('format method works with only examples', () => {
    const error = new ThinktankError('Test error', {
      examples: ['Example 1', 'Example 2']
    });
    
    const formatted = error.format();
    expect(formatted).toContain('Error (Unknown): Test error');
    expect(formatted).not.toContain('Suggestions:');
    expect(formatted).toContain('Examples:');
    expect(formatted).toContain('Example 1');
    expect(formatted).toContain('Example 2');
  });
});

describe('Specialized error classes', () => {
  test('ConfigError has correct defaults', () => {
    const error = new ConfigError('Config error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ConfigError');
    expect(error.category).toBe(errorCategories.CONFIG);
  });

  test('ConfigError passes suggestions and examples to base class', () => {
    const error = new ConfigError('Config error', {
      suggestions: ['Fix config.json format'],
      examples: ['{ "key": "value" }']
    });
    
    expect(error.suggestions).toEqual(['Fix config.json format']);
    expect(error.examples).toEqual(['{ "key": "value" }']);
    
    const formatted = error.format();
    expect(formatted).toContain('Error (Configuration)');
    expect(formatted).toContain('Fix config.json format');
    expect(formatted).toContain('{ "key": "value" }');
  });

  test('ApiError has correct defaults and providerId', () => {
    const error = new ApiError('API error', { providerId: 'openai' });
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ApiError');
    expect(error.category).toBe(errorCategories.API);
    expect(error.providerId).toBe('openai');
    expect(error.message).toBe('[openai] API error');
  });

  test('ApiError works without providerId', () => {
    const error = new ApiError('API error');
    expect(error.providerId).toBeUndefined();
    expect(error.message).toBe('API error');
  });

  test('ApiError passes suggestions and examples to base class', () => {
    const error = new ApiError('API error', {
      providerId: 'openai',
      suggestions: ['Check API key'],
      examples: ['OPENAI_API_KEY=your-key-here']
    });
    
    expect(error.providerId).toBe('openai');
    expect(error.suggestions).toEqual(['Check API key']);
    expect(error.examples).toEqual(['OPENAI_API_KEY=your-key-here']);
  });

  test('FileSystemError has correct defaults and filePath', () => {
    const error = new FileSystemError('File error', { filePath: '/path/to/file' });
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('FileSystemError');
    expect(error.category).toBe(errorCategories.FILESYSTEM);
    expect(error.filePath).toBe('/path/to/file');
  });

  test('FileSystemError passes suggestions and examples to base class', () => {
    const error = new FileSystemError('File not found', {
      filePath: '/path/to/file.txt',
      suggestions: ['Check file permissions'],
      examples: ['ls -la /path/to/file.txt']
    });
    
    expect(error.filePath).toBe('/path/to/file.txt');
    expect(error.suggestions).toEqual(['Check file permissions']);
    expect(error.examples).toEqual(['ls -la /path/to/file.txt']);
  });

  test('ValidationError has correct defaults', () => {
    const error = new ValidationError('Validation error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ValidationError');
    expect(error.category).toBe(errorCategories.VALIDATION);
  });

  test('NetworkError has correct defaults', () => {
    const error = new NetworkError('Network error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('NetworkError');
    expect(error.category).toBe(errorCategories.NETWORK);
  });

  test('PermissionError has correct defaults', () => {
    const error = new PermissionError('Permission error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('PermissionError');
    expect(error.category).toBe(errorCategories.PERMISSION);
  });

  test('InputError has correct defaults', () => {
    const error = new InputError('Input error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('InputError');
    expect(error.category).toBe(errorCategories.INPUT);
  });

  test('specialized errors can chain causes', () => {
    const originalError = new Error('Original error');
    const validationError = new ValidationError('Input validation failed', { cause: originalError });
    
    expect(validationError.cause).toBe(originalError);
  });
});

describe('Error factory functions', () => {
  // Save and mock process.cwd
  const originalCwd = process.cwd;
  beforeEach(() => {
    process.cwd = jest.fn().mockReturnValue('/test/current/dir');
  });
  
  afterEach(() => {
    process.cwd = originalCwd;
  });

  describe('createFileNotFoundError', () => {
    test('createFileNotFoundError creates proper FileSystemError', () => {
      const filePath = '/path/to/file.txt';
      const error = createFileNotFoundError(filePath);
      
      expect(error).toBeInstanceOf(FileSystemError);
      expect(error.message).toContain('File not found');
      expect(error.filePath).toBe(filePath);
      expect(error.category).toBe(errorCategories.FILESYSTEM);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.length).toBeGreaterThan(0);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBeGreaterThan(0);
    });

    test('createFileNotFoundError accepts custom error message', () => {
      const customMessage = 'Could not locate the configuration file';
      const error = createFileNotFoundError('/path/to/config.json', customMessage);
      
      expect(error.message).toBe(customMessage);
    });

    test('createFileNotFoundError handles relative paths correctly', () => {
      const error = createFileNotFoundError('relative/path/file.txt');
      
      // Should use current working directory in suggestions
      expect(error.suggestions?.some((s: string) => 
        s.includes('/test/current/dir/relative/path')
      )).toBe(true);
    });

    test('createFileNotFoundError adds extension suggestions for files without extensions', () => {
      const error = createFileNotFoundError('myconfig');
      
      // Should suggest adding an extension
      expect(error.suggestions?.some((s: string) => 
        s.includes('myconfig.txt') || s.includes('myconfig.md')
      )).toBe(true);
    });
  });

  describe('createModelFormatError', () => {
    test('createModelFormatError creates proper ConfigError for invalid format', () => {
      const modelSpec = 'gpt4'; // Missing colon
      const error = createModelFormatError(modelSpec);
      
      expect(error).toBeInstanceOf(ConfigError);
      expect(error.message).toContain('Invalid model format');
      expect(error.message).toContain('gpt4');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('colon'))).toBe(true);
    });

    test('createModelFormatError handles model spec ending with colon', () => {
      const modelSpec = 'openai:';
      const error = createModelFormatError(modelSpec);
      
      expect(error.message).toContain('Missing model ID after provider');
      expect(error.suggestions?.some((s: string) => 
        s.includes('Specify a model ID after the provider')
      )).toBe(true);
    });

    test('createModelFormatError handles model spec starting with colon', () => {
      const modelSpec = ':gpt4';
      const error = createModelFormatError(modelSpec);
      
      expect(error.message).toContain('Missing provider name before model ID');
      expect(error.suggestions?.some((s: string) => 
        s.includes('Specify a provider before the model ID')
      )).toBe(true);
    });

    test('createModelFormatError includes provider suggestions when available', () => {
      const modelSpec = ':gpt4'; // Missing provider
      const providers = ['openai', 'anthropic', 'google'];
      
      const error = createModelFormatError(modelSpec, providers);
      
      expect(error.suggestions?.some((s: string) => s.includes('openai'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('anthropic'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('google'))).toBe(true);
    });

    test('createModelFormatError includes available models in suggestions', () => {
      const modelSpec = 'openai:';
      const providers = ['openai'];
      const models = ['openai:gpt-4o', 'openai:gpt-3.5-turbo'];
      
      const error = createModelFormatError(modelSpec, providers, models);
      
      expect(error.suggestions?.some((s: string) => s.includes('openai:gpt-4o'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('openai:gpt-3.5-turbo'))).toBe(true);
    });

    test('createModelFormatError accepts custom error message', () => {
      const error = createModelFormatError('invalid', [], [], 'Custom error message');
      
      expect(error.message).toBe('Custom error message');
    });
  });

  describe('createMissingApiKeyError', () => {
    test('createMissingApiKeyError creates proper ApiError', () => {
      const missingModels = [
        { provider: 'openai', modelId: 'gpt-4o' },
        { provider: 'anthropic', modelId: 'claude-3-opus' }
      ];
      
      const error = createMissingApiKeyError(missingModels);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.message).toContain('Missing API keys for 2 models');
      expect(error.category).toBe(errorCategories.API);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('openai:gpt-4o'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('anthropic:claude-3-opus'))).toBe(true);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBe(2); // One example per provider
    });

    test('createMissingApiKeyError handles singular case correctly', () => {
      const missingModels = [{ provider: 'openai', modelId: 'gpt-4o' }];
      const error = createMissingApiKeyError(missingModels);
      
      expect(error.message).toContain('Missing API key for 1 model');
      expect(error.message).not.toContain('keys');
      expect(error.message).not.toContain('models');
    });

    test('createMissingApiKeyError provides provider-specific instructions', () => {
      const missingModels = [
        { provider: 'openai', modelId: 'gpt-4o' },
        { provider: 'anthropic', modelId: 'claude-3' },
        { provider: 'google', modelId: 'gemini-pro' },
        { provider: 'openrouter', modelId: 'mixtral' },
        { provider: 'unknown', modelId: 'custom-model' }
      ];
      
      const error = createMissingApiKeyError(missingModels);
      
      // Should include provider-specific instructions
      expect(error.suggestions?.some((s: string) => s.includes('platform.openai.com/api-keys'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('console.anthropic.com/keys'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('aistudio.google.com/app/apikey'))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('openrouter.ai/keys'))).toBe(true);
      
      // Should handle unknown providers
      expect(error.suggestions?.some((s: string) => s.includes('Get an API key for unknown'))).toBe(true);
      
      // Should include environment variable setup instructions
      expect(error.suggestions?.some((s: string) => s.includes('export PROVIDER_API_KEY='))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('set PROVIDER_API_KEY='))).toBe(true);
      expect(error.suggestions?.some((s: string) => s.includes('$env:PROVIDER_API_KEY'))).toBe(true);
      
      // Should include examples for all providers
      expect(error.examples).toHaveLength(5);
      expect(error.examples?.some((s: string) => s.includes('OPENAI_API_KEY'))).toBe(true);
      expect(error.examples?.some((s: string) => s.includes('UNKNOWN_API_KEY'))).toBe(true);
    });

    test('createMissingApiKeyError accepts custom error message', () => {
      const missingModels = [{ provider: 'openai', modelId: 'gpt-4o' }];
      const error = createMissingApiKeyError(missingModels, 'Custom error message');
      
      expect(error.message).toBe('Custom error message');
    });
  });

  describe('createModelNotFoundError', () => {
    test('createModelNotFoundError creates proper ConfigError', () => {
      const modelSpec = 'openai:nonexistent-model';
      const availableModels = [
        'openai:gpt-4o', 
        'openai:gpt-3.5-turbo', 
        'anthropic:claude-3-opus'
      ];
      
      const error = createModelNotFoundError(modelSpec, availableModels);
      
      expect(error).toBeInstanceOf(ConfigError);
      expect(error.message).toContain('Model "openai:nonexistent-model" not found in configuration');
      expect(error.category).toBe(errorCategories.CONFIG);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('openai:gpt-4o'))).toBe(true);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBeGreaterThan(0);
    });

    test('createModelNotFoundError includes group context when provided', () => {
      const modelSpec = 'openai:nonexistent-model';
      const groupName = 'premium';
      
      const error = createModelNotFoundError(modelSpec, [], groupName);
      
      expect(error.message).toContain(`not found in group "${groupName}"`);
      expect(error.suggestions?.some((s: string) => s.includes(groupName))).toBe(true);
    });

    test('createModelNotFoundError suggests models from same provider', () => {
      const modelSpec = 'openai:nonexistent-model';
      const availableModels = [
        'openai:gpt-4o', 
        'openai:gpt-3.5-turbo', 
        'anthropic:claude-3-opus'
      ];
      
      const error = createModelNotFoundError(modelSpec, availableModels);
      
      expect(error.suggestions?.some((s: string) => 
        s.includes('Available models from openai') && 
        s.includes('gpt-4o') && 
        s.includes('gpt-3.5-turbo')
      )).toBe(true);
    });

    test('createModelNotFoundError suggests similar models by ID', () => {
      const modelSpec = 'openai:gpt4';
      const availableModels = [
        'openai:gpt-4o', 
        'openai:gpt-4-turbo', 
        'anthropic:claude-3-opus'
      ];
      
      const error = createModelNotFoundError(modelSpec, availableModels);
      
      // Verify we have suggestions
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.length).toBeGreaterThan(0);
      
      // Check that it includes similar models from the same provider
      expect(error.suggestions?.some(s => s.includes('Available models from openai'))).toBe(true);
      expect(error.suggestions?.some(s => s.includes('gpt-4o'))).toBe(true);
      expect(error.suggestions?.some(s => s.includes('gpt-4-turbo'))).toBe(true);
    });

    test('createModelNotFoundError suggests available providers when provider not found', () => {
      const modelSpec = 'unknown:model';
      const availableModels = [
        'openai:gpt-4o', 
        'anthropic:claude-3-opus'
      ];
      
      const error = createModelNotFoundError(modelSpec, availableModels);
      
      expect(error.suggestions?.some((s: string) => s.includes('Provider "unknown" not found'))).toBe(true);
      expect(error.suggestions?.some((s: string) => 
        s.includes('Available providers') && 
        s.includes('openai') && 
        s.includes('anthropic')
      )).toBe(true);
    });

    test('createModelNotFoundError accepts custom error message', () => {
      const error = createModelNotFoundError(
        'openai:nonexistent-model',
        [],
        undefined,
        'Custom error message'
      );
      
      expect(error.message).toBe('Custom error message');
    });
  });
});

describe('Error Integration', () => {
  test('errors should chain through cause relationship', () => {
    const originalError = new Error('Network timeout');
    const networkError = new NetworkError('Connection failed', { cause: originalError });
    const apiError = new ApiError('API request failed', { 
      providerId: 'openai', 
      cause: networkError 
    });
    
    // Verify chain of errors
    expect(apiError.cause).toBe(networkError);
    expect((apiError.cause as NetworkError).cause).toBe(originalError);
    
    // Message should include provider
    expect(apiError.message).toBe('[openai] API request failed');
  });
  
  test('complex error formatting with multiple suggestions and examples', () => {
    const error = new ConfigError('Invalid configuration', {
      suggestions: [
        'Check your config file format',
        'Make sure all required fields are present',
        'Validate the model configuration'
      ],
      examples: [
        '{ "models": [{ "provider": "openai", "modelId": "gpt-4o", "enabled": true }] }',
        'thinktank config validate'
      ]
    });
    
    const formatted = error.format();
    
    // Verify all suggestions are included
    expect(formatted).toContain('Check your config file format');
    expect(formatted).toContain('Make sure all required fields are present');
    expect(formatted).toContain('Validate the model configuration');
    
    // Verify all examples are included
    expect(formatted).toContain('{ "models": [{ "provider": "openai", "modelId": "gpt-4o", "enabled": true }] }');
    expect(formatted).toContain('thinktank config validate');
  });
  
  test('factory errors contain detailed help information', () => {
    const models = [
      'openai:gpt-4o',
      'openai:gpt-3.5-turbo',
      'anthropic:claude-3-opus'
    ];
    
    const error = createModelNotFoundError('openai:gpt4', models);
    const formatted = error.format();
    
    // Should contain helpful guidance
    expect(formatted).toContain('Model "openai:gpt4" not found');
    expect(formatted).toContain('Suggestions:');
    expect(formatted).toContain('Examples:');
  });
});