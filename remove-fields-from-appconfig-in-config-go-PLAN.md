# Remove fields from AppConfig in config.go

## Goal
Remove any `ClarifyTask` field from the AppConfig struct in the `internal/config/config.go` file as part of the ongoing effort to remove the clarify feature from the codebase.

## Implementation Approach
1. Examine the `internal/config/config.go` file to identify the AppConfig struct
2. Check if the AppConfig struct contains a `ClarifyTask` field
3. If present, remove the field and any direct references to it
4. Ensure backward compatibility for config loading by handling any legacy configurations
5. Verify that the code still compiles and tests pass
6. If needed, implement temporary compatibility measures to ensure the codebase continues to function correctly

## Reasoning
This approach is the most direct and straightforward method to remove the `ClarifyTask` field from the AppConfig struct. By examining the config.go file first, we can determine if the field exists and how it's used within the codebase.

Alternative approaches considered:
1. **Deprecate but keep the field**: While this would minimize risk, it doesn't fully remove the feature and leaves technical debt in the codebase, which doesn't align with the project's goal of completely removing the clarify feature.
2. **Add conditional loading**: We could implement logic that skips the field when loading. However, this adds complexity without providing significant benefits, especially since we're systematically removing all references to the feature.

The chosen approach aligns with the incremental strategy laid out in the project plan, focusing on removing one component at a time while ensuring the codebase remains functional. If we discover that other parts of the code depend on this field, we'll implement temporary workarounds to maintain compatibility until those dependencies can be addressed in subsequent tasks.