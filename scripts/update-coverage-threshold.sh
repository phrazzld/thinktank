#!/bin/bash
set -e

# update-coverage-threshold.sh - Manage incremental coverage threshold progression
# Usage: ./scripts/update-coverage-threshold.sh [target_threshold]

TARGET_THRESHOLD=${1:-90}
CURRENT_THRESHOLD=$(grep "check-coverage-cached.sh" .pre-commit-config.yaml | grep -o "[0-9]\+" | head -1)

echo "Current pre-commit coverage threshold: ${CURRENT_THRESHOLD}%"
echo "Target coverage threshold: ${TARGET_THRESHOLD}%"

# Calculate current coverage to see if we can increase threshold
echo "Checking current coverage..."
CURRENT_COVERAGE=$(./scripts/check-coverage.sh 1 2>/dev/null | grep "Total code coverage:" | \
  awk '{print $4}' | tr -d '%' | cut -d. -f1 || echo "0")

echo "Current coverage: ${CURRENT_COVERAGE}%"

# Determine if we can safely increase threshold
if [ "$CURRENT_COVERAGE" -gt "$CURRENT_THRESHOLD" ]; then
  # Calculate safe increment (current coverage minus 1% buffer)
  SAFE_THRESHOLD=$((CURRENT_COVERAGE - 1))

  # Don't exceed target threshold
  if [ "$SAFE_THRESHOLD" -gt "$TARGET_THRESHOLD" ]; then
    SAFE_THRESHOLD=$TARGET_THRESHOLD
  fi

  # Only increase threshold if it's meaningful (at least 1% improvement)
  if [ "$SAFE_THRESHOLD" -gt "$CURRENT_THRESHOLD" ]; then
    echo "✅ Coverage allows threshold increase to ${SAFE_THRESHOLD}%"

    # Update .pre-commit-config.yaml
    sed -i.bak "s/check-coverage-cached.sh ${CURRENT_THRESHOLD}/check-coverage-cached.sh ${SAFE_THRESHOLD}/" .pre-commit-config.yaml

    # Update description
    if [ "$SAFE_THRESHOLD" -eq "$TARGET_THRESHOLD" ]; then
      sed -i.bak "s/target ${TARGET_THRESHOLD}%/achieved ${TARGET_THRESHOLD}% target/" .pre-commit-config.yaml
    else
      sed -i.bak "s/baseline protection/incremental progress/" .pre-commit-config.yaml
    fi

    rm .pre-commit-config.yaml.bak

    echo "📝 Updated pre-commit threshold from ${CURRENT_THRESHOLD}% to ${SAFE_THRESHOLD}%"

    # Test the new threshold
    if ./scripts/check-coverage-cached.sh "$SAFE_THRESHOLD"; then
      echo "✅ New threshold validated successfully"
    else
      echo "❌ New threshold validation failed, reverting..."
      # Revert changes
      sed -i.bak "s/check-coverage-cached.sh ${SAFE_THRESHOLD}/check-coverage-cached.sh ${CURRENT_THRESHOLD}/" .pre-commit-config.yaml
      rm .pre-commit-config.yaml.bak
      exit 1
    fi
  else
    echo "ℹ️  Coverage (${CURRENT_COVERAGE}%) not high enough for meaningful threshold increase"
  fi
else
  echo "⚠️  Coverage (${CURRENT_COVERAGE}%) is below current threshold (${CURRENT_THRESHOLD}%)"
  echo "    Focus on improving coverage before increasing threshold"
fi

echo ""
echo "Coverage improvement suggestions:"
echo "1. Run: ./scripts/check-coverage.sh 90 # See packages below 90%"
echo "2. Focus on packages with lowest coverage first"
echo "3. Use: go test -coverprofile=coverage.out ./package && go tool cover -html=coverage.out"
