# Remove ClarifyTask assignment in parseFlags()

## Task Goal
Remove the assignment to config.ClarifyTask in the parseFlags() function in main.go as part of the ongoing work to remove the clarify functionality from the codebase.

## Implementation Approach
After analyzing the codebase, I'll implement a direct removal approach with these steps:

1. Locate and remove the line in parseFlags() that assigns the clarifyTaskFlag value to config.ClarifyTask
2. The Configuration struct field has already been removed in a previous task (based on the TODO.md status)
3. The clarifyTaskFlag declaration has already been removed in a previous task
4. No new code needs to be added to replace this functionality

## Key Reasoning
This simple approach is best because:

1. **Isolated Change**: This is a straightforward removal that aligns with the incremental approach being taken to remove the clarify functionality
2. **Reduced Risk**: By focusing solely on removing the assignment without adding any replacement code, we minimize the chance of introducing new issues
3. **Consistency**: This approach maintains the pattern established in previous tasks of methodically removing clarify-related code
4. **Clean Codebase**: Removing unused code improves maintainability and reduces cognitive load for developers

The build will likely fail after this change since other parts of the code still reference config.ClarifyTask, but those will be addressed in subsequent tasks as outlined in the TODO.md.