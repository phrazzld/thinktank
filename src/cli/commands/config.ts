/**
 * Config command implementation
 * 
 * This command manages thinktank configuration
 */
import { Command } from 'commander';
import { handleError } from '../index';
import * as loadConfig from '../../core/configManager';
import { colors } from '../../utils/consoleUtils';
import { fileExists, getConfigFilePath } from '../../utils/fileReader';
import { AppConfig } from '../../core/types';

// Create the main config command
const configCommand = new Command('config');
configCommand.description('Manage thinktank configuration');

// Path command to show the config file path
const pathCommand = new Command('path');
pathCommand
  .description('Show the path to the config file')
  .action(async () => {
    try {
      // Get the XDG config path - this is now our canonical configuration location
      const configPath = await getConfigFilePath();
      
      // Show the path with more descriptive context
      // eslint-disable-next-line no-console
      console.log(`Configuration file location: ${colors.cyan(configPath)}`);
      
      // Check if the file exists and show additional info
      if (await fileExists(configPath)) {
        // eslint-disable-next-line no-console
        console.log(`Status: ${colors.green('File exists')}`);
      } else {
        // eslint-disable-next-line no-console
        console.log(`Status: ${colors.yellow('File does not exist yet')} (will be created when needed)`);
      }
      
      // Show more helpful contextual information
      // eslint-disable-next-line no-console
      console.log(colors.dim('\nThis is the default location where thinktank stores its configuration.'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('You can override this location with the --config option when running commands.'));
      
      // For backward compatibility, also mention the project-local config option
      const projectLocalPath = loadConfig.getDefaultConfigPath();
      if (projectLocalPath !== configPath) {
        // eslint-disable-next-line no-console
        console.log(colors.dim(`\nA project-local configuration can be placed at: ${projectLocalPath}`));
        // eslint-disable-next-line no-console
        console.log(colors.dim('Project-local configs can be used to share settings with a team via version control.'));
      }
    } catch (error) {
      handleError(error);
    }
  });

// Show command to display the current configuration
const showCommand = new Command('show');
showCommand
  .description('Show the current configuration')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-j, --json', 'Output as JSON')
  .option('-m, --models-only', 'Show only models configuration')
  .option('-g, --groups-only', 'Show only groups configuration')
  .action(async (options: { 
    config?: string;
    json?: boolean;
    modelsOnly?: boolean;
    groupsOnly?: boolean;
  }) => {
    try {
      // Load the configuration
      const config = await loadConfig.loadConfig({ 
        configPath: options.config,
        mergeWithDefaults: true
      });
      
      // If JSON output is requested, display as formatted JSON
      if (options.json) {
        let outputConfig = config;
        
        // Filter based on options
        if (options.modelsOnly) {
          outputConfig = { models: config.models, groups: {} };
        } else if (options.groupsOnly) {
          outputConfig = { models: [], groups: config.groups || {} };
        }
        
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(outputConfig, null, 2));
        return;
      }
      
      // Display models if requested or no specific filter
      if (!options.groupsOnly) {
        // eslint-disable-next-line no-console
        console.log(colors.blue('📋 Configured Models:'));
        
        if (config.models.length === 0) {
          // eslint-disable-next-line no-console
          console.log(colors.yellow('  No models configured'));
        } else {
          // Group models by provider for better display
          const modelsByProvider: Record<string, typeof config.models> = {};
          
          config.models.forEach(model => {
            if (!modelsByProvider[model.provider]) {
              modelsByProvider[model.provider] = [];
            }
            modelsByProvider[model.provider].push(model);
          });
          
          // Display models grouped by provider
          Object.entries(modelsByProvider).forEach(([provider, models]) => {
            // eslint-disable-next-line no-console
            console.log(`\n  ${colors.cyan(provider)} (${models.length}):`);
            
            models.forEach(model => {
              const status = model.enabled 
                ? colors.green('Enabled') 
                : colors.red('Disabled');
              
              // eslint-disable-next-line no-console
              console.log(`    - ${colors.yellow(model.modelId)} [${status}]`);
              
              // Show model options if available
              if (model.options) {
                const options = Object.entries(model.options)
                  .map(([key, value]) => `${key}: ${JSON.stringify(value)}`)
                  .join(', ');
                
                // eslint-disable-next-line no-console
                console.log(`      Options: ${colors.dim(options)}`);
              }
              
              // Show API key environment variable if available
              if (model.apiKeyEnvVar) {
                // eslint-disable-next-line no-console
                console.log(`      API Key: ${colors.dim(`\${${model.apiKeyEnvVar}}`)}`);
              }
            });
          });
        }
      }
      
      // Display groups if requested or no specific filter
      if (!options.modelsOnly && config.groups) {
        // eslint-disable-next-line no-console
        console.log(colors.blue('\n📂 Configured Groups:'));
        
        const groupEntries = Object.entries(config.groups);
        if (groupEntries.length === 0) {
          // eslint-disable-next-line no-console
          console.log(colors.yellow('  No groups configured'));
        } else {
          groupEntries.forEach(([groupName, group]) => {
            const modelCount = group.models.length;
            
            // eslint-disable-next-line no-console
            console.log(`\n  ${colors.green(groupName)}${group.description ? ` - ${group.description}` : ''}`);
            // eslint-disable-next-line no-console
            console.log(`    Models: ${colors.yellow(modelCount.toString())}`);
            
            // Display system prompt if available
            if (group.systemPrompt && group.systemPrompt.text) {
              const promptText = group.systemPrompt.text.length > 50
                ? `${group.systemPrompt.text.substring(0, 50)}...`
                : group.systemPrompt.text;
              
              // eslint-disable-next-line no-console
              console.log(`    System Prompt: ${colors.dim(promptText)}`);
            }
            
            // Display model IDs in the group
            if (modelCount > 0) {
              const modelsList = group.models
                .map(model => `${model.provider}:${model.modelId}`)
                .join(', ');
              
              // eslint-disable-next-line no-console
              console.log(`    Includes: ${colors.dim(modelsList)}`);
            }
          });
        }
      }
      
      // Show a helpful tip at the end
      // eslint-disable-next-line no-console
      console.log(colors.dim('\nTip: Use --json for machine-readable output'));
    } catch (error) {
      handleError(error);
    }
  });

