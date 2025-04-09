/**
 * Factory functions for creating AppConfig test data
 * 
 * This module provides functions for generating AppConfig objects 
 * with standard defaults and customizable overrides.
 */
import { AppConfig, ModelGroup } from '../../src/core/types';
import { createModelConfig } from './modelConfigFactory';

/**
 * Creates an AppConfig object with default values and optional overrides
 * 
 * @param overrides - Optional partial AppConfig to override default values
 * @returns A complete AppConfig object for testing
 * 
 * @example
 * ```typescript
 * const defaultConfig = createAppConfig();
 * const customConfig = createAppConfig({ 
 *   models: [createModelConfig({ provider: 'openai' })]
 * });
 * ```
 */
export function createAppConfig(overrides: Partial<AppConfig> = {}): AppConfig {
  const defaultModel = createModelConfig();
  const defaultModels = [defaultModel];
  
  const defaultGroups: Record<string, ModelGroup> = {
    default: { 
      name: 'default', 
      models: defaultModels,
      systemPrompt: { text: 'Default system prompt for testing' }
    }
  };
  
  return {
    models: defaultModels,
    groups: defaultGroups,
    ...overrides
  };
}
