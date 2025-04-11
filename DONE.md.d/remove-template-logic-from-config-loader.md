**Task: Remove Template Logic from Config Loader (`internal/config/loader.go`)**

**Completed:** April 10, 2025

**Summary:**
Removed all template-related logic from the configuration loader to support the transition to a simpler instructions-based design:

- Removed the template-specific methods and functions:
  - Removed the `GetTemplatePath` function which was used to locate template files
  - Removed the `getTemplatePathFromConfig` helper function 
  - Removed the `GetUserTemplateDir` and `GetSystemTemplateDirs` functions
  
- Updated the `GetConfigDirs` function to return only the core configuration directories

- Updated the `setViperDefaults` function to remove template-related default settings

- Updated the `MergeWithFlags` function to remove template-specific mapping logic

- Simplified the `EnsureConfigDirs` function to no longer create template directories

- Updated the initialization message to remove template-related output

This change continues the refactoring process to replace the template-based approach with an instructions-based model. As with previous tasks, this intentionally breaks compilation, which is expected and will be addressed as we complete the remaining refactoring tasks.