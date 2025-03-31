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
import { getModelConfigKey, resolveOutputDirectory, sanitizeFilename } from '../atoms/helpers';
import ora from 'ora';
import fs from 'fs/promises';
import path from 'path';

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
   * Path to the output directory (optional)
   * If provided, this will be used as the parent directory for the run-specific output folder
   * If not provided, a default directory in the current working directory will be used
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
 * Formats an LLM response as Markdown
 * 
 * @param response - The LLM response to format
 * @param includeMetadata - Whether to include metadata in the output
 * @returns The formatted Markdown
 */
function formatResponseAsMarkdown(
  response: LLMResponse & { configKey: string },
  includeMetadata = false
): string {
  const { text, error, metadata, configKey } = response;
  
  // Start with a header
  let markdown = `# ${configKey}\n\n`;
  
  // Add timestamp
  const timestamp = new Date().toISOString();
  markdown += `Generated: ${timestamp}\n\n`;
  
  // Add error if present
  if (error) {
    markdown += `## Error\n\n\`\`\`\n${error}\n\`\`\`\n\n`;
  }
  
  // Add the response text (if available)
  if (text) {
    markdown += `## Response\n\n${text}\n\n`;
  }
  
  // Include metadata if requested
  if (includeMetadata && metadata) {
    markdown += '## Metadata\n\n```json\n';
    markdown += JSON.stringify(metadata, null, 2);
    markdown += '\n```\n';
  }
  
  return markdown;
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
  
  // Track the output directory path for later use
  let outputDirectoryPath: string | undefined;
  
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    
    // 2. Read input file
    spinner.text = 'Reading input file...';
    const prompt = await readFileContent(options.input);
    
    // 2.5 Create output directory if needed
    if (options.output) {
      // Resolve the output directory path
      outputDirectoryPath = resolveOutputDirectory(options.output);
      
      spinner.text = `Creating output directory: ${outputDirectoryPath}`;
      try {
        // Create the directory with recursive option to ensure parent directories exist
        await fs.mkdir(outputDirectoryPath, { recursive: true });
        spinner.info(`Output directory created: ${outputDirectoryPath}`);
      } catch (error) {
        spinner.fail(`Failed to create output directory: ${outputDirectoryPath}`);
        throw new ThinktankError(
          `Failed to create output directory: ${error instanceof Error ? error.message : String(error)}`,
          error instanceof Error ? error : undefined
        );
      }
    }
    
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
    
    // 7. Write individual files if output directory is specified
    if (outputDirectoryPath) {
      spinner.text = 'Writing model responses to individual files...';
      
      // Track stats for reporting
      let succeededWrites = 0;
      let failedWrites = 0;
      const fileWritePromises: Promise<void>[] = [];
      
      // Process each result
      results.forEach((result) => {
        // Create sanitized filename from provider and model
        const sanitizedProvider = sanitizeFilename(result.provider);
        const sanitizedModelId = sanitizeFilename(result.modelId);
        const filename = `${sanitizedProvider}-${sanitizedModelId}.md`;
        
        // Full path to output file
        const filePath = path.join(outputDirectoryPath!, filename);
        
        // Format the response as Markdown
        const markdownContent = formatResponseAsMarkdown(result, options.includeMetadata);
        
        // Create file write promise with error handling
        const writePromise = fs.writeFile(filePath, markdownContent)
          .then(() => {
            succeededWrites++;
          })
          .catch((error) => {
            failedWrites++;
            // Log error but continue with other files
            console.error(`Error writing ${filename}: ${error instanceof Error ? error.message : String(error)}`);
          });
        
        fileWritePromises.push(writePromise);
      });
      
      // Wait for all file writes to complete
      await Promise.all(fileWritePromises);
      
      // Report results
      if (failedWrites === 0) {
        spinner.succeed(`All ${succeededWrites} model responses written to ${outputDirectoryPath}`);
      } else {
        spinner.warn(`Completed with issues: ${succeededWrites} successful, ${failedWrites} failed writes in ${outputDirectoryPath}`);
      }
    } else {
      // Format results for console output only if no output directory
      spinner.text = 'Formatting results...';
      spinner.succeed('Processing completed.');
    }
    
    // Always return formatted results for potential console display by CLI
    const formattedResults = formatResults(results, {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors,
    });
    return formattedResults;
  } catch (error) {
    spinner.fail('An error occurred');
    
    if (error instanceof Error) {
      throw new ThinktankError(`Error running Thinktank: ${error.message}`, error);
    }
    
    throw new ThinktankError('Unknown error running Thinktank');
  }
}