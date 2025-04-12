```markdown
# TODO

## [Issue 1: Redundant Validation Logic]
- [x] **Task Title:** Remove Redundant `validateInputs` Function from `internal/architect/app.go`
  - **Action:** Delete the `validateInputs` function located at `internal/architect/app.go:950-969`. Verify that all necessary input validation is comprehensively handled by `cmd/architect/cli.go:ValidateInputs` before `architect.Execute` is called.
  - **Depends On:** None
  - **AC Ref:** Issue 1, `CORE_PRINCIPLES.md` (Simplicity), `ARCHITECTURE_GUIDELINES.md` (Separation of Concerns)

## [Issue 2: Output Saving Backward Compatibility Complexity]
- [ ] **Task Title:** Decide on Output File Backward Compatibility Strategy
  - **Action:** Analyze test dependencies (`integration_test.go`, `xml_integration_test.go`) and any other potential consumers of the legacy `outputDir/output.md` file. Choose one strategy: A) Update tests to use new model-specific paths (`outputDir/model-name.md`) and remove the legacy write, OR B) Keep the legacy write for backward compatibility. Document the chosen strategy and rationale (e.g., in commit message, issue tracker).
  - **Depends On:** None
  - **AC Ref:** Issue 2, `CORE_PRINCIPLES.md` (Simplicity)
- [ ] **Task Title:** [If Strategy A] Update Integration Tests for Model-Specific Outputs
  - **Action:** Modify `internal/integration/integration_test.go` and `internal/integration/xml_integration_test.go` to assert against the new model-specific output files (e.g., `filepath.Join(outputDir, "test-model.md")`) instead of relying on `output.md`.
  - **Depends On:** Decide on Output File Backward Compatibility Strategy
  - **AC Ref:** Issue 2
- [ ] **Task Title:** [If Strategy A] Simplify `savePlanToFile` Function
  - **Action:** Remove the conditional logic and file write operation related to the legacy `output.md` path within the `savePlanToFile` function in `internal/architect/app.go:956-1016`. Ensure only the model-specific file is written.
  - **Depends On:** [If Strategy A] Update Integration Tests for Model-Specific Outputs
  - **AC Ref:** Issue 2, `CORE_PRINCIPLES.md` (Simplicity)
- [ ] **Task Title:** [If Strategy B] Add Explanatory Comments to `savePlanToFile`
  - **Action:** Add clear comments within the `savePlanToFile` function (`internal/architect/app.go:956-1016`) explaining *why* the code writes to both the model-specific path and the legacy `output.md` path, explicitly mentioning the backward compatibility requirement.
  - **Depends On:** Decide on Output File Backward Compatibility Strategy
  - **AC Ref:** Issue 2

## [Issue 3: Inclusion of Development Artifact (`TODO.md`)]
- [ ] **Task Title:** Remove Development `TODO.md` from Repository
  - **Action:** Delete the `TODO.md` file from the project repository. Before deletion, ensure any relevant pending tasks or future considerations mentioned within it are captured in the project's official issue tracker or backlog.
  - **Depends On:** None
  - **AC Ref:** Issue 3, `DOCUMENTATION_APPROACH.md` (Principles 1 & 5)

## [Issue 4: HTML Escape Sequences in Go Test Files]
- [ ] **Task Title:** Replace HTML Escape Sequences in Go Test Files
  - **Action:** Perform a search-and-replace in the specified Go test files (`internal/runutil/runutil_test.go`, `internal/integration/xml_integration_test.go`). Replace all occurrences of `&lt;` with the literal character `<` and `&gt;` with the literal character `>`.
  - **Depends On:** None
  - **AC Ref:** Issue 4, `CODING_STANDARDS.md` (Readability/Clarity)

## [Issue 5: Minor Documentation Inconsistencies]
- [ ] **Task Title:** Standardize "architect" Casing in README.md
  - **Action:** Edit `README.md`. Change the main title from `# Code Review: ...` (or similar) to `# architect`. Ensure all other references to the tool name within the body text, examples, and configuration sections consistently use the lowercase "architect".
  - **Depends On:** None
  - **AC Ref:** Issue 5, `DOCUMENTATION_APPROACH.md` (Clarity and Consistency)
- [ ] **Task Title:** Add Missing Newline Before License Link in README.md
  - **Action:** Edit `README.md`. Locate the license link at the very end (e.g., `[MIT License](LICENSE)`) and ensure there is a blank line immediately preceding it.
  - **Depends On:** None
  - **AC Ref:** Issue 5, `DOCUMENTATION_APPROACH.md` (Clarity and Consistency)

## [Issue 6: Filename Sanitization Character]
- [ ] **Task Title:** Correct Characters in `sanitizeFilename` Replacer
  - **Action:** Modify the `strings.NewReplacer` call within the `sanitizeFilename` function (`internal/architect/app.go:540-542`). Change the placeholder `&lt;` to the literal character `<` and `&gt;` to the literal character `>`.
  - **Depends On:** None
  - **AC Ref:** Issue 6 (Potential Bug)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** Assumed the decision for Issue 2 (Backward Compatibility) will be made and documented before dependent tasks are started.
  - **Context:** Issue 2 tasks depend on the initial decision (Strategy A vs. Strategy B).
- [ ] **Issue/Assumption:** Assumed that removing `TODO.md` (Issue 3) implies transferring any necessary information to a formal issue tracker if one exists.
  - **Context:** Issue 3 suggests removing the file, but valuable context might be lost if not migrated.
```