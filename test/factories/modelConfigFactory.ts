/**
 * Factory functions for creating ModelConfig test data
 * 
 * This module provides functions for generating ModelConfig objects 
 * with standard defaults and customizable overrides.
 */
import { ModelConfig } from '../../src/core/types';

/**
 * Creates a ModelConfig object with default values and optional overrides
 * 
 * @param overrides - Optional partial ModelConfig to override default values
 * @returns A complete ModelConfig object for testing
 * 
 * @example
 * ```typescript
 * const defaultModel = createModelConfig();
 * const customModel = createModelConfig({ provider: 'openai', modelId: 'gpt-4o' });
 * ```
 */
export function createModelConfig(overrides: Partial<ModelConfig> = {}): ModelConfig {
  return {
    provider: 'test-provider',
    modelId: `test-model-${Math.random().toString(36).substring(7)}`,
    enabled: true,
    options: { temperature: 0.7 },
    ...overrides
  };
}