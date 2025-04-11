**Task: Remove Template Fields from Config (`internal/config/`)**

**Completed:** April 10, 2025

**Summary:**
Modified the configuration system to remove template-related fields as part of the transition to a simpler instructions-based design:

- Removed the `TemplateConfig` struct definition entirely
- Removed template-related fields from other structures:
  - Removed `Templates TemplateConfig` field from the `AppConfig` struct
  - Removed `UserTemplateDir` and `SystemTemplateDirs` fields from the `ConfigDirectories` struct
- Updated the `DefaultConfig` function to remove template initialization

This change is part of the larger refactoring effort to simplify the application's design by replacing the template-based approach with a direct instructions-based model. While this change intentionally breaks compilation, subsequent tasks will address the dependencies and references to ensure a complete transition.