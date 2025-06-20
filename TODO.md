# TODO: PR #97 Fixes

## Critical Issues
- [x] Fix remaining test method signatures in `cmd/thinktank/cli_benchmark_test.go` - some still use old `ModelCompleted` signature
- [x] Add error logging when terminal width detection fails in `internal/logutil/layout.go:calculateOptimalWidth()`

## CI Failure Resolution (High Priority)
- [x] Fix failing `TestCIEnvironmentDetection` test expectation in `internal/cli/flags_integration_test.go:345` - change expected text from "Starting processing" to "Processing" to match modern CLI output format
- [x] Run CLI tests locally to verify all 7 CI environment detection subtests pass after fix
- [x] Search for other tests with outdated output expectations using `grep -r "Starting processing" internal/` and update any found
- [x] Verify CI environment detection still properly validates all CI variables (CI, GITHUB_ACTIONS, GITLAB_CI, TRAVIS, CIRCLECI, JENKINS_URL, CONTINUOUS_INTEGRATION)
- [x] Run full test suite `go test ./...` to ensure no regressions introduced by test expectation changes

## Regression Fixes (High Priority)
- [x] Fix TestConsoleWriter_OutputFormatting failures - update symbol/color expectations to match modern output (multiple test cases failing)
- [x] Fix TestConsoleWriter_ErrorWarningSuccessMessages failures - update emoji expectations to match Unicode symbols
- [x] Fix TestConsoleWriter_Modern* test failures - update bullet point expectations from ● to * for CI mode

## Nice to Have
- [ ] Add `THINKTANK_ASCII_ONLY=true` env var to force ASCII output when Unicode detection is wrong
- [ ] Cache terminal width instead of calling `terminal.GetSize()` every time

## Ship It
- [ ] Merge the PR
