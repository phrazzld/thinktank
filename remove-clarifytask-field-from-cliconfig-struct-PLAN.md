# Remove ClarifyTask field from CliConfig struct

## Goal
Remove the `ClarifyTask bool` field from the `CliConfig` struct in `cmd/architect/cli.go` to continue the incremental removal of the clarify feature from the codebase.

## Chosen Implementation Approach
I'll implement this task using the following targeted approach:

1. Examine the `CliConfig` struct definition in `cmd/architect/cli.go` to confirm the exact location and context of the `ClarifyTask` field.
2. Remove the `ClarifyTask bool` field from the struct definition while keeping all other fields intact.
3. Identify and handle any immediate compilation issues caused by removing this field:
   - Keep the temporary `clarifyTaskFlag` variable we added in the previous task to maintain compilation
   - Add a no-op assignment to maintain compatibility with the flag usage until subsequent tasks are completed
4. Verify the application builds successfully and all tests pass after this change.

## Reasoning for Approach
I chose this approach for the following reasons:

* **Incremental and controlled**: This approach continues the step-by-step removal process, focusing only on the struct field without disrupting other code paths yet. This makes potential issues easier to isolate and fix.

* **Maintains compilation**: While removing the field creates a direct incompatibility with code that assigns to or reads from this field, we'll add a temporary workaround to keep the code compiling until those references are addressed in subsequent tasks.

* **Follows dependency chain**: This approach aligns with the dependency structure in the TODO list, where "Remove clarify key-value pair from ConvertConfigToMap" depends on this task. We're preparing the ground for that task without getting ahead of ourselves.

Alternative approaches considered:
1. **Remove all related code at once**: This would be more efficient but riskier as it would make debugging more complex if anything went wrong.
2. **Refactor struct to avoid field removal**: We could have kept the field but made it unused. However, this would leave technical debt and confuse future developers about the field's purpose.

The chosen incremental approach strikes the right balance between progress and safety.