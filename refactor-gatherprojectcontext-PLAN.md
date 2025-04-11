# Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)

## Goal
Change the signature and implementation of the `GatherProjectContext` function to return a slice of `FileMeta` structs instead of a formatted string. This change will separate the content gathering logic from the formatting logic, which will be moved to a dedicated prompt stitching function.

## Implementation Approach
1. Modify the signature of `GatherProjectContext` to return a slice of `FileMeta` structs:
   ```go
   func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error)
   ```

2. Modify the `processFile` function to add to a `[]FileMeta` slice instead of writing to a string builder:
   ```go
   func processFile(path string, files *[]FileMeta, config *Config)
   ```

3. Update the implementation of `GatherProjectContext` to:
   - Initialize an empty `[]FileMeta` slice instead of a `strings.Builder`
   - Remove the `<context>` wrapper tags (as formatting will be handled elsewhere)
   - Pass the slice to `processFile` for population
   - Return the populated slice along with the file count and error

4. Update the implementation of `processFile` to:
   - Remove usage of `config.Format` for formatting
   - Create a new `FileMeta` struct with the file path and content
   - Append the struct to the slice

## Alternative Approaches Considered

### Alternative 1: Keep Format Field for Future Flexibility
One alternative approach would be to keep the `config.Format` field and use it during the prompt stitching phase. This would maintain flexibility for different formatting styles. However, this would complicate the refactoring process and the format seems to be standardized to `<{path}>...</{path}>` in the new design.

### Alternative 2: Add More Fields to FileMeta
Another approach would be to add more fields to the `FileMeta` struct, such as `FileType`, `Size`, or metadata generated during processing. This could provide more context for the prompt stitching phase. However, the current requirements only specify path and content, and we can always extend the struct later if needed.

## Key Reasoning
1. **Separation of Concerns**: This refactoring cleanly separates the content gathering from the content formatting, following the single responsibility principle.

2. **Simplified Interface**: The new implementation provides structured data that can be processed by the prompt stitching logic in a more flexible way, rather than pre-formatting the content as a string.

3. **Maintainability**: By removing the format-specific logic from `fileutil`, we make that package more focused on its core responsibility of file handling, which will make future changes easier.

4. **Performance**: The change may also offer a slight performance improvement as we no longer need to format and concatenate strings during the gathering phase, only to potentially reformat them later.

5. **Adherence to Plan**: This approach follows the project's refactoring plan by moving from template-based to a more direct instructions + structured context approach.