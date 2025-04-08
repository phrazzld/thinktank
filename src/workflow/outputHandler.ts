/**
 * OutputHandler module for formatting and writing LLM responses
 * 
 * Handles the formatting of LLM responses for console output and file writing,
 * as well as the actual file writing operations.
 */
import path from 'path';
import { LLMResponse } from '../core/types';
import { FileSystem } from '../core/interfaces';
import { formatResults } from '../utils/outputFormatter';
import { 
  sanitizeFilename, 
  generateOutputDirectoryPath 
} from '../utils/helpers';
import { errorCategories } from '../core/errors';

/**
 * Sanitizes control characters from a string
 * 
 * @param input - The input string to sanitize
 * @returns The sanitized string with control characters removed
 */
function sanitizeControlChars(input: string): string {
  let result = '';
  for (let i = 0; i < input.length; i++) {
    const charCode = input.charCodeAt(i);
    if (
      !(charCode >= 0 && charCode <= 8) &&
      !(charCode >= 11 && charCode <= 12) &&
      !(charCode >= 14 && charCode <= 31)
    ) {
      result += input[i];
    }
  }
  return result;
}

/**
 * Error thrown by the OutputHandler module
 */
export class OutputHandlerError extends Error {
  /**
   * The category of error (e.g., "Filesystem", "Formatting", etc.)
   */
  category?: string;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'OutputHandlerError';
    this.category = errorCategories.FILESYSTEM;
  }
}

/**
 * Status of a file write operation
 */
export type FileWriteStatus = 'pending' | 'success' | 'error';

/**
 * Detail of a file write operation
 */
export interface FileWriteDetail {
  /**
   * The model identifier
   */
  modelKey: string;
  
  /**
   * The filename
   */
  filename: string;
  
  /**
   * The full file path
   */
  filePath: string;
  
  /**
   * The status of the file write
   */
  status: FileWriteStatus;
  
  /**
   * Error message, if status is 'error'
   */
  error?: string;
  
  /**
   * Start time of the write operation in milliseconds
   */
  startTime?: number;
  
  /**
   * End time of the write operation in milliseconds
   */
  endTime?: number;
  
  /**
   * Duration of the write operation in milliseconds
   */
  durationMs?: number;
}

/**
 * Represents data for a file to be written
 */
export interface FileData {
  /**
   * The filename
   */
  filename: string;
  
  /**
   * The content to be written to the file
   */
  content: string;
  
  /**
   * The model key (provider:modelId) associated with this file
   */
  modelKey: string;
}

/**
 * Options for file output
 */
export interface FileOutputOptions {
  /**
   * The base output directory path
   * If not provided, a default path will be used
   */
  outputDirectory?: string;
  
  /**
   * Optional identifier for the output directory (e.g., model name, group name)
   */
  directoryIdentifier?: string;
  
  /**
   * Optional friendly name for the output directory
   * If provided, this will be used instead of a timestamp
   */
  friendlyRunName?: string;
  
  /**
   * Whether to include metadata in the output files
   */
  includeMetadata?: boolean;
  
  /**
   * Whether to throw errors or just record them in the results
   */
  throwOnError?: boolean;
  
  /**
   * Status update callback
   * 
   * Called when a file's write status changes
   */
  onStatusUpdate?: (
    fileDetail: FileWriteDetail, 
    allDetails: FileWriteDetail[]
  ) => void;
}

/**
 * Result of file write operations
 */
export interface FileOutputResult {
  /**
   * The directory where files were written
   */
  outputDirectory: string;
  
  /**
   * Details of file write operations
   */
  files: FileWriteDetail[];
  
  /**
   * Count of successful writes
   */
  succeededWrites: number;
  
  /**
   * Count of failed writes
   */
  failedWrites: number;
  
  /**
   * Timing information
   */
  timing: {
    /**
     * Start time of all write operations in milliseconds
     */
    startTime: number;
    
    /**
     * End time of all write operations in milliseconds
     */
    endTime: number;
    
    /**
     * Duration of all write operations in milliseconds
     */
    durationMs: number;
  };
}

/**
 * Options for console output
 */
export interface ConsoleOutputOptions {
  /**
   * Whether to include metadata in the output
   */
  includeMetadata?: boolean;
  
  /**
   * Whether to use colors in the output
   */
  useColors?: boolean;
  
  /**
   * Whether to include thinking output in the results
   */
  includeThinking?: boolean;
  
  /**
   * Whether to use tabular format for results
   */
  useTable?: boolean;
}

/**
 * Formats an LLM response as Markdown for file output
 * 
 * @param response - The LLM response to format
 * @param includeMetadata - Whether to include metadata in the output
 * @returns The formatted Markdown
 */
