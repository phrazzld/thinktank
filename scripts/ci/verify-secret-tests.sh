#!/bin/bash
# verify-secret-tests.sh
# Verifies that provider secret detection tests are mandatory for build success

set -e  # Exit immediately if a command exits with a non-zero status

# Get module path
MODULE_PATH=$(grep -E '^module\s+' go.mod | awk '{print $2}')

# Check Gemini provider
provider="gemini"
test_function="TestGeminiProviderSecretHandling"
test_file="internal/providers/${provider}/provider_secrets_test.go"
if [ -f "$test_file" ]; then
  echo "✓ Secret detection tests exist for ${provider} provider"
else
  echo "✗ ERROR: Secret detection tests missing for ${provider} provider"
  exit 1
fi

package="${MODULE_PATH}/internal/providers/${provider}"
if echo "$package" | grep -q -E "internal/integration|internal/e2e|/disabled/"; then
  echo "✗ ERROR: ${provider} secret detection tests are excluded from the CI test run"
  exit 1
else
  echo "✓ Secret detection tests for ${provider} are included in CI test run"
fi

echo "Running secret detection tests for ${provider} (${test_function})..."
go test -run "${test_function}" "${package}" || {
  echo "✗ ERROR: Secret detection tests failed for ${provider} provider"
  exit 1
}

# Check OpenAI provider
provider="openai"
test_function="TestOpenAIProviderSecretHandling"
test_file="internal/providers/${provider}/provider_secrets_test.go"
if [ -f "$test_file" ]; then
  echo "✓ Secret detection tests exist for ${provider} provider"
else
  echo "✗ ERROR: Secret detection tests missing for ${provider} provider"
  exit 1
fi

package="${MODULE_PATH}/internal/providers/${provider}"
if echo "$package" | grep -q -E "internal/integration|internal/e2e|/disabled/"; then
  echo "✗ ERROR: ${provider} secret detection tests are excluded from the CI test run"
  exit 1
else
  echo "✓ Secret detection tests for ${provider} are included in CI test run"
fi

echo "Running secret detection tests for ${provider} (${test_function})..."
go test -run "${test_function}" "${package}" || {
  echo "✗ ERROR: Secret detection tests failed for ${provider} provider"
  exit 1
}

# Check OpenRouter provider
provider="openrouter"
test_function="TestOpenRouterProviderSecretHandling"
test_file="internal/providers/${provider}/provider_secrets_test.go"
if [ -f "$test_file" ]; then
  echo "✓ Secret detection tests exist for ${provider} provider"
else
  echo "✗ ERROR: Secret detection tests missing for ${provider} provider"
  exit 1
fi

package="${MODULE_PATH}/internal/providers/${provider}"
if echo "$package" | grep -q -E "internal/integration|internal/e2e|/disabled/"; then
  echo "✗ ERROR: ${provider} secret detection tests are excluded from the CI test run"
  exit 1
else
  echo "✓ Secret detection tests for ${provider} are included in CI test run"
fi

echo "Running secret detection tests for ${provider} (${test_function})..."
go test -run "${test_function}" "${package}" || {
  echo "✗ ERROR: Secret detection tests failed for ${provider} provider"
  exit 1
}

echo "✓ All provider secret detection tests verified as mandatory in CI"
exit 0
