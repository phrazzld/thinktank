# Testing Guide

This document describes the established testing patterns, infrastructure, and best practices for the thinktank project.

## Core Testing Principles

### 1. No Internal Mocking

**Fundamental Rule**: We do NOT mock internal components within our application boundaries.

- **Mock only external dependencies**: HTTP APIs, filesystem operations, environment variables
- **Use real implementations**: For all internal packages and components
- **Test integration points**: Focus on how components work together, not in isolation

This approach ensures tests validate actual component interactions and catch integration issues.

### 2. Boundary Testing Pattern

We follow a "boundary testing" approach where external system boundaries are abstracted and mocked:

```go
// External boundaries that we mock
type ExternalAPICaller interface {
    Call(ctx context.Context, request *http.Request) (*http.Response, error)
}

type FilesystemIO interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm int) error
    // ... other filesystem operations
}

type EnvironmentProvider interface {
    GetEnv(key string) string
    SetEnv(key, value string) error
}
```

### 3. Test-Driven Development

- Write tests first, make them fail, then implement code to pass
- Focus on integration and workflow tests over unit tests
- Maintain 90%+ code coverage across all packages

## Testing Infrastructure

### Overview

The testing infrastructure is located in `/internal/testutil/` and provides:

- **Data Factories**: Builder patterns for creating test data
- **Mock Implementations**: For external dependencies only
- **HTTP Mocking**: For testing external API interactions
- **Filesystem Abstraction**: In-memory filesystem for tests
- **Integration Utilities**: Temporary files/directories, cleanup
- **Test Scenarios**: Common test setups and assertions

### Key Components

```text
internal/testutil/
├── factories.go          # Data builder patterns
├── mocklogger.go         # Logger and audit logger mocks
├── mockregistry.go       # Registry mock (internal interface)
├── providers.go          # HTTP server mocking utilities
├── integration.go        # Temporary file/directory utilities
├── memfs.go             # In-memory filesystem implementation
├── realfs.go            # Real filesystem implementation
├── testfixtures.go      # Standard test data constants
├── testscenarios.go     # Common test scenario helpers
└── security.go          # Secure API key handling in tests
```

## Test Data Factories

Use builder patterns to create test data with sensible defaults and easy customization:

### Basic Usage

```go
// Create a test model with defaults
model := testutil.NewModelDefinition().Build()

// Customize specific fields
model := testutil.NewModelDefinition().
    WithName("custom-model").
    WithProvider("openai").
    WithContextWindow(8192).
    Build()

// Create invalid data for error testing
model := testutil.NewModelDefinition().
    InvalidName().  // Sets empty name
    Build()
```

### Available Builders

- `ProviderDefinitionBuilder`: For creating provider configurations
- `ModelDefinitionBuilder`: For creating model configurations
- `ParameterDefinitionBuilder`: For model parameters
- `SafetyBuilder`: For safety/content filtering results
- `ProviderResultBuilder`: For LLM response data
- `ChatCompletionMessageBuilder`: For chat messages
- `ChatCompletionRequestBuilder`: For API requests
- `ChatCompletionResponseBuilder`: For API responses

### Builder Pattern Examples

```go
// Provider with custom base URL
provider := testutil.NewProviderDefinition().
    WithName("custom-provider").
    WithBaseURL("https://api.custom.example.com/v1").
    Build()

// Model with Gemini-specific parameters
model := testutil.NewModelDefinition().
    WithName("gemini-pro").
    WithProvider("gemini").
    WithGeminiParameters().
    Build()

// LLM response with safety blocking
response := testutil.NewProviderResult().
    SafetyBlocked().
    Build()

// Chat completion request with custom parameters
request := testutil.NewChatCompletionRequest().
    WithModel("gpt-4").
    WithTemperature(0.3).
    AddMessage(testutil.NewChatCompletionMessage().
        AsSystem().
        WithContent("You are a helpful assistant").
        Build()).
    Build()
```

## Mock Implementations

### MockLogger

