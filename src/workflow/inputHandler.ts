/**
 * Input handler module for processing prompt input from various sources
 * 
 * This module handles loading prompts from files, stdin, or direct text,
 * with proper error handling and validation.
 */
import fs from 'fs/promises';
import path from 'path';
import { normalizeText } from '../utils/helpers';
import { FileSystem } from '../core/interfaces';
// Create a local error class to avoid circular dependencies
export class ThinktankError extends Error {
  /**
   * The category of error (e.g., "File System", "API", etc.)
   */
  category?: string;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands related to this error context
   */
  examples?: string[];
  
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'ThinktankError';
  }
}
import { errorCategories } from '../core/errors';

/**
 * Input source types supported by the InputHandler
 */
export enum InputSourceType {
  FILE = 'file',
  STDIN = 'stdin',
  TEXT = 'text',
}

/**
 * Options for input handling
 */
export interface InputOptions {
  /**
   * Path to the input file, content directly, or '-' for stdin
   */
  input: string;
  
  /**
   * Whether to normalize the input text (trim whitespace, etc.)
   * Defaults to true
   */
  normalize?: boolean;
  
  /**
   * The source type of the input
   * If not provided, it will be determined based on the input
   */
  sourceType?: InputSourceType;
  
  /**
   * Timeout for reading from stdin in milliseconds
   * Defaults to 30000 (30 seconds)
   */
  stdinTimeout?: number;
  
  /**
   * Optional FileSystem interface for file operations
   * If provided, it will be used instead of direct fs operations
   */
  fileSystem?: FileSystem;
}

/**
 * Result of processing input, including metadata about the source
 */
export interface InputResult {
  /**
   * The processed prompt content
   */
  content: string;
  
  /**
   * The source type that was used
   */
  sourceType: InputSourceType;
  
  /**
   * Source path if applicable (for files)
   */
  sourcePath?: string;
  
  /**
   * Metadata about the input processing
   */
  metadata: {
    /**
     * Processing time in milliseconds
     */
    processingTimeMs: number;
    
    /**
     * Original content length before any processing
     */
    originalLength: number;
    
    /**
     * Final content length after processing
     */
    finalLength: number;
    
    /**
     * Whether the content was normalized
     */
    normalized: boolean;
  };
}

/**
 * Error class for input handling errors
 */
export class InputError extends ThinktankError {
  constructor(message: string, public readonly cause?: Error) {
    super(message, cause);
    this.name = 'InputError';
    this.category = errorCategories.INPUT;
  }
}

/**
 * Reads and processes input from a file
 * 
 * @param filePath - Path to the file to read
 * @param normalize - Whether to normalize the content
 * @param fileSystem - Optional FileSystem interface for file operations
 * @returns The processed content and metadata
 * @throws {InputError} If the file cannot be read
 */
async function processFileInput(
  filePath: string, 
  normalize = true,
  fileSystem?: FileSystem
): Promise<InputResult> {
  // Track processing time
  const startTime = Date.now();
  
  try {
    // Resolve to absolute path if relative path is provided
    const resolvedPath = path.isAbsolute(filePath) 
      ? filePath 
      : path.resolve(process.cwd(), filePath);
    
    // Read file content using the provided FileSystem or direct fs operations
    let content: string;
    
    if (fileSystem) {
      try {
        // Use the FileSystem interface if provided
        content = await fileSystem.readFileContent(resolvedPath, { normalize: false });
      } catch (error) {
        // Convert FileSystem errors to InputError for consistent handling
        if (error instanceof Error) {
          if (error.message.includes('not found') || error.message.includes('ENOENT')) {
            const inputError = new InputError(`Input file not found: ${filePath}`);
            inputError.suggestions = [
              'Check that the file exists and the path is correct',
              'Use an absolute path if you are not in the same directory',
              `Current working directory: ${process.cwd()}`
            ];
            throw inputError;
          } else if (error.message.includes('permission') || error.message.includes('EACCES')) {
            const inputError = new InputError(`Permission denied to read file: ${filePath}`);
            inputError.suggestions = [
              'Check that you have read permissions for the file',
              'Try running the command with elevated permissions'
            ];
            throw inputError;
          }
        }
        
        throw new InputError(`Error reading file: ${filePath}`, error instanceof Error ? error : undefined);
      }
    } else {
      // Fall back to direct fs operations when no FileSystem is provided
      try {
        // Check if file exists and is readable
        await fs.access(resolvedPath, fs.constants.R_OK);
      } catch (error) {
        // Handle specific file access errors
        const nodeError = error as NodeJS.ErrnoException;
        if (nodeError.code === 'ENOENT') {
          const inputError = new InputError(`Input file not found: ${filePath}`);
          inputError.suggestions = [
            'Check that the file exists and the path is correct',
            'Use an absolute path if you are not in the same directory',
            `Current working directory: ${process.cwd()}`
          ];
          throw inputError;
        } else if (nodeError.code === 'EACCES') {
          const inputError = new InputError(`Permission denied to read file: ${filePath}`);
          inputError.suggestions = [
            'Check that you have read permissions for the file',
            'Try running the command with elevated permissions'
          ];
          throw inputError;
        }
        
        // Generic error case
        throw new InputError(`Error accessing file: ${filePath}`, error instanceof Error ? error : undefined);
      }
      
      // Read file content
      try {
        content = await fs.readFile(resolvedPath, 'utf-8');
      } catch (error) {
        throw new InputError(`Error reading file: ${filePath}`, error instanceof Error ? error : undefined);
      }
    }
    
    // Track original length before any processing
    const originalLength = content.length;
    
    // Normalize content if requested
    if (normalize) {
      content = normalizeText(content);
    }
    
    // Calculate processing time
    const processingTimeMs = Date.now() - startTime;
    
    // Return result with metadata
    return {
      content,
      sourceType: InputSourceType.FILE,
      sourcePath: resolvedPath,
      metadata: {
        processingTimeMs,
        originalLength,
        finalLength: content.length,
        normalized: normalize
      }
    };
  } catch (error) {
    // If it's already an InputError, rethrow it
    if (error instanceof InputError) {
      throw error;
    }
    
    // Otherwise, wrap it in an InputError
    if (error instanceof Error) {
      throw new InputError(`Error processing file input: ${filePath}`, error);
    }
    
    // Generic error case
    throw new InputError(`Unknown error processing file input: ${filePath}`);
  }
}

