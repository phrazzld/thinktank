# Secret Detection CI Verification

This document explains how to verify that the secret detection tests are mandatory in CI and will fail the build if they fail.

## Background

As part of the project's security measures, we require that no API keys or other secrets are logged, especially as part of error handling. Secret detection tests in the `internal/providers/*/provider_secrets_test.go` files verify this behavior.

The CI pipeline has been configured to run these tests and fail if they detect any issues.

## Verifying Secret Detection Tests

The `scripts/ci/verify-secret-tests.sh` script checks that:

1. Secret detection test files exist for all providers (Gemini, OpenAI, OpenRouter)
2. These test files are included in the CI test run
3. The tests are passing

This script is now part of the CI pipeline and will fail the build if any of these conditions are not met.

## Manual Verification

To manually verify that CI will fail when a secret detection test fails, you can use the `scripts/ci/break-secret-test.sh` script:

```bash
# Temporarily break a secret detection test
./scripts/ci/break-secret-test.sh

# Verify that the test now fails
./scripts/ci/break-secret-test.sh --verify  # This should fail

# Restore the original test file
mv ./internal/providers/openai/provider_secrets_test.go.bak ./internal/providers/openai/provider_secrets_test.go
```

## CI Configuration

The CI pipeline (`.github/workflows/ci.yml`) includes a dedicated step to verify the secret detection tests:

```yaml
# Verify secret detection tests are running and passing
- name: Verify secret detection tests
  run: ./scripts/ci/verify-secret-tests.sh
  timeout-minutes: 2
```

This ensures that all provider secret detection tests are considered mandatory for the build to succeed.
