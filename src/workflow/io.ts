/**
 * I/O module for centralizing all workflow I/O operations
 *
 * This module provides centralized I/O handling for the workflow module, abstracting
 * away direct file system and console interactions from the core business logic.
 * It serves as a dedicated layer for all I/O operations used across the workflow.
 */
import path from 'path';
import { FileSystem, ConsoleLogger, UISpinner } from '../core/interfaces';
import { FileWriteDetail, FileData, FileOutputResult } from './outputHandler';
import { FileSystemError } from '../core/errors/types/filesystem';
import { styleSuccess, styleInfo, styleWarning, styleError } from '../utils/consoleUtils';

/**
 * Options for file I/O operations
 */
export interface FileIOOptions {
  /**
   * Whether to throw errors or just record them in the results
   */
  throwOnError?: boolean;

  /**
   * Whether to display detailed error information
   */
  verbose?: boolean;
}

/**
 * Result of a directory creation operation
 */
export interface DirectoryCreationResult {
  /**
   * The path to the created directory
   */
  directoryPath: string;

  /**
   * Whether the directory was created successfully
   */
  success: boolean;

  /**
   * Error message if creation failed
   */
  error?: string;
}

/**
 * Creates a directory in the file system
 *
 * @param dirPath - The path to the directory to create
 * @param fileSystem - The file system interface to use
 * @param options - Options for directory creation
 * @returns Promise resolving to a DirectoryCreationResult
 */
export async function createDirectory(
  dirPath: string,
  fileSystem: FileSystem,
  options: FileIOOptions = {}
): Promise<DirectoryCreationResult> {
  try {
    // Ensure the directory exists (create it if it doesn't)
    await fileSystem.mkdir(dirPath, { recursive: true });

    return {
      directoryPath: dirPath,
      success: true,
    };
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);

    // If throwOnError is true, throw the error
    if (options.throwOnError) {
      throw new FileSystemError(`Failed to create directory: ${dirPath}`, {
        cause: error instanceof Error ? error : undefined,
        filePath: dirPath,
      });
    }

    // Otherwise, return a result object with the error
    return {
      directoryPath: dirPath,
      success: false,
      error: errorMessage,
    };
  }
}

/**
 * Writes a collection of files to the file system
 *
 * This function handles all file I/O operations, including directory creation,
 * file writing, error handling, and status tracking.
 *
 * @param files - Array of FileData objects containing files to write
 * @param outputDirectoryPath - Directory where files should be written
 * @param fileSystem - File system interface for I/O operations
 * @param options - Options for file writing
 * @returns Promise resolving to a FileOutputResult with success/failure statistics
 * @throws {FileSystemError} If directory creation fails and throwOnError is true
 */
export async function writeFiles(
  files: FileData[],
  outputDirectoryPath: string,
  fileSystem: FileSystem,
  options: FileIOOptions = {}
): Promise<FileOutputResult> {
  // Start timing for file operations
  const fileWriteStartTime = Date.now();

  // Track file write stats
  let succeededWrites = 0;
  let failedWrites = 0;
  const fileDetails: FileWriteDetail[] = [];

  // Ensure output directory exists
  try {
    await fileSystem.mkdir(outputDirectoryPath, { recursive: true });
  } catch (error) {
    if (options.throwOnError) {
      throw new FileSystemError(`Failed to create output directory: ${outputDirectoryPath}`, {
        cause: error instanceof Error ? error : undefined,
        filePath: outputDirectoryPath,
      });
    }

    // Return early with failure result if we can't create the directory
    return {
      outputDirectory: outputDirectoryPath,
      files: [],
      succeededWrites: 0,
      failedWrites: files.length,
      timing: {
        startTime: fileWriteStartTime,
        endTime: Date.now(),
        durationMs: Date.now() - fileWriteStartTime,
      },
    };
  }

  // Process each file
  for (const file of files) {
    const filePath = path.join(outputDirectoryPath, file.filename);

    // Create file detail for tracking
    const fileDetail: FileWriteDetail = {
      modelKey: file.modelKey,
      filename: file.filename,
      filePath,
      status: 'pending',
      startTime: Date.now(),
    };

    fileDetails.push(fileDetail);

    try {
      // Create parent directory if needed (for nested paths)
      const parentDir = path.dirname(filePath);
      await fileSystem.mkdir(parentDir, { recursive: true });

      // Write the file
      await fileSystem.writeFile(filePath, file.content);

      // Update stats
      succeededWrites++;

      // Mark as success
      fileDetail.status = 'success';
      fileDetail.endTime = Date.now();
      fileDetail.durationMs = fileDetail.endTime - (fileDetail.startTime || fileDetail.endTime);
    } catch (error) {
      // Update stats
      failedWrites++;

      // Mark as error
      fileDetail.status = 'error';
      fileDetail.error = error instanceof Error ? error.message : String(error);
      fileDetail.endTime = Date.now();
      fileDetail.durationMs = fileDetail.endTime - (fileDetail.startTime || fileDetail.endTime);

      // Throw if requested
      if (options.throwOnError) {
        throw new FileSystemError(`Failed to write file: ${filePath}`, {
          cause: error instanceof Error ? error : undefined,
          filePath,
        });
      }
    }
  }

  // Calculate overall timing
  const fileWriteEndTime = Date.now();
  const fileWriteDurationMs = fileWriteEndTime - fileWriteStartTime;

  // Create file output result object
  return {
    outputDirectory: outputDirectoryPath,
    files: fileDetails,
    succeededWrites,
    failedWrites,
    timing: {
      startTime: fileWriteStartTime,
      endTime: fileWriteEndTime,
      durationMs: fileWriteDurationMs,
    },
  };
}

/**
 * Log a file output result to the console
 *
 * @param result - The file output result to log
 * @param logger - The console logger to use
 * @param options - Options for console output
 */
export function logFileOutputResult(
  result: FileOutputResult,
  logger: ConsoleLogger,
  options: { verbose?: boolean } = {}
): void {
  const { succeededWrites, failedWrites, outputDirectory } = result;

  // Log summary
  if (failedWrites === 0) {
    logger.success(
      `${styleSuccess('✓')} Wrote ${succeededWrites} ${succeededWrites === 1 ? 'file' : 'files'} to ${outputDirectory}`
    );
  } else if (succeededWrites > 0) {
    logger.warn(
      `${styleWarning('⚠')} Wrote ${succeededWrites} ${succeededWrites === 1 ? 'file' : 'files'} to ${outputDirectory} (${failedWrites} failed)`
    );
  } else {
    logger.error(`${styleError('✗')} Failed to write any files to ${outputDirectory}`);
  }

  // If verbose, log individual file details
  if (options.verbose && (failedWrites > 0 || result.files.length > 1)) {
    logger.info(styleInfo('File details:'));

    // Log each file detail
    for (const file of result.files) {
      if (file.status === 'success') {
        logger.info(`  ${styleSuccess('✓')} ${file.filename}`);
      } else if (file.status === 'error') {
        logger.error(`  ${styleError('✗')} ${file.filename}: ${file.error || 'Unknown error'}`);
      }
    }
  }
}

/**
 * Update a spinner with file output progress
 *
 * @param result - The file output result to update from
 * @param spinner - The spinner to update
 */
export function updateSpinnerWithFileOutput(
  result: FileOutputResult,
  spinner: UISpinner | { text: string }
): void {
  const { succeededWrites, failedWrites } = result;

  if (failedWrites === 0) {
    spinner.text = `Files written: ${succeededWrites} succeeded`;
  } else {
    spinner.text = `Files written: ${succeededWrites} succeeded, ${failedWrites} failed`;
  }
}
