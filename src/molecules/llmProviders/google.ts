/**
 * Google Gemini provider implementation for thinktank
 */

/**
 * Google provider error class
 */
export class GoogleProviderError extends Error {
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'GoogleProviderError';
  }
}