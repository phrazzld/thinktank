# Remove ClarifyTask field from internal CliConfig

## Goal
Remove the `ClarifyTask` field from any configuration structures in the `internal/architect/types.go` file to continue the removal of the clarify feature from the codebase.

## Implementation Approach
1. Examine the `internal/architect/types.go` file to locate any configuration structures that contain a `ClarifyTask` field
2. Remove the `ClarifyTask` field from identified structures
3. Update any references to this field in other parts of the code
4. Ensure the code still compiles and all tests pass
5. If any code is relying on this field for functionality, implement an appropriate workaround to maintain compatibility until subsequent tasks can address the dependency

## Reasoning
This is the most direct approach for removing the field from internal configurations. By examining the types defined in the `types.go` file, we can identify and remove references to the deprecated feature. If we find code that depends on this field, we'll maintain temporary compatibility to ensure the codebase continues to work until other tasks can remove those dependencies.

Alternative approaches might include:
1. Marking the field as deprecated with a comment and leaving it in place - This is less desirable as it leaves technical debt and doesn't fully remove the feature.
2. Refactoring all dependent code at once - While more comprehensive, this would be riskier and goes against the incremental approach outlined in the task breakdown.

The selected approach aligns with the incremental, careful removal strategy in the project plan, minimizing risk while making steady progress toward eliminating the clarify feature.