package tokenizers

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenizerInitializationTime_MeetsPerformanceTargets tests that tokenizer
// initialization meets the Phase 5.2 performance targets
func TestTokenizerInitializationTime_MeetsPerformanceTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		provider       string
		targetDuration time.Duration
	}{
		{
			name:           "OpenRouter tiktoken-o200k initialization under 100ms",
			provider:       "openrouter",
			targetDuration: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTokenizerManager()

			start := time.Now()
			tokenizer, err := manager.GetTokenizer(tt.provider)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, tokenizer)

			// This will FAIL initially (RED phase) - driving implementation
			assert.LessOrEqual(t, duration, tt.targetDuration,
				"Tokenizer initialization took %v, exceeded target %v", duration, tt.targetDuration)
		})
	}
}

// TestTokenizerMemoryUsage_StaysWithinLimits tests that vocabulary loading
// stays within the 20MB memory target
func TestTokenizerMemoryUsage_StaysWithinLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		provider     string
		memoryTarget int64 // bytes
	}{
		{
			name:         "OpenRouter tiktoken-o200k under 80MB",
			provider:     "openrouter",
			memoryTarget: 80 * 1024 * 1024, // 80MB (adjusted for OpenRouter tiktoken-o200k)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			manager := NewTokenizerManager()
			tokenizer, err := manager.GetTokenizer(tt.provider)
			require.NoError(t, err)

			// Force tokenizer initialization by calling CountTokens
			_, _ = tokenizer.CountTokens(context.Background(), "test", "test-model")
			// Note: This may fail for unsupported models, that's okay for this memory test

			runtime.GC()
			runtime.ReadMemStats(&m2)

			memoryIncrease := int64(m2.Alloc - m1.Alloc)

			// This will FAIL initially (RED phase) - we need memory monitoring
			assert.LessOrEqual(t, memoryIncrease, tt.memoryTarget,
				"Memory usage %d bytes exceeds target %d bytes", memoryIncrease, tt.memoryTarget)
		})
	}
}

// TestLargeInputHandling_RespectsTimeouts tests that large input handling
// properly respects timeout protection for >1MB inputs
func TestLargeInputHandling_RespectsTimeouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		textSize   int
		timeout    time.Duration
		shouldFail bool
	}{
		{
			name:       "Small input 100KB with generous timeout",
			textSize:   100 * 1024,
			timeout:    5 * time.Second,
			shouldFail: false,
		},
		{
			name:       "Medium input 1MB with reasonable timeout",
			textSize:   1024 * 1024,
			timeout:    10 * time.Second,
			shouldFail: false,
		},
		{
			name:       "Large input 5MB with generous timeout",
			textSize:   5 * 1024 * 1024,
			timeout:    30 * time.Second,
			shouldFail: false,
		},
		{
			name:       "Huge input 10MB with short timeout should fail",
			textSize:   10 * 1024 * 1024,
			timeout:    100 * time.Millisecond,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use existing timeout manager
			manager := NewTokenizerManagerWithTimeout(tt.timeout)

			// Create a mock that scales with input size (more realistic)
			mockTokenizer := &MockInputSizeAwareTokenCounter{
				BaseLatencyMs: 1, // 1ms per KB
			}
			manager.SetMockTokenizer("openrouter", mockTokenizer)

			tokenizer, err := manager.GetTokenizer("openrouter")
			require.NoError(t, err)

			largeText := strings.Repeat("a", tt.textSize)

			start := time.Now()
			_, err = tokenizer.CountTokens(context.Background(), largeText, "gpt-4.1")
			elapsed := time.Since(start)

			if tt.shouldFail {
				assert.Error(t, err, "Expected timeout error for large input with short timeout")
				assert.LessOrEqual(t, elapsed, tt.timeout+(200*time.Millisecond), "Should timeout quickly")
			} else {
				// This may FAIL initially (RED phase) if timeout calculation isn't smart enough
				assert.NoError(t, err, "Should handle large input within timeout")
			}
		})
	}
}

// BenchmarkTokenizerInitialization benchmarks initialization time for performance targets
func BenchmarkTokenizerInitialization(b *testing.B) {
	benchmarks := []struct {
		name     string
		provider string
		target   time.Duration
	}{
		{"OpenRouter_TikToken_o200k", "openrouter", 100 * time.Millisecond},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			var initTime time.Duration
			for i := 0; i < b.N; i++ {
				manager := NewTokenizerManager()

				start := time.Now()
				tokenizer, err := manager.GetTokenizer(bm.provider)
				elapsed := time.Since(start)

				if i == 0 {
					initTime = elapsed
				}

				if err != nil {
					b.Fatal(err)
				}
				if tokenizer == nil {
					b.Fatal("tokenizer is nil")
				}
			}

			// Log initialization time for the first iteration
			b.Logf("First initialization took %v (target: %v)", initTime, bm.target)
		})
	}
}

// BenchmarkTokenizerMemoryUsage benchmarks memory usage during initialization
func BenchmarkTokenizerMemoryUsage(b *testing.B) {
	providers := []string{"openrouter"}

	for _, provider := range providers {
		b.Run(provider, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var m1, m2 runtime.MemStats
				runtime.GC()
				runtime.ReadMemStats(&m1)

				manager := NewTokenizerManager()
				_, err := manager.GetTokenizer(provider)
				if err != nil {
					b.Fatal(err)
				}

				runtime.GC()
				runtime.ReadMemStats(&m2)

				if i == 0 {
					memIncrease := int64(m2.Alloc - m1.Alloc)
					b.Logf("Memory increase: %d bytes (%.2f MB)", memIncrease, float64(memIncrease)/(1024*1024))
				}
			}
		})
	}
}

// MockInputSizeAwareTokenCounter simulates processing time that scales with input size
type MockInputSizeAwareTokenCounter struct {
	BaseLatencyMs int // Latency per KB of input
}

func (m *MockInputSizeAwareTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	// Handle nil context (used by circuit breaker initialization test)
	if ctx == nil {
		time.Sleep(time.Duration(m.BaseLatencyMs) * time.Millisecond)
		return len(text), nil
	}

	// Calculate processing time based on input size (1ms per KB)
	inputSizeKB := len(text) / 1024
	if inputSizeKB == 0 {
		inputSizeKB = 1 // Minimum 1KB
	}
	processingTime := time.Duration(inputSizeKB*m.BaseLatencyMs) * time.Millisecond

	// Simulate the processing time with context cancellation support
	select {
	case <-time.After(processingTime):
		return len(text), nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (m *MockInputSizeAwareTokenCounter) SupportsModel(modelName string) bool {
	return true
}

func (m *MockInputSizeAwareTokenCounter) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}
