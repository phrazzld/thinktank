/**
 * Console utility module for terminal styling and formatting
 * 
 * Centralizes all terminal styling logic to maintain consistency
 * and provide reusable formatting helpers across the application.
 */

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-return */
/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
import chalk from 'chalk';

// Re-export our configured chalk instance
export const colors = chalk;

// Define commonly used Unicode symbols
export const symbols = {
  tick: '✓',
  cross: '✖',
  warning: '⚠',
  info: 'ℹ',
  pointer: '❯',
  line: '─',
  bullet: '•',
};

/**
 * Styles text as a success message
 * @param text - The text to style
 * @returns Styled text with a success indicator
 */
export function styleSuccess(text: string): string {
  return `${colors.green(symbols.tick)} ${text}`;
}

/**
 * Styles text as an error message
 * @param text - The text to style
 * @returns Styled text with an error indicator
 */
export function styleError(text: string): string {
  return `${colors.red(symbols.cross)} ${text}`;
}

/**
 * Styles text as a warning message
 * @param text - The text to style
 * @returns Styled text with a warning indicator
 */
export function styleWarning(text: string): string {
  return `${colors.yellow(symbols.warning)} ${text}`;
}

/**
 * Styles text as an informational message
 * @param text - The text to style
 * @returns Styled text with an info indicator
 */
export function styleInfo(text: string): string {
  return `${colors.blue(symbols.info)} ${text}`;
}

/**
 * Styles text as a header
 * @param text - The text to style
 * @returns Styled text as a header
 */
export function styleHeader(text: string): string {
  return colors.bold.blue(text);
}

/**
 * Styles text as dimmed/secondary content
 * @param text - The text to style
 * @returns Styled dimmed text
 */
export function styleDim(text: string): string {
  return colors.dim(text);
}

/**
 * A divider line for visual separation
 * @param length - The length of the divider line
 * @returns A styled divider line
 */
export function divider(length = 80): string {
  return styleDim('─'.repeat(length));
}

/**
 * Categories of errors for consistent categorization
 */
export const errorCategories = {
  API: 'API',
  CONFIG: 'Configuration',
  NETWORK: 'Network',
  FILESYSTEM: 'File System',
  PERMISSION: 'Permission',
  VALIDATION: 'Validation',
  UNKNOWN: 'Unknown',
};

/**
 * Formats an error message with consistent styling
 * @param error - The error object or message
 * @param category - Optional error category
 * @param tip - Optional troubleshooting tip
 * @returns Formatted error message
 */
export function formatError(
  error: Error | string, 
  category: string = errorCategories.UNKNOWN,
  tip?: string
): string {
  const errorMsg = error instanceof Error ? error.message : error;
  let output = `${colors.red.bold('Error')}${category ? ` (${colors.yellow(category)})` : ''}: ${errorMsg}`;
  
  if (tip) {
    output += `\n  ${colors.cyan(symbols.info)} Tip: ${tip}`;
  }
  
  return output;
}

/**
 * Tries to categorize an error based on its message or type
 * @param error - The error to categorize
 * @returns The error category
 */
export function categorizeError(error: Error | string): string {
  const message = error instanceof Error ? error.message : error;
  const lowerMsg = message.toLowerCase();
  
  if (lowerMsg.includes('api key') || lowerMsg.includes('authentication') || 
      lowerMsg.includes('auth') || lowerMsg.includes('401') || lowerMsg.includes('403')) {
    return errorCategories.API;
  }
  
  if (lowerMsg.includes('econnrefused') || lowerMsg.includes('etimedout') || 
      lowerMsg.includes('enotfound') || lowerMsg.includes('network')) {
    return errorCategories.NETWORK;
  }
  
  if (lowerMsg.includes('config') || lowerMsg.includes('settings')) {
    return errorCategories.CONFIG;
  }
  
  if (lowerMsg.includes('enoent') || lowerMsg.includes('file not found') || 
      lowerMsg.includes('directory') || lowerMsg.includes('path')) {
    return errorCategories.FILESYSTEM;
  }
  
  if (lowerMsg.includes('permission') || lowerMsg.includes('access denied') ||
      lowerMsg.includes('eacces')) {
    return errorCategories.PERMISSION;
  }
  
  if (lowerMsg.includes('validation') || lowerMsg.includes('invalid') || 
      lowerMsg.includes('schema') || lowerMsg.includes('required')) {
    return errorCategories.VALIDATION;
  }
  
  return errorCategories.UNKNOWN;
}

/**
 * Gets a troubleshooting tip based on the error category
 * @param error - The error message or object
 * @param category - The error category
 * @returns A helpful tip or undefined if none available
 */
export function getTroubleshootingTip(error: Error | string, category: string): string | undefined {
  const message = error instanceof Error ? error.message : error;
  const lowerMsg = message.toLowerCase();
  
  switch(category) {
    case errorCategories.API:
      if (lowerMsg.includes('api key')) {
        return 'Check your API key in your environment variables or config file';
      }
      if (lowerMsg.includes('rate limit') || lowerMsg.includes('429')) {
        return 'You\'ve hit the rate limit. Wait a while before trying again';
      }
      return 'Verify your API credentials and permissions';
      
    case errorCategories.NETWORK:
      return 'Check your internet connection and try again';
      
    case errorCategories.CONFIG:
      return 'Verify your thinktank.config.json file for errors';
      
    case errorCategories.FILESYSTEM:
      if (lowerMsg.includes('permission') || lowerMsg.includes('eacces')) {
        return 'Check file permissions or run with elevated privileges';
      }
      if (lowerMsg.includes('file not found') || lowerMsg.includes('enoent') || lowerMsg.includes('no such file')) {
        return 'Check that the file exists at the specified path and you have permission to read it';
      }
      return 'Verify the file path and ensure it exists';
      
    default:
      return undefined;
  }
}

