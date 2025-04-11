**Task: Update CLI Tests (`cmd/architect/cli_test.go`)**

**Completed:** April 10, 2025

**Summary:**
Updated the CLI tests to align with the new instructions-based design:
- Replaced tests for `--task-file` with tests for the new `--instructions` flag
- Added a test for dry run mode without an instructions file
- Added a comprehensive test for the usage message to verify it contains the correct flags and examples
- Added a test for advanced configuration options
- Removed the template example flags tests
- Updated validation tests to verify instructions file requirement
- Ensured test coverage for all core CLI functionality

This work ensures that the test suite correctly validates the new CLI interface and behavior after removal of the template-based functionality.