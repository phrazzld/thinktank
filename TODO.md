# TODO

## Core Error System
- [x] Create src/core/errors.ts file
  - Description: Create new file with error categories and base ThinktankError class
  - Dependencies: None
  - Priority: High

- [x] Implement specialized error subclasses
  - Description: Add ConfigError, ApiError, FileSystemError, etc. extending ThinktankError
  - Dependencies: Base ThinktankError class
  - Priority: High

- [x] Implement error factory functions
  - Description: Move existing error creation logic into specialized factory functions
  - Dependencies: Error subclasses
  - Priority: High

- [x] Add error formatting capabilities
  - Description: Implement format() method on ThinktankError base class
  - Dependencies: Base ThinktankError class
  - Priority: High

## Migration and Backwards Compatibility
- [x] Update consoleUtils.ts to use new error system
  - Description: Update existing utility functions to use new error classes, add deprecation notices
  - Dependencies: Core error system implementation
  - Priority: High

- [x] Remove ThinktankError from runThinktank.ts
  - Description: Remove the existing ThinktankError class definition
  - Dependencies: Core error system implementation
  - Priority: High

- [x] Update error handling in runThinktank.ts
  - Description: Update try/catch blocks to use new error classes
  - Dependencies: Core error system implementation
  - Priority: Medium

- [x] Update ModelSelectionError handling
  - Description: Convert ModelSelectionError to use the new error hierarchy
  - Dependencies: Core error system implementation
  - Priority: Medium

## Provider-Specific Updates
- [x] Update anthropic.ts error handling
  - Description: Use ApiError in catch blocks with providerId
  - Dependencies: Core error system implementation
  - Priority: Medium

- [ ] Update openai.ts error handling
  - Description: Use ApiError in catch blocks with providerId
  - Dependencies: Core error system implementation
  - Priority: Medium

- [ ] Update google.ts error handling
  - Description: Use ApiError in catch blocks with providerId
  - Dependencies: Core error system implementation
  - Priority: Medium

- [ ] Update openrouter.ts error handling
  - Description: Use ApiError in catch blocks with providerId
  - Dependencies: Core error system implementation
  - Priority: Medium

## CLI Error Handling
- [ ] Update CLI error handler in index.ts
  - Description: Update handleError function to use format() method from ThinktankError
  - Dependencies: Core error system implementation
  - Priority: Medium

- [ ] Update CLI command error handling
  - Description: Update try/catch blocks in command files to use new error classes
  - Dependencies: Core error system implementation
  - Priority: Low

## Testing
- [x] Create errors.test.ts
  - Description: Create comprehensive test suite for new error classes and factory functions
  - Dependencies: Core error system implementation
  - Priority: High

- [ ] Update existing error handling tests
  - Description: Update runThinktank-error-handling.test.ts and cli-error-handling.test.ts
  - Dependencies: Updated modules and error system
  - Priority: Medium

- [ ] Test cross-module error propagation
  - Description: Test how errors propagate from providers through workflow to CLI
  - Dependencies: All updated modules
  - Priority: Medium

## Documentation and Cleanup
- [ ] Add JSDoc comments
  - Description: Ensure comprehensive documentation for all error classes and functions
  - Dependencies: Core error system implementation
  - Priority: Low

- [ ] Update REFACTOR_PLAN.md
  - Description: Mark T3 task as completed with implementation details
  - Dependencies: All implementation complete
  - Priority: Low

## Assumptions and Clarifications
- The plan assumes we should maintain backward compatibility during the transition
- The PLAN.md suggests moving formatting logic into the error classes themselves, which is a good approach but different from the current implementation
- We're focusing on making this change with minimal impact to the rest of the codebase
- We're deprecating rather than immediately removing old utility functions to facilitate gradual migration