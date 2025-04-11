**Task: Update Main Entry Point Logic (`cmd/architect/main.go`)**

**Completed:** April 10, 2025

**Summary:**
Modified the `cmd/architect/main.go` file to align with the simplified instructions-based design:
- Updated the `convertToArchitectConfig` function to map the new `InstructionsFile` field between CLI and internal configs
- Removed mapping of obsolete template-related fields: `TaskDescription`, `TaskFile`, `Template`, `ListExamples`, `ShowExample`
- Removed the call to `architect.HandleSpecialCommands` which was used for template examples functionality

This change ensures that the main entry point correctly translates between the CLI and internal configuration models and removes template-related functionality from the execution flow.