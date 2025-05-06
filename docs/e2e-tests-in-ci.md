# E2E Tests in CI

This document describes how End-to-End (E2E) tests for the `thinktank` project are enforced in the Continuous Integration (CI) pipeline.

## Overview

E2E tests are a critical part of our testing strategy. They verify that the `thinktank` CLI functions correctly from the user's perspective by testing the compiled binary with various arguments and verifying the output.

## CI Configuration

E2E tests are integrated into our CI pipeline in the following ways:

1. **Dedicated CI Step**: The E2E tests are run as a dedicated step in the `test` job of our CI workflow.
2. **Full Test Coverage**: The tests are run without the `--short` flag to ensure comprehensive coverage.
3. **Proper Timeout**: A 15-minute timeout is set to accommodate the full test suite.
4. **Blocking Status**: The E2E tests are mandatory and must pass for the CI build to succeed.

## Branch Protection Setup

To properly enforce E2E tests in your GitHub repository, follow these steps:

1. Go to your GitHub repository
2. Click on "Settings" -> "Branches"
3. Under "Branch protection rules", click "Add rule"
4. In "Branch name pattern", enter "master" (or your main branch name)
5. Enable "Require status checks to pass before merging"
6. In the search box, find and select the status check for your CI workflow (usually named "Test")
7. Make sure "Require branches to be up to date before merging" is checked
8. Click "Create" or "Save changes"

This configuration ensures that:
- All pull requests to the master branch require the CI checks to pass
- The E2E tests are part of these required checks
- No code can be merged to master if the E2E tests fail

## Verifying the Setup

To verify the setup is working correctly:

1. Create a branch with a failing E2E test
2. Push the branch and create a pull request
3. Observe that the CI build fails and the PR cannot be merged
4. Fix the test and push again
5. Observe that the CI build passes and the PR can be merged

## Troubleshooting

If the E2E tests fail in CI but pass locally, consider the following:

1. **Environment Differences**: CI environment might be different from your local environment
2. **API Limitations**: Mock API behavior might differ in the CI environment
3. **Timing Issues**: Some tests might be sensitive to timing issues that are more pronounced in CI

Check the CI logs for detailed error messages to help diagnose the issue.

## Local Testing

To run the E2E tests locally with the same configuration as CI, use:

```bash
./internal/e2e/run_e2e_tests.sh --verbose
```

This will run the full E2E test suite without the `--short` flag, matching the CI configuration.
