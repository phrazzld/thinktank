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

// Load environment variables from .env file
dotenv.config();

// Create the program
const program = new Command();

// Configure the CLI
program
  .name('thinktank')
  .description('A CLI tool for querying multiple LLMs with the same prompt')
  .version('0.1.0'); // This should be loaded from package.json in the future

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
    // eslint-disable-next-line no-console
    console.error(`${colors.red('Error')}${colors.yellow(category)}: ${error.message}`);
    
    // Display cause if available
    if (error.cause) {
      // eslint-disable-next-line no-console
      console.error(`${colors.dim('Cause:')} ${error.cause.message}`);
    }
    
    // Show suggestions if available
    if (error.suggestions && error.suggestions.length > 0) {
      // eslint-disable-next-line no-console
      console.error('\nSuggestions:');
      error.suggestions.forEach(suggestion => {
        // eslint-disable-next-line no-console
        console.error(`  ${colors.cyan('•')} ${suggestion}`);
      });
    }
    
    // Show examples if available
    if (error.examples && error.examples.length > 0) {
      // eslint-disable-next-line no-console
      console.error('\nExample commands:');
      error.examples.forEach(example => {
        // eslint-disable-next-line no-console
        console.error(`  ${colors.green('>')} ${example}`);
      });
    }
    
    // Show general help for common errors
    if (error.category === errorCategories.FILESYSTEM) {
      // eslint-disable-next-line no-console
      console.error('\nCorrect usage:');
      // eslint-disable-next-line no-console
      console.error(`  ${colors.green('>')} thinktank run prompt.txt [--group=group]`);
      // eslint-disable-next-line no-console
      console.error(`  ${colors.green('>')} thinktank run prompt.txt --models=provider:model`);
    }
  } else if (error instanceof Error) {
    // eslint-disable-next-line no-console
    console.error(`${colors.red('Unexpected error:')} ${error.message}`);
  } else {
    // eslint-disable-next-line no-console
    console.error(`${colors.red('An unknown error occurred')}`);
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