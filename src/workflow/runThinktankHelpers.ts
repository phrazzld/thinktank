/**
 * Helper functions for the runThinktank workflow
 * 
 * This file contains the implementation of helper functions that encapsulate
 * distinct phases of the runThinktank workflow, making the main function more
 * modular and easier to maintain.
 */
import { loadConfig } from '../core/configManager';
import { generateFunName } from '../utils/nameGenerator';
import { createOutputDirectory } from './outputHandler';
import { 
  ThinktankError,
  ConfigError,
  FileSystemError,
  PermissionError,
  errorCategories
} from '../core/errors';
import { styleInfo, styleSuccess } from '../utils/consoleUtils';
import {
  SetupWorkflowParams,
  SetupWorkflowResult
} from './runThinktankTypes';

/**
 * Setup workflow helper function
 * 
 * Handles configuration loading, run name generation, and output directory creation
 * with proper error handling and spinner updates.
 * 
 * @param params - Parameters containing the spinner and options
 * @returns An object containing the configuration, run name, and output directory path
 * @throws 
 *   - ConfigError when config loading fails
 *   - FileSystemError when directory creation fails
 *   - PermissionError when permission issues occur
 */
export async function _setupWorkflow({ 
  spinner, 
  options 
}: SetupWorkflowParams): Promise<SetupWorkflowResult> {
  try {
    // 1. Load configuration
    spinner.text = 'Loading configuration...';
    const config = await loadConfig({ configPath: options.configPath });
    spinner.text = 'Configuration loaded successfully';
    
    // 2. Generate a friendly run name
    spinner.text = 'Generating run identifier...';
    const friendlyRunName = generateFunName();
    spinner.info(styleInfo(`Run name: ${styleSuccess(friendlyRunName)}`));
    spinner.start(); // Restart spinner for next step
    
    // 3. Create output directory
    spinner.text = 'Creating output directory...';
    // Determine directory identifier based on options
    const directoryIdentifier = options.specificModel || options.groupName;
    const outputDirectoryPath = await createOutputDirectory({
      outputDirectory: options.output,
      directoryIdentifier,
      friendlyRunName
    });
    spinner.info(styleInfo(`Output directory: ${outputDirectoryPath} (Run: ${friendlyRunName})`));
    spinner.start(); // Restart spinner for next step
    
    // Return the result object with all required properties
    return {
      config,
      friendlyRunName,
      outputDirectoryPath
    };
  } catch (error) {
    // Handle specific error types according to the error handling contract
    
    // If it's already a ConfigError, just rethrow it
    if (error instanceof ConfigError) {
      throw error;
    }
    
    // If it's a FileSystemError or PermissionError, rethrow it
    if (error instanceof FileSystemError || error instanceof PermissionError) {
      throw error;
    }
    
    // Handle NodeJS.ErrnoException for file system errors
    if (
      error instanceof Error && 
      'code' in error && 
      typeof (error as NodeJS.ErrnoException).code === 'string'
    ) {
      const nodeError = error as NodeJS.ErrnoException;
      
      // Permission errors
      if (nodeError.code === 'EACCES' || nodeError.code === 'EPERM') {
        throw new PermissionError(`Permission denied: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check that you have sufficient permissions for the directory',
            'Try specifying a different output directory with --output'
          ]
        });
      }
      
      // Directory/file not found
      if (nodeError.code === 'ENOENT') {
        throw new FileSystemError(`File or directory not found: ${error.message}`, {
          cause: error,
          suggestions: [
            'Check that the file exists at the specified path',
            `Current working directory: ${process.cwd()}`
          ]
        });
      }
      
      // Other file system errors
      throw new FileSystemError(`File system error: ${error.message}`, {
        cause: error,
        suggestions: [
          'Check disk space and permissions',
          'Verify the path is valid'
        ]
      });
    }
    
    // For config-related errors during setup, wrap in ConfigError
    if (spinner.text.includes('Loading configuration') || spinner.text.includes('configuration')) {
      throw new ConfigError(`Configuration error: ${error instanceof Error ? error.message : String(error)}`, {
        cause: error instanceof Error ? error : undefined,
        suggestions: [
          'Check that your configuration file is valid JSON',
          'Verify the configuration path is correct'
        ]
      });
    }
    
    // For directory creation errors, wrap in FileSystemError
    if (spinner.text.includes('Creating output directory') || spinner.text.includes('directory')) {
      throw new FileSystemError(`Error creating output directory: ${error instanceof Error ? error.message : String(error)}`, {
        cause: error instanceof Error ? error : undefined,
        suggestions: [
          'Check that the parent directory exists and is writable',
          'Verify that there is sufficient disk space'
        ]
      });
    }
    
    // Generic ThinktankError for other cases
    throw new ThinktankError(`Error during workflow setup: ${error instanceof Error ? error.message : String(error)}`, {
      cause: error instanceof Error ? error : undefined,
      category: errorCategories.UNKNOWN
    });
  }
}