/**
 * Reads and processes input from stdin
 * 
 * @param timeout - Timeout in milliseconds
 * @param normalize - Whether to normalize the content
 * @returns The processed content and metadata
 * @throws {InputError} If stdin cannot be read or times out
 */
async function processStdinInput(
  timeout = 30000,
  normalize = true
): Promise<InputResult> {
  // Track processing time
  const startTime = Date.now();
  
  return new Promise((resolve, reject) => {
    // Set timeout to prevent indefinite waiting
    const timeoutId = setTimeout(() => {
      // Clean up listeners
      process.stdin.removeAllListeners('data');
      process.stdin.removeAllListeners('end');
      process.stdin.removeAllListeners('error');
      process.stdin.pause();
      
      // Reject with timeout error
      reject(new InputError(`Stdin read timeout after ${timeout}ms`));
    }, timeout);
    
    // Buffer to collect chunks
    const chunks: Buffer[] = [];
    
    // Setup stdin handlers
    process.stdin.on('data', (chunk) => {
      chunks.push(Buffer.from(chunk));
    });
    
    process.stdin.on('end', () => {
      // Clear timeout since we received the full input
      clearTimeout(timeoutId);
      
      // Combine chunks and convert to string
      const buffer = Buffer.concat(chunks);
      let content = buffer.toString('utf-8');
      
      // Track original length
      const originalLength = content.length;
      
      // Normalize if requested
      if (normalize) {
        content = normalizeText(content);
      }
      
      // Calculate processing time
      const processingTimeMs = Date.now() - startTime;
      
      // Resolve with result
      resolve({
        content,
        sourceType: InputSourceType.STDIN,
        metadata: {
          processingTimeMs,
          originalLength,
          finalLength: content.length,
          normalized: normalize
        }
      });
    });
    
    process.stdin.on('error', (error) => {
      // Clear timeout since we got an error
      clearTimeout(timeoutId);
      
      // Reject with input error
      reject(new InputError('Error reading from stdin', error));
    });
    
    // Ensure stdin is in flowing mode
    process.stdin.resume();
  });
}

/**
 * Processes direct text input
 * 
 * @param text - The text to process
 * @param normalize - Whether to normalize the content
 * @returns The processed content and metadata
 */
function processTextInput(text: string, normalize = true): InputResult {
  // Track processing time
  const startTime = Date.now();
  
  // Track original length
  const originalLength = text.length;
  
  // Normalize if requested
  if (normalize) {
    text = normalizeText(text);
  }
  
  // Calculate processing time
  const processingTimeMs = Date.now() - startTime;
  
  // Return result with metadata
  return {
    content: text,
    sourceType: InputSourceType.TEXT,
    metadata: {
      processingTimeMs,
      originalLength,
      finalLength: text.length,
      normalized: normalize
    }
  };
}

/**
 * Determines the input source type based on the input
 * 
 * @param input - The input to check
 * @returns The determined input source type
 */
function determineInputSourceType(input: string): InputSourceType {
  // If input is '-', use stdin
  if (input === '-') {
    return InputSourceType.STDIN;
  }
  
  // If input starts with certain prefixes indicating direct text, use text
  if (input.startsWith('`') || 
      input.startsWith('"') || 
      input.startsWith("'") || 
      input.includes('\n')) {
    return InputSourceType.TEXT;
  }
  
  // Otherwise, treat as file path
  return InputSourceType.FILE;
}

/**
 * Main function to process input from various sources
 * 
 * @param options - Options for input processing
 * @returns The processed input content and metadata
 * @throws {InputError} If input cannot be processed
 */
export async function processInput(options: InputOptions): Promise<InputResult> {
  const { 
    input, 
    normalize = true, 
    sourceType: explicitSourceType,
    stdinTimeout = 30000,
    fileSystem
  } = options;
  
  // Validate input
  if (!input) {
    throw new InputError('Input is required');
  }
  
  // Determine source type if not explicitly provided
  const sourceType = explicitSourceType || determineInputSourceType(input);
  
  // Process based on source type
  switch (sourceType) {
    case InputSourceType.FILE:
      return processFileInput(input, normalize, fileSystem);
    
    case InputSourceType.STDIN:
      return processStdinInput(stdinTimeout, normalize);
    
    case InputSourceType.TEXT:
      return processTextInput(input, normalize);
    
    default:
      throw new InputError(`Unsupported input source type: ${String(sourceType)}`);
  }
}
