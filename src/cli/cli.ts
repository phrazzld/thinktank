#!/usr/bin/env node
/**
 * Command-line interface for thinktank
 * 
 * Simplified CLI that supports three core use cases:
 * 1. Running a prompt through a group of models (or default group)
 * 2. Running a prompt through one specific model
 * 3. Listing all available models
 */
import { runThinktank, ThinktankError } from '../workflow/runThinktank';

// Interface for error objects with metadata
interface ErrorWithMetadata {
  message: string;
  category?: string;
  suggestions?: string[];
  examples?: string[];
}
import { listAvailableModels } from '../workflow/listModelsWorkflow';
import { 
  colors, 
  errorCategories, 
  createFileNotFoundError, 
  createModelFormatError
} from '../utils/consoleUtils';
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
  console.error('Options:');
  // eslint-disable-next-line no-console
  console.error('  --thinking                                # Enable Claude\'s thinking capability (for supported models)');
  // eslint-disable-next-line no-console
  console.error('  --show-thinking                           # Display thinking output in the results');
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
  console.error('  thinktank prompt.txt anthropic:claude-3.7-sonnet-20250219 --thinking --show-thinking');
  // eslint-disable-next-line no-console
  console.error('  thinktank models                          # List all available models');
}

// Define interfaces for the helper function return types to avoid 'any' usage
interface ConfigModel {
  provider: string;
  modelId: string;
  enabled: boolean;
}

interface ModelGroup {
  name: string;
  systemPrompt: { text: string; metadata?: Record<string, unknown> };
  models: ConfigModel[];
  description?: string;
}

interface Config {
  models: ConfigModel[];
  groups?: Record<string, ModelGroup>;
  defaultGroup?: string;
}

interface ConfigHelpers {
  loadConfig: () => Promise<Config>;
  getEnabledModels: (config: { models: Array<{ provider: string; modelId: string; enabled: boolean }> }) => Array<{
    provider: string;
    modelId: string;
  }>;
}

interface ProviderHelpers {
  getProviderIds: () => string[];
}

// Fixes for dynamic import in Jest tests - removed as unused
// If needed in future, would be implemented with proper return type

// Fixes for dynamic import in Jest tests
async function getConfigHelper(): Promise<ConfigHelpers> {
  try {
    const configModule = await import('../core/configManager');
    return configModule;
  } catch (error) {
    // Default mock for tests
    return {
      loadConfig: async () => {
        await new Promise(resolve => setTimeout(resolve, 0)); // Add await to satisfy linter
        return { models: [] };
      },
      getEnabledModels: () => []
    };
  }
}

// Fixes for dynamic import in Jest tests
async function getProviderHelper(): Promise<ProviderHelpers> {
  try {
    const providerModule = await import('../core/llmRegistry');
    return providerModule;
  } catch (error) {
    // Default mock for tests
    return {
      getProviderIds: () => ['openai', 'anthropic']
    };
  }
}

/**
 * Main CLI entry point
 * @export For testing purposes
 */
export async function main(): Promise<void> {
  try {
    // Get command-line arguments (excluding node and script path)
    const args = process.argv.slice(2);
    
    // Check for --help flag before processing other arguments
    const helpFlagIndex = args.indexOf('--help');
    if (helpFlagIndex !== -1) {
      showHelp();
      process.exit(0); // Exit successfully after showing help
    }
    
    // Parse options
    const options: {
      thinking?: boolean;
      showThinking?: boolean;
    } = {};
    
    // Check for --thinking flag (to enable thinking capability)
    const thinkingFlagIndex = args.indexOf('--thinking');
    if (thinkingFlagIndex !== -1) {
      options.thinking = true;
      args.splice(thinkingFlagIndex, 1); // Remove the flag from args
    }
    
    // Check for --show-thinking flag (to display thinking outputs)
    const showThinkingFlagIndex = args.indexOf('--show-thinking');
    if (showThinkingFlagIndex !== -1) {
      options.showThinking = true;
      args.splice(showThinkingFlagIndex, 1); // Remove the flag from args
    }
    
    // No arguments provided - show help
    if (args.length === 0) {
      showHelp();
      process.exit(1); // Error code for missing required arguments
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
      const typedError = baseError as ErrorWithMetadata;
      if (typedError.category) {
        fileError.category = typedError.category;
      }
      
      if (typedError.suggestions) {
        fileError.suggestions = typedError.suggestions;
      }
      
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
          try {
            // Get helper utils and providers 
            const { getProviderIds } = await getProviderHelper();
            const { loadConfig, getEnabledModels } = await getConfigHelper();
            
            // Fetch available providers and models for better error messages
            const availableProviders = getProviderIds();
            const config = await loadConfig();
            const enabledModels = getEnabledModels(config);
            
            // Create model strings in provider:modelId format
            const availableModels = enabledModels.map(model => `${model.provider}:${model.modelId}`);
            
            // Use the utility function directly (from import or helper)
            const modelError = createModelFormatError(
              secondArg,
              availableProviders,
              availableModels
            );
            
            // Convert to ThinktankError
            const modelFormatError = new ThinktankError(modelError.message);
            const typedModelError = modelError as ErrorWithMetadata;
            
            if (typedModelError.category) {
              modelFormatError.category = typedModelError.category;
            }
            
            if (typedModelError.suggestions) {
              modelFormatError.suggestions = typedModelError.suggestions;
            }
            
            if (typedModelError.examples) {
              modelFormatError.examples = typedModelError.examples;
            }
            
            throw modelFormatError;
          } catch (importError) {
            // Fallback to simpler error if we can't get providers/models
            throw new ThinktankError(
              `Invalid model format: "${secondArg}". Use "provider:modelId" format (e.g., "openai:gpt-4o").`
            );
          }
        }
        
        // Call runThinktank with the specific model
        await runThinktank({
          input: promptFile,
          specificModel: secondArg,
          enableThinking: options.thinking,
          includeThinking: options.showThinking
        });
      } else {
        // Running with a group name
        await runThinktank({
          input: promptFile,
          groupName: secondArg,
          enableThinking: options.thinking,
          includeThinking: options.showThinking
        });
      }
    } else {
      // No second argument, use default group
      await runThinktank({
        input: promptFile,
        // Leave groupName undefined to use default group
        enableThinking: options.thinking,
        includeThinking: options.showThinking
      });
    }
    
    // Clean up any lingering connections by explicitly exiting
    // This ensures we don't have hanging HTTP connections from axios/openai/anthropic libraries
    process.nextTick(() => {
      process.exit(0);
    });
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
    
    // Clean up any lingering connections on error
    process.nextTick(() => {
      process.exit(1);
    });
  }
}

// Execute main function
main().catch(error => {
  // eslint-disable-next-line no-console
  console.error('Fatal error:', error);
  process.nextTick(() => {
    process.exit(1);
  });
});