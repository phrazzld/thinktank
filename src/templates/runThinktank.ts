/**
 * Main orchestration for the thinktank application
 * 
 * This template connects all the components and orchestrates the workflow.
 */
import { readFileContent } from '../molecules/fileReader';
import { loadConfig, filterModels, getEnabledModels, validateModelApiKeys } from '../organisms/configManager';
import { getProvider } from '../organisms/llmRegistry';
import { formatResults } from '../molecules/outputFormatter';
import { LLMResponse, ModelConfig } from '../atoms/types';
import { getModelConfigKey, generateOutputDirectoryPath, sanitizeFilename } from '../atoms/helpers';
import ora from 'ora';
import fs from 'fs/promises';
import path from 'path';

// Import provider modules to ensure they're registered
import '../molecules/llmProviders/openai';
import '../molecules/llmProviders/anthropic';
// Future providers will be imported here

/**
 * Options for running thinktank
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
   * If not provided, './thinktank-reports/' in the current working directory will be used
   * Note: Model responses are always written to files in a timestamped subdirectory
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
 * Error class for thinktank runtime errors
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
 * Main function to run thinktank
 * 
 * @param options - Options for running thinktank
 * @returns The formatted results
 * @throws {ThinktankError} If an error occurs during execution
 */
export async function runThinktank(options: RunOptions): Promise<string> {
  const spinner = ora('Starting thinktank...').start();
  
  // Track the output directory path for later use
  let outputDirectoryPath: string | undefined;
  
  // For tracking model statuses
  const modelStatuses: Record<string, { status: 'pending' | 'success' | 'error', message?: string }> = {};
  
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    
    // 2. Read input file
    spinner.text = 'Reading input file...';
    const prompt = await readFileContent(options.input);
    
    // 2.5 Create output directory - this is now always done
    // Generate the output directory path with timestamped subdirectory
    outputDirectoryPath = generateOutputDirectoryPath(options.output);
    
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
    spinner.text = `Preparing to query ${models.length} model${models.length === 1 ? '' : 's'}...`;
    
    // List models being used
    const modelList = models.map(model => getModelConfigKey(model)).join(', ');
    spinner.info(`Models: ${modelList}`);
    
    // Initialize status tracking
    models.forEach(model => {
      const configKey = getModelConfigKey(model);
      modelStatuses[configKey] = { status: 'pending' };
    });
    
    const callPromises: Array<Promise<LLMResponse & { configKey: string }>> = [];
    
    // For each model, get provider and send prompt
    models.forEach(model => {
      const provider = getProvider(model.provider);
      const configKey = getModelConfigKey(model);
      
      if (!provider) {
        spinner.warn(`Provider not found for ${configKey}`);
        modelStatuses[configKey] = { 
          status: 'error', 
          message: 'Provider not found' 
        };
        return;
      }
      
      // Create promise for this model
      const responsePromise = provider.generate(prompt, model.modelId, model.options)
        .then(response => {
          // Update status
          modelStatuses[configKey] = { status: 'success' };
          
          // Update spinner text with progress
          const pendingCount = Object.values(modelStatuses).filter(s => s.status === 'pending').length;
          const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
          const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
          spinner.text = `Processing models: ${successCount} complete, ${pendingCount} pending, ${errorCount} failed`;
          
          return {
            ...response,
            configKey,
          };
        })
        .catch(error => {
          // Update status
          modelStatuses[configKey] = { 
            status: 'error', 
            message: error instanceof Error ? error.message : String(error)
          };
          
          // Update spinner text with progress
          const pendingCount = Object.values(modelStatuses).filter(s => s.status === 'pending').length;
          const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
          const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
          spinner.text = `Processing models: ${successCount} complete, ${pendingCount} pending, ${errorCount} failed`;
          
          return {
            provider: model.provider,
            modelId: model.modelId,
            text: '',
            error: error instanceof Error ? error.message : String(error),
            configKey,
          };
        });
      
      callPromises.push(responsePromise);
    });
    
    // Initial status message
    spinner.text = `Sending prompt to ${models.length} model${models.length === 1 ? '' : 's'}...`;
    
    // 6. Execute calls concurrently
    const results = await Promise.all(callPromises);
    
    // 7. Show model completion summary
    const successCount = Object.values(modelStatuses).filter(s => s.status === 'success').length;
    const errorCount = Object.values(modelStatuses).filter(s => s.status === 'error').length;
    
    if (errorCount > 0) {
      // Log models with errors
      spinner.warn(`${successCount} of ${successCount + errorCount} models completed successfully`);
      
      // Display error details
      const errorModels = Object.entries(modelStatuses)
        .filter(([_, status]) => status.status === 'error')
        .map(([model, status]) => `  - ${model}: ${status.message || 'Unknown error'}`);
      
      console.log('\nModels with errors:');
      console.log(errorModels.join('\n'));
    } else {
      spinner.succeed(`All ${successCount} models completed successfully`);
    }
    
    // 8. Write individual files (now always done)
    spinner.text = 'Writing model responses to individual files...';
    
    // Track stats for reporting
    let succeededWrites = 0;
    let failedWrites = 0;
    const fileWritePromises: Promise<void>[] = [];
    type FileDetail = {
      model: string;
      filename: string;
      status: 'pending' | 'success' | 'error';
      error?: string;
    };
    
    const fileDetails: FileDetail[] = [];
    
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
      
      // Add to tracking array
      const fileDetail: FileDetail = {
        model: result.configKey,
        filename,
        status: 'pending'
      };
      fileDetails.push(fileDetail);
      
      // Create file write promise with error handling
      const writePromise = fs.writeFile(filePath, markdownContent)
        .then(() => {
          succeededWrites++;
          fileDetail.status = 'success';
          // Update progress after each file
          const progressPercent = Math.round((succeededWrites + failedWrites) / results.length * 100);
          spinner.text = `Writing files: ${progressPercent}% complete (${succeededWrites + failedWrites}/${results.length})`;
        })
        .catch((error) => {
          failedWrites++;
          fileDetail.status = 'error';
          fileDetail.error = error instanceof Error ? error.message : String(error);
          // Update progress after each file
          const progressPercent = Math.round((succeededWrites + failedWrites) / results.length * 100);
          spinner.text = `Writing files: ${progressPercent}% complete (${succeededWrites + failedWrites}/${results.length})`;
        });
      
      fileWritePromises.push(writePromise);
    });
    
    // Wait for all file writes to complete
    await Promise.all(fileWritePromises);
    
    // Report results
    if (failedWrites === 0) {
      spinner.succeed(`All ${succeededWrites} model responses written to ${outputDirectoryPath}`);
      console.log(`\nOutput directory: ${outputDirectoryPath}`);
    } else {
      spinner.warn(`Completed with issues: ${succeededWrites} successful, ${failedWrites} failed writes`);
      console.log(`\nOutput directory: ${outputDirectoryPath}`);
      
      // Show files with errors
      const failedFiles = fileDetails.filter(file => file.status === 'error');
      console.log('\nFiles with errors:');
      failedFiles.forEach(file => {
        console.log(`  - ${file.filename}: ${file.error || 'Unknown error'}`);
      });
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
      throw new ThinktankError(`Error running thinktank: ${error.message}`, error);
    }
    
    throw new ThinktankError('Unknown error running thinktank');
  }
}