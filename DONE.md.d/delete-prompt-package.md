**Task: Delete Prompt Package (`internal/prompt/`)**

**Completed:** April 10, 2025

**Summary:**
Deleted the entire `internal/prompt/` directory, which contained the templating functionality that is being replaced with a simpler instructions-based approach. This included:

- Removal of all Go source files:
  - `config_adapter.go` and `config_adapter_test.go`
  - `embed.go`
  - `integration.go`
  - `mock_manager.go`
  - `prompt.go`, `prompt_test.go`, and `prompt_examples_test.go`

- Removal of all template files:
  - Main templates: `custom.tmpl`, `default.tmpl`, `test.tmpl`
  - Example templates: `basic.tmpl`, `bugfix.tmpl`, `detailed.tmpl`, `feature.tmpl`

This is an intentional breaking change as part of the larger refactoring effort. The codebase will not compile at this stage, which is expected. Subsequent tasks will address the dependencies and references to this package throughout the rest of the codebase.