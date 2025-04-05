/**
 * Centralized error handling system for thinktank
 * 
 * This module defines a consistent error hierarchy and error handling utilities
 * to improve error reporting and troubleshooting across the application.
 * 
 * The error system is built around a hierarchy of error classes extending from the base
 * {@link ThinktankError} class, with specialized subclasses for different error categories.
 * 
 * Key features:
 * - Consistent error categorization with {@link errorCategories}
 * - Rich error objects with suggestions and examples
 * - Support for error chaining with the `cause` property
 * - Standardized formatting through the {@link ThinktankError.format} method
 * - Factory functions for common error types
 * 
 * @example
 * ```typescript
 * // Using error class directly
 * throw new ConfigError('Invalid configuration', {
 *   suggestions: ['Check your thinktank.config.json file']
 * });
 * 
 * // Using a factory function
 * throw createFileNotFoundError('/path/to/missing-file.txt');
 * ```
 * 
 * @module errors
 */

import { colors } from '../utils/consoleUtils';

/**
 * Error categories for consistent categorization across the application.
 * 
 * These categories are used to classify errors in a standardized way,
 * allowing for consistent error handling, display, and filtering.
 * 
 * @property {string} API - Errors related to external API interactions (e.g., OpenAI, Anthropic)
 * @property {string} CONFIG - Configuration-related errors (e.g., invalid settings, missing config)
 * @property {string} NETWORK - Network connectivity issues (e.g., timeouts, connection failures)
 * @property {string} FILESYSTEM - File system operation errors (e.g., file not found, permission denied)
 * @property {string} PERMISSION - Permission-related errors (e.g., insufficient permissions)
 * @property {string} VALIDATION - Input validation errors (e.g., invalid format, schema violations)
 * @property {string} INPUT - User input processing errors (e.g., invalid prompts)
 * @property {string} UNKNOWN - Uncategorized or internal errors
 */
export const errorCategories = {
  API: 'API',
  CONFIG: 'Configuration',
  NETWORK: 'Network',
  FILESYSTEM: 'File System',
  PERMISSION: 'Permission',
  VALIDATION: 'Validation',
  INPUT: 'Input',
  UNKNOWN: 'Unknown',
};

/**
 * Base error class for all thinktank errors.
 * 
 * This class extends the native JavaScript Error class and provides additional
 * properties and methods for enhanced error reporting, troubleshooting assistance,
 * and consistent formatting across the application.
 * 
 * All specialized error types in the thinktank application should extend this class
 * rather than the native Error class to ensure consistent behavior and properties.
 * 
 * @example
 * ```typescript
 * // Creating a basic ThinktankError
 * const error = new ThinktankError('Something went wrong');
 * 
 * // Creating a more detailed error
 * const detailedError = new ThinktankError('Configuration file is invalid', {
 *   category: errorCategories.CONFIG,
 *   suggestions: [
 *     'Check the JSON syntax in your configuration file',
 *     'Ensure all required fields are present'
 *   ],
 *   examples: [
 *     '{ "models": [], "groups": {} }'
 *   ]
 * });
 * 
 * // Error with a cause
 * try {
 *   JSON.parse(invalidJson);
 * } catch (parseError) {
 *   throw new ThinktankError('Failed to parse configuration', {
 *     cause: parseError,
 *     category: errorCategories.CONFIG
 *   });
 * }
 * ```
 * 
 * @extends {Error}
 */
export class ThinktankError extends Error {
  /**
   * The error category for grouping similar errors.
   * 
   * This property allows for categorization of errors to help with filtering,
   * handling, and displaying errors in a more structured way. Default is 'Unknown'.
   * 
   * @type {string}
   * @default errorCategories.UNKNOWN
   */
  category: string = errorCategories.UNKNOWN;
  
  /**
   * List of suggestions to help resolve the error.
   * 
   * These suggestions are displayed to the user to provide actionable
   * guidance on how to fix the issue.
   * 
   * @type {string[] | undefined}
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands, configurations, or usage patterns.
   * 
   * These examples help users understand the correct way to use the
   * functionality that caused the error.
   * 
   * @type {string[] | undefined}
   */
  examples?: string[];

  /**
   * The original error that caused this error.
   * 
   * This property facilitates error chaining, allowing for more detailed
   * error diagnostics by preserving the underlying cause.
   * 
   * @type {Error | undefined}
   */
  cause?: Error;
  
