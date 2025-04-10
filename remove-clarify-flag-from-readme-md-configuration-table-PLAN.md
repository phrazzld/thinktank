# Remove clarify flag from README.md configuration table

## Goal
Edit the README.md file to remove the `--clarify` entry from the configuration options table, as part of the overall effort to completely remove the clarify functionality from the codebase.

## Implementation Approach
The implementation is straightforward:

1. Locate the configuration options table in the README.md file
2. Remove the entire row that contains the `--clarify` flag description
3. Ensure that the markdown table formatting remains correct after the removal

Examining the README.md file, the relevant line to remove is line 97:
```markdown
| `--clarify` | Enable interactive task clarification to refine your task description | `false` |
```

This is the full entry in the configuration table for the clarify flag that needs to be removed.

## Reasoning
This change is necessary as part of the systematic removal of the clarify functionality from the codebase. Since all the code, function calls, and template files related to the clarify feature have been removed in previous tasks, we need to also update the documentation to ensure it accurately reflects the current state of the application.

Removing just this single line from the configuration table is the minimal change needed to achieve the task's goal. This approach:

1. Maintains the integrity of the rest of the documentation
2. Ensures that users won't try to use a flag that no longer exists
3. Is consistent with the overall goal of completely removing all references to the clarify functionality

Note that this is just one part of the documentation updates. There are separate tasks listed in the TODO.md to remove the "Task Clarification" section and clarify examples from the README.md, which will be handled in subsequent tasks.