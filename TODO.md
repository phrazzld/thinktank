# TODO

## Pre-commit Framework Migration

- [x] **Create `scripts` Directory:**
  - **Action:** Create a `scripts` directory at the root of the project if it doesn't already exist.
  - **Depends On:** None
  - **AC Ref:** PLAN.md Step 3

- [x] **Create `scripts/check-large-files.sh` Script:**
  - **Action:** Create the bash script at `scripts/check-large-files.sh` with the exact content provided in PLAN.md Step 3. Ensure the script has execute permissions (`chmod +x scripts/check-large-files.sh`).
  - **Depends On:** Create `scripts` Directory
  - **AC Ref:** PLAN.md Step 3

- [x] **Create `.pre-commit-config.yaml`:**
  - **Action:** Create the `.pre-commit-config.yaml` file at the project root. Populate it with the YAML configuration specified in PLAN.md Step 2, including hooks for `pre-commit-hooks`, `golangci-lint`, `dnephin/pre-commit-golang`, and the `local` hook referencing the script from the previous task.
  - **Depends On:** Create `scripts/check-large-files.sh` Script
  - **AC Ref:** PLAN.md Step 2

- [x] **Run `pre-commit autoupdate`:**
  - **Action:** Run `pre-commit autoupdate` in the terminal to ensure hook revisions are up-to-date. If this command modifies the `rev` fields in `.pre-commit-config.yaml`, review the changes to ensure they are acceptable.
  - **Depends On:** Create `.pre-commit-config.yaml`
  - **AC Ref:** PLAN.md Step 4

- [x] **Install pre-commit hook:**
  - **Action:** Run `pre-commit install` from the project root to install the pre-commit hook in the git repository.
  - **Depends On:** Run `pre-commit autoupdate`
  - **AC Ref:** PLAN.md Step 4

- [x] **Test pre-commit hooks:**
  - **Action:** Run `pre-commit run --all-files` to test all hooks against all files in the repository. Verify hooks work correctly.
  - **Depends On:** Install pre-commit hook
  - **AC Ref:** PLAN.md Step 4

- [x] **Verify hook works on real changes:**
  - **Action:** Make a test change (e.g., add trailing whitespace to a Go file) and attempt to commit to verify the hook works correctly. Then remove the whitespace and commit successfully.
  - **Depends On:** Test pre-commit hooks
  - **AC Ref:** PLAN.md Step 4

- [x] **Update `hooks/README.md`:**
  - **Action:** Replace the entire content of the existing `hooks/README.md` file with the new Markdown content provided in PLAN.md Step 5. This new content should accurately describe the `pre-commit` framework setup, installation, and usage.
  - **Depends On:** Verify hook works on real changes
  - **AC Ref:** PLAN.md Step 5

- [x] **Remove Old `hooks/pre-commit` Script:**
  - **Action:** Delete the legacy shell script located at `hooks/pre-commit` after confirming the new pre-commit framework setup is fully functional and documented.
  - **Depends On:** Update `hooks/README.md`
  - **AC Ref:** PLAN.md Step 6

- [ ] **Fix golangci-lint issues:**
  - **Action:** Address pre-existing golangci-lint issues that were identified during pre-commit hook testing. This ensures the linting process is enforced properly.
  - **Depends On:** None
  - **AC Ref:** PLAN.md (Clean Implementation)

- [ ] **Resolve go-unit-tests hook issues:**
  - **Action:** Investigate and fix the problems with the go-unit-tests pre-commit hook. Ensure tests run consistently in both the pre-commit environment and direct invocation.
  - **Depends On:** None
  - **AC Ref:** PLAN.md (Test Verification)

- [ ] **Add and commit changes:**
  - **Action:** Add and commit the following files:
    - `.pre-commit-config.yaml`
    - `scripts/check-large-files.sh`
    - Updated `hooks/README.md`
    - Updated `TODO.md`
  - **Depends On:** Fix golangci-lint issues, Resolve go-unit-tests hook issues
  - **AC Ref:** PLAN.md Step 7

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** The pre-commit tool needs to be installed by each developer. The installation instructions are included in the `hooks/README.md` file, but this doesn't guarantee that every developer will have pre-commit installed when setting up the project.
  - **Context:** PLAN.md Step 1 (Install pre-commit) and Step 5 (Update Documentation)

- [ ] **Issue/Assumption:** The specific hook revisions (`rev:` values) listed in PLAN.md Step 2 are the desired starting point, and any updates made by `pre-commit autoupdate` are acceptable after review.
  - **Context:** PLAN.md Step 2 (Create Configuration File) and Step 4 (autoupdate part)

- [ ] **Issue/Assumption:** The approach doesn't include changes to the CI workflow to ensure pre-commit checks run in CI as well, which could lead to inconsistencies between local development and CI environments.
  - **Context:** The overall implementation plan doesn't mention CI integration