  /**
   * Creates a new ThinktankError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.category - The error category from errorCategories
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    category?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message);
    this.name = 'ThinktankError';
    
    if (options?.cause) {
      this.cause = options.cause;
    }
    
    if (options?.category) {
      this.category = options.category;
    }
    
    if (options?.suggestions) {
      this.suggestions = options.suggestions;
    }
    
    if (options?.examples) {
      this.examples = options.examples;
    }
  }
  
  /**
   * Formats the error for display in CLI and logging contexts.
   * 
   * This method generates a user-friendly, formatted representation of the error
   * that includes the error message, category, suggestions, and examples.
   * The output uses ANSI colors for better readability in terminal environments.
   * 
   * @returns A formatted string representation of the error
   * 
   * @example
   * ```typescript
   * const error = new ConfigError('Invalid model format');
   * console.log(error.format());
   * // Output: 
   * // Error (Configuration): Invalid model format
   * ```
   */
  format(): string {
    let output = `${colors.red.bold('Error')} (${colors.yellow(this.category)}): ${this.message}`;
    
    if (this.suggestions?.length) {
      output += '\n\nSuggestions:';
      this.suggestions.forEach(suggestion => {
        output += `\n  • ${suggestion}`;
      });
    }
    
    if (this.examples?.length) {
      output += '\n\nExamples:';
      this.examples.forEach(example => {
        output += `\n  - ${example}`;
      });
    }
    
    return output;
  }
}

/**
 * Configuration-related error class.
 * 
 * This error class is used for issues related to the application configuration,
 * such as invalid configuration files, missing required properties, or invalid
 * configuration values.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new ConfigError('Invalid model configuration');
 * 
 * // With suggestions and examples
 * throw new ConfigError('Required configuration file not found', {
 *   suggestions: [
 *     'Create a thinktank.config.json file in your project root',
 *     'Copy the template from the examples directory'
 *   ],
 *   examples: [
 *     'cp templates/thinktank.config.default.json ./thinktank.config.json'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class ConfigError extends ThinktankError {
  /**
   * Creates a new ConfigError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.CONFIG,
    });
    this.name = 'ConfigError';
  }
}

/**
 * API interaction error class.
 * 
 * This error class is used for issues related to external API interactions,
 * such as authentication failures, rate limits, or server errors from LLM
 * providers like OpenAI, Anthropic, or Google.
 * 
 * The `providerId` property allows for provider-specific error handling and
 * guidance.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new ApiError('Failed to generate response', {
 *   providerId: 'openai'
 * });
 * 
 * // With detailed options
 * throw new ApiError('API rate limit exceeded', {
 *   providerId: 'anthropic',
 *   suggestions: [
 *     'Wait and try again later',
 *     'Reduce the frequency of requests',
 *     'Consider upgrading your API tier'
 *   ],
 *   cause: originalError
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class ApiError extends ThinktankError {
  /**
   * The identifier of the provider that generated the error.
   * 
   * This allows for provider-specific error handling and guidance
   * (e.g., 'openai', 'anthropic', 'google').
   * 
   * @type {string | undefined}
   */
  providerId?: string;
  
  /**
   * Creates a new ApiError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.providerId - Identifier of the provider that generated the error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    providerId?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    // Prefix the message with the provider ID if available
    const formattedMessage = options?.providerId 
      ? `[${options.providerId}] ${message}`
      : message;
    
    super(formattedMessage, {
      ...options,
      category: errorCategories.API,
    });
    
    this.name = 'ApiError';
    this.providerId = options?.providerId;
  }
}

/**
 * File system related error class.
 * 
 * This error class is used for issues related to file system operations,
 * such as file not found, permission denied, or directory creation failures.
 * 
 * The `filePath` property contains the path to the file or directory that 
 * caused the error, which can be used for error reporting and troubleshooting.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new FileSystemError('Failed to read file', {
 *   filePath: '/path/to/file.txt'
 * });
 * 
 * // With suggestions
 * throw new FileSystemError('Permission denied while writing to file', {
 *   filePath: '/path/to/file.txt',
 *   suggestions: [
 *     'Check file permissions',
 *     'Ensure you have write access to the directory'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class FileSystemError extends ThinktankError {
  /**
   * The path to the file or directory that caused the error.
   * 
   * This property is used for error reporting and troubleshooting file system issues.
   * 
   * @type {string | undefined}
   */
  filePath?: string;
  
  /**
   * Creates a new FileSystemError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.filePath - Path to the file or directory that caused the error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    filePath?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.FILESYSTEM,
    });
    
    this.name = 'FileSystemError';
    this.filePath = options?.filePath;
  }
}

/**
 * Input validation error class.
 * 
 * This error class is used for issues related to validation of user inputs,
 * such as invalid formats, missing required fields, or incorrect values.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new ValidationError('Invalid parameter format');
 * 
 * // With suggestions
 * throw new ValidationError('Prompt exceeds maximum allowed length', {
 *   suggestions: [
 *     'Limit your prompt to 4000 characters',
 *     'Break up long prompts into multiple requests'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class ValidationError extends ThinktankError {
  /**
   * Creates a new ValidationError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.VALIDATION,
    });
    this.name = 'ValidationError';
  }
}

/**
 * Network-related error class.
 * 
 * This error class is used for issues related to network connectivity,
 * such as timeouts, connection failures, or DNS resolution problems.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new NetworkError('Connection timeout');
 * 
 * // With cause and suggestions
 * try {
 *   // Network operation
 * } catch (error) {
 *   throw new NetworkError('Failed to connect to API endpoint', {
 *     cause: error,
 *     suggestions: [
 *       'Check your internet connection',
 *       'Verify the API endpoint is correct and accessible',
 *       'Try again later as the service might be temporarily down'
 *     ]
 *   });
 * }
 * ```
 * 
 * @extends {ThinktankError}
 */
