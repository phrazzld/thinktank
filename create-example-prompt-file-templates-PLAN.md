# Implementation Plan: Create Example Prompt File Templates

## Goal
Create example prompt file templates that users can reference when creating their own prompt files for use with the Architect CLI tool's `--task-file` flag, providing clear guidance on how to structure and format prompt files.

## Chosen Approach
**Approach 1: Embedded Examples with CLI Access**

After analyzing the thinktank output and considering all three approaches, I've selected Approach 1 which involves embedding the example templates in the binary and providing CLI commands to list and view them. 

This approach is preferred because:
1. It ensures examples are always distributed with the specific version of the tool
2. Doesn't rely on external dependencies or require users to find files online
3. Maintains version consistency between the tool and its examples
4. Leverages the existing `embed` mechanism already used for default templates
5. Provides the best testability compared to the other approaches

## Implementation Details

### 1. Create Example Template Files
Create the following example template files:
- `basic.tmpl`: A simple task file format with minimal formatting
- `detailed.tmpl`: A comprehensive task file with sections for requirements, context, and constraints
- `bugfix.tmpl`: A template specifically designed for bug fixes
- `feature.tmpl`: A template designed for new feature implementation

Each template will demonstrate proper use of the Architect tool's template variables and structure.

### 2. Update Prompt Package
- Add a new `examples` subdirectory under `internal/prompt/templates/`
- Modify the `embed.go` file to include the examples
- Enhance the `prompt.Manager` to support accessing and listing example templates

### 3. Add CLI Commands
- Add a new `list-examples` command to list available example templates
- Add a `show-example` command to display the content of a specific template

### 4. Update Documentation
- Update README.md with a new section explaining example templates
- Add usage examples showing how to list, view, and save examples
- Add clear documentation for the new CLI commands

### Testability Considerations
This approach offers good testability:
- Unit tests can verify that the list/show commands work correctly
- Tests can confirm that the expected templates are embedded and accessible
- The functionality is contained within the binary, making tests self-contained
- No external dependencies or filesystem operations to mock (beyond showing the templates)

The chosen approach avoids the filesystem complexity of Approach 3 and the documentation-only limitations of Approach 2, providing a robust, testable solution that ensures users always have access to quality examples appropriate for their installed version of the tool.