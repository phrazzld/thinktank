# GitHub Actions CI/CD Pipeline

This document explains the GitHub Actions continuous integration and continuous deployment (CI/CD) pipeline used in the thinktank project.

## Overview

thinktank uses GitHub Actions to automate the following processes:
- Code quality checks (linting)
- Automated testing
- Build verification

This automation helps ensure that all code changes maintain quality standards and don't introduce regressions before they're merged into the main codebase.

## Workflow Architecture

The CI/CD pipeline follows a **single sequential job** approach, where linting, testing, and building steps run in sequence within a single GitHub Actions job. This approach was chosen for its simplicity, clarity, and alignment with local development workflows.

Key architectural decisions:
- **Single Job**: All steps run sequentially in one job for simplicity and clarity
- **Fail-Fast Principle**: If linting fails, the workflow stops before running tests or building
- **Caching Strategy**: Uses GitHub's caching action to store pnpm dependencies between runs
- **Node.js Environment**: Uses Node.js 20 running on Ubuntu latest

## Workflow Configuration

The workflow is defined in [.github/workflows/ci.yml](.github/workflows/ci.yml) and includes the following components:

### Triggers

The workflow runs on:
- Push to the `master` branch
- Pull requests targeting the `master` branch

```yaml
on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]
```

### Job Configuration

The workflow consists of a single job named "Lint, Test, Build" that runs on the latest Ubuntu runner:

```yaml
jobs:
  ci:
    name: Lint, Test, Build
    runs-on: ubuntu-latest
    # ...steps defined below
```

### Steps

The workflow executes the following steps in sequence:

1. **Checkout Code**
   - Uses `actions/checkout@v4` to fetch the repository code
   - Essential first step for any workflow that needs to access the repository contents

2. **Set up Node.js**
   - Uses `actions/setup-node@v4` to configure Node.js environment
   - Sets up Node.js 20, which is the current LTS version

3. **Set up pnpm**
   - Uses `pnpm/action-setup@v4` to install and configure pnpm
   - Uses the latest stable version of pnpm

4. **Configure pnpm Cache**
   - Gets the pnpm store path and configures caching using `actions/cache@v4`
   - Significantly improves workflow performance by reusing cached dependencies
   - Cache key is based on the pnpm lock file to ensure proper invalidation when dependencies change

5. **Install Dependencies**
   - Runs `pnpm install --frozen-lockfile` to install project dependencies
   - Uses frozen lockfile to ensure consistent installations

6. **Run Linter**
   - Executes `pnpm run lint` to check code quality
   - Fails the workflow if there are linting errors

7. **Run Tests**
   - Executes `pnpm test` to run the test suite
   - Verifies functionality and prevents regressions

8. **Build Project**
   - Executes `pnpm run build` to verify the build process
   - Ensures the project can be built successfully

## Interpreting Workflow Results

The workflow results can be found in the "Actions" tab of the GitHub repository. Here's how to interpret the results:

- **Green Checkmark**: All steps completed successfully
- **Red X**: At least one step failed
- **Yellow Dot**: Workflow is currently running

Click on a specific workflow run to see detailed results for each step. If a step fails, you can expand it to see the error message and logs.

## Common Issues and Troubleshooting

### Linting Errors

If the workflow fails at the "Run linter" step:
1. Look at the error message in the workflow logs
2. Fix the linting issues locally using `pnpm run lint`
3. Consider using `pnpm run lint:fix` to automatically fix some issues
4. Commit and push the fixes

### Test Failures

If the workflow fails at the "Run tests" step:
1. Check the test output in the workflow logs
2. Run the tests locally to reproduce the issue
3. Fix the failing tests or the code causing the failures
4. Commit and push the fixes

### Build Errors

If the workflow fails at the "Build project" step:
1. Review the build error messages
2. Try building locally with `pnpm run build` to reproduce
3. Fix any type errors or build configuration issues
4. Commit and push the fixes

### Caching Issues

If you suspect caching problems:
1. Look for the "Setup pnpm cache" step in the workflow logs
2. Check if it shows "Cache hit" or "Cache miss"
3. If you need to force a cache refresh, you can:
   - Update the `pnpm-lock.yaml` file by adding or updating a dependency
   - Clear the cache manually in the GitHub Actions cache management UI

## Best Practices for Contributors

1. **Run Checks Locally Before Pushing**
   - Run `pnpm run lint`, `pnpm test`, and `pnpm run build` locally
   - Fix any issues before pushing to avoid failing workflow runs

2. **Watch Your Workflow Runs**
   - After pushing, monitor the workflow status on GitHub
   - Address any failures promptly

3. **Keep PRs Focused**
   - Smaller, focused PRs are easier to test and review
   - They also run faster in the CI pipeline

4. **Understand Caching**
   - The workflow caches dependencies to speed up runs
   - Major dependency changes may require a full cache refresh

5. **Reference Workflow Runs in PR Discussions**
   - When discussing issues, reference specific workflow runs by linking to them
   - This provides context for troubleshooting

## Future Enhancements

Potential future enhancements to the workflow include:

1. **Parallel Jobs**: Splitting linting, testing, and building into parallel jobs for speed
2. **Matrix Testing**: Testing across multiple Node.js versions or operating systems
3. **Code Coverage Reports**: Adding steps to generate and publish code coverage information
4. **Automated Releases**: Adding steps to automate version bumping and release creation
5. **Deployment**: Configuring automatic deployment to staging or production environments

## Workflow Development

If you need to modify the workflow configuration:

1. Create a feature branch
2. Edit `.github/workflows/ci.yml`
3. Test your changes by pushing the branch and observing the workflow run
4. Request a review before merging to `master`

Remember that workflow configuration changes affect the entire project, so careful testing and review are essential.