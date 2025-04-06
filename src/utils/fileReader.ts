/**
 * File reader module for handling prompt file input and configuration files
 */
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { normalizeText } from './helpers';
import { shouldIgnorePath } from './gitignoreUtils';

/**
 * Custom error for file reading operations
 */
export class FileReadError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'FileReadError';
  }
}

/**
 * Result of reading a context file
 */
export interface ContextFileResult {
  /**
   * Path to the file that was read (original path provided)
   */
  path: string;
  
  /**
   * Content of the file, or null if there was an error
   */
  content: string | null;
  
  /**
   * Error information if reading failed, or null if successful
   */
  error: {
    /**
     * Error code (e.g., 'ENOENT', 'EACCES', 'NOT_FILE', etc.)
     */
    code: string;
    
    /**
     * Human-readable error message
     */
    message: string;
  } | null;
}

/**
 * Application name used for XDG paths
 */
const APP_NAME = 'thinktank';

/**
 * Options for reading file content
 */
export interface ReadFileOptions {
  normalize?: boolean;
}

/**
 * Reads content from a file at the specified path
 * 
 * @param filePath - Path to the file to read
 * @param options - Options for reading the file
 * @returns The file content as a string
 * @throws {FileReadError} If the file cannot be read
 */
export async function readFileContent(
  filePath: string,
  options: ReadFileOptions = {}
): Promise<string> {
  const { normalize = true } = options;
  
  try {
    // Resolve to absolute path if relative path is provided
    const resolvedPath = path.isAbsolute(filePath) 
      ? filePath 
      : path.resolve(process.cwd(), filePath);
    
    // Check if file exists and is readable
    await fs.access(resolvedPath, fs.constants.R_OK);
    
    // Read file content
    const content = await fs.readFile(resolvedPath, 'utf-8');
    
    // Return normalized content if requested, otherwise raw content
    return normalize ? normalizeText(content) : content;
  } catch (error) {
    // Handle specific error types
    if (error instanceof Error) {
      if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
        throw new FileReadError(`File not found: ${filePath}`, error);
      } else if ((error as NodeJS.ErrnoException).code === 'EACCES') {
        throw new FileReadError(`Permission denied to read file: ${filePath}`, error);
      }
      throw new FileReadError(`Error reading file: ${filePath}`, error);
    }
    
    // Generic error case
    throw new FileReadError(`Unknown error reading file: ${filePath}`);
  }
}

/**
 * Checks if a file exists at the specified path
 * 
 * @param filePath - Path to check
 * @returns True if the file exists, false otherwise
 */
export async function fileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath, fs.constants.F_OK);
    return true;
  } catch {
    return false;
  }
}

/**
 * Writes content to a file
 * 
 * @param filePath - Path to the file to write
 * @param content - Content to write to the file
 * @throws {FileReadError} If writing fails
 */
export async function writeFile(filePath: string, content: string): Promise<void> {
  try {
    // Ensure the directory exists
    const dir = path.dirname(filePath);
    await fs.mkdir(dir, { recursive: true });
    
    // Write the file
    await fs.writeFile(filePath, content, { encoding: 'utf-8' });
  } catch (error) {
    if (error instanceof Error) {
      // Cast to get access to error codes
      const errnoError = error as NodeJS.ErrnoException;
      
      // Handle Windows-specific write errors
      if (process.platform === 'win32') {
        if (errnoError.code === 'EPERM' || errnoError.code === 'EACCES') {
          throw new FileReadError(
            `Permission denied writing file at ${filePath}. On Windows, this may happen if the file is read-only or in use by another process.`,
            error
          );
        }
        
        if (errnoError.code === 'ENOENT') {
          throw new FileReadError(
            `Failed to write file at ${filePath}: The directory path may not exist or may not be accessible.`,
            error
          );
        }
        
        if (errnoError.code === 'EBUSY') {
          throw new FileReadError(
            `Failed to write file at ${filePath}: The file is in use by another process. Close any programs that might be using this file.`,
            error
          );
        }
      }
      
      // Handle macOS-specific write errors
      if (process.platform === 'darwin') {
        if (errnoError.code === 'EPERM' || errnoError.code === 'EACCES') {
          throw new FileReadError(
            `Permission denied writing file at ${filePath}. On macOS, this may happen if System Integrity Protection (SIP) is restricting access or if the file permissions are incorrect. Check folder permissions.`,
            error
          );
        }
        
        if (errnoError.code === 'ENOENT') {
          throw new FileReadError(
            `Failed to write file at ${filePath}: The parent directory may not exist or may not be accessible on macOS.`,
            error
          );
        }
        
        if (errnoError.code === 'EROFS') {
          throw new FileReadError(
            `Failed to write file at ${filePath}: The file system is read-only. On macOS, check if your disk is mounted with write permissions.`,
            error
          );
        }
        
        if (errnoError.code === 'EMFILE') {
          throw new FileReadError(
            `Failed to write file at ${filePath}: Too many open files. On macOS, you may need to increase the open file limit or close some applications.`,
            error
          );
        }
      }
      
      // Generic error for all platforms
      throw new FileReadError(`Failed to write file at ${filePath}: ${error.message}`, error);
    }
    
    throw new FileReadError(`Unknown error writing file at ${filePath}`);
  }
}

