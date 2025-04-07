/**
 * Test helpers for CLI error handling
 * 
 * This file contains extracted copies of the error handling functions from cli/index.ts
 * for use in tests. This avoids the need to import the actual CLI module and trigger
 * its initialization code.
 */
import { 
  ThinktankError, 
  ApiError, 
  ConfigError, 
  FileSystemError, 
  NetworkError,
  errorCategories 
} from '../../../core/errors';
import { colors } from '../../../utils/consoleUtils';
import { logger } from '../../../utils/logger';

/**
 * Handle errors consistently across all commands
 * This is a copy of the handleError function from cli/index.ts
 * 
 * @param error - The error to handle
 */
export function handleError(error: unknown): void {
  // Handle errors with enhanced formatting
  if (error instanceof ThinktankError) {
    // Use the built-in format method for consistent error display
    logger.error(error.format());
    
    // Display cause if available and not already shown in format()
    if (error.cause && !error.format().includes('Cause:')) {
      logger.error(`${colors.dim('Cause:')} ${error.cause.message}`);
    }
    
    // Provide category-specific guidance
    // This adds contextual help based on error category
    addCategorySpecificGuidance(error);
    
  } else if (error instanceof Error) {
    // Convert standard Error to ThinktankError for consistent formatting
    const wrappedError = wrapStandardError(error);
    logger.error(wrappedError.format());
    
    // Add contextual help for the wrapped error
    addCategorySpecificGuidance(wrappedError);
  } else {
    // Handle unknown errors (non-Error objects)
    const genericError = new ThinktankError('An unknown error occurred', {
      category: errorCategories.UNKNOWN,
      suggestions: [
        'This is likely an internal error in thinktank',
        'Check for updates to thinktank as this may be a fixed issue',
        'Report this issue if it persists'
      ]
    });
    logger.error(genericError.format());
  }
  
  // In tests, we mock process.exit, but in the real CLI it would exit here
  // process.exit(1);
}

/**
 * Adds category-specific guidance based on error type
 * This is a copy of the addCategorySpecificGuidance function from cli/index.ts
 * 
 * @param error - The ThinktankError to provide guidance for
 */
export function addCategorySpecificGuidance(error: ThinktankError): void {
  // File System errors
  if (error.category === errorCategories.FILESYSTEM) {
    logger.error('\nCorrect usage:');
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt [--group=group]`);
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt --models=provider:model`);
  }
  
  // Configuration errors
  else if (error.category === errorCategories.CONFIG) {
    logger.error('\nConfiguration help:');
    logger.error(`  ${colors.green('>')} thinktank config view`);
    logger.error(`  ${colors.green('>')} thinktank config set key value`);
    logger.error(`  ${colors.green('>')} Edit ~/.config/thinktank/config.json directly`);
  }
  
  // API errors
  else if (error.category === errorCategories.API) {
    // For API errors, check if it's provider-specific
    if (error instanceof ApiError && error.providerId) {
      // Provider-specific guidance
      switch (error.providerId.toLowerCase()) {
        case 'openai':
          logger.error('\nOpenAI API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://platform.openai.com/api-keys`);
          logger.error(`  ${colors.green('>')} Set with: export OPENAI_API_KEY=your_key_here`);
          break;
          
        case 'anthropic':
          logger.error('\nAnthropic API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://console.anthropic.com/keys`);
          logger.error(`  ${colors.green('>')} Set with: export ANTHROPIC_API_KEY=your_key_here`);
          break;
          
        case 'google':
          logger.error('\nGoogle AI API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://aistudio.google.com/app/apikey`);
          logger.error(`  ${colors.green('>')} Set with: export GEMINI_API_KEY=your_key_here`);
          break;
          
        case 'openrouter':
          logger.error('\nOpenRouter API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://openrouter.ai/keys`);
          logger.error(`  ${colors.green('>')} Set with: export OPENROUTER_API_KEY=your_key_here`);
          break;
          
        default:
          logger.error('\nAPI help:');
          logger.error(`  ${colors.green('>')} Ensure you have the correct API key for ${error.providerId}`);
          logger.error(`  ${colors.green('>')} Set with: export ${error.providerId.toUpperCase()}_API_KEY=your_key_here`);
      }
    } else {
      // Generic API error guidance
      logger.error('\nAPI help:');
      logger.error(`  ${colors.green('>')} Check your API credentials`);
      logger.error(`  ${colors.green('>')} Verify network connectivity to API services`);
      logger.error(`  ${colors.green('>')} Run with --debug flag for more information`);
    }
  }
  
  // Network errors
  else if (error.category === errorCategories.NETWORK) {
    logger.error('\nNetwork troubleshooting:');
    logger.error(`  ${colors.green('>')} Check your internet connection`);
    logger.error(`  ${colors.green('>')} Verify you can access the API endpoints (no firewall blocking)`);
    logger.error(`  ${colors.green('>')} Try again in a few minutes if service might be down`);
  }
  
  // Validation errors (including input errors)
  else if (error.category === errorCategories.VALIDATION || error.category === errorCategories.INPUT) {
    logger.error('\nInput help:');
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt [options]`);
    logger.error(`  ${colors.green('>')} Use --help with any command for detailed usage`);
  }
  
  // For unknown or other errors, offer general debugging help
  else if (error.category === errorCategories.UNKNOWN) {
    logger.error('\nTroubleshooting help:');
    logger.error(`  ${colors.green('>')} Run with --debug flag for more information`);
    logger.error(`  ${colors.green('>')} Check thinktank documentation for guidance`);
    logger.error(`  ${colors.green('>')} Report bugs at: https://github.com/phrazzld/thinktank/issues`);
  }
}

/**
 * Wraps standard Error objects in ThinktankError for consistent formatting
 * This is a copy of the wrapStandardError function from cli/index.ts
 * 
 * @param error - The standard Error to wrap
 * @returns A ThinktankError with appropriate category and cause
 */
export function wrapStandardError(error: Error): ThinktankError {
  // Try to categorize based on message content
  const message = error.message.toLowerCase();
  
  // Network-related errors
  if (message.includes('network') || 
      message.includes('econnrefused') || 
      message.includes('timeout') ||
      message.includes('socket')) {
    return new NetworkError(`Network error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check your internet connection',
        'Verify that required services are accessible from your network',
        'The service might be down or experiencing issues'
      ]
    });
  }
  
  // File-related errors
  else if (message.includes('file') || 
           message.includes('directory') || 
           message.includes('enoent') ||
           message.includes('permission denied')) {
    return new FileSystemError(`File system error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check that the file or directory exists',
        'Verify that you have appropriate permissions',
        'Ensure the path is correct'
      ]
    });
  }
  
  // Configuration errors
  else if (message.includes('config') || 
           message.includes('settings') || 
           message.includes('option')) {
    return new ConfigError(`Configuration error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check your thinktank configuration file',
        'Try resetting to default configuration with: thinktank config reset',
        'Verify that configuration values are in the correct format'
      ]
    });
  }
  
  // Default to unknown category
  return new ThinktankError(`Unexpected error: ${error.message}`, {
    category: errorCategories.UNKNOWN,
    cause: error,
    suggestions: [
      'Run with --debug flag for more detailed information',
      'Check documentation for this feature',
      'This may be an internal error in thinktank'
    ]
  });
}
