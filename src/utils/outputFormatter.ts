/**
 * Output formatter for displaying LLM responses in a readable format
 * 
 * Functions in this module are pure - they take data structures as input
 * and return formatted strings as output. They do not perform any direct I/O
 * operations or modify global state.
 */

/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
import Table from 'cli-table3';
import type { Cell } from 'cli-table3';
import { 
  colors, 
  styleSuccess, 
  styleError, 
  styleDim, 
  styleHeader 
} from './consoleUtils';
import { LLMResponse, LLMAvailableModel } from '../core/types';
// Import types from workflow/types when needed in refactored functions
import { formatCompletionSummary } from './formatCompletionSummary';

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
   * Whether to include thinking output when available
   */
  includeThinking?: boolean;
  
  /**
   * Custom separator between model outputs
   */
  separator?: string;
  
  /**
   * Whether to use tabular format for results display
   */
  useTable?: boolean;
}

/**
 * Default format options
 */
const DEFAULT_FORMAT_OPTIONS: FormatOptions = {
  includeMetadata: false,
  useColors: true,
  includeText: true,
  includeErrors: true,
  includeThinking: false,
  separator: '\n\n' + '-'.repeat(80) + '\n\n',
  useTable: false,
};

/**
 * Formats a single LLM response for output to markdown file
 * 
 * @param response - The LLM response to format
 * @param options - Format options
 * @returns The formatted response as a markdown string
 */
export function formatResponseForMarkdownFile(
  response: LLMResponse, 
  options: FormatOptions = {}
): string {
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  const { includeMetadata, includeText, includeErrors, includeThinking } = opts;
  
  const configKey = `${response.provider}:${response.modelId}`;
  const lines: string[] = [];
  
  // Format the header with provider and model info
  const header = `# Model: ${configKey}`;
  lines.push(header);
  
  // Include group information if available
  if (response.groupInfo && response.groupInfo.name !== 'default') {
    const groupLine = `Group: ${response.groupInfo.name}`;
    lines.push(groupLine);
  }
  
  // Add timestamp
  const timestamp = new Date().toISOString();
  lines.push(`Generated: ${timestamp}`);
  lines.push('');
  
  // Include error if present and requested
  if (response.error && includeErrors) {
    lines.push(`## Error\n\n\`\`\`\n${response.error}\n\`\`\`\n`);
  }
  
  // Include the response text if requested and available
  if (includeText && response.text) {
    lines.push('## Response\n');
    lines.push(response.text);
  }
  
  // Include thinking output if requested and available
  if (includeThinking && response.metadata?.thinking) {
    lines.push('\n## Thinking\n');
    
    // Type guard for thinking data with a process field
    const thinking = response.metadata.thinking;
    if (typeof thinking === 'object' && thinking !== null && 
        'process' in thinking && typeof thinking.process === 'string') {
      lines.push('```text');
      lines.push(thinking.process);
      lines.push('```');
    } else {
      lines.push('```json');
      lines.push(JSON.stringify(thinking, null, 2));
      lines.push('```');
    }
  }
  
  // Include metadata if requested and available
  if (includeMetadata && response.metadata) {
    lines.push('\n## Metadata\n');
    lines.push('```json');
    
    // Create a copy of metadata to modify safely
    const metadataCopy = { ...response.metadata };
    
    // Skip thinking in metadata if we've already displayed it separately
    if (includeThinking && 'thinking' in metadataCopy) {
      delete metadataCopy.thinking;
    }
    
    lines.push(JSON.stringify(metadataCopy, null, 2));
    lines.push('```');
  }
  
  return lines.join('\n');
}

/**
 * Formats a single LLM response (alias for backwards compatibility)
 * @deprecated Use formatResponseForMarkdownFile instead
 */
