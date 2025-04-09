# Remove NoSpinner field from AppConfig struct

## Task Goal
Remove the `NoSpinner` boolean field from the `AppConfig` struct definition in internal/config/config.go as part of the ongoing spinner removal process.

## Implementation Approach
I will use a straightforward field removal approach with careful validation:

1. Remove the `NoSpinner` field from the `AppConfig` struct in internal/config/config.go
2. Remove the NoSpinner field initialization in the DefaultConfig() function
3. Remove the NoSpinner field from the setViperDefaults function
4. Verify tests continue to pass to ensure nothing breaks with this change

## Reasoning for Selected Approach
This is a simple, direct approach that focuses on removing the field while ensuring tests continue to pass. I considered other approaches:

1. **Field Removal Only**: Just remove the field without addressing DefaultConfig() and setViperDefaults(). This would be risky as it could lead to runtime errors when the code tries to access or set a non-existent field.

2. **Comprehensive Refactoring**: Completely refactor all code that might reference NoSpinner. This is excessive since we're already addressing references in a step-by-step manner through other tasks.

3. **Field Deprecation**: Keep the field but mark it as deprecated. This doesn't align with the task goal of complete spinner removal.

The chosen approach provides a good balance - it completely removes the field while ensuring we catch any immediate issues through test validation. The task is part of a larger effort to remove all spinner functionality, so this specific change should be focused on the struct definition in config.go.