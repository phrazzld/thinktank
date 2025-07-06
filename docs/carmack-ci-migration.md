# Carmack-Style CI/CD Migration

## What Changed

We've dramatically simplified the CI/CD setup following John Carmack's philosophy of pragmatic, fast, and simple solutions.

### Before
- **7 workflow files** totaling 100KB+ of YAML
- **897-line main workflow** with complex retry logic
- **15+ pre-commit hooks** taking 8+ minutes
- **Complex quality gates** with override mechanisms
- **Excessive logging** and debugging output

### After
- **1 workflow file** with 45 lines
- **5 pre-commit hooks** taking <5 seconds
- **No retry logic** - fix flaky tests instead
- **Simple pass/fail** - no complex gates
- **Minimal output** - only failures matter

## The New CI Pipeline

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - Checkout
    - Setup Go with caching
    - Verify formatting (fail if needed)
    - Run go vet
    - Run tests with race detection (5min timeout)
    - Build binary

  security:
    # Only on main branch
    # Simple vulnerability scan
```

## Key Principles

1. **Speed First**: CI completes in <2 minutes
2. **Fix, Don't Retry**: No flaky tests allowed
3. **Fail Fast**: Stop on first error
4. **Developer Experience**: Get feedback before context switching

## Maintaining Simplicity

### DO:
- Keep the single workflow file under 100 lines
- Fix flaky tests immediately
- Use Go's built-in tools (go fmt, go vet, go test)
- Run full test suite locally before pushing

### DON'T:
- Add retry mechanisms
- Create complex quality gates
- Add verbose logging
- Split into multiple workflows
- Add coverage requirements to CI

## Local Development

Pre-commit hooks now only do essential checks:
- Fix trailing whitespace
- Fix end-of-file
- Check YAML syntax
- Run go fmt
- Run go vet

Total time: <5 seconds

## Testing Strategy

Follow the existing integration-first approach but with focus on speed:
- Use `go test -short` in CI
- Run full suite locally
- Use `t.Parallel()` aggressively
- 5-minute hard timeout

## Why This Works

1. **Fast feedback** keeps developers in flow
2. **Simple config** means less maintenance
3. **No flaky tests** means reliable CI
4. **Direct approach** uses Go's excellent built-in tooling

## The Carmack Test

> "Can I push code and get feedback before I've context-switched?"

If no, the CI is too slow. Our new setup passes this test.
