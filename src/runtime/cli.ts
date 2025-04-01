#!/usr/bin/env node
/**
 * Command-line interface for thinktank
 * 
 * Simplified CLI that supports three core use cases:
 * 1. Running a prompt through a group of models (or default group)
 * 2. Running a prompt through one specific model
 * 3. Listing all available models
 */
import { runThinktank, ThinktankError } from '../templates/runThinktank';
import { listAvailableModels } from '../templates/listModelsWorkflow';
import fs from 'fs/promises';
import dotenv from 'dotenv';

// Load environment variables early
dotenv.config();

/**
 * Display usage help message
 */
function showHelp(): void {
  // eslint-disable-next-line no-console
  console.error('Usage:');
  // eslint-disable-next-line no-console
  console.error('  thinktank prompt.txt [group]              # Run prompt through a group (or default group)');
  // eslint-disable-next-line no-console
  console.error('  thinktank prompt.txt provider:model       # Run prompt through one specific model');
  // eslint-disable-next-line no-console
  console.error('  thinktank models                          # List all available models');
  // eslint-disable-next-line no-console
  console.error('');
  // eslint-disable-next-line no-console
  console.error('Examples:');
  // eslint-disable-next-line no-console
  console.error('  thinktank prompt.txt                      # Run prompt through default group');
  // eslint-disable-next-line no-console
  console.error('  thinktank prompt.txt coding               # Run prompt through "coding" group');
  // eslint-disable-next-line no-console
  console.error('  thinktank prompt.txt openai:gpt-4o        # Run prompt through specific model');
  // eslint-disable-next-line no-console
  console.error('  thinktank models                          # List all available models');
}

/**
 * Main CLI entry point
 * @export For testing purposes
 */
export async function main(): Promise<void> {
  try {
    // Get command-line arguments (excluding node and script path)
    const args = process.argv.slice(2);
    
    // No arguments provided - show help
    if (args.length === 0) {
      showHelp();
      process.exit(1);
    }
    
    // Check for "models" command
    if (args[0] === 'models') {
      const result = await listAvailableModels({});
      // eslint-disable-next-line no-console
      console.log(result);
      process.exit(0);
    }
    
    // All other commands require a prompt file as first argument
    const promptFile = args[0];
    
    // Validate prompt file exists
    try {
      await fs.access(promptFile);
    } catch (error) {
      throw new ThinktankError(`Input file not found: ${promptFile}`);
    }
    
    // The second argument can be either a group name or a specific model
    // We'll implement this in the next task
    // Uncomment when we implement the parsing logic:
    // const secondArg = args[1];
    
    // We'll implement the actual parsing logic for the group vs model
    // in the next task, but set up the structure here
    
    // For now, we'll pass through the input file and handle other options
    // in future implementations
    await runThinktank({
      input: promptFile,
      // Other options to be determined based on second argument
    });
    
    process.exit(0);
  } catch (error) {
    // Handle errors
    if (error instanceof ThinktankError) {
      // eslint-disable-next-line no-console
      console.error(`Error: ${error.message}`);
      
      if (error.cause) {
        // eslint-disable-next-line no-console
        console.error(`Cause: ${error.cause.message}`);
      }
    } else if (error instanceof Error) {
      // eslint-disable-next-line no-console
      console.error(`Unexpected error: ${error.message}`);
    } else {
      // eslint-disable-next-line no-console
      console.error('An unknown error occurred');
    }
    
    process.exit(1);
  }
}

// Execute main function
main().catch(error => {
  // eslint-disable-next-line no-console
  console.error('Fatal error:', error);
  process.exit(1);
});