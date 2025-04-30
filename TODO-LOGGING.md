# Todo: Logging & Comment Pruning (R8)

*Priority: P3 (Medium Impact, Very Low Risk)*

## Logging & Comment Pruning (R8)
- [ ] **T040 · Refactor · P3**: audit and prune verbose logs in `internal/thinktank/orchestrator`
    - **Context:** SHRINK_PLAN.md § 3.7 (R8), Implementation
    - **Action:**
        1. Review log statements in the orchestrator package.
        2. Remove/lower level of trivial success or excessively verbose logs.
    - **Done‑when:**
        1. Logging is less verbose, focusing on significant events.
        2. Tests pass.
    - **Depends‑on:** none
- [ ] **T041 · Refactor · P3**: audit and prune verbose logs in `internal/thinktank/modelproc`
    - **Context:** SHRINK_PLAN.md § 3.7 (R8), Implementation
    - **Action:**
        1. Review log statements in the model processing package.
        2. Remove/lower level of trivial or verbose logs.
    - **Done‑when:**
        1. Logging is less verbose.
        2. Tests pass.
    - **Depends‑on:** none
- [ ] **T042 · Refactor · P3**: audit and prune verbose logs in `internal/providers/*`
    - **Context:** SHRINK_PLAN.md § 3.7 (R8), Implementation
    - **Action:**
        1. Review log statements in provider packages.
        2. Remove/lower level of trivial or verbose logs.
    - **Done‑when:**
        1. Logging is less verbose.
        2. Tests pass.
    - **Depends‑on:** none
- [ ] **T043 · Refactor · P3**: audit and remove comments explaining obvious code
    - **Context:** SHRINK_PLAN.md § 3.7 (R8), Implementation
    - **Action:**
        1. Scan codebase for comments that merely restate what the code does.
        2. Remove such comments, keeping "why" comments.
    - **Done‑when:**
        1. Redundant "what" comments are removed.
    - **Depends‑on:** none