export class NetworkError extends ThinktankError {
  /**
   * Creates a new NetworkError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.NETWORK,
    });
    this.name = 'NetworkError';
  }
}

/**
 * Permission-related error class.
 * 
 * This error class is used for issues related to permissions,
 * such as insufficient file system permissions, API access restrictions,
 * or unauthorized operations.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new PermissionError('Permission denied');
 * 
 * // With suggestions
 * throw new PermissionError('Permission denied when writing to output directory', {
 *   suggestions: [
 *     'Check file system permissions for the output directory',
 *     'Run the command with appropriate privileges',
 *     'Select a different output directory that you have write access to'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class PermissionError extends ThinktankError {
  /**
   * Creates a new PermissionError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.PERMISSION,
    });
    this.name = 'PermissionError';
  }
}

/**
 * Input processing error class.
 * 
 * This error class is used for issues related to processing user inputs,
 * such as invalid prompt formats, unsupported file types, or content issues.
 * Unlike ValidationError which focuses on format/schema validation, this error
 * is related to semantic issues or processing failures.
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw new InputError('Failed to process input file');
 * 
 * // With suggestions
 * throw new InputError('Input file contains unsupported markdown syntax', {
 *   suggestions: [
 *     'Use only basic markdown syntax',
 *     'Remove complex tables or diagrams',
 *     'Check for syntax errors in your markdown'
 *   ],
 *   examples: [
 *     '# Heading\n\nSimple paragraph with **bold** and *italic* text'
 *   ]
 * });
 * ```
 * 
 * @extends {ThinktankError}
 */
export class InputError extends ThinktankError {
  /**
   * Creates a new InputError instance.
   * 
   * @param message - The error message describing what went wrong
   * @param options - Optional configuration for the error
   * @param options.cause - The original error that caused this error
   * @param options.suggestions - List of suggestions to help resolve the error
   * @param options.examples - Examples of valid usage or configuration
   */
  constructor(message: string, options?: {
    cause?: Error;
    suggestions?: string[];
    examples?: string[];
  }) {
    super(message, {
      ...options,
      category: errorCategories.INPUT,
    });
    this.name = 'InputError';
  }
}

/**
 * Creates a file not found error with helpful suggestions and examples.
 * 
 * This factory function generates a FileSystemError with context-aware suggestions
 * based on the provided file path. It includes guidance on checking file existence,
 * path formatting, file extensions, and permissions.
 * 
 * @param filePath - The path to the file that wasn't found
 * @param errorMessage - Optional custom error message (defaults to "File not found: {filePath}")
 * @returns A FileSystemError with relevant suggestions and examples
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw createFileNotFoundError('/path/to/config.json');
 * 
 * // With custom error message
 * throw createFileNotFoundError('./prompt.txt', 'Unable to locate prompt file');
 * ```
 */
export function createFileNotFoundError(filePath: string, errorMessage?: string): FileSystemError {
  const message = errorMessage || `File not found: ${filePath}`;
  const currentDir = process.cwd();
  
  // Determine if path is absolute
  const isAbsolutePath = filePath.startsWith('/');
  
  // Get dirname and basename
  const lastSlashIndex = filePath.lastIndexOf('/');
  const dirname = lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '.';
  const basename = lastSlashIndex > 0 ? filePath.substring(lastSlashIndex + 1) : filePath;
  
  // Build suggestions
  const suggestions = [
    `Check that the file exists at the specified path: ${isAbsolutePath ? filePath : `${currentDir}/${filePath}`}`,
    `Current working directory: ${currentDir}`
  ];
  
  // Add path-specific suggestions
  if (!isAbsolutePath && dirname !== '.') {
    suggestions.push(`Ensure the directory exists: ${currentDir}/${dirname}`);
  }
  
  // Add common filename pattern suggestions
  if (!basename.includes('.')) {
    suggestions.push(`The file may need an extension (e.g., ${basename}.txt, ${basename}.md)`);
  }
  
  // Add general suggestions
  suggestions.push(
    `Use a relative path (./path/to/file.txt) or absolute path (/full/path/to/file.txt)`,
    `Make sure the file has read permissions`
  );
  
  // Add examples
  const examples = [
    `path/to/${basename}.txt`,
    `./path/to/${basename}.txt`,
    `${currentDir}/${basename}.txt`
  ];
  
  return new FileSystemError(message, {
    filePath,
    suggestions,
    examples
  });
}

/**
 * Creates an error for invalid model format with context-aware suggestions.
 * 
 * This factory function generates a ConfigError for issues with model specification
 * format. It analyzes the given model string and generates appropriate suggestions
 * based on the specific issue detected (missing colon, missing provider, etc.).
 * 
 * When provided with available providers and models, it enhances the suggestions
 * with specific examples from the available options.
 * 
 * @param modelSpecification - The invalid model specification (e.g., "openai-gpt4" instead of "openai:gpt-4")
 * @param availableProviders - Optional array of available provider IDs (e.g., ["openai", "anthropic"])
 * @param availableModels - Optional array of available model specifications (e.g., ["openai:gpt-4o", "anthropic:claude-3"])
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with helpful suggestions based on the specific format issue
 * 
 * @example
 * ```typescript
 * // Basic usage
 * throw createModelFormatError('openai-gpt4');
 * 
 * // With available providers and models
 * throw createModelFormatError(
 *   'openai-gpt4',
 *   ['openai', 'anthropic', 'google'],
 *   ['openai:gpt-4o', 'openai:gpt-3.5-turbo', 'anthropic:claude-3-opus']
 * );
 * ```
 */
export function createModelFormatError(
  modelSpecification: string,
  availableProviders: string[] = [],
  availableModels: string[] = [],
  errorMessage?: string
): ConfigError {
  // Create the error with a default message if none provided
  let message = errorMessage;
  
  if (!message) {
    if (!modelSpecification.includes(':')) {
      message = `Invalid model format: "${modelSpecification}". Model must be specified as "provider:modelId".`;
    } else if (modelSpecification.endsWith(':')) {
      message = `Invalid model format: "${modelSpecification}". Missing model ID after provider.`;
    } else if (modelSpecification.startsWith(':')) {
      message = `Invalid model format: "${modelSpecification}". Missing provider name before model ID.`;
    } else {
      message = `Model not found: "${modelSpecification}". Use "provider:modelId" format.`;
    }
  }
  
  // Parse the model specification
  const [provider] = modelSpecification.split(':');
  
  // Build suggestions
  const suggestions: string[] = [
    `Model specifications must use the format "provider:modelId" (e.g., "openai:gpt-4o")`
  ];
  
  // Specific error cases
  if (!modelSpecification.includes(':')) {
    suggestions.push(`Add a colon between provider and model ID: "${modelSpecification}" → "provider:${modelSpecification}"`);
  } else if (modelSpecification.endsWith(':')) {
    suggestions.push(`Specify a model ID after the provider: "${modelSpecification}modelId"`);
    
    // If we have models from this provider, suggest some
    if (availableModels.length > 0) {
      const matchingModels = availableModels.filter(m => m.startsWith(`${provider}:`));
      if (matchingModels.length > 0) {
        const models = matchingModels.slice(0, 3).join(', ') + 
          (matchingModels.length > 3 ? ', ...' : '');
        suggestions.push(`Available models for ${provider}: ${models}`);
      }
    }
  } else if (modelSpecification.startsWith(':')) {
    suggestions.push(`Specify a provider before the model ID: "provider${modelSpecification}"`);
    
    // If we have providers, suggest some
    if (availableProviders.length > 0) {
      const providersList = availableProviders.slice(0, 5).join(', ') + 
        (availableProviders.length > 5 ? ', ...' : '');
      suggestions.push(`Available providers: ${providersList}`);
    }
  }
  
  // Add general provider/model suggestions
  if (availableProviders.length > 0) {
    suggestions.push(`Available providers: ${availableProviders.join(', ')}`);
  }
  
  if (availableModels.length > 0) {
    // Limit to a reasonable number of examples
    const modelExamples = availableModels.slice(0, 5);
    const exampleList = modelExamples.join(', ') + 
      (availableModels.length > 5 ? ', ...' : '');
    suggestions.push(`Example models: ${exampleList}`);
  }
  
  // Add examples
  const examples = [
    'openai:gpt-4o',
    'anthropic:claude-3-7-sonnet-20250219',
    'google:gemini-pro'
  ];
  
  return new ConfigError(message, { suggestions, examples });
}

/**
 * Creates an error for missing API keys with provider-specific guidance.
 * 
 * This factory function generates an ApiError for missing API keys, with
 * customized suggestions based on the specific providers that need API keys.
 * It includes provider-specific links to obtain API keys and instructions
 * for setting environment variables.
 * 
 * @param missingModels - Array of models with missing API keys (each with provider and modelId)
 * @param errorMessage - Optional custom error message
 * @returns An ApiError with provider-specific guidance on obtaining and setting API keys
 * 
 * @example
 * ```typescript
 * // Single missing API key
 * throw createMissingApiKeyError([{ provider: 'openai', modelId: 'gpt-4o' }]);
 * 
 * // Multiple missing API keys
 * throw createMissingApiKeyError([
 *   { provider: 'openai', modelId: 'gpt-4o' },
 *   { provider: 'anthropic', modelId: 'claude-3-opus' }
 * ]);
 * ```
 */
export function createMissingApiKeyError(
  missingModels: Array<{ provider: string; modelId: string }>,
  errorMessage?: string
): ApiError {
  // Create the error with a default message if none provided
  const message = errorMessage || 
    `Missing API key${missingModels.length > 1 ? 's' : ''} for ${missingModels.length} model${missingModels.length > 1 ? 's' : ''}`;
  
  // Group models by provider for better suggestions
  const providerModels: Record<string, string[]> = {};
  missingModels.forEach(model => {
    if (!providerModels[model.provider]) {
      providerModels[model.provider] = [];
    }
    providerModels[model.provider].push(`${model.provider}:${model.modelId}`);
  });
  
  // Build suggestions for each provider
  const suggestions: string[] = [];
  
  // Add suggestions for each provider
  Object.entries(providerModels).forEach(([provider, models]) => {
    const modelsText = models.join(', ');
    suggestions.push(`Missing API key for ${provider} model${models.length > 1 ? 's' : ''}: ${modelsText}`);
    
    const envVarName = `${provider.toUpperCase()}_API_KEY`;
    
    // Provider-specific instructions
    switch (provider.toLowerCase()) {
      case 'openai':
        suggestions.push(
          `To use OpenAI models, get your API key from: https://platform.openai.com/api-keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'anthropic':
        suggestions.push(
          `To use Anthropic Claude models, get your API key from: https://console.anthropic.com/keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'google':
        suggestions.push(
          `To use Google AI models, get your API key from: https://aistudio.google.com/app/apikey`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      case 'openrouter':
        suggestions.push(
          `To use OpenRouter models, get your API key from: https://openrouter.ai/keys`,
          `Set the ${envVarName} environment variable with your key`
        );
        break;
        
      default:
        suggestions.push(
          `Get an API key for ${provider} from their developer portal`,
          `Set the ${envVarName} environment variable with your key`
        );
    }
  });
  
  // Add general environment variable setup instructions
  suggestions.push(
    `\nTo set environment variables:`,
    '• For Bash/Zsh: Add `export PROVIDER_API_KEY=your_key_here` to your ~/.bashrc or ~/.zshrc',
    '• For Windows Command Prompt: Use `set PROVIDER_API_KEY=your_key_here`',
    '• For PowerShell: Use `$env:PROVIDER_API_KEY = "your_key_here"`',
    '• For a local project: Create a .env file with `PROVIDER_API_KEY=your_key_here`'
  );
  
  // Add example commands
  const examples = Object.keys(providerModels).map(provider => {
    const envVarName = `${provider.toUpperCase()}_API_KEY`;
    return `export ${envVarName}=your_${provider}_key_here`;
  });
  
  return new ApiError(message, { suggestions, examples });
}

/**
 * Creates an error for model not found in configuration with specific suggestions.
 * 
 * This factory function generates a ConfigError when a requested model cannot
 * be found in the configuration. It provides context-specific suggestions based
 * on whether the issue is related to a model group or general configuration.
 * 
 * When provided with available models, it enhances the suggestions with similar
 * models or available providers to help users correct their model selection.
 * 
 * @param modelSpecification - The model specification that wasn't found (e.g., "openai:gpt-5")
 * @param availableModels - Optional array of available model specifications to suggest alternatives
 * @param groupName - Optional group name if relevant to the context (e.g., when model is missing from a specific group)
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with context-aware suggestions for resolving the issue
 * 
 * @example
 * ```typescript
 * // Basic usage - model not found
 * throw createModelNotFoundError('openai:nonexistent-model');
 * 
 * // When a model is not found in a specific group
 * throw createModelNotFoundError(
 *   'openai:gpt-4o', 
 *   ['openai:gpt-3.5-turbo', 'anthropic:claude-3-haiku'],
 *   'fast-models'
 * );
 * ```
 */
export function createModelNotFoundError(
  modelSpecification: string,
  availableModels: string[] = [],
  groupName?: string,
  errorMessage?: string
): ConfigError {
  const [provider, modelId] = modelSpecification.split(':');
  
  // Create the error with a default message if none provided
  let message = errorMessage;
  
  if (!message) {
    if (groupName) {
      message = `Model "${modelSpecification}" not found in group "${groupName}".`;
    } else {
      message = `Model "${modelSpecification}" not found in configuration.`;
    }
  }
  
  // Build suggestions
  const suggestions: string[] = [];
  
  // Suggest similar models by partial matching
  if (availableModels.length > 0) {
    // Find models with the same provider
    const sameProviderModels = availableModels.filter(m => m.startsWith(`${provider}:`));
    
    if (sameProviderModels.length > 0) {
      const providerModelList = sameProviderModels.slice(0, 5).join(', ') + 
        (sameProviderModels.length > 5 ? ', ...' : '');
      suggestions.push(`Available models from ${provider}: ${providerModelList}`);
    } else {
      // Provider not found
      suggestions.push(`Provider "${provider}" not found.`);
      
      // Find all available providers
      const availableProviders = new Set<string>();
      availableModels.forEach(m => {
        const parts = m.split(':');
        if (parts.length === 2) {
          availableProviders.add(parts[0]);
        }
      });
      
      if (availableProviders.size > 0) {
        suggestions.push(`Available providers: ${Array.from(availableProviders).join(', ')}`);
      }
    }
    
    // For specific model ID matching
    if (modelId) {
      // Find models with similar model IDs
      const similarModelIds = availableModels.filter(m => {
        const parts = m.split(':');
        return parts.length === 2 && parts[1].includes(modelId);
      });
      
      if (similarModelIds.length > 0) {
        const similarList = similarModelIds.slice(0, 3).join(', ') + 
          (similarModelIds.length > 3 ? ', ...' : '');
        suggestions.push(`Models with similar IDs: ${similarList}`);
      }
    }
    
    // Add a list of example models regardless
    const exampleList = availableModels.slice(0, 5).join(', ') + 
      (availableModels.length > 5 ? ', ...' : '');
    suggestions.push(`Available models: ${exampleList}`);
  }
  
  // Add configuration suggestions
  suggestions.push(
    'Check your thinktank.config.json file to ensure the model is properly defined',
    'Models must be enabled in the configuration to be usable'
  );
  
  if (groupName) {
    suggestions.push(`Ensure the model is included in the "${groupName}" group configuration`);
  }
  
  // Add examples
  const examples = availableModels.length > 0 
    ? availableModels.slice(0, 3) 
    : [
        'openai:gpt-4o',
        'anthropic:claude-3-7-sonnet-20250219',
        'google:gemini-pro'
      ];
  
  return new ConfigError(message, { suggestions, examples });
}