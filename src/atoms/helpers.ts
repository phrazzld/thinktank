/**
 * Helper functions for the Thinktank application
 */
import { ModelConfig } from './types';
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
 * First checks the custom apiKeyEnvVar if specified, then falls back to default naming
 * 
 * @param config - The model configuration
 * @returns The API key if found, undefined otherwise
 */
export function getApiKey(config: ModelConfig): string | undefined {
  // First try the custom environment variable if specified
  if (config.apiKeyEnvVar && process.env[config.apiKeyEnvVar]) {
    return process.env[config.apiKeyEnvVar];
  }
  
  // Fall back to the default environment variable pattern
  const defaultEnvVar = getDefaultApiKeyEnvVar(config.provider);
  return process.env[defaultEnvVar];
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
 * Generates a complete output directory path including a timestamped run subdirectory
 * 
 * @param outputOption - The output option from CLI/config, if provided
 * @returns The resolved path to the run-specific output directory
 */
export function generateOutputDirectoryPath(outputOption?: string): string {
  // Get the base output directory
  const baseOutputPath = resolveOutputDirectory(outputOption);
  
  // Generate the unique run directory name
  const runDirectoryName = generateRunDirectoryName();
  
  // Return the full path to the run-specific directory
  return path.join(baseOutputPath, runDirectoryName);
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