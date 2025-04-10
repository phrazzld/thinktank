# Remove ClarifyTask field from AppConfig struct in config.go

## Goal
Remove the `ClarifyTask` field from the `AppConfig` struct in `internal/config/config.go` including its mapstructure and toml tags, as part of eliminating the clarify flag and all related code from the Architect tool.

## Chosen Approach: Direct Removal
I'll implement a direct removal approach for this task:

1. Locate the `AppConfig` struct definition in `internal/config/config.go` (line 62)
2. Remove the line containing the `ClarifyTask bool \`mapstructure:"clarify_task" toml:"clarify_task"\`` field declaration
3. Remove the related default value setting in the `setViperDefaults()` method in `internal/config/loader.go` (line 262)
4. Make no additional changes at this step, focusing only on removing the field declaration and immediate initialization logic

This is an atomic change that addresses just the field removal from the struct and its immediate default value setting. The code will not compile immediately after this change, but subsequent tasks in TODO.md will address all references to this field.

## Reasoning for this Choice
I've chosen the direct removal approach for the following reasons:

1. **Atomicity**: This approach creates a clean, atomic change that's focused on a single logical modification, making it easy to review and test.

2. **Task Alignment**: It precisely follows the task breakdown in TODO.md, which separates field removal from updating references.

3. **Testability**: This approach is highly testable according to TESTING_PHILOSOPHY.MD:
   - It follows the "Simplicity & Clarity" principle with a straightforward removal
   - It aligns with "Testability is a Design Goal" by simplifying the AppConfig struct
   - It doesn't require complex mocking or setup to test - we simply need to verify the field no longer exists

4. **Best Practices**: It adheres to the "Conventional Commits & Atomic Changes" principle from BEST_PRACTICES.md by keeping the change small and focused on a single purpose.

5. **Consistency**: This approach is consistent with the previous task (removing ClarifyTask from Configuration struct), maintaining a consistent pattern throughout the implementation.

While this change on its own will cause compilation errors, that's expected and will be addressed by subsequent tasks. This approach maintains a clear step-by-step implementation path and makes it easier to verify each change independently.

The field removal is the logical next step in the process of removing the clarify functionality, and subsequent tasks will handle updating all references to this field to complete the removal process.