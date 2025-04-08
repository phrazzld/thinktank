/**
 * Interface definitions for external dependencies used within thinktank.
 * 
 * These interfaces define the contracts for interacting with external systems
 * like LLM APIs, the filesystem, and configuration management, facilitating
 * dependency injection and testability.
 */

import type { Stats } from 'fs';
import type { 
  LLMResponse, 
  ModelOptions, 
  SystemPrompt, 
  AppConfig,
  ModelConfig, 
  ModelGroup
} from './types';
import type { LoadConfigOptions } from './configManager';

// --- Filesystem Interface ---

/**
 * Interface abstracting file system operations.
 * This allows swapping the real filesystem with a virtual one for testing.
 */
export interface FileSystem {
  /**
   * Reads the content of a file.
   * @param filePath - The path to the file.
   * @param options - Optional settings like whether to normalize the content
   * @returns A promise resolving to the file content as a string.
   * @throws {FileSystemError} If reading fails (e.g., not found, permission denied).
   */
  readFileContent(filePath: string, options?: { normalize?: boolean }): Promise<string>;

  /**
   * Writes content to a file, creating directories if necessary.
   * @param filePath - The path to the file.
   * @param content - The content to write.
   * @returns A promise resolving when the write is complete.
   * @throws {FileSystemError} If writing fails.
   */
  writeFile(filePath: string, content: string): Promise<void>;

  /**
   * Checks if a path exists.
   * @param path - The path to check.
   * @returns A promise resolving to true if the path exists, false otherwise.
   */
  fileExists(path: string): Promise<boolean>;

  /**
   * Creates a directory, including parent directories if needed.
   * @param dirPath - The path to the directory to create.
   * @param options - Options, e.g., { recursive?: boolean }.
   * @returns A promise resolving when the directory is created.
   * @throws {FileSystemError} If creation fails.
   */
  mkdir(dirPath: string, options?: { recursive?: boolean }): Promise<void>;
  
  /**
   * Reads the names of entries in a directory.
   * @param dirPath - The path to the directory.
   * @returns A promise resolving to an array of entry names (files and directories).
   * @throws {FileSystemError} If reading fails.
   */
  readdir(dirPath: string): Promise<string[]>;

  /**
   * Gets statistics for a file or directory path.
   * @param path - The path to get stats for.
   * @returns A promise resolving to a fs.Stats object.
   * @throws {FileSystemError} If getting stats fails.
   */
  stat(path: string): Promise<Stats>;

  /**
   * Tests a user's permissions for accessing a file.
   * @param path - The path to check.
   * @param mode - Optional mode to check (e.g., fs.constants.R_OK).
   * @returns A promise that resolves if access is allowed, rejects otherwise.
   */
  access(path: string, mode?: number): Promise<void>;

  /**
   * Gets the path to the application's config directory.
   * @returns A promise resolving to the config directory path.
   */
  getConfigDir(): Promise<string>;

  /**
   * Gets the path to the application's config file.
   * @returns A promise resolving to the config file path.
   */
  getConfigFilePath(): Promise<string>;
}

// --- LLM Client Interface ---

/**
 * Interface abstracting interactions with a Large Language Model provider API.
 * 
 * Note: This interface is similar to the existing LLMProvider interface in types.ts
 * We're creating a separate interface for abstraction purposes, but with the intention
 * that concrete implementations may satisfy both interfaces.
 */
export interface LLMClient {
  /**
   * Generates a response from the language model.
   * @param prompt - The main user prompt.
   * @param modelId - The specific model identifier to use.
   * @param options - Optional parameters for the generation (e.g., temperature, maxTokens).
   * @param systemPrompt - Optional system prompt to guide the model's behavior.
   * @returns A promise resolving to the standardized LLMResponse.
   * @throws {ApiError} If the API call fails.
   */
  generate(
    prompt: string,
    modelId: string,
    options?: ModelOptions,
    systemPrompt?: SystemPrompt
  ): Promise<LLMResponse>;
}

// --- Configuration Manager Interface ---

/**
 * Interface abstracting configuration management operations.
 * Handles loading, saving, and retrieving configuration paths.
 */
export interface ConfigManagerInterface {
  /**
   * Loads the application configuration.
   * Handles finding the correct config file (XDG, custom path), validation,
   * and default creation if necessary.
   * @param options - Options for loading, such as a specific config path.
   * @returns A promise resolving to the loaded and validated AppConfig.
   * @throws {ConfigError} If loading or validation fails.
   */
  loadConfig(options?: LoadConfigOptions): Promise<AppConfig>;

  /**
   * Saves the application configuration to the appropriate location.
   * Handles validation before saving.
   * @param config - The configuration object to save.
   * @param configPath - Optional specific path to save to (otherwise uses the standard path).
   * @returns A promise resolving when the save is complete.
   * @throws {ConfigError} If saving fails or the config is invalid.
   */
  saveConfig(config: AppConfig, configPath?: string): Promise<void>;

  /**
   * Gets the active configuration file path (typically the XDG path).
   * @returns A promise resolving to the absolute path of the active config file.
   * @throws {ConfigError} If the config directory cannot be determined or accessed.
   */
  getActiveConfigPath(): Promise<string>;

  /**
   * Gets the default project-local configuration file path (e.g., ./thinktank.config.json).
   * @returns The absolute path to the default project-local config file.
   */
  getDefaultConfigPath(): string;
}