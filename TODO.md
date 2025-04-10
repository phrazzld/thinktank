# TODO

## Command-Line Interface Updates

## Input Validation

- [ ] **Improve error messages for task file validation**
  - **Action:** Enhance error handling for file existence, readability, and content validation with clear error messages.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Error handling (Technical Risk 2)

- [ ] **Add deprecation warning for --task flag**
  - **Action:** When users provide --task without --task-file, log a clear deprecation warning suggesting migration to --task-file.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Backward compatibility (Technical Risk 1)

## Prompt Template Processing
- [ ] **Update generateAndSavePlanWithPromptManager function**
  - **Action:** Modify this function to handle the task file content appropriately, adding logic to determine if the file should be treated as a template or raw content.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Template handling update (Detailed Task Breakdown 3)

- [ ] **Add fileIsTemplate detection mechanism**
  - **Action:** Implement logic to determine if a file should be processed as a template (with variable substitution) or used as raw content.
  - **Depends On:** Update generateAndSavePlanWithPromptManager function
  - **AC Ref:** Template processing decision (Technical Risk 3, Open Question 1)

## Documentation
- [ ] **Update README.md with new usage instructions**
  - **Action:** Add clear documentation in README.md about the requirement for --task-file and deprecation of --task.
  - **Depends On:** None
  - **AC Ref:** Documentation update (Detailed Task Breakdown 4)

- [ ] **Create example prompt file templates**
  - **Action:** Create example prompt files that users can reference when creating their own prompt files.
  - **Depends On:** None
  - **AC Ref:** User guidance (Open Question 2)

## Testing
- [ ] **Add unit tests for task file requirement**
  - **Action:** Create unit tests that verify the application properly requires the --task-file flag.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Unit testing (Detailed Task Breakdown 5, Testing Strategy)

- [ ] **Add unit tests for file existence and readability checks**
  - **Action:** Create unit tests that verify error handling when the file doesn't exist or is unreadable.
  - **Depends On:** Improve error messages for task file validation
  - **AC Ref:** Error handling tests (Testing Strategy)

- [ ] **Add integration tests for prompt file usage**
  - **Action:** Create integration tests that validate the complete workflow with various prompt file formats.
  - **Depends On:** Update generateAndSavePlanWithPromptManager function, Add fileIsTemplate detection mechanism
  - **AC Ref:** Integration testing (Testing Strategy)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Assumption: Template Processing Decision**
  - **Context:** The plan mentions "Decision needs to be made whether user-provided files should be processed as templates themselves (with variable substitution) or used as raw content" but doesn't definitively resolve this.
  - **Assumption:** We will implement a detection mechanism to determine if a file contains template variables (like {{.Task}} or {{.Context}}) and process it accordingly. Files without template variables will be used as raw content.

- [ ] **Assumption: Backward Compatibility Approach**
  - **Context:** The plan mentions "Consider providing a clear deprecation message rather than removing support immediately" for the --task flag.
  - **Assumption:** We will maintain support for the --task flag with a deprecation warning for a transition period, rather than removing it completely.

- [ ] **Clarification: Error Messaging for Prompt File Format**
  - **Context:** The plan asks "Should we enhance the error messaging to guide users toward proper prompt file formatting?"
  - **Clarification Needed:** What specific format guidance should be provided to users regarding prompt file structure and content?

- [ ] **Clarification: Example Templates Scope**
  - **Context:** The plan asks "Should we provide example prompt files to help users get started with creating their own prompts?"
  - **Clarification Needed:** How many example templates should be provided, and what variations should they cover?