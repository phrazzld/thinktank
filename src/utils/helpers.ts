/**
 * Helper functions for the thinktank application
 */
import { ModelConfig } from '../core/types';
import path from 'path';

/**
 * Generates a unique key for a model configuration in the format "provider:modelId"
 * Used for model identification throughout the application
 * 
 * @param config - The model configuration object
 * @returns The unique model key
 */
export function getModelConfigKey(config: ModelConfig): string {
  return `${config.provider}:${config.modelId}`;
}

/**
 * Determines the conventional environment variable name for a provider's API key
 * Example: "openai" -> "OPENAI_API_KEY"
 * 
 * @param provider - The provider name (e.g., "openai", "anthropic")
 * @returns The standard environment variable name for the provider's API key
 */
export function getDefaultApiKeyEnvVar(provider: string): string {
  // Convert to uppercase and append _API_KEY
  return `${provider.toUpperCase()}_API_KEY`;
}

/**
 * Safely extracts a model's API key from environment variables
 * First checks the custom apiKeyEnvVar if specified, then tries standard mappings,
 * then falls back to default naming pattern
 * 
 * @param config - The model configuration
 * @returns The API key if found, null otherwise
 */
export function getApiKey(config: ModelConfig): string | null {
  // First try the custom environment variable if specified
  if (config.apiKeyEnvVar) {
    const key = process.env[config.apiKeyEnvVar];
    if (key) {
      return key;
    }
  }
  
  // Standard environment variable mappings by provider
  const envVarMappings: Record<string, string[]> = {
    'openai': ['OPENAI_API_KEY'],
    'anthropic': ['ANTHROPIC_API_KEY'],
    'google': ['GEMINI_API_KEY', 'GOOGLE_API_KEY'], // Check multiple possible names
    'openrouter': ['OPENROUTER_API_KEY']
  };
  
  // Handle case-insensitive provider matching
  const provider = config.provider.toLowerCase();
  const possibleVars = envVarMappings[provider] || [`${provider.toUpperCase()}_API_KEY`];
  
  // Try each possible environment variable
  for (const envVar of possibleVars) {
    const key = process.env[envVar];
    if (key) {
      return key;
    }
  }
  
  return null;
}

/**
 * Trims and normalizes a string to remove extra whitespace
 * Useful for handling user-provided prompts
 * 
 * @param text - The input text to normalize
 * @returns The normalized text
 */
export function normalizeText(text: string): string {
  // Remove leading/trailing whitespace and normalize internal whitespace
  return text.trim().replace(/\s+/g, ' ');
}

/**
 * Generates a timestamped directory name for model outputs
 * Format: thinktank_run_YYYYMMDD_HHmmss_SSS
 * 
 * @returns The generated directory name
 */
export function generateRunDirectoryName(): string {
  const now = new Date();
  
  // Format: YYYYMMDD_HHmmss_SSS
  const timestamp = now.toISOString()
    .replace(/[-:T.Z]/g, match => match === 'T' ? '_' : match === '.' ? '_' : '')
    .replace(/^(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})_(\d{3}).*$/, '$1$2$3_$4$5$6_$7');
  
  return `thinktank_run_${timestamp}`;
}

/**
 * Resolves the output directory path based on the provided output option
 * If output is specified, it's used as the output directory
 * Otherwise, a default directory in the current working directory is used
 * 
 * @param outputOption - The output option from CLI/config, if provided
 * @param defaultDirName - The default directory name to use if outputOption is not provided
 * @returns The resolved path to the output directory
 */
export function resolveOutputDirectory(
  outputOption?: string,
  defaultDirName: string = 'thinktank-reports'
): string {
  // Use provided path or default to 'thinktank-reports' in current working directory
  const targetPath = outputOption
    ? path.resolve(outputOption) 
    : path.resolve(process.cwd(), defaultDirName);
  
  return targetPath;
}

/**
 * Generates a complete output directory path including a model/group name and timestamp
 * 
 * @param outputOption - The output option from CLI/config, if provided
 * @param identifier - The model or group identifier (optional)
 * @returns The resolved path to the run-specific output directory
 */
export function generateOutputDirectoryPath(
  outputOption?: string, 
  identifier?: string
): string {
  // Get the base output directory (always use 'thinktank-output' as default)
  const baseOutputPath = resolveOutputDirectory(outputOption, 'thinktank-output');
  
  // Generate a timestamp in a simpler format: YYYYMMDD-HHmmss
  const now = new Date();
  const timestamp = now.toISOString()
    .replace(/[-:T.Z]/g, match => match === 'T' ? '-' : match === '.' ? '' : '')
    .substring(0, 15); // Get YYYYMMDD-HHmmss format
  
  // Generate the directory name
  let dirName = `run-${timestamp}`;
  
  // If an identifier was provided, include it in the directory name
  if (identifier) {
    // For provider:model format, replace colon with hyphen
    const safeIdentifier = identifier.replace(/:/g, '-');
    dirName = `${safeIdentifier}-${timestamp}`;
  }
  
  // Return the full path to the run-specific directory
  return path.join(baseOutputPath, dirName);
}

/**
 * Sanitizes a string for safe use as a filename
 * Replaces invalid characters with underscores
 * 
 * @param input - The input string to sanitize
 * @returns A sanitized string safe for use as a filename
 */
export function sanitizeFilename(input: string): string {
  if (!input) {
    return 'unnamed';
  }
  
  // Replace characters that are invalid in filenames across common operating systems
  // This includes: / \ : * ? " < > | and control characters
  
  // Temporarily disable eslint for the next line as we need to match control characters
  // eslint-disable-next-line no-control-regex
  const sanitized = input.replace(/[\x00-\x1F]/g, '');
  
  return sanitized
    .replace(/[/\\:*?"<>|]/g, '_')      // Replace invalid chars with underscore
    .replace(/\s+/g, '_')                   // Replace whitespace with underscore
    .replace(/__+/g, '_')                   // Replace multiple underscores with one
    .replace(/^[.-]+|[.-]+$/g, '')          // Remove leading/trailing dots and hyphens
    .substring(0, 255);                     // Limit length to 255 characters
}

/**
 * Helper function to debug API key availability
 * Returns status of common API key environment variables
 * Only provides information in development mode to avoid leaking sensitive information
 */
export function debugApiKeyAvailability(): Record<string, boolean> {
  const result: Record<string, boolean> = {};
  
  if (process.env.NODE_ENV === 'development') {
    const keys = [
      'OPENAI_API_KEY', 
      'ANTHROPIC_API_KEY', 
      'GEMINI_API_KEY', 
      'GOOGLE_API_KEY', 
      'OPENROUTER_API_KEY'
    ];
    
    keys.forEach(key => {
      result[key] = !!process.env[key];
    });
  }
  
  return result;
}