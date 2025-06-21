# Task: Fix Issues with Running All Models Simultaneously in thinktank

## Executive Summary

When running thinktank with all 15 available models simultaneously, several issues emerge that reduce reliability and prevent 100% success rates. Through testing, I identified specific failure patterns, rate limiting bottlenecks, and error handling deficiencies that need to be addressed to achieve robust multi-model processing.

## Test Results Overview

### Test Configuration
- **Models tested**: All 15 available models from `internal/models/models.go`
- **Input**: `README.md` (338 lines, 12,664 characters)
- **Instructions**: Simple refactoring analysis request
- **Concurrency**: Default 5 simultaneous requests
- **Rate limit**: Default 60 RPM per model

### Success Rate Analysis
- **Consistent failures**: 1 model consistently fails (`openrouter/deepseek/deepseek-r1-0528:free`)
- **Success rate**: 93-94% (14/15 models succeed)
- **Total runtime**: ~3-4 minutes for all models
- **Rate limiting behavior**: Extensive queuing and retry cycles

## Specific Issues Identified

### 1. Model-Specific Failure: `openrouter/deepseek/deepseek-r1-0528:free`

**Issue**: This model consistently fails in batch runs but succeeds when run individually.

**Error Pattern**:
```
Failed model 9/15: openrouter/deepseek/deepseek-r1-0528:free (model openrouter/deepseek/deepseek-r1-0528:free processing failed: model processing failed)
```

**Evidence**:
- Individual run: ✅ Success (61.3s runtime, 2,806 chars output)
- Batch run: ❌ Consistent failure
- API key: Valid (length 73, properly configured)
- Logs: Show successful request/response cycle when run individually

**Root Cause Hypothesis**:
- **Concurrency conflict**: Model may have provider-side concurrency limits
- **Rate limiting interaction**: May be hitting undocumented rate limits when combined with other OpenRouter models
- **Request timeout**: Could be timing out in batch context due to resource contention

### 2. Extensive Rate Limiting Creates Performance Bottlenecks

**Issue**: Current rate limiting strategy causes excessive queuing and delays.

**Observed Pattern**:
```
Processing model 6/15: openrouter/deepseek/deepseek-chat-v3-0324
Completed model 13/15: openrouter/x-ai/grok-3-mini-beta (16.1s)
Rate limited for model 1/15: gpt-4.1 (retry in 16.1s)
Processing model 1/15: gpt-4.1
```

**Problems**:
- Models wait unnecessarily long for rate limiter slots
- 60 RPM limit may be too conservative for some providers
- No differentiation between provider capabilities
- Rate limiting applies globally rather than per-provider

### 3. Insufficient Error Detail and Recovery

**Issue**: Generic error messages don't provide actionable debugging information.

**Current Error Output**:
```
Failed model X/15: model-name (model model-name processing failed: model processing failed)
```

**Missing Information**:
- HTTP status codes from provider APIs
- Specific error messages from provider responses
- Timeout vs network vs authentication vs rate limit differentiation
- Retry attempt details
- Context about concurrent requests when failure occurred

### 4. Provider-Specific Rate Limiting Not Optimized

**Issue**: All providers use the same 60 RPM limit regardless of their actual capabilities.

**Provider Capabilities** (based on documentation):
- **OpenAI**: High rate limits (thousands of RPM for paid tiers)
- **Gemini**: Moderate rate limits (~60 RPM for free tier)
- **OpenRouter**: Varies by model, some have very low limits

**Current State**: One-size-fits-all approach doesn't optimize for provider differences.

## Recommended Solutions

### 1. Fix Model-Specific Failure (Priority: HIGH)

#### Investigation Steps:
```bash
# Test individual vs batch behavior
./thinktank --model "openrouter/deepseek/deepseek-r1-0528:free" --verbose README.md
./thinktank --model "openrouter/deepseek/deepseek-r1-0528:free" --model "openrouter/deepseek/deepseek-chat-v3-0324" README.md
```

