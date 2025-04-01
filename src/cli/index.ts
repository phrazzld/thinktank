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
import { ThinktankError } from '../workflow/runThinktank';
import { colors, errorCategories } from '../utils/consoleUtils';
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
    // Display the main error message with appropriate category if available
    const category = error.category ? ` (${error.category})` : '';
    const message = `${colors.red('Error')}${colors.yellow(category)}: ${error.message}`;
    
    // Use the logger but directly with console.error to ensure it's shown
    logger.error(message);
    
    // Display cause if available
    if (error.cause) {
      logger.error(`${colors.dim('Cause:')} ${error.cause.message}`);
    }
    
    // Show suggestions if available
    if (error.suggestions && error.suggestions.length > 0) {
      logger.error('\nSuggestions:');
      error.suggestions.forEach(suggestion => {
        logger.error(`  ${colors.cyan('•')} ${suggestion}`);
      });
    }
    
    // Show examples if available
    if (error.examples && error.examples.length > 0) {
      logger.error('\nExample commands:');
      error.examples.forEach(example => {
        logger.error(`  ${colors.green('>')} ${example}`);
      });
    }
    
    // Show general help for common errors
    if (error.category === errorCategories.FILESYSTEM) {
      logger.error('\nCorrect usage:');
      logger.error(`  ${colors.green('>')} thinktank run prompt.txt [--group=group]`);
      logger.error(`  ${colors.green('>')} thinktank run prompt.txt --models=provider:model`);
    }
  } else if (error instanceof Error) {
    logger.error(`Unexpected error: ${error.message}`, error);
  } else {
    logger.error('An unknown error occurred');
  }
  
  process.exit(1);
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