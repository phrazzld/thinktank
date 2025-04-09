# TODO

## GitHub Actions CI/CD Setup

- [x] **Create GitHub workflows directory**
  - **Action:** Create the `.github/workflows` directory in the project root.
  - **Depends On:** None
  - **AC Ref:** Implementation Steps 1 (PLAN.md line 124)

- [x] **Create CI workflow file**
  - **Action:** Create the `ci.yml` file in the `.github/workflows` directory with the basic structure and workflow name.
  - **Depends On:** Create GitHub workflows directory
  - **AC Ref:** Implementation Steps 2 (PLAN.md line 125), Basic GitHub Actions Workflow (PLAN.md lines 16-34)

- [x] **Configure workflow triggers**
  - **Action:** Add the trigger configuration for push and pull requests to the master branch.
  - **Depends On:** Create CI workflow file
  - **AC Ref:** Workflow Triggers (PLAN.md lines 21-34), Success Criteria 1 (PLAN.md line 135)

- [x] **Set up workflow job configuration**
  - **Action:** Configure the single job with name, runner (ubuntu-latest), and initial steps structure.
  - **Depends On:** Configure workflow triggers
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 36-88)

- [x] **Add code checkout step**
  - **Action:** Add the step to checkout the repository code using `actions/checkout@v4`.
  - **Depends On:** Set up workflow job configuration
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 47-48)

- [x] **Configure Node.js setup**
  - **Action:** Add the step to set up Node.js using `actions/setup-node@v4`, specifying at least version 18 as per package.json requirements.
  - **Depends On:** Add code checkout step
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 50-53)

- [x] **Configure pnpm setup**
  - **Action:** Add the step to set up pnpm using `pnpm/action-setup@v4`.
  - **Depends On:** Configure Node.js setup
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 55-58)

- [x] **Implement pnpm caching**
  - **Action:** Add steps to get the pnpm store path and set up caching using `actions/cache@v4`.
  - **Depends On:** Configure pnpm setup
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 60-72), Caching Strategy (PLAN.md lines 104-106)

- [x] **Add dependency installation step**
  - **Action:** Add the step to install project dependencies using `pnpm install --frozen-lockfile`.
  - **Depends On:** Implement pnpm caching
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 74-75)

- [x] **Add linting step**
  - **Action:** Add the step to run the linter using `pnpm run lint`.
  - **Depends On:** Add dependency installation step
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 77-78), Key Goals 1 (PLAN.md line 9), Success Criteria 2 (PLAN.md line 136)

- [x] **Add testing step**
  - **Action:** Add the step to run tests using `pnpm test`.
  - **Depends On:** Add linting step
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 80-84), Key Goals 2 (PLAN.md line 10), Success Criteria 3 (PLAN.md line 137)

- [x] **Add build step**
  - **Action:** Add the step to build the project using `pnpm run build`.
  - **Depends On:** Add testing step
  - **AC Ref:** Define Single Job Workflow (PLAN.md lines 86-87), Key Goals 3 (PLAN.md line 11), Success Criteria 4 (PLAN.md line 138)

- [x] **Test workflow with deliberate lint error**
  - **Action:** Create a temporary branch with a deliberate lint error to verify that the CI workflow fails at the lint step.
  - **Depends On:** Add build step
  - **AC Ref:** Testing the CI Workflow 1 (PLAN.md line 145), Success Criteria 2 (PLAN.md line 136)
  - **Test Results:** Created file with lint errors (src/test-ci-lint-error.ts), pushed to GitHub. The workflow should fail at the lint step. To verify: check GitHub Actions tab for the workflow run.

- [ ] **Test workflow with deliberate test failure**
  - **Action:** Create a temporary branch with a failing test to verify that the CI workflow fails at the test step.
  - **Depends On:** Test workflow with deliberate lint error
  - **AC Ref:** Testing the CI Workflow 2 (PLAN.md line 146), Success Criteria 3 (PLAN.md line 137)

- [ ] **Test workflow with deliberate build error**
  - **Action:** Create a temporary branch with a build error to verify that the CI workflow fails at the build step.
  - **Depends On:** Test workflow with deliberate test failure
  - **AC Ref:** Testing the CI Workflow 3 (PLAN.md line 147), Success Criteria 4 (PLAN.md line 138)

- [ ] **Test workflow with clean code**
  - **Action:** Create a clean branch to verify that the CI workflow passes all checks.
  - **Depends On:** Test workflow with deliberate build error
  - **AC Ref:** Testing the CI Workflow 4 (PLAN.md line 148), Success Criteria 1-5 (PLAN.md lines 135-139)

- [ ] **Verify workflow completion time**
  - **Action:** Measure and verify that the workflow completes in less than 5 minutes.
  - **Depends On:** Test workflow with clean code
  - **AC Ref:** Success Criteria 5 (PLAN.md line 139)

- [ ] **Document GitHub Actions workflow**
  - **Action:** Add documentation about the CI/CD workflow in README.md or a separate documentation file.
  - **Depends On:** Verify workflow completion time
  - **AC Ref:** Key Goals 4 (PLAN.md line 12)

- [ ] **Final review and PR preparation**
  - **Action:** Review all implementation details, ensure all tests pass, and prepare for PR submission.
  - **Depends On:** Document GitHub Actions workflow
  - **AC Ref:** Implementation Steps 5-6 (PLAN.md lines 128-129)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Environment Variables for Tests**
  - **Context:** PLAN.md lines 82-84 mention potential environment variables for tests, but doesn't specify which ones might be needed for this project.
  - **Assumption:** We will initially implement without environment variables in the workflow, and add them later if specific tests require them.

- [ ] **Node.js Version Selection**
  - **Context:** PLAN.md line 53 suggests using Node.js 20 or matching package.json engines field.
  - **Assumption:** We'll use Node.js 20 as suggested, even though package.json only requires >=18.0.0, as using a more recent LTS version is generally preferable unless there are specific compatibility issues.

- [ ] **Maintenance of Workflow Performance**
  - **Context:** Success Criteria 5 (PLAN.md line 139) specifies that workflow should complete in <5 minutes.
  - **Assumption:** We'll monitor the workflow execution time during initial implementation and testing phases. If it approaches or exceeds 5 minutes, we'll investigate optimizations within the single-job approach before considering multi-job parallelism.

- [ ] **Branch Protection Rules**
  - **Context:** The plan doesn't mention configuring branch protection rules for the main branch.
  - **Assumption:** Branch protection rules (requiring CI to pass before merging to main) will be set up separately in GitHub repository settings, not as part of this implementation task.

- [ ] **Testing Strategy for Workflow File**
  - **Context:** Testing the CI Workflow section (PLAN.md lines 142-148) outlines a manual approach to testing.
  - **Assumption:** We will follow manual testing as described, rather than implementing automated tests for the workflow itself, which would add significant complexity.