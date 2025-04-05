/**
 * Run command implementation
 * 
 * This command handles sending prompts to LLMs
 */
import { Command } from 'commander';
import fs from 'fs/promises';
import path from 'path';
import { runThinktank } from '../../workflow/runThinktank';
import { 
  ThinktankError, 
  ApiError, 
  ConfigError, 
  FileSystemError,
  createFileNotFoundError,
  createModelFormatError
} from '../../core/errors';
import { colors } from '../../utils/consoleUtils';
import { handleError } from '../index';
import * as configManager from '../../core/configManager';
import { logger } from '../../utils/logger';

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
        // Use the specialized factory function to create a file not found error
        const fileError = createFileNotFoundError(promptFile, `Input file not found: ${promptFile}`);
        
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
            throw createModelFormatError(model, ['openai', 'anthropic', 'google', 'openrouter']);
          }
          
          const [provider, modelId] = model.split(':');
          if (!provider || !modelId) {
            throw createModelFormatError(model, ['openai', 'anthropic', 'google', 'openrouter']);
          }
        }
        
        // Give user feedback about the selected models
        logger.info(
          colors.cyan(`Running prompt against ${specificModels.length} model${specificModels.length === 1 ? '' : 's'}:`)
        );
        specificModels.forEach((model, index) => {
          logger.info(`  ${index + 1}. ${colors.yellow(model)}`);
        });
        logger.info('');
      }
      
      // Check for invalid combinations
      if (options.group && specificModels && specificModels.length > 1) {
        logger.warn('Both --group and --models (multiple) options were provided. ' +
                   'The --models option will filter models within the specified group.');
      }
      
      // Display thinking capability info
      if (options.thinking) {
        logger.info('Thinking capability enabled for models that support it (Claude models)');
      }
      
      // If a specific model is requested but not yet in the config, consider adding it
      if (specificModels && specificModels.length === 1) {
        const [provider, modelId] = specificModels[0].split(':');
        
        // Load the config to check if the model exists
        const config = await configManager.loadConfig({ configPath: options.config });
        const modelExists = configManager.findModel(config, provider, modelId);
        
        if (!modelExists) {
          logger.warn(`Model ${specificModels[0]} not found in configuration. ` +
                    `Will attempt to use it anyway. If this fails, add it with:`);
          logger.info(
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
      // Handle specialized error types
      if (error instanceof ThinktankError) {
        // Error is already properly formatted, just pass it through
        handleError(error);
        return;
      } 
      
      // If not a ThinktankError, check for common error patterns
      if (error instanceof Error) {
        const message = error.message.toLowerCase();
        
        // Handle file-related errors
        if (message.includes('file') || 
            message.includes('directory') || 
            message.includes('path') ||
            message.includes('enoent')) {
          handleError(new FileSystemError(`File system error: ${error.message}`, {
            cause: error,
            filePath: promptFile,
            suggestions: [
              'Check that the prompt file exists and is accessible',
              'Verify that the path is correct',
              'Make sure you have read permissions for the file'
            ],
            examples: [
              `thinktank run prompt.txt`,
              `thinktank run ./path/to/prompt.txt`
            ]
          }));
          return;
        }
        
        // Handle model-related errors
        if (message.includes('model')) {
          handleError(new ConfigError(`Model error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Ensure the model is specified in the correct format (provider:modelId)',
              'Check that the model exists in your configuration',
              options.models ?
                `The provided models were: ${options.models}` :
                'No specific models were provided - try specifying models explicitly'
            ],
            examples: [
              'thinktank run prompt.txt --models=openai:gpt-4o',
              'thinktank run prompt.txt --group=default'
            ]
          }));
          return;
        }
        
        // Handle API-related errors
        if (message.includes('api') || 
            message.includes('key') ||
            message.includes('token') ||
            message.includes('authentication') ||
            message.includes('authorization')) {
          handleError(new ApiError(`API error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Check your API credentials',
              'Verify that you have the correct API keys set in your environment variables',
              'Make sure the provider services are accessible from your network'
            ]
          }));
          return;
        }
        
        // Handle config-related errors
        if (message.includes('config') || message.includes('configuration')) {
          handleError(new ConfigError(`Configuration error: ${error.message}`, {
            cause: error,
            suggestions: [
              'Ensure your configuration file is valid JSON',
              options.config ?
                `The specified config path was: ${options.config}` :
                'No custom config path was provided - using default config location',
              'Try using the default configuration by omitting the --config option'
            ]
          }));
          return;
        }
      }
      
      // Default to passing the error through to the central handler
      handleError(error);
    }
  });

export default runCommand;