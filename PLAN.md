# Automated Semantic Versioning via Conventional Commits (Go-Native Implementation)

## Chosen Approach (One‑liner)

Implement a robust, automated semantic versioning and release process using Go-native tools (`goreleaser`, `git-chglog`, `svu`, `lefthook`, `go-conventionalcommits`) for Conventional Commit enforcement, version calculation, changelog generation, and release automation, fully integrated into CI/CD.

## Architecture Blueprint

-   **Modules / Packages (Key Configuration & Tooling Components)**
    -   `.github/workflows/release.yml`: GitHub Actions workflow defining the CI/CD pipeline for linting, testing, versioning, and releasing.
    -   `.goreleaser.yml`: Configuration file for `goreleaser`, defining build, packaging, changelog, and release steps.
    -   `.lefthook.yml`: Configuration for `lefthook` defining pre-commit and commit-msg hooks, primarily for Conventional Commit linting.
    -   `.chglog/`: Directory containing `git-chglog` configuration (`config.yml`) and templates (`CHANGELOG.tpl.md`).
    -   `tools.go` (optional but recommended): A file to pin Go tool dependencies (e.g., `lefthook`, `go-conventionalcommits`, `svu`, `git-chglog`, `goreleaser`) ensuring consistent versions.
    -   `Makefile` (optional): Provides helper targets for installing tools, running hooks, and local release dry-runs.

-   **Public Interfaces / Contracts (Tool Interactions)**
    -   `lefthook` (via Git hooks):
        -   `commit-msg` hook: Validates commit message format using `go-conventionalcommits`.
    -   `go-conventionalcommits validate --message=<file>`: CLI command for commit message validation.
    -   `svu <next|current|major|minor|patch>`: CLI command to calculate/suggest semantic versions based on Git history and Conventional Commits.
    -   `git-chglog --next-tag <version> -o CHANGELOG.md`: CLI command to generate changelog entries.
    -   `goreleaser release --clean [--snapshot | --skip-publish]`: CLI command to orchestrate the release (tagging, building, changelog integration, publishing).
    -   `goreleaser check`: CLI command to validate `.goreleaser.yml` configuration.

-   **Data Flow Diagram**

    ```mermaid
    graph TD
        subgraph Local Development
            A[Developer writes code] --> B{git commit};
            B -- Commit Message --> C[Lefthook: commit-msg hook];
            C -- Runs --> D[go-conventionalcommits validate];
            D -- Valid --> E[Commit Successful];
            D -- Invalid --> F[Commit Aborted, Fix Message];
            E --> G{git push};
        end

        subgraph CI/CD Pipeline (e.g., on merge to main)
            G --> H[Checkout Code];
            H --> I[Setup Go & Tools];
            I --> J[Lint & Test Application];
            J -- Success --> K[Validate Commit Messages (go-conventionalcommits, optional redundancy)];
            K -- Success --> L[Determine Next Version (svu next)];
            L -- Version --> M[Generate Changelog (git-chglog)];
            M -- Changelog --> N[Execute Release (goreleaser release)];
            N -- Creates --> O[Git Tag];
            N -- Updates --> P[CHANGELOG.md in repo];
            N -- Publishes --> Q[GitHub Release + Artifacts];
        end
    ```

-   **Error & Edge‑Case Strategy**
    *   **Invalid Commit Message (Local):** `lefthook` + `go-conventionalcommits` aborts commit, forcing developer correction.
    *   **Invalid Commit Message (CI):** Redundant check in CI fails the build if hooks were bypassed.
    *   **Build/Test/Lint Failure in CI:** Pipeline stops before versioning/release steps.
    *   **`svu` Version Calculation Failure:** (e.g., no new conventional commits, ambiguous history) `svu` exits with error, CI fails. Requires manual inspection/correction of commit history or manual tagging.
    *   **`git-chglog` Failure:** (e.g., template error, git history issue) CI fails. Requires config/template fix or history adjustment.
    *   **`goreleaser check` Failure:** Invalid `.goreleaser.yml` fails CI early.
    *   **`goreleaser release` Failure:** (e.g., Git tagging conflict, artifact build error, GitHub API issue) CI fails. Requires investigation (e.g., token permissions, existing tag).
    *   **No Releasable Commits:** `svu` might indicate no version change. The pipeline should handle this gracefully (e.g., `goreleaser` might skip if no version bump).
    *   **Manual Version Override/Hotfix:** Document process for manual tagging (e.g., `git tag vX.Y.Z-fix`) and potentially manual changelog entry. CI should detect pushed tags and run release process.
    *   **GitHub Token Issues:** Insufficient permissions or invalid token for `goreleaser` to publish will cause release failure. Ensure correct scopes.

