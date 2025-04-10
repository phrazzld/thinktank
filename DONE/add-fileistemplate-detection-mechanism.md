# Add fileIsTemplate detection mechanism

**Completed:** April, 2025

## Task Description
- **Action:** Implement logic to determine if a file should be processed as a template (with variable substitution) or used as raw content.
- **Depends On:** Update generateAndSavePlanWithPromptManager function
- **AC Ref:** Template processing decision (Technical Risk 3, Open Question 1)

## Implementation Details

### Approach
The implementation follows a hybrid approach:
1. First check if the file has a `.tmpl` extension - if so, it's always treated as a template
2. If not, fall back to content inspection using the existing `IsTemplate` function to check for template variables like `{{.Task}}` or `{{.Context}}`

This approach provides an explicit opt-in mechanism (the `.tmpl` extension) while maintaining backward compatibility with the content-based detection implemented in the previous task.

### Changes Made
1. Added a new `FileIsTemplate` function to the `prompt` package:

```go
// FileIsTemplate determines if a file should be processed as a template based on
// its file path and content. It checks the file extension first, and if it's not
// a .tmpl file, falls back to content inspection.
func FileIsTemplate(filePath string, content string) bool {
    // Check if the file has a .tmpl extension
    if filepath.Ext(filePath) == ".tmpl" {
        return true
    }
    
    // Fall back to content inspection
    return IsTemplate(content)
}
```

2. Added comprehensive unit tests in `prompt_test.go` to verify the functionality works as expected

### Key Benefits
- Clear, intuitive way for users to mark files as templates using standard `.tmpl` extension
- Backward compatibility with existing content-based detection
- Lightweight solution that doesn't require additional configuration
- Well-tested with comprehensive unit tests

### Future Work
- Consider integrating this detection mechanism with the file reading functions for better cohesion
- Possibly expand the template variable detection to support additional variables in the future
- Add documentation for users about how template detection works
- Implement integration tests for the full workflow with various prompt file formats