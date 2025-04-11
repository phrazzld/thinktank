# Remove fields from TemplateConfig

## Goal
Remove the `Clarify` and `Refine` fields from the TemplateConfig struct in `internal/config/config.go` if they exist, as part of the ongoing effort to completely remove the clarify feature from the codebase.

## Implementation Approach
1. Examine the `internal/config/config.go` file to verify if the TemplateConfig struct contains `Clarify` and `Refine` fields
2. If present, remove these fields from the struct definition
3. Check for any direct references or usage of these fields in the codebase and update accordingly
4. Ensure the code continues to compile and tests still pass
5. If needed, implement temporary compatibility measures to handle legacy configurations that might reference these fields

## Reasoning
This straightforward approach directly addresses the task requirements by removing the specified fields from the TemplateConfig struct. The approach is clear and focused, targeting only the specific fields mentioned in the task while maintaining compatibility with the rest of the codebase.

Alternative approaches considered:
1. **Deprecate fields with comments but keep them**: This would involve marking the fields as deprecated but not actually removing them. While this would minimize immediate risk, it doesn't align with the goal of completely removing the clarify feature and would leave technical debt in the codebase.

2. **Rename fields to indicate deprecation**: Another approach could be to rename the fields to have a "deprecated" prefix. However, this still keeps the fields in the codebase and doesn't actually achieve the removal goal.

The chosen approach is superior because it completely removes the deprecated fields, aligning with the project's goal of purging the clarify feature from the codebase. It follows the incremental pattern established in previous tasks, which has proven successful, and it maintains the careful balance between removing deprecated code and ensuring the system continues to function properly.