/**
 * Output formatter for displaying LLM responses in a readable format
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
} from '../atoms/consoleUtils';
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
  const { useColors, includeMetadata, includeText, includeErrors, includeThinking } = opts;
  
  const configKey = `${response.provider}:${response.modelId}`;
  const lines: string[] = [];
  
  // Format the header with provider and model info
  const header = `Model: ${configKey}`;
  lines.push(useColors ? styleHeader(header) : header);
  
  // Include error if present and requested
  if (response.error && includeErrors) {
    const errorText = `Error: ${response.error}`;
    lines.push(useColors ? styleError(errorText) : errorText);
  }
  
  // Include the response text if requested and available
  if (includeText && response.text) {
    lines.push('');
    lines.push(response.text);
  }
  
  // Include thinking output if requested and available
  if (includeThinking && response.metadata?.thinking) {
    lines.push('');
    const thinkingHeader = 'Thinking:';
    lines.push(useColors ? styleHeader(thinkingHeader) : thinkingHeader);
    
    const thinking = response.metadata.thinking as { process?: string };
    if (thinking.process) {
      lines.push(thinking.process);
    } else {
      lines.push(JSON.stringify(thinking, null, 2));
    }
  }
  
  // Include metadata if requested and available
  if (includeMetadata && response.metadata) {
    lines.push('');
    const metadataHeader = 'Metadata:';
    lines.push(useColors ? styleDim(metadataHeader) : metadataHeader);
    
    // Format each metadata entry
    Object.entries(response.metadata).forEach(([key, value]) => {
      // Skip thinking in metadata if we've already displayed it separately
      if (key === 'thinking' && includeThinking) return;
      
      const metadataLine = `  ${key}: ${JSON.stringify(value)}`;
      lines.push(useColors ? styleDim(metadataLine) : metadataLine);
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
    const responseTime = result.metadata?.responseTime || '-';
    
    // Get token count if available
    const usage = result.metadata?.usage as Record<string, number> | undefined;
    const tokens = usage?.total_tokens || 
                  usage?.completion_tokens || 
                  usage?.prompt_tokens || 
                  '-';
    
    // Add to the table
    table.push([
      result.configKey as Cell,
      displayGroupName as Cell,
      statusText as Cell,
      responseTime as Cell,
      tokens as Cell
    ]);
    
    // Add to grouped results for statistics
    if (!groupedResults.has(groupName)) {
      groupedResults.set(groupName, []);
    }
    groupedResults.get(groupName)!.push(result);
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