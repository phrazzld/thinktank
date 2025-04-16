# TODO-FIX: Tasks to Fix Build After Parameter Constraints Implementation

## Interface and Method Signature Issues

- [x] Update `GenerateContent` method in all provider client implementations (openai, gemini) to use the new signature: `GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*ProviderResult, error)`

- [x] Fix `APIServiceAdapter` implementation in `adapters.go` to match interfaces (no actual duplicate methods found):
  - Update method signatures to match `interfaces.APIService`
  - Add missing required methods from the interface
  - Fix type assertions for delegation

- [ ] Update all mock and test implementations of `GenerateContent` to use the new method signature with parameters

## Adapter Pattern Issues

- [ ] Fix `TokenManagerAdapter` implementation in `adapters.go`:
  - Update method signatures to match `interfaces.TokenManager`
  - Fix type assertions and method delegation

- [ ] Fix `ContextGathererAdapter` implementation:
  - Update parameter types to match internal vs interface types
  - Fix `ContextStats` type compatibility issue

- [ ] Fix `FileWriterAdapter` implementation to match interface

## Architecture Alignment

- [ ] Align internal package types with interface package types:
  - Make `TokenResult` in internal package compatible with `interfaces.TokenResult`
  - Make `ContextStats` in internal package compatible with `interfaces.ContextStats`
  - Make `GatherConfig` in internal package compatible with `interfaces.GatherConfig`

- [ ] Update constructor arguments to match new signatures:
  - Update all `NewTokenManager` calls to include registry parameter
  - Fix adapter struct initialization in `app.go`

## Test Fixes

- [ ] Fix test helpers in `cmd/architect/api_test_helper.go` and `internal/architect/api_test_helper.go`

- [ ] Update mock implementations in test files to support the new interface methods

- [ ] Fix type assertions in tests to accommodate new interface methods

## Integration Fixes

- [ ] Fix error: `too many arguments in call to NewTokenManager` in app.go

- [ ] Fix error: `undefined: APIServiceAdapter` in app.go

- [ ] Fix error: `undefined: TokenManagerAdapter` in app.go

## Verification Steps

- [ ] Run `go fmt ./...` to format all code

- [ ] Run `go vet ./...` to check for any subtle issues

- [ ] Run package-specific tests to verify each component:
  - [ ] `go test ./internal/registry` (already passing)
  - [ ] `go test ./internal/architect`
  - [ ] `go test ./internal/gemini`
  - [ ] `go test ./internal/openai`

- [ ] Run all tests: `go test ./...`

- [ ] Build the application: `go build`

- [ ] Verify runtime functionality with a simple test command:
  `go run main.go --verbose --instructions test.txt --model gpt-4.1-mini --model gemini-2.0-flash ./`