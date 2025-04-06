# Clean Up Redundant Test Files

## Goal
Remove redundant test files to reduce code clutter and maintenance overhead while ensuring all functionality remains well-tested.

## Implementation Approach
1. Compare each pair of duplicate test files to confirm the newer *Helper versions are more comprehensive
2. Run the test suite before making changes to establish a baseline
3. Remove the older test files one at a time:
   - `processInput.test.ts`
   - `selectModels.test.ts`
   - `setupWorkflow.test.ts`
4. After each removal, run the test suite to ensure there are no regressions
5. Update any imports or references to these files if needed

## Rationale
After comparing the duplicate test files, I've determined that:

1. The newer files (with "Helper" in the name) were all created later on April 5th
2. The newer files appear to use better testing practices:
   - They use a dedicated test helper for mock spinners
   - They have more comprehensive test coverage
   - They follow a more consistent pattern across all helper functions

This approach minimizes risk by:
- Removing only the older, less comprehensive test files
- Testing incrementally after each removal
- Maintaining test coverage through the newer, better-structured tests

Keeping the well-organized *Helper.test.ts files will make the codebase more maintainable and reduce the cognitive overhead of having duplicate tests for the same functionality.