export function formatResponseAsMarkdown(
  response: LLMResponse & { configKey: string },
  includeMetadata = false
): string {
  const { text, error, metadata, configKey, groupInfo } = response;
  
  // Start with a header including group information if available
  let markdown = `# ${configKey}`;
  if (groupInfo && groupInfo.name !== 'default') {
    markdown += ` (${groupInfo.name} group)`;
  }
  markdown += '\n\n';
  
  // Add timestamp
  const timestamp = new Date().toISOString();
  markdown += `Generated: ${timestamp}\n\n`;
  
  // Add group information if available and not default
  if (groupInfo && groupInfo.name !== 'default') {
    markdown += `Group: ${groupInfo.name}\n`;
    if (groupInfo.systemPrompt && includeMetadata) {
      markdown += `System Prompt: "${groupInfo.systemPrompt.text}"\n`;
    }
    markdown += '\n';
  }
  
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
 * Generate a filename for a model response
 * 
 * @param response - The model response
 * @param options - Options for filename generation
 * @returns The generated filename
 */
export function generateFilename(
  response: LLMResponse & { configKey: string },
  options: { includeGroup?: boolean } = {}
): string {
  const { provider, modelId, groupInfo } = response;
  const { includeGroup = true } = options;
  
  // Sanitize components
  const sanitizedProvider = sanitizeFilename(provider);
  const sanitizedModelId = sanitizeFilename(modelId);
  
  // Format the filename to include group information when relevant
  if (includeGroup && groupInfo?.name && groupInfo.name !== 'default') {
    // Include group in filename
    const sanitizedGroupName = sanitizeFilename(groupInfo.name);
    return `${sanitizedGroupName}-${sanitizedProvider}-${sanitizedModelId}.md`;
  } else {
    return `${sanitizedProvider}-${sanitizedModelId}.md`;
  }
}

/**
 * Format responses for console output
 * 
 * @param responses - Array of LLM responses with their config keys
 * @param options - Console output options
 * @returns Formatted string for console display
 */
export function formatForConsole(
  responses: Array<LLMResponse & { configKey: string }>,
  options: ConsoleOutputOptions = {}
): string {
  if (responses.length === 0) {
    return 'No results to display.';
  }
  
  // Use the generic formatter with console-specific options
  return formatResults(responses, {
    includeMetadata: options.includeMetadata,
    useColors: options.useColors !== false, // Default to true
    includeThinking: options.includeThinking,
    useTable: options.useTable
  });
}

/**
 * Create output directory for file writing
 * 
 * @param options - File output options
 * @param fileSystem - File system interface to use for operations
 * @returns Created directory path
 * @throws {OutputHandlerError} If directory creation fails
 */
export async function createOutputDirectory(
  options: Pick<FileOutputOptions, 'outputDirectory' | 'directoryIdentifier' | 'friendlyRunName'> = {},
  fileSystem: FileSystem
): Promise<string> {
  // Generate output directory path
  const outputDirectoryPath = generateOutputDirectoryPath(
    options.outputDirectory,
    options.directoryIdentifier,
    options.friendlyRunName
  );
  
  try {
    // Create the directory with recursive option to ensure parent directories exist
    await fileSystem.mkdir(outputDirectoryPath, { recursive: true });
    return outputDirectoryPath;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new OutputHandlerError(
      `Failed to create output directory: ${errorMessage}`,
      error instanceof Error ? error : undefined
    );
  }
}

/**
 * Write LLM responses to files
 * 
 * @param responses - Array of LLM responses with their config keys
 * @param outputDirectory - Directory to write files to
 * @param options - File output options
 * @param fileSystem - File system interface to use for operations
 * @returns Result of file write operations
 */
export async function writeResponsesToFiles(
  responses: Array<LLMResponse & { configKey: string }>,
  outputDirectory: string,
  options: Omit<FileOutputOptions, 'outputDirectory' | 'directoryIdentifier'> = {},
  fileSystem: FileSystem
): Promise<FileOutputResult> {
  // Start timing
  const startTime = Date.now();
  
  // Initialize result objects
  const fileDetails: FileWriteDetail[] = [];
  let succeededWrites = 0;
  let failedWrites = 0;
  
  // Process each response
  const processFile = async (response: LLMResponse & { configKey: string }): Promise<void> => {
    // Generate filename
    const filename = generateFilename(response);
    
    // Create file path
    const filePath = path.join(outputDirectory, filename);
    
    // Format the response as Markdown
    const markdownContent = formatResponseAsMarkdown(response, options.includeMetadata);
    
    // Create file detail object for tracking
    const fileDetail: FileWriteDetail = {
      modelKey: response.configKey,
      filename,
      filePath,
      status: 'pending',
      startTime: Date.now()
    };
    
    // Add to tracking array
    fileDetails.push(fileDetail);
    
    // Call status update callback if provided
    if (options.onStatusUpdate) {
      options.onStatusUpdate(fileDetail, fileDetails);
    }
    
    try {
      // Sanitize content to handle control characters
      const sanitizedContent = typeof markdownContent === 'string' 
        ? sanitizeControlChars(markdownContent)
        : markdownContent;
      
      // Create parent directory if it doesn't exist (for extra safety)
      const parentDir = path.dirname(filePath);
      await fileSystem.mkdir(parentDir, { recursive: true });
      
      // Write file with atomic operation if possible
      const tempPath = `${filePath}.tmp`;
      await fileSystem.writeFile(tempPath, sanitizedContent);
      
      // Handle rename - note: FileSystem interface might not have rename directly
      // We'll need to read and write again if that's the case
      try {
        // Try using writeFile to effectively "rename" by overwriting
        await fileSystem.writeFile(filePath, sanitizedContent);
        
        // If we reached here, try to clean up the temp file
        try {
          await fileSystem.writeFile(tempPath, ''); // Can't unlink, so empty the file
        } catch {
          // Ignore cleanup errors
        }
      } catch (renameError) {
        // If we can't rename/move, try direct write
        await fileSystem.writeFile(filePath, sanitizedContent);
      }
      
      // Calculate duration for success case
      const endTime = Date.now();
      const durationMs = endTime - (fileDetail.startTime || endTime);
      
      // Update status with success
      fileDetail.status = 'success';
      fileDetail.endTime = endTime;
      fileDetail.durationMs = durationMs;
      
      succeededWrites++;
      
      // Call status update callback if provided
      if (options.onStatusUpdate) {
        options.onStatusUpdate(fileDetail, fileDetails);
      }
    } catch (error) {
      // Calculate duration for error case
      const endTime = Date.now();
      const durationMs = endTime - (fileDetail.startTime || endTime);
      
      // Update status with error
      fileDetail.status = 'error';
      fileDetail.error = error instanceof Error ? error.message : String(error);
      fileDetail.endTime = endTime;
      fileDetail.durationMs = durationMs;
      
      failedWrites++;
      
      // Call status update callback if provided
      if (options.onStatusUpdate) {
        options.onStatusUpdate(fileDetail, fileDetails);
      }
      
      // Throw if requested
      if (options.throwOnError) {
        throw new OutputHandlerError(
          `Failed to write file ${fileDetail.filename}: ${fileDetail.error}`,
          error instanceof Error ? error : undefined
        );
      }
      
      // Clean up temp file if we failed during the rename operation
      try {
        const tempPath = `${filePath}.tmp`;
        await fileSystem.writeFile(tempPath, ''); // Clean up by emptying
      } catch {
        // Ignore errors during cleanup
      }
    }
  };
  
  // Create promises for all files
  const fileWritePromises = responses.map(response => processFile(response));
  
  // Wait for all file writes to complete
  await Promise.allSettled(fileWritePromises);
  
  // Calculate overall timing
  const endTime = Date.now();
  const durationMs = endTime - startTime;
  
  // Return the results
  return {
    outputDirectory,
    files: fileDetails,
    succeededWrites,
    failedWrites,
    timing: {
      startTime,
      endTime,
      durationMs
    }
  };
}

/**
 * Process responses for output, preparing both file data and console output
 * 
 * This is a pure function that doesn't perform I/O operations like file writing
 * or directory creation. Instead, it returns structured data that can be used
 * by the caller to perform those operations.
 * 
 * @param responses - Array of LLM responses with their config keys
 * @param options - Output options
 * @returns Result object with file data and console formatted string
 */
export function processOutput(
  responses: Array<LLMResponse & { configKey: string }>,
  options: FileOutputOptions & ConsoleOutputOptions = {}
): {
  files: FileData[];
  directoryPath: string;
  consoleOutput: string;
} {
  // Generate output directory path
  const directoryPath = generateOutputDirectoryPath(
    options.outputDirectory,
    options.directoryIdentifier,
    options.friendlyRunName
  );
  
  // Generate file data for each response
  const files: FileData[] = responses.map(response => {
    // Generate filename
    const filename = generateFilename(response, {
      // Control if group name is included based on how models were selected
      includeGroup: true // Default behavior, caller can override if needed
    });
    
    // Format content for file output
    const content = formatResponseAsMarkdown(
      response, 
      options.includeMetadata
    );
    
    return {
      filename,
      content,
      modelKey: response.configKey,
    };
  });
  
  // Format for console output
  const consoleOutput = formatForConsole(
    responses,
    {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors,
      includeThinking: options.includeThinking,
      useTable: options.useTable
    }
  );
  
  return {
    files,
    directoryPath,
    consoleOutput
  };
}
