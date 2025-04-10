# Extract prompt building functions to prompt.go

## Task Description
Move the prompt building functions (`buildPrompt`, `buildPromptWithConfig`, and `buildPromptWithManager`) from main.go to cmd/architect/prompt.go, maintaining the same functionality while ensuring testability.

## Relevant Context

### From TODO.md:
- **Action:** Move buildPrompt, buildPromptWithConfig, and buildPromptWithManager functions from main.go to prompt.go.
- **Depends On:** Create skeleton files in cmd/architect/.
- **AC Ref:** Implementation Steps 5.

### Current State:
- The skeleton file `cmd/architect/prompt.go` has been created with stub implementations of the PromptBuilder interface.
- The functions to extract currently exist in main.go:
  - `buildPrompt` - A simple wrapper that calls buildPromptWithManager with a basic prompt manager
  - `buildPromptWithConfig` - Creates a prompt manager with config support and calls buildPromptWithManager
  - `buildPromptWithManager` - Core implementation that handles prompt template loading and data binding

### Dependencies:
- These functions depend on the `internal/prompt` package, which provides the `ManagerInterface` and template handling functionality.
- The Configuration struct from main.go contains prompt template configuration.

## Requirements

1. Move the identified functions to cmd/architect/prompt.go, implementing the appropriate methods on the PromptBuilder interface.
2. Maintain the same behavior as the original functions.
3. Follow the testability principles from TESTING_PHILOSOPHY.md, particularly:
   - Design for testability from the start
   - Prefer testing behavior over implementation
   - Minimize mocking, especially of internal components
   - Focus on clear interfaces and separation of concerns
4. Consider how to handle the Configuration struct dependency - should it be passed directly, or should specific parameters be extracted?
5. Update main.go to use the new implementation.
6. Add appropriate transitional comments to the original functions in main.go.

## Considerations for Implementation Approaches

Please provide 2-3 different approaches for implementing this task, considering:

1. Interface design and parameter passing
2. Management of dependencies (Configuration, logger, etc.)
3. Error handling
4. Testing strategy

For each approach, explain the pros and cons, and recommend the best approach based on the project's testing philosophy and design principles.