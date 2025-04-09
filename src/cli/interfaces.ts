import {
  RunCommandOptions,
  ModelsListCommandOptions,
  ConfigCommandOptions,
  ConfigModelsAddCommandOptions,
  ConfigModelsRemoveCommandOptions
} from './types';

/**
 * Generic interface for all CLI command handlers.
 * @template TOptions The type defining the options specific to the command.
 */
export interface CommandHandler<TOptions = Record<string, unknown>> {
  /**
   * Executes the command handler's logic.
   * @param options Options specific to the command being handled.
   * @returns Promise resolving when execution is complete.
   */
  execute(options: TOptions): Promise<void>;
}

/**
 * Handles the execution of the 'run' command.
 * This handler is responsible for executing the core thinktank functionality.
 */
export interface RunCommandHandler extends CommandHandler<RunCommandOptions> {}

/**
 * Handles the execution of the 'models list' command.
 * This handler is responsible for listing available models.
 */
export interface ModelsListCommandHandler extends CommandHandler<ModelsListCommandOptions> {}

/**
 * Handles the execution of the 'config' command.
 * This handler is responsible for configuration management.
 */
export interface ConfigCommandHandler extends CommandHandler<ConfigCommandOptions> {}

/**
 * Handles the execution of the 'config models add' command.
 * This handler is responsible for adding models to the configuration.
 */
export interface ConfigModelsAddCommandHandler extends CommandHandler<ConfigModelsAddCommandOptions> {}

/**
 * Handles the execution of the 'config models remove' command.
 * This handler is responsible for removing models from the configuration.
 */
export interface ConfigModelsRemoveCommandHandler extends CommandHandler<ConfigModelsRemoveCommandOptions> {}
