/**
 * Concrete implementation of the FileSystem interface
 * Provides a consistent abstraction over filesystem operations 
 * by wrapping existing file operations from fileReader.ts where available,
 * and implementing direct fs/promises calls where needed.
 */
// Core Node.js modules
import fs from 'fs/promises';
import { Stats } from 'fs';
import { FileSystem } from './interfaces';
import * as fileReader from '../utils/fileReader';
import { ReadFileOptions } from '../utils/fileReaderTypes';
import { FileSystemError } from './errors/types/filesystem';
import { createFileNotFoundError } from './errors/factories/filesystem';

/**
 * ConcreteFileSystem implements the FileSystem interface to provide
 * filesystem operations with consistent error handling.
 * 
 * This class serves as the primary concrete implementation of the FileSystem
 * interface, allowing other components to interact with the filesystem through
 * dependency injection, improving testability.
 */
export class ConcreteFileSystem implements FileSystem {
  /**
   * Reads the content of a file
   * 
   * @param filePath - The path to the file
   * @param options - Optional settings like whether to normalize content
   * @returns Promise resolving to the file content as a string
   * @throws {FileSystemError} If the file cannot be read (not found, permission denied, etc.)
   */
  async readFileContent(filePath: string, options?: ReadFileOptions): Promise<string> {
    try {
      return await fileReader.readFileContent(filePath, options);
    } catch (error) {
      // Convert FileReadError to FileSystemError with appropriate metadata
      if (error instanceof Error) {
        const errorMsg = error.message;
        
        // Handle common error cases with specialized error messages
        if (errorMsg.includes('not found')) {
          throw createFileNotFoundError(filePath);
        } else if (errorMsg.includes('Permission denied')) {
          throw new FileSystemError(`Permission denied reading file: ${filePath}`, {
            cause: error,
            filePath,
            suggestions: [
              'Check file permissions',
              'Ensure you have read access to the file and its directory',
              `Current working directory: ${process.cwd()}`
            ]
          });
        }
        
        // Generic file reading error
        throw new FileSystemError(`Error reading file: ${filePath}`, {
          cause: error,
          filePath
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Unknown error reading file: ${filePath}`, {
        filePath
      });
    }
  }

  /**
   * Writes content to a file, creating directories if necessary
   * 
   * @param filePath - The path to the file
   * @param content - The content to write
   * @returns Promise resolving when the write is complete
   * @throws {FileSystemError} If writing fails (permission denied, disk full, etc.)
   */
  async writeFile(filePath: string, content: string): Promise<void> {
    try {
      await fileReader.writeFile(filePath, content);
    } catch (error) {
      if (error instanceof Error) {
        const errorMsg = error.message;
        
        // Handle common error cases with specialized error messages
        if (errorMsg.includes('Permission denied')) {
          throw new FileSystemError(`Permission denied writing to file: ${filePath}`, {
            cause: error,
            filePath,
            suggestions: [
              'Check file and directory permissions',
              'Ensure you have write access to the specified location',
              `If on Windows, ensure the file is not open in another program`,
              `Current working directory: ${process.cwd()}`
            ]
          });
        } else if (errorMsg.includes('directory') && errorMsg.includes('not exist')) {
          throw new FileSystemError(`Cannot write file, parent directory does not exist: ${filePath}`, {
            cause: error,
            filePath,
            suggestions: [
              'Create the parent directory before writing the file',
              'Use the recursive option when creating the directory'
            ]
          });
        }
        
        // Generic write error
        throw new FileSystemError(`Failed to write file: ${filePath}`, {
          cause: error,
          filePath
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Unknown error writing file: ${filePath}`, {
        filePath
      });
    }
  }

  /**
   * Checks if a path exists
   * 
   * @param path - The path to check
   * @returns Promise resolving to true if the path exists, false otherwise
   */
  async fileExists(path: string): Promise<boolean> {
    try {
      return await fileReader.fileExists(path);
    } catch (error) {
      // This should rarely throw since fileExists is designed to return false
      // rather than throw for non-existent files, but handle it just in case
      throw new FileSystemError(`Error checking if file exists: ${path}`, {
        cause: error instanceof Error ? error : undefined,
        filePath: path
      });
    }
  }

  /**
   * Creates a directory, including parent directories if needed
   * 
   * @param dirPath - The path to the directory to create
   * @param options - Options, e.g., { recursive?: boolean }
   * @returns Promise resolving when the directory is created
   * @throws {FileSystemError} If directory creation fails
   */
  async mkdir(dirPath: string, options?: { recursive?: boolean }): Promise<void> {
    try {
      await fs.mkdir(dirPath, options);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'EPERM' || nodeError.code === 'EACCES') {
          throw new FileSystemError(`Permission denied creating directory: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'Check parent directory permissions',
              'Ensure you have write access to the parent directory',
              `Current working directory: ${process.cwd()}`
            ]
          });
        } else if (nodeError.code === 'EEXIST' && options?.recursive) {
          // Directory already exists and recursive is true, this is fine
          return;
        } else if (nodeError.code === 'EEXIST') {
          throw new FileSystemError(`Directory already exists: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'Use { recursive: true } option to ignore this error',
              'Check if you need to use a different directory name'
            ]
          });
        } else if (nodeError.code === 'ENOENT') {
          throw new FileSystemError(`Cannot create directory, parent directory does not exist: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'Use { recursive: true } option to create parent directories',
              'Create parent directories first'
            ]
          });
        }
        
        // Generic error
        throw new FileSystemError(`Failed to create directory: ${dirPath}`, {
          cause: error,
          filePath: dirPath
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Failed to create directory: ${dirPath}`, {
        filePath: dirPath
      });
    }
  }

