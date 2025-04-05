/**
 * Centralized error handling system for thinktank
 * 
 * This module defines a consistent error hierarchy and error handling utilities
 * to improve error reporting and troubleshooting across the application.
 */

import { colors } from '../utils/consoleUtils';

/**
 * Error categories for consistent categorization across the application
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
 * Base error class for all thinktank errors
 * Provides a consistent structure for error reporting
 */
export class ThinktankError extends Error {
  /**
   * Error category for grouping similar errors
   */
  category: string = errorCategories.UNKNOWN;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands or configurations
   */
  examples?: string[];

  /**
   * The original error that caused this error
   */
  cause?: Error;
  
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
   * Formats the error for display
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
 * Configuration-related errors
 */
export class ConfigError extends ThinktankError {
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
 * API interaction errors
 */
export class ApiError extends ThinktankError {
  providerId?: string;
  
  constructor(message: string, options?: {
    cause?: Error;
    providerId?: string;
    suggestions?: string[];
    examples?: string[];
  }) {
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
 * File system related errors
 */
export class FileSystemError extends ThinktankError {
  filePath?: string;
  
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
 * Input validation errors
 */
export class ValidationError extends ThinktankError {
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
 * Network-related errors
 */
export class NetworkError extends ThinktankError {
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
 * Permission-related errors
 */
export class PermissionError extends ThinktankError {
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
 * Input processing errors
 */
export class InputError extends ThinktankError {
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
 * Creates a file not found error with helpful suggestions
 * 
 * @param filePath - The path to the file that wasn't found
 * @param errorMessage - Optional custom error message
 * @returns A FileSystemError with relevant suggestions
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
 * Creates an error for invalid model format
 * 
 * @param modelSpecification - The invalid model specification
 * @param availableProviders - Optional array of available provider IDs
 * @param availableModels - Optional array of available model specifications
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with helpful suggestions
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
 * Creates an error for missing API keys
 * 
 * @param missingModels - Array of models with missing API keys
 * @param errorMessage - Optional custom error message
 * @returns An ApiError with helpful suggestions
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
 * Creates an error for model not found in configuration
 * 
 * @param modelSpecification - The model specification that wasn't found
 * @param availableModels - Optional array of available model specifications
 * @param groupName - Optional group name if relevant to the context
 * @param errorMessage - Optional custom error message
 * @returns A ConfigError with helpful suggestions
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