# Quality Gate Feature Flags

This document describes the feature flag system for enabling and disabling quality gates in CI/CD workflows.

## Overview

The quality gate feature flag system allows you to:

- **Enable/disable individual quality gates** without modifying workflow YAML files
- **Control whether gate failures block the pipeline** (required vs optional)
- **Gradually roll out new quality gates** with safe defaults
- **Temporarily disable problematic gates** during incidents
- **Maintain backwards compatibility** when introducing new gates

## Configuration File

The feature flags are controlled by `.github/quality-gates-config.yml`:

```yaml
version: "1.0"

# CI Workflow Gates (ci.yml)
ci_gates:
  lint:
    enabled: true     # Whether the gate runs at all
    required: true    # Whether failure blocks the pipeline
    description: "Code formatting, linting, and style checks"

  vulnerability_scan:
    enabled: true
    required: true
    description: "Dependency vulnerability scanning with govulncheck"

  test:
    enabled: true
    required: true
    description: "Test execution and coverage validation"

  build:
    enabled: true
    required: true
    description: "Binary compilation and build artifact generation"

# Security Gates Workflow (security-gates.yml)
security_gates:
  secret_scan:
    enabled: true
    required: true
    description: "Secret and credential detection in codebase"

  license_scan:
    enabled: true
    required: true
    description: "Dependency license compliance validation"

  sast_scan:
    enabled: true
    required: true
    description: "Static security analysis with gosec"

# Additional gates
additional_gates:
  quality_dashboard:
    enabled: true
    required: false    # Dashboard generation shouldn't block PRs
    description: "Quality metrics dashboard generation and deployment"

# Override behavior
override_settings:
  allow_emergency_override: true    # Whether emergency overrides can bypass feature flags
  track_disabled_gates: false      # Whether to create issues for disabled required gates
```

## Gate States

Each quality gate can be in one of these states:

| enabled | required | Behavior |
|---------|----------|----------|
| `true`  | `true`   | Gate runs and failure blocks the pipeline |
| `true`  | `false`  | Gate runs but failure doesn't block (informational) |
| `false` | `true`   | Gate is completely skipped |
| `false` | `false`  | Gate is completely skipped |

When `enabled: false`, the `required` setting is ignored and the gate is skipped entirely.

## How It Works

### 1. Configuration Reading

Each workflow includes a `read-config` job that:
- Reads `.github/quality-gates-config.yml`
- Parses the YAML and extracts gate settings
- Outputs boolean values for each gate's enabled/required status
- Falls back to safe defaults if the config file is missing

### 2. Conditional Job Execution

Quality gate jobs use conditional `if:` statements:

```yaml
jobs:
  read-config:
    name: Read Quality Gate Configuration
    runs-on: ubuntu-latest
    outputs:
      lint_enabled: ${{ steps.config.outputs.lint_enabled }}
      # ... other outputs

  lint:
    name: Lint and Format
    runs-on: ubuntu-latest
    needs: [read-config]
    if: needs.read-config.outputs.lint_enabled == 'true'
    # Job only runs if lint is enabled
```

### 3. Integration with Emergency Overrides

Feature flags work alongside the existing emergency override system:

```yaml
test:
  needs: [read-config, check-override]
  if: |
    always() &&
    needs.read-config.outputs.test_enabled == 'true' &&
    (needs.check-override.outputs.bypass_tests != 'true' || github.event_name == 'push')
```

The gate must be both:
- Enabled via feature flag AND
- Not bypassed via emergency override

## Usage Examples

### Temporarily Disable a Problematic Gate

If the license scan is causing false positives:

```yaml
security_gates:
  license_scan:
    enabled: false  # Temporarily disable
    required: true
```

### Make a Gate Informational Only

To make the SAST scan informational while tuning it:

```yaml
security_gates:
  sast_scan:
    enabled: true
    required: false  # Won't block PRs but will provide feedback
```

### Gradually Roll Out a New Gate

When introducing a new performance regression gate:

```yaml
additional_gates:
  performance_regression:
    enabled: false    # Start disabled
    required: false   # Non-blocking when enabled
```