Implements both `logutil.LoggerInterface` and `auditlog.AuditLogger` interfaces:

```go
func TestLogging(t *testing.T) {
    logger := testutil.NewMockLogger()

    // Use logger in your code
    logger.InfoContext(ctx, "Processing request")
    logger.ErrorContext(ctx, "Failed to process: %v", err)

    // Verify logging in tests
    messages := logger.GetInfoMessages()
    if len(messages) != 1 {
        t.Errorf("Expected 1 info message, got %d", len(messages))
    }

    // Check for specific content
    if !logger.ContainsMessage("Processing request") {
        t.Error("Expected log message not found")
    }

    // Verify correlation ID propagation
    entries := logger.GetLogEntriesByCorrelationID("test-correlation-id")
    if len(entries) == 0 {
        t.Error("Expected log entries with correlation ID")
    }
}
```

### MockRegistry

Implements `registry.Registry` interface with configurable behavior:

```go
func TestWithMockRegistry(t *testing.T) {
    registry := testutil.NewMockRegistry()

    // Add test data
    registry.AddModel(testutil.NewModelDefinition().
        WithName("test-model").
        WithProvider("test-provider").
        Build())

    registry.AddProvider(testutil.NewProviderDefinition().
        WithName("test-provider").
        Build())

    // Configure error cases
    registry.SetGetModelError(errors.New("model not found"))

    // Use registry in your code
    model, err := registry.GetModel(ctx, "test-model")

    // Verify method calls
    calls := registry.GetMethodCalls("GetModel")
    if len(calls) != 1 {
        t.Errorf("Expected 1 GetModel call, got %d", len(calls))
    }
}
```

## HTTP Mocking

For testing external API calls:

```go
func TestExternalAPI(t *testing.T) {
    // Create mock HTTP server
    server := testutil.SetupMockHTTPServer(t)

    // Configure responses
    server.AddJSONHandler("/v1/chat/completions", 200, map[string]interface{}{
        "choices": []map[string]interface{}{
            {
                "message": map[string]interface{}{
                    "content": "Test response",
                },
                "finish_reason": "stop",
            },
        },
    })

    // Configure your client to use server.URL
    client := createClientWithBaseURL(server.URL)

    // Test your code
    response, err := client.CreateCompletion(ctx, request)
    // ... assertions
}
```

### HTTP Mock Helpers

```go
// Standard response helpers
successResponse := testutil.CreateHTTPSuccessResponse("Test content")
errorResponse := testutil.CreateHTTPErrorResponse("rate_limit", "Too many requests")
authErrorResponse := testutil.CreateHTTPAuthErrorResponse()

// Advanced handlers
server.AddAuthHandler("/secure", "Bearer token123", 200, response)
server.AddMethodHandler("/api", "POST", 200, response)
server.AddSlowHandler("/slow", 200, response, func() {
    time.Sleep(100 * time.Millisecond)
})
server.AddMalformedJSONHandler("/broken")  // For error testing
```

## Filesystem Abstraction

Use `FilesystemIO` interface to make code testable:

### In Production Code

```go
type MyService struct {
    fs FilesystemIO
}

func (s *MyService) SaveOutput(ctx context.Context, path string, content []byte) error {
    return s.fs.WriteFileWithContext(ctx, path, content, 0644)
}
```

### In Tests

```go
func TestSaveOutput(t *testing.T) {
    // Use in-memory filesystem
    fs := testutil.NewMemFS()
    service := &MyService{fs: fs}

    // Create test directory
    fs.MkdirAll("/test/output", 0755)

    // Test the service
    err := service.SaveOutput(ctx, "/test/output/file.txt", []byte("content"))
    if err != nil {
        t.Fatalf("Failed to save output: %v", err)
    }

    // Verify file was created
    if !fs.FileExists("/test/output/file.txt") {
        t.Error("Expected file was not created")
    }

    // Verify content
    content, err := fs.ReadFile("/test/output/file.txt")
    if err != nil {
        t.Fatalf("Failed to read file: %v", err)
    }
    if string(content) != "content" {
        t.Errorf("Expected 'content', got %q", string(content))
    }
}
```

