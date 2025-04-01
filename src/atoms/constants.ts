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
 * Paths where the application will look for config files, in order of precedence
 */
export const CONFIG_SEARCH_PATHS: string[] = [
  // Current working directory
  path.resolve(process.cwd(), 'thinktank.config.json'),
  
  // User's config directory
  path.resolve(getUserConfigDir(), 'thinktank/config.json'),
  
  // Application directory
  path.resolve(__dirname, '../../templates/thinktank.config.default.json'),
];

/**
 * Get the user's config directory based on platform
 */
function getUserConfigDir(): string {
  const { platform } = process;
  
  // Windows: %APPDATA% (e.g., C:\Users\Username\AppData\Roaming)
  if (platform === 'win32') {
    return process.env.APPDATA || path.join(process.env.USERPROFILE || '', 'AppData', 'Roaming');
  }
  
  // macOS: ~/Library/Preferences
  if (platform === 'darwin') {
    return path.join(process.env.HOME || '', 'Library', 'Preferences');
  }
  
  // Linux/Unix: ~/.config (XDG Base Directory Specification)
  return path.join(process.env.HOME || '', '.config');
}