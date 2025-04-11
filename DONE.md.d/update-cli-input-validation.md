# Update CLI Input Validation (`cmd/architect/cli.go`)

- **Completed:** 2025-04-10
- **Action:** Modified `ValidateInputs` function to remove the check for `TaskFile` and the validation bypass logic related to example commands. Added a check ensuring `InstructionsFile` is provided (unless in `--dry-run` mode) with appropriate error messages.
- **AC Ref:** Plan Section 3 (Task 1)