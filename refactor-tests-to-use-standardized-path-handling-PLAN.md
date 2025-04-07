# Refactor tests to use standardized path handling

## Goal
Update all test files to use the new path normalization helper function, ensuring consistent behavior across the test suite.

## Chosen Implementation Approach
A hybrid approach that combines:

1. **Centralized Test Helper Update**: First updating core test utility functions in `virtualFsUtils.ts` and other helper files to use the `normalizePath` function internally, ensuring consistent path normalization at the utility level.

2. **Systematic Manual Refactoring**: Then methodically reviewing and updating each test file that uses path handling, focusing on:
   - Direct path string literals
   - Path manipulations (joining, concatenation)
   - Path comparisons in assertions
   - Path specifications for virtual filesystem setup

The refactoring will be done file-by-file to ensure careful consideration of each usage context, with tests run after each file update to verify functionality.

## Reasoning
This approach was chosen because:

- **Safety and Precision**: Manual review ensures `normalizePath` is applied correctly in each specific context (e.g., in fixtures vs. assertions), minimizing the risk of introducing subtle bugs.

- **Centralized Impact**: Updating core test helpers first provides the most impact with minimal changes, as suggested by multiple models.

- **Maintainability**: As noted in both gemini-2.5-pro and qwen-turbo responses, centralizing path normalization in shared functions improves long-term maintainability.

- **Alignment with Project Philosophy**: The project prioritizes careful, deliberate changes over potentially risky automation, and this approach respects that philosophy.

- **Manageability**: While automated approaches like codemods were considered, they introduce complexity and risks that aren't justified given the manageable scope of the codebase.

By first focusing on utility functions and then systematically addressing each test file, we'll achieve comprehensive path normalization with minimal risk and maximum reliability.