// Add the subcommands to the main config command
configCommand.addCommand(pathCommand);
configCommand.addCommand(showCommand);

// Add a models command group
const modelsCommand = new Command('models');
modelsCommand.description('Manage model configurations');

// List models command
const listModelsCommand = new Command('list');
listModelsCommand
  .description('List all configured models')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-p, --provider <provider>', 'Filter models by provider')
  .option('-e, --enabled-only', 'Show only enabled models')
  .option('-d, --disabled-only', 'Show only disabled models')
  .option('-j, --json', 'Output as JSON')
  .action(async (options: { 
    config?: string;
    provider?: string;
    enabledOnly?: boolean;
    disabledOnly?: boolean;
    json?: boolean;
  }) => {
    try {
      // Load the configuration
      const config = await loadConfig.loadConfig({ configPath: options.config });
      
      // Filter models based on options
      let filteredModels = config.models;
      
      // Filter by provider if specified
      if (options.provider) {
        filteredModels = filteredModels.filter(model => 
          model.provider === options.provider
        );
      }
      
      // Filter by enabled/disabled status if specified
      if (options.enabledOnly) {
        filteredModels = filteredModels.filter(model => model.enabled);
      } else if (options.disabledOnly) {
        filteredModels = filteredModels.filter(model => !model.enabled);
      }
      
      // If JSON output is requested, display the filtered models as JSON
      if (options.json) {
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(filteredModels, null, 2));
        return;
      }
      
      // Format and display the filtered models
      // eslint-disable-next-line no-console
      console.log(colors.blue('📋 Configured Models:'));
      
      if (filteredModels.length === 0) {
        // eslint-disable-next-line no-console
        console.log(
          options.provider 
            ? colors.yellow(`  No models found for provider '${options.provider}'`)
            : colors.yellow('  No models configured')
        );
        return;
      }
      
      // Group models by provider for better display
      const modelsByProvider: Record<string, typeof config.models> = {};
      
      filteredModels.forEach(model => {
        if (!modelsByProvider[model.provider]) {
          modelsByProvider[model.provider] = [];
        }
        modelsByProvider[model.provider].push(model);
      });
      
      // Display models grouped by provider
      Object.entries(modelsByProvider).forEach(([provider, models]) => {
        // eslint-disable-next-line no-console
        console.log(`\n  ${colors.cyan(provider)} (${models.length}):`);
        
        models.forEach((model, index) => {
          const status = model.enabled 
            ? colors.green('Enabled') 
            : colors.red('Disabled');
          
          // eslint-disable-next-line no-console
          console.log(`    ${index + 1}. ${colors.yellow(model.modelId)} [${status}]`);
          
          // Find which groups this model belongs to
          if (config.groups) {
            const groups = Object.entries(config.groups)
              .filter(([_, group]) => 
                group.models.some(m => 
                  m.provider === model.provider && m.modelId === model.modelId
                )
              )
              .map(([name]) => name);
              
            if (groups.length > 0) {
              // eslint-disable-next-line no-console
              console.log(`      Groups: ${colors.dim(groups.join(', '))}`);
            }
          }
          
          // Show model options if available
          if (model.options) {
            const options = Object.entries(model.options)
              .map(([key, value]) => `${key}: ${JSON.stringify(value)}`)
              .join(', ');
            
            // eslint-disable-next-line no-console
            console.log(`      Options: ${colors.dim(options)}`);
          }
          
          // Show API key environment variable if available
          if (model.apiKeyEnvVar) {
            // eslint-disable-next-line no-console
            console.log(`      API Key: ${colors.dim(`\${${model.apiKeyEnvVar}}`)}`);
          }
        });
      });
      
      // Show command tips
      // eslint-disable-next-line no-console
      console.log(colors.dim('\nTips:'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use --provider to filter models by provider'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use --enabled-only or --disabled-only to filter by status'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use --json for machine-readable output'));
    } catch (error) {
      handleError(error);
    }
  });

// Add model command
const addModelCommand = new Command('add');
addModelCommand
  .description('Add a new model definition')
  .argument('<provider>', 'Provider ID (e.g., openai, anthropic)')
  .argument('<modelId>', 'Model ID (e.g., gpt-4o, claude-3-opus)')
  .option('-o, --options <json>', 'JSON string of model options')
  .option('-e, --enable', 'Enable the model (default)')
  .option('-d, --disable', 'Disable the model')
  .option('-k, --api-key-env <variable>', 'Environment variable name for API key')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (provider: string, modelId: string, options: { 
    options?: string;
    enable?: boolean;
    disable?: boolean;
    apiKeyEnv?: string;
    config?: string;
  }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Parse the options JSON if provided
      let modelOptions: Record<string, unknown> | undefined;
      if (options.options) {
        try {
          modelOptions = JSON.parse(options.options) as Record<string, unknown>;
        } catch (error) {
          throw new loadConfig.ConfigError(`Invalid options JSON: ${options.options}`);
        }
      }
      
      // Determine if the model should be enabled
      // Default to true unless disable flag is set
      const enabled = options.disable ? false : true;
      
      // Create the model configuration
      const modelConfig = {
        provider,
        modelId,
        enabled,
        ...(modelOptions ? { options: modelOptions } : {}),
        ...(options.apiKeyEnv ? { apiKeyEnvVar: options.apiKeyEnv } : {}),
      };
      
      // Add or update the model in the configuration
      const updatedConfig = loadConfig.addOrUpdateModel(config, modelConfig);
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Determine if this was an add or update operation
      const existingModel = loadConfig.findModel(config, provider, modelId);
      const operation = existingModel ? 'updated' : 'added';
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully ${operation} model ${colors.cyan(`${provider}:${modelId}`)}`)
      );
      
      // Show model details
      // eslint-disable-next-line no-console
      console.log('Model details:');
      // eslint-disable-next-line no-console
      console.log(`  Provider: ${colors.cyan(provider)}`);
      // eslint-disable-next-line no-console
      console.log(`  Model ID: ${colors.cyan(modelId)}`);
      // eslint-disable-next-line no-console
      console.log(`  Status: ${enabled ? colors.green('Enabled') : colors.red('Disabled')}`);
      
      if (modelOptions) {
        // eslint-disable-next-line no-console
        console.log('  Options:');
        for (const [key, value] of Object.entries(modelOptions)) {
          // eslint-disable-next-line no-console
          console.log(`    ${key}: ${JSON.stringify(value)}`);
        }
      }
      
      if (options.apiKeyEnv) {
        // eslint-disable-next-line no-console
        console.log(`  API Key Environment Variable: ${colors.cyan(options.apiKeyEnv)}`);
      }
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Remove model command
const removeModelCommand = new Command('remove');
removeModelCommand
  .description('Remove a model definition')
  .argument('<identifier>', 'Model identifier (provider:modelId or index number from list)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-f, --force', 'Force removal without confirmation')
  .action(async (identifier: string, options: { 
    config?: string;
    force?: boolean;
  }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Parse the identifier (either provider:modelId or an index number)
      let provider: string;
      let modelId: string;
      
      if (identifier.includes(':')) {
        // Format is provider:modelId
        [provider, modelId] = identifier.split(':');
        
        if (!provider || !modelId) {
          throw new loadConfig.ConfigError(
            `Invalid model identifier: ${identifier}. Expected format: provider:modelId`
          );
        }
      } else {
        // Try to parse as an index number
        const index = parseInt(identifier, 10);
        if (isNaN(index) || index < 1 || index > config.models.length) {
          throw new loadConfig.ConfigError(
            `Invalid model index: ${identifier}. Must be a number between 1 and ${config.models.length}`
          );
        }
        
        // Get the model at the specified index (1-based for user, 0-based for array)
        const model = config.models[index - 1];
        provider = model.provider;
        modelId = model.modelId;
      }
      
      // Find the model to verify it exists
      const model = loadConfig.findModel(config, provider, modelId);
      if (!model) {
        throw new loadConfig.ConfigError(`Model ${provider}:${modelId} not found in configuration`);
      }
      
      // Warn if the model is part of any groups (but still remove it with --force)
      let groupsContainingModel: string[] = [];
      if (config.groups) {
        groupsContainingModel = Object.entries(config.groups)
          .filter(([_, group]) => 
            group.models.some(m => 
              m.provider === provider && m.modelId === modelId
            )
          )
          .map(([name]) => name);
      }
      
      const isInGroups = groupsContainingModel.length > 0;
      if (isInGroups && !options.force) {
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow(
            `Warning: Model ${colors.cyan(`${provider}:${modelId}`)} is used in the following groups: ` +
            `${groupsContainingModel.join(', ')}`
          )
        );
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow(
            'Use --force to remove this model from all groups and the configuration'
          )
        );
        return;
      }
      
      // Remove the model
      const updatedConfig = loadConfig.removeModel(config, provider, modelId);
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully removed model ${colors.cyan(`${provider}:${modelId}`)}`)
      );
      
      // Show additional details if relevant
      if (isInGroups) {
        // eslint-disable-next-line no-console
        console.log(
          colors.dim(
            `The model was also removed from the following groups: ${groupsContainingModel.join(', ')}`
          )
        );
      }
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Enable model command
const enableModelCommand = new Command('enable');
enableModelCommand
  .description('Enable a model')
  .argument('<identifier>', 'Model identifier (provider:modelId or index number from list)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (identifier: string, options: { config?: string }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Parse the identifier (either provider:modelId or an index number)
      let provider: string;
      let modelId: string;
      
      if (identifier.includes(':')) {
        // Format is provider:modelId
        [provider, modelId] = identifier.split(':');
        
        if (!provider || !modelId) {
          throw new loadConfig.ConfigError(
            `Invalid model identifier: ${identifier}. Expected format: provider:modelId`
          );
        }
      } else {
        // Try to parse as an index number
        const index = parseInt(identifier, 10);
        if (isNaN(index) || index < 1 || index > config.models.length) {
          throw new loadConfig.ConfigError(
            `Invalid model index: ${identifier}. Must be a number between 1 and ${config.models.length}`
          );
        }
        
        // Get the model at the specified index (1-based for user, 0-based for array)
        const model = config.models[index - 1];
        provider = model.provider;
        modelId = model.modelId;
      }
      
      // Find the model to verify it exists
      const model = loadConfig.findModel(config, provider, modelId);
      if (!model) {
        throw new loadConfig.ConfigError(`Model ${provider}:${modelId} not found in configuration`);
      }
      
      // Check if the model is already enabled
      if (model.enabled) {
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow(
            `Info: Model ${colors.cyan(`${provider}:${modelId}`)} is already enabled`
          )
        );
        return;
      }
      
      // Update the model
      const updatedConfig = loadConfig.addOrUpdateModel(config, {
        ...model,
        enabled: true
      });
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully enabled model ${colors.cyan(`${provider}:${modelId}`)}`)
      );
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Disable model command
const disableModelCommand = new Command('disable');
disableModelCommand
  .description('Disable a model')
  .argument('<identifier>', 'Model identifier (provider:modelId or index number from list)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (identifier: string, options: { config?: string }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Parse the identifier (either provider:modelId or an index number)
      let provider: string;
      let modelId: string;
      
      if (identifier.includes(':')) {
        // Format is provider:modelId
        [provider, modelId] = identifier.split(':');
        
        if (!provider || !modelId) {
          throw new loadConfig.ConfigError(
            `Invalid model identifier: ${identifier}. Expected format: provider:modelId`
          );
        }
      } else {
        // Try to parse as an index number
        const index = parseInt(identifier, 10);
        if (isNaN(index) || index < 1 || index > config.models.length) {
          throw new loadConfig.ConfigError(
            `Invalid model index: ${identifier}. Must be a number between 1 and ${config.models.length}`
          );
        }
        
        // Get the model at the specified index (1-based for user, 0-based for array)
        const model = config.models[index - 1];
        provider = model.provider;
        modelId = model.modelId;
      }
      
      // Find the model to verify it exists
      const model = loadConfig.findModel(config, provider, modelId);
      if (!model) {
        throw new loadConfig.ConfigError(`Model ${provider}:${modelId} not found in configuration`);
      }
      
      // Check if the model is already disabled
      if (!model.enabled) {
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow(
            `Info: Model ${colors.cyan(`${provider}:${modelId}`)} is already disabled`
          )
        );
        return;
      }
      
      // Update the model
      const updatedConfig = loadConfig.addOrUpdateModel(config, {
        ...model,
        enabled: false
      });
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully disabled model ${colors.cyan(`${provider}:${modelId}`)}`)
      );
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Add a groups command group
const groupsCommand = new Command('groups');
groupsCommand.description('Manage model group configurations');

