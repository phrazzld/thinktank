**Task: Delete Prompt Package (`internal/prompt/`)**

## Goal
Remove the entire prompt package which contains the templating functionality that is being replaced with the simpler instructions-based approach. This includes all Go files, template files, and tests within the `internal/prompt/` directory.

## Implementation Approach
1. First, examine the contents of the `internal/prompt/` directory to understand its structure and contents:
   - Identify all `.go` files, `.tmpl` files, and subdirectories
   - Note any templates that may need to be preserved for reference or documentation

2. Determine if there are any test fixtures or mocks that might be used by other packages

3. Delete the entire directory:
   - Use `git rm -r internal/prompt/` to remove it from the repository
   - This will stage the deletion for the next commit

## Reasoning
This approach directly addresses the task requirements in a straightforward manner. Since the templating system is being completely replaced with a simpler instructions-based approach, we need to remove all code related to the old system.

Removing the entire package in one step is appropriate because:
1. The code is being replaced, not refactored, so keeping any parts would lead to confusion
2. It allows us to make a clean break from the old system
3. It reduces the chances of leaving behind unused code that might cause issues later

This is an intentional breaking change as part of a larger refactoring effort. The codebase will not compile after this step, which is expected. Subsequent tasks will address the dependencies and references to this package throughout the rest of the codebase.