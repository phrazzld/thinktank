/**
 * Main orchestration for the Thinktank application
 * 
 * This template connects all the components and orchestrates the workflow.
 */
import { readFileContent } from '../molecules/fileReader';
import { loadConfig, filterModels, getEnabledModels, validateModelApiKeys } from '../organisms/configManager';
import { getProvider } from '../organisms/llmRegistry';
import { formatResults } from '../molecules/outputFormatter';
import { LLMResponse, ModelConfig } from '../atoms/types';
import { getModelConfigKey } from '../atoms/helpers';
import ora from 'ora';
import fs from 'fs/promises';

// Import provider modules to ensure they're registered
import '../molecules/llmProviders/openai';
// Future providers will be imported here

/**
 * Options for running Thinktank
 */
export interface RunOptions {
  /**
   * Path to the input prompt file
   */
  input: string;
  
  /**
   * Path to the configuration file (optional)
   */
  configPath?: string;
  
  /**
   * Path to the output file (optional)
   */
  output?: string;
  
  /**
   * Array of model identifiers to use (optional)
   * If not provided, all enabled models will be used
   */
  models?: string[];
  
  /**
   * Whether to include metadata in the output
   */
  includeMetadata?: boolean;
  
  /**
   * Whether to use colors in the output
   */
  useColors?: boolean;
}

/**
 * Error class for Thinktank runtime errors
 */
export class ThinktankError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'ThinktankError';
  }
}

/**
 * Main function to run Thinktank
 * 
 * @param options - Options for running Thinktank
 * @returns The formatted results
 * @throws {ThinktankError} If an error occurs during execution
 */
export async function runThinktank(options: RunOptions): Promise<string> {
  const spinner = ora('Starting Thinktank...').start();
  
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    
    // 2. Read input file
    spinner.text = 'Reading input file...';
    const prompt = await readFileContent(options.input);
    
    // 3. Filter models based on CLI args
    spinner.text = 'Preparing models...';
    let models: ModelConfig[];
    
    if (options.models && options.models.length > 0) {
      // Filter models by CLI args
      models = options.models.flatMap(modelFilter => 
        filterModels(config, modelFilter)
      );
      
      // Remove duplicates
      const modelKeys = new Set<string>();
      models = models.filter(model => {
        const key = getModelConfigKey(model);
        if (modelKeys.has(key)) {
          return false;
        }
        modelKeys.add(key);
        return true;
      });
      
      // Filter to only enabled models
      models = models.filter(model => model.enabled);
      
      if (models.length === 0) {
        spinner.warn('No enabled models matched the specified filters.');
        return 'No enabled models matched the specified filters.';
      }
    } else {
      // Use all enabled models
      models = getEnabledModels(config);
      
      if (models.length === 0) {
        spinner.warn('No enabled models found in configuration.');
        return 'No enabled models found in configuration.';
      }
    }
    
    // 4. Validate API keys
    const { missingKeyModels } = validateModelApiKeys(config);
    
    // Log warnings for models with missing API keys
    if (missingKeyModels.length > 0) {
      const modelNames = missingKeyModels.map(getModelConfigKey).join(', ');
      spinner.warn(`Missing API keys for models: ${modelNames}`);
      
      // Filter out models with missing keys
      models = models.filter(model => 
        !missingKeyModels.some(m => 
          m.provider === model.provider && m.modelId === model.modelId
        )
      );
      
      if (models.length === 0) {
        spinner.fail('No models with valid API keys available.');
        return 'No models with valid API keys available.';
      }
    }
    
    // 5. Prepare API calls
    spinner.text = `Sending prompt to ${models.length} model(s)...`;
    const callPromises: Array<Promise<LLMResponse & { configKey: string }>> = [];
    
    // For each model, get provider and send prompt
    models.forEach(model => {
      const provider = getProvider(model.provider);
      const configKey = getModelConfigKey(model);
      
      if (!provider) {
        spinner.warn(`Provider not found for ${configKey}`);
        return;
      }
      
      // Create promise for this model
      const responsePromise = provider.generate(prompt, model.modelId, model.options)
        .then(response => ({
          ...response,
          configKey,
        }))
        .catch(error => ({
          provider: model.provider,
          modelId: model.modelId,
          text: '',
          error: error instanceof Error ? error.message : String(error),
          configKey,
        }));
      
      callPromises.push(responsePromise);
    });
    
    // 6. Execute calls concurrently
    const results = await Promise.all(callPromises);
    
    // 7. Format results
    spinner.text = 'Formatting results...';
    const formattedResults = formatResults(results, {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors,
    });
    
    // 8. Write to file if specified
    if (options.output) {
      spinner.text = `Writing results to ${options.output}...`;
      await fs.writeFile(options.output, formattedResults);
      spinner.succeed(`Results written to ${options.output}`);
    } else {
      spinner.succeed('Done!');
    }
    
    return formattedResults;
  } catch (error) {
    spinner.fail('An error occurred');
    
    if (error instanceof Error) {
      throw new ThinktankError(`Error running Thinktank: ${error.message}`, error);
    }
    
    throw new ThinktankError('Unknown error running Thinktank');
  }
}