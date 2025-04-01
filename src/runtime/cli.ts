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
import { colors, errorCategories, createFileNotFoundError } from '../atoms/consoleUtils';
import fs from 'fs/promises';
import dotenv from 'dotenv';
import path from 'path';

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
    
    // Handle "models" command (list all available models)
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
      // Create a helpful file not found error using our utility
      const errorMessage = `Input file not found: ${promptFile}`;
      const baseError = createFileNotFoundError(promptFile, errorMessage);
      
      // Convert to ThinktankError for consistency
      const fileError = new ThinktankError(baseError.message);
      
      // Copy properties from baseError
      fileError.category = (baseError as any).category;
      fileError.suggestions = (baseError as any).suggestions;
      
      // Customize examples for CLI usage
      const basename = path.basename(promptFile);
      fileError.examples = [
        `thinktank ${basename}.txt`,
        `thinktank ${basename}.txt default`,
        `thinktank ${basename}.txt openai:gpt-4o`
      ];
      
      throw fileError;
    }
    
    // The second argument can be either a group name or a specific model
    const secondArg = args[1];
    
    if (secondArg) {
      // Check if the second argument contains a colon, indicating a provider:model format
      if (secondArg.includes(':')) {
        // Running with a specific model
        const [provider, modelId] = secondArg.split(':');
        
        // Validate provider and modelId are present
        if (!provider || !modelId) {
          throw new ThinktankError(
            `Invalid model format: "${secondArg}". Use "provider:modelId" format (e.g., "openai:gpt-4o").`
          );
        }
        
        // Call runThinktank with the specific model
        await runThinktank({
          input: promptFile,
          specificModel: secondArg
        });
      } else {
        // Running with a group name
        await runThinktank({
          input: promptFile,
          groupName: secondArg
        });
      }
    } else {
      // No second argument, use default group
      await runThinktank({
        input: promptFile,
        // Leave groupName undefined to use default group
      });
    }
    
    process.exit(0);
  } catch (error) {
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
        console.error(`  ${colors.green('>')} thinktank prompt.txt [group]`);
        // eslint-disable-next-line no-console
        console.error(`  ${colors.green('>')} thinktank prompt.txt provider:model`);
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
}

// Execute main function
main().catch(error => {
  // eslint-disable-next-line no-console
  console.error('Fatal error:', error);
  process.exit(1);
});