Then gradually enable it:
1. Set `enabled: true, required: false` to gather data
2. Monitor results and tune thresholds
3. Set `required: true` when confident

## Testing

Test the feature flag system locally:

```bash
# Run the test suite
./scripts/test-feature-flags.sh

# Test specific configurations
yq eval '.ci_gates.lint.enabled' .github/quality-gates-config.yml
```

Test in GitHub Actions:
1. Create a PR that modifies `.github/quality-gates-config.yml`
2. Disable a gate (e.g., set `ci_gates.lint.enabled: false`)
3. Verify the corresponding job is skipped in the workflow run

## Workflow Integration

### CI Workflow (`ci.yml`)

Gates controlled by feature flags:
- `lint` - Code quality checks
- `vulnerability-scan` - Dependency vulnerability scanning
- `test` - Test execution and coverage
- `build` - Binary compilation

### Security Gates Workflow (`security-gates.yml`)

Gates controlled by feature flags:
- `secret-scan` - Secret detection with TruffleHog
- `license-scan` - License compliance checking
- `sast-scan` - Static security analysis

## Best Practices

### 1. Default to Enabled

New gates should default to `enabled: true` to maintain security posture:

```yaml
new_gate:
  enabled: true     # Safe default
  required: true    # Enforce by default
```

### 2. Use Descriptive Names

Gate names should be clear and match the workflow job names:

```yaml
ci_gates:
  lint:              # Matches job name in ci.yml
    enabled: true
```

### 3. Document Changes

When modifying feature flags, document the reason in the commit message:

```
feat: disable SAST scan temporarily

SAST scan is generating false positives for crypto/rand usage.
Disabling temporarily while we tune the gosec configuration.

Tracking issue: #123
```

### 4. Monitor Disabled Gates

When disabling required gates:
- Create a tracking issue
- Set a timeline for re-enabling
- Monitor for security impact

### 5. Test Thoroughly

Before disabling production gates:
- Test in a feature branch first
- Verify the gate is actually skipped
- Check that dependent jobs still run correctly

## Emergency Procedures

### Incident Response

During incidents, you can quickly disable problematic gates:

1. **Immediate relief**: Modify `.github/quality-gates-config.yml`
2. **Create tracking issue**: Document what was disabled and why
3. **Fix root cause**: Address the underlying issue
4. **Re-enable gates**: Restore normal operation

### Example: Disable All Security Gates

```yaml
security_gates:
  secret_scan:
    enabled: false
  license_scan:
    enabled: false
  sast_scan:
    enabled: false
```

### Example: Make All Gates Informational

```yaml
ci_gates:
  lint:
    enabled: true
    required: false  # Won't block
  test:
    enabled: true
    required: false  # Won't block
```

## Troubleshooting

### Config File Not Found

If `.github/quality-gates-config.yml` is missing, the system falls back to safe defaults (all gates enabled and required).

### YAML Syntax Errors

The action includes error handling for malformed YAML:
- Invalid YAML falls back to defaults
- Warnings are logged in the workflow output

### Gate Still Running When Disabled

Check:
1. YAML syntax is correct
2. Gate name matches exactly (case-sensitive)
3. The `read-config` job completed successfully
4. The job's `if:` condition references the correct output

### Dependencies Between Gates

When disabling a gate, consider its dependents:

```yaml
# If you disable 'lint', 'test' might need adjustment
test:
  needs: [read-config, check-override, lint, vulnerability-scan]
  # Will fail if 'lint' is skipped due to disabled feature flag
```

Use `always()` in dependent jobs to handle skipped dependencies:

```yaml
test:
  needs: [read-config, check-override, lint, vulnerability-scan]
  if: always() && needs.read-config.outputs.test_enabled == 'true' && ...
```

## Future Enhancements

Potential improvements to the feature flag system:

1. **Repository Variables Integration**: Allow override via GitHub repository variables
2. **Time-based Flags**: Automatically re-enable gates after a specified time
3. **Branch-specific Flags**: Different settings for different branches
4. **Metrics Integration**: Automatically disable gates with high false positive rates
5. **Notification System**: Alert when required gates are disabled
