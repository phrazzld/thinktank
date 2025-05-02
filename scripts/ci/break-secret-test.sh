#!/bin/bash
# break-secret-test.sh
# Temporarily breaks a secret detection test to verify that CI fails
# DO NOT USE IN PRODUCTION - only for testing the CI configuration

set -e

# Check if --verify flag is provided (to restore the file)
if [ "$1" == "--verify" ]; then
  echo "Verifying CI by running the secret detection tests..."
  ./scripts/ci/verify-secret-tests.sh
  exit $?
fi

# Get the path to a provider_secrets_test.go file
test_file="./internal/providers/openai/provider_secrets_test.go"

# Create a backup of the file
cp "$test_file" "${test_file}.bak"

# Break the test by modifying it to intentionally leak a secret
# This should cause the CI check to fail
sed -i.broken 's/secretLogger.HasDetectedSecrets()/false/' "$test_file"

echo "WARNING: Test has been intentionally broken to verify CI failure."
echo "Run '$0 --verify' to check if the break works."
echo "Remember to restore the file by running: mv '${test_file}.bak' '$test_file'"
