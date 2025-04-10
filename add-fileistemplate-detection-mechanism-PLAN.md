# Implementation Plan: Add fileIsTemplate detection mechanism

## Task
- **Action:** Implement logic to determine if a file should be processed as a template (with variable substitution) or used as raw content.
- **Depends On:** Update generateAndSavePlanWithPromptManager function (completed)
- **AC Ref:** Template processing decision (Technical Risk 3, Open Question 1)

## Goal
Provide a robust mechanism to detect whether an input file should be processed as a Go template (with variable substitution) or used as raw content. This detection should build upon the existing `IsTemplate` function but be more broadly applicable to file-based detection.

## Analysis of Current Implementation
The codebase already has a foundation for template detection:

1. The `IsTemplate` function in `prompt.go` uses regex to detect template variables:
   ```go
   func IsTemplate(content string) bool {
       re := regexp.MustCompile(`{{\s*\.(?:Task|Context)\s*}}`)
       return re.MatchString(content)
   }
   ```

2. In `main.go`, the `generateAndSavePlanWithPromptManager` function checks if a task file contains template variables:
   ```go
   if prompt.IsTemplate(config.TaskDescription) {
       // Process as a template
   } else {
       // Standard approach
   }
   ```

## Implementation Approaches Considered

### Approach 1: File Extension-Based Detection with Content Fallback
**Description:**
Implement a detection mechanism that primarily uses file extension (`.tmpl`) to identify template files, but falls back to content inspection for files without the extension.

**Implementation:**
1. Create a new `FileIsTemplate` function in the `prompt` package that:
   - First checks if the file has a `.tmpl` extension
   - If not, falls back to content analysis using the existing `IsTemplate` function

```go
// FileIsTemplate determines if a file should be processed as a template
func FileIsTemplate(filePath string, content string) bool {
    // Check extension first
    if filepath.Ext(filePath) == ".tmpl" {
        return true
    }
    
    // Fall back to content inspection
    return IsTemplate(content)
}
```

**Pros:**
- Simple implementation with clear rules
- Provides explicit opt-in via file extension
- Falls back to content detection for flexibility
- Maintains backward compatibility

**Cons:**
- Users must know about and use the `.tmpl` extension for explicit template files
- Content detection might have false positives in rare cases

### Approach 2: Configuration-Based Detection with Overrides
**Description:**
Extend the configuration system to allow users to specify template detection rules, including directories or patterns for template files, with content-based detection as a fallback.

**Implementation:**
1. Add template detection configuration to `config.go`
2. Implement a `FileIsTemplate` function that checks against configuration rules first, then falls back to content inspection

**Pros:**
- Highly customizable and flexible
- Allows for project-wide configuration 
- Can handle complex use cases

**Cons:**
- More complex implementation
- Requires users to understand and configure template rules
- Increases configuration complexity

### Approach 3: Pure Content-Based Detection (Current Approach Extended)
**Description:**
Extend the current content-based detection to a dedicated `FileIsTemplate` function that uses only content inspection, regardless of the file name or extension.

**Implementation:**
1. Create a simple wrapper function that applies the existing `IsTemplate` logic to files

```go
// FileIsTemplate determines if a file should be processed as a template 
// based solely on its content
func FileIsTemplate(filePath string, content string) bool {
    return IsTemplate(content)
}
```

**Pros:**
- Simplest implementation
- Consistent behavior regardless of file naming
- No additional configuration needed

**Cons:**
- No way to explicitly mark files as templates without using template variables
- May have false positives/negatives in edge cases
- Limited flexibility

## Chosen Approach

**Approach 1: File Extension-Based Detection with Content Fallback** is the recommended approach because:

1. It provides a clear, explicit mechanism for users to mark files as templates (using `.tmpl` extension)
2. It maintains the current content-based detection as a fallback for backward compatibility
3. It balances simplicity with flexibility
4. It doesn't require additional configuration
5. It's the most intuitive for users to understand and apply
6. It aligns with the existing pattern in the codebase where `.tmpl` extension is used for templates

## Implementation Details

### Files to Modify:
1. `/Users/phaedrus/Development/architect-context-files/internal/prompt/prompt.go` - Add the `FileIsTemplate` function
2. `/Users/phaedrus/Development/architect-context-files/internal/prompt/prompt_test.go` - Add tests for the new function

### High-Level Implementation Plan:
1. Add the `FileIsTemplate` function to `prompt.go`
2. Add unit tests for the function in `prompt_test.go`
3. Update relevant documentation

The implementation will be focused on testability, with clear separation of concerns and well-defined test cases to cover various scenarios.