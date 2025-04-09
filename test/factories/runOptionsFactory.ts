/**
 * Factory functions for creating RunOptions test data
 * 
 * This module provides functions for generating RunOptions objects 
 * with standard defaults and customizable overrides.
 */
import { RunOptions } from '../../src/workflow/runThinktank';

/**
 * Creates a RunOptions object with default values and optional overrides
 * 
 * @param overrides - Optional partial RunOptions to override default values
 * @returns A complete RunOptions object for testing
 * 
 * @example
 * ```typescript
 * const defaultOptions = createRunOptions();
 * const customOptions = createRunOptions({ 
 *   models: ['openai:gpt-4o'],
 *   includeMetadata: true
 * });
 * ```
 */
export function createRunOptions(overrides: Partial<RunOptions> = {}): RunOptions {
  return {
    input: 'test-prompt.txt',
    ...overrides
  };
}