/**
 * Gets the config directory path following platform-specific conventions:
 * 
 * - Windows: '%APPDATA%\thinktank' or '%USERPROFILE%\AppData\Roaming\thinktank'
 *   - Uses APPDATA environment variable if available
 *   - Falls back to homedir/AppData/Roaming if APPDATA is not set
 * 
 * - Unix-like systems (Linux, macOS): '~/.config/thinktank'
 *   - Uses $XDG_CONFIG_HOME/thinktank if XDG_CONFIG_HOME environment variable is set
 *   - Falls back to ~/.config/thinktank for both Linux and macOS for consistency
 * 
 * Note: While macOS traditionally uses ~/Library/Preferences for app configs,
 * we've chosen to use ~/.config for consistency across Unix-like platforms.
 * 
 * @returns Promise resolving to the platform-specific config directory path
 * @throws {FileReadError} If directory creation fails with platform-specific error details
 */
export async function getConfigDir(): Promise<string> {
  try {
    let configDir: string;
    
    // Check for XDG_CONFIG_HOME environment variable (Linux/Unix/macOS)
    if (process.env.XDG_CONFIG_HOME && process.env.XDG_CONFIG_HOME.trim() !== '') {
      configDir = path.join(process.env.XDG_CONFIG_HOME, APP_NAME);
    } 
    // Windows: %APPDATA%\thinktank or %USERPROFILE%\AppData\Roaming\thinktank
    else if (process.platform === 'win32') {
      // Check for the APPDATA environment variable
      const appDataEnv = process.env.APPDATA;
      
      // Use APPDATA if it exists and is not empty, otherwise construct from homedir
      if (appDataEnv && appDataEnv.trim() !== '') {
        configDir = path.join(appDataEnv, APP_NAME);
      } else {
        // Use the standard Windows AppData location as fallback
        configDir = path.join(os.homedir(), 'AppData', 'Roaming', APP_NAME);
      }
    }
    // Unix-like systems (Linux, macOS): ~/.config/thinktank
    else {
      const homeDir = os.homedir();
      
      // Ensure the home directory path is valid
      if (!homeDir || homeDir.trim() === '') {
        throw new FileReadError('Unable to determine home directory. Check user environment.');
      }
      
      // Use the same path for both Linux and macOS for consistency
      configDir = path.join(homeDir, '.config', APP_NAME);
    }
    
    // Ensure the directory exists
    await fs.mkdir(configDir, { recursive: true });
    
    return configDir;
  } catch (error) {
    if (error instanceof Error) {
      const errnoError = error as NodeJS.ErrnoException;
      
      // Handle Windows-specific error codes
      if (process.platform === 'win32') {
        if (errnoError.code === 'EPERM' || errnoError.code === 'EACCES') {
          throw new FileReadError(
            `Permission denied creating config directory. On Windows, this may happen if you don't have administrator rights or the folder is read-only. Try running the application with administrative privileges.`,
            error
          );
        }
        
        if (errnoError.code === 'ENOENT') {
          throw new FileReadError(
            `Unable to create configuration directory. The AppData folder may not exist or may not be accessible.`,
            error
          );
        }
      }
      
      // Handle macOS-specific error codes
      if (process.platform === 'darwin') {
        if (errnoError.code === 'EPERM' || errnoError.code === 'EACCES') {
          throw new FileReadError(
            `Permission denied creating config directory. On macOS, this may happen if you don't have write access to ~/Library/Preferences or if System Integrity Protection (SIP) is restricting access. Check folder permissions.`,
            error
          );
        }
        
        if (errnoError.code === 'ENOENT') {
          throw new FileReadError(
            `Unable to create configuration directory. The Library/Preferences path may not exist or may not be accessible. Check if your user account is properly set up.`,
            error
          );
        }
        
        if (errnoError.code === 'EROFS') {
          throw new FileReadError(
            `Configuration directory is on a read-only file system. On macOS, this may happen if your disk is mounted read-only or if System Integrity Protection is restricting write access.`,
            error
          );
        }
      }
      
      // Generic error with message for all platforms
      throw new FileReadError(
        `Failed to create or access config directory: ${error.message}`, 
        error
      );
    }
    throw new FileReadError('Unknown error accessing config directory');
  }
}

/**
 * Gets the full path to the configuration file
 * 
 * @returns Promise resolving to the full config file path
 * @throws {FileReadError} If directory creation fails
 */
export async function getConfigFilePath(): Promise<string> {
  const configDir = await getConfigDir();
  return path.join(configDir, 'config.json');
}

/**
 * Reads content from a file for use as context in prompts
 * Instead of throwing errors, returns an object with path, content, and error information
 * 
 * @param filePath - Path to the file to read
 * @returns Promise resolving to an object containing the file path, content, and any error information
 */