#### Code Changes Needed:

**A. Add Provider-Specific Concurrency Limits**
```go
// internal/models/models.go
type ModelInfo struct {
    // ... existing fields
    MaxConcurrentRequests *int `json:"max_concurrent_requests,omitempty"`
}

// Update problematic model definition
"openrouter/deepseek/deepseek-r1-0528:free": {
    // ... existing config
    MaxConcurrentRequests: &[]int{1}[0], // Force sequential processing
}
```

**B. Implement Model-Specific Rate Limiting**
```go
// internal/ratelimit/ratelimit.go
func NewProviderLimiter(provider string, model string) RateLimiter {
    // Check if model has specific limits
    if modelInfo, err := models.GetModelInfo(model); err == nil {
        if modelInfo.MaxConcurrentRequests != nil {
            return NewConcurrencyLimiter(*modelInfo.MaxConcurrentRequests)
        }
    }
    // Fallback to provider defaults
    return getProviderDefaultLimiter(provider)
}
```

### 2. Implement Provider-Aware Rate Limiting (Priority: MEDIUM)

#### Enhanced Rate Limiting Strategy:
```go
// internal/providers/limits.go
var ProviderLimits = map[string]RateLimitConfig{
    "openai": {
        RequestsPerMinute: 500, // Much higher for paid tiers
        ConcurrentRequests: 10,
    },
    "gemini": {
        RequestsPerMinute: 60,  // Conservative for free tier
        ConcurrentRequests: 5,
    },
    "openrouter": {
        RequestsPerMinute: 20,  // Conservative due to model variation
        ConcurrentRequests: 3,  // Lower due to some models having strict limits
    },
}
```

#### CLI Configuration:
```bash
# Allow users to override provider limits
./thinktank --provider-rpm openai:1000,gemini:60,openrouter:30 --models...
```

### 3. Enhanced Error Reporting and Recovery (Priority: MEDIUM)

#### Detailed Error Structure:
```go
// internal/llm/errors.go
type ModelProcessingError struct {
    Model           string        `json:"model"`
    Provider        string        `json:"provider"`
    HTTPStatusCode  int           `json:"http_status_code,omitempty"`
    ProviderError   string        `json:"provider_error,omitempty"`
    ErrorCategory   string        `json:"error_category"` // "timeout", "rate_limit", "auth", "network"
    RetryAttempts   int           `json:"retry_attempts"`
    Duration        time.Duration `json:"duration"`
    ConcurrentReqs  int           `json:"concurrent_requests"` // How many other requests were active
}
```

#### Error Categorization:
```go
func CategorizeError(err error, statusCode int) string {
    switch {
    case statusCode == 429:
        return "rate_limit"
    case statusCode == 401 || statusCode == 403:
        return "authentication"
    case statusCode >= 500:
        return "provider_error"
    case errors.Is(err, context.DeadlineExceeded):
        return "timeout"
    default:
        return "unknown"
    }
}
```

### 4. Automatic Retry with Backoff (Priority: LOW)

#### Intelligent Retry Logic:
```go
// internal/orchestrator/retry.go
type RetryConfig struct {
    MaxAttempts      int
    BaseDelay        time.Duration
    MaxDelay         time.Duration
    RetryableErrors  []string // ["rate_limit", "timeout", "provider_error"]
}

func (o *Orchestrator) processModelWithRetry(model string, prompt string) (*ModelOutput, error) {
    config := getRetryConfig(model)

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        result, err := o.processModel(model, prompt)
        if err == nil {
            return result, nil
        }

        if !isRetryable(err) {
            return nil, fmt.Errorf("non-retryable error: %w", err)
        }

        if attempt < config.MaxAttempts {
            delay := calculateBackoff(attempt, config)
            time.Sleep(delay)
        }
    }

    return nil, fmt.Errorf("failed after %d attempts", config.MaxAttempts)
}
```

### 5. Testing Infrastructure Improvements (Priority: MEDIUM)

