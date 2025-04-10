# TODO

## Initial Setup

- [x] **Create GitHub workflows directory**
  - **Action:** Create the `.github/workflows` directory in the repository root
  - **Depends On:** None
  - **AC Ref:** Primary Workflow setup

- [ ] **Configure GitHub repository settings**
  - **Action:** Enable required status checks for pull requests, set branch protection rules
  - **Depends On:** None
  - **AC Ref:** Workflow Trigger Events

## Primary CI Workflow

- [x] **Create CI workflow file**
  - **Action:** Create the main `.github/workflows/ci.yml` file with basic structure and trigger events for master branch
  - **Depends On:** Create GitHub workflows directory
  - **AC Ref:** Primary Workflow, Trigger Events

- [x] **Implement checkout job step**
  - **Action:** Add checkout code step using actions/checkout@v4
  - **Depends On:** Create CI workflow file
  - **AC Ref:** Setup step 1

- [ ] **Configure Go environment setup**
  - **Action:** Add step to set up Go environment with version caching using actions/setup-go@v5
  - **Depends On:** Implement checkout job step
  - **AC Ref:** Setup steps 2-3, Environment

- [ ] **Add dependency verification step**
  - **Action:** Add step to verify dependencies with go mod verify
  - **Depends On:** Configure Go environment setup
  - **AC Ref:** Lint job

- [ ] **Implement formatting check**
  - **Action:** Add step to check code formatting with go fmt
  - **Depends On:** Configure Go environment setup
  - **AC Ref:** Code Quality Checks step 1

- [ ] **Implement vet check**
  - **Action:** Add step to run go vet
  - **Depends On:** Configure Go environment setup
  - **AC Ref:** Code Quality Checks step 2

- [ ] **Implement golangci-lint**
  - **Action:** Add step to install and run golangci-lint
  - **Depends On:** Configure Go environment setup
  - **AC Ref:** Code Quality Checks step 3

- [ ] **Create test job**
  - **Action:** Create the test job with proper configuration
  - **Depends On:** Create CI workflow file
  - **AC Ref:** Testing

- [ ] **Add race detection testing**
  - **Action:** Implement step to run tests with race detection
  - **Depends On:** Create test job
  - **AC Ref:** Testing step 1

- [ ] **Add coverage reporting**
  - **Action:** Add step to generate test coverage reports
  - **Depends On:** Add race detection testing
  - **AC Ref:** Testing step 2

- [ ] **Implement coverage threshold check**
  - **Action:** Add step to verify coverage against threshold (80%)
  - **Depends On:** Add coverage reporting
  - **AC Ref:** Testing step 3

- [ ] **Create build job**
  - **Action:** Create the build job with proper configuration
  - **Depends On:** Create CI workflow file
  - **AC Ref:** Build

- [ ] **Implement build step**
  - **Action:** Add step to build the project
  - **Depends On:** Create build job
  - **AC Ref:** Build step 1

- [ ] **Add artifact upload**
  - **Action:** Add step to upload build artifacts with correct binary name 'architect'
  - **Depends On:** Implement build step
  - **AC Ref:** Build step 2

## Testing and Validation

- [ ] **Validate workflow syntax**
  - **Action:** Use GitHub Actions linter to validate workflow file syntax
  - **Depends On:** All CI workflow implementation tasks
  - **AC Ref:** Implementation Roadmap step 1

- [ ] **Create test pull request**
  - **Action:** Create a test PR to trigger and validate the workflow
  - **Depends On:** Validate workflow syntax
  - **AC Ref:** Implementation Roadmap step 2

- [ ] **Document workflow usage**
  - **Action:** Update README with information about the CI workflow
  - **Depends On:** Validate workflow syntax
  - **AC Ref:** General documentation

## Future Enhancements (Preparatory Tasks)

- [ ] **Research multi-platform testing setup**
  - **Action:** Research and document approach for multi-platform testing
  - **Depends On:** None
  - **AC Ref:** Future Enhancements 1

- [ ] **Research release automation workflow**
  - **Action:** Research and document approach for release automation
  - **Depends On:** None
  - **AC Ref:** Future Enhancements 2

- [ ] **Research CodeCov integration**
  - **Action:** Research and document approach for CodeCov integration
  - **Depends On:** Add coverage reporting
  - **AC Ref:** Future Enhancements 3

- [ ] **Research Dependabot setup**
  - **Action:** Research and document approach for Dependabot integration
  - **Depends On:** None
  - **AC Ref:** Future Enhancements 4

- [ ] **Research benchmark testing in CI**
  - **Action:** Research and document approach for benchmark testing in CI
  - **Depends On:** None
  - **AC Ref:** Future Enhancements 5

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** Specific coverage threshold value
  - **Context:** PLAN.md specifies failing if coverage drops below 80%. Need to confirm if 80% is the appropriate threshold for this project.

- [x] **Issue/Assumption:** Binary name for artifact upload
  - **Context:** Confirmed the binary name is 'architect' (not "architect-github-actions").

- [x] **Issue/Assumption:** Branch name for PR targets
  - **Context:** Confirmed the workflow should be configured for PRs targeting the master branch (not main).

- [ ] **Issue/Assumption:** Dependency on bc command
  - **Context:** The coverage threshold check uses the bc command which may not be available by default on all runners. Assuming it's available or should be installed.

- [ ] **Issue/Assumption:** Go version specification
  - **Context:** The workflow uses 'stable' for Go version. Need to confirm if a specific version should be pinned instead for consistency.

- [ ] **Issue/Assumption:** Required GitHub repository settings
  - **Context:** The workflow assumes certain repository settings for branch protection. Need to confirm the specific settings required.

- [ ] **Issue/Assumption:** Need for custom golangci-lint configuration
  - **Context:** The workflow adds golangci-lint but does not specify a configuration file. Need to confirm if a custom configuration is needed.