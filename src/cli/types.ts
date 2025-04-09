/**
 * Type definitions for CLI command options
 */

/**
 * Base options common to multiple commands
 */
export interface BaseCommandOptions {
  /** Path to custom config file */
  config?: string;
  /** Enable verbose output */
  verbose?: boolean;
}

/**
 * Options for the 'run' command
 */
export interface RunCommandOptions extends BaseCommandOptions {
  /** Path to the prompt file */
  promptFile: string;
  /** Paths to context directories or files */
  contextPaths?: string[];
  /** Specific models to use (IDs) */
  models?: string[];
  /** Single specific model to use */
  specificModel?: string;
  /** Model group name to use */
  groupName?: string;
  /** Output directory for results */
  outputDir?: string;
}

/**
 * Options for the 'models list' command
 */
export interface ModelsListCommandOptions extends BaseCommandOptions {
  /** Filter models by provider */
  provider?: string;
  /** Show detailed information for each model */
  detailed?: boolean;
}

/**
 * Options for the 'config' command base
 */
export interface ConfigCommandOptions extends BaseCommandOptions {
  /** Command action (list, get, set, etc.) */
  action?: string;
  /** Configuration key */
  key?: string;
  /** Configuration value */
  value?: string;
}

/**
 * Options for 'config models add' command
 */
export interface ConfigModelsAddCommandOptions extends BaseCommandOptions {
  /** Provider ID */
  provider: string;
  /** Model ID */
  modelId: string;
  /** JSON string of model options */
  options?: string;
  /** Enable the model */
  enable?: boolean;
  /** Disable the model */
  disable?: boolean;
  /** Environment variable name for API key */
  apiKeyEnv?: string;
}

/**
 * Options for 'config models remove' command
 */
export interface ConfigModelsRemoveCommandOptions extends BaseCommandOptions {
  /** Provider ID */
  provider: string;
  /** Model ID */
  modelId: string;
}
