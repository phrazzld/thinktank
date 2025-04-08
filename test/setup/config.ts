/**
 * Configuration setup utilities for tests
 * 
 * This module provides specialized setup helpers for configuration-related tests.
 */
import path from 'path';
import { setupBasicFs } from './fs';
import { normalizePathGeneral } from '../../src/utils/pathUtils';
import type { AppConfig } from '../../src/core/types';

/**
 * Creates a minimal application configuration for testing
 * 
 * @returns A minimal AppConfig object
 * 
 * Usage:
 * ```typescript
 * const config = createMinimalConfig();
 * ```
 */
export function createMinimalConfig(): AppConfig {
  return {
    models: [],
    groups: {}
  };
}

/**
 * Creates a realistic application configuration for testing
 * 
 * @param customConfig - Optional configuration overrides
 * @returns A populated AppConfig object
 * 
 * Usage:
 * ```typescript
 * const config = createRealisticConfig();
 * // or
 * const config = createRealisticConfig({ models: [{ id: 'custom:model' }] });
 * ```
 */
export function createRealisticConfig(customConfig: Partial<AppConfig> = {}): AppConfig {
  const defaultConfig: AppConfig = {
    models: [
      { provider: 'openai', modelId: 'gpt-4o', enabled: true },
      { provider: 'anthropic', modelId: 'claude-3-7-sonnet-20240229', enabled: true },
      { provider: 'google', modelId: 'gemini-pro', enabled: true }
    ],
    groups: {
      default: { 
        name: 'default',
        systemPrompt: { text: 'Default system prompt' },
        models: [{ provider: 'openai', modelId: 'gpt-4o', enabled: true }]
      },
      faves: { 
        name: 'faves',
        systemPrompt: { text: 'Favorite models prompt' },
        models: [
          { provider: 'openai', modelId: 'gpt-4o', enabled: true },
          { provider: 'anthropic', modelId: 'claude-3-7-sonnet-20240229', enabled: true },
          { provider: 'google', modelId: 'gemini-pro', enabled: true }
        ]
      }
    }
  };
  
  return {
    ...defaultConfig,
    ...customConfig,
    // Handle nested objects properly
    models: [...(customConfig.models || defaultConfig.models)],
    groups: { ...defaultConfig.groups, ...(customConfig.groups || {}) }
  };
}

/**
 * Creates a virtual configuration file for tests
 * 
 * @param configPath - Path where the config file should be created
 * @param config - Configuration object to serialize
 * 
 * Usage:
 * ```typescript
 * createVirtualConfigFile('/project/thinktank.config.json', createMinimalConfig());
 * ```
 */
export function createVirtualConfigFile(configPath: string, config: AppConfig): void {
  const normalizedPath = normalizePathGeneral(configPath, true);
  setupBasicFs({
    [normalizedPath]: JSON.stringify(config, null, 2)
  }, { reset: false });
}

/**
 * Sets up a test environment with a configuration file
 * 
 * @param baseDir - Base directory for the test environment
 * @param configName - Filename for the configuration file (defaults to thinktank.config.json)
 * @param config - Configuration object (defaults to minimal config)
 * @returns Object containing paths to created files
 * 
 * Usage:
 * ```typescript
 * const { configPath } = setupConfigTest('/project');
 * ```
 */
export function setupConfigTest(
  baseDir: string = '/project',
  configName: string = 'thinktank.config.json',
  config: AppConfig = createMinimalConfig()
): { configPath: string } {
  const normalizedBaseDir = normalizePathGeneral(baseDir, true);
  const configPath = path.join(normalizedBaseDir, configName);
  
  // Create directory structure and config file
  setupBasicFs({
    [configPath]: JSON.stringify(config, null, 2)
  });
  
  return { configPath };
}

/**
 * Sets up multiple configuration files for testing config precedence
 * 
 * @param baseDir - Base directory for the test environment
 * @returns Object containing paths to created files
 * 
 * Usage:
 * ```typescript
 * const { userConfigPath, defaultConfigPath } = setupConfigPrecedenceTest();
 * ```
 */
export function setupConfigPrecedenceTest(baseDir: string = '/project'): {
  defaultConfigPath: string;
  userConfigPath: string;
  xdgConfigPath: string;
} {
  const normalizedBaseDir = normalizePathGeneral(baseDir, true);
  const xdgConfigDir = path.join(normalizedBaseDir, '.config', 'thinktank');
  
  const defaultConfigPath = path.join(normalizedBaseDir, 'thinktank.config.default.json');
  const userConfigPath = path.join(normalizedBaseDir, 'thinktank.config.json');
  const xdgConfigPath = path.join(xdgConfigDir, 'config.json');
  
  // Create configs with different settings to test precedence
  const defaultConfig = createMinimalConfig();
  const userConfig = createRealisticConfig({ models: [{ provider: 'user', modelId: 'model', enabled: true }] });
  const xdgConfig = createRealisticConfig({ models: [{ provider: 'xdg', modelId: 'model', enabled: true }] });
  
  // Create all relevant files
  setupBasicFs({
    [defaultConfigPath]: JSON.stringify(defaultConfig, null, 2),
    [userConfigPath]: JSON.stringify(userConfig, null, 2),
    [xdgConfigPath]: JSON.stringify(xdgConfig, null, 2)
  });
  
  return { defaultConfigPath, userConfigPath, xdgConfigPath };
}
