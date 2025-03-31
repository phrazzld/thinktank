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
  ],
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