/**
 * Formats an error with automatically determined category and tip
 * @param error - The error to format
 * @returns A formatted error message with category and tip
 */
export function formatErrorWithTip(error: Error | string): string {
  const category = categorizeError(error);
  const tip = getTroubleshootingTip(error, category);
  return formatError(error, category, tip);
}

/**
 * Creates a detailed file not found error with helpful suggestions
 * 
 * @param filePath - The path to the file that wasn't found
 * @param errorMessage - Optional custom error message
 * @returns A ThinktankError with suggestions
 */
export function createFileNotFoundError(filePath: string, errorMessage?: string): Error {
  // Import path if available, otherwise use a simpler approach
  let path: any;
  try {
    path = require('path');
  } catch {
    // No path module, use simpler approach
    path = {
      isAbsolute: (p: string) => p.startsWith('/'),
      dirname: (p: string) => {
        const parts = p.split('/');
        return parts.slice(0, -1).join('/') || '.';
      },
      basename: (p: string) => {
        const parts = p.split('/');
        return parts[parts.length - 1];
      },
      join: (...parts: string[]) => parts.join('/')
    };
  }

  // Get current working directory
  const currentDir = process.cwd();

  // Create the error with a default message if none provided
  const message = errorMessage || `File not found: ${filePath}`;
  
  // This requires the ThinktankError class which we can't import here
  // to avoid circular dependencies, so we'll create a regular Error
  // and add the properties the consumer can convert it if needed
  const error = new Error(message);
  
  // Add metadata
  (error as any).category = errorCategories.FILESYSTEM;
  
  // Extract path components
  const isAbsolutePath = path.isAbsolute(filePath);
  const dirname = path.dirname(filePath);
  const basename = path.basename(filePath);
  
  // Build suggestions
  const suggestions = [
    `Check that the file exists at the specified path: ${isAbsolutePath ? filePath : path.join(currentDir, filePath)}`,
    `Current working directory: ${currentDir}`
  ];
  
  // Add path-specific suggestions
  if (!isAbsolutePath && dirname !== '.') {
    suggestions.push(`Ensure the directory exists: ${path.join(currentDir, dirname)}`);
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
  
  (error as any).suggestions = suggestions;
  
  // Add examples
  (error as any).examples = [
    `path/to/${basename}.txt`,
    `./path/to/${basename}.txt`,
    `${path.join(currentDir, basename)}.txt`
  ];
  
  return error;
}

/**
 * Handles common types of model format errors
 * 
 * @param modelSpecification - The problematic model specification
 * @param availableProviders - Optional array of available provider IDs
 * @param availableModels - Optional array of available model specifications
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions
 */
export function createModelFormatError(
  modelSpecification: string,
  availableProviders: string[] = [],
  availableModels: string[] = [],
  errorMessage?: string
): Error {
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
  
  const error = new Error(message);
  
  // Add metadata
  (error as any).category = errorCategories.CONFIG;
  
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
  
  (error as any).suggestions = suggestions;
  
  // Add examples
  (error as any).examples = [
    'openai:gpt-4o',
    'anthropic:claude-3-7-sonnet-20250219',
    'google:gemini-pro'
  ];
  
  return error;
}

/**
 * Handles errors when a specific model is not found or unavailable
 * 
 * @param modelSpecification - The model specification that wasn't found
 * @param availableModels - Optional array of available model specifications
 * @param groupName - Optional group name if relevant to the context
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions
 */
/**
 * Creates a helpful error message for missing API keys
 * 
 * @param missingModels - Array of models with missing API keys
 * @param errorMessage - Optional custom error message
 * @returns An Error object with helpful suggestions for setting API keys
 */
export function createMissingApiKeyError(
  missingModels: Array<{ provider: string; modelId: string }>,
  errorMessage?: string
): Error {
  // Create the error with a default message if none provided
  const message = errorMessage || 
    `Missing API key${missingModels.length > 1 ? 's' : ''} for ${missingModels.length} model${missingModels.length > 1 ? 's' : ''}`;
  
  const error = new Error(message);
  
  // Add metadata
  (error as any).category = errorCategories.API;
  
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
  
  (error as any).suggestions = suggestions;
  
  // Add example commands
  const examples = Object.keys(providerModels).map(provider => {
    const envVarName = `${provider.toUpperCase()}_API_KEY`;
    return `export ${envVarName}=your_${provider}_key_here`;
  });
  
  (error as any).examples = examples;
  
  return error;
}

export function createModelNotFoundError(
  modelSpecification: string,
  availableModels: string[] = [],
  groupName?: string,
  errorMessage?: string
): Error {
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
  
  const error = new Error(message);
  
  // Add metadata
  (error as any).category = errorCategories.CONFIG;
  
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
  
  (error as any).suggestions = suggestions;
  
  // Add examples
  (error as any).examples = availableModels.length > 0 
    ? availableModels.slice(0, 3) 
    : [
        'openai:gpt-4o',
        'anthropic:claude-3-7-sonnet-20250219',
        'google:gemini-pro'
      ];
  
  return error;
}