## Detailed Build Steps

1.  **Tool Installation & Pinning:**
    *   Create `tools.go` (e.g., in a `tools/` directory or project root) to pin versions of `github.com/evilmartians/lefthook`, `github.com/leodido/go-conventionalcommits`, `github.com/caarlos0/svu`, `github.com/git-chglog/git-chglog`, `github.com/goreleaser/goreleaser/v2/cmd/goreleaser`.
    *   Update `README.md` or create a `CONTRIBUTING.md` section with instructions for developers to install these tools (e.g., using a `Makefile` target `make tools` that runs `go install` for each tool listed in `tools.go`).

2.  **Local Commit Message Enforcement (Lefthook):**
    *   Initialize Lefthook: `lefthook install`.
    *   Create/configure `.lefthook.yml`:
        ```yaml
        # .lefthook.yml
        commit-msg:
          commands:
            validate-commit:
              run: go-conventionalcommits validate --message-file {1}
        ```
    *   Ensure developers are instructed to run `lefthook install` as part of project setup.

3.  **Semantic Version Utility (svu) Integration:**
    *   No specific configuration file for `svu`; it operates on Git history.
    *   In CI, this will be the first step in the release phase to determine the next version.

4.  **Changelog Generation (git-chglog):**
    *   Create `.chglog/config.yml` defining commit types, scopes, and changelog sections. Example:
        ```yaml
        # .chglog/config.yml
        style: conventional
        template: CHANGELOG.tpl.md
        info:
          title: CHANGELOG
          repository_url: https://github.com/thinktank/project # Replace with actual URL
        options:
          commits:
            filters:
              Type:
                - feat
                - fix
                - perf
                - refactor
                - revert
                - docs
                - style
                - chore
                - test
                - build
                - ci
          commit_groups:
            title_maps:
              feat: Features
              fix: Bug Fixes
              perf: Performance Improvements
              refactor: Code Refactoring
              # ... and so on
          header:
            pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
            pattern_maps:
              - Type
              - Scope
              - Subject
          notes:
            keywords:
              - BREAKING CHANGE
        ```
    *   Create `.chglog/CHANGELOG.tpl.md` template. A good default can be found in `git-chglog` examples.

5.  **Release Automation (Goreleaser):**
    *   Create `.goreleaser.yml`:
        *   Define `project_name`.
        *   Configure `before` hooks (e.g., `go mod tidy`).
        *   Define `builds` section for the Go application.
        *   Configure `release`:
            *   Set `disable: '{{ .Env.SKIP_RELEASE | default "false" }}'` to control actual publishing.
            *   `github`: Specify owner/repo.
            *   `prerelease: auto` (or based on commit messages if using `svu` for prereleases).
        *   Configure `changelog`:
            *   Set `use: git` and `skip: true`. This tells Goreleaser *not* to generate its own changelog but to expect one (which `git-chglog` will provide).
            *   Alternatively, configure Goreleaser's built-in changelog to match `git-chglog`'s output if preferred for tighter integration, but using `git-chglog` explicitly ensures adherence to its specific templating. A common pattern is for Goreleaser to pick up the `CHANGELOG.md` file.
        *   Configure `semver`:
            *   This section might not be needed if `svu` output is passed to `goreleaser` via an environment variable or file, and Goreleaser is instructed to use that version for tagging. Goreleaser can also infer version from Git tags.
    *   Ensure the changelog is correctly committed back to the repository as part of the release process. Goreleaser can be configured to commit it.

