# DONE

## 2025-04-09

- [x] **Create example prompt file templates**
  - **Action:** Create example prompt files that users can reference when creating their own prompt files.
  - **Depends On:** None
  - **AC Ref:** User guidance (Open Question 2)

- [x] **Add deprecation warning for --task flag**
  - **Action:** When users provide --task without --task-file, log a clear deprecation warning suggesting migration to --task-file.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Backward compatibility (Technical Risk 1)
- [x] **Update parseFlags() function to modify task-file flag description**
  - **Action:** Modify the task-file flag description in parseFlags() to indicate it's required. Update the taskFlag description to indicate it's deprecated.
  - **Depends On:** None
  - **AC Ref:** Requirement to make --task-file required (Detailed Task Breakdown 1)

- [x] **Update flag.Usage() message in parseFlags()**
  - **Action:** Update the usage message to show --task-file as the primary means of providing input, as shown in the Implementation Specifications.
  - **Depends On:** None
  - **AC Ref:** Documentation update (Detailed Task Breakdown 4)

- [x] **Modify validateInputs() to require task file**
  - **Action:** Update the validateInputs() function to enforce the requirement for --task-file, following the implementation specification in PLAN.md.
  - **Depends On:** None
  - **AC Ref:** Validation logic update (Detailed Task Breakdown 2)

- [x] **Improve error messages for task file validation**
  - **Action:** Enhance error handling for file existence, readability, and content validation with clear error messages.
  - **Depends On:** Modify validateInputs() to require task file
  - **AC Ref:** Error handling (Technical Risk 2)

- [x] **Update README.md with new usage instructions**
  - **Action:** Add clear documentation in README.md about the requirement for --task-file and deprecation of --task.
  - **Depends On:** None
  - **AC Ref:** Documentation update (Detailed Task Breakdown 4)