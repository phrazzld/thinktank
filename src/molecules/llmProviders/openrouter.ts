/**
 * OpenRouter provider implementation for thinktank
 */

/**
 * OpenRouter provider error class
 */
export class OpenRouterProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'OpenRouterProviderError';
  }
}