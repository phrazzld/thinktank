/**
 * Provider testing setup utilities
 * 
 * This module provides specialized setup helpers for provider/API-related tests.
 */
import type { LLMResponse } from '../../src/core/types';

/**
 * Creates a mock LLM response object for provider tests
 * 
 * @param providerId - Provider ID (e.g., 'openai', 'anthropic')
 * @param modelId - Model ID (e.g., 'gpt-4o', 'claude-3-7-sonnet-20240229')
 * @param responseText - Text response content
 * @param error - Optional error object
 * @returns A mock LLMResponse object
 * 
 * Usage:
 * ```typescript
 * const response = createMockLlmResponse('openai', 'gpt-4o', 'Generated text');
 * ```
 */
export function createMockLlmResponse(
  providerId: string,
  modelId: string,
  responseText: string,
  error?: Error | null
): LLMResponse {
  return {
    provider: providerId,
    modelId,
    text: responseText,
    error: error ? error.message : undefined,
    metadata: {
      usage: {
        promptTokens: 100,
        completionTokens: 50,
        totalTokens: 150
      },
      timestamp: new Date().toISOString()
    }
  };
}

/**
 * Creates a mock for network requests (fetch)
 * 
 * @param responseData - Response data object
 * @param statusCode - HTTP status code
 * @returns Mock function for fetch
 * 
 * Usage:
 * ```typescript
 * const mockFetch = createMockFetch({ data: 'test' });
 * global.fetch = mockFetch;
 * ```
 */
export function createMockFetch<T>(responseData: T, statusCode = 200): jest.Mock {
  return jest.fn().mockResolvedValue({
    ok: statusCode >= 200 && statusCode < 300,
    status: statusCode,
    statusText: statusCode === 200 ? 'OK' : 'Error',
    json: jest.fn().mockResolvedValue(responseData),
    text: jest.fn().mockResolvedValue(JSON.stringify(responseData))
  });
}

/**
 * Creates a mock for network error responses
 * 
 * @param errorMessage - Error message
 * @param statusCode - HTTP status code
 * @returns Mock function for fetch that returns an error response
 * 
 * Usage:
 * ```typescript
 * const mockFetch = createMockFetchError('Rate limit exceeded', 429);
 * global.fetch = mockFetch;
 * ```
 */
export function createMockFetchError(errorMessage: string, statusCode = 500): jest.Mock {
  return jest.fn().mockResolvedValue({
    ok: false,
    status: statusCode,
    statusText: errorMessage,
    json: jest.fn().mockResolvedValue({ error: errorMessage }),
    text: jest.fn().mockResolvedValue(JSON.stringify({ error: errorMessage }))
  });
}

/**
 * Creates a mock for network fetch failures (e.g., connection issues)
 * 
 * @param errorMessage - Error message
 * @returns Mock function for fetch that throws a network error
 * 
 * Usage:
 * ```typescript
 * const mockFetch = createMockFetchFailure('Network error');
 * global.fetch = mockFetch;
 * ```
 */
export function createMockFetchFailure(errorMessage: string): jest.Mock {
  return jest.fn().mockRejectedValue(new Error(errorMessage));
}

/**
 * Sets up mocks for specific provider API responses
 * 
 * @param providerId - Provider ID (e.g., 'openai', 'anthropic')
 * @param responseText - Text to return in the response
 * @returns Mock functions and data for provider API tests
 * 
 * Usage:
 * ```typescript
 * const { mockFetch, mockResponse } = setupProviderMock('openai', 'Generated text');
 * global.fetch = mockFetch;
 * ```
 */
export function setupProviderMock(providerId: string, responseText: string): {
  mockFetch: jest.Mock;
  mockResponse: Record<string, unknown>;
} {
  let mockResponse;
  
  // Create provider-specific response format
  switch (providerId) {
    case 'openai':
      mockResponse = {
        choices: [{ message: { content: responseText } }],
        usage: { prompt_tokens: 100, completion_tokens: 50, total_tokens: 150 }
      };
      break;
    case 'anthropic':
      mockResponse = {
        content: [{ text: responseText }],
        usage: { input_tokens: 100, output_tokens: 50 }
      };
      break;
    case 'google':
      mockResponse = {
        candidates: [{ content: { parts: [{ text: responseText }] } }],
        usageMetadata: { promptTokenCount: 100, candidatesTokenCount: 50 }
      };
      break;
    default:
      mockResponse = { text: responseText };
  }
  
  // Create fetch mock with the appropriate response
  const mockFetch = createMockFetch(mockResponse);
  
  return { mockFetch, mockResponse };
}
