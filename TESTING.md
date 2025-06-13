# Testing Guide

This document outlines the testing strategy, patterns, and infrastructure for the thinktank project, following our core tenets of testability, simplicity, and maintainability.

## Core Testing Principles

### 1. No Internal Mocking
- **Never mock internal components** - only mock external system boundaries
- Use real implementations for all internal collaborators
- Mock only: HTTP APIs, file systems, databases, system clock
- If testing is hard, refactor the code for better testability

### 2. Integration-Focused Testing
- Prefer integration tests over isolated unit tests
- Test component interactions with real implementations
- Focus on observable behavior, not implementation details
- Test end-to-end workflows whenever possible

### 3. Behavior-Driven Testing
- Test what the code should do, not how it does it
- Verify outcomes and side effects, not internal calls
- Write tests that survive refactoring
- Tests should serve as living documentation

## API Key Management Strategy

### Test Environment Security

#### 1. Test API Keys in CI/CD
Our CI pipeline uses **hardcoded test API keys** which is the recommended approach:

```yaml
# .github/workflows/ci.yml
-e GEMINI_API_KEY=test-api-key \
-e OPENAI_API_KEY=test-api-key \
-e OPENROUTER_API_KEY=test-api-key \
```

**Why this approach:**
- ✅ No real API keys exposed in CI logs
- ✅ No dependency on repository secrets for basic testing
- ✅ Predictable, deterministic test behavior
- ✅ No risk of accidental production API usage

#### 2. Local Development API Keys
For local testing that requires real API integration:

```bash
# Optional: Set environment variables for manual API testing
export OPENAI_TEST_KEY=test-openai-key-12345
export GEMINI_TEST_KEY=test-gemini-key-67890
export OPENROUTER_TEST_KEY=test-openrouter-key-11111

# Run manual API tests (requires build tag)
go test -tags=manual_api_test ./internal/e2e/...
```

**Security Requirements:**
- Test keys **MUST** have `test-` prefix
- Never commit real API keys to repository
- Use dedicated test accounts with usage limits
- Test keys should have minimal permissions

#### 3. Test API Key Validation
Use the secure test configuration helper:

```go
import "github.com/phrazzld/thinktank/internal/testutil"

func TestWithAPIKey(t *testing.T) {
    // Safely retrieve test API key
    apiKey := testutil.GetTestAPIKey(t, "OPENAI_TEST_KEY")

    // Key will be validated as test-only
    provider := openai.NewProvider(logger)
    client, err := provider.CreateClient(ctx, apiKey, "gpt-4", "")
    // ...
}
```

### External API Mocking Strategy

#### 1. HTTP Test Servers
For provider testing, use in-memory HTTP servers:

```go
func TestProviderAPI(t *testing.T) {
    // Create mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Simulate provider API responses
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(mockResponse)
    }))
    defer server.Close()

    // Use real provider implementation with mock server
    provider := openai.NewProvider(logger)
    client, err := provider.CreateClient(ctx, "test-api-key", "gpt-4", server.URL)

    // Test real behavior against mock API
    result, err := client.GenerateContent(ctx, "test prompt", nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Content)
}
```

#### 2. In-Memory Implementations
For complex testing scenarios:

```go
// Example: In-memory provider for integration tests
type InMemoryProvider struct {
    responses map[string]*llm.ProviderResult
    errors    map[string]error
}

func (p *InMemoryProvider) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
    if err, exists := p.errors[prompt]; exists {
        return nil, err
    }
    if resp, exists := p.responses[prompt]; exists {
        return resp, nil
    }
    return testutil.BasicSuccessResponse, nil
}
```

## Test Infrastructure

### 1. Test Utilities (`internal/testutil`)

#### File System Helpers
```go
// Create temporary test directories with automatic cleanup
func setupTestDirectory(t *testing.T) string {
    dir, err := os.MkdirTemp("", "thinktank-test-*")
    require.NoError(t, err)

    t.Cleanup(func() {
        os.RemoveAll(dir)
    })

    return dir
}
```

