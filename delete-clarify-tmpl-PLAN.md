# Delete clarify.tmpl

## Goal
Delete the file internal/prompt/templates/clarify.tmpl as part of the overall effort to remove the clarify functionality from the codebase.

## Implementation Approach
The implementation approach is straightforward - simply delete the file internal/prompt/templates/clarify.tmpl. 

The file has already been confirmed to exist at the specified path, and based on previous tasks in the project, references to this template file have already been removed from the codebase:

1. The "Update SetupPromptManagerWithConfig to remove clarify/refine templates" task has been completed, which modified internal/prompt/integration.go to only load default.tmpl and remove clarify.tmpl and refine.tmpl from the template loading loop.

2. The "Remove Clarify and Refine fields from TemplateConfig struct in config.go" task has been completed, which removed the Clarify and Refine string fields from the TemplateConfig struct.

3. The "Update DefaultConfig() to remove clarify template references" task has been completed, which removed Clarify and Refine template references from the DefaultConfig() function.

## Reasoning
This is a necessary step in the overall refactoring process to remove the clarify functionality. The file needs to be deleted because:

1. It's no longer referenced in the code since the previous refactoring steps have removed all references to it.
2. Keeping unused template files would create confusion for future developers.
3. Complete removal of all clarify-related artifacts is part of the project requirements.

The risk of this change is minimal since the code that used this template has already been removed, and we've verified through previous tasks that the template loading code no longer references this file.