// List groups command
const listGroupsCommand = new Command('list');
listGroupsCommand
  .description('List all configured groups')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-j, --json', 'Output as JSON')
  .option('-v, --verbose', 'Show detailed group information including full system prompts')
  .action(async (options: { 
    config?: string;
    json?: boolean;
    verbose?: boolean;
  }) => {
    try {
      // Load the configuration
      const config = await loadConfig.loadConfig({ configPath: options.config });
      
      // If JSON output is requested, display the groups as JSON
      if (options.json) {
        // eslint-disable-next-line no-console
        const result: AppConfig = {
          models: config.models,
          groups: config.groups || {}
        };
        // eslint-disable-next-line no-console
        console.log(JSON.stringify(result, null, 2));
        return;
      }
      
      // Check if groups exist
      if (!config.groups || Object.keys(config.groups).length === 0) {
        // eslint-disable-next-line no-console
        console.log(colors.yellow('No model groups configured.'));
        return;
      }
      
      // Format and display the groups
      // eslint-disable-next-line no-console
      console.log(colors.blue('📂 Configured Groups:'));
      
      let index = 1;
      for (const [groupName, group] of Object.entries(config.groups)) {
        // eslint-disable-next-line no-console
        console.log(`\n${index}. ${colors.green(groupName)}${group.description ? ` - ${group.description}` : ''}`);
        
        // Show model count and names
        const modelCount = group.models.length;
        // eslint-disable-next-line no-console
        console.log(`   Models: ${colors.yellow(modelCount.toString())}`);
        
        if (modelCount > 0) {
          // Group models by provider for cleaner display
          const modelsByProvider: Record<string, string[]> = {};
          
          group.models.forEach(model => {
            if (!modelsByProvider[model.provider]) {
              modelsByProvider[model.provider] = [];
            }
            modelsByProvider[model.provider].push(model.modelId);
          });
          
          // Display models grouped by provider
          Object.entries(modelsByProvider).forEach(([provider, modelIds]) => {
            const modelsList = modelIds.join(', ');
            // eslint-disable-next-line no-console
            console.log(`   ${colors.cyan(provider)}: ${colors.dim(modelsList)}`);
          });
        } else {
          // eslint-disable-next-line no-console
          console.log(`   ${colors.dim('No models in this group')}`);
        }
        
        // Display system prompt if available
        if (group.systemPrompt && group.systemPrompt.text) {
          // In verbose mode, show the full system prompt
          if (options.verbose) {
            // eslint-disable-next-line no-console
            console.log(`\n   System Prompt: "${group.systemPrompt.text}"`);
          } else {
            // Otherwise truncate it
            const promptText = group.systemPrompt.text.length > 50
              ? `${group.systemPrompt.text.substring(0, 50)}...`
              : group.systemPrompt.text;
            
            // eslint-disable-next-line no-console
            console.log(`   System Prompt: ${colors.dim(`"${promptText}"`)}`);
          }
        } else {
          // eslint-disable-next-line no-console
          console.log(`   ${colors.dim('No system prompt defined')}`);
        }
        
        index++;
      }
      
      // Show tips for managing groups
      // eslint-disable-next-line no-console
      console.log(colors.dim('\nTips:'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use --verbose to see full system prompts'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use --json for machine-readable output'));
      // eslint-disable-next-line no-console
      console.log(colors.dim('  • Use "config groups create" to create a new group'));
    } catch (error) {
      handleError(error);
    }
  });

