#!/usr/bin/env node
/**
 * Main entry point for the thinktank CLI
 * 
 * Sets up Commander.js and loads all available commands
 */
import { Command } from 'commander';
import fs from 'fs/promises';
import path from 'path';
import dotenv from 'dotenv';
import { 
  ThinktankError, 
  ApiError, 
  ConfigError, 
  FileSystemError, 
  NetworkError,
  errorCategories 
} from '../core/errors';
import { colors } from '../utils/consoleUtils';
import { configureLogger, logger } from '../utils/logger';

// Load environment variables from .env file
dotenv.config();

// Create the program
const program = new Command();

// Configure the CLI
program
  .name('thinktank')
  .description('A CLI tool for querying multiple LLMs with the same prompt')
  .version('0.1.0') // This should be loaded from package.json in the future
  .option('-v, --verbose', 'Enable verbose output with detailed information')
  .option('-q, --quiet', 'Suppress all output except errors')
  .option('-d, --debug', 'Enable debug mode with extra information')
  .option('--no-color', 'Disable colored output')
  .hook('preAction', (thisCommand) => {
    // Get options and safely cast them to the expected types
    const options = thisCommand.opts();
    
    // Configure the logger based on command-line options
    configureLogger({
      verbose: Boolean(options.verbose),
      quiet: Boolean(options.quiet),
      debug: Boolean(options.debug),
      noColor: !options.color
    });
    
    // Log debug info about the environment
    logger.debug(`Node.js version: ${process.version}`);
    logger.debug(`Platform: ${process.platform}`);
    logger.debug(`Command: ${process.argv.join(' ')}`);
  });

/**
 * Handle errors consistently across all commands
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
  
  process.exit(1);
}

/**
 * Adds category-specific guidance based on error type
 * 
 * @param error - The ThinktankError to provide guidance for
 */
function addCategorySpecificGuidance(error: ThinktankError): void {
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
    logger.error(`  ${colors.green('>')} Edit ~/.thinktank/config.json directly`);
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
 * 
 * @param error - The standard Error to wrap
 * @returns A ThinktankError with appropriate category and cause
 */
function wrapStandardError(error: Error): ThinktankError {
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

// Main execution function
async function main(): Promise<void> {
  try {
    // Create a directory for command modules
    const commandsDir = path.join(__dirname, 'commands');
    
    // Check if the commands directory exists
    try {
      await fs.mkdir(commandsDir, { recursive: true });
    } catch (error) {
      // Directory may already exist, which is fine
    }
    
    // Import and register the built-in commands
    // For now, we'll do this manually, but later we could automate it
    // by reading the command directory
    
    // Import run command
    const { default: runCommand } = await import('./commands/run');
    program.addCommand(runCommand);
    
    // Import models command
    const { default: modelsCommand } = await import('./commands/models');
    program.addCommand(modelsCommand);
    
    // Import config command
    const { default: configCommand } = await import('./commands/config');
    program.addCommand(configCommand);
    
    // Parse command-line arguments
    await program.parseAsync(process.argv);
  } catch (error) {
    handleError(error);
  }
}

// Execute the main function
main().catch(handleError);