  /**
   * Reads the names of entries in a directory
   * 
   * @param dirPath - The path to the directory
   * @returns Promise resolving to an array of entry names
   * @throws {FileSystemError} If directory reading fails
   */
  async readdir(dirPath: string): Promise<string[]> {
    try {
      return await fs.readdir(dirPath);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw new FileSystemError(`Directory not found: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'Check the directory path is correct',
              'Ensure the directory exists before trying to read it',
              `Current working directory: ${process.cwd()}`
            ]
          });
        } else if (nodeError.code === 'EACCES') {
          throw new FileSystemError(`Permission denied reading directory: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'Check directory permissions',
              'Ensure you have read access to the directory'
            ]
          });
        } else if (nodeError.code === 'ENOTDIR') {
          throw new FileSystemError(`Not a directory: ${dirPath}`, {
            cause: error,
            filePath: dirPath,
            suggestions: [
              'The specified path exists but is not a directory',
              'Check if you meant to use a file reading operation instead'
            ]
          });
        }
        
        // Generic error
        throw new FileSystemError(`Failed to read directory: ${dirPath}`, {
          cause: error,
          filePath: dirPath
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Failed to read directory: ${dirPath}`, {
        filePath: dirPath
      });
    }
  }

  /**
   * Gets statistics for a file or directory path
   * 
   * @param path - The path to get stats for
   * @returns Promise resolving to a fs.Stats object
   * @throws {FileSystemError} If stat operation fails
   */
  async stat(path: string): Promise<Stats> {
    try {
      return await fs.stat(path);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw createFileNotFoundError(path);
        } else if (nodeError.code === 'EACCES') {
          throw new FileSystemError(`Permission denied accessing path: ${path}`, {
            cause: error,
            filePath: path,
            suggestions: [
              'Check file and directory permissions',
              'Ensure you have sufficient permissions to access the path'
            ]
          });
        } else if (nodeError.code === 'ELOOP') {
          throw new FileSystemError(`Too many symbolic links encountered: ${path}`, {
            cause: error,
            filePath: path,
            suggestions: [
              'Check for circular symbolic links',
              'Ensure the path does not contain symbolic link loops'
            ]
          });
        }
        
        // Generic error
        throw new FileSystemError(`Failed to get stats for path: ${path}`, {
          cause: error,
          filePath: path
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Failed to get stats for path: ${path}`, {
        filePath: path
      });
    }
  }

  /**
   * Tests a user's permissions for accessing a file
   * 
   * @param path - The path to check
   * @param mode - Optional mode to check (e.g., fs.constants.R_OK)
   * @returns Promise that resolves if access is allowed, rejects otherwise
   * @throws {FileSystemError} If access check fails
   */
  async access(path: string, mode?: number): Promise<void> {
    try {
      await fs.access(path, mode);
    } catch (error) {
      if (error instanceof Error) {
        const nodeError = error as NodeJS.ErrnoException;
        
        if (nodeError.code === 'ENOENT') {
          throw createFileNotFoundError(path);
        } else if (nodeError.code === 'EACCES') {
          // Determine what kind of access was denied based on mode
          let accessType = 'accessing';
          if (mode) {
            // Use specific constants to avoid require()
            const R_OK = 4; // Read permission
            const W_OK = 2; // Write permission
            const X_OK = 1; // Execute permission
            
            if (mode & R_OK) accessType = 'reading';
            if (mode & W_OK) accessType = mode & R_OK ? 'reading/writing' : 'writing to';
            if (mode & X_OK) accessType = 'executing';
          }
          
          throw new FileSystemError(`Permission denied ${accessType} path: ${path}`, {
            cause: error,
            filePath: path,
            suggestions: [
              'Check file and directory permissions',
              `Ensure you have ${accessType} access to the specified path`,
              `Current working directory: ${process.cwd()}`
            ]
          });
        }
        
        // Generic error
        throw new FileSystemError(`Failed to access path: ${path}`, {
          cause: error,
          filePath: path
        });
      }
      
      // Unknown error type
      throw new FileSystemError(`Failed to access path: ${path}`, {
        filePath: path
      });
    }
  }

  /**
   * Gets the path to the application's config directory
   * 
   * @returns Promise resolving to the config directory path
   * @throws {FileSystemError} If config directory cannot be determined or created
   */
  async getConfigDir(): Promise<string> {
    try {
      return await fileReader.getConfigDir();
    } catch (error) {
      if (error instanceof Error) {
        throw new FileSystemError(`Failed to access or create config directory`, {
          cause: error,
          suggestions: [
            'Check permissions for the user home directory or XDG_CONFIG_HOME location',
            'Ensure the environment variables are set correctly',
            `If running in a container or restricted environment, check filesystem permissions`
          ]
        });
      }
      
      throw new FileSystemError('Failed to access or create config directory');
    }
  }

  /**
   * Gets the path to the application's config file
   * 
   * @returns Promise resolving to the config file path
   * @throws {FileSystemError} If config directory or file path cannot be determined
   */
  async getConfigFilePath(): Promise<string> {
    try {
      return await fileReader.getConfigFilePath();
    } catch (error) {
      if (error instanceof Error) {
        throw new FileSystemError(`Failed to determine config file path`, {
          cause: error,
          suggestions: [
            'Check permissions for the user home directory or XDG_CONFIG_HOME location',
            'Ensure the environment variables are set correctly'
          ]
        });
      }
      
      throw new FileSystemError('Failed to determine config file path');
    }
  }
}