// Create group command
const createGroupCommand = new Command('create');
createGroupCommand
  .description('Create a new model group')
  .argument('<groupName>', 'Name for the new group')
  .option('-p, --prompt <text>', 'System prompt text for the group')
  .option('-m, --models <models>', 'Comma-separated list of models (provider:modelId) to include')
  .option('-d, --description <text>', 'Description of the group')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (groupName: string, options: { 
    prompt?: string;
    models?: string;
    description?: string;
    config?: string;
  }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Check if the group already exists
      if (config.groups && config.groups[groupName]) {
        throw new loadConfig.ConfigError(
          `Group "${groupName}" already exists. Use the update command to modify it.`
        );
      }
      
      // Create system prompt
      const systemPrompt = options.prompt 
        ? { text: options.prompt } 
        : { 
            text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.'
          };
      
      // Parse models if provided
      const modelsList: string[] = options.models 
        ? options.models.split(',')
        : [];
      
      // Validate and find models
      const groupModels = [];
      for (const modelIdentifier of modelsList) {
        // Split identifier into provider and modelId
        const [provider, modelId] = modelIdentifier.split(':');
        
        if (!provider || !modelId) {
          throw new loadConfig.ConfigError(
            `Invalid model identifier: ${modelIdentifier}. Expected format: provider:modelId`
          );
        }
        
        // Find the model in configuration
        const model = loadConfig.findModel(config, provider, modelId);
        if (!model) {
          throw new loadConfig.ConfigError(
            `Model ${provider}:${modelId} not found in configuration.\n` +
            `Add it first with: thinktank config models add ${provider} ${modelId}`
          );
        }
        
        // Add the model to the group
        groupModels.push(model);
      }
      
      // Create the group
      const updatedConfig = loadConfig.addOrUpdateGroup(config, groupName, {
        systemPrompt,
        models: groupModels,
        description: options.description
      });
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully created group "${colors.cyan(groupName)}"`)
      );
      
      // Show group details
      // eslint-disable-next-line no-console
      console.log('Group details:');
      // eslint-disable-next-line no-console
      console.log(`  Name: ${colors.cyan(groupName)}`);
      
      if (options.description) {
        // eslint-disable-next-line no-console
        console.log(`  Description: ${options.description}`);
      }
      
      // Show truncated system prompt
      // eslint-disable-next-line no-console
      console.log(`  System Prompt: "${systemPrompt.text.substring(0, 50)}${systemPrompt.text.length > 50 ? '...' : ''}"`);
      
      // Show models if any
      if (groupModels.length > 0) {
        // eslint-disable-next-line no-console
        console.log(`  Models: ${groupModels.length}`);
        
        // Group models by provider for cleaner display
        const modelsByProvider: Record<string, string[]> = {};
        
        groupModels.forEach(model => {
          if (!modelsByProvider[model.provider]) {
            modelsByProvider[model.provider] = [];
          }
          modelsByProvider[model.provider].push(model.modelId);
        });
        
        // Display models grouped by provider
        Object.entries(modelsByProvider).forEach(([provider, modelIds]) => {
          const modelsList = modelIds.join(', ');
          // eslint-disable-next-line no-console
          console.log(`    ${colors.cyan(provider)}: ${colors.dim(modelsList)}`);
        });
      } else {
        // eslint-disable-next-line no-console
        console.log(`  Models: ${colors.yellow('None')}`);
        // eslint-disable-next-line no-console
        console.log(colors.dim('  Use "config groups add-model" to add models to this group'));
      }
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Add model to group command
const addModelToGroupCommand = new Command('add-model');
addModelToGroupCommand
  .description('Add a model to a group')
  .argument('<groupName>', 'Name of the group')
  .argument('<modelId>', 'Model identifier (provider:modelId)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (groupName: string, modelIdentifier: string, options: { config?: string }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Check if the group exists
      if (!config.groups || !config.groups[groupName]) {
        throw new loadConfig.ConfigError(
          `Group "${groupName}" not found in configuration`
        );
      }
      
      // Parse the model identifier
      const [provider, modelId] = modelIdentifier.split(':');
      
      if (!provider || !modelId) {
        throw new loadConfig.ConfigError(
          `Invalid model identifier: ${modelIdentifier}. Expected format: provider:modelId`
        );
      }
      
      // Find the model in configuration
      const model = loadConfig.findModel(config, provider, modelId);
      if (!model) {
        throw new loadConfig.ConfigError(
          `Model ${provider}:${modelId} not found in configuration.\n` +
          `Add it first with: thinktank config models add ${provider} ${modelId}`
        );
      }
      
      // Check if the model is already in the group
      const isInGroup = config.groups[groupName].models.some(
        m => m.provider === provider && m.modelId === modelId
      );
      
      if (isInGroup) {
        // eslint-disable-next-line no-console
        console.log(
          colors.yellow(
            `Info: Model ${colors.cyan(`${provider}:${modelId}`)} is already in group "${colors.cyan(groupName)}"`
          )
        );
        return;
      }
      
      // Add the model to the group
      const updatedConfig = loadConfig.addModelToGroup(config, groupName, provider, modelId);
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully added model ${colors.cyan(`${provider}:${modelId}`)} to group "${colors.cyan(groupName)}"`)      
      );
      
      // Show updated group information
      const group = updatedConfig.groups![groupName];
      // eslint-disable-next-line no-console
      console.log(`Group now has ${colors.yellow(group.models.length.toString())} models:`);      
      
      // Group models by provider for cleaner display
      const modelsByProvider: Record<string, string[]> = {};
      
      group.models.forEach(model => {
        if (!modelsByProvider[model.provider]) {
          modelsByProvider[model.provider] = [];
        }
        modelsByProvider[model.provider].push(model.modelId);
      });
      
      // Display models grouped by provider
      Object.entries(modelsByProvider).forEach(([provider, modelIds]) => {
        const modelsList = modelIds.join(', ');
        // eslint-disable-next-line no-console
        console.log(`  ${colors.cyan(provider)}: ${colors.dim(modelsList)}`);
      });
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Remove model from group command
const removeModelFromGroupCommand = new Command('remove-model');
removeModelFromGroupCommand
  .description('Remove a model from a group')
  .argument('<groupName>', 'Name of the group')
  .argument('<modelId>', 'Model identifier (provider:modelId)')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (groupName: string, modelIdentifier: string, options: { config?: string }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Check if the group exists
      if (!config.groups || !config.groups[groupName]) {
        throw new loadConfig.ConfigError(
          `Group "${groupName}" not found in configuration`
        );
      }
      
      // Parse the model identifier
      const [provider, modelId] = modelIdentifier.split(':');
      
      if (!provider || !modelId) {
        throw new loadConfig.ConfigError(
          `Invalid model identifier: ${modelIdentifier}. Expected format: provider:modelId`
        );
      }
      
      // Check if the model is in the group
      const isInGroup = config.groups[groupName].models.some(
        m => m.provider === provider && m.modelId === modelId
      );
      
      if (!isInGroup) {
        throw new loadConfig.ConfigError(
          `Model ${provider}:${modelId} is not in group "${groupName}"`
        );
      }
      
      // Remove the model from the group
      const updatedConfig = loadConfig.removeModelFromGroup(config, groupName, provider, modelId);
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully removed model ${colors.cyan(`${provider}:${modelId}`)} from group "${colors.cyan(groupName)}"`)      
      );
      
      // Show updated group information
      const group = updatedConfig.groups![groupName];
      
      if (group.models.length === 0) {
        // eslint-disable-next-line no-console
        console.log(`Group now has ${colors.yellow('no')} models`);
      } else {
        // eslint-disable-next-line no-console
        console.log(`Group now has ${colors.yellow(group.models.length.toString())} models:`);
        
        // Group models by provider for cleaner display
        const modelsByProvider: Record<string, string[]> = {};
        
        group.models.forEach(model => {
          if (!modelsByProvider[model.provider]) {
            modelsByProvider[model.provider] = [];
          }
          modelsByProvider[model.provider].push(model.modelId);
        });
        
        // Display models grouped by provider
        Object.entries(modelsByProvider).forEach(([provider, modelIds]) => {
          const modelsList = modelIds.join(', ');
          // eslint-disable-next-line no-console
          console.log(`  ${colors.cyan(provider)}: ${colors.dim(modelsList)}`);
        });
      }
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Set group system prompt command
const setGroupPromptCommand = new Command('set-prompt');
setGroupPromptCommand
  .description('Set or update a group\'s system prompt')
  .argument('<groupName>', 'Name of the group')
  .requiredOption('-p, --prompt <text>', 'System prompt text')
  .option('-c, --config <path>', 'Path to a custom config file')
  .action(async (groupName: string, options: { 
    prompt: string;
    config?: string;
  }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Check if the group exists
      if (!config.groups || !config.groups[groupName]) {
        throw new loadConfig.ConfigError(
          `Group "${groupName}" not found in configuration`
        );
      }
      
      // Create the system prompt
      const systemPrompt = { text: options.prompt };
      
      // Update the group with the new system prompt
      const updatedConfig = loadConfig.addOrUpdateGroup(config, groupName, {
        ...config.groups[groupName],
        systemPrompt
      });
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully updated system prompt for group "${colors.cyan(groupName)}"`)      
      );
      
      // Show the new prompt
      // eslint-disable-next-line no-console
      console.log(`New system prompt: "${options.prompt}"`);
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Remove group command
const removeGroupCommand = new Command('remove');
removeGroupCommand
  .description('Delete a model group')
  .argument('<groupName>', 'Name of the group to remove')
  .option('-c, --config <path>', 'Path to a custom config file')
  .option('-f, --force', 'Force removal of the default group')
  .action(async (groupName: string, options: { 
    config?: string;
    force?: boolean;
  }) => {
    try {
      // Load the existing configuration
      const configPath = options.config || await loadConfig.getActiveConfigPath();
      const config = await loadConfig.loadConfig({ configPath });
      
      // Check if the group exists
      if (!config.groups || !config.groups[groupName]) {
        throw new loadConfig.ConfigError(
          `Group "${groupName}" not found in configuration`
        );
      }
      
      // Special handling for default group
      if (groupName === 'default' && !options.force) {
        throw new loadConfig.ConfigError(
          `Cannot remove the default group. Use --force if you really need to remove it.`
        );
      }
      
      // Count models in the group for reporting
      const modelCount = config.groups[groupName].models.length;
      
      // Remove the group
      const updatedConfig = loadConfig.removeGroup(config, groupName);
      
      // Save the updated configuration
      await loadConfig.saveConfig(updatedConfig, configPath);
      
      // Display success message
      // eslint-disable-next-line no-console
      console.log(
        colors.green(`Successfully removed group "${colors.cyan(groupName)}"`)      
      );
      
      // Show additional details if relevant
      if (modelCount > 0) {
        // eslint-disable-next-line no-console
        console.log(
          colors.dim(
            `The group contained ${modelCount} model${modelCount === 1 ? '' : 's'}. ` +
            `These models are still available in the configuration.`
          )
        );
      }
      
      // Show configuration file path
      // eslint-disable-next-line no-console
      console.log(colors.dim(`\nConfiguration saved to: ${configPath}`));
    } catch (error) {
      handleError(error);
    }
  });

// Add all the model subcommands to the models command
modelsCommand.addCommand(listModelsCommand);
modelsCommand.addCommand(addModelCommand);
modelsCommand.addCommand(removeModelCommand);
modelsCommand.addCommand(enableModelCommand);
modelsCommand.addCommand(disableModelCommand);

// Add all the group subcommands to the groups command
groupsCommand.addCommand(listGroupsCommand);
groupsCommand.addCommand(createGroupCommand);
groupsCommand.addCommand(addModelToGroupCommand);
groupsCommand.addCommand(removeModelFromGroupCommand);
groupsCommand.addCommand(setGroupPromptCommand);
groupsCommand.addCommand(removeGroupCommand);

// Add the main command groups to the config command
configCommand.addCommand(modelsCommand);
configCommand.addCommand(groupsCommand);

export default configCommand;