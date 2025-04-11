**Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)**

## Goal
Remove all template-related logic from the configuration loader, which includes the `GetTemplatePath` function, templates directory handling, and any logic related to template names or settings within the configuration loading process.

## Implementation Approach
1. Remove the template-specific methods and functions:
   - Remove the `GetTemplatePath` function entirely
   - Remove the `getTemplatePathFromConfig` function
   - Remove the `GetUserTemplateDir` and `GetSystemTemplateDirs` functions 

2. Update the `GetConfigDirs` function to return the new `ConfigDirectories` struct without template directories

3. Update the `setViperDefaults` function:
   - Remove template-related default settings
   - Keep the defaults for all other configuration options

4. Update the `MergeWithFlags` function:
   - Remove the template-specific mapping logic
   - Keep the logic for other nested configuration fields like `excludes`

5. Update the `EnsureConfigDirs` function:
   - Remove the template directory creation
   - Keep the user config directory creation

6. Update the `displayInitializationMessage` function:
   - Remove references to templates in the output message

## Reasoning
This approach systematically removes all template-related logic from the configuration loader while preserving the core configuration loading functionality. By focusing only on template-related code, we maintain the structure and behavior of the configuration system for all other settings.

The removal of template-specific methods like `GetTemplatePath` is straightforward as they'll no longer be needed in the new instructions-based design. Similarly, removing the defaults for template settings in `setViperDefaults` ensures that no template configuration options are saved or loaded.

This change is part of the incremental refactoring strategy. As noted in the task, removal of these elements will temporarily break compilation, which is expected and will be resolved as we complete subsequent tasks in the refactoring plan.