6.  **CI/CD Pipeline Integration (`.github/workflows/release.yml`):**
    *   **Trigger:** On pushes to `main` branch and on `v*` tags.
    *   **Jobs:**
        *   `lint_test_build`: Standard Go linting, testing, and application build checks.
        *   `release`: (runs after `lint_test_build` succeeds, conditional on `main` branch or tag)
            1.  Checkout code: `actions/checkout@v4` (with `fetch-depth: 0` for full history).
            2.  Set up Go: `actions/setup-go@v5`.
            3.  Install tools (from `tools.go` or cache them).
            4.  **Determine Next Version:** `VERSION=$(svu next)` and `echo "NEXT_VERSION=$VERSION" >> $GITHUB_ENV`.
            5.  **Generate Changelog:** `git-chglog --next-tag $VERSION -o RELEASE_CHANGELOG.md`.
            6.  **Configure Git User:** For committing changelog.
            7.  **Run Goreleaser:**
                *   For `main` branch builds (not tags): `goreleaser release --clean --snapshot --release-notes=RELEASE_CHANGELOG.md`. This validates the process.
                *   For `v*` tag builds OR if `NEXT_VERSION` indicates a new release on `main`:
                    *   Update `CHANGELOG.md` (prepend `RELEASE_CHANGELOG.md` content).
                    *   `git add CHANGELOG.md`
                    *   `git commit -m "chore(release): prepare for release $VERSION [skip ci]"`
                    *   `git tag $VERSION`
                    *   `git push origin $VERSION`
                    *   `git push` (to push changelog commit)
                    *   `goreleaser release --clean --release-notes=RELEASE_CHANGELOG.md` (Goreleaser will use the tag).
                    *   *Alternative Goreleaser Flow:* Let Goreleaser create the tag and commit the changelog. This requires careful configuration of `goreleaser`'s `git` and `changelog` sections. Goreleaser can auto-tag and then build.
            8.  Ensure `GITHUB_TOKEN` with `contents: write` (and `packages: write` if publishing Go modules/containers) permissions is available to the `goreleaser release` step.

7.  **Documentation Updates:**
    *   `README.md`: Add sections on commit conventions, release process, and local tooling setup.
    *   `CONTRIBUTING.md` (if exists): Detail commit message format and link to Conventional Commits spec.
    *   Document the first automated release in `CHANGELOG.md` manually or as part of the first automated run.