#### Provider Test Factories
```go
// Use standardized test data
func TestProviderConfig(t *testing.T) {
    // Valid configuration
    config := testutil.OpenAIProvider

    // Invalid configurations for error testing
    invalidConfig := testutil.OpenAIProvider
    invalidConfig.BaseURL = "invalid-url"
}
```

#### Mock Server Utilities
```go
// Set up provider mock server
func setupProviderMockServer(t *testing.T) *httptest.Server {
    return testutil.NewProviderMockServer(t, testutil.ProviderMockConfig{
        SuccessResponse: testutil.BasicSuccessResponse,
        ErrorResponses: map[string]error{
            "auth-error": testutil.CreateAuthError("openai"),
        },
    })
}
```

### 2. Test Logging (`internal/logutil`)

#### Structured Test Logging
```go
func TestWithLogging(t *testing.T) {
    logger := logutil.NewTestLogger()
    logger.SetCorrelationID("test-" + t.Name())

    t.Cleanup(func() {
        // Verify no error logs during test
        if logger.HasErrors() {
            t.Errorf("Test generated error logs: %v", logger.GetErrors())
        }
    })

    // Use logger in test...
}
```

#### Secret Detection Testing
```go
func TestSecretHandling(t *testing.T) {
    testLogger := logutil.NewBufferLogger(logutil.DebugLevel)
    secretLogger := logutil.WithSecretDetection(testLogger)
    secretLogger.SetFailOnSecretDetect(false)

    // Perform operation that should not leak secrets
    provider.CreateClient(ctx, "test-api-key", "model", "")

    // Verify no secrets were logged
    if secretLogger.HasDetectedSecrets() {
        t.Errorf("API key leaked in logs: %v", secretLogger.GetDetectedSecrets())
    }
}
```

### 3. Property-Based Testing

#### Library Choice: Rapid

