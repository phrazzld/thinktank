#!/usr/bin/env node
/**
 * Command-line interface for thinktank
 */
import yargs from 'yargs';
import { hideBin } from 'yargs/helpers';
import { runthinktank, thinktankError } from '../templates/runthinktank';
import { listAvailableModels } from '../templates/listModelsWorkflow';
import fs from 'fs/promises';
import dotenv from 'dotenv';

// Load environment variables early
dotenv.config();

/**
 * Main CLI entry point
 * @export For testing purposes
 */
export async function main(): Promise<void> {
  // Define CLI arguments
  const yargsInstance = yargs(hideBin(process.argv))
    .usage('Usage: $0 [command] [options]')
    .option('config', {
      alias: 'c',
      describe: 'Path to configuration file',
      type: 'string',
    })
    // Command for querying LLMs
    .command('$0', 'Send a prompt to LLMs and compare responses', (yargs) => {
      return yargs
        .option('input', {
          alias: 'i',
          describe: 'Path to input prompt file',
          type: 'string',
          demandOption: true,
        })
        .option('output', {
          alias: 'o',
          describe: 'Custom path for output directory (default: ./thinktank-reports/)',
          type: 'string',
        })
        .option('model', {
          alias: 'm',
          describe: 'Models to use (provider:model, provider, or model)',
          type: 'array',
        })
        .option('metadata', {
          describe: 'Include metadata in output',
          type: 'boolean',
          default: false,
        })
        .option('no-color', {
          describe: 'Disable colored output',
          type: 'boolean',
          default: false,
        })
        .example('$0 -i prompt.txt', 'Send prompt.txt to all enabled models (output to ./thinktank_outputs/)')
        .example('$0 -i prompt.txt -m openai:gpt-4o', 'Send prompt to specific model')
        .example('$0 -i prompt.txt -c custom-config.json', 'Use custom config file')
        .example('$0 -i prompt.txt -o ./custom-outputs', 'Use custom directory for output files');
    })
    // Command for listing available models
    .command('list-models', 'List available models from providers', (yargs) => {
      return yargs
        .option('provider', {
          alias: 'p',
          describe: 'Filter models by provider ID',
          type: 'string',
        })
        .example('$0 list-models', 'List all available models')
        .example('$0 list-models -p anthropic', 'List only Anthropic models')
        .example('$0 list-models -c custom-config.json', 'Use custom config file');
    })
    .help()
    .alias('help', 'h')
    .version()
    .alias('version', 'v')
    .epilogue('For more information, visit https://github.com/phrazzld/thinktank');
    
  const argv = await yargsInstance.parseAsync();
  
  try {
    // Determine the command to run based on argv._
    const command = argv._[0];
    
    if (command === 'list-models') {
      // Run the list-models command
      const result = await listAvailableModels({
        config: argv.config,
        provider: argv.provider as string | undefined,
      });
      
      // Display the results
      // eslint-disable-next-line no-console
      console.log(result);
      
      process.exit(0);
    } else {
      // Default command (thinktank with input file)
      
      // Verify input file exists
      if (!argv.input) {
        throw new thinktankError('Input file is required. Use --input or -i to specify the input file.');
      }
      
      try {
        await fs.access(argv.input as string);
      } catch (error) {
        throw new thinktankError(`Input file not found: ${argv.input as string}`);
      }
      
      // Run thinktank - all model responses are written to the output directory
      // We're running thinktank but not using the returned results
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      await runthinktank({
        input: argv.input as string,
        configPath: argv.config,
        output: argv.output as string | undefined,
        models: argv.model as string[] | undefined,
        includeMetadata: argv.metadata as boolean | undefined,
        useColors: !(argv['no-color'] as boolean | undefined),
      });
      
      process.exit(0);
    }
  } catch (error) {
    // Handle errors
    if (error instanceof thinktankError) {
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