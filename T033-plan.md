# T033 Plan: Fix import paths and package structure

## Overview
This task focuses on resolving issues with import paths, package structure, and vendor directory configuration in the codebase. The goal is to ensure consistent imports, eliminate circular dependencies, and handle vendor dependencies properly.

## Steps

1. **Analyze Current Import Structure**
   - Check for circular dependencies between packages
   - Identify inconsistent import paths
   - Examine vendor directory configuration and usage

2. **Fix Import Paths**
   - Ensure all imports follow the same pattern (github.com/phrazzld/thinktank/...)
   - Fix any relative imports causing issues
   - Ensure package names match directory names

3. **Address Circular Dependencies**
   - Identify any circular dependencies between packages
   - Refactor code to break circular dependencies by:
     - Moving shared types to interface packages
     - Using dependency injection
     - Restructuring package boundaries

4. **Configure Vendor Directory**
   - Review current vendor directory structure
   - Ensure all dependencies are properly vendored
   - Run with -mod=vendor to verify correctness

5. **Verify Changes**
   - Build the project with clean dependencies
   - Run the test suite to ensure everything works correctly
   - Verify that all circular dependencies are resolved

## Testing Plan
- Run `go build -mod=vendor ./...` to verify building with vendored dependencies
- Execute full test suite with `go test -mod=vendor ./...`
- Run specific tests from packages that had import issues before

## Success Criteria
- Clean build with no import-related errors
- No vendor-related errors in the build process
- All tests pass with the fixed import structure
EOL < /dev/null
