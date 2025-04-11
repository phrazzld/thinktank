# Consultation Request: Systematic Approach to Fix Build and Test Failures

## Goal
Fix all build errors and test failures in the architect project after the major refactoring that replaced the template-based system with an XML-structured approach.

## Problem/Blocker
We have several related issues that are causing build failures and test errors:

1. **Interface Mismatch**: The interfaces in cmd/architect/context.go and internal/architect/context.go are misaligned. The cmd version expects a string return type while the internal version returns a []fileutil.FileMeta slice.

2. **Log Level Configuration Issues**: The log level parsing in cli.go and cli_test.go isn't working properly, causing test failures.

3. **Incomplete Template Removal**: The config.ManagerInterface still requires template-related methods that have been removed from the actual implementation.

These issues violate several principles from our standards:
- **Simplicity (CORE_PRINCIPLES.md)**: We're currently trying to patch individual issues rather than addressing the root architectural causes.
- **Architecture (ARCHITECTURE_GUIDELINES.md)**: The interface mismatch between packages indicates a violation of proper API design and separation of concerns.
- **Testing (TESTING_STRATEGY.md)**: Our current approach makes testing unnecessarily complex due to broken interfaces.

## Context/History
We've been working through a series of tasks to refactor the architect tool:
1. Removed the template-based prompt system
2. Replaced it with a simpler XML-structured approach
3. Updated integration tests
4. Updated documentation

We're now in the "cleanup" phase where we need to ensure everything builds and tests pass, but we're finding several interface mismatches and configuration issues.

## Key Files/Code Sections
- `/Users/phaedrus/Development/architect/cmd/architect/context.go` - Contains a ContextGatherer interface that expects a string return type
- `/Users/phaedrus/Development/architect/internal/architect/context.go` - Contains a ContextGatherer interface that returns a []fileutil.FileMeta slice
- `/Users/phaedrus/Development/architect/cmd/architect/cli.go` - Has issues with log level parsing
- `/Users/phaedrus/Development/architect/cmd/architect/cli_test.go` - Test fails due to log level issues
- `/Users/phaedrus/Development/architect/internal/config/interfaces.go` - Defines the ManagerInterface with template methods
- `/Users/phaedrus/Development/architect/internal/config/loader.go` - Implementation that should support the interface

## Error Messages (if any)
```
cmd/architect/context.go:51:9: cannot use &contextGatherer{…} (value of type *contextGatherer) as ContextGatherer value in return statement: *contextGatherer does not implement ContextGatherer (wrong type for method GatherContext)
  have GatherContext(context.Context, gemini.Client, GatherConfig) (string, *ContextStats, error)
  want GatherContext(context.Context, gemini.Client, GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)

--- FAIL: TestAdvancedConfiguration (0.00s)
    --- FAIL: TestAdvancedConfiguration/Custom_format_and_model (0.00s)
        cli_test.go:241: Expected LogLevel to be "debug", got "info"
```

## Desired Outcome
A systematic approach to fix all the build issues and test failures by:
1. Properly aligning all interfaces between the cmd and internal packages
2. Ensuring log level configuration works consistently across the codebase
3. Cleaning up any remaining references to the removed template system
4. Implementing a solution that doesn't just patch individual symptoms but addresses the root architectural causes