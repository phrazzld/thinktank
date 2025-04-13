# Bug Analysis and Fix Instructions

## Bug Report
- **Bug Description:** The CI workflow is failing at the test step with the following message: "Error: The action 'Run integration tests with parallel execution' has timed out after 5 minutes."
- **Expected Behavior:** Integration tests should complete within the 5-minute timeout limit.
- **Actual Behavior:** Integration tests are taking longer than 5 minutes to complete, causing the CI workflow to time out and fail.
- **Key Components/Files Mentioned:** 
  - CI workflow (in `.github/workflows/ci.yml`)
  - Integration tests (in `internal/integration/` directory)
- **Status:** Investigating

## Current Investigation
I've identified three main hypotheses:

1. **Intentional Delays in Rate Limit Tests**: The `rate_limit_test.go` tests explicitly use `time.Sleep()` for simulating rate limiting and are marked to be skipped in short mode, but the CI workflow doesn't use the `-short` flag.

2. **Excessive Sleep in Concurrency Tests**: The `multi_model_test.go` file's concurrency tests use substantial sleeps (50-150ms per model) to simulate API delays.

3. **Race Detector Overhead**: Running with `-race` adds significant overhead, especially with parallel execution and goroutines.

My planned first test is to add the `-short` flag to the CI workflow's integration test command.

You are an expert AI debugger. Your task is to analyze a reported bug, systematically investigate its root cause using the provided context, and formulate a precise fix.

## Instructions

Based on the Bug Report and Codebase Context:

1. **Analyze Current State:** Review existing hypotheses and test results (if any).

2. **Formulate/Refine Hypotheses:** If the root cause isn't clear, brainstorm or refine plausible hypotheses (potential causes, reasoning, validation ideas). Prioritize them.

3. **Design Next Test:** Propose the *next single, minimal test* to validate or refute the top hypothesis. Define:
   * `Hypothesis Tested:`
   * `Test Description:` (Specific action)
   * `Execution Method:` (e.g., Run test, add logging, use debugger)
   * `Expected Result (if true):`
   * `Expected Result (if false):`

4. **Identify Root Cause:** If the evidence strongly points to a root cause explaining the *entire* bug, state it clearly.

5. **Propose Fix:** If the root cause is identified, design and describe the specific code changes needed for the fix. Include reasoning and suggest inline comment format (`// BUGFIX: ..., CAUSE: ..., FIX: ...`).

6. **Verification:** Describe how to verify the fix (re-run reproduction steps, specific tests to run).

## Output

Provide the *next logical step* in the debugging process based on the instructions above. This could be:
* A list of new/refined hypotheses.
* The definition of the next test to run.
* The identified root cause.
* The proposed fix description and verification steps.

Format the output clearly, suitable for appending to the `Hypotheses`, `Test Log`, `Root Cause`, or `Fix Description` sections of `BUGFIXPLAN.md`.