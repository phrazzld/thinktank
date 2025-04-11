# Update Core CliConfig Types

- **Completed:** 2025-04-10
- **Action:** Modified the `CliConfig` struct in both `internal/architect/types.go` and `cmd/architect/cli.go`. Removed fields: `TaskDescription`, `Template`, `ListExamples`, `ShowExample` from `internal/architect/types.go` and `TaskFile`, `PromptTemplate`, `ListExamples`, `ShowExample` from `cmd/architect/cli.go`. Added field: `InstructionsFile string` to both.
- **AC Ref:** Plan Section 2 (Type Definitions), Plan Section 4 (CLIConfig Structs)