8.  **Initial Tagging & Migration:**
    *   If no tags exist, the first run of `svu next` will likely suggest `v0.1.0` (if there's a `feat:` commit).
    *   If existing tags are not semantic, decide on a starting semantic version and tag manually (e.g., `git tag v1.0.0`).

## Testing Strategy

-   **Test Layers:**
    -   **Local Hook Testing:** Developers verify `lefthook` + `go-conventionalcommits` by attempting valid/invalid commits.
    -   **CI Unit/Integration (Application):** Existing Go application tests (unaffected by this plan).
    -   **CI Integration (Release Process):**
        *   `goreleaser check` in CI to validate `.goreleaser.yml`.
        *   `goreleaser release --snapshot` runs on PRs or main branch pushes (without tagging/publishing) to verify build, changelog generation logic, and version calculation with `svu`.
    -   **E2E (Release Process):** The first few actual releases on `main` (or a staging branch) serve as E2E tests. Monitor closely. Dry-run modes for `git-chglog` and `svu` can be used in scripts for testing specific scenarios.

-   **What to Mock:** No mocking of Go code. For testing the release *pipeline scripts*, one might use a temporary Git repository with a crafted commit history to simulate different scenarios for `svu` and `git-chglog`.

-   **Coverage Targets & Edge‑Case Notes:**
    *   Ensure `lefthook` correctly blocks non-conventional commits.
    *   Verify `svu` correctly determines `major`, `minor`, `patch`, and prerelease versions based on commit messages (e.g., `feat:`, `fix:`, `feat!:`, `chore:`, `BREAKING CHANGE:` footer).
    *   Verify `git-chglog` generates `CHANGELOG.md` correctly, including all specified commit types and grouping.
    *   Test scenario: no releasable commits since last tag (should not create new release/tag).
    *   Test scenario: first release (no prior tags).

## Logging & Observability

-   **Log Events:**
    -   `lefthook`: Output to console on commit.
    -   `go-conventionalcommits`: Output to console on validation.
    -   `svu`, `git-chglog`, `goreleaser`: Standard CLI output captured in CI logs.
    -   CI workflow logs provide a full audit trail of the release process.
-   **Structured Fields:** Not directly applicable to these CLI tools, but CI logs should include commit SHAs, determined versions, and timestamps.
-   **Correlation ID Propagation:** The CI run ID (e.g., GitHub Actions `run_id`) serves as the correlation ID for the entire automated process.

## Security & Config

-   **Input Validation Hotspots:**
    -   Commit messages: Handled by `go-conventionalcommits`.
    -   `.goreleaser.yml`, `.lefthook.yml`, `.chglog/config.yml`: Schema validation by the respective tools (`goreleaser check`).
-   **Secrets Handling:**
    -   `GITHUB_TOKEN`: Stored as a GitHub Actions secret. Passed to `goreleaser` via environment variable in the CI workflow. **Never hardcode tokens.**
    -   Ensure CI logs do not inadvertently print secrets. `goreleaser` is generally good about this.
-   **Least‑Privilege Notes:**
    -   The `GITHUB_TOKEN` used for releases should have the minimum necessary permissions (e.g., `contents: write` for tagging and creating releases, `packages: write` if publishing packages).
    -   Local hooks (`lefthook`) run with developer's local permissions; they should not require network access for validation.

## Documentation

-   **Code Self‑Doc Patterns:**
    -   Comments in `.goreleaser.yml`, `.lefthook.yml`, `.github/workflows/release.yml` explaining non-obvious configurations.
    -   `Makefile` targets should be commented.
-   **Required Readme/Contributing Updates:**
    -   `README.md` or `CONTRIBUTING.md`:
        *   Section on "Commit Message Guidelines" referencing Conventional Commits.
        *   Examples of valid commit messages for `feat`, `fix`, `BREAKING CHANGE`.
        *   Instructions for setting up local development environment, including `lefthook install` and installing Go tools (via `tools.go` and `Makefile`).
    -   `CHANGELOG.md`: Will be auto-generated. A note about its auto-generation can be added to its header or `README.md`.
    -   Briefly describe the automated release process in `README.md`.

## Risk Matrix

| Risk                                                     | Severity | Mitigation                                                                                                                               |
| -------------------------------------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Developers bypass local commit hooks                     | Medium   | CI enforces Conventional Commits; build fails on non-compliant messages. Document policy.                                                |
| Incorrect version calculation by `svu`                   | High     | Thorough testing with diverse commit histories. Manual review of `svu next` output in CI before tagging. Clear error messages from `svu`.   |
| Flawed `CHANGELOG.md` generation by `git-chglog`         | Medium   | Test templates thoroughly. Review generated changelog during `--snapshot` CI runs. Manually correct if necessary before final release.         |
| `goreleaser` misconfiguration leading to failed release  | High     | Use `goreleaser check`. Run `goreleaser release --snapshot` in CI. Incremental configuration and testing.                                    |
| GitHub token exposure or insufficient permissions        | Critical | Store token as CI secret. Use least privilege. Audit CI workflow for token handling.                                                       |
| Tool versioning conflicts/drift                          | Low      | Pin tool versions using `tools.go` and `go.mod`. Ensure CI uses pinned versions.                                                           |
| Complexity introduces high maintenance overhead          | Medium   | Start with simplest viable configuration. Document choices. Ensure team understanding. Favor Goreleaser's built-in features where possible. |
| Migration from existing versioning (if any) is disruptive | Medium   | Clearly document migration path. Perform initial semantic tag carefully. Communicate changes to the team.                                  |
| CI pipeline becomes slow due to added steps              | Low      | Optimize tool installation (caching). Ensure tools run efficiently. Most steps are quick.                                                  |

## Open Questions

1.  **Initial Version:** What should be the initial semantic version if the project already has untagged history or non-semantic tags? (Recommendation: Manually decide and `git tag vX.Y.Z` before enabling full automation).
2.  **Prerelease Workflow:** How will prereleases (e.g., `alpha`, `beta`, `rc`) be handled? `svu` and `goreleaser` support this, but the workflow needs definition (e.g., triggered from specific branches or specific commit message conventions).
3.  **Hotfix Workflow:** What is the precise process for hotfixes on older released versions? (e.g., create branch from tag, cherry-pick fix, commit with `fix:`, tag `vX.Y.Z+1`, push tag to trigger release).
4.  **Handling of `CHANGELOG.md` Commits:** Who/what commits the updated `CHANGELOG.md`? (Recommendation: The CI release job should commit it back to the repository before tagging and pushing the tag).
5.  **Developer Tool Installation:** Is `make tools` using `tools.go` sufficient, or are pre-built binaries/other installation methods needed for developers without a full Go environment for hooks? (Recommendation: Stick to Go toolchain as per project context).