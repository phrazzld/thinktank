/**
 * Adapter implementation of the FileSystem interface
 * Wraps existing file operations from fileReader.ts
 */
// Core Node.js modules
import fs from 'fs/promises';
import { Stats } from 'fs';
import { FileSystem } from './interfaces';
import * as fileReader from '../utils/fileReader';
import { FileReadError, ReadFileOptions } from '../utils/fileReaderTypes';

/**
 * FileSystemAdapter implementation of the FileSystem interface that wraps
 * existing file operations from fileReader.ts where available,
 * and falls back to direct fs/promises calls as needed.
 */
export class FileSystemAdapter implements FileSystem {
  /**
   * Reads the content of a file
   * @param filePath - The path to the file
   * @param options - Optional settings
   * @returns Promise resolving to the file content
   */
  async readFileContent(filePath: string, options?: ReadFileOptions): Promise<string> {
    return fileReader.readFileContent(filePath, options);
  }

  /**
   * Writes content to a file, creating directories if necessary
   * @param filePath - The path to the file
   * @param content - The content to write
   * @returns Promise resolving when the write is complete
   */
  async writeFile(filePath: string, content: string): Promise<void> {
    return fileReader.writeFile(filePath, content);
  }

  /**
   * Checks if a path exists
   * @param path - The path to check
   * @returns Promise resolving to true if the path exists, false otherwise
   */
  async fileExists(path: string): Promise<boolean> {
    return fileReader.fileExists(path);
  }

  /**
   * Creates a directory, including parent directories if needed
   * @param dirPath - The path to the directory to create
   * @param options - Options, e.g., { recursive?: boolean }
   * @returns Promise resolving when the directory is created
   */
  async mkdir(dirPath: string, options?: { recursive?: boolean }): Promise<void> {
    try {
      await fs.mkdir(dirPath, options);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'EPERM' || nodeError.code === 'EACCES') {
          throw new FileReadError(`Permission denied creating directory: ${dirPath}`, error);
        } else if (nodeError.code === 'EEXIST' && options?.recursive) {
          // Directory already exists and recursive is true, this is fine
          return;
        }
        
        throw new FileReadError(`Failed to create directory: ${dirPath}`, error);
      }
      
      throw new FileReadError(`Failed to create directory: ${dirPath}`);
    }
  }

  /**
   * Reads the names of entries in a directory
   * @param dirPath - The path to the directory
   * @returns Promise resolving to an array of entry names
   */
  async readdir(dirPath: string): Promise<string[]> {
    try {
      return await fs.readdir(dirPath);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw new FileReadError(`Directory not found: ${dirPath}`, error);
        } else if (nodeError.code === 'EACCES') {
          throw new FileReadError(`Permission denied reading directory: ${dirPath}`, error);
        }
        
        throw new FileReadError(`Failed to read directory: ${dirPath}`, error);
      }
      
      throw new FileReadError(`Failed to read directory: ${dirPath}`);
    }
  }

  /**
   * Gets statistics for a file or directory path
   * @param path - The path to get stats for
   * @returns Promise resolving to a fs.Stats object
   */
  async stat(path: string): Promise<Stats> {
    try {
      return await fs.stat(path);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw new FileReadError(`Path not found: ${path}`, error);
        } else if (nodeError.code === 'EACCES') {
          throw new FileReadError(`Permission denied accessing path: ${path}`, error);
        }
        
        throw new FileReadError(`Failed to get stats for path: ${path}`, error);
      }
      
      throw new FileReadError(`Failed to get stats for path: ${path}`);
    }
  }

  /**
   * Tests a user's permissions for accessing a file
   * @param path - The path to check
   * @param mode - Optional mode to check
   * @returns Promise that resolves if access is allowed, rejects otherwise
   */
  async access(path: string, mode?: number): Promise<void> {
    try {
      await fs.access(path, mode);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw new FileReadError(`Path not found: ${path}`, error);
        } else if (nodeError.code === 'EACCES') {
          throw new FileReadError(`Permission denied accessing path: ${path}`, error);
        }
        
        throw new FileReadError(`Failed to access path: ${path}`, error);
      }
      
      throw new FileReadError(`Failed to access path: ${path}`);
    }
  }

  /**
   * Gets the path to the application's config directory
   * @returns Promise resolving to the config directory path
   */
  async getConfigDir(): Promise<string> {
    return fileReader.getConfigDir();
  }

  /**
   * Gets the path to the application's config file
   * @returns Promise resolving to the config file path
   */
  async getConfigFilePath(): Promise<string> {
    return fileReader.getConfigFilePath();
  }
}