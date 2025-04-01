/**
 * Run command implementation
 * 
 * This command handles sending prompts to LLMs
 */
import { Command } from 'commander';
import fs from 'fs/promises';
import path from 'path';
import { runThinktank, ThinktankError } from '../../workflow/runThinktank';

// Interface for error objects with metadata
interface ErrorWithMetadata {
  message: string;
  category?: string;
  suggestions?: string[];
  examples?: string[];
}
import { createFileNotFoundError, colors } from '../../utils/consoleUtils';
import { handleError } from '../index';
import * as configManager from '../../core/configManager';

// Create the command
const runCommand = new Command('run');

// Configure the command
runCommand
  .description('Run a prompt through one or more LLM models')
  .argument('<prompt-file>', 'Path to the file containing the prompt')
  .option('-g, --group <name>', 'Group name to run the prompt against')
  .option('-m, --models <models>', 'Comma-separated list of models in provider:model format (e.g., "openai:gpt-4o,anthropic:claude-3-opus-20240229")')
  .option('-t, --thinking', 'Enable thinking capability for models that support it')
  .option('-s, --show-thinking', 'Display thinking output in the results')
  .option('-o, --output <path>', 'Output directory path (defaults to thinktank-output)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-v, --verbose', 'Show verbose output including detailed model information')
  .option('--include-metadata', 'Include metadata in the output')
  .option('--system-prompt <text>', 'Override system prompt for all models')
  .action(async (promptFile: string, options: {
    group?: string;
    models?: string;
    thinking?: boolean;
    showThinking?: boolean;
    output?: string;
    config?: string;
    verbose?: boolean;
    includeMetadata?: boolean;
    systemPrompt?: string;
  }) => {
    try {
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
          `thinktank run ${basename}.txt`,
          `thinktank run ${basename}.txt --group=default`,
          `thinktank run ${basename}.txt --models=openai:gpt-4o`
        ];
        
        throw fileError;
      }
      
      // Parse and validate the models string if provided
      let specificModels: string[] | undefined;
      if (options.models) {
        specificModels = options.models.split(',');
        
        // Validate that each model follows the provider:model format
        for (const model of specificModels) {
          if (!model.includes(':')) {
            throw new ThinktankError(
              `Invalid model format: "${model}". Models must be in provider:modelId format (e.g., openai:gpt-4o).`
            );
          }
          
          const [provider, modelId] = model.split(':');
          if (!provider || !modelId) {
            throw new ThinktankError(
              `Invalid model format: "${model}". Provider and modelId must not be empty.`
            );
          }
        }
        
        // Give user feedback about the selected models
        if (options.verbose) {
          // eslint-disable-next-line no-console
          console.log(
            colors.cyan(`Running prompt against ${specificModels.length} model${specificModels.length === 1 ? '' : 's'}:`)
          );
          specificModels.forEach((model, index) => {
            // eslint-disable-next-line no-console
            console.log(`  ${index + 1}. ${colors.yellow(model)}`);
          });
          // eslint-disable-next-line no-console
          console.log('');
        }
      }
      
      // Check for invalid combinations
      if (options.group && specificModels && specificModels.length > 1) {
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow('Warning: Both --group and --models (multiple) options were provided. ' +
                     'The --models option will filter models within the specified group.')
        );
      }
      
      // Display thinking capability info
      if (options.thinking && options.verbose) {
        // eslint-disable-next-line no-console
        console.log(
          colors.cyan('Thinking capability enabled for models that support it (Claude models)')
        );
      }
      
      // If a specific model is requested but not yet in the config, consider adding it
      if (specificModels && specificModels.length === 1) {
        const [provider, modelId] = specificModels[0].split(':');
        
        // Load the config to check if the model exists
        const config = await configManager.loadConfig({ configPath: options.config });
        const modelExists = configManager.findModel(config, provider, modelId);
        
        if (!modelExists && options.verbose) {
          // eslint-disable-next-line no-console
          console.log(
            colors.yellow(`Model ${specificModels[0]} not found in configuration. ` +
                        `Will attempt to use it anyway. If this fails, add it with:`)
          );
          // eslint-disable-next-line no-console
          console.log(
            colors.dim(`  thinktank config models add ${provider} ${modelId} --enable`)
          );
        }
      }
      
      // Call runThinktank with the parsed options
      await runThinktank({
        input: promptFile,
        configPath: options.config,
        groupName: options.group,
        // For backward compatibility, still set specificModel if there's only one model
        specificModel: specificModels && specificModels.length === 1 ? specificModels[0] : undefined,
        // Always pass the models array to support multiple model selection
        models: specificModels,
        enableThinking: options.thinking,
        includeThinking: options.showThinking,
        output: options.output,
        includeMetadata: options.includeMetadata,
        systemPrompt: options.systemPrompt,
        useColors: true
      });
    } catch (error) {
      handleError(error);
    }
  });

export default runCommand;