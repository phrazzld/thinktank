# Deprecated API Removal Plan

This document outlines the process and timeline for completely removing the deprecated provider-specific methods (`InitClient`, `ProcessResponse`) and adapter (`llmToGeminiClientAdapter`) from the Architect tool. This constitutes Phase 3 of our deprecation plan for transitioning to provider-agnostic interfaces.

## Background

The Architect tool was initially built with tight integration to the Gemini API. As part of our effort to support multiple LLM providers, we've been migrating from Gemini-specific methods to provider-agnostic methods. This migration follows a three-phase plan:

1. **Phase 1 (Completed)**: Add deprecation notices to original methods and implement provider-agnostic alternatives.
2. **Phase 2 (Completed)**: Move deprecated methods to a dedicated compatibility package with clear documentation.
3. **Phase 3 (Current)**: Plan and execute the complete removal of deprecated methods.

## Timeline

| Milestone | Version | Date | Action |
|-----------|---------|------|--------|
| Notice Period Begins | v0.6.0 | Q2 2024 | Announce upcoming removal with a minimum 3-month notice period |
| Warning Logs | v0.7.0 | Q3 2024 | Add runtime warning logs when deprecated methods are used |
| Removal | v0.8.0 | Q4 2024 | Complete removal of deprecated methods and compatibility package |

## Migration Guide for Users

### Who Is Affected

Users who directly consume any of the following components will need to update their code:

- `InitClient` method from `internal/architect/api.go`
- `ProcessResponse` method from `internal/architect/api.go`
- `llmToGeminiClientAdapter` from `internal/architect/api.go`
- Any code importing the `internal/architect/compat` package

### Migration Steps

1. **Replace InitClient with InitLLMClient**:

   ```go
   // Old code (deprecated)
   client, err := apiService.InitClient(ctx, apiKey, modelName, apiEndpoint)

   // New code (provider-agnostic)
   client, err := apiService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
   ```

2. **Replace ProcessResponse with ProcessLLMResponse**:

   ```go
   // Old code (deprecated)
   content, err := apiService.ProcessResponse(result)

   // New code (provider-agnostic)
   content, err := apiService.ProcessLLMResponse(result)
   ```

3. **Update interface usage**:

   If your code depends on the Gemini-specific client interface, update it to use the provider-agnostic `llm.LLMClient` interface instead. This interface provides equivalent functionality but works with all supported LLM providers.

4. **Check error handling**:

   While we've maintained consistent error types and messages, review your error handling to ensure it works with the provider-agnostic approach.

## Communication Strategy

### Release Notes Template

The following template will be used for release notes when the deprecated methods are removed:

```
# Breaking Changes in v0.8.0

## Removal of Deprecated Gemini-Specific Methods

As announced in v0.6.0, we have removed the following deprecated methods:

- `InitClient` - Use `InitLLMClient` instead
- `ProcessResponse` - Use `ProcessLLMResponse` instead
- `llmToGeminiClientAdapter` - Use the provider-agnostic `llm.LLMClient` interface instead

The entire `internal/architect/compat` package has been removed.

These changes complete our transition to a provider-agnostic API that supports multiple LLM providers beyond Gemini.

### Migration Guide

Please refer to our [migration guide](MIGRATION-GUIDE.md) for detailed instructions on updating your code.
```

### Communication Channels

1. **GitHub Releases**: Detailed release notes will be published with each affected version.
2. **README Notice**: The README will be updated in v0.6.0 to highlight the upcoming breaking change.
3. **Deprecation Warnings**: Runtime warnings will be added in v0.7.0 when deprecated methods are called.
4. **Issue Tracker**: A pinned issue will track the deprecation process and serve as a place for users to ask questions.
5. **Documentation Update**: All documentation referring to the deprecated methods will be updated or removed.

## Implementation Steps

1. **v0.6.0 Preparation (Q2 2024)**:
   - Update README with deprecation notice
   - Create a pinned GitHub issue
   - Create a draft migration guide
   - Tag all relevant documentation with deprecation notices

2. **v0.7.0 Implementation (Q3 2024)**:
   - Add runtime warning logs in compatibility methods
   - Finalize migration guide
   - Update documentation to prioritize provider-agnostic methods
   - Prepare removal PRs in a feature branch

3. **v0.8.0 Removal (Q4 2024)**:
   - Remove the entire `internal/architect/compat` package
   - Update any remaining code references
   - Remove deprecation mentions from docs
   - Publish detailed release notes with migration links

## Testing Strategy

Before each release, we will:

1. Run the full test suite to ensure no regressions
2. Verify that proper warnings are displayed (v0.7.0)
3. Ensure all previously deprecated code is fully removed (v0.8.0)
4. Test examples and documentation to ensure they use the new APIs
5. Verify that error messages and handling are consistent

## Rollback Plan

If significant issues arise after removal:

1. For critical issues, consider releasing a hotfix (v0.8.1) that temporarily restores the compatibility package
2. For less critical issues, provide workarounds in documentation while users migrate
3. Extend the deprecation timeline if broad feedback indicates more migration time is needed

## Conclusion

This plan provides a clear timeline and process for removing deprecated Gemini-specific methods from the Architect tool. By following this plan, we aim to complete our transition to a fully provider-agnostic API while minimizing disruption for users.
