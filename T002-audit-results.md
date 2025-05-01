# Secret Detection Test Coverage Audit

## Overview
This audit analyzes the current secret detection test coverage for each provider (Gemini, OpenAI, OpenRouter) and identifies gaps in test coverage for secure handling of API keys and other sensitive information.

## Current Test Coverage

### Common Coverage Across All Providers
All three providers have tests covering:
- Client creation with API key (`provider.CreateClient`)
- Error handling during client creation with an invalid model name

### Test Structure
All provider secret tests:
- Use `logutil.WithSecretDetection` to wrap a buffer logger
- Configure the logger not to panic on detection (using `SetFailOnSecretDetect(false)`)
- Check for API key leakage using `HasDetectedSecrets()`
- Clear detected secrets between test cases using `ClearDetectedSecrets()`

## Coverage Gaps

### 1. OpenRouter Provider
**Issues:**
- Line 52-60 in `provider.go` logs API key prefixes directly:
```go
p.logger.Debug("Using provided API key (length: %d, starts with: %s)", len(effectiveAPIKey), effectiveAPIKey[:min(5, len(effectiveAPIKey))])
```
```go
p.logger.Debug("Using API key from OPENROUTER_API_KEY environment variable (starts with: %s)", effectiveAPIKey[:min(5, len(effectiveAPIKey))])
```
- No tests for API key prefixes not being logged
- No tests covering token authentication in the actual HTTP calls (Authorization header)
- Missing tests for when environment variable is used vs. directly provided key

### 2. All Providers - Missing Test Cases
1. **API Call Execution Tests**
   - No tests verifying that API keys aren't logged during actual API calls to the LLM services
   - No tests covering how Authorization headers are constructed in requests

2. **Parameter Handling**
   - No tests for parameter handling with `SetParameters` method and whether it might accidentally log sensitive information

3. **Client Lifecycle**
   - No tests for lifecycle methods like `Close()` to ensure they don't log sensitive data

4. **Error Path Testing**
   - Limited error path coverage (only invalid model errors)
   - Missing tests for network errors, API errors, timeouts, etc.

5. **Edge Cases**
   - No tests with empty API keys
   - No tests with malformed API keys
   - No tests with extremely long API keys that might get truncated in logs

### 3. Initialization Flow
- Missing secret detection tests for provider initialization and logger creation
- Tests only cover the main API key paths, not initialization configuration

### 4. Client Adapter Testing
Each provider has a client adapter pattern (`GeminiClientAdapter`, `OpenAIClientAdapter`, etc.) that lacks dedicated secret detection tests.

## Recommended New Tests

Based on the identified gaps, the following new test cases should be implemented (T003):

### For All Providers:
1. Test that API key prefixes/fragments are never logged (critical for OpenRouter)
2. Test that API keys aren't leaked during actual API calls
3. Test that environment variable fallbacks are handled securely
4. Test adapter parameter handling for potential secret leakage
5. Test error handling for network and API errors
6. Test edge cases (empty/malformed keys)
7. Test that the Authorization header is properly protected in debug logs
8. Test that client adapter methods maintain secret security

### Priority Focus Areas:
1. OpenRouter provider requires immediate fixes to stop logging key prefixes
2. API call execution tests are needed for all providers
3. Error path testing should be expanded
4. Edge cases with key handling should be validated

## Conclusion
While basic secret detection tests exist, there are significant gaps in coverage, particularly around actual API calls, error handling, and the OpenRouter provider implementation. Implementing the recommended tests will ensure comprehensive protection against secret leakage in logs.
