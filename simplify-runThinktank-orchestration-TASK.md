# Task: Simplify runThinktank orchestration

## Context
This task is part of the Phase 4 refactoring effort to improve testability by separating I/O operations from data transformation. The goal is to rewrite `runThinktank` in `src/workflow/runThinktank.ts` to compose pure functions and perform I/O operations at the boundaries.

## Requirements

1. **Primary Objective**: Rewrite the `runThinktank` function to:
   - Call pure functions for data transformation
   - Consolidate I/O operations at the boundaries of the workflow
   - Improve testability by reducing the need for complex mocking

2. **Current State**:
   - The current `runThinktank` function mixes orchestration flow, data processing, and I/O operations.
   - Some helper functions have already been refactored to be pure (e.g., `_processOutput`), but direct file I/O still happens in `runThinktank.ts`.
   - The workflow stages (setup, input processing, model selection, query execution, output processing) are well-defined, but the I/O boundaries are blurred.

3. **Technical Guidelines**:
   - Follow the dependency injection pattern already established
   - Use the existing `FileSystem` and other interfaces for all I/O operations
   - Update type definitions in `runThinktankTypes.ts` as needed
   - Maintain the same overall workflow structure and feature set
   - Ensure error handling patterns remain consistent with the rest of the codebase

4. **Acceptance Criteria**:
   - `runThinktank` should orchestrate the workflow by composing pure functions
   - All I/O operations should occur at the workflow boundaries
   - File system operations should use the injected `FileSystem` interface
   - Console output should use structured data returned from pure functions
   - Existing tests should pass (or be updated if the interface changes)
   - The function should maintain the same behavior from the user's perspective

## Constraints
- Must follow the project's established error handling patterns
- Must maintain backward compatibility with the existing function signature
- Must handle all the same error cases as the current implementation
- Tests should be updated to reflect the new pure approach

## Request
Please provide 2-3 approaches for refactoring the `runThinktank` function to achieve the goal of separating pure data transformation from I/O operations. Include:

1. A high-level description of each approach
2. The pros and cons of each approach
3. Implementation details and any interface changes needed
4. A recommended approach with justification, especially considering testability principles from TESTING_PHILOSOPHY.MD

Consider the role of dependency injection, handling of file I/O operations, and how to make the orchestration code more testable without complex mocking.