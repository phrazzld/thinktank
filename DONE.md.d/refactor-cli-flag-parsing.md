# Refactor CLI Flag Parsing (`cmd/architect/cli.go`)

- **Completed:** 2025-04-10
- **Action:** Updated `ParseFlagsWithEnv` function: Removed definitions for `--task-file`, `--prompt-template`, `--list-examples`, `--show-example`. Added definition for a required `--instructions` string flag. Updated flag parsing logic to populate `InstructionsFile` in `CliConfig`. Removed population of deleted fields.
- **AC Ref:** Plan Section 3 (Task 1), Plan Section 4 (CLIConfig Structs)