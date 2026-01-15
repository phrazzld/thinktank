# ADR-001: OpenRouter as Unified Provider

## Status
Accepted

## Context
thinktank needs to access multiple LLM providers (OpenAI, Google, Anthropic, etc.). Initial implementation had direct integrations with each provider, requiring:
- Multiple API keys (`OPENAI_API_KEY`, `GOOGLE_API_KEY`, etc.)
- Provider-specific client implementations
- Different error handling per provider
- Complex configuration

Users wanted simpler setup and the ability to compare models across providers easily.

## Decision
Route all LLM requests through OpenRouter's unified API.

- Single API key: `OPENROUTER_API_KEY`
- Single provider implementation: `internal/providers/openrouter/`
- Model IDs map to OpenRouter format: `openai/gpt-5`, `google/gemini-3-flash`
- OpenRouter handles provider-specific protocols

## Consequences

**Positive:**
- Single API key for all models
- Simplified codebase (one provider to maintain)
- Consistent error handling across providers
- Automatic failover capabilities from OpenRouter
- Price comparison and routing optimization

**Negative:**
- Dependency on OpenRouter availability
- Slightly higher latency (extra hop)
- OpenRouter usage fees on top of provider fees
- Can't use provider-specific features not exposed by OpenRouter

## Alternatives Considered

1. **Direct provider integrations**: More complex, multiple API keys, but lower latency
2. **LiteLLM proxy**: Self-hosted, but adds operational complexity
3. **Multiple provider support**: Keep both options - rejected for simplicity
