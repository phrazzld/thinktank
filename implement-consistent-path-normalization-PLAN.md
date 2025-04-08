# Implement Consistent Path Normalization - Implementation Plan

## Task Title
Implement consistent path normalization

## Goal
Create and consistently use a standard path normalization utility across all tests to handle both Unix and Windows-style paths correctly.

## Chosen Approach
After analyzing the thinktank suggestions, I've chosen the **Dedicated Context-Specific Normalizers** approach, which involves creating separate utility functions with clear, focused responsibilities.

### Implementation Steps

1. **Centralize Path Utilities**
   - Create a dedicated utility module at `src/utils/pathUtils.ts` (moving from `src/__tests__/utils/pathUtils.ts`)
   - Keep `normalizePathForMemfs` in `src/__tests__/utils/virtualFsUtils.ts` but refine its implementation
   - Add unit tests for all path utility functions

2. **Create Specific Normalization Functions**
   - **`normalizePathGeneral(inputPath: string, keepLeadingSlash?: boolean): string`**
     - Main general-purpose path normalization utility
     - Standardizes separators to forward slashes
     - Handles redundant slashes, `.`, and `..`
     - Option to keep or remove leading slashes via parameter
     
   - **`normalizePathForComparison(path1: string, path2: string): [string, string]`**
     - Normalizes two paths for accurate comparison
     - Returns both normalized paths as a tuple
     - Ensures consistent format for both paths (both absolute or both relative)
     
   - **`normalizePathForGitignore(inputPath: string, basePath: string): string`**
     - Creates a properly formatted path for gitignore checks
     - Makes the path relative to the gitignore location
     - Uses forward slashes, handles special cases

3. **Refactor Existing Code**
   - Update all tests to use the appropriate normalization function
   - Remove any manual path normalization and separator handling
   - Update imports to point to the new centralized utilities
   - Ensure consistent cache clearing for gitignore tests

4. **Document Normalization Utilities**
   - Add clear JSDoc comments to explain each function's purpose
   - Include examples for common use cases
   - Explicitly note which function to use in which context

## Reasoning for This Approach

### Testability Considerations
1. **Simplified Testing Context**: 
   - Each function has a single responsibility, making it easier to test
   - Functions with specific use cases (memfs, gitignore) handle the quirks of those contexts
   - Clear purpose makes it obvious which function to use in each situation

2. **Clear Intent**: 
   - Function names indicate purpose (`normalizePathForMemfs`, `normalizePathForGitignore`)
   - Parameters clearly show what's needed (basePath for gitignore paths)
   - Tests become more readable when intent is clear

3. **Reduced Edge Cases**:
   - Specialization means each function handles just its specific edge cases
   - Less risk of platform-specific bugs since each function normalizes for a specific purpose

### Alignment with Testing Philosophy
1. **Simplicity & Clarity**: 
   - Each function does one thing well, matching the "Simplicity & Clarity" principle
   - Simple functions are easier to understand and maintain
   - Clear purpose reduces cognitive load

2. **Minimal Abstraction**:
   - Avoids one complex function with many options and internal branches
   - No unnecessary indirection or complex abstractions
   - Uses plain Node.js path module where appropriate

3. **Robustness**:
   - Dedicated functions ensure paths work correctly in each specific context
   - Reduces risk of cross-platform test failures
   - Makes tests more deterministic

### Implementation Details
The chosen approach will:
1. Make path handling consistent across all tests
2. Respect the specific needs of different contexts (memfs, gitignore)
3. Remove redundant normalization code
4. Make tests more robust against platform differences
5. Be easy to understand and use correctly

This solution strikes the optimal balance between simplicity and correctness while maintaining the clear, single-responsibility approach valued in the project's testing philosophy.