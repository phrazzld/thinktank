/**
 * Factory functions for creating LLMResponse test data
 * 
 * This module provides functions for generating LLMResponse objects 
 * with standard defaults and customizable overrides.
 */
import { LLMResponse } from '../../src/core/types';

/**
 * Creates an LLMResponse object with default values and optional overrides
 * 
 * @param overrides - Optional partial LLMResponse to override default values
 * @returns A complete LLMResponse object for testing, with configKey added
 * 
 * @example
 * ```typescript
 * const defaultResponse = createLlmResponse();
 * const errorResponse = createLlmResponse({ 
 *   error: 'API rate limit exceeded'
 * });
 * ```
 */
export function createLlmResponse(
  overrides: Partial<LLMResponse> = {}
): LLMResponse & { configKey: string } {
  const provider = overrides.provider ?? 'test-provider';
  const modelId = overrides.modelId ?? 'test-model';
  
  return {
    provider,
    modelId,
    text: 'Default test response for testing purposes',
    metadata: { 
      responseTime: 1000,
      usage: { 
        input_tokens: 100,
        output_tokens: 200,
        total_tokens: 300
      }
    },
    configKey: `${provider}:${modelId}`,
    ...overrides
  };
}