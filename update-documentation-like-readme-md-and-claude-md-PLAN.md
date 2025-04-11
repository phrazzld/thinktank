# Task: Update documentation, like README.md and CLAUDE.md

## Goal
Update the README.md and CLAUDE.md files to reflect the new architecture and command structure of the Architect tool, removing references to the old template-based approach and documenting the new XML-structured approach with the `--instructions` flag.

## Implementation Approach
I'll take a comprehensive approach to update both documents:

1. For README.md:
   - Update the "Important Update" section to reflect the switch from `--task-file` to `--instructions`
   - Revise the examples throughout the document to use the new flag
   - Remove all references to template-related flags (`--prompt-template`, `--list-examples`, `--show-example`)
   - Remove the "Custom Prompt Templates" section entirely
   - Add a new section explaining the XML-structured approach for instructions and context
   - Update the "Configuration Options" table to match the current flags

2. For CLAUDE.md:
   - Update the "Run" command example to use `--instructions` instead of `--task`

## Reasoning
This approach directly addresses the task requirements by ensuring all documentation accurately reflects the new architecture. Since the tool has undergone significant changes (removing templates in favor of XML structure, changing flag names), it's critical that the documentation is updated to prevent user confusion.

The README.md requires more extensive changes since it contains detailed examples and explanations of the removed template system. The CLAUDE.md file will need minimal changes as it only contains a basic command example.

I'm choosing to maintain the overall structure of the documents while making these targeted changes to ensure consistency with previous documentation style and format.