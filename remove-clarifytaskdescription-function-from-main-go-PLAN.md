# Remove clarifyTaskDescription() function from main.go

## Task Goal
Remove the entire `clarifyTaskDescription()` function from the main.go file as part of the ongoing effort to remove the clarify functionality from the codebase.

## Implementation Approach
After analyzing the code, I've determined that the most effective approach is to:

1. **Identify Function Dependencies:** The function is only referenced in one place - inside the `clarifyTaskDescriptionWithConfig()` function as a fallback option.

2. **Direct Function Removal:** Remove the entire `clarifyTaskDescription()` function (lines 107-112 in main.go).

3. **Update References:** Modify the `clarifyTaskDescriptionWithConfig()` function to handle the error case differently, since it can no longer call `clarifyTaskDescription()` as a fallback. Instead, it will directly create a prompt manager and call `clarifyTaskDescriptionWithPromptManager()`.

## Key Reasoning

1. **Minimal Impact:** This targeted removal approach minimizes the risk of unintended side effects. Only one reference exists, which we can easily update.

2. **Clean Encapsulation:** By directly creating the prompt manager in the error handler of `clarifyTaskDescriptionWithConfig()`, we maintain the function's intent while removing the dependency on the function we're deleting.

3. **Incremental Approach:** This change is part of a larger refactoring effort, and this small step keeps the codebase functional while progressing toward the goal of removing all clarify functionality.

4. **No Build Errors:** The approach ensures the code will still compile and build successfully, as we're updating all references to the removed function.

5. **Low Testability Impact:** Since this is a pure removal that preserves existing behavior (until the entire feature is removed), we don't need additional tests beyond verifying the build still succeeds.