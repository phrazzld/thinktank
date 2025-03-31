#!/usr/bin/env node
/**
 * Command-line interface for Thinktank
 */
import yargs from 'yargs';
import { hideBin } from 'yargs/helpers';
import { runThinktank, ThinktankError } from '../templates/runThinktank';
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
  const argv = await yargs(hideBin(process.argv))
    .usage('Usage: $0 --input <file> [options]')
    .option('input', {
      alias: 'i',
      describe: 'Path to input prompt file',
      type: 'string',
      demandOption: true,
    })
    .option('config', {
      alias: 'c',
      describe: 'Path to configuration file',
      type: 'string',
    })
    .option('output', {
      alias: 'o',
      describe: 'Path to output file',
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
    .help()
    .alias('help', 'h')
    .version()
    .alias('version', 'v')
    .example('$0 -i prompt.txt', 'Send prompt.txt to all enabled models')
    .example('$0 -i prompt.txt -m openai:gpt-4o', 'Send prompt to specific model')
    .example('$0 -i prompt.txt -c custom-config.json', 'Use custom config file')
    .example('$0 -i prompt.txt -o results.txt', 'Write results to file')
    .epilogue('For more information, visit https://github.com/phrazzld/thinktank')
    .parseAsync();
  
  try {
    // Verify input file exists
    try {
      await fs.access(argv.input);
    } catch (error) {
      throw new ThinktankError(`Input file not found: ${argv.input}`);
    }
    
    // Run Thinktank
    const result = await runThinktank({
      input: argv.input,
      configPath: argv.config,
      output: argv.output,
      models: argv.model as string[] | undefined,
      includeMetadata: argv.metadata,
      useColors: !argv['no-color'],
    });
    
    // Print result to console if not writing to file
    if (!argv.output) {
      // eslint-disable-next-line no-console
      console.log(result);
    }
    
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