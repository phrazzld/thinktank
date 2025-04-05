# Remove ThinktankError from runThinktank.ts

## Goal
Remove the existing ThinktankError class definition from runThinktank.ts since we now have a centralized error system in src/core/errors.ts.

## Implementation Approach
Based on analysis of the code, the task is straightforward as ThinktankError is already imported from the core/errors.ts file. The implementation strategy will be:

1. Ensure the existing import is working correctly
2. Update ModelSelectionError in modelSelector.ts to extend from the core ThinktankError class
3. Verify all error handling in runThinktank.ts is properly using the imported ThinktankError

This approach ensures a clean transition to the centralized error system while maintaining all existing functionality.

## Reasoning
This task is part of the overall error handling refactoring effort (T3 in REFACTOR_PLAN.md). The existing code already imports ThinktankError from core/errors.ts, so there's no actual class definition to remove. However, we need to ensure ModelSelectionError in modelSelector.ts extends the core ThinktankError class to maintain consistent error handling and to make sure runThinktank.ts continues to work properly with the error system.

The selected approach:
- Focuses on proper integration of the error system
- Preserves backward compatibility
- Follows the refactoring plan's goal of centralizing error handling
- Makes minimal changes to ensure stability