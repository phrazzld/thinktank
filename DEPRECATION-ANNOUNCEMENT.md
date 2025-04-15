# 📢 Important: Provider-Specific APIs Deprecation and Removal Timeline

## Summary

We are deprecating the Gemini-specific APIs in favor of provider-agnostic methods that work with all supported LLM providers. This change affects internal APIs used by developers extending or customizing Architect.

**Removal Timeline:** v0.8.0 (Q4 2024)

## Affected APIs

The following APIs are deprecated and will be removed:

| API | Replacement | Status |
|-----|-------------|--------|
| `InitClient` | `InitLLMClient` | Deprecated, scheduled for removal in v0.8.0 |
| `ProcessResponse` | `ProcessLLMResponse` | Deprecated, scheduled for removal in v0.8.0 |
| `llmToGeminiClientAdapter` | Use `llm.LLMClient` directly | Deprecated, scheduled for removal in v0.8.0 |
| Entire `internal/architect/compat` package | N/A | Created in v0.6.0, removal in v0.8.0 |

## Documentation

We've prepared comprehensive documentation to help you migrate:

- [Migration Guide](MIGRATION-GUIDE.md) - Step-by-step instructions for updating your code
- [Deprecation Plan](DEPRECATED-API-REMOVAL-PLAN.md) - Detailed plan and timeline for removal

## Why We're Making This Change

This change is part of our transition to a provider-agnostic architecture that supports multiple LLM providers (Gemini, OpenAI, etc.) with a consistent interface. Benefits include:

- **Simplified APIs:** Consistent interface regardless of the underlying provider
- **Provider Flexibility:** Switch between providers without code changes
- **Future-Proof:** Easier to add support for new providers
- **Reduced Complexity:** Removes adapter layer and duplicate code

## Timeline

1. **v0.6.0 (Q2 2024):**
   - Initial announcement
   - Documentation updates
   - Deprecated APIs moved to compatibility package

2. **v0.7.0 (Q3 2024):**
   - Warning logs added when deprecated APIs are used
   - Documentation emphasis on provider-agnostic methods

3. **v0.8.0 (Q4 2024):**
   - Complete removal of deprecated APIs
   - Removal of compatibility package

## Action Required

If you're using any of the deprecated APIs in your own code that extends or customizes Architect:

1. Read the [Migration Guide](MIGRATION-GUIDE.md)
2. Update your code to use the provider-agnostic replacements
3. Test your changes before v0.8.0 is released

## Support

If you have questions or need help migrating:

- Comment on this issue
- Open a new issue with the "api-migration" label
- Reference the migration guide for common patterns

## Related Issues

- #123: Initial provider-agnostic API implementation
- #234: Move deprecated methods to compatibility package
- #345: Add warning logs for deprecated methods (planned for v0.7.0)

---

This issue will remain open and pinned until after v0.8.0 is released to track the deprecation process and serve as a central place for questions.
