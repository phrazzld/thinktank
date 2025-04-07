# Standardize Virtual FS Setup

## Goal
Standardize the setup of virtual filesystem tests by replacing direct FS calls with the `createVirtualFs` helper consistently across all test files, creating a unified approach to filesystem testing.

## Implementation Approach
The approach I've chosen is to refactor all test files that set up their own filesystem structures to use the `createVirtualFs` helper from `virtualFsUtils.ts`. This will:

1. Identify all test files currently using custom file/directory creation methods
2. Replace direct filesystem setup (like separate `mkdirSync` and `writeFileSync` calls) with single `createVirtualFs` calls
3. Update directory path handling to use `path.join()` instead of string concatenation for cross-platform compatibility
4. Ensure proper reset of the virtual filesystem before each test with `resetVirtualFs()`
5. Adapt any test assertions to work with the standardized virtual filesystem approach

## Reasoning for this Approach
This approach provides several benefits:

1. **Consistency**: Using a single pattern for filesystem setup makes tests more readable and maintainable
2. **Simplicity**: The `createVirtualFs` helper simplifies setup by eliminating separate calls for each file and directory
3. **Cross-platform compatibility**: Using `path.join()` ensures tests work on all platforms
4. **Better isolation**: Proper reset before each test prevents cross-contamination
5. **Easier migration**: Having a consistent pattern makes further changes (like migrating to different virtual FS libraries) simpler

This aligns with the overall project goal of Phase 1, which is to complete the memfs migration for filesystem tests and create a consistent approach to filesystem testing across the codebase.