### MemFS vs RealFS

- **MemFS**: Use in tests for speed and isolation
- **RealFS**: Use in production and integration tests that need real filesystem

```go
// In tests
fs := testutil.NewMemFS()

// In production
fs := testutil.NewRealFS()

// Both implement the same FilesystemIO interface
```

## Integration Testing

### Temporary Resources

```go
func TestWithTempResources(t *testing.T) {
    // Create temporary directory
    tempDir := testutil.SetupTempDir(t, "test-output")

    // Create temporary file
    tempFile, file := testutil.SetupTempFile(t, "test", ".txt")
    defer file.Close()

    // Create test files with content
    files := map[string][]byte{
        "input.txt":  []byte("test input"),
        "config.yaml": []byte("key: value"),
    }
    createdFiles := testutil.CreateTestFiles(t, tempDir, files)

    // Resources are automatically cleaned up via t.Cleanup()
}
```

### Boundary Testing Pattern

```go
func TestWithBoundaries(t *testing.T) {
    // Use integration test helper
    IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
        // Configure mock external boundaries
        env.MockFS.MkdirAll("/output", 0755)
        env.MockAPI.AddJSONHandler("/v1/models", 200, modelsResponse)

        // Set environment variables
        env.MockEnv.SetEnv("API_KEY", "test-key-123")

        // Run your integration test
        err := env.Run(ctx, "test instructions")
        if err != nil {
            t.Fatalf("Integration test failed: %v", err)
        }

        // Verify results using boundary mocks
        files := env.MockFS.GetFileContent()
        if len(files) == 0 {
            t.Error("Expected output files to be created")
        }
    })
}
```

## Context and Correlation ID Handling

### Context Propagation

Always use context in tests to validate proper propagation:

```go
func TestContextPropagation(t *testing.T) {
    // Create context with correlation ID
    ctx := logutil.WithCorrelationID(context.Background(), "test-123")

    // Use in your code
    result, err := service.ProcessRequest(ctx, request)

    // Verify correlation ID was propagated in logs
    logger := testutil.NewMockLogger()
    entries := logger.GetLogEntriesByCorrelationID("test-123")
    if len(entries) == 0 {
        t.Error("Expected log entries with correlation ID")
    }
}
```

### Context Deadlines

```go
func TestContextDeadlines(t *testing.T) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    // Test that your code respects context cancellation
    err := service.LongRunningOperation(ctx)
    if err == nil {
        t.Error("Expected context deadline exceeded error")
    }
    if !errors.Is(err, context.DeadlineExceeded) {
        t.Errorf("Expected context.DeadlineExceeded, got %v", err)
    }
}
```

## Security in Tests

### API Key Handling

```go
func TestAPIKeyHandling(t *testing.T) {
    // Ensure test environment
    testutil.EnsureTestEnvironment(t)

    // Get test API key (enforces "test-" prefix)
    apiKey := testutil.GetTestAPIKey(t, "OPENAI_API_KEY")

    // Use in tests
    client := createClient(apiKey)

    // Verify secure handling
    if !strings.HasPrefix(apiKey, "test-") {
        t.Error("API key should have test prefix for safety")
    }
}
```

### Test Configuration Validation

```go
func TestSecureConfig(t *testing.T) {
    config := testutil.CreateSecureTestConfig(map[string]string{
        "api_key": "test-key-123",
        "base_url": "https://api.test.example.com",
    })

    // Validate configuration
    err := testutil.ValidateTestConfiguration(config)
    if err != nil {
        t.Fatalf("Test configuration validation failed: %v", err)
    }
}
```

## Common Patterns and Examples

### Table-Driven Tests with Builders

