**Task: Remove Template Fields from Config (`internal/config/`)**

## Goal
Remove the template-related fields from the configuration system as part of transitioning from the template-based approach to a simpler instructions-based design. This includes removing the `Templates TemplateConfig` field from the `AppConfig` struct and updating the `DefaultConfig` function accordingly.

## Implementation Approach
1. Examine the current structure of `config.go`:
   - Identify template-related types, fields, and constants
   - Note the impact of their removal on other parts of the codebase

2. Remove the `TemplateConfig` struct definition entirely, as it will no longer be needed with the new instructions-based design.

3. Remove template-related fields from other structs:
   - Remove the `Templates TemplateConfig` field from the `AppConfig` struct
   - Remove template-related fields from the `ConfigDirectories` struct (`UserTemplateDir` and `SystemTemplateDirs`)

4. Update the `DefaultConfig` function:
   - Remove the initialization of the `Templates` field from the returned `AppConfig` instance

5. Leave TaskDescription and TaskFile in place for now. These will be removed in a coordinated step later, likely when we refactor the core application flow as they're currently still referenced in other parts of the application.

## Reasoning
This approach focuses narrowly on removing template-specific configuration while minimizing impact on the rest of the codebase. By removing the `TemplateConfig` type and related fields, we're taking a concrete step toward simplifying the application's design.

The reason for not removing all template-related implementation details at once is to allow for an incremental refactoring approach. This way, we can make targeted, focused changes that are easier to review and less likely to introduce errors. Other parts of the codebase that reference template configuration will be addressed in subsequent tasks, particularly when we remove template logic from the config loader.

This is an intentional breaking change as part of the larger refactoring effort. The codebase will not compile after this change, but that's expected and will be resolved as we complete subsequent tasks in the refactoring plan.