We use **[Rapid](https://github.com/flyingmutant/rapid)** (`pgregory.net/rapid`) for property-based testing.

**Why Rapid was chosen over alternatives like Gopter:**

- **Modern API**: Leverages Go generics for type-safe data generation
- **Simplicity**: Intuitive API with minimal boilerplate compared to traditional QuickCheck-style libraries
- **Automatic shrinking**: Intelligently minimizes failing test cases without requiring user configuration
- **Active maintenance**: Recent releases and ongoing development (v1.2.0, February 2025)
- **No dependencies**: Clean, minimal footprint aligning with our project values
- **Performance**: Optimized for quick feedback loops during development

#### Usage Patterns

For algorithmic functions and data processing:

```go
import "pgregory.net/rapid"

func TestContentProcessing_Properties(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate arbitrary input using type-safe generators
        content := rapid.String().Draw(t, "content")
        maxLength := rapid.IntRange(1, 1000).Draw(t, "maxLength")

        // Process with real implementation
        processed := processContent(content, maxLength)

        // Verify invariants hold for all generated inputs
        assert.LessOrEqual(t, len(processed), maxLength)
        assert.True(t, strings.Contains(processed, extractKey(content)))
    })
}
```

#### Testing Data Structures

```go
func TestConfigValidation_Properties(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate arbitrary configuration
        config := ProviderConfig{
            Temperature: rapid.Float64Range(0.0, 2.0).Draw(t, "temperature"),
            MaxTokens:   rapid.IntRange(1, 4096).Draw(t, "maxTokens"),
            TopP:        rapid.Float64Range(0.0, 1.0).Draw(t, "topP"),
        }

        // Validate configuration
        err := validateConfig(config)

        // Assert invariants about validation
        assert.NoError(t, err) // All generated configs should be valid
    })
}
```

#### When to Use Property-Based Testing

- **Data processing algorithms**: Functions that transform input data
- **Configuration validation**: Parameter boundary checking
- **Parsing and serialization**: Round-trip testing for data formats
- **Mathematical operations**: Numeric calculations with invariants
- **Text processing**: String manipulation functions

## Test Categories

### 1. Unit Tests
- **Focus**: Pure functions, data processing, algorithms
- **Mocking**: External dependencies only
- **Coverage Target**: 90%+ for critical business logic

### 2. Integration Tests
- **Focus**: Component interactions, workflow testing
- **Mocking**: HTTP APIs, file system, external services
- **Coverage Target**: 90%+ for integration paths

### 3. End-to-End Tests
- **Focus**: Complete application workflows
- **Execution**: Containerized environment with compiled binary
- **API Keys**: Test-only keys or mock servers

### 4. Manual API Tests
- **Purpose**: Validate real provider integration
- **Execution**: `go test -tags=manual_api_test`
- **Requirements**: Real test API keys with `test-` prefix

## CI/CD Integration

### Coverage Enforcement
```bash
# Overall coverage threshold (current: 35%, target: 90%)
./scripts/check-coverage.sh 35

# Package-specific coverage enforcement
./scripts/ci/check-package-specific-coverage.sh
```

### Quality Gates
- **Lint**: Code formatting and style checks (required)
- **Test**: Unit, integration, and coverage validation (required)
- **Security**: Secret scanning with TruffleHog (required)
- **Vulnerability**: Dependency scanning with govulncheck (required)

### Emergency Override
Use emergency override labels for critical hotfixes:
- `emergency-override`: Bypass all quality gates
- `bypass-tests`: Skip test execution only
- `bypass-coverage`: Skip coverage enforcement only

## Security Guidelines

### 1. Secret Detection
- All tests must pass secret detection validation
- Use `logutil.WithSecretDetection()` in provider tests
- Never log raw API keys or sensitive data

### 2. Test Key Management
- Test keys must have `test-` prefix
- Use environment variables, never hardcode
- Implement key validation in test helpers
- Use minimal permissions for test accounts

### 3. Production Isolation
- Separate test and production environments completely
- Never use production keys in any test environment
- Use mock servers for standard test suites
- Reserve real API calls for manual integration tests only

## Common Patterns

### Error Scenario Testing
```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name           string
        serverResponse func(w http.ResponseWriter, r *http.Request)
        expectedError  string
    }{
        {
            name: "API key invalid",
            serverResponse: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusUnauthorized)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "Invalid API key",
                })
            },
            expectedError: "Invalid API key",
        },
        // ... more error scenarios
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
            defer server.Close()

            // Test with real provider implementation
            provider := openai.NewProvider(logger)
            client, err := provider.CreateClient(ctx, "test-api-key", "gpt-4", server.URL)
            require.NoError(t, err)

            _, err = client.GenerateContent(ctx, "test prompt", nil)
            assert.Contains(t, err.Error(), tt.expectedError)
        })
    }
}
```

### Table-Driven Parameter Testing
```go
func TestParameterValidation(t *testing.T) {
    tests := []struct {
        name        string
        params      map[string]interface{}
        expectError bool
        errorMsg    string
    }{
        {
            name:        "Valid temperature",
            params:      map[string]interface{}{"temperature": 0.7},
            expectError: false,
        },
        {
            name:        "Temperature too high",
            params:      map[string]interface{}{"temperature": 2.5},
            expectError: true,
            errorMsg:    "temperature must be between 0.0 and 2.0",
        },
        // ... more parameter tests
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := setupProvider(t)
            client := setupClient(t, provider)

            _, err := client.GenerateContent(ctx, "test prompt", tt.params)

            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Getting Started

### 1. Run Standard Tests
```bash
# Run all tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/providers/openai/...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2. Run Integration Tests
```bash
# Run integration tests
go test -v ./internal/integration/...

# Run E2E tests (requires compiled binary)
./internal/e2e/run_e2e_tests.sh -v
```

### 3. Run Manual API Tests
```bash
# Set test API keys (optional)
export OPENAI_TEST_KEY=test-openai-key-12345
export GEMINI_TEST_KEY=test-gemini-key-67890

# Run manual API tests
go test -tags=manual_api_test -v ./internal/e2e/...
```

### 4. Check Coverage
```bash
# Check overall coverage
./scripts/check-coverage.sh 90

# Check package-specific coverage
./scripts/check-package-coverage.sh 90
```

This testing guide ensures our codebase maintains high quality while following our core development tenets of testability, simplicity, and maintainability.