```go
func TestModelProcessing(t *testing.T) {
    tests := []struct {
        name     string
        model    registry.ModelDefinition
        wantErr  bool
    }{
        {
            name: "valid model",
            model: testutil.NewModelDefinition().
                WithName("gpt-4").
                WithProvider("openai").
                Build(),
            wantErr: false,
        },
        {
            name: "invalid model name",
            model: testutil.NewModelDefinition().
                InvalidName().
                Build(),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := processor.ProcessModel(ctx, tt.model)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessModel() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Multi-Provider Testing

```go
func TestMultiProviderIsolation(t *testing.T) {
    registry := testutil.NewMockRegistry()

    // Add multiple providers
    providers := []string{"openai", "gemini", "openrouter"}
    for _, provider := range providers {
        registry.AddProvider(testutil.NewProviderDefinition().
            WithName(provider).
            Build())

        registry.AddModel(testutil.NewModelDefinition().
            WithProvider(provider).
            WithName(provider + "-model").
            Build())
    }

    // Test API key isolation
    for _, provider := range providers {
        client, err := registry.CreateLLMClient(ctx, "test-"+provider+"-key", provider+"-model")
        if err != nil {
            t.Errorf("Failed to create client for %s: %v", provider, err)
        }
        // ... test provider-specific behavior
    }
}
```

### Error Scenario Testing

```go
func TestErrorScenarios(t *testing.T) {
    scenarios := []struct {
        name           string
        setupMock      func(*testutil.MockHTTPServer)
        expectedError  string
    }{
        {
            name: "rate limit exceeded",
            setupMock: func(server *testutil.MockHTTPServer) {
                server.AddJSONHandler("/v1/chat/completions", 429,
                    testutil.CreateHTTPRateLimitResponse())
            },
            expectedError: "rate limit exceeded",
        },
        {
            name: "authentication failed",
            setupMock: func(server *testutil.MockHTTPServer) {
                server.AddJSONHandler("/v1/chat/completions", 401,
                    testutil.CreateHTTPAuthErrorResponse())
            },
            expectedError: "authentication",
        },
    }

    for _, scenario := range scenarios {
        t.Run(scenario.name, func(t *testing.T) {
            testutil.WithMockProvider(t, scenario.setupMock, func(baseURL string) {
                client := createClientWithBaseURL(baseURL)
                _, err := client.CreateCompletion(ctx, request)

                if err == nil {
                    t.Error("Expected error but got none")
                }
                if !strings.Contains(err.Error(), scenario.expectedError) {
                    t.Errorf("Expected error containing %q, got %q",
                        scenario.expectedError, err.Error())
                }
            })
        })
    }
}
```

## Best Practices

### 1. Test Organization

- Group related tests in the same file
- Use descriptive test names that explain the scenario
- Use table-driven tests for multiple similar test cases
- Use subtests (`t.Run`) to organize complex test scenarios

### 2. Resource Management

- Always use `testutil` helpers for temporary resources
- Let `t.Cleanup()` handle resource cleanup automatically
- Don't manually manage temporary files/directories

### 3. Assertion Patterns

```go
// Good: Specific error checking
if err == nil {
    t.Fatal("Expected error but got none")
}
if !strings.Contains(err.Error(), "expected message") {
    t.Errorf("Expected error containing 'expected message', got %q", err.Error())
}

// Good: Use integration helpers for content verification
VerifyFileContent(t, env, expectedPath, expectedContent)

// Good: Check mock interactions
calls := mockRegistry.GetMethodCalls("GetModel")
if len(calls) != 1 {
    t.Errorf("Expected 1 GetModel call, got %d", len(calls))
}
```

### 4. Context Usage

- Always pass context to functions that accept it
- Test context cancellation and timeouts
- Verify correlation ID propagation through the call stack

### 5. Coverage and Quality

- Maintain 90%+ test coverage
- Test both success and error paths
- Use builders to test invalid data scenarios
- Focus on integration tests over isolated unit tests

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

### 3. Check Coverage
```bash
# Check overall coverage
./scripts/check-coverage.sh 90

# Check package-specific coverage
./scripts/check-package-coverage.sh 90
```

This testing infrastructure ensures reliable, maintainable tests that validate real system behavior while providing fast feedback during development.