export async function readContextFile(filePath: string): Promise<ContextFileResult> {
  // Initialize the result with the provided path
  const result: ContextFileResult = {
    path: filePath,
    content: null,
    error: null
  };
  
  try {
    // Resolve to absolute path if relative path is provided
    const resolvedPath = path.isAbsolute(filePath) 
      ? filePath 
      : path.resolve(process.cwd(), filePath);
    
    // Check if file exists and is readable
    await fs.access(resolvedPath, fs.constants.R_OK);
    
    // Check if path is a file (not a directory or other non-file)
    const stats = await fs.stat(resolvedPath);
    if (!stats.isFile()) {
      return {
        ...result,
        error: {
          code: 'NOT_FILE',
          message: `Path is not a file: ${filePath}`
        }
      };
    }
    
    // Read file content
    const content = await fs.readFile(resolvedPath, 'utf-8');
    
    // Return successful result with content
    return {
      ...result,
      content,
      error: null
    };
  } catch (error) {
    // Handle specific error types
    if (error instanceof Error) {
      const errnoError = error as NodeJS.ErrnoException;
      
      if (errnoError.code === 'ENOENT') {
        return {
          ...result,
          error: {
            code: 'ENOENT',
            message: `File not found: ${filePath}`
          }
        };
      } else if (errnoError.code === 'EACCES') {
        return {
          ...result,
          error: {
            code: 'EACCES',
            message: `Permission denied to read file: ${filePath}`
          }
        };
      }
      
      // Generic error case with Error object
      return {
        ...result,
        error: {
          code: 'READ_ERROR',
          message: `Error reading file: ${filePath}`
        }
      };
    }
    
    // Unknown error type
    return {
      ...result,
      error: {
        code: 'UNKNOWN',
        message: `Unknown error reading file: ${filePath}`
      }
    };
  }
}

/**
 * Fallback directories to ignore during traversal when .gitignore is not available
 * Also serves as a safety net for critical directories
 */
const DEFAULT_IGNORED_DIRECTORIES = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'coverage',
  '.cache',
  '.next',
  '.nuxt',
  '.output',
  '.vscode',
  '.idea'
];

/**
 * Recursively reads all files in a directory and its subdirectories
 * 
 * @param dirPath - Path to the directory to read
 * @returns Promise resolving to an array of ContextFileResult objects
 */
export async function readDirectoryContents(dirPath: string): Promise<ContextFileResult[]> {
  const results: ContextFileResult[] = [];
  
  try {
    // Resolve to absolute path if relative path is provided
    const resolvedPath = path.isAbsolute(dirPath) 
      ? dirPath 
      : path.resolve(process.cwd(), dirPath);
    
    // Check if directory exists and is readable
    await fs.access(resolvedPath, fs.constants.R_OK);
    
    // Check if path is a directory
    const stats = await fs.stat(resolvedPath);
    if (!stats.isDirectory()) {
      // If it's a file, just read it and return the result
      const fileResult = await readContextFile(dirPath);
      return [fileResult];
    }
    
    // Read directory contents
    const entries = await fs.readdir(resolvedPath);
    
    // Process each entry
    for (const entry of entries) {
      const entryPath = path.join(dirPath, entry);
      
      try {
        const entryStats = await fs.stat(entryPath);
        
        if (entryStats.isFile()) {
          // Check if the file should be ignored based on gitignore rules
          if (await shouldIgnorePath(dirPath, entryPath)) {
            continue;
          }
          
          // If not ignored, read the file and add to results
          const fileResult = await readContextFile(entryPath);
          results.push(fileResult);
        } else if (entryStats.isDirectory()) {
          // Always skip certain critical directories regardless of gitignore rules
          if (DEFAULT_IGNORED_DIRECTORIES.includes(entry)) {
            continue;
          }
          
          // Check if the directory should be ignored based on gitignore rules
          if (await shouldIgnorePath(dirPath, entryPath)) {
            continue;
          }
          
          // If it's a directory and not ignored, recursively read its contents
          const subdirResults = await readDirectoryContents(entryPath);
          results.push(...subdirResults);
        }
        // Skip other types (symlinks, etc.)
      } catch (error) {
        // If we can't process an entry, add an error result for it
        results.push({
          path: entryPath,
          content: null,
          error: {
            code: 'READ_ERROR',
            message: `Error processing directory entry: ${entryPath}`
          }
        });
      }
    }
    
    return results;
  } catch (error) {
    // Handle directory access errors
    if (error instanceof Error) {
      return [{
        path: dirPath,
        content: null,
        error: {
          code: 'READ_ERROR',
          message: `Error reading directory: ${dirPath} - ${error.message}`
        }
      }];
    }
    
    // Unknown error type
    return [{
      path: dirPath,
      content: null,
      error: {
        code: 'UNKNOWN',
        message: `Unknown error reading directory: ${dirPath}`
      }
    }];
  }
}