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
  .description('Run a prompt against LLM models')
  .argument('<promptFile>', 'Path to the file containing the prompt')
  .option('-m, --models <models>', 'Comma-separated list of specific models to use (e.g., openai:gpt-4o,anthropic:claude-3-opus)')
  .option('-g, --group <group>', 'Name of a model group from config to use')
  .option('-o, --output <directory>', 'Directory to save results to')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-v, --verbose', 'Show detailed output during execution')
  .option('--include-metadata', 'Include raw API response metadata in output')
  .option('--system-prompt <system-prompt>', 'Custom system prompt for supported models')
  .action(async (promptFile: string, options: {
  models?: string;
  group?: string;
  output?: string;
  config?: string;
  verbose?: boolean;
  includeMetadata?: boolean;
  systemPrompt?: string;
}): Promise<void> => {
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
      
      // Validate group exists in config if provided
      if (options.group) {
        const configOptions: configManager.LoadConfigOptions = options.config ? { configPath: options.config } : {};
        const config = await configManager.loadConfig(configOptions);
        // Define a proper interface for the config structure
        interface ConfigWithGroups {
          groups?: Record<string, {name: string}>;
        }
        // Use a type guard to handle different config formats
        const configWithGroups = config as ConfigWithGroups;
        
        // Check if the groups object exists and if the specified group exists in it
        const groupExists = configWithGroups.groups && configWithGroups.groups[options.group];
        if (!groupExists) {
          // Get available groups for error message
          const availableGroups = configWithGroups.groups ? Object.keys(configWithGroups.groups) : [];
          
          throw new ConfigError(`Model group "${options.group}" not found in config`, {
            suggestions: [
              'Check that the group name matches exactly (case-sensitive)',
              'Available groups: ' + (availableGroups.length > 0 
                ? availableGroups.join(', ') 
                : 'none defined'),
              'Define your groups in thinktank.config.json'
            ],
            cause: new Error(`Model group "${options.group}" not found in config`)
          });
        }
        
        // If verbose, show group models
        if (options.verbose) {
          // Pass correct types to this function
          const groupModels = configManager.getEnabledModelsFromGroups(config, [options.group]);
          logger.info(
            colors.cyan(`Running with model group "${options.group}" (${groupModels.length} models)`)
          );
          if (groupModels.length > 0) {
            groupModels.forEach((model, index) => {
              logger.info(`  ${index + 1}. ${colors.yellow(`${model.provider}:${model.modelId}`)}`);
            });
            logger.info('');
          } else {
            logger.warn(
              colors.yellow('Warning: Selected group has no enabled models')
            );
          }
        }
      }
      
      // Validate output directory if provided
      if (options.output) {
        try {
          await fs.mkdir(options.output, { recursive: true });
        } catch (error) {
          throw new FileSystemError(`Failed to create output directory: ${options.output}`, {
            cause: error instanceof Error ? error : undefined,
            suggestions: [
              'Check that the directory path is valid',
              'Ensure you have write permissions to this location',
              'Try creating the directory manually first'
            ]
          });
        }
      }
      
      // If systemPrompt is provided, validate it
      let systemPrompt = options.systemPrompt;
      if (systemPrompt) {
        // Attempt to read from file if it starts with @ 
        if (systemPrompt.startsWith('@')) {
          const systemPromptFile = systemPrompt.substring(1);
          try {
            systemPrompt = await fs.readFile(systemPromptFile, 'utf-8');
          } catch (error) {
            throw new FileSystemError(`System prompt file not found: ${systemPromptFile}`, {
              cause: error instanceof Error ? error : undefined,
              filePath: systemPromptFile,
              suggestions: [
                'Check that the file exists and is accessible',
                'Use the format --system-prompt=@path/to/prompt.txt',
                'Or provide the system prompt directly: --system-prompt="Your prompt here"'
              ]
            });
          }
        }
      }
      
      // Run the core function
      await runThinktank({
        input: promptFile,
        specificModel: specificModels ? specificModels.join(',') : undefined,
        groupName: options.group,
        output: options.output,
        configPath: options.config,
        includeMetadata: options.includeMetadata,
        systemPrompt
      });
      
      // Output completion message if verbose
      if (options.verbose) {
        logger.info(colors.green('\nExecution complete!'));
        
        if (options.output) {
          logger.info(`Results saved to: ${options.output}`);
        } else {
          logger.info('Results displayed above (no output directory specified)');
        }
      }
      
      // Just return, no need to return results (void return type)
      return;
    } catch (error) {
      // Convert to ThinktankError if it's not already one
      // We'll use this for error handling
      if (!(error instanceof ThinktankError)) {
        handleError(new ThinktankError(
          error instanceof Error ? error.message : String(error),
          { cause: error instanceof Error ? error : undefined }
        ));
        return;
      }
      
      // Try to provide specific guidance for common error types
      if (error instanceof Error) {
        const message = error.message.toLowerCase();
        
        // Handle API key errors
        if (message.includes('api') && 
            (message.includes('key') || 
             message.includes('token') ||
             message.includes('authentication') ||
             message.includes('authorization'))) {
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