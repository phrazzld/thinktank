# TODO

## Initial Setup
- [x] **Create skeleton files in cmd/architect/**
  - **Action:** Create the following files with proper package declarations and import statements: api.go, context.go, token.go, output.go, prompt.go.
  - **Depends On:** None.
  - **AC Ref:** Implementation Plan 1.

- [x] **Review existing cmd/architect/ implementation**
  - **Action:** Thoroughly analyze the existing cmd/architect/main.go and cmd/architect/cli.go to understand current functionality and integration points.
  - **Depends On:** None.
  - **AC Ref:** Refactoring Approach 1.

## Token Management
- [x] **Extract token counting logic to token.go**
  - **Action:** Move the tokenInfoResult struct and getTokenInfo function from main.go to token.go, maintaining the same function signatures.
  - **Depends On:** Create skeleton files in cmd/architect/.
  - **AC Ref:** Implementation Steps 2.

- [x] **Extract token limit checking to token.go**
  - **Action:** Move the checkTokenLimit function from main.go to token.go.
  - **Depends On:** Extract token counting logic to token.go.
  - **AC Ref:** Implementation Steps 2.

- [x] **Add token confirmation logic to token.go**
  - **Action:** Move the promptForConfirmation function from main.go to token.go.
  - **Depends On:** Extract token limit checking to token.go.
  - **AC Ref:** Implementation Steps 2.

- [x] **Create unit tests for token package functions**
  - **Action:** Add new tests for token.go functions to ensure they work correctly in isolation.
  - **Depends On:** All token management functions moved.
  - **AC Ref:** Implementation Steps 8, Testing Strategy 4.

## API Client
- [x] **Extract API client initialization to api.go**
  - **Action:** Move the initGeminiClient function from main.go to api.go.
  - **Depends On:** Create skeleton files in cmd/architect/.
  - **AC Ref:** Implementation Steps 3.

- [x] **Extract API response processing to api.go**
  - **Action:** Move the processApiResponse function from main.go to api.go.
  - **Depends On:** Extract API client initialization to api.go.
  - **AC Ref:** Implementation Steps 3.

- [x] **Implement robust error handling in api.go**
  - **Action:** Ensure consistent error handling across API operations, maintaining compatibility with existing error reporting.
  - **Depends On:** Extract API response processing to api.go.
  - **AC Ref:** Implementation Steps 3.

- [x] **Create unit tests for API functions**
  - **Action:** Add new tests for api.go functions to verify behavior in isolation.
  - **Depends On:** All API functions moved.
  - **AC Ref:** Implementation Steps 8, Testing Strategy 4.

## Context Gathering
- [x] **Extract context gathering to context.go**
  - **Action:** Move the gatherContext function from main.go to context.go.
  - **Depends On:** Create skeleton files in cmd/architect/.
  - **AC Ref:** Implementation Steps 4.

- [x] **Extract dry run info display to context.go**
  - **Action:** Move the displayDryRunInfo function from main.go to context.go.
  - **Depends On:** Extract context gathering to context.go.
  - **AC Ref:** Implementation Steps 4.

- [x] **Ensure file filtering logic remains intact**
  - **Action:** Verify that all file filtering functionality (include/exclude patterns) works exactly as before.
  - **Depends On:** Extract dry run info display to context.go.
  - **AC Ref:** Implementation Steps 4.

- [x] **Create unit tests for context package functions**
  - **Action:** Add new tests for context.go functions to ensure proper file handling.
  - **Depends On:** All context gathering functions moved.
  - **AC Ref:** Implementation Steps 8, Testing Strategy 4.

## Prompt Building
- [x] **Extract prompt building functions to prompt.go**
  - **Action:** Move buildPrompt, buildPromptWithConfig, and buildPromptWithManager functions from main.go to prompt.go.
  - **Depends On:** Create skeleton files in cmd/architect/.
  - **AC Ref:** Implementation Steps 5.

- [x] **Extract template loading to prompt.go**
  - **Action:** Move template related functions (listExampleTemplates, showExampleTemplate) from main.go to prompt.go.
  - **Depends On:** Extract prompt building functions to prompt.go.
  - **AC Ref:** Implementation Steps 5.

- [x] **Extract task file reading to prompt.go**
  - **Action:** Move the readTaskFromFile function from main.go to prompt.go.
  - **Depends On:** Extract template loading to prompt.go.
  - **AC Ref:** Implementation Steps 5.

- [x] **Create unit tests for prompt package functions**
  - **Action:** Add new tests for prompt.go functions to verify template handling.
  - **Depends On:** All prompt building functions moved.
  - **AC Ref:** Implementation Steps 8, Testing Strategy 4.

## Output Handling
- [x] **Extract file writing logic to output.go**
  - **Action:** Move the saveToFile function from main.go to output.go.
  - **Depends On:** Create skeleton files in cmd/architect/.
  - **AC Ref:** Implementation Steps 6.

- [x] **Extract plan generation and saving to output.go**
  - **Action:** Move the generateAndSavePlan, generateAndSavePlanWithConfig, and generateAndSavePlanWithPromptManager functions from main.go to output.go.
  - **Depends On:** Extract file writing logic to output.go.
  - **AC Ref:** Implementation Steps 6.

- [x] **Create unit tests for output package functions**
  - **Action:** Add new tests for output.go functions to ensure proper file writing.
  - **Depends On:** All output handling functions moved.
  - **AC Ref:** Implementation Steps 8, Testing Strategy 4.

## Main Entry Point
- [x] **Update cmd/architect/main.go**
  - **Action:** Expand the Main() function to use all the refactored components, gradually phasing out the original implementation.
  - **Depends On:** All extraction tasks completed (Token Management, API Client, Context Gathering, Prompt Building, Output Handling).
  - **AC Ref:** Implementation Steps 7.

- [x] **Add transitional comments to original main.go**
  - **Action:** Mark functions in the original main.go with "Transitional implementation" comments as they're moved to the new structure.
  - **Depends On:** Each function moved to a new file.
  - **AC Ref:** Implementation Steps 7.

- [x] **Enhance OriginalMain transitional implementation**
  - **Action:** Modify the OriginalMain function to gradually use the new components while maintaining backward compatibility.
  - **Depends On:** Update cmd/architect/main.go.
  - **AC Ref:** Implementation Steps 7.

- [x] **Create a new main.go entry point**
  - **Action:** Replace the root-level main.go with a simple entry point that calls cmd/architect/main.go's Main function.
  - **Depends On:** Enhance OriginalMain transitional implementation.
  - **AC Ref:** Implementation Steps 7.

## Testing


## Documentation
- [ ] **Update help text for commands**
  - **Action:** Verify the help text is consistent across all commands.
  - **Depends On:** Update README.md usage examples.
  - **AC Ref:** Implementation Steps 5.

- [ ] **Add refactoring notes to documentation**
  - **Action:** Document the architectural decisions made during refactoring, particularly regarding configuration management.
  - **Depends On:** All implementation completed.
  - **AC Ref:** Implementation Steps 5.

## Validation
- [ ] **Perform manual testing of key user journeys**
  - **Action:** Test the main user journeys to ensure the tool works correctly after refactoring.
  - **Depends On:** All other tasks completed.
  - **AC Ref:** Refactoring Approach 3.

## Refactoring Guidelines

Based on clarifications from architectural review:

1. **Pure Refactoring Approach:**
   - Focus solely on code reorganization with no functional changes
   - If improvements are identified, document them for future tasks
   - Maintain exact behavior compatibility throughout

2. **Incremental Testing:**
   - Use existing tests as a baseline
   - Add new unit tests for each extracted component
   - Run full test suite frequently to catch regressions early

3. **Deprecation Strategy:**
   - Add comments in code to mark transitional functions
   - Focus on phasing out the original main.go code, not on formal API deprecation
   - Eventually main.go should become minimal, just calling architect.Main()

4. **Import Path Considerations:**
   - Import path stability for external users is not a concern for main package refactoring
   - Focus on maintaining correct internal imports between components
   - Code in cmd/ and internal/ cannot be imported by external packages

5. **Code Duplication:**
   - Focus solely on refactoring main.go as defined in the plan
   - Document any duplication found for future improvement tasks
   - Avoid expanding scope to address duplication in other packages

6. **Configuration Management:**
   - Move away from passing the entire Configuration struct
   - Utilize config.ManagerInterface as the central configuration point
   - Pass only necessary configuration values to components
   - Components should declare specific configuration dependencies