#### Comprehensive Test Suite:
```go
// internal/integration/multi_model_test.go
func TestAllModelsSimultaneous(t *testing.T) {
    tests := []struct {
        name           string
        models         []string
        expectedFails  []string // Models we expect to fail
        maxDuration    time.Duration
    }{
        {
            name:          "all_models_basic",
            models:        models.ListAllModels(),
            expectedFails: []string{"openrouter/deepseek/deepseek-r1-0528:free"},
            maxDuration:   5 * time.Minute,
        },
    }
    // ... test implementation
}
```

#### Load Testing:
```bash
# scripts/load_test.sh
#!/bin/bash
for i in {1..5}; do
    echo "Run $i: $(date)"
    ./thinktank --all-models --partial-success-ok README.md
    echo "Success rate: $(calculate_success_rate)"
    sleep 30
done
```

## Implementation Priority

### Phase 1 (Immediate - 1 week)
1. ✅ **Fix `openrouter/deepseek/deepseek-r1-0528:free` failure**
   - Add MaxConcurrentRequests field to ModelInfo
   - Set limit to 1 for problematic model
   - Test batch execution achieves 100% success rate

### Phase 2 (Short-term - 2 weeks)
2. ✅ **Enhanced error reporting**
   - Implement detailed error structure
   - Add error categorization
   - Improve CLI output with actionable details

### Phase 3 (Medium-term - 1 month)
3. ✅ **Provider-aware rate limiting**
   - Implement per-provider rate limit configuration
   - Add CLI flags for user customization
   - Document optimal settings per provider

### Phase 4 (Long-term - 2 months)
4. ✅ **Automatic retry with backoff**
   - Implement intelligent retry logic
   - Add configuration for retry policies
   - Test with various failure scenarios

5. ✅ **Comprehensive testing infrastructure**
   - Create multi-model integration tests
   - Add load testing scripts
   - Set up CI to catch regressions

## Success Criteria

### Immediate Goals:
- [ ] 100% success rate when running all models simultaneously
- [ ] No more "model processing failed" generic errors
- [ ] Runtime under 3 minutes for all 15 models

### Long-term Goals:
- [ ] Sub-2-minute runtime for all models
- [ ] Graceful degradation when providers have issues
- [ ] Clear error messages that guide users to solutions
- [ ] Robust retry behavior for transient failures

## Testing Commands

```bash
# Test current state
./thinktank --instructions test-instructions.txt --model gpt-4.1 --model o4-mini --model gemini-2.5-pro --model gemini-2.5-flash --model o3 --model "openrouter/deepseek/deepseek-chat-v3-0324" --model "openrouter/deepseek/deepseek-r1-0528" --model "openrouter/deepseek/deepseek-chat-v3-0324:free" --model "openrouter/deepseek/deepseek-r1-0528:free" --model "openrouter/meta-llama/llama-3.3-70b-instruct" --model "openrouter/meta-llama/llama-4-maverick" --model "openrouter/meta-llama/llama-4-scout" --model "openrouter/x-ai/grok-3-mini-beta" --model "openrouter/x-ai/grok-3-beta" --model "openrouter/google/gemma-3-27b-it" --partial-success-ok README.md

# Test problematic model individually
./thinktank --instructions test-instructions.txt --model "openrouter/deepseek/deepseek-r1-0528:free" --verbose README.md

# Test with timeout
timeout 300 ./thinktank --instructions test-instructions.txt [all models] README.md
```

## Related Files to Modify

- `internal/models/models.go` - Add concurrency limits
- `internal/ratelimit/ratelimit.go` - Provider-aware limiting
- `internal/llm/errors.go` - Enhanced error types
- `internal/orchestrator/orchestrator.go` - Retry logic
- `cmd/thinktank/cli.go` - New CLI flags
- Tests: `internal/integration/multi_model_test.go`
- Docs: Update README.md with new rate limiting options

This task represents a comprehensive approach to achieving 100% reliability when running all thinktank models simultaneously, with clear priorities and measurable success criteria.
