# Analysis of Sharing httptest.Server in E2E Tests

## Current Implementation

- Each test creates its own `TestEnv` instance with `NewTestEnv(t)` in setup
- `TestEnv` contains a `MockServer` field of type `*httptest.Server`
- During initialization, `startMockServer()` is called to create a new server instance
- Tests configure mock behavior using the `MockConfig` struct:
  ```go
  env.MockConfig.HandleTokenCount = func(req *http.Request) (int, error) {
      return 100000, nil
  }
  ```
- Tests clean up with `defer env.Cleanup()` which closes the MockServer

## Potential Issues with Sharing a Server

1. **Test Isolation**: Each test requires specific server behaviors:
   - `TestMultipleModels`: Customizes server responses based on model name
   - `TestModelError`: Deliberately returns errors for specific models
   - `TestTokenLimit`: Configures token count to exceed limits
   - `TestUserConfirmation`: Configures token count to trigger confirmation prompts

2. **State Management**: Tests modify handler functions and then restore them:
   ```go
   originalHandleTokenCount := env.MockConfig.HandleTokenCount
   env.MockConfig.HandleTokenCount = func(req *http.Request) (int, error) {
       return 100000, nil
   }
   defer func() {
       env.MockConfig.HandleTokenCount = originalHandleTokenCount
   }()
   ```

3. **Concurrency Concerns**: If tests were to run in parallel:
   - Concurrent tests modifying the same handler functions would conflict
   - Response tracking in a shared server would become ambiguous

## Potential Solutions

### 1. Shared Server with Test-Specific Request Routing

A possible approach would be:
- Create a shared MockServer in TestMain
- Enhance the server handler to route requests based on a test identifier:
  ```go
  handler.HandleFunc("/v1beta/models/", func(w http.ResponseWriter, r *http.Request) {
      testID := r.Header.Get("X-Test-ID")
      // Use testID to select the appropriate handler function
  })
  ```
- Tests would need to set a unique header on their requests
- Tests would register their handlers with the shared server

### 2. Shared Server with Register/Unregister Pattern

- Create a shared MockServer in TestMain
- Add a mechanism for tests to "register" their custom handlers:
  ```go
  type HandlerConfig struct {
      HandleTokenCount  func(req *http.Request) (int, error)
      HandleGeneration  func(req *http.Request) (string, int, string, error)
      HandleModelInfo   func(req *http.Request) (string, int, int, error)
  }
  
  func (e *TestEnv) RegisterHandlers(conf HandlerConfig) func() {
      // Store original handlers
      // Set new handlers
      // Return a function that restores original handlers
  }
  ```
- Tests would call this and receive an unregister function to defer

### 3. Keep Separate Servers but Optimize Creation

This is the simplest approach:
- Continue using separate servers for each test
- Optimize the server creation process (if performance is an issue)
- Add support for parallel test execution by ensuring complete isolation

## Recommendation

**The simplest and safest approach is to continue using separate MockServer instances for each test.**

Advantages:
- Maintains complete test isolation
- Avoids complex synchronization mechanisms
- Tests can customize server behavior freely
- Tests can run in parallel (with t.Parallel())

The only disadvantage is a slight overhead from creating multiple servers, but this is likely negligible compared to the complexity of implementing and maintaining a shared server with proper isolation.

## Implementation Decision

Based on this analysis, we recommend **not moving the httptest.Server setup to TestMain** at this time. Instead, focus on ensuring tests are isolated and can run in parallel by continuing to use separate server instances for each test.