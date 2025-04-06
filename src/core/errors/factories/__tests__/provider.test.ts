/* eslint-disable @typescript-eslint/no-unused-vars */
import {
  createProviderApiKeyMissingError,
  createProviderRateLimitError,
  createProviderModelNotFoundError,
  createProviderTokenLimitError,
  createProviderContentPolicyError,
  createProviderUnknownError,
  createProviderNetworkError,
  isProviderRateLimitError,
  isProviderTokenLimitError,
  isProviderContentPolicyError,
  isProviderAuthError,
  isProviderNetworkError
} from '../provider';
import { ApiError } from '../../types/api';

describe('Provider error factory functions', () => {
  describe('createProviderApiKeyMissingError', () => {
    it('creates an API key missing error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const consoleUrl = 'https://test-provider.com/keys';
      
      const error = createProviderApiKeyMissingError(providerId, providerName, consoleUrl);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain(`${providerName} API key is missing`);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.length).toBe(3);
      expect(error.suggestions?.[0]).toContain(`${providerName.toUpperCase()}_API_KEY`);
      expect(error.suggestions?.[1]).toContain(consoleUrl);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBe(2);
      expect(error.examples?.[0]).toContain(`${providerName.toUpperCase()}_API_KEY`);
    });
  });

  describe('createProviderRateLimitError', () => {
    it('creates a rate limit error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const originalError = new Error('Rate limit exceeded');
      
      const error = createProviderRateLimitError(providerId, providerName, originalError);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain('Rate limit exceeded');
      expect(error.cause).toBe(originalError);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.length).toBe(4);
      expect(error.suggestions?.[2]).toContain(providerName);
      expect(error.examples).toBeDefined();
      expect(error.examples?.join('')).toContain('exponential backoff');
    });
  });

  describe('createProviderModelNotFoundError', () => {
    it('creates a model not found error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const modelId = 'nonexistent-model';
      
      const error = createProviderModelNotFoundError(providerId, providerName, modelId);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain(`Model '${modelId}' not found`);
      expect(error.message).toContain(providerName);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.[0]).toContain('available models');
    });
  });

  describe('createProviderTokenLimitError', () => {
    it('creates a token limit error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const originalError = new Error('Token limit exceeded');
      
      const error = createProviderTokenLimitError(providerId, providerName, originalError);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain('Token limit');
      expect(error.cause).toBe(originalError);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('shorter'))).toBe(true);
    });
  });

  describe('createProviderUnknownError', () => {
    it('creates an unknown provider error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      
      const error = createProviderUnknownError(providerId, providerName);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain(`Unknown error occurred while generating text from ${providerName}`);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('network'))).toBe(true);
    });
  });

  describe('createProviderContentPolicyError', () => {
    it('creates a content policy error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const originalError = new Error('Content policy violation');
      
      const error = createProviderContentPolicyError(providerId, providerName, originalError);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain('Content policy violation');
      expect(error.cause).toBe(originalError);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('policy'))).toBe(true);
    });
  });

  describe('createProviderNetworkError', () => {
    it('creates a network error with correct properties', () => {
      const providerId = 'test-provider';
      const providerName = 'Test Provider';
      const originalError = new Error('Network connection failed');
      
      const error = createProviderNetworkError(providerId, providerName, originalError);
      
      expect(error).toBeInstanceOf(ApiError);
      expect(error.category).toBe('API');
      expect(error.providerId).toBe(providerId);
      expect(error.message).toContain('Network error');
      expect(error.cause).toBe(originalError);
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some((s: string) => s.includes('connection'))).toBe(true);
    });
  });

  describe('Error detection utility functions', () => {
    it('isProviderRateLimitError detects rate limit errors', () => {
      expect(isProviderRateLimitError('Rate limit exceeded')).toBe(true);
      expect(isProviderRateLimitError('Status 429: Too many requests')).toBe(true);
      expect(isProviderRateLimitError('You have exceeded your quota')).toBe(true);
      expect(isProviderRateLimitError('Regular error message')).toBe(false);
    });

    it('isProviderTokenLimitError detects token limit errors', () => {
      expect(isProviderTokenLimitError('Token limit exceeded')).toBe(true);
      expect(isProviderTokenLimitError('Maximum context length exceeded')).toBe(true);
      expect(isProviderTokenLimitError('Regular error message')).toBe(false);
    });

    it('isProviderContentPolicyError detects content policy errors', () => {
      expect(isProviderContentPolicyError('Content policy violation')).toBe(true);
      expect(isProviderContentPolicyError('Content filtered by safety system')).toBe(true);
      expect(isProviderContentPolicyError('Content violates our terms')).toBe(true);
      expect(isProviderContentPolicyError('Regular error message')).toBe(false);
    });

    it('isProviderAuthError detects authentication errors', () => {
      expect(isProviderAuthError('Invalid API key')).toBe(true);
      expect(isProviderAuthError('Authentication failed with status 401')).toBe(true);
      expect(isProviderAuthError('Regular error message')).toBe(false);
    });

    it('isProviderNetworkError detects network errors', () => {
      expect(isProviderNetworkError('Network connection failed')).toBe(true);
      expect(isProviderNetworkError('ECONNREFUSED: Connection refused')).toBe(true);
      expect(isProviderNetworkError('Socket timeout')).toBe(true);
      expect(isProviderNetworkError('Regular error message')).toBe(false);
    });
  });
});