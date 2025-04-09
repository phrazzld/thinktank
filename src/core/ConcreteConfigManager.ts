/**
 * Concrete implementation of the ConfigManagerInterface.
 *
 * This module provides a concrete implementation of the ConfigManagerInterface
 * that wraps the existing configManager functionality.
 */

import { ConfigManagerInterface } from './interfaces';
import { AppConfig, ModelConfig, ModelGroup } from './types';
import * as configFns from './configManager';
import type { LoadConfigOptions } from './configManager';
import { ConfigError, ThinktankError } from './errors';

/**
 * Concrete implementation of the ConfigManagerInterface that wraps
 * the existing configManager functionality.
 */
export class ConcreteConfigManager implements ConfigManagerInterface {
  /**
   * Loads the application configuration.
   *
   * @param options - Options for loading, such as a specific config path
   * @returns Promise resolving to the loaded and validated AppConfig
   * @throws {ConfigError} If loading or validation fails
   */
  async loadConfig(options?: LoadConfigOptions): Promise<AppConfig> {
    try {
      // Directly delegate to the existing function
      return await configFns.loadConfig(options);
    } catch (error) {
      // Re-throw known Thinktank errors
      if (error instanceof ThinktankError) {
        throw error;
      }

      // Wrap other errors with context
      const message = `Failed to load configuration: ${error instanceof Error ? error.message : String(error)}`;
      throw new ConfigError(message, { cause: error instanceof Error ? error : undefined });
    }
  }

  /**
   * Saves the application configuration to the appropriate location.
   *
   * @param config - The configuration object to save
   * @param configPath - Optional specific path to save to
   * @returns Promise resolving when the save is complete
   * @throws {ConfigError} If saving fails or the config is invalid
   */
  async saveConfig(config: AppConfig, configPath?: string): Promise<void> {
    try {
      // Directly delegate to the existing function
      await configFns.saveConfig(config, configPath);
    } catch (error) {
      // Re-throw known Thinktank errors
      if (error instanceof ThinktankError) {
        throw error;
      }

      // Wrap other errors with context
      const message = `Failed to save configuration: ${error instanceof Error ? error.message : String(error)}`;
      throw new ConfigError(message, { cause: error instanceof Error ? error : undefined });
    }
  }

  /**
   * Gets the active configuration file path (typically the XDG path).
   *
   * @returns Promise resolving to the absolute path of the active config file
   * @throws {ConfigError} If the config directory cannot be determined or accessed
   */
  async getActiveConfigPath(): Promise<string> {
    try {
      // Directly delegate to the existing function
      return await configFns.getActiveConfigPath();
    } catch (error) {
      // Re-throw known Thinktank errors
      if (error instanceof ThinktankError) {
        throw error;
      }

      // Wrap other errors with context
      const message = `Failed to get active config path: ${error instanceof Error ? error.message : String(error)}`;
      throw new ConfigError(message, { cause: error instanceof Error ? error : undefined });
    }
  }

  /**
   * Gets the default project-local configuration file path.
   *
   * @returns The absolute path to the default project-local config file
   */
  getDefaultConfigPath(): string {
    // Directly delegate to the existing function
    return configFns.getDefaultConfigPath();
  }

  /**
   * Adds or updates a model in the configuration.
   *
   * @param config - The application configuration to modify
   * @param model - The model to add or update
   * @returns The updated configuration
   */
  addOrUpdateModel(config: AppConfig, model: ModelConfig): AppConfig {
    return configFns.addOrUpdateModel(config, model);
  }

  /**
   * Removes a model from the configuration.
   *
   * @param config - The application configuration to modify
   * @param provider - The provider ID
   * @param modelId - The model ID
   * @returns The updated configuration
   */
  removeModel(config: AppConfig, provider: string, modelId: string): AppConfig {
    return configFns.removeModel(config, provider, modelId);
  }

  /**
   * Adds or updates a group in the configuration.
   *
   * @param config - The application configuration to modify
   * @param groupName - The name of the group
   * @param groupDetails - The group details
   * @returns The updated configuration
   */
  addOrUpdateGroup(
    config: AppConfig,
    groupName: string,
    groupDetails: Partial<Omit<ModelGroup, 'name'>>
  ): AppConfig {
    return configFns.addOrUpdateGroup(config, groupName, groupDetails);
  }

  /**
   * Removes a group from the configuration.
   *
   * @param config - The application configuration to modify
   * @param groupName - The name of the group to remove
   * @returns The updated configuration
   */
  removeGroup(config: AppConfig, groupName: string): AppConfig {
    return configFns.removeGroup(config, groupName);
  }

  /**
   * Adds a model to a group.
   *
   * @param config - The application configuration to modify
   * @param groupName - The name of the group
   * @param provider - The provider ID
   * @param modelId - The model ID
   * @returns The updated configuration
   */
  addModelToGroup(
    config: AppConfig,
    groupName: string,
    provider: string,
    modelId: string
  ): AppConfig {
    return configFns.addModelToGroup(config, groupName, provider, modelId);
  }

  /**
   * Removes a model from a group.
   *
   * @param config - The application configuration to modify
   * @param groupName - The name of the group
   * @param provider - The provider ID
   * @param modelId - The model ID
   * @returns The updated configuration
   */
  removeModelFromGroup(
    config: AppConfig,
    groupName: string,
    provider: string,
    modelId: string
  ): AppConfig {
    return configFns.removeModelFromGroup(config, groupName, provider, modelId);
  }
}
