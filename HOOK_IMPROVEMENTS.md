# Git Hook Improvements

## Current Issues

1. **Slow pre-commit hooks**: Running full test suite and builds on every commit
2. **Poor developer experience**: Long wait times discourage frequent commits
3. **Redundancy**: Same checks run in pre-commit AND CI
4. **Outdated tooling**: golangci-lint v2.1.1 is very old (latest is v1.61.0)

## Improved Strategy

### Pre-commit (Fast, Essential Only)
- **Goal**: Complete in <10 seconds
- **Checks**:
  - Code formatting (`go fmt`)
  - Basic validation (`go vet`)
  - File hygiene (whitespace, large files)
  - Secret detection (ward)
  - YAML validation

### Pre-push (Comprehensive, Before Sharing)
- **Goal**: Complete in <2 minutes
- **Checks**:
  - Full linting (golangci-lint)
  - Complete test suite with race detection
  - Build verification
  - Coverage thresholds
  - Vulnerability scanning (govulncheck)

## Implementation

### Files Created
- `.pre-commit-config.yaml.improved` - Streamlined pre-commit config
- `.pre-push-config.yaml` - Comprehensive pre-push checks
- `scripts/setup-hooks.sh` - Automated setup script

### Usage

```bash
# Apply the improved configuration
cp .pre-commit-config.yaml.improved .pre-commit-config.yaml

# Set up both pre-commit and pre-push hooks
./scripts/setup-hooks.sh

# Test the setup
git add .
git commit -m "test: fast pre-commit"  # Should be fast
git push  # Will run comprehensive checks
```

### Benefits

1. **Faster commits**: Pre-commit completes in seconds, not minutes
2. **Better developer experience**: Encourages frequent, small commits
3. **Comprehensive quality gates**: Nothing reaches remote without full validation
4. **Clear separation**: Fast feedback vs. thorough validation
5. **Modern tooling**: Latest versions of all tools

### Migration Notes

- Current config runs tests/builds in pre-commit (slow)
- New config moves expensive checks to pre-push
- Maintains same quality standards with better UX
- CI pipeline remains unchanged as final safety net

## Troubleshooting

### If pre-commit is slow
```bash
# Skip for urgent commits
git commit --no-verify -m "urgent fix"
```

### If pre-push fails
```bash
# Run checks manually
./scripts/setup-hooks.sh  # Re-run setup
golangci-lint run         # Check linting
go test -race ./...       # Run tests
```

### Update golangci-lint
```bash
# Install latest version
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
```
