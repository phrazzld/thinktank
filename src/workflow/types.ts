/**
 * Data interfaces for pure functions refactored to separate I/O from business logic
 * 
 * This file defines the data structures returned by pure functions after their refactoring
 * to remove direct I/O operations. These interfaces represent the data that would have been
 * written to files or logged to the console, but is now returned as structured data instead.
 */

/**
 * Represents data for a file to be written.
 * Used as a core type across multiple interfaces.
 */
export interface FileData {
  /** The filename (e.g., 'openai-gpt-4o.md') */
  filename: string;
  /** The formatted content string to be written to the file */
  content: string;
  /** Associated model key, useful for linking back to results */
  modelKey: string;
}

/**
 * Result returned by the refactored _processOutput function.
 * Note: This is different from the original ProcessOutputResult in runThinktankTypes.ts
 * and will replace it during the refactoring process.
 */
export interface PureProcessOutputResult {
  /** Array of file data objects to be written to the filesystem */
  files: FileData[];
  /** String content intended for console display */
  consoleOutput: string;
}

/**
 * Raw data needed to generate a completion summary.
 * Contains all the statistics and information needed to format a summary.
 */
export interface CompletionSummaryData {
  /** Total number of models queried */
  totalModels: number;
  /** Number of models that returned successful responses */
  successCount: number;
  /** Number of models that failed to return responses */
  failureCount: number;
  /** Detailed information about errors that occurred */
  errors: Array<{
    /** The model identifier */
    modelKey: string;
    /** The error message */
    message: string;
    /** The error category if available */
    category?: string;
  }>;
  /** Optional friendly name for the run */
  runName?: string;
  /** The output directory path where files were written */
  outputDirectoryPath: string;
  /** Total execution time in milliseconds */
  totalExecutionTimeMs?: number;
}

/**
 * Result returned by the refactored _logCompletionSummary function.
 * Contains formatted text ready for console output.
 */
export interface PureCompletionSummaryResult {
  /** Formatted summary text ready for display */
  summaryText: string;
  /** Optional detailed error information strings */
  errorDetails?: string[];
}
