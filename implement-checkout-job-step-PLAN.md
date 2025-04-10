# Implement checkout job step

## Goal
Replace the placeholder checkout step in the CI workflow with a proper implementation using the actions/checkout@v4 GitHub Action.

## Implementation Approach
Update the `.github/workflows/ci.yml` file to:
1. Remove the placeholder checkout step in the lint job
2. Add a proper implementation using the actions/checkout@v4 GitHub Action
3. Add the checkout step to the test and build jobs as well to ensure all jobs have proper source code access

## Reasoning
1. **Consistency across jobs**: Adding the checkout step to all jobs (lint, test, and build) ensures each job has proper access to the source code. Even though we're currently only directed to update the lint job, it makes sense to ensure all jobs have the checkout step.

2. **Use standard implementation**: The actions/checkout@v4 action is the standard way to check out code in GitHub Actions. We'll use a standard implementation without any customization to keep things simple and consistent with the plan.

3. **Latest available version**: We're using v4 of the action as specified in the task, which is the latest version at the time of implementation.

4. **Clean implementation**: The checkout step should be the first step in each job since subsequent steps will need access to the repository.