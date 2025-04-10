# Implementation Plan: Modify validateInputs() to require task file

## Task
Update the validateInputs() function to enforce the requirement for --task-file, following the implementation specification in PLAN.md.

## Chosen Approach
I'll implement **Approach 3: Phased Deprecation with Warning** from the analysis.

This approach modifies the validateInputs() function to prefer --task-file but allows --task temporarily while issuing a deprecation warning. This provides a smooth transition period for users who are still using the --task flag while clearly indicating the direction toward requiring --task-file in the future.

### Implementation Steps:
1. Restructure the validateInputs() function to have more explicit flow control with a taskLoaded flag
2. Handle the case where --task-file is provided first, treating it as the preferred path
3. If --task-file isn't provided but --task is (and not in dry-run mode), issue a deprecation warning
4. If neither flag is provided (and not in dry-run mode), show an error indicating --task-file is required
5. Maintain existing checks for Paths and ApiKey

### Code Changes:

```go
func validateInputs(config *Configuration, logger logutil.LoggerInterface) {
    // --- Start Modifications ---
    taskLoaded := false
    if config.TaskFile != "" {
        // Task file provided - this is the preferred path
        taskContent, err := readTaskFromFile(config.TaskFile, logger)
        if err != nil {
            logger.Error("Failed to load task file: %v", err)
            flag.Usage()
            os.Exit(1)
        }
        config.TaskDescription = taskContent
        taskLoaded = true
        logger.Debug("Loaded task description from file: %s", config.TaskFile)

        // Check if --task was also unnecessarily provided
        if flag.Lookup("task").Value.String() != "" {
             logger.Warn("Both --task and --task-file flags were provided. Using task from --task-file. The --task flag is deprecated.")
        }

    } else if flag.Lookup("task").Value.String() != "" && !config.DryRun {
         // Task file NOT provided, but deprecated --task IS provided (and not dry run)
         logger.Warn("The --task flag is deprecated and will be removed in a future version. Please use --task-file instead.")
         // config.TaskDescription is already set from parseFlags
         taskLoaded = true
    }

    // Check if a task is loaded (unless in dry-run mode)
    if !taskLoaded && !config.DryRun {
         logger.Error("The required --task-file flag is missing.")
         flag.Usage()
         os.Exit(1)
    }
    // --- End Modifications ---

    // Keep existing checks for other required inputs
    // ...
}
```

## Reasoning for Choice

While Approach 2 (Refined Implementation) offers a cleaner solution by strictly requiring --task-file, I've selected Approach 3 for the following reasons:

1. **Backward Compatibility**: This approach provides a transition period that prevents immediately breaking existing scripts or user workflows that rely on the --task flag, while clearly signaling the deprecation.

2. **User Experience**: By providing warning messages rather than immediate errors, users are guided toward the new approach while still being able to complete their immediate tasks.

3. **Gradual Migration**: The implementation allows for a phased migration strategy, where we can first deprecate with warnings before making a hard requirement in a future version.

4. **Assumption Alignment**: This approach aligns with the assumption noted in the TODO.md: "We will maintain support for the --task flag with a deprecation warning for a transition period, rather than removing it completely."

5. **Balance**: While this approach has more complex logic, it strikes a balance between enforcing the new requirement and providing a smooth user experience during transition.

The slight increase in code complexity is an acceptable trade-off for the improved user experience and reduced risk of disrupting existing workflows. In a future release, once users have had time to update their scripts and workflows, we can simplify the code by removing the deprecated path entirely.

## Testability Considerations

Testing this implementation will require:
- Mocking os.Exit to prevent test termination
- Testing multiple paths (--task-file provided, --task provided, neither provided)
- Verifying log messages for the deprecation warnings
- Checking that config.TaskDescription is properly set in each case

The added complexity means more test cases, but they can be structured clearly to test each distinct path through the function.