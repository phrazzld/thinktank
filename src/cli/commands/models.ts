/**
 * Models command implementation
 * 
 * This command lists available LLM models
 */
import { Command } from 'commander';
import { listAvailableModels } from '../../workflow/listModelsWorkflow';
import { handleError } from '../index';
import {
  ConfigError,
  ValidationError,
  ApiError,
  NetworkError
} from '../../core/errors';

// Create the command
const modelsCommand = new Command('models');

// Configure the command
modelsCommand
  .description('List all available LLM models')
  .option('-p, --provider <provider>', 'Filter models by provider')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (options: {
    provider?: string;
    config?: string;
  }) => {
    try {
      // Validate provider if specified
      if (options.provider && typeof options.provider !== 'string') {
        throw new ValidationError('Provider option must be a string', {
          suggestions: [
            'Specify a valid provider ID (e.g., openai, anthropic, google)',
            'Use --provider=openai format to specify the provider'
          ],
          examples: [
            'thinktank models --provider=openai',
            'thinktank models --provider=anthropic'
          ]
        });
      }
      
      // Call listAvailableModels with the parsed options
      const result = await listAvailableModels({
        provider: options.provider,
        config: options.config
      });
      
      // Display the result
      // eslint-disable-next-line no-console
      console.log(result);
    } catch (error) {
      // If error is already a structured error, pass it through
      if (error instanceof ApiError || 
          error instanceof ConfigError ||
          error instanceof ValidationError ||
          error instanceof NetworkError) {
        handleError(error);
        return;
      }
      
      // Handle other errors with more helpful messages
      if (error instanceof Error) {
        // Check for common error patterns in this command
        const message = error.message.toLowerCase();
        
        // Provider-related errors
        if (message.includes('provider') && 
            (message.includes('invalid') || message.includes('not found'))) {
          handleError(new ConfigError(`Invalid provider: ${options.provider}`, {
            suggestions: [
              'Check that the provider ID is correct',
              'Available providers include: openai, anthropic, google, openrouter',
              'Run without --provider to see all available models'
            ],
            examples: [
              'thinktank models',
              'thinktank models --provider=openai'
            ]
          }));
          return;
        }
        
        // Config-related errors
        if (message.includes('config') || message.includes('configuration')) {
          handleError(new ConfigError(`Configuration error: ${error.message}`, {
            suggestions: [
              'Ensure your configuration file is valid JSON',
              'Check that the config path points to a valid file',
              'Try using the default configuration'
            ],
            examples: [
              'thinktank models',
              'thinktank models --config=/path/to/valid/config.json'
            ]
          }));
          return;
        }
      }
      
      // Default to passing the error through to the central handler
      handleError(error);
    }
  });

export default modelsCommand;