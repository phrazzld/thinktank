# Update DefaultConfig() to remove clarify template references

## Task Goal
Remove references to Clarify and Refine templates from the DefaultConfig() function in internal/config/config.go as part of the ongoing effort to remove the clarify functionality from the codebase.

## Chosen Implementation Approach
After analyzing the codebase, I've chosen to implement a two-part approach:

1. **Verify No-Op in DefaultConfig()**: The DefaultConfig() function in config.go doesn't need modification because the Clarify and Refine fields have already been removed from the TemplateConfig struct in a previous task. The code would not compile if it attempted to assign values to non-existent fields.

2. **Update Template Loading in integration.go**: Instead, we need to modify the SetupPromptManagerWithConfig function in internal/prompt/integration.go to stop loading the clarify.tmpl and refine.tmpl templates. This change involves updating the slice of templates to remove these entries:

```go
// Before
for _, tmplName := range []string{"default.tmpl", "clarify.tmpl", "refine.tmpl"} {
    // ...
}

// After
for _, tmplName := range []string{"default.tmpl"} { // Only load default template
    // ...
}
```

## Key Reasoning

This approach is best because:

1. **Correctness**: It addresses the actual issue - stopping the application from loading the clarify and refine templates that we're removing.

2. **Completeness**: Although the task description focuses on DefaultConfig(), the real issue is in the template loading mechanism. This approach addresses both aspects - confirming no changes are needed in DefaultConfig() and stopping the actual template loading elsewhere.

3. **Testability**: This change affects a key behavior (which templates are loaded) rather than just configuration values. Tests can verify the application functions correctly without the removed templates, aligning with the project's testing philosophy of focusing on behavior over implementation details.

4. **Minimal Risk**: The change is simple and focused, reducing the chance of introducing errors. The change to integration.go is straightforward and directly addresses the functional issue.