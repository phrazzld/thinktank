/**
 * Output formatter for displaying LLM responses in a readable format
 */
import chalk from 'chalk';
import { LLMResponse, LLMAvailableModel } from '../atoms/types';

/**
 * Format options for the output
 */
export interface FormatOptions {
  /**
   * Whether to include metadata in the output
   */
  includeMetadata?: boolean;
  
  /**
   * Whether to use colors in the output
   */
  useColors?: boolean;
  
  /**
   * Whether to include the response text
   */
  includeText?: boolean;
  
  /**
   * Whether to include errors in the output
   */
  includeErrors?: boolean;
  
  /**
   * Custom separator between model outputs
   */
  separator?: string;
}

/**
 * Default format options
 */
const DEFAULT_FORMAT_OPTIONS: FormatOptions = {
  includeMetadata: false,
  useColors: true,
  includeText: true,
  includeErrors: true,
  separator: '\n\n' + '-'.repeat(80) + '\n\n',
};

/**
 * Formats a single LLM response
 * 
 * @param response - The LLM response to format
 * @param options - Format options
 * @returns The formatted response as a string
 */
export function formatResponse(
  response: LLMResponse, 
  options: FormatOptions = {}
): string {
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  const { useColors, includeMetadata, includeText, includeErrors } = opts;
  
  const configKey = `${response.provider}:${response.modelId}`;
  const lines: string[] = [];
  
  // Format the header with provider and model info
  const header = `Model: ${configKey}`;
  lines.push(useColors ? chalk.bold.blue(header) : header);
  
  // Include error if present and requested
  if (response.error && includeErrors) {
    const errorText = `Error: ${response.error}`;
    lines.push(useColors ? chalk.red(errorText) : errorText);
  }
  
  // Include the response text if requested and available
  if (includeText && response.text) {
    lines.push('');
    lines.push(response.text);
  }
  
  // Include metadata if requested and available
  if (includeMetadata && response.metadata) {
    lines.push('');
    const metadataHeader = 'Metadata:';
    lines.push(useColors ? chalk.gray(metadataHeader) : metadataHeader);
    
    // Format each metadata entry
    Object.entries(response.metadata).forEach(([key, value]) => {
      const metadataLine = `  ${key}: ${JSON.stringify(value)}`;
      lines.push(useColors ? chalk.gray(metadataLine) : metadataLine);
    });
  }
  
  return lines.join('\n');
}

/**
 * Formats multiple LLM responses
 * 
 * @param responses - Array of LLM responses
 * @param options - Format options
 * @returns The formatted responses as a string
 */
export function formatResponses(
  responses: LLMResponse[], 
  options: FormatOptions = {}
): string {
  if (responses.length === 0) {
    return 'No responses to display.';
  }
  
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  
  // Format each response
  const formattedResponses = responses.map(response => 
    formatResponse(response, opts)
  );
  
  // Join with separator
  return formattedResponses.join(opts.separator);
}

/**
 * Formats multiple LLM responses with their respective config keys
 * 
 * @param results - Array of LLM responses with their config keys
 * @param options - Format options
 * @returns The formatted responses as a string
 */
export function formatResults(
  results: Array<LLMResponse & { configKey: string }>,
  options: FormatOptions = {}
): string {
  if (results.length === 0) {
    return 'No results to display.';
  }
  
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  
  // Format each result
  const formattedResults = results.map(result => {
    // Extract config key components
    const [provider, modelId] = result.configKey.split(':');
    
    // Create a normalized response with the config key components
    const response: LLMResponse = {
      ...result,
      provider: provider || result.provider,
      modelId: modelId || result.modelId,
    };
    
    return formatResponse(response, opts);
  });
  
  // Join with separator
  return formattedResults.join(opts.separator);
}

/**
 * Formats a list of available models grouped by provider
 * 
 * @param modelsByProvider - Record mapping provider ID to array of models or error
 * @param options - Format options
 * @returns The formatted model list as a string
 */
export function formatModelList(
  modelsByProvider: Record<string, LLMAvailableModel[] | { error: string }>,
  options: FormatOptions = {}
): string {
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  const { useColors } = opts;
  
  const lines: string[] = [];
  
  // Header
  const header = 'Available Models:';
  lines.push(useColors ? chalk.bold.blue(header) : header);
  lines.push('');
  
  // Check if there are any providers
  if (Object.keys(modelsByProvider).length === 0) {
    const noProvidersMessage = 'No providers configured.';
    lines.push(useColors ? chalk.gray(noProvidersMessage) : noProvidersMessage);
    return lines.join('\n');
  }
  
  // Format each provider's models
  for (const providerId in modelsByProvider) {
    // Provider header
    const providerHeader = `--- ${providerId} ---`;
    lines.push(useColors ? chalk.bold.blue(providerHeader) : providerHeader);
    
    const result = modelsByProvider[providerId];
    
    // Check if this provider returned an error
    if ('error' in result) {
      const errorLine = `  Error fetching models: ${result.error}`;
      lines.push(useColors ? chalk.red(errorLine) : errorLine);
    } else {
      // Check if the provider returned any models
      if (result.length === 0) {
        const noModelsLine = '  (No models available)';
        lines.push(useColors ? chalk.gray(noModelsLine) : noModelsLine);
      } else {
        // Format each model
        result.forEach(model => {
          let modelLine = `  - ${model.id}`;
          if (model.description) {
            modelLine += ` (${model.description})`;
          }
          lines.push(modelLine);
        });
      }
    }
    
    // Add a blank line between providers
    lines.push('');
  }
  
  // Remove the last empty line if present
  if (lines[lines.length - 1] === '') {
    lines.pop();
  }
  
  return lines.join('\n');
}