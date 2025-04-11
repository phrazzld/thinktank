**Task: Update Config-Flag Mapping (`cmd/architect/cli.go`)**

**Completed:** April 10, 2025

**Summary:**
Modified the `ConvertConfigToMap` function in `cmd/architect/cli.go` to:
- Remove map keys related to removed template flags: `promptTemplate`, `listExamples`, `showExample`, `taskFile`
- Add the new `instructionsFile` key for configuration merging
- Update the associated test in `flags_test.go` to verify the correct mapping

This change aligns the configuration mapping with the previously implemented CLI flag changes, ensuring consistency throughout the codebase.