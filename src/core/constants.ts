/**
 * Core constants for the thinktank application
 */
import path from 'path';
import { AppConfig } from './types';

/**
 * Default configuration structure used when no config file exists
 * or as a fallback if the template file is not available
 */
export const DEFAULT_CONFIG: AppConfig = {
  models: [
    {
      provider: 'openai',
      modelId: 'gpt-4o',
      enabled: true,
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
    {
      provider: 'openai',
      modelId: 'gpt-3.5-turbo',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-opus-20240229',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-sonnet-20240229',
      enabled: false,
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
    {
      provider: 'anthropic',
      modelId: 'claude-3-haiku-20240307',
      enabled: true,
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
    {
      provider: 'google',
      modelId: 'gemini-1.5-pro',
      enabled: false,
      apiKeyEnvVar: 'GEMINI_API_KEY',
      options: {
        temperature: 0.7,
        maxTokens: 2000,
      },
    },
  ],
  groups: {
    default: {
      name: 'default',
      description: 'Default group with general-purpose models',
      systemPrompt: {
        text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information. If you are unsure about something, admit it rather than making up an answer.',
        metadata: {
          source: 'default-configuration'
        }
      },
      models: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true
        },
        {
          provider: 'anthropic',
          modelId: 'claude-3-haiku-20240307',
          enabled: true
        }
      ]
    },
    coding: {
      name: 'coding',
      description: 'Models optimized for programming tasks',
      systemPrompt: {
        text: 'You are an expert software engineer with deep knowledge of programming languages, algorithms, design patterns, and software architecture. Provide accurate, efficient, and well-documented code. Consider edge cases, performance, and security best practices.',
        metadata: {
          source: 'default-configuration'
        }
      },
      models: [
        {
          provider: 'openai',
          modelId: 'gpt-4o',
          enabled: true
        }
      ]
    },
    creative: {
      name: 'creative',
      description: 'Models optimized for creative writing and brainstorming',
      systemPrompt: {
        text: 'You are a creative assistant with a talent for generating innovative ideas, compelling narratives, and engaging content. Think outside the box and offer diverse perspectives and original concepts.',
        metadata: {
          source: 'default-configuration'
        }
      },
      models: [
        {
          provider: 'anthropic',
          modelId: 'claude-3-haiku-20240307',
          enabled: true
        }
      ]
    }
  }
};

/**
 * Path to the default configuration template file in the package
 * This is used to initialize a new configuration file if none exists
 */
export const DEFAULT_CONFIG_TEMPLATE_PATH = path.resolve(__dirname, '../../config/thinktank.config.default.json');
