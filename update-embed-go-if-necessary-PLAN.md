# Update embed.go if necessary

## Goal
Examine the `internal/prompt/embed.go` file to determine if it contains any specific references to the clarify template files that need to be updated, ensuring proper handling of the embed functionality after the removal of clarify-related templates.

## Implementation Approach
1. Review the current implementation of `internal/prompt/embed.go` to understand how templates are embedded
2. Check if there are any explicit references to clarify template files that need to be removed
3. If specific references exist, update them to maintain proper functionality
4. If the file uses wildcard patterns (like *.tmpl) that handle missing files gracefully, no changes may be necessary
5. Ensure the embed functionality continues to work correctly after any changes

## Reasoning
This approach directly examines the embed.go file to determine if changes are needed based on how templates are referenced. Go's embed package typically handles missing files gracefully when using wildcard patterns, but we need to verify there are no explicit references to the removed template files.

Alternative approaches considered:
1. **Automatically update the embed.go file**: We could have just assumed changes were needed without checking the current implementation, but this could lead to unnecessary modifications that might introduce bugs.

2. **Do nothing**: We could have assumed the embed pattern would handle missing files and skip this task, but this would risk leaving broken references in the code.

The chosen approach is the most careful and comprehensive. By examining the current implementation of the embed.go file, we can make an informed decision about whether changes are needed, ensuring that the code continues to function correctly while removing all references to the clarify feature.