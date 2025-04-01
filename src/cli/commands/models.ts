/**
 * Models command implementation
 * 
 * This command lists available LLM models
 */
import { Command } from 'commander';
import { listAvailableModels } from '../../workflow/listModelsWorkflow';
import { handleError } from '../index';

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
      // Call listAvailableModels with the parsed options
      const result = await listAvailableModels({
        provider: options.provider,
        config: options.config
      });
      
      // Display the result
      // eslint-disable-next-line no-console
      console.log(result);
    } catch (error) {
      handleError(error);
    }
  });

export default modelsCommand;