# Task: Replace spinner update message calls with logger in main.go

## Goal
Replace all `spinnerInstance.UpdateMessage(msg)` calls with equivalent `logger.Info(msg)` calls in main.go, maintaining the same level of user feedback.

## Approaches Considered

### Approach 1: Direct One-to-One Replacement
Replace each `spinnerInstance.UpdateMessage(msg)` call with a direct `logger.Info(msg)` equivalent without modifying surrounding code.

**Pros:**
- Straightforward implementation
- Minimal code changes reduce risk of errors
- Consistent with the approach used for replacing Start/Stop messages
- Maintains existing message structure and content

**Cons:**
- Could leave the code with a mixture of old and new patterns until further tasks are completed
- May result in duplicate logging if UpdateMessage is already being paired with logger calls

### Approach 2: Replacement with Context Adjustment
Replace UpdateMessage calls with logger.Info calls and modify surrounding code to better fit the pure logging approach (e.g., removing redundant logging, adjusting log levels).

**Pros:**
- Results in cleaner code with better-integrated logging
- Addresses potential redundancy issues
- Better maintains the intent of providing progress updates

**Cons:**
- More complex changes that might go beyond the task's scope
- Potentially higher risk of introducing errors
- Less consistent with the incremental approach taken for Start/Stop replacement

### Approach 3: Dual-Level Logging Replacement
Replace UpdateMessage calls with both Info and Debug level logging to maintain the same behavior as the original UpdateMessage implementation.

**Pros:**
- Most closely replicates the original spinner behavior (which logs at both Info and Debug levels)
- Ensures no loss of information in log files
- Maintains the same output behavior at different log levels

**Cons:**
- Adds more logger calls, potentially cluttering the code
- May lead to duplicate or redundant logging
- Makes a simple change more complex than necessary

## Chosen Implementation Approach
**Approach 1: Direct One-to-One Replacement**

I will use a direct replacement approach, replacing each `spinnerInstance.UpdateMessage(msg)` call with `logger.Info(msg)`, maintaining the exact same message content.

Implementation steps:
1. Identify the 2 UpdateMessage calls in main.go (lines 155 and 469 according to our documentation)
2. Replace each call with the logger.Info equivalent
3. Ensure any fmt.Sprintf formatting or variables are preserved correctly

## Reasoning for Approach
I've chosen the direct one-to-one replacement approach for several reasons:

1. **Consistency with Previous Task**: This approach aligns with how we handled the Start/Stop replacements, maintaining consistency in our refactoring approach.

2. **Task Scope Adherence**: The task specifically calls for replacing UpdateMessage calls with logger.Info calls, and this approach directly fulfills that requirement without introducing additional changes.

3. **Minimized Risk**: With only 2 UpdateMessage calls to replace, the simplest approach with minimal changes reduces the risk of introducing errors.

4. **Preservation of Original Behavior**: Looking at the spinner implementation, we can see that when a spinner is disabled, UpdateMessage falls back to logger.Info, so this replacement maintains the same behavior.

While Approach 3 (dual-level logging) would more precisely replicate the spinner's behavior by preserving the Debug-level logging, the debug logs can be handled in a separate upcoming task ("Ensure debug-level logging is preserved") that specifically focuses on this aspect. For now, we'll focus on the primary user-facing message replacement.

This approach continues our incremental refactoring strategy, making one focused change at a time to ensure a smooth transition away from the spinner functionality.