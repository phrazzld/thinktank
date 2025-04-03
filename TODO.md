# TODO

## Name Generator Refactoring
- [x] Define Word Lists
  - Description: Create and add arrays of at least 50 adjectives and 50 nouns to nameGenerator.ts
  - Dependencies: None
  - Priority: High

- [ ] Refactor generateFunName Function
  - Description: Modify generateFunName to use word lists for random name generation, remove API call logic
  - Dependencies: Word Lists
  - Priority: High

- [ ] Update Integration Point
  - Description: Modify call to generateFunName in runThinktank.ts to handle synchronous nature
  - Dependencies: Refactored generateFunName
  - Priority: High

- [ ] Update Unit Tests
  - Description: Rewrite tests in nameGenerator.test.ts to verify random generation logic
  - Dependencies: Refactored generateFunName
  - Priority: Medium

- [ ] Code Cleanup
  - Description: Remove unused imports and dead code related to API implementation
  - Dependencies: Refactored functions
  - Priority: Medium

## Optional Enhancements
- [ ] Implement Themed Word Lists
  - Description: Add themed sets of adjectives/nouns and logic to select a theme
  - Dependencies: Basic word list implementation
  - Priority: Low

## Testing & Verification
- [ ] Manual Verification
  - Description: Run application to ensure names are generated correctly and appear in logs
  - Dependencies: All implementation tasks
  - Priority: Medium

- [ ] Add Integration Tests
  - Description: Add tests that verify runThinktank workflow uses new name generator
  - Dependencies: All implementation tasks
  - Priority: Low