export function formatResponse(
  response: LLMResponse, 
  options: FormatOptions = {}
): string {
  return formatResponseForMarkdownFile(response, options);
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
 * Formats results in a tabular format
 * 
 * @param results - Array of LLM responses with their config keys
 * @param options - Format options
 * @returns Formatted results table as a string
 */
export function formatResultsTable(
  results: Array<LLMResponse & { configKey: string }>,
  options: FormatOptions = {}
): string {
  if (results.length === 0) {
    return 'No results to display.';
  }
  
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  const { useColors, includeMetadata } = opts;
  
  // Define the table columns
  const table = new Table({
    head: [
      useColors ? colors.bold('Model') : 'Model',
      useColors ? colors.bold('Group') : 'Group',
      useColors ? colors.bold('Status') : 'Status',
      useColors ? colors.bold('Time') : 'Time',
      useColors ? colors.bold('Tokens') : 'Tokens',
    ],
    colAligns: ['left', 'left', 'center', 'right', 'right'],
    style: useColors ? { head: ['blue'] } : undefined
  });
  
  // Sort results by group then model
  const sortedResults = [...results].sort((a, b) => {
    const groupA = a.groupInfo?.name || 'default';
    const groupB = b.groupInfo?.name || 'default';
    if (groupA !== groupB) return groupA.localeCompare(groupB);
    return a.configKey.localeCompare(b.configKey);
  });
  
  // Group the results for potential group statistics
  const groupedResults = new Map<string, Array<LLMResponse & { configKey: string }>>();
  
  // Add each result to the table
  sortedResults.forEach(result => {
    // Determine the status text
    let statusText: string;
    if (result.error) {
      statusText = useColors ? styleError('Error') : 'Error';
    } else {
      statusText = useColors ? styleSuccess('Success') : 'Success';
    }
    
    // Get the group name
    const groupName = result.groupInfo?.name || 'default';
    
    // Format the group name
    const displayGroupName = groupName === 'default' 
      ? (useColors ? styleDim('default') : 'default')
      : groupName;
    
    // Get response time if available
    let responseTime: string | number = '-';
    
    if (result.metadata?.responseTime !== undefined) {
      const metadataTime = result.metadata.responseTime;
      // Handle different types that might be stored in responseTime
      if (typeof metadataTime === 'number' || typeof metadataTime === 'string') {
        responseTime = metadataTime;
      }
    }
    
    // Get token count if available
    // Type guard for usage object with token properties
    const usage = result.metadata?.usage;
    let tokens: string | number = '-';
    
    if (usage && typeof usage === 'object' && usage !== null) {
      // Check for known token properties in order of preference
      const usageRecord = usage as Record<string, unknown>;
      if ('total_tokens' in usageRecord && typeof usageRecord.total_tokens === 'number') {
        tokens = usageRecord.total_tokens;
      } else if ('completion_tokens' in usageRecord && typeof usageRecord.completion_tokens === 'number') {
        tokens = usageRecord.completion_tokens;
      } else if ('prompt_tokens' in usageRecord && typeof usageRecord.prompt_tokens === 'number') {
        tokens = usageRecord.prompt_tokens;
      }
    }
    
    // Add to the table
    // Convert all values to proper Cell type for cli-table3
    const row: Cell[] = [
      result.configKey,
      displayGroupName,
      statusText,
      responseTime,
      tokens
    ];
    table.push(row);
    
    // Add to grouped results for statistics
    if (!groupedResults.has(groupName)) {
      groupedResults.set(groupName, []);
    }
    
    // Since we just checked or created the entry, we know it exists
    const groupResults = groupedResults.get(groupName);
    if (groupResults) {
      groupResults.push(result);
    }
  });
  
  let output = table.toString();
  
  // Add group statistics if we have metadata
  if (includeMetadata) {
    const groupStats: string[] = [];
    
    groupedResults.forEach((groupResults, groupName) => {
      const total = groupResults.length;
      const success = groupResults.filter(r => !r.error).length;
      const error = total - success;
      
      // Calculate average response time if available
      const timesWithValues = groupResults
        .map(r => Number(r.metadata?.responseTime || 0))
        .filter(t => t > 0);
      
      const avgTime = timesWithValues.length > 0
        ? Math.round(timesWithValues.reduce((a, b) => a + b, 0) / timesWithValues.length)
        : undefined;
      
      const groupTitle = groupName === 'default' 
        ? 'Default Group' 
        : `Group: ${groupName}`;
      
      let groupStat = `\n${useColors ? styleHeader(groupTitle) : groupTitle}`;
      groupStat += `\n  Models: ${total}, Success: ${success}, Errors: ${error}`;
      if (avgTime !== undefined) {
        groupStat += `, Avg. Time: ${avgTime}ms`;
      }
      
      groupStats.push(groupStat);
    });
    
    if (groupStats.length > 0) {
      output += '\n\nGroup Statistics:' + groupStats.join('\n');
    }
  }
  
  return output;
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
  
  // Use table format if requested
  if (opts.useTable) {
    return formatResultsTable(results, opts);
  }
  
  // Format each result using the traditional approach
  const formattedResults = results.map(result => {
    // Extract config key components
    const [provider, modelId] = result.configKey.split(':');
    
    // Create a normalized response with the config key components
    const response: LLMResponse = {
      ...result,
      provider: provider || result.provider,
      modelId: modelId || result.modelId,
    };
    
    return formatResponseForMarkdownFile(response, opts);
  });
  
  // Join with separator
  return formattedResults.join(opts.separator);
}

/**
 * Formats LLM responses for console display
 * 
 * This is a specialized function for formatting results specifically for
 * console output. It acts as a thin wrapper around formatResults or formatResultsTable
 * depending on the options.
 * 
 * @param results - Array of LLM responses with their config keys
 * @param options - Format options
 * @returns The formatted console output as a string
 */
export function formatResultsForConsole(
  results: Array<LLMResponse & { configKey: string }>,
  options: FormatOptions = {}
): string {
  // Set up specific options for console output
  const consoleOptions: FormatOptions = {
    ...options,
    includeMetadata: options.includeMetadata ?? false,
    useColors: options.useColors ?? true,
    includeText: options.includeText ?? true,
    includeErrors: options.includeErrors ?? true,
  };
  
  return formatResults(results, consoleOptions);
}

/**
 * Formats a list of available models grouped by provider
 * 
 * @param modelsByProvider - Record mapping provider ID to array of models or error
 * @param options - Format options
 * @returns The formatted model list as a string
 */
// Re-export formatCompletionSummary from the separate module
export { formatCompletionSummary };

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
  lines.push(useColors ? styleHeader(header) : header);
  lines.push('');
  
  // Check if there are any providers
  if (Object.keys(modelsByProvider).length === 0) {
    const noProvidersMessage = 'No providers configured.';
    lines.push(useColors ? styleDim(noProvidersMessage) : noProvidersMessage);
    return lines.join('\n');
  }
  
  // Format each provider's models
  for (const providerId in modelsByProvider) {
    // Provider header
    const providerHeader = `--- ${providerId} ---`;
    lines.push(useColors ? styleHeader(providerHeader) : providerHeader);
    
    const result = modelsByProvider[providerId];
    
    // Check if this provider returned an error
    if ('error' in result) {
      const errorLine = `  Error fetching models: ${result.error}`;
      lines.push(useColors ? styleError(errorLine) : errorLine);
    } else {
      // Check if the provider returned any models
      if (result.length === 0) {
        const noModelsLine = '  (No models available)';
        lines.push(useColors ? styleDim(noModelsLine) : noModelsLine);
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
