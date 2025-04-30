# Todo: API Key Resolution Centralization (R5)

*Priority: P2 (Low Impact, Low-Medium Risk)*

## API Key Resolution Centralization (R5)
- [ ] **T030 · Refactor · P2**: centralize API key resolution in `InitLLMClient`
    - **Context:** SHRINK_PLAN.md § 3.4 (R5), Implementation Step 1
    - **Action:**
        1. Ensure `registryAPIService.InitLLMClient` implements full key logic (CLI param, env vars via `ModelsConfig.APIKeySources`, global keys).
    - **Done‑when:**
        1. `InitLLMClient` contains the single source of truth for API key resolution.
    - **Depends‑on:** none
- [ ] **T031 · Refactor · P2**: remove API key lookups from provider `CreateClient` methods
    - **Context:** SHRINK_PLAN.md § 3.4 (R5), Implementation Step 2
    - **Action:**
        1. Remove redundant key lookups (env, flags) from individual provider `CreateClient` implementations.
    - **Done‑when:**
        1. Provider `CreateClient` methods rely solely on the key passed via arguments.
        2. Provider tests pass.
    - **Depends‑on:** [T030]
- [ ] **T032 · Refactor · P2**: remove redundant API key checks from `cmd/thinktank/cli.go`
    - **Context:** SHRINK_PLAN.md § 3.4 (R5), Implementation Step 3
    - **Action:**
        1. Remove logic from CLI parsing that duplicates key resolution performed later by `InitLLMClient`.
    - **Done‑when:**
        1. CLI code relies on `InitLLMClient` for key validation/resolution during execution.
        2. CLI tests pass.
    - **Depends‑on:** [T030]
- [ ] **T033 · Test · P2**: update tests for API key precedence in `InitLLMClient`
    - **Context:** SHRINK_PLAN.md § 3.4 (R5), Implementation Step 4
    - **Action:**
        1. Write/update unit tests for `InitLLMClient` verifying correct precedence (CLI > Env > Global).
    - **Done‑when:**
        1. Tests pass and cover all key precedence scenarios.
    - **Depends‑on:** [T030, T031, T032]
