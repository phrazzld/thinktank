# Delete clarify template files

## Goal
Check for and delete the `clarify.tmpl` and `refine.tmpl` files from the `internal/prompt/templates/` directory as part of the ongoing effort to completely remove the clarify feature from the codebase.

## Implementation Approach
1. Check the `internal/prompt/templates/` directory for the existence of `clarify.tmpl` and `refine.tmpl` files
2. If the files exist, delete them using appropriate Git commands to track the removal
3. Verify that removing these files doesn't break any existing functionality
4. Ensure all tests continue to pass after the files are removed
5. Document which files were found and removed, or confirm that the files did not exist

## Reasoning
This straightforward approach directly addresses the task by checking for the specified template files and removing them if they exist. Since the references to these templates have already been removed from the configuration in previous tasks, deleting the actual template files is a logical next step.

Alternative approaches considered:
1. **Rename the files with a ".deprecated" extension**: This would allow recovery if needed but would leave obsolete files in the repository, which doesn't align with the goal of completely removing the clarify feature.

2. **Move the files to a backup location outside the repository**: Similar to renaming, this would preserve the files but wouldn't fully remove them from the system, which contradicts the clear removal goal.

The chosen approach is the most direct and complete. It fully removes the template files while ensuring that the application continues to function correctly. This clean removal is consistent with the overall project goal of completely eliminating the clarify feature from the codebase. If the files don't exist, we can simply document this finding, which would indicate that the template removal aspect of the task has already been effectively completed.