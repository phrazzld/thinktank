# Todo: Standard Library Helpers (R7)

*Priority: P3 (Very Low Impact, Very Low Risk)*

## Standard Library Helpers (R7)
- [ ] **T037 · Chore · P3**: ensure `go.mod` specifies Go 1.21+
    - **Context:** SHRINK_PLAN.md § 3.6 (R7), Implementation Step 3
    - **Action:**
        1. Check `go.mod` file.
        2. If needed, update the `go` directive to `1.21` or higher. Run `go mod tidy`.
    - **Done‑when:**
        1. `go.mod` specifies Go 1.21+.
    - **Depends‑on:** none
- [ ] **T038 · Refactor · P3**: remove custom min/max implementations
    - **Context:** SHRINK_PLAN.md § 3.6 (R7), Implementation Step 1
    - **Action:**
        1. Find and delete custom `min`/`max` function definitions.
    - **Done‑when:**
        1. No custom `min`/`max` functions remain.
    - **Depends‑on:** none
- [ ] **T039 · Refactor · P3**: replace min/max calls with stdlib equivalents
    - **Context:** SHRINK_PLAN.md § 3.6 (R7), Implementation Step 2
    - **Action:**
        1. Replace calls to custom `min`/`max` with standard library `min()` / `max()` (available since Go 1.21).
    - **Done‑when:**
        1. All calls use standard library functions.
        2. Code compiles and tests pass.
    - **Depends‑on:** [T037, T038]
