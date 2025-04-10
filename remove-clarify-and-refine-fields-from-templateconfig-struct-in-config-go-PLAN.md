# Remove Clarify and Refine fields from TemplateConfig struct in config.go

## Goal
Remove the `Clarify` and `Refine` string fields from the `TemplateConfig` struct in `internal/config/config.go` as part of eliminating the clarify flag and all related code from the Architect tool.

## Chosen Approach: Direct Removal with Dependency Updates
I'll implement a comprehensive removal approach that addresses both the fields and their direct dependencies:

1. Remove the Clarify and Refine fields from the TemplateConfig struct in internal/config/config.go
2. Remove the corresponding default values in the DefaultConfig() function in config.go
3. Remove the corresponding SetDefault calls in the setViperDefaults() function in loader.go
4. Update the getTemplatePathFromConfig() function in loader.go to remove the case statements for "clarify" and "refine"

This approach ensures all direct references to these fields are removed in a single cohesive change, keeping the codebase in a compilable state after implementation.

## Reasoning for this Choice
I've chosen the direct removal with dependency updates approach for the following reasons:

1. **Completeness**: This approach provides a complete solution that addresses both the fields and their direct dependencies in a single change, making it more effective than simply removing the fields alone.

2. **Maintainability**: It keeps the codebase in a compilable state, which is more maintainable than leaving the code with broken references to removed fields.

3. **Testability**: This approach aligns well with the testing philosophy in TESTING_PHILOSOPHY.MD:
   - It follows the "Behavior Over Implementation" principle by ensuring the system still functions correctly after removing these fields
   - It supports the "Testability is a Design Goal" principle by simplifying the code structure
   - It doesn't introduce any new complexities that would require extensive mocking

4. **Cohesion**: While it spans multiple files, all the changes are focused on a single purpose (removing the template references for clarify/refine functionality) and don't introduce new complexity.

5. **Practical Development**: From a development perspective, making all the necessary changes in one go rather than requiring multiple commits to fix compilation errors is more efficient and reduces the chance of overlooking dependencies.

This approach represents a good balance between maintaining code quality, ensuring testability, and making progress on the overall goal of removing the clarify functionality. The changes form a cohesive unit that removes a specific piece of functionality while keeping the code in a working state.