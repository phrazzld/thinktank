/**
 * Core constants for the thinktank application
 */
import path from 'path';
import { AppConfig } from './types';

/**
 * Default configuration structure used when no config file exists
 * or for merging with user-provided config
 */
export const DEFAULT_CONFIG: AppConfig = {
  models: [
    {
      provider: 'openai',
      modelId: 'gpt-4o',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-opus-20240229',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-sonnet-20240229',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-haiku-20240307',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-5-sonnet-20240620',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-7-sonnet-20250219',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'google',
      modelId: 'gemini-1.5-flash',
      enabled: false,
      apiKeyEnvVar: 'GEMINI_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'google',
      modelId: 'gemini-1.5-pro',
      enabled: false,
      apiKeyEnvVar: 'GEMINI_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'google',
      modelId: 'gemini-pro',
      enabled: false,
      apiKeyEnvVar: 'GEMINI_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'openrouter',
      modelId: 'openai/gpt-4o',
      enabled: false,
      apiKeyEnvVar: 'OPENROUTER_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'openrouter',
      modelId: 'anthropic/claude-3-opus-20240229',
      enabled: false,
      apiKeyEnvVar: 'OPENROUTER_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
    {
      provider: 'openrouter',
      modelId: 'meta-llama/llama-3-70b-instruct',
      enabled: false,
      apiKeyEnvVar: 'OPENROUTER_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 1000,
      },
    },
  ],
  groups: {
    default: {
      name: 'default',
      description: 'Default model group with general-purpose system prompt',
      systemPrompt: {
        text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information. If you are unsure about something, admit it rather than making up an answer.',
        metadata: {
          source: 'default-configuration'
        }
      },
      models: [] // Will be populated with all enabled models during config normalization
    },
    coding: {
      name: 'coding',
      description: 'Models optimized for code generation and programming tasks',
      systemPrompt: {
        text: 'You are an expert software engineer with deep knowledge of programming languages, algorithms, design patterns, and software architecture. Provide accurate, efficient, and well-documented code. Always consider edge cases, performance implications, and security best practices.',
        metadata: {
          source: 'default-configuration'
        }
      },
      models: [] // Will be populated during configuration
    }
  }
};

/**
 * Path to the default configuration template file in the package
 * This is used to initialize a new configuration file if none exists
 */
export const DEFAULT_CONFIG_TEMPLATE_PATH = path.resolve(__dirname, '../../config/thinktank.config.default.json');