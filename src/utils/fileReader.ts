/**
 * File reader module for handling prompt file input and configuration files
 */
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { normalizeText } from './helpers';

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
      throw new FileReadError(`Failed to write file at ${filePath}: ${error.message}`, error);
    }
    
    throw new FileReadError(`Unknown error writing file at ${filePath}`);
  }
}

/**
 * Gets the XDG config directory path following the XDG Base Directory Specification
 * 
 * @returns Promise resolving to the platform-specific config directory path
 * @throws {FileReadError} If directory creation fails
 */
export async function getConfigDir(): Promise<string> {
  try {
    let configDir: string;
    
    // Check for XDG_CONFIG_HOME environment variable (Linux/Unix/macOS)
    if (process.env.XDG_CONFIG_HOME && process.env.XDG_CONFIG_HOME.trim() !== '') {
      configDir = path.join(process.env.XDG_CONFIG_HOME, APP_NAME);
    } 
    // Windows: %APPDATA%\thinktank
    else if (process.platform === 'win32') {
      configDir = path.join(process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming'), APP_NAME);
    }
    // macOS: ~/Library/Preferences/thinktank (unless XDG_CONFIG_HOME is set)
    else if (process.platform === 'darwin') {
      configDir = path.join(os.homedir(), 'Library', 'Preferences', APP_NAME);
    }
    // Linux/Unix/Default fallback: ~/.config/thinktank
    else {
      configDir = path.join(os.homedir(), '.config', APP_NAME);
    }
    
    // Ensure the directory exists
    await fs.mkdir(configDir, { recursive: true });
    
    return configDir;
  } catch (error) {
    if (error instanceof Error) {
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