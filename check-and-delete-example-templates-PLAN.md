# Check and delete example templates

## Goal
Examine the example templates in the `internal/prompt/templates/examples/` directory for any files related to the clarify feature and delete them if found, as part of the ongoing effort to completely remove the clarify feature from the codebase.

## Implementation Approach
1. Check all template files in the `internal/prompt/templates/examples/` directory for any that are specifically related to the clarify feature or contain clarify-related content
2. If any clarify-related templates are found, delete them
3. Verify that removing these files doesn't impact the existing functionality
4. If no clarify-related templates are found, document this finding

## Reasoning
This approach directly addresses the task requirement by examining the example templates directory for any clarify-related files and removing them if found. This ensures that no references to the deprecated clarify feature remain in the example templates.

Since the removal of clarify-related templates is a simple, discrete action with a clear purpose (complete removal of the feature), there are no viable alternative approaches that would achieve the same goal. The only decision point is how to handle the case where no clarify-related templates are found, and in that case, we can simply document that finding.

This task is a necessary step in the complete removal of the clarify feature from the codebase. By ensuring that no example templates reference or demonstrate the clarify feature, we prevent users from attempting to use functionality that no longer exists.