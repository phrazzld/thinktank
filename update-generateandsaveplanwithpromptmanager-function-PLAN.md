# Update generateAndSavePlanWithPromptManager function

## Goal
Modify the generateAndSavePlanWithPromptManager function to handle task file content appropriately, adding logic to determine if the file should be treated as a template or raw content.

## Approaches Considered

### Approach 1: Simple Template Detection
Add logic to detect template variables (like {{.Task}} or {{.Context}}) in the task file content. If detected, process the file as a template; otherwise, use it as raw content directly.

**Pros:**
- Simple and straightforward implementation
- Minimal changes to existing code
- Works with both formats transparently
- No additional configuration needed from users

**Cons:**
- May have false positives if the content contains {{ }} for other reasons
- Limited to detecting specific template variables

### Approach 2: File Extension-Based Detection
Use file extensions (.tmpl or .template) to determine if a file should be processed as a template.

**Pros:**
- Very explicit and clear for users
- No ambiguity in detection
- Standard practice in many template systems

**Cons:**
- Requires users to follow specific naming conventions
- Less flexible for users who want to use templates without renaming files

### Approach 3: Configuration Option
Add a new flag (--template-file) to explicitly specify that the file should be processed as a template, separate from --task-file.

**Pros:**
- Most explicit approach
- Gives users full control
- No ambiguity or detection issues

**Cons:**
- Adds complexity to the CLI
- Requires users to understand the difference between the flags
- Goes against the goal of simplifying the interface

## Chosen Approach
**Approach 1: Simple Template Detection**

I'll implement a simple template variable detection mechanism in the generateAndSavePlanWithPromptManager function. The logic will check if the task file content contains template variables like {{.Task}} or {{.Context}}. If detected, it will process the file as a template; otherwise, it will use the content directly.

This approach provides the best balance of user simplicity and functionality. It works transparently for both use cases without requiring users to follow special naming conventions or understand additional flags. The risk of false positives is low since the specific template variables like {{.Task}} and {{.Context}} are unlikely to appear in regular prompt content.

The implementation will:
1. Add a new `isTemplate` function to detect template variables in the content
2. Modify generateAndSavePlanWithPromptManager to check if the task file content is a template
3. If it's a template, process it using the existing template engine
4. If not, use the content directly as the prompt
