# TODO

## Refactor runThinktank Workflow Improvements

- [x] **Clean Up Redundant Test Files**: Consolidate duplicate test files
  - **Action:** Remove redundant test files (`processInput.test.ts`, `selectModels.test.ts`, `setupWorkflow.test.ts`), keeping the more comprehensive `*Helper.test.ts` versions.
  - **Depends On:** None.
  - **AC Ref:** Low-risk issue identified in PLAN.md. Reduces code clutter and maintenance overhead.
  - **Completed:** Removed redundant test files, keeping the more comprehensive and newer *Helper.test.ts versions. Build and tests pass successfully.

- [x] **Simplify Error Handling Logic**: Make error categorization more maintainable
  - **Action:** Refactor the complex error handling in `_handleWorkflowError` by creating or enhancing utility functions for error categorization.
  - **Depends On:** None.
  - **AC Ref:** Medium-risk issue identified in PLAN.md. Could lead to maintenance challenges if not addressed.
  - **Completed:** Created a new `createContextualError` utility function in `categorization.ts` to centralize error handling, simplified the `_handleWorkflowError` function significantly. Successfully refactored to improve maintainability.

- [x] **Optimize Spinner Updates**: Reduce terminal flicker
  - **Action:** Implement debouncing or batching for spinner updates to reduce visual flicker and improve user experience.
  - **Depends On:** None.
  - **AC Ref:** Low-Medium risk issue identified in PLAN.md. Affects user experience.
  - **Completed:** Implemented a throttled spinner wrapper that limits update frequency to reduce flickering. Created a configurable factory to easily toggle between regular and throttled spinners.

- [x] **Refactor Duplicated Spinner Logic**: Centralize spinner update code
  - **Action:** Create a single function for updating spinner text based on model status to eliminate repetition in the code.
  - **Depends On:** "Optimize Spinner Updates"
  - **AC Ref:** Medium-risk issue identified in PLAN.md. Could lead to inconsistencies if not addressed.
  - **Completed:** Enhanced the ThrottledSpinner class with specialized methods to handle different types of status updates. Implemented a duck-typing approach that gracefully falls back to basic text updates for compatibility with regular Ora spinners.

- [x] **Add Missing JSDoc Comments**: Improve code documentation
  - **Action:** Add comprehensive JSDoc comments to helper functions in `runThinktankTypes.ts` and `runThinktankHelpers.ts`.
  - **Depends On:** None.
  - **AC Ref:** Low-risk issue identified in PLAN.md. Affects code understandability.
  - **Completed:** Added comprehensive JSDoc comments to both files, enhancing documentation of the workflow structure, function behavior, error contracts, and type definitions. Used consistent style with detailed descriptions for all components.

- [x] **Simplify Return Types**: Improve type clarity
  - **Action:** Refactor the `_selectModels` return type to eliminate unnecessary nesting by returning `ModelSelectionResult & { modeDescription: string }` directly.
  - **Depends On:** None.
  - **AC Ref:** Low-risk issue identified in PLAN.md. Minor code readability improvement.
  - **Completed:** Refactored SelectModelsResult to be an intersection type that combines ModelSelectionResult with the modeDescription property. Updated all code using this type to work with the flattened structure while maintaining backward compatibility.

- [x] **Improve Error Factory Accessibility**: Make error factories more discoverable
  - **Action:** Re-export error factory functions from `core/errors/index.ts` to improve discoverability and ease of use.
  - **Depends On:** None.
  - **AC Ref:** Low-risk issue identified in PLAN.md. Addresses code discoverability issue.
  - **Completed:** Re-exported all error factory functions including provider-specific factories from core/errors/index.ts. Updated provider modules to use the centralized imports. All tests pass except for unrelated existing failures.

## Additional Tasks

- [ ] **Ensure Test Consistency**: Verify test coverage after consolidation
  - **Action:** After removing redundant test files, ensure that all functionality is still properly tested by the remaining test files.
  - **Depends On:** "Clean Up Redundant Test Files"
  - **AC Ref:** Ensures test coverage remains complete after removing redundant tests.

- [x] **Update Documentation**: Reflect recent refactoring changes
  - **Action:** Update documentation to reflect the XDG-compliant configuration approach and remove any references to thinktank.config.json in the project root.
  - **Depends On:** None
  - **AC Ref:** Ensures documentation is consistent with actual implementation.
  - **Completed:** Updated README.md to correctly document XDG-compliant configuration paths. Updated error messages and examples to refer to the proper configuration file location. Fixed CLI examples and command reference to show the correct way to locate configuration files.
