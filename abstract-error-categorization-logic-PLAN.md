# Abstract error categorization logic

## Goal
Create utility functions or mappings to standardize error message categorization, replacing string-based conditionals with more maintainable approaches.

## Implementation Approach

### Approach 1: Pattern Matching Functions
Create specific pattern matching functions for each error category that check for keywords in error messages. These functions would return true/false based on pattern matches.

### Approach 2: Error Type Maps with Regexes
Create a mapping of error categories to arrays of regular expressions that match error messages belonging to that category. This would allow for a more maintainable and centralized approach.

### Approach 3: Error Type Registry with Categorization Functions
Create a registry system where error types register their categorization functions, eliminating the need for string-based conditionals entirely.

## Selected Approach
I've chosen Approach 2: Error Type Maps with Regexes because:

1. It provides a centralized, maintainable solution that removes string conditionals
2. Regular expressions offer more flexibility than simple string matching
3. It's more structured than Approach 1 but less complex than Approach 3
4. It aligns with the existing codebase patterns and doesn't require major refactoring
5. It will be easier to maintain, as new error patterns can be added to the maps

## Implementation Details
1. Create a new file for error categorization utilities
2. Define error category maps that associate categories with regex patterns
3. Implement utility functions that use these maps to categorize errors
4. Replace existing string-based conditionals with calls to these utility functions
5. Update tests to use the new categorization approach
