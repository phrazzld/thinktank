# Todo: Adapters Flattening (R6)

*Priority: P2 (Low-Medium Impact, Low-Medium Risk)*

## Adapters Flattening (R6)
- [ ] **T034 · Chore · P2**: verify concrete types satisfy orchestrator interfaces
    - **Context:** SHRINK_PLAN.md § 3.5 (R6), Implementation Step 1
    - **Action:**
        1. Check if `registryAPIService`, `contextGatherer`, `fileWriter` directly implement required interfaces (`interfaces.APIService`, etc.).
    - **Done‑when:**
        1. Interface satisfaction confirmed (yes/no).
    - **Depends‑on:** none
- [ ] **T035 · Refactor · P2**: remove `adapters.go` and update orchestrator injection (if applicable)
    - **Context:** SHRINK_PLAN.md § 3.5 (R6), Implementation Steps 2-3
    - **Action:**
        1. If T034 confirms satisfaction, `git rm internal/thinktank/adapters.go`.
        2. Update `NewOrchestrator` and call sites (`app.go`) to inject concrete types/interfaces directly.
    - **Done‑when:**
        1. `adapters.go` is removed (if applicable).
        2. Code compiles and relevant tests pass.
    - **Depends‑on:** [T034]
- [ ] **T036 · Test · P2**: run tests after adapter removal
    - **Context:** SHRINK_PLAN.md § 3.5 (R6), Implementation Step 4
    - **Action:**
        1. Execute `go test ./...`.
        2. Execute `./internal/e2e/run_e2e_tests.sh`.
    - **Done‑when:**
        1. All unit and E2E tests pass.
    - **Depends‑on:** [T035]
