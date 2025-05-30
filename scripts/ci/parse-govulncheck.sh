#!/bin/bash
# scripts/ci/parse-govulncheck.sh
# Parses govulncheck JSON output and fails on Critical/High vulnerabilities

set -eo pipefail

# Read JSON stream from govulncheck
VULN_DATA=$(cat)

if [ -z "$VULN_DATA" ]; then
  echo "✅ No vulnerabilities detected by govulncheck"
  exit 0
fi

# Parse for Critical/High severity vulnerabilities
# govulncheck outputs JSON Lines format (one JSON object per line)
CRITICAL_HIGH_COUNT=0
CRITICAL_HIGH_DETAILS=""

while IFS= read -r line; do
  if [ -n "$line" ]; then
    # Extract severity if this is a finding with OSV data
    SEVERITY=$(echo "$line" | jq -r '.finding?.osv?.database_specific?.severity // .osv?.severity // empty' 2>/dev/null || true)

    if [ "$SEVERITY" = "CRITICAL" ] || [ "$SEVERITY" = "HIGH" ]; then
      CRITICAL_HIGH_COUNT=$((CRITICAL_HIGH_COUNT + 1))
      # Extract key details
      ID=$(echo "$line" | jq -r '.finding?.osv?.id // .osv?.id // "Unknown"' 2>/dev/null || true)
      PKG=$(echo "$line" | jq -r '.finding?.trace[0]?.module // "Unknown"' 2>/dev/null || true)
      CRITICAL_HIGH_DETAILS="${CRITICAL_HIGH_DETAILS}\n  - ${SEVERITY}: ${ID} in ${PKG}"
    fi
  fi
done <<< "$VULN_DATA"

if [ $CRITICAL_HIGH_COUNT -gt 0 ]; then
  echo "::error::❌ Found $CRITICAL_HIGH_COUNT Critical/High severity vulnerabilities:"
  echo -e "$CRITICAL_HIGH_DETAILS"
  exit 1
else
  echo "✅ No Critical or High severity vulnerabilities